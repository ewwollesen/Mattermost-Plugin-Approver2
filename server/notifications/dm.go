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

	// Format timestamp as YYYY-MM-DD HH:MM:SS UTC (AC2 requirement)
	timestamp := time.UnixMilli(record.CreatedAt).UTC()
	timestampStr := timestamp.Format("2006-01-02 15:04:05 MST")

	// Construct DM message with exact format from AC2
	message := fmt.Sprintf("📋 **Approval Request**\n\n"+
		"**From:** @%s (%s)\n"+
		"**Requested:** %s\n"+
		"**Description:**\n%s\n\n"+
		"**Request ID:** `%s`",
		record.RequesterUsername,
		record.RequesterDisplayName,
		timestampStr,
		record.Description,
		record.Code)

	// Story 8.4: Add playbook context if this is a playbook-linked approval (AC1, AC5)
	playbookContext := formatPlaybookContext(api, record)
	if playbookContext != "" {
		message += playbookContext
	}

	// Standard post with interactive buttons (baseline - working)
	// Note: Timestamps will show in UTC (not user timezone) for DM notifications
	// Future: Story needed to implement timezone support for DM buttons
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Message:   message, // Markdown message
		Props: model.StringInterface{
			"attachments": []any{
				map[string]any{
					"actions": []any{
						// Approve button (green/primary style)
						map[string]any{
							"name": "Approve",
							"type": "button",
							"integration": map[string]any{
								"url": "/plugins/com.mattermost.plugin-approver2/action",
								"context": map[string]any{
									"approval_id": record.ID,
									"action":      "approve",
								},
							},
							"style": "primary",
						},
						// Deny button (red/danger style)
						map[string]any{
							"name": "Deny",
							"type": "button",
							"integration": map[string]any{
								"url": "/plugins/com.mattermost.plugin-approver2/action",
								"context": map[string]any{
									"approval_id": record.ID,
									"action":      "deny",
								},
							},
							"style": "danger",
						},
					},
				},
			},
		},
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

	// Get or create DM channel between bot and requester
	channelID, err := GetDMChannelID(api, botUserID, record.RequesterID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for requester %s: %w", record.RequesterID, err)
	}

	// Format decision timestamp as YYYY-MM-DD HH:MM:SS UTC
	timestamp := time.UnixMilli(record.DecidedAt).UTC()
	timestampStr := timestamp.Format("2006-01-02 15:04:05 MST")

	// Determine header and status based on decision
	var header, status string
	switch record.Status {
	case approval.StatusApproved:
		header = "✅ **Approval Request Approved**"
		status = "**Status:** You may proceed with this action."
	case approval.StatusDenied:
		header = "❌ **Approval Request Denied**"
		status = "**Status:** This request has been denied."
	default:
		return "", fmt.Errorf("invalid status for outcome notification: %s", record.Status)
	}

	// Construct base message
	message := fmt.Sprintf("%s\n\n"+
		"**Approver:** @%s (%s)\n"+
		"**Decision Time:** %s\n"+
		"**Request ID:** `%s`\n\n"+
		"**Original Request:**\n> %s",
		header,
		record.ApproverUsername,
		record.ApproverDisplayName,
		timestampStr,
		record.Code,
		record.Description)

	// Add comment section if decision comment is present
	if record.DecisionComment != "" {
		message += fmt.Sprintf("\n\n**Comment:**\n%s", record.DecisionComment)
	}

	// Add status statement
	message += fmt.Sprintf("\n\n%s", status)

	// Create standard post
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Message:   message,
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

	// Get or create DM channel between bot and approver
	channelID, err := GetDMChannelID(api, botUserID, record.ApproverID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for approver %s: %w", record.ApproverID, err)
	}

	// Format cancellation timestamp as "Jan 02, 2006 3:04 PM"
	canceledAt := time.UnixMilli(record.CanceledAt).UTC()
	canceledAtStr := canceledAt.Format("Jan 02, 2006 3:04 PM")

	// Handle cancellation reason (may be empty)
	canceledReason := record.CanceledReason
	if canceledReason == "" {
		canceledReason = "Not specified"
	}

	// Construct DM message (Story 7.3: include details if present)
	message := fmt.Sprintf("🚫 **Approval Request Canceled**\n\n"+
		"**Reference:** `%s`\n"+
		"**Requester:** @%s\n"+
		"**Reason:** %s",
		record.Code,
		record.RequesterUsername,
		canceledReason,
	)

	// Add details if present (Story 7.3)
	if record.CanceledDetails != "" {
		message += fmt.Sprintf("\n**Details:** %s", record.CanceledDetails)
	}

	message += fmt.Sprintf("\n**Canceled:** %s\n\n"+
		"The approval request you received has been canceled by the requester.",
		canceledAtStr,
	)

	// Create standard post
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Message:   message,
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

	// Construct DM message per AC3 requirements
	message := fmt.Sprintf("⏱️ **Approval Request Timed Out**\n\n"+
		"**Request ID:** `%s`\n\n"+
		"**Original Request:**\n> %s\n\n"+
		"**Approver:** @%s (%s)\n\n"+
		"**Reason:** No response within 30 minutes\n\n"+
		"**Status:** This request has been automatically canceled. You may create a new request if still needed.",
		record.Code,
		record.Description,
		record.ApproverUsername,
		record.ApproverDisplayName)

	// Create standard post
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Message:   message,
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

	// Get or create DM channel with requestor
	channelID, err := GetDMChannelID(api, botUserID, record.RequesterID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel with requestor %s: %w", record.RequesterID, err)
	}

	// Format cancellation timestamp
	cancelTime := time.UnixMilli(record.CanceledAt).UTC().Format("Jan 02, 2006 3:04 PM")

	// Build notification message (requestor perspective, Story 7.3: include details if present)
	message := fmt.Sprintf(`🚫 **Your Approval Request Was Canceled**

**Request ID:** `+"`%s`"+`
**Original Request:** %s
**Approver:** @%s
**Reason:** %s`,
		record.Code,
		record.Description,
		record.ApproverUsername,
		record.CanceledReason,
	)

	// Add details if present (Story 7.3)
	if record.CanceledDetails != "" {
		message += fmt.Sprintf(`
**Details:** %s`, record.CanceledDetails)
	}

	message += fmt.Sprintf(`
**Canceled:** %s

---

The approver has canceled this approval request. You may submit a new request if needed.`,
		cancelTime,
	)

	// Create standard post
	post := &model.Post{
		ChannelId: channelID,
		UserId:    botUserID,
		Message:   message,
	}

	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send cancellation notification to requestor %s: %w", record.RequesterID, appErr)
	}

	return createdPost.Id, nil
}

// SendVerificationNotificationDM sends a DM notification to the approver when the requester marks an approved request as verified.
// Story 6.2: Notifies approver that the requester has confirmed completion of the approved action.
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

	// Get or create DM channel between bot and approver
	channelID, err := GetDMChannelID(api, botUserID, record.ApproverID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for approver %s: %w", record.ApproverID, err)
	}

	// Format verification timestamp
	timestamp := time.UnixMilli(record.VerifiedAt).UTC()
	timestampStr := timestamp.Format("2006-01-02 15:04:05 MST")

	// Construct DM message
	message := fmt.Sprintf("✅ **Approval Request Verified**\n\n"+
		"**Request ID:** `%s`\n\n"+
		"**Original Request:**\n> %s\n\n"+
		"**Requester:** @%s (%s)\n"+
		"**Verified:** %s",
		record.Code,
		record.Description,
		record.RequesterUsername,
		record.RequesterDisplayName,
		timestampStr)

	// Add verification comment if provided
	if record.VerificationComment != "" {
		message += fmt.Sprintf("\n\n**Verification Note:**\n> %s", record.VerificationComment)
	}

	// Create standard post
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Message:   message,
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
