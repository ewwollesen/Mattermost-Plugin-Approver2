package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/command"
	"github.com/mattermost/mattermost-plugin-approver2/server/notifications"
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

// handleApproveNew creates and saves a new approval request.
// Validation occurs in two layers:
// 1. Field presence validation (HandleDialogSubmission) - keeps modal open
// 2. Business logic validation (ValidateApprover, ValidateDescription) - keeps modal open
//
// All errors are logged at this highest layer per Mattermost conventions.
// Field-specific errors return to modal; general errors close modal.
func (p *Plugin) handleApproveNew(payload *model.SubmitDialogRequest) *model.SubmitDialogResponse {
	// Layer 1: Basic field presence validation
	response := command.HandleDialogSubmission(payload.Submission)
	if len(response.Errors) > 0 {
		return response
	}

	// Extract validated data with safe type assertions
	approverID, ok := payload.Submission["approver"].(string)
	if !ok {
		p.API.LogError("Invalid approver type in submission", "type", fmt.Sprintf("%T", payload.Submission["approver"]))
		return &model.SubmitDialogResponse{
			Error: "Invalid submission format. Please try again.",
		}
	}

	description, ok := payload.Submission["description"].(string)
	if !ok {
		p.API.LogError("Invalid description type in submission", "type", fmt.Sprintf("%T", payload.Submission["description"]))
		return &model.SubmitDialogResponse{
			Error: "Invalid submission format. Please try again.",
		}
	}

	// Layer 2: Business logic validation

	// Validate description length (AC4: Validate Description Length)
	if err := approval.ValidateDescription(description); err != nil {
		p.API.LogError("Description validation failed", "error", err.Error(), "description_length", len(description))
		return &model.SubmitDialogResponse{
			Errors: map[string]string{
				"description": err.Error(),
			},
		}
	}

	// Validate approver exists and is active (AC3: Validate Invalid Approver)
	// AC6: Handle Mattermost API Errors - error wrapping is done in ValidateApprover
	// Returns validated user object to avoid redundant API call
	approver, err := approval.ValidateApprover(approverID, p.API)
	if err != nil {
		p.API.LogError("Approver validation failed", "error", err.Error(), "approver_id", approverID)
		return &model.SubmitDialogResponse{
			Errors: map[string]string{
				"approver": err.Error(),
			},
		}
	}

	// Get requester info
	requester, appErr := p.API.GetUser(payload.UserId)
	if appErr != nil {
		p.API.LogError("Failed to get requester user", "user_id", payload.UserId, "error", appErr.Error())
		return &model.SubmitDialogResponse{
			Error: "Failed to retrieve requester information",
		}
	}

	// Create KV store for code uniqueness checking
	kvStore := store.NewKVStore(p.API)

	// Create approval record with unique code
	record, err := approval.NewApprovalRecord(
		kvStore,
		requester.Id, requester.Username, requester.GetDisplayName(model.ShowFullName),
		approver.Id, approver.Username, approver.GetDisplayName(model.ShowFullName),
		description,
		payload.ChannelId,
		payload.TeamId,
	)
	if err != nil {
		// Log with full context (Task 4: AC6)
		p.API.LogError("Failed to create approval record",
			"error", err.Error(),
			"requester_id", requester.Id,
			"approver_id", approver.Id,
		)
		// General error closes modal (system failure, not validation failure)
		return &model.SubmitDialogResponse{
			Error: "Failed to generate unique approval code. Please try again.",
		}
	}

	// Task 4 (AC5): Handle KV Store Unavailability with proper error wrapping
	if err := kvStore.SaveApproval(record); err != nil {
		// Log with full context at highest layer (Mattermost convention)
		p.API.LogError("Failed to save approval record to KV store",
			"error", err.Error(),
			"record_id", record.ID,
			"code", record.Code,
			"requester_id", requester.Id,
			"approver_id", approver.Id,
		)
		// AC5: User-friendly message for KV store failures
		return &model.SubmitDialogResponse{
			Error: "Failed to create approval request. The system is temporarily unavailable. Please try again.",
		}
	}

	// Story 2.1: Send DM notification to approver (best effort, graceful degradation)
	if err := notifications.SendApprovalRequestDM(p.API, p.botUserID, record); err != nil {
		// Log warning but continue - approval record already saved (data integrity priority)
		p.API.LogWarn("DM notification failed but approval created",
			"approval_id", record.ID,
			"code", record.Code,
			"approver_id", record.ApproverID,
			"requester_id", record.RequesterID,
			"error", err.Error(),
		)
		// NotificationSent flag remains false (default value)
	} else {
		// Notification sent successfully - update flag (best effort)
		record.NotificationSent = true
		if err := kvStore.SaveApproval(record); err != nil {
			// Log warning but don't fail - notification already sent
			p.API.LogWarn("Failed to update NotificationSent flag",
				"approval_id", record.ID,
				"code", record.Code,
				"error", err.Error(),
			)
		}
	}

	// Send ephemeral confirmation message to requester (visible only to them)
	confirmMsg := fmt.Sprintf("✅ **Approval Request Submitted**\n\n"+
		"**Approver:** @%s (%s)\n"+
		"**Request ID:** `%s`\n\n"+
		"You will be notified when a decision is made.",
		approver.Username, approver.GetDisplayName(model.ShowFullName),
		record.Code)

	post := &model.Post{
		UserId:    "", // Empty for system/bot message
		ChannelId: payload.ChannelId,
		Message:   confirmMsg,
	}

	// Send ephemeral post (only requester sees it)
	ephemeralPost := p.API.SendEphemeralPost(payload.UserId, post)
	if ephemeralPost == nil {
		p.API.LogError("Failed to send ephemeral confirmation post", "user_id", payload.UserId, "record_id", record.ID, "code", record.Code)
		// AC3: Fallback to regular post as generic success indicator if ephemeral fails
		// This ensures user sees confirmation even if ephemeral delivery fails
		post.UserId = payload.UserId
		if _, appErr := p.API.CreatePost(post); appErr != nil {
			p.API.LogError("Failed to send fallback confirmation post", "user_id", payload.UserId, "record_id", record.ID, "code", record.Code, "error", appErr.Error())
			// Don't fail the whole operation - record is already saved
		}
	}

	p.API.LogInfo("Approval request created successfully",
		"record_id", record.ID,
		"code", record.Code,
		"requester", requester.Username,
		"approver", approver.Username)

	return &model.SubmitDialogResponse{}
}
