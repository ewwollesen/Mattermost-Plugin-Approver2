package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	// Button action endpoint (no auth required - authenticated via PostActionIntegrationRequest)
	router.HandleFunc("/action", p.handleAction).Methods(http.MethodPost)

	// Dialog submission endpoint (no auth required - handled by Mattermost)
	router.HandleFunc("/dialog/submit", p.handleDialogSubmit).Methods(http.MethodPost)

	// API routes with authentication middleware
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(p.MattermostAuthorizationRequired)
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

	// Route based on callback ID
	switch {
	case strings.HasPrefix(payload.CallbackId, "confirm_"):
		response = p.handleConfirmDecision(payload)
	case payload.CallbackId == "approve_new":
		response = p.handleApproveNew(payload)
	case strings.HasPrefix(payload.CallbackId, "cancel_approval_"):
		response = p.handleCancelModalSubmission(payload)
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
	err = kvStore.SaveApproval(record)
	if err != nil {
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
	postID, err := notifications.SendApprovalRequestDM(p.API, p.botUserID, record)
	if err != nil {
		// Story 2.6: Classify error and provide resolution suggestion (AC6)
		errorType, suggestion := notifications.ClassifyDMError(err)

		// Log warning but continue - approval record already saved (data integrity priority)
		p.API.LogWarn("DM notification failed but approval created",
			"approval_id", record.ID,
			"code", record.Code,
			"approver_id", record.ApproverID,
			"requester_id", record.RequesterID,
			"error", err.Error(),
			"error_type", errorType,
			"suggestion", suggestion,
		)
		// NotificationSent flag remains false (default value)
	} else {
		// Notification sent successfully - update flags and post ID (best effort)
		record.NotificationSent = true
		record.NotificationPostID = postID
		if err := kvStore.SaveApproval(record); err != nil {
			// Log warning but don't fail - notification already sent
			p.API.LogWarn("Failed to update notification tracking fields",
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

// handleAction processes button click actions from approval request notifications
func (p *Plugin) handleAction(w http.ResponseWriter, r *http.Request) {
	// Parse request body (Mattermost sends PostActionIntegrationRequest)
	var request model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.API.LogError("Failed to decode action request", "error", err.Error())
		p.writeActionError(w, "Invalid request")
		return
	}

	// Extract context data - Context is a map[string]any
	contextData := request.Context

	approvalID, ok := contextData["approval_id"].(string)
	if !ok || approvalID == "" {
		p.API.LogError("Missing or invalid approval_id in context")
		p.writeActionError(w, "Invalid request")
		return
	}

	action, ok := contextData["action"].(string)
	if !ok || (action != "approve" && action != "deny") {
		p.API.LogError("Missing or invalid action in context", "action", action)
		p.writeActionError(w, "Invalid request")
		return
	}

	// Get approval record from store
	record, err := p.store.GetApproval(approvalID)
	if err != nil {
		p.API.LogError("Failed to get approval record",
			"approval_id", approvalID,
			"error", err.Error(),
		)
		p.writeActionError(w, "Approval not found")
		return
	}

	// Verify authenticated user is the designated approver
	approverID := request.UserId
	if record.ApproverID != approverID {
		p.API.LogError("Unauthorized approval attempt",
			"approval_id", approvalID,
			"authenticated_user", approverID,
			"designated_approver", record.ApproverID,
		)
		p.writeActionError(w, "Permission denied")
		return
	}

	// Check status (immutability guard)
	if record.Status != approval.StatusPending {
		p.writeActionError(w, fmt.Sprintf("Decision already recorded: %s", record.Status))
		return
	}

	// Open confirmation modal
	if err := p.openConfirmationModal(request.TriggerId, record, action); err != nil {
		p.API.LogError("Failed to open confirmation modal",
			"approval_id", approvalID,
			"action", action,
			"error", err.Error(),
		)
		p.writeActionError(w, "Failed to open confirmation modal")
		return
	}

	// Return success response
	p.writeActionSuccess(w)
}

// openConfirmationModal opens an interactive dialog for approval/denial confirmation
func (p *Plugin) openConfirmationModal(triggerID string, record *approval.ApprovalRecord, action string) error {
	// Determine modal title and confirmation text
	title := "Confirm Approval"
	confirmText := "Confirm you are approving:"
	if action == "deny" {
		title = "Confirm Denial"
		confirmText = "Confirm you are denying:"
	}

	// Format description (quoted)
	quotedDescription := fmt.Sprintf("> %s", record.Description)

	// Construct modal introduction text
	introText := fmt.Sprintf(
		"%s\n\n%s\n\n**From:** @%s (%s)\n**Request ID:** `%s`\n\n**This decision will be recorded and cannot be edited.**",
		confirmText,
		quotedDescription,
		record.RequesterUsername,
		record.RequesterDisplayName,
		record.Code,
	)

	// Create interactive dialog
	dialog := model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       "/plugins/com.mattermost.plugin-approver2/dialog/submit",
		Dialog: model.Dialog{
			CallbackId:       fmt.Sprintf("confirm_%s_%s", action, record.ID),
			Title:            title,
			IntroductionText: introText,
			Elements: []model.DialogElement{
				{
					DisplayName: "Add comment (optional)",
					Name:        "comment",
					Type:        "textarea",
					Placeholder: "Optional: Provide context or reasoning for your decision",
					MaxLength:   500,
					Optional:    true,
				},
			},
			SubmitLabel: "Confirm",
		},
	}

	if appErr := p.API.OpenInteractiveDialog(dialog); appErr != nil {
		return fmt.Errorf("failed to open dialog: %w", appErr)
	}

	return nil
}

// writeActionSuccess writes a successful PostActionIntegrationResponse
func (p *Plugin) writeActionSuccess(w http.ResponseWriter) {
	response := &model.PostActionIntegrationResponse{}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// writeActionError writes an error PostActionIntegrationResponse
func (p *Plugin) writeActionError(w http.ResponseWriter, message string) {
	response := &model.PostActionIntegrationResponse{
		EphemeralText: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(response)
}

// handleConfirmDecision processes confirmation modal submissions (approve/deny decisions)
func (p *Plugin) handleConfirmDecision(payload *model.SubmitDialogRequest) *model.SubmitDialogResponse {
	// Parse callback ID: "confirm_approve_recordID" or "confirm_deny_recordID"
	parts := strings.Split(payload.CallbackId, "_")
	if len(parts) != 3 {
		p.API.LogError("Invalid callback ID format", "callback_id", payload.CallbackId)
		return &model.SubmitDialogResponse{
			Error: "Invalid request format",
		}
	}

	action := parts[1]     // "approve" or "deny"
	approvalID := parts[2] // record ID

	// Validate action
	if action != "approve" && action != "deny" {
		p.API.LogError("Invalid action in callback ID", "action", action, "callback_id", payload.CallbackId)
		return &model.SubmitDialogResponse{
			Error: "Invalid action",
		}
	}

	// Extract comment (optional)
	comment := ""
	if commentVal, ok := payload.Submission["comment"]; ok {
		if commentStr, ok := commentVal.(string); ok {
			comment = strings.TrimSpace(commentStr)
		}
	}

	// Verify approver identity
	approverID := payload.UserId
	record, err := p.store.GetApproval(approvalID)
	if err != nil {
		p.API.LogError("Failed to get approval record in confirm dialog",
			"approval_id", approvalID,
			"error", err.Error(),
		)
		return &model.SubmitDialogResponse{
			Error: "Approval not found",
		}
	}

	if record.ApproverID != approverID {
		p.API.LogError("Unauthorized decision attempt",
			"approval_id", approvalID,
			"authenticated_user", approverID,
			"designated_approver", record.ApproverID,
		)
		return &model.SubmitDialogResponse{
			Error: "Permission denied",
		}
	}

	// Re-check status (race condition guard)
	if record.Status != approval.StatusPending {
		p.API.LogWarn("Cannot record decision for non-pending approval",
			"approval_id", approvalID,
			"current_status", record.Status,
		)
		return &model.SubmitDialogResponse{
			Error: fmt.Sprintf("Decision already recorded: %s", record.Status),
		}
	}

	// Record decision (Story 2.4 integration point)
	decision := "approved"
	if action == "deny" {
		decision = "denied"
	}

	updatedRecord, err := p.service.RecordDecision(approvalID, approverID, decision, comment)
	if err != nil {
		p.API.LogError("Failed to record decision",
			"approval_id", approvalID,
			"action", action,
			"error", err.Error(),
		)
		return &model.SubmitDialogResponse{
			Error: "Failed to record decision. Please try again.",
		}
	}

	// BEST EFFORT: Send outcome notification to requester (Story 2.5, graceful degradation)
	postID, notifErr := notifications.SendOutcomeNotificationDM(p.API, p.botUserID, updatedRecord)
	if notifErr != nil {
		// Story 2.6: Classify error and provide resolution suggestion (AC6)
		errorType, suggestion := notifications.ClassifyDMError(notifErr)

		// Log warning but DO NOT return error (decision is already recorded successfully)
		p.API.LogWarn("Failed to send outcome notification",
			"approval_id", approvalID,
			"requester_id", updatedRecord.RequesterID,
			"error", notifErr.Error(),
			"error_type", errorType,
			"suggestion", suggestion,
		)
	} else {
		// Success - update OutcomeNotified flag (also best effort)
		p.API.LogInfo("Outcome notification sent",
			"approval_id", approvalID,
			"code", updatedRecord.Code,
			"decision", decision,
			"requester_id", updatedRecord.RequesterID,
			"post_id", postID,
		)
		updatedRecord.OutcomeNotified = true
		if flagErr := p.store.SaveApproval(updatedRecord); flagErr != nil {
			p.API.LogError("Failed to update OutcomeNotified flag",
				"approval_id", approvalID,
				"error", flagErr.Error(),
			)
			// Continue anyway - notification was sent successfully
		}
	}

	// Disable buttons in original DM notification (best effort)
	if err := p.disableButtonsInDM(record, decision); err != nil {
		// Log warning but continue - decision already recorded
		p.API.LogWarn("Failed to disable buttons in DM notification",
			"approval_id", approvalID,
			"decision", decision,
			"error", err.Error(),
		)
	}

	p.API.LogInfo("Approval decision recorded",
		"approval_id", approvalID,
		"code", record.Code,
		"decision", decision,
		"approver_id", approverID,
	)

	// Return success - modal will close
	return &model.SubmitDialogResponse{}
}

// disableButtonsInDM disables the action buttons in the original DM notification
func (p *Plugin) disableButtonsInDM(record *approval.ApprovalRecord, decision string) error {
	// Check if we have the notification post ID
	if record.NotificationPostID == "" {
		// Fallback: send new message if post ID not available
		return p.sendDecisionConfirmationFallback(record, decision)
	}

	// Get the original post
	post, appErr := p.API.GetPost(record.NotificationPostID)
	if appErr != nil {
		return fmt.Errorf("failed to get original post: %w", appErr)
	}

	// Create disabled buttons with updated message
	statusEmoji := "✅"
	statusText := "Approved"
	if decision == "denied" {
		statusEmoji = "❌"
		statusText = "Denied"
	}

	// Update the message to show decision recorded
	post.Message = fmt.Sprintf("%s **Decision Recorded: %s**\n\n%s", statusEmoji, statusText, post.Message)

	// Remove action buttons by clearing all Props (Story 4.7: same approach as cancellation)
	// WHY clear all Props instead of selective removal:
	// - Avoids complex nested map manipulation (reduced from 9 lines to 3)
	// - More maintainable and less error-prone than traversing attachments[0].actions
	// - Consistent with cancellation workflow (server/notifications/dm.go:220)
	// - No side effects: we only put action buttons in Props, so clearing all is safe
	post.Props = model.StringInterface{}

	// Update the post
	if _, appErr := p.API.UpdatePost(post); appErr != nil {
		return fmt.Errorf("failed to update post: %w", appErr)
	}

	return nil
}

// sendDecisionConfirmationFallback sends a new DM when the original post cannot be updated
func (p *Plugin) sendDecisionConfirmationFallback(record *approval.ApprovalRecord, decision string) error {
	// Get DM channel
	dmChannelID, err := notifications.GetDMChannelID(p.API, p.botUserID, record.ApproverID)
	if err != nil {
		return fmt.Errorf("failed to get DM channel: %w", err)
	}

	// Create informational message
	statusEmoji := "✅"
	statusText := "Approved"
	if decision == "denied" {
		statusEmoji = "❌"
		statusText = "Denied"
	}

	message := fmt.Sprintf(
		"%s **Decision Recorded: %s**\n\nApproval request `%s` has been decided.\n\nThe requester will be notified of your decision.",
		statusEmoji,
		statusText,
		record.Code,
	)

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: dmChannelID,
		Message:   message,
	}

	if _, appErr := p.API.CreatePost(post); appErr != nil {
		return fmt.Errorf("failed to create post: %w", appErr)
	}

	return nil
}

// handleCancelModalSubmission processes cancellation modal submissions
// Story 4.3: Handles user-selected cancellation reason and performs cancellation
func (p *Plugin) handleCancelModalSubmission(payload *model.SubmitDialogRequest) *model.SubmitDialogResponse {
	// Parse callback ID: "cancel_approval_{approvalID}"
	parts := strings.Split(payload.CallbackId, "_")
	if len(parts) != 3 {
		p.API.LogError("Invalid callback ID format", "callback_id", payload.CallbackId)
		return &model.SubmitDialogResponse{
			Error: "Invalid request format",
		}
	}

	approvalID := parts[2]

	// Extract form data
	reasonCode, ok := payload.Submission["reason_code"].(string)
	if !ok || reasonCode == "" {
		return &model.SubmitDialogResponse{
			Errors: map[string]string{
				"reason_code": "Reason is required",
			},
		}
	}

	// Extract additional details (optional, always captured - Story 7.3)
	additionalDetails := ""
	if detailsVal, ok := payload.Submission["additional_details"]; ok {
		if detailsStr, ok := detailsVal.(string); ok {
			additionalDetails = strings.TrimSpace(detailsStr)
		}
	}

	// Validate "Other" reason requires additional details
	if reasonCode == "other" && additionalDetails == "" {
		return &model.SubmitDialogResponse{
			Errors: map[string]string{
				"additional_details": "Please provide details when selecting 'Other reason'",
			},
		}
	}

	// Map reason code to human-readable text (Story 7.3: no longer passes details)
	reasonText := p.mapCancellationReason(reasonCode)

	// Get approval record
	record, err := p.store.GetApproval(approvalID)
	if err != nil {
		p.API.LogError("Failed to get approval record in cancel modal",
			"approval_id", approvalID,
			"error", err.Error(),
		)
		return &model.SubmitDialogResponse{
			Error: "Approval request not found",
		}
	}

	// Verify user is the requester (permission check - AC5)
	if record.RequesterID != payload.UserId {
		p.API.LogError("Unauthorized cancellation attempt",
			"approval_id", approvalID,
			"authenticated_user", payload.UserId,
			"requester_id", record.RequesterID,
		)
		return &model.SubmitDialogResponse{
			Error: "Only the requester can cancel this approval",
		}
	}

	// Get requester user for post update
	requester, appErr := p.API.GetUser(payload.UserId)
	if appErr != nil {
		p.API.LogError("Failed to get requester user",
			"error", appErr.Error(),
			"user_id", payload.UserId,
		)
		return &model.SubmitDialogResponse{
			Error: "Failed to retrieve user information. Please try again.",
		}
	}

	// Call service to cancel approval (Story 7.3: pass details separately)
	err = p.service.CancelApproval(record.Code, payload.UserId, reasonText, additionalDetails)
	if err != nil {
		p.API.LogError("Failed to cancel approval",
			"error", err.Error(),
			"approval_code", record.Code,
			"user_id", payload.UserId,
		)
		return &model.SubmitDialogResponse{
			Error: p.formatCancelError(err, record.Code),
		}
	}

	// Post-cancellation actions (Stories 4.1, 4.2) - best effort
	updatedRecord, err := p.store.GetByCode(record.Code)
	if err != nil {
		p.API.LogWarn("Failed to retrieve updated record for post update",
			"error", err.Error(),
			"approval_code", record.Code,
		)
	} else {
		// Update the original post (Story 4.1)
		err = notifications.UpdateApprovalPostForCancellation(p.API, updatedRecord, requester.Username)
		if err != nil {
			p.API.LogWarn("Failed to update approver post",
				"error", err.Error(),
				"approval_code", record.Code,
				"approver_post_id", updatedRecord.NotificationPostID,
			)
		}

		// Send cancellation notification DM to approver (Story 4.2)
		_, err = notifications.SendCancellationNotificationDM(p.API, p.botUserID, updatedRecord, requester.Username)
		if err != nil {
			errorType, suggestion := notifications.ClassifyDMError(err)
			p.API.LogWarn("Failed to send cancellation notification to approver",
				"error", err.Error(),
				"error_type", errorType,
				"suggestion", suggestion,
				"approval_code", updatedRecord.Code,
				"approver_id", updatedRecord.ApproverID,
			)
			// Continue - cancellation already recorded, notification is best-effort
		}

		// Send cancellation notification DM to requestor (Story 7.1)
		_, err = notifications.SendRequesterCancellationNotificationDM(p.API, p.botUserID, updatedRecord)
		if err != nil {
			errorType, suggestion := notifications.ClassifyDMError(err)
			p.API.LogWarn("Failed to send cancellation notification to requestor",
				"error", err.Error(),
				"error_type", errorType,
				"suggestion", suggestion,
				"approval_code", updatedRecord.Code,
				"requester_id", updatedRecord.RequesterID,
			)
			// Continue - cancellation already recorded, notification is best-effort
		}
	}

	p.API.LogInfo("Approval canceled successfully via modal",
		"approval_code", record.Code,
		"user_id", payload.UserId,
		"reason", reasonText,
	)

	// Success - return empty response (modal will close)
	return &model.SubmitDialogResponse{}
}

// mapCancellationReason maps reason codes to human-readable text
// Story 7.3: Details stored separately in CanceledDetails field
func (p *Plugin) mapCancellationReason(code string) string {
	switch code {
	case "no_longer_needed":
		return "No longer needed"
	case "wrong_approver":
		return "Wrong approver"
	case "sensitive_info":
		return "Sensitive information"
	case "other":
		return "Other" // Details stored separately in CanceledDetails
	default:
		return "Unknown reason"
	}
}
