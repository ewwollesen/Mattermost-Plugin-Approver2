package notifications

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
)

// PluginID is the unique identifier for this plugin used in Integration URLs
// Matches the ID in plugin.json/manifest.go
const PluginID = "com.mattermost.plugin-approver2"

// CustomApprovalDMPostType is the custom post type for DM approval notifications
// This enables webapp rendering with the ApprovalDMPost component
const CustomApprovalDMPostType = "custom_approval_dm"

// Notification type constants for the notification_type prop
const (
	NotificationTypeApprovalRequest       = "approval_request"
	NotificationTypeOutcome               = "outcome"
	NotificationTypeCancellation          = "cancellation"           // Sent to approver when requester cancels
	NotificationTypeRequesterCancellation = "requester_cancellation" // Sent to requester when approver cancels
	NotificationTypeTimeout               = "timeout"
	NotificationTypeVerification          = "verification"
)

// CreateApproveAction creates a PostAction for the Approve button.
// Uses URL path parameter pattern (Matterpoll style) for the approval code.
// Integration URL: /plugins/{PluginID}/api/v1/approval/{code}/approve
func CreateApproveAction(code string) *model.PostAction {
	return &model.PostAction{
		Id:    "approve",
		Name:  "Approve",
		Type:  model.PostActionTypeButton,
		Style: "success",
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("/plugins/%s/api/v1/approval/%s/approve", PluginID, code),
		},
	}
}

// CreateDenyAction creates a PostAction for the Deny button.
// Uses URL path parameter pattern (Matterpoll style) for the approval code.
// Integration URL: /plugins/{PluginID}/api/v1/approval/{code}/deny
func CreateDenyAction(code string) *model.PostAction {
	return &model.PostAction{
		Id:    "deny",
		Name:  "Deny",
		Type:  model.PostActionTypeButton,
		Style: "danger",
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("/plugins/%s/api/v1/approval/%s/deny", PluginID, code),
		},
	}
}

// CreateInteractiveApprovalPost creates a post with interactive buttons using the Matterpoll pattern.
// This is the CRITICAL function that uses model.ParseSlackAttachment() to preserve Integration URLs
// with custom post types.
//
// Parameters:
//   - botUserID: The bot user ID (for post.UserId)
//   - channelID: The target channel ID (for post.ChannelId)
//   - record: The approval record containing all data
//   - notificationType: One of the NotificationType* constants
//
// Returns the prepared post, or nil if inputs are invalid.
// Caller is responsible for creating via API.CreatePost().
func CreateInteractiveApprovalPost(botUserID, channelID string, record *approval.ApprovalRecord, notificationType string) *model.Post {
	// Validate required inputs (matches dm.go validation pattern)
	if botUserID == "" || channelID == "" || record == nil {
		return nil
	}

	// Build PostAction slice - only include buttons for pending approval requests
	var actions []*model.PostAction
	if notificationType == NotificationTypeApprovalRequest && record.Status == approval.StatusPending {
		actions = []*model.PostAction{
			CreateApproveAction(record.Code),
			CreateDenyAction(record.Code),
		}
	}

	// Determine title based on notification type
	title := getNotificationTitle(notificationType, record.Status)

	// Create SlackAttachment with title, text, and actions (AC3)
	attachment := &model.SlackAttachment{
		Title:   title,
		Text:    record.Description,
		Actions: actions,
	}

	// Create post with custom type (AC1)
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Type:      CustomApprovalDMPostType,
		Message:   FormatMarkdownFallback(record, notificationType), // AC5: Markdown fallback
	}

	// Populate post.Props with approval data (AC4)
	props := FormatApprovalPropsForDM(record, notificationType)
	for k, v := range props {
		post.AddProp(k, v)
	}

	// CRITICAL: Use ParseSlackAttachment to preserve Integration URLs (Matterpoll pattern)
	// This is what makes buttons work with custom post types!
	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})

	return post
}

