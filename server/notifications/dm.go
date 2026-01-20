package notifications

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// SendApprovalRequestDM sends a DM notification to the approver when a new approval request is created.
// The message includes complete context: requester info, timestamp, description, and request ID.
// Returns the post ID and error. Error returned if DM send fails (caller should log and handle gracefully).
//
// Story 10.3: Uses Matterpoll pattern with CreateInteractiveApprovalPost() for interactive buttons.
// Button URLs use /api/v1/approval/{code}/approve|deny pattern (Story 10.2 handlers).
func SendApprovalRequestDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}

	// Get or create DM channel between bot and approver
	channelID, err := GetDMChannelID(api, botUserID, record.ApproverID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for approver %s: %w", record.ApproverID, err)
	}

	// Story 10.3: Use Matterpoll pattern helper (Story 10.1 infrastructure)
	// This uses model.ParseSlackAttachment() to preserve Integration URLs with custom post types
	post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeApprovalRequest)
	if post == nil {
		return "", fmt.Errorf("failed to create interactive approval post")
	}

	// Story 8.4: Add playbook context to markdown fallback if this is a playbook-linked approval
	// The FormatMarkdownFallback doesn't include playbook context, so we append it here
	playbookContext := formatPlaybookContext(api, record)
	if playbookContext != "" {
		post.Message += playbookContext
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send DM to approver %s: %w", record.ApproverID, appErr)
	}

	return createdPost.Id, nil
}

// SendOutcomeNotificationDM sends a DM notification to the requester when their approval request is decided.
// The message includes complete context: approver info, decision time, original request, decision comment, and status.
//
// Story 10.5: Uses Matterpoll pattern with CreateInteractiveApprovalPost() for custom webapp component rendering.
// The webapp ApprovalDMPost component handles "outcome" notification type to display timestamps in user's timezone.
//
// IMPORTANT: This function implements graceful degradation (Architecture Decision 2.2). The caller MUST NOT
// fail the approval decision recording if this notification fails. Decision integrity is non-negotiable.
//
// Returns the post ID on success, or error if DM send fails (e.g., DM channel creation failure, CreatePost failure).
// The caller should log errors at WARN level and continue - notification failures are best-effort only.
func SendOutcomeNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}

	// Validate status - outcome notifications require approved or denied status
	if record.Status != approval.StatusApproved && record.Status != approval.StatusDenied {
		return "", fmt.Errorf("invalid status for outcome notification: %s", record.Status)
	}

	// Get or create DM channel between bot and requester
	channelID, err := GetDMChannelID(api, botUserID, record.RequesterID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for requester %s: %w", record.RequesterID, err)
	}

	// Story 10.5: Use Matterpoll pattern helper (Story 10.1 infrastructure)
	// Creates custom_approval_dm post with outcome notification type
	// No buttons for outcome notifications (handled by CreateInteractiveApprovalPost)
	// Markdown fallback in post.Message for non-webapp clients
	post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeOutcome)
	if post == nil {
		return "", fmt.Errorf("failed to create interactive approval post")
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send outcome DM to requester %s: %w", record.RequesterID, appErr)
	}

	return createdPost.Id, nil
}

// UpdateApprovalPostForCancellation updates the approver's DM post to show canceled state.
// This function:
// - Updates the message to show cancellation with plain description text
// - Removes interactive buttons (fixes ghost buttons bug)
// - Shows who canceled and when
//
// Returns error if post update fails. Caller should log but continue with cancellation.
func UpdateApprovalPostForCancellation(api plugin.API, record *approval.ApprovalRecord, canceledByUsername string) error {
	// Validate inputs
	if record == nil {
		return fmt.Errorf("approval record is nil")
	}
	if record.NotificationPostID == "" {
		api.LogWarn("Cannot update approver post: no post ID stored", "request_id", record.ID)
		return fmt.Errorf("no approver post ID found")
	}

	// Get the original post
	post, appErr := api.GetPost(record.NotificationPostID)
	if appErr != nil {
		api.LogError("Failed to get post for update", "post_id", record.NotificationPostID, "error", appErr.Error())
		return fmt.Errorf("failed to get post: %w", appErr)
	}

	// Build updated message with cancellation info
	canceledAt := time.UnixMilli(record.CanceledAt).UTC()
	canceledAtStr := canceledAt.Format("Jan 02, 2006 3:04 PM")

	updatedMessage := fmt.Sprintf("🚫 **Approval Request (Canceled)**\n\n"+
		"**From:** @%s\n"+
		"**Request ID:** `%s`\n"+
		"**Description:**\n%s\n\n"+
		"---\n"+
		"_Canceled by @%s at %s_",
		record.RequesterUsername,
		record.Code,
		record.Description,
		canceledByUsername,
		canceledAtStr,
	)

	// Story 9.10: Clear interactive buttons from approval request post
	// Note: Approval request DMs are standard posts (not custom_approval) to preserve button functionality
	// Simply clear props to remove the buttons
	post.Message = updatedMessage
	post.Props = model.StringInterface{} // Clear all interactive elements

	// Update the post
	_, appErr = api.UpdatePost(post)
	if appErr != nil {
		api.LogError("Failed to update post", "post_id", record.NotificationPostID, "error", appErr.Error())
		return fmt.Errorf("failed to update post: %w", appErr)
	}

	return nil
}

