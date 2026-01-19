package playbooks

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/stretchr/testify/assert"
)

// Story 8.5: Unit tests for playbook status message formatters

func TestFormatApprovedStatusMessage(t *testing.T) {
	t.Run("formats approved message with all fields", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:             "TUZ-2RK",
			Description:      "Deploy v2.1.0 to production",
			ApproverUsername: "jane.doe",
			DecidedAt:        time.Date(2024, 1, 17, 14, 23, 0, 0, time.UTC).UnixMilli(),
			Status:           approval.StatusApproved,
		}

		result := FormatApprovedStatusMessage(record)

		// AC1: Contains emoji, reference code, details, approver, timestamp (markdown table format)
		assert.Contains(t, result, "✅")
		assert.Contains(t, result, "Approval Approved")
		assert.Contains(t, result, "TUZ-2RK")
		assert.Contains(t, result, "Deploy v2.1.0 to production")
		assert.Contains(t, result, "@jane.doe")
		assert.Contains(t, result, "14:23")
		// Verify table structure
		assert.Contains(t, result, "| Field | Value |")
		assert.Contains(t, result, "| **Request ID** |")
		assert.Contains(t, result, "| **Description** |")
		assert.Contains(t, result, "| **Approved By** |")
		assert.Contains(t, result, "| **Time** |")
	})

	t.Run("truncates long description to 100 characters", func(t *testing.T) {
		longDesc := strings.Repeat("x", 105)
		record := &approval.ApprovalRecord{
			Code:             "A-TEST",
			Description:      longDesc,
			ApproverUsername: "approver",
			DecidedAt:        time.Now().UnixMilli(),
		}

		result := FormatApprovedStatusMessage(record)

		// Should truncate at 97 chars + "..." = 100 chars
		assert.NotContains(t, result, strings.Repeat("x", 105))
		assert.Contains(t, result, "...")
	})

	t.Run("handles UTF-8 multibyte characters in description", func(t *testing.T) {
		// Description with emoji (4-byte UTF-8 character)
		record := &approval.ApprovalRecord{
			Code:             "B-TEST",
			Description:      "Deploy 🚀 Production Release with new features " + strings.Repeat("x", 60),
			ApproverUsername: "deployer",
			DecidedAt:        time.Now().UnixMilli(),
		}

		result := FormatApprovedStatusMessage(record)

		// Should not corrupt emoji during truncation
		assert.Contains(t, result, "🚀")
		assert.Contains(t, result, "...")
	})

	t.Run("includes approval note when provided", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:             "C-TEST",
			Description:      "Deploy to production",
			ApproverUsername: "jane.doe",
			DecisionComment:  "Looks good, deployment approved",
			DecidedAt:        time.Now().UnixMilli(),
			Status:           approval.StatusApproved,
		}

		result := FormatApprovedStatusMessage(record)

		// Should include the approval note
		assert.Contains(t, result, "Looks good, deployment approved")
		assert.Contains(t, result, "| **Note** |")
	})

	t.Run("omits note row when no comment provided", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:             "D-TEST",
			Description:      "Deploy to production",
			ApproverUsername: "jane.doe",
			DecisionComment:  "", // No comment
			DecidedAt:        time.Now().UnixMilli(),
			Status:           approval.StatusApproved,
		}

		result := FormatApprovedStatusMessage(record)

		// Should not include note row
		assert.NotContains(t, result, "| **Note** |")
	})
}

