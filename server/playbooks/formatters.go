package playbooks

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
)

// FormatApprovedStatusMessage formats the playbook channel status message for approved requests
// Story 8.5 AC1: "✅ **Approved:** [CODE] | [Details] | Approved by @approver at [time]"
func FormatApprovedStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 100)
	timeStr := formatTimestamp(record.DecidedAt)

	return fmt.Sprintf("✅ **Approved:** %s | %s | Approved by @%s at %s",
		record.Code,
		details,
		record.ApproverUsername,
		timeStr)
}

// FormatDeniedStatusMessage formats the playbook channel status message for denied requests
// Story 8.5 AC2: "❌ **Denied:** [CODE] | [Details] | Denied by @approver | Reason: [reason]"
func FormatDeniedStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 100)

	message := fmt.Sprintf("❌ **Denied:** %s | %s | Denied by @%s",
		record.Code,
		details,
		record.ApproverUsername)

	// Add reason if provided (Story 8.5 AC2)
	if record.DecisionComment != "" {
		reason := truncateString(record.DecisionComment, 100)
		message += fmt.Sprintf(" | Reason: %s", reason)
	}

	return message
}

// FormatCanceledStatusMessage formats the playbook channel status message for canceled requests
// Story 8.5 AC3: "🚫 **Canceled:** [CODE] | [Details] | Reason: [cancellation reason]"
func FormatCanceledStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 100)
	reason := record.CanceledReason
	if reason == "" {
		reason = "Not specified"
	}

	// Include additional details if present (Story 7.3 integration)
	if record.CanceledDetails != "" {
		detailsText := truncateString(record.CanceledDetails, 50)
		reason = fmt.Sprintf("%s (%s)", reason, detailsText)
	}

	return fmt.Sprintf("🚫 **Canceled:** %s | %s | Reason: %s",
		record.Code,
		details,
		reason)
}

// FormatTimedOutStatusMessage formats the playbook channel status message for timed-out requests
// Story 8.5 AC4: "⏱️ **Timeout:** [CODE] | [Details] | No response from @approver"
func FormatTimedOutStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 100)

	return fmt.Sprintf("⏱️ **Timeout:** %s | %s | No response from @%s",
		record.Code,
		details,
		record.ApproverUsername)
}

// formatTimestamp formats a Unix millisecond timestamp as human-readable time
// Story 8.5 AC7: "14:23" or "Jan 17, 14:23" format
// Uses simple HH:MM format in UTC for consistency across timezones
func formatTimestamp(unixMillis int64) string {
	t := time.UnixMilli(unixMillis).UTC()
	return t.Format("15:04") // "14:23" format
}

// truncateString truncates a string to maxLen characters using UTF-8-safe rune truncation
// Story 8.5 AC: Subtask 1.7 - Truncate long details/reasons to 100 characters
// Uses rune-based truncation to avoid corrupting UTF-8 multibyte characters (learned from Story 8.4 review)
func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}