// SendCancellationNotificationDM sends a DM notification to the approver when a request is canceled.
// The message includes complete context: reference code, requester, cancellation reason, and timestamp.
//
// Story 10.6: Uses Matterpoll pattern with CreateInteractiveApprovalPost() for custom webapp component rendering.
// The webapp ApprovalDMPost component handles "cancellation" notification type to display timestamps in user's timezone.
// No interactive buttons are rendered for cancellation notifications (read-only).
//
// Note: The canceledByUsername parameter is preserved for API compatibility but not used -
// the cancellation message shows requester info from the record.
//
// IMPORTANT: This function implements graceful degradation (Architecture Decision 2.2). The caller MUST NOT
// fail the cancellation operation if this notification fails. Cancellation integrity is non-negotiable.
//
// Returns the post ID on success, or error if DM send fails (e.g., DM channel creation failure, CreatePost failure).
// The caller should log errors at WARN level and continue - notification failures are best-effort only.
func SendCancellationNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord, canceledByUsername string) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}
	if record.ApproverID == "" {
		return "", fmt.Errorf("approver ID is empty")
	}

	// Get or create DM channel between bot and APPROVER (not requester)
	channelID, err := GetDMChannelID(api, botUserID, record.ApproverID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for approver %s: %w", record.ApproverID, err)
	}

	// Story 10.6: Use Matterpoll pattern helper (Story 10.1 infrastructure)
	// Creates custom_approval_dm post with cancellation notification type
	// No buttons for cancellation notifications (handled by CreateInteractiveApprovalPost)
	// Props include canceled_at, canceled_reason (FormatApprovalPropsForDM lines 176-181)
	// Markdown fallback in post.Message for non-webapp clients
	post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeCancellation)
	if post == nil {
		return "", fmt.Errorf("failed to create interactive approval post")
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send cancellation notification to approver %s: %w", record.ApproverID, appErr)
	}

	return createdPost.Id, nil
}

// SendTimeoutNotificationDM sends a DM notification to the requester when their approval request times out.
// The message includes complete context: request details, approver info, timeout reason, and actionable guidance.
//
// Story 10.7: Uses Matterpoll pattern with CreateInteractiveApprovalPost() for custom webapp component rendering.
// The webapp ApprovalDMPost component handles "timeout" notification type to display timestamps in user's timezone.
// No interactive buttons are rendered for timeout notifications (read-only).
//
// IMPORTANT: This function implements graceful degradation (Architecture Decision 2.2). The caller MUST NOT
// fail the auto-cancellation operation if this notification fails. Data integrity is non-negotiable.
//
// Returns the post ID on success, or error if DM send fails (e.g., DM channel creation failure, CreatePost failure).
// The caller should log errors at WARN level and continue - notification failures are best-effort only.
func SendTimeoutNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}
	if record.RequesterID == "" {
		return "", fmt.Errorf("requester ID is empty")
	}

	// Get or create DM channel between bot and requester
	channelID, err := GetDMChannelID(api, botUserID, record.RequesterID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for requester %s: %w", record.RequesterID, err)
	}

	// Story 10.7: Use Matterpoll pattern helper (Story 10.1 infrastructure)
	// Creates custom_approval_dm post with timeout notification type
	// No buttons for timeout notifications (handled by CreateInteractiveApprovalPost)
	// Props include created_at for timeout display (FormatApprovalPropsForDM)
	// Markdown fallback in post.Message for non-webapp clients
	post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeTimeout)
	if post == nil {
		return "", fmt.Errorf("failed to create interactive approval post")
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send timeout notification to requester %s: %w", record.RequesterID, appErr)
	}

	return createdPost.Id, nil
}