func TestFormatDeniedStatusMessage(t *testing.T) {
	t.Run("formats denied message with reason", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:             "A-X7K9Q2",
			Description:      "Emergency DB access",
			ApproverUsername: "security.manager",
			DecisionComment:  "Insufficient justification for P3 incident",
			DecidedAt:        time.Now().UnixMilli(),
			Status:           approval.StatusDenied,
		}

		result := FormatDeniedStatusMessage(record)

		// AC2: Contains emoji, reference code, details, approver, reason (markdown table format)
		assert.Contains(t, result, "❌")
		assert.Contains(t, result, "Approval Denied")
		assert.Contains(t, result, "A-X7K9Q2")
		assert.Contains(t, result, "Emergency DB access")
		assert.Contains(t, result, "@security.manager")
		assert.Contains(t, result, "Insufficient justification for P3 incident")
		// Verify table structure
		assert.Contains(t, result, "| **Request ID** |")
		assert.Contains(t, result, "| **Description** |")
		assert.Contains(t, result, "| **Denied By** |")
		assert.Contains(t, result, "| **Reason** |")
	})

	t.Run("formats denied message without reason", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:             "B-3M8PN",
			Description:      "Purchase software license",
			ApproverUsername: "finance.lead",
			DecisionComment:  "", // No reason provided
			DecidedAt:        time.Now().UnixMilli(),
			Status:           approval.StatusDenied,
		}

		result := FormatDeniedStatusMessage(record)

		// Should not include Reason row if no comment provided (markdown table format)
		assert.Contains(t, result, "❌")
		assert.Contains(t, result, "Approval Denied")
		assert.Contains(t, result, "B-3M8PN")
		assert.NotContains(t, result, "| **Reason** |")
	})

	t.Run("truncates long denial reason", func(t *testing.T) {
		longReason := strings.Repeat("x", 105)
		record := &approval.ApprovalRecord{
			Code:             "C-TEST",
			Description:      "Test request",
			ApproverUsername: "approver",
			DecisionComment:  longReason,
			DecidedAt:        time.Now().UnixMilli(),
		}

		result := FormatDeniedStatusMessage(record)

		// Reason should be truncated
		assert.NotContains(t, result, strings.Repeat("x", 105))
		assert.Contains(t, result, "...")
	})
}

func TestFormatCanceledStatusMessage(t *testing.T) {
	t.Run("formats canceled message with reason", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:            "B-3M8PN",
			Description:     "Purchase software license",
			CanceledReason:  "Duplicate request",
			CanceledDetails: "",
			CanceledAt:      time.Now().UnixMilli(),
			Status:          approval.StatusCanceled,
		}

		result := FormatCanceledStatusMessage(record)

		// AC3: Contains emoji, reference code, details, reason (markdown table format)
		assert.Contains(t, result, "🚫")
		assert.Contains(t, result, "Approval Canceled")
		assert.Contains(t, result, "B-3M8PN")
		assert.Contains(t, result, "Purchase software license")
		assert.Contains(t, result, "Duplicate request")
		// Verify table structure
		assert.Contains(t, result, "| **Request ID** |")
		assert.Contains(t, result, "| **Description** |")
		assert.Contains(t, result, "| **Reason** |")
	})

	t.Run("formats canceled message with reason and additional details", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:            "D-TEST",
			Description:     "Deploy hotfix",
			CanceledReason:  "Other",
			CanceledDetails: "Manager decided to postpone until Monday",
			CanceledAt:      time.Now().UnixMilli(),
			Status:          approval.StatusCanceled,
		}

		result := FormatCanceledStatusMessage(record)

		// Should include both reason and details
		assert.Contains(t, result, "Other")
		assert.Contains(t, result, "Manager decided to postpone until Monday")
	})

	t.Run("handles empty cancellation reason", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:           "E-TEST",
			Description:    "Test request",
			CanceledReason: "", // Empty reason
			CanceledAt:     time.Now().UnixMilli(),
			Status:         approval.StatusCanceled,
		}

		result := FormatCanceledStatusMessage(record)

		// Should show "Not specified" for empty reason (AC3) - in table format
		assert.Contains(t, result, "Not specified")
		assert.Contains(t, result, "| **Reason** |")
	})

	t.Run("truncates long cancellation details", func(t *testing.T) {
		longDetails := strings.Repeat("x", 70) // More than 60 to trigger truncation
		record := &approval.ApprovalRecord{
			Code:            "F-TEST",
			Description:     "Test request",
			CanceledReason:  "Other",
			CanceledDetails: longDetails,
			CanceledAt:      time.Now().UnixMilli(),
		}

		result := FormatCanceledStatusMessage(record)

		// Details should be truncated to 60 chars (57 + "..." = 60)
		assert.Contains(t, result, "...")
		assert.NotContains(t, result, strings.Repeat("x", 70))
	})
}