// getNotificationTitle returns the appropriate title for the notification type and status
func getNotificationTitle(notificationType, status string) string {
	switch notificationType {
	case NotificationTypeApprovalRequest:
		return "Approval Request"
	case NotificationTypeOutcome:
		if status == approval.StatusApproved {
			return "Approval Request Approved"
		}
		return "Approval Request Denied"
	case NotificationTypeCancellation:
		return "Approval Request Canceled"
	case NotificationTypeRequesterCancellation:
		return "Your Approval Request Was Canceled"
	case NotificationTypeTimeout:
		return "Approval Request Timed Out"
	case NotificationTypeVerification:
		return "Approval Request Verified"
	default:
		return "Approval Notification"
	}
}

// FormatApprovalPropsForDM formats approval record data for DM notification custom post type.
// Returns map suitable for post.Props that matches webapp ApprovalDMPost component expectations.
//
// Field names use snake_case to match webapp expectations (Story 9.7 AC3):
//   - approval_code, approval_status, requester_username, etc.
//
// Timestamps are int64 Unix milliseconds (NOT formatted strings).
// All fields are required except: decided_at, decision_comment (optional for pending status).
// DM-specific fields: notification_type, is_dm.
func FormatApprovalPropsForDM(record *approval.ApprovalRecord, notificationType string) map[string]any {
	if record == nil {
		return make(map[string]any)
	}

	props := map[string]any{
		// Standard approval fields (same as playbook posts)
		"approval_code":          record.Code,
		"approval_status":        record.Status,
		"requester_username":     record.RequesterUsername,
		"requester_display_name": record.RequesterDisplayName,
		"approver_username":      record.ApproverUsername,
		"approver_display_name":  record.ApproverDisplayName,
		"description":            record.Description,
		"created_at":             record.CreatedAt, // int64 Unix millis

		// DM-specific fields
		"notification_type": notificationType,
		"is_dm":             true,
	}

	// Optional fields - only include if they have meaningful values
	if record.DecidedAt > 0 {
		props["decided_at"] = record.DecidedAt // int64 Unix millis
	}
	if record.DecisionComment != "" {
		props["decision_comment"] = record.DecisionComment
	}

	// Cancellation fields (if applicable)
	if record.CanceledAt > 0 {
		props["canceled_at"] = record.CanceledAt
	}
	if record.CanceledReason != "" {
		props["canceled_reason"] = record.CanceledReason
	}

	// Verification fields (if applicable)
	if record.VerifiedAt > 0 {
		props["verified_at"] = record.VerifiedAt
	}
	if record.VerificationComment != "" {
		props["verification_comment"] = record.VerificationComment
	}

	// Playbook context (if available)
	if record.PlaybookRunID != "" {
		props["playbook_id"] = record.PlaybookRunID
	}
	if record.PlaybookName != "" {
		props["playbook_title"] = record.PlaybookName
	}
	if record.PlaybookChannelID != "" {
		props["playbook_channel_id"] = record.PlaybookChannelID
	}

	return props
}