// SendRequesterCancellationNotificationDM sends a DM notification to the requestor when their approval request is canceled by an approver.
// Epic 7: 1.0 Polish & UX Improvements, Story 7.1: Completes the feedback loop by notifying requestors of cancellation.
//
// Story 10.6: Uses Matterpoll pattern with CreateInteractiveApprovalPost() for custom webapp component rendering.
// The webapp ApprovalDMPost component handles "requester_cancellation" notification type to display:
// - Who canceled the request (approver info)
// - Cancellation reason and timestamp in user's timezone
// No interactive buttons are rendered for cancellation notifications (read-only).
//
// IMPORTANT: This function implements graceful degradation (Architecture Decision 2.2). The caller MUST NOT
// fail the cancellation operation if this notification fails - it is best-effort only. The cancellation has
// already been recorded in the KV store before this notification is sent.
//
// Returns the post ID on success, or error if DM send fails (e.g., DM channel creation failure, CreatePost failure).
// The caller should log errors at WARN level with ClassifyDMError() and continue - notification failures are best-effort only.
func SendRequesterCancellationNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}
	if record.RequesterID == "" {
		return "", fmt.Errorf("requester ID is empty")
	}

	// Get or create DM channel with REQUESTER (not approver)
	channelID, err := GetDMChannelID(api, botUserID, record.RequesterID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel with requestor %s: %w", record.RequesterID, err)
	}

	// Story 10.6: Use Matterpoll pattern helper (Story 10.1 infrastructure)
	// Creates custom_approval_dm post with requester_cancellation notification type
	// No buttons for cancellation notifications (handled by CreateInteractiveApprovalPost)
	// Props include canceled_at, canceled_reason (FormatApprovalPropsForDM)
	// Markdown fallback in post.Message for non-webapp clients
	post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeRequesterCancellation)
	if post == nil {
		return "", fmt.Errorf("failed to create interactive approval post")
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send cancellation notification to requestor %s: %w", record.RequesterID, appErr)
	}

	return createdPost.Id, nil
}

// SendVerificationNotificationDM sends a DM notification to the approver when the requester marks an approved request as verified.
// Story 6.2: Notifies approver that the requester has confirmed completion of the approved action.
//
// Story 10.8: Uses Matterpoll pattern with CreateInteractiveApprovalPost() for custom webapp component rendering.
// The webapp ApprovalDMPost component handles "verification" notification type to display timestamps in user's timezone.
// No interactive buttons are rendered for verification notifications (read-only).
//
// IMPORTANT: This function implements graceful degradation (Architecture Decision 2.2). The caller MUST NOT
// fail the verification operation if this notification fails - it is best-effort only. The notification is
// informational and does not affect the approval workflow.
//
// Returns the post ID on success, or error if DM send fails (e.g., DM channel creation failure, CreatePost failure).
// The caller should log errors at WARN level and continue - notification failures are best-effort only.
func SendVerificationNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}
	if record.ApproverID == "" {
		return "", fmt.Errorf("approver ID is empty")
	}

	// Get or create DM channel between bot and APPROVER
	channelID, err := GetDMChannelID(api, botUserID, record.ApproverID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for approver %s: %w", record.ApproverID, err)
	}

	// Story 10.8: Use Matterpoll pattern helper (Story 10.1 infrastructure)
	// Creates custom_approval_dm post with verification notification type
	// No buttons for verification notifications (handled by CreateInteractiveApprovalPost)
	// Props include verified_at, verification_comment (FormatApprovalPropsForDM lines 186-191)
	// Markdown fallback in post.Message for non-webapp clients
	post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeVerification)
	if post == nil {
		return "", fmt.Errorf("failed to create interactive approval post")
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send verification notification to approver %s: %w", record.ApproverID, appErr)
	}

	return createdPost.Id, nil
}