func TestFormatTimedOutStatusMessage(t *testing.T) {
	t.Run("formats timeout message", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:             "C-4R7QT",
			Description:      "Deploy hotfix to staging",
			ApproverUsername: "lead.engineer",
			Status:           approval.StatusCanceled, // Timeout is a type of cancellation
			CanceledReason:   "Timed out",
		}

		result := FormatTimedOutStatusMessage(record)

		// AC4: Contains emoji, reference code, details, approver (markdown table format)
		assert.Contains(t, result, "⏱️")
		assert.Contains(t, result, "Timed Out")
		assert.Contains(t, result, "C-4R7QT")
		assert.Contains(t, result, "Deploy hotfix to staging")
		assert.Contains(t, result, "@lead.engineer")
		assert.Contains(t, result, "(no response)")
		// Verify table structure
		assert.Contains(t, result, "| **Request ID** |")
		assert.Contains(t, result, "| **Description** |")
		assert.Contains(t, result, "| **Approver** |")
	})

	t.Run("truncates long description", func(t *testing.T) {
		longDesc := strings.Repeat("x", 105)
		record := &approval.ApprovalRecord{
			Code:             "G-TEST",
			Description:      longDesc,
			ApproverUsername: "approver",
		}

		result := FormatTimedOutStatusMessage(record)

		// Should truncate description
		assert.NotContains(t, result, strings.Repeat("x", 105))
		assert.Contains(t, result, "...")
	})
}

func TestFormatTimestamp(t *testing.T) {
	t.Run("formats timestamp as HH:MM", func(t *testing.T) {
		// January 17, 2024 at 14:23:45
		timestamp := time.Date(2024, 1, 17, 14, 23, 45, 0, time.UTC).UnixMilli()

		result := formatTimestamp(timestamp)

		// AC7: Human-readable format "14:23"
		assert.Equal(t, "14:23", result)
	})

	t.Run("handles midnight correctly", func(t *testing.T) {
		timestamp := time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC).UnixMilli()

		result := formatTimestamp(timestamp)

		assert.Equal(t, "00:00", result)
	})

	t.Run("handles end of day correctly", func(t *testing.T) {
		timestamp := time.Date(2024, 1, 17, 23, 59, 0, 0, time.UTC).UnixMilli()

		result := formatTimestamp(timestamp)

		assert.Equal(t, "23:59", result)
	})
}

func TestTruncateString(t *testing.T) {
	t.Run("does not truncate short strings", func(t *testing.T) {
		input := "Short string"
		result := truncateString(input, 100)
		assert.Equal(t, input, result)
	})

	t.Run("truncates long strings at maxLen", func(t *testing.T) {
		input := strings.Repeat("x", 105)
		result := truncateString(input, 100)

		// Should be 97 chars + "..." = 100 total
		assert.Equal(t, 100, utf8.RuneCountInString(result))
		assert.Contains(t, result, "...")
	})

	t.Run("handles UTF-8 multibyte characters correctly", func(t *testing.T) {
		// 55 runes: "Deploy 🚀 " (10 runes) + 45 more runes
		input := "Deploy 🚀 Production Release v2.1.0 for Customer " + strings.Repeat("x", 5)
		result := truncateString(input, 50)

		// Should truncate at 47 runes + "..." = 50 runes total
		assert.Equal(t, 50, utf8.RuneCountInString(result))
		// Emoji should not be corrupted
		assert.Contains(t, result, "🚀")
		assert.Contains(t, result, "...")
	})

	t.Run("handles exactly maxLen characters", func(t *testing.T) {
		input := strings.Repeat("x", 100)
		result := truncateString(input, 100)

		// Should not truncate if exactly at limit
		assert.Equal(t, input, result)
		assert.NotContains(t, result, "...")
	})
}

// Story 9.8: Unit tests for webapp custom post type props formatter

