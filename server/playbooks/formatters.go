package playbooks

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
)

// FormatApprovedStatusMessage formats the playbook channel status message for approved requests
// GitHub Issue #2: Using markdown table for nice formatting without Playbooks API side effects
func FormatApprovedStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 80)
	timeStr := formatTimestamp(record.DecidedAt)

	noteRow := ""
	if record.DecisionComment != "" {
		note := truncateString(record.DecisionComment, 80)
		noteRow = fmt.Sprintf("\n| **Note** | %s |", note)
	}

	return fmt.Sprintf(`### ✅ Approval Approved

| Field | Value |
|:------|:------|
| **Request ID** | %s |
| **Description** | %s |
| **Approved By** | @%s |
| **Time** | %s |%s`,
		record.Code,
		details,
		record.ApproverUsername,
		timeStr,
		noteRow)
}

// FormatDeniedStatusMessage formats the playbook channel status message for denied requests
// GitHub Issue #2: Using markdown table for nice formatting without Playbooks API side effects
func FormatDeniedStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 80)

	reasonRow := ""
	if record.DecisionComment != "" {
		reason := truncateString(record.DecisionComment, 80)
		reasonRow = fmt.Sprintf("\n| **Reason** | %s |", reason)
	}

	return fmt.Sprintf(`### ❌ Approval Denied

| Field | Value |
|:------|:------|
| **Request ID** | %s |
| **Description** | %s |
| **Denied By** | @%s |%s`,
		record.Code,
		details,
		record.ApproverUsername,
		reasonRow)
}

// FormatCanceledStatusMessage formats the playbook channel status message for canceled requests
// GitHub Issue #2: Using markdown table for nice formatting without Playbooks API side effects
func FormatCanceledStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 80)
	reason := record.CanceledReason
	if reason == "" {
		reason = "Not specified"
	}

	// Include additional details if present (Story 7.3 integration)
	if record.CanceledDetails != "" {
		detailsText := truncateString(record.CanceledDetails, 60)
		reason = fmt.Sprintf("%s (%s)", reason, detailsText)
	}

	return fmt.Sprintf(`### 🚫 Approval Canceled

| Field | Value |
|:------|:------|
| **Request ID** | %s |
| **Description** | %s |
| **Reason** | %s |`,
		record.Code,
		details,
		reason)
}

// FormatTimedOutStatusMessage formats the playbook channel status message for timed-out requests
// GitHub Issue #2: Using markdown table for nice formatting without Playbooks API side effects
func FormatTimedOutStatusMessage(record *approval.ApprovalRecord) string {
	details := truncateString(record.Description, 80)

	return fmt.Sprintf(`### ⏱️ Approval Timed Out

| Field | Value |
|:------|:------|
| **Request ID** | %s |
| **Description** | %s |
| **Approver** | @%s (no response) |`,
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