// // FormatApprovalPropsForDM formats approval record data for DM notification custom post type
// // Returns map suitable for post.Props that matches webapp ApprovalPost component expectations
// // Story 9.10: DM notification conversion to custom post type
// //
// // Field names use snake_case to match webapp expectations (Story 9.7 AC3):
// //   - approval_code, approval_status, requester_username, etc.
// //
// // Timestamps are int64 Unix milliseconds (NOT formatted strings)
// // All fields are required except: decided_at, decision_comment (optional for pending status)
// // DM-specific fields: notification_type, is_dm
// func FormatApprovalPropsForDM(record *approval.ApprovalRecord, notificationType string) map[string]any {
// 	if record == nil {
// 		return make(map[string]any)
// 	}
//
// 	props := map[string]any{
// 		// Standard approval fields (same as playbook posts)
// 		"approval_code":          record.Code,
// 		"approval_status":        record.Status,
// 		"requester_username":     record.RequesterUsername,
// 		"requester_display_name": record.RequesterDisplayName,
// 		"approver_username":      record.ApproverUsername,
// 		"approver_display_name":  record.ApproverDisplayName,
// 		"description":            record.Description,
// 		"created_at":             record.CreatedAt, // int64 Unix millis
//
// 		// DM-specific fields
// 		"notification_type": notificationType,
// 		"is_dm":             true,
// 	}
//
// 	// Optional fields - only include if they have meaningful values
// 	if record.DecidedAt > 0 {
// 		props["decided_at"] = record.DecidedAt // int64 Unix millis
// 	}
// 	if record.DecisionComment != "" {
// 		props["decision_comment"] = record.DecisionComment
// 	}
//
// 	// Playbook context (if available) - use same field names as playbooks/formatters.go
// 	if record.PlaybookRunID != "" {
// 		props["playbook_id"] = record.PlaybookRunID
// 	}
// 	if record.PlaybookName != "" {
// 		props["playbook_title"] = record.PlaybookName
// 	}
// 	if record.PlaybookChannelID != "" {
// 		props["playbook_channel_id"] = record.PlaybookChannelID
// 	}
//
// 	return props
// }

// formatPlaybookContext formats the playbook context section for DM notifications
// Returns formatted string with playbook name and channel link (Story 8.4: AC2, AC3, AC8)
// Returns empty string if no playbook context is available
func formatPlaybookContext(api plugin.API, record *approval.ApprovalRecord) string {
	// Only format if this is a playbook-linked approval
	if record.PlaybookRunID == "" || record.PlaybookName == "" {
		return ""
	}

	// Truncate playbook name if > 50 characters (AC8)
	// Use rune count for proper UTF-8 handling (emojis, CJK characters, etc.)
	playbookName := record.PlaybookName
	if utf8.RuneCountInString(playbookName) > 50 {
		runes := []rune(playbookName)
		playbookName = string(runes[:47]) + "..."
	}

	// Get channel name for clickable link (requires channel name, not ID)
	var channelLink string
	if record.PlaybookChannelID != "" {
		channel, err := api.GetChannel(record.PlaybookChannelID)
		if err != nil || channel == nil || channel.Name == "" {
			// Fallback to ID if we can't fetch channel name or name is empty
			api.LogDebug("Failed to get playbook channel for link",
				"channel_id", record.PlaybookChannelID,
				"error", err)
			channelLink = fmt.Sprintf("~%s", record.PlaybookChannelID)
		} else {
			// Use channel name for proper clickable link
			channelLink = fmt.Sprintf("~%s", channel.Name)
		}
	} else {
		channelLink = "~unknown"
	}

	return fmt.Sprintf("\n**Playbook Context:**\n"+
		"- Playbook: %s\n"+
		"- Channel: %s\n",
		playbookName,
		channelLink)
}

// GetDMChannelID gets or creates a DM channel between the bot and the target user.
// Returns the channel ID if successful, or an error if the channel cannot be created.
func GetDMChannelID(api plugin.API, botUserID, targetUserID string) (string, error) {
	// Get or create DM channel (creates if doesn't exist)
	channel, appErr := api.GetDirectChannel(botUserID, targetUserID)
	if appErr != nil {
		return "", fmt.Errorf("failed to get DM channel for user %s: %w", targetUserID, appErr)
	}

	if channel == nil {
		return "", fmt.Errorf("DM channel is nil for user %s", targetUserID)
	}

	return channel.Id, nil
}