func TestFormatApprovalPropsForWebapp(t *testing.T) {
	t.Run("formats props for pending approval", func(t *testing.T) {
		createdAt := time.Date(2024, 1, 17, 14, 23, 0, 0, time.UTC).UnixMilli()
		record := &approval.ApprovalRecord{
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterUsername:    "john.doe",
			RequesterDisplayName: "John Doe",
			ApproverUsername:     "jane.smith",
			ApproverDisplayName:  "Jane Smith",
			Description:          "Deploy v2.1.0 to production",
			CreatedAt:            createdAt,
			DecidedAt:            0, // Pending, no decision yet
			DecisionComment:      "",
		}

		props := FormatApprovalPropsForWebapp(record)

		// AC3: Verify all required fields with correct types
		assert.Equal(t, "A-X7K9Q2", props["approval_code"])
		assert.Equal(t, approval.StatusPending, props["approval_status"])
		assert.Equal(t, "john.doe", props["requester_username"])
		assert.Equal(t, "John Doe", props["requester_display_name"])
		assert.Equal(t, "jane.smith", props["approver_username"])
		assert.Equal(t, "Jane Smith", props["approver_display_name"])
		assert.Equal(t, "Deploy v2.1.0 to production", props["description"])

		// Verify timestamp types (int64, not strings)
		assert.IsType(t, int64(0), props["created_at"])
		assert.Equal(t, createdAt, props["created_at"])
		assert.IsType(t, int64(0), props["decided_at"])
		assert.Equal(t, int64(0), props["decided_at"])

		// Optional fields
		assert.Equal(t, "", props["decision_comment"])
		assert.Equal(t, "", props["note"])
	})

	t.Run("formats props for approved approval with note", func(t *testing.T) {
		createdAt := time.Date(2024, 1, 17, 14, 23, 0, 0, time.UTC).UnixMilli()
		decidedAt := time.Date(2024, 1, 17, 15, 10, 0, 0, time.UTC).UnixMilli()
		record := &approval.ApprovalRecord{
			Code:                 "B-3M8PN",
			Status:               approval.StatusApproved,
			RequesterUsername:    "developer",
			RequesterDisplayName: "Developer User",
			ApproverUsername:     "tech.lead",
			ApproverDisplayName:  "Tech Lead",
			Description:          "Emergency hotfix deployment",
			CreatedAt:            createdAt,
			DecidedAt:            decidedAt,
			DecisionComment:      "Looks good, proceed with deployment",
		}

		props := FormatApprovalPropsForWebapp(record)

		assert.Equal(t, approval.StatusApproved, props["approval_status"])
		assert.Equal(t, decidedAt, props["decided_at"])
		assert.Equal(t, "Looks good, proceed with deployment", props["decision_comment"])
		assert.Equal(t, "Looks good, proceed with deployment", props["note"]) // Note mirrors decision_comment
	})

	t.Run("formats props for denied approval with reason", func(t *testing.T) {
		createdAt := time.Date(2024, 1, 17, 14, 23, 0, 0, time.UTC).UnixMilli()
		decidedAt := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC).UnixMilli()
		record := &approval.ApprovalRecord{
			Code:                 "C-4R7QT",
			Status:               approval.StatusDenied,
			RequesterUsername:    "contractor",
			RequesterDisplayName: "Contractor User",
			ApproverUsername:     "security.lead",
			ApproverDisplayName:  "Security Lead",
			Description:          "Access to production database",
			CreatedAt:            createdAt,
			DecidedAt:            decidedAt,
			DecisionComment:      "Insufficient justification for P3 incident",
		}

		props := FormatApprovalPropsForWebapp(record)

		assert.Equal(t, approval.StatusDenied, props["approval_status"])
		assert.Equal(t, decidedAt, props["decided_at"])
		assert.Equal(t, "Insufficient justification for P3 incident", props["decision_comment"])
	})

	t.Run("formats props for canceled approval", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "D-9XP2Q",
			Status:               approval.StatusCanceled,
			RequesterUsername:    "requester",
			RequesterDisplayName: "Requester",
			ApproverUsername:     "approver",
			ApproverDisplayName:  "Approver",
			Description:          "Duplicate request",
			CreatedAt:            time.Now().UnixMilli(),
			DecidedAt:            0,
			DecisionComment:      "", // Canceled might not have decision comment
		}

		props := FormatApprovalPropsForWebapp(record)

		assert.Equal(t, approval.StatusCanceled, props["approval_status"])
		assert.Equal(t, "", props["decision_comment"])
	})

	t.Run("formats props for timeout status", func(t *testing.T) {
		createdAt := time.Date(2024, 1, 17, 14, 23, 0, 0, time.UTC).UnixMilli()
		timeoutAt := time.Date(2024, 1, 17, 16, 23, 0, 0, time.UTC).UnixMilli() // 2 hours later
		record := &approval.ApprovalRecord{
			Code:                 "E-TEST",
			Status:               "timeout", // Story 9.7 mentions timeout status
			RequesterUsername:    "requester",
			RequesterDisplayName: "Requester",
			ApproverUsername:     "approver",
			ApproverDisplayName:  "Approver",
			Description:          "Timed out request",
			CreatedAt:            createdAt,
			DecidedAt:            timeoutAt,
			DecisionComment:      "",
		}

		props := FormatApprovalPropsForWebapp(record)

		assert.Equal(t, "timeout", props["approval_status"])
		assert.Equal(t, timeoutAt, props["decided_at"])
	})

	t.Run("handles nil record gracefully", func(t *testing.T) {
		props := FormatApprovalPropsForWebapp(nil)

		// Should return empty map, not crash
		assert.NotNil(t, props)
		assert.Empty(t, props)
	})

	t.Run("handles empty strings in optional fields", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "F-TEST",
			Status:               approval.StatusPending,
			RequesterUsername:    "",
			RequesterDisplayName: "",
			ApproverUsername:     "",
			ApproverDisplayName:  "",
			Description:          "",
			CreatedAt:            time.Now().UnixMilli(),
			DecidedAt:            0,
			DecisionComment:      "",
		}

		props := FormatApprovalPropsForWebapp(record)

		// Should not crash with empty strings
		assert.Equal(t, "", props["requester_username"])
		assert.Equal(t, "", props["description"])
		assert.NotNil(t, props)
	})

	t.Run("preserves timestamp precision (int64)", func(t *testing.T) {
		// Use precise timestamp with milliseconds
		preciseTimestamp := int64(1705502580123) // Jan 17, 2024 14:23:00.123
		record := &approval.ApprovalRecord{
			Code:                 "G-TEST",
			Status:               approval.StatusPending,
			RequesterUsername:    "user",
			RequesterDisplayName: "User",
			ApproverUsername:     "approver",
			ApproverDisplayName:  "Approver",
			Description:          "Test",
			CreatedAt:            preciseTimestamp,
			DecidedAt:            0,
		}

		props := FormatApprovalPropsForWebapp(record)

		// Verify exact timestamp (no precision loss)
		assert.Equal(t, preciseTimestamp, props["created_at"])
		assert.IsType(t, int64(0), props["created_at"])
	})

	t.Run("uses snake_case field names for webapp", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "H-TEST",
			Status:               approval.StatusPending,
			RequesterUsername:    "user",
			RequesterDisplayName: "User Display",
			ApproverUsername:     "approver",
			ApproverDisplayName:  "Approver Display",
			Description:          "Test",
			CreatedAt:            time.Now().UnixMilli(),
		}

		props := FormatApprovalPropsForWebapp(record)

		// Verify snake_case naming (not camelCase)
		_, hasSnakeCase := props["approval_code"]
		assert.True(t, hasSnakeCase, "Should use snake_case 'approval_code'")

		_, hasRequesterUsername := props["requester_username"]
		assert.True(t, hasRequesterUsername, "Should use snake_case 'requester_username'")

		_, hasDisplayName := props["requester_display_name"]
		assert.True(t, hasDisplayName, "Should use snake_case 'requester_display_name'")

		// Verify NOT camelCase
		_, hasCamelCase := props["approvalCode"]
		assert.False(t, hasCamelCase, "Should NOT use camelCase 'approvalCode'")
	})
}