// FormatMarkdownFallback creates a markdown message for non-webapp clients.
// This ensures approval information is readable even without custom post rendering.
func FormatMarkdownFallback(record *approval.ApprovalRecord, notificationType string) string {
	if record == nil {
		return ""
	}

	timestamp := time.UnixMilli(record.CreatedAt).UTC()
	timestampStr := timestamp.Format("2006-01-02 15:04:05 MST")

	switch notificationType {
	case NotificationTypeApprovalRequest:
		return fmt.Sprintf("📋 **Approval Request**\n\n"+
			"**From:** @%s (%s)\n"+
			"**Requested:** %s\n"+
			"**Description:**\n%s\n\n"+
			"**Request ID:** `%s`",
			record.RequesterUsername,
			record.RequesterDisplayName,
			timestampStr,
			record.Description,
			record.Code)

	case NotificationTypeOutcome:
		decisionTimestamp := time.UnixMilli(record.DecidedAt).UTC()
		decisionStr := decisionTimestamp.Format("2006-01-02 15:04:05 MST")

		var header, status string
		if record.Status == approval.StatusApproved {
			header = "✅ **Approval Request Approved**"
			status = "**Status:** You may proceed with this action."
		} else {
			header = "❌ **Approval Request Denied**"
			status = "**Status:** This request has been denied."
		}

		msg := fmt.Sprintf("%s\n\n"+
			"**Approver:** @%s (%s)\n"+
			"**Decision Time:** %s\n"+
			"**Request ID:** `%s`\n\n"+
			"**Original Request:**\n> %s",
			header,
			record.ApproverUsername,
			record.ApproverDisplayName,
			decisionStr,
			record.Code,
			record.Description)

		if record.DecisionComment != "" {
			msg += fmt.Sprintf("\n\n**Comment:**\n%s", record.DecisionComment)
		}
		msg += fmt.Sprintf("\n\n%s", status)
		return msg

	case NotificationTypeCancellation:
		canceledAt := time.UnixMilli(record.CanceledAt).UTC()
		canceledAtStr := canceledAt.Format("Jan 02, 2006 3:04 PM")
		canceledReason := record.CanceledReason
		if canceledReason == "" {
			canceledReason = "Not specified"
		}

		return fmt.Sprintf("🚫 **Approval Request Canceled**\n\n"+
			"**Reference:** `%s`\n"+
			"**Requester:** @%s\n"+
			"**Reason:** %s\n"+
			"**Canceled:** %s",
			record.Code,
			record.RequesterUsername,
			canceledReason,
			canceledAtStr)

	case NotificationTypeRequesterCancellation:
		// Sent to requester when approver cancels - shows approver info
		canceledAt := time.UnixMilli(record.CanceledAt).UTC()
		canceledAtStr := canceledAt.Format("Jan 02, 2006 3:04 PM")
		canceledReason := record.CanceledReason
		if canceledReason == "" {
			canceledReason = "Not specified"
		}

		msg := fmt.Sprintf("🚫 **Your Approval Request Was Canceled**\n\n"+
			"**Request ID:** `%s`\n"+
			"**Original Request:** %s\n"+
			"**Canceled By:** @%s (%s)\n"+
			"**Reason:** %s",
			record.Code,
			record.Description,
			record.ApproverUsername,
			record.ApproverDisplayName,
			canceledReason)

		if record.CanceledDetails != "" {
			msg += fmt.Sprintf("\n**Details:** %s", record.CanceledDetails)
		}

		msg += fmt.Sprintf("\n**Canceled:** %s\n\n"+
			"The approver has canceled this approval request. You may submit a new request if needed.",
			canceledAtStr)
		return msg

	case NotificationTypeTimeout:
		return fmt.Sprintf("⏱️ **Approval Request Timed Out**\n\n"+
			"**Request ID:** `%s`\n\n"+
			"**Original Request:**\n> %s\n\n"+
			"**Approver:** @%s (%s)\n\n"+
			"**Reason:** No response within 30 minutes\n\n"+
			"**Status:** This request has been automatically canceled.",
			record.Code,
			record.Description,
			record.ApproverUsername,
			record.ApproverDisplayName)

	case NotificationTypeVerification:
		verifiedTimestamp := time.UnixMilli(record.VerifiedAt).UTC()
		verifiedStr := verifiedTimestamp.Format("2006-01-02 15:04:05 MST")

		msg := fmt.Sprintf("✅ **Approval Request Verified**\n\n"+
			"**Request ID:** `%s`\n\n"+
			"**Original Request:**\n> %s\n\n"+
			"**Requester:** @%s (%s)\n"+
			"**Verified:** %s",
			record.Code,
			record.Description,
			record.RequesterUsername,
			record.RequesterDisplayName,
			verifiedStr)

		if record.VerificationComment != "" {
			msg += fmt.Sprintf("\n\n**Verification Note:**\n> %s", record.VerificationComment)
		}
		return msg

	default:
		return fmt.Sprintf("**Approval Notification**\n\n**Request ID:** `%s`", record.Code)
	}
}
