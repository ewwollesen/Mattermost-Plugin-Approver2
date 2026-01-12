package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/command"
	"github.com/mattermost/mattermost-plugin-approver2/server/store"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
// The root URL is currently <siteUrl>/plugins/com.mattermost.plugin-starter-template/api/v1/. Replace com.mattermost.plugin-starter-template with the plugin ID.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	// Dialog submission endpoint (no auth required - handled by Mattermost)
	router.HandleFunc("/dialog/submit", p.handleDialogSubmit).Methods(http.MethodPost)

	// Middleware to require that the user is logged in for API routes
	router.Use(p.MattermostAuthorizationRequired)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/hello", p.HelloWorld).Methods(http.MethodGet)

	router.ServeHTTP(w, r)
}

func (p *Plugin) MattermostAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) HelloWorld(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Hello, world!")); err != nil {
		p.API.LogError("Failed to write response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleDialogSubmit processes dialog submissions
func (p *Plugin) handleDialogSubmit(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		p.API.LogError("Failed to read dialog submission body", "error", err.Error())
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			p.API.LogError("Failed to close request body", "error", closeErr.Error())
		}
	}()

	// Parse dialog submission payload
	payload, err := command.ParseDialogSubmissionPayload(body)
	if err != nil {
		p.API.LogError("Failed to parse dialog submission", "error", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate submission based on callback ID
	var response *model.SubmitDialogResponse
	switch payload.CallbackId {
	case "approve_new":
		response = p.handleApproveNew(payload)
	default:
		p.API.LogWarn("Unknown dialog callback ID", "callback_id", payload.CallbackId)
		response = &model.SubmitDialogResponse{
			Error: "Unknown dialog type",
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.API.LogError("Failed to encode dialog response", "error", err.Error())
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
}

// handleApproveNew creates and saves a new approval request
func (p *Plugin) handleApproveNew(payload *model.SubmitDialogRequest) *model.SubmitDialogResponse {
	// First validate the submission
	response := command.HandleDialogSubmission(payload.Submission)
	if len(response.Errors) > 0 {
		return response
	}

	// Extract validated data
	approverID := payload.Submission["approver"].(string)
	description := payload.Submission["description"].(string)

	// Get requester info
	requester, appErr := p.API.GetUser(payload.UserId)
	if appErr != nil {
		p.API.LogError("Failed to get requester user", "user_id", payload.UserId, "error", appErr.Error())
		return &model.SubmitDialogResponse{
			Error: "Failed to retrieve requester information",
		}
	}

	// Get approver info
	approver, appErr := p.API.GetUser(approverID)
	if appErr != nil {
		p.API.LogError("Failed to get approver user", "user_id", approverID, "error", appErr.Error())
		return &model.SubmitDialogResponse{
			Errors: map[string]string{
				"approver": "Failed to retrieve approver information",
			},
		}
	}

	// Create KV store for code uniqueness checking
	kvStore := store.NewKVStore(p.API)

	// Create approval record with unique code
	record, err := approval.NewApprovalRecord(
		kvStore,
		requester.Id, requester.Username, requester.GetDisplayName(model.ShowUsername),
		approver.Id, approver.Username, approver.GetDisplayName(model.ShowUsername),
		description,
		payload.ChannelId,
		payload.TeamId,
	)
	if err != nil {
		p.API.LogError("Failed to create approval record", "error", err.Error())
		return &model.SubmitDialogResponse{
			Error: "Failed to generate unique approval code",
		}
	}
	if err := kvStore.SaveApproval(record); err != nil {
		p.API.LogError("Failed to save approval record", "record_id", record.ID, "error", err.Error())
		return &model.SubmitDialogResponse{
			Error: "Failed to save approval request",
		}
	}

	// Send confirmation message to requester
	confirmMsg := fmt.Sprintf("✅ **Approval request created!**\n\n"+
		"**Reference Code:** `%s`\n"+
		"**Approver:** @%s\n"+
		"**Description:** %s\n\n"+
		"_Your approver will be notified shortly._",
		record.Code, approver.Username, description)

	post := &model.Post{
		UserId:    payload.UserId,
		ChannelId: payload.ChannelId,
		Message:   confirmMsg,
	}

	if _, appErr := p.API.CreatePost(post); appErr != nil {
		p.API.LogError("Failed to send confirmation post", "error", appErr.Error())
		// Don't fail the whole operation if we can't send the confirmation
	}

	p.API.LogInfo("Approval request created successfully",
		"record_id", record.ID,
		"code", record.Code,
		"requester", requester.Username,
		"approver", approver.Username)

	return &model.SubmitDialogResponse{}
}
