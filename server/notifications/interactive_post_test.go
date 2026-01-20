package notifications

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateApproveAction(t *testing.T) {
	code := "A-TEST01"
	action := CreateApproveAction(code)

	// AC2: Verify PostAction structure
	assert.Equal(t, "approve", action.Id)
	assert.Equal(t, "Approve", action.Name)
	assert.Equal(t, "button", action.Type)
	assert.Equal(t, "success", action.Style)

	// AC2: Verify Integration URL format
	require.NotNil(t, action.Integration)
	expectedURL := "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-TEST01/approve"
	assert.Equal(t, expectedURL, action.Integration.URL)
}

func TestCreateDenyAction(t *testing.T) {
	code := "A-TEST02"
	action := CreateDenyAction(code)

	// AC2: Verify PostAction structure
	assert.Equal(t, "deny", action.Id)
	assert.Equal(t, "Deny", action.Name)
	assert.Equal(t, "button", action.Type)
	assert.Equal(t, "danger", action.Style)

	// AC2: Verify Integration URL format
	require.NotNil(t, action.Integration)
	expectedURL := "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-TEST02/deny"
	assert.Equal(t, expectedURL, action.Integration.URL)
}

func TestCreateInteractiveApprovalPost_ApprovalRequest(t *testing.T) {
	record := &approval.ApprovalRecord{
		ID:                   "test-id-123",
		Code:                 "A-XYZ123",
		Status:               approval.StatusPending,
		RequesterID:          "user1",
		RequesterUsername:    "alice",
		RequesterDisplayName: "Alice Smith",
		ApproverID:           "user2",
		ApproverUsername:     "bob",
		ApproverDisplayName:  "Bob Jones",
		Description:          "Test approval request",
		CreatedAt:            1705600000000,
	}

	post := CreateInteractiveApprovalPost("bot-user", "channel-123", record, NotificationTypeApprovalRequest)

	// AC1: Verify custom post type
	require.NotNil(t, post)
	assert.Equal(t, CustomApprovalDMPostType, post.Type)
	assert.Equal(t, "bot-user", post.UserId)
	assert.Equal(t, "channel-123", post.ChannelId)

	// AC5: Verify markdown fallback message is set
	assert.Contains(t, post.Message, "Approval Request")
	assert.Contains(t, post.Message, "A-XYZ123")

	// AC4: Verify Props schema
	assert.Equal(t, "A-XYZ123", post.GetProp("approval_code"))
	assert.Equal(t, "pending", post.GetProp("approval_status"))
	assert.Equal(t, "alice", post.GetProp("requester_username"))
	assert.Equal(t, "Alice Smith", post.GetProp("requester_display_name"))
	assert.Equal(t, "bob", post.GetProp("approver_username"))
	assert.Equal(t, "Bob Jones", post.GetProp("approver_display_name"))
	assert.Equal(t, "Test approval request", post.GetProp("description"))
	assert.Equal(t, int64(1705600000000), post.GetProp("created_at"))
	assert.Equal(t, NotificationTypeApprovalRequest, post.GetProp("notification_type"))
	assert.Equal(t, true, post.GetProp("is_dm"))

	// Verify ParseSlackAttachment was called (attachments should be populated)
	attachments := post.GetProp("attachments")
	require.NotNil(t, attachments, "ParseSlackAttachment should populate attachments")
}

func TestCreateInteractiveApprovalPost_NilRecord(t *testing.T) {
	post := CreateInteractiveApprovalPost("bot-user", "channel-123", nil, NotificationTypeApprovalRequest)
	assert.Nil(t, post)
}

func TestCreateInteractiveApprovalPost_EmptyBotUserID(t *testing.T) {
	record := &approval.ApprovalRecord{
		Code:              "A-TEST01",
		Status:            approval.StatusPending,
		RequesterUsername: "alice",
		ApproverUsername:  "bob",
		CreatedAt:         1705600000000,
	}
	post := CreateInteractiveApprovalPost("", "channel-123", record, NotificationTypeApprovalRequest)
	assert.Nil(t, post, "should return nil for empty botUserID")
}

func TestCreateInteractiveApprovalPost_EmptyChannelID(t *testing.T) {
	record := &approval.ApprovalRecord{
		Code:              "A-TEST01",
		Status:            approval.StatusPending,
		RequesterUsername: "alice",
		ApproverUsername:  "bob",
		CreatedAt:         1705600000000,
	}
	post := CreateInteractiveApprovalPost("bot-user", "", record, NotificationTypeApprovalRequest)
	assert.Nil(t, post, "should return nil for empty channelID")
}

func TestCreateInteractiveApprovalPost_Outcome(t *testing.T) {
	record := &approval.ApprovalRecord{
		ID:                   "test-id-456",
		Code:                 "A-ABC789",
		Status:               approval.StatusApproved,
		RequesterID:          "user1",
		RequesterUsername:    "alice",
		RequesterDisplayName: "Alice Smith",
		ApproverID:           "user2",
		ApproverUsername:     "bob",
		ApproverDisplayName:  "Bob Jones",
		Description:          "Test approval request",
		CreatedAt:            1705600000000,
		DecidedAt:            1705600300000,
		DecisionComment:      "Looks good!",
	}

	post := CreateInteractiveApprovalPost("bot-user", "channel-123", record, NotificationTypeOutcome)

	require.NotNil(t, post)
	assert.Equal(t, CustomApprovalDMPostType, post.Type)
	assert.Equal(t, NotificationTypeOutcome, post.GetProp("notification_type"))
	assert.Equal(t, int64(1705600300000), post.GetProp("decided_at"))
	assert.Equal(t, "Looks good!", post.GetProp("decision_comment"))

	// Outcome posts should NOT have buttons (verify attachments has no actions)
	// The post should still have attachments but with empty actions
}

func TestFormatApprovalPropsForDM(t *testing.T) {
	t.Run("full record with all fields", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-FULL01",
			Status:               approval.StatusApproved,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Full test",
			CreatedAt:            1705600000000,
			DecidedAt:            1705600300000,
			DecisionComment:      "Approved!",
			PlaybookRunID:        "run-123",
			PlaybookName:         "Test Playbook",
			PlaybookChannelID:    "channel-pb",
		}

		props := FormatApprovalPropsForDM(record, NotificationTypeOutcome)

		// Required fields
		assert.Equal(t, "A-FULL01", props["approval_code"])
		assert.Equal(t, "approved", props["approval_status"])
		assert.Equal(t, "alice", props["requester_username"])
		assert.Equal(t, "Alice Smith", props["requester_display_name"])
		assert.Equal(t, "bob", props["approver_username"])
		assert.Equal(t, "Bob Jones", props["approver_display_name"])
		assert.Equal(t, "Full test", props["description"])
		assert.Equal(t, int64(1705600000000), props["created_at"])
		assert.Equal(t, NotificationTypeOutcome, props["notification_type"])
		assert.Equal(t, true, props["is_dm"])

		// Optional fields
		assert.Equal(t, int64(1705600300000), props["decided_at"])
		assert.Equal(t, "Approved!", props["decision_comment"])

		// Playbook fields
		assert.Equal(t, "run-123", props["playbook_id"])
		assert.Equal(t, "Test Playbook", props["playbook_title"])
		assert.Equal(t, "channel-pb", props["playbook_channel_id"])
	})

	t.Run("nil record returns empty map", func(t *testing.T) {
		props := FormatApprovalPropsForDM(nil, NotificationTypeApprovalRequest)
		assert.Empty(t, props)
	})

	t.Run("pending record without optional fields", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-PEND01",
			Status:               approval.StatusPending,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob",
			Description:          "Pending test",
			CreatedAt:            1705600000000,
		}

		props := FormatApprovalPropsForDM(record, NotificationTypeApprovalRequest)

		// decided_at should not be present (it's 0)
		_, hasDecidedAt := props["decided_at"]
		assert.False(t, hasDecidedAt, "decided_at should not be included when 0")

		// decision_comment should not be present (empty string)
		_, hasDecisionComment := props["decision_comment"]
		assert.False(t, hasDecisionComment, "decision_comment should not be included when empty")
	})

	t.Run("cancellation fields included when present", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-CANC01",
			Status:               approval.StatusCanceled,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob",
			Description:          "Canceled test",
			CreatedAt:            1705600000000,
			CanceledAt:           1705600600000,
			CanceledReason:       "No longer needed",
		}

		props := FormatApprovalPropsForDM(record, NotificationTypeCancellation)

		assert.Equal(t, int64(1705600600000), props["canceled_at"])
		assert.Equal(t, "No longer needed", props["canceled_reason"])
	})

	t.Run("verification fields included when present", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-VERI01",
			Status:               approval.StatusApproved,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob",
			Description:          "Verified test",
			CreatedAt:            1705600000000,
			VerifiedAt:           1705600900000,
			VerificationComment:  "Task completed",
		}

		props := FormatApprovalPropsForDM(record, NotificationTypeVerification)

		assert.Equal(t, int64(1705600900000), props["verified_at"])
		assert.Equal(t, "Task completed", props["verification_comment"])
	})
}

func TestFormatMarkdownFallback(t *testing.T) {
	t.Run("approval request format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD001",
			Status:               approval.StatusPending,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
		}

		msg := FormatMarkdownFallback(record, NotificationTypeApprovalRequest)

		assert.Contains(t, msg, "📋 **Approval Request**")
		assert.Contains(t, msg, "@alice")
		assert.Contains(t, msg, "Alice Smith")
		assert.Contains(t, msg, "Deploy to production")
		assert.Contains(t, msg, "`A-MD001`")
	})

	t.Run("outcome approved format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD002",
			Status:               approval.StatusApproved,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			DecidedAt:            1705600300000,
			DecisionComment:      "Looks good!",
		}

		msg := FormatMarkdownFallback(record, NotificationTypeOutcome)

		assert.Contains(t, msg, "✅ **Approval Request Approved**")
		assert.Contains(t, msg, "@bob")
		assert.Contains(t, msg, "Bob Jones")
		assert.Contains(t, msg, "Looks good!")
		assert.Contains(t, msg, "You may proceed with this action")
	})

	t.Run("outcome denied format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD003",
			Status:               approval.StatusDenied,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			DecidedAt:            1705600300000,
		}

		msg := FormatMarkdownFallback(record, NotificationTypeOutcome)

		assert.Contains(t, msg, "❌ **Approval Request Denied**")
		assert.Contains(t, msg, "This request has been denied")
	})

	t.Run("cancellation format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD004",
			Status:               approval.StatusCanceled,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			CanceledAt:           1705600600000,
			CanceledReason:       "Changed my mind",
		}

		msg := FormatMarkdownFallback(record, NotificationTypeCancellation)

		assert.Contains(t, msg, "🚫 **Approval Request Canceled**")
		assert.Contains(t, msg, "Changed my mind")
	})

	t.Run("cancellation format with empty reason", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD005",
			Status:               approval.StatusCanceled,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			CanceledAt:           1705600600000,
		}

		msg := FormatMarkdownFallback(record, NotificationTypeCancellation)

		assert.Contains(t, msg, "Not specified")
	})

	// Story 10.6: Test for requester_cancellation (sent to requester when approver cancels)
	t.Run("requester cancellation format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD009",
			Status:               approval.StatusCanceled,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			CanceledAt:           1705600600000,
			CanceledReason:       "Budget constraints",
		}

		msg := FormatMarkdownFallback(record, NotificationTypeRequesterCancellation)

		assert.Contains(t, msg, "🚫 **Your Approval Request Was Canceled**")
		assert.Contains(t, msg, "**Canceled By:** @bob (Bob Jones)")
		assert.Contains(t, msg, "Budget constraints")
		assert.Contains(t, msg, "Deploy to production")
		assert.Contains(t, msg, "The approver has canceled this approval request")
	})

	t.Run("requester cancellation format with empty reason", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD010",
			Status:               approval.StatusCanceled,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			CanceledAt:           1705600600000,
			CanceledReason:       "", // Empty reason
		}

		msg := FormatMarkdownFallback(record, NotificationTypeRequesterCancellation)

		assert.Contains(t, msg, "**Reason:** Not specified")
	})

	t.Run("timeout format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD006",
			Status:               approval.StatusTimeout,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
		}

		msg := FormatMarkdownFallback(record, NotificationTypeTimeout)

		assert.Contains(t, msg, "⏱️ **Approval Request Timed Out**")
		assert.Contains(t, msg, "No response within 30 minutes")
		assert.Contains(t, msg, "automatically canceled")
	})

	t.Run("verification format", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code:                 "A-MD007",
			Status:               approval.StatusApproved,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1705600000000,
			VerifiedAt:           1705600900000,
			VerificationComment:  "Deployment successful",
		}

		msg := FormatMarkdownFallback(record, NotificationTypeVerification)

		assert.Contains(t, msg, "✅ **Approval Request Verified**")
		assert.Contains(t, msg, "Deployment successful")
	})

	t.Run("nil record returns empty string", func(t *testing.T) {
		msg := FormatMarkdownFallback(nil, NotificationTypeApprovalRequest)
		assert.Empty(t, msg)
	})

	t.Run("unknown type returns fallback", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			Code: "A-MD008",
		}

		msg := FormatMarkdownFallback(record, "unknown")

		assert.Contains(t, msg, "Approval Notification")
		assert.Contains(t, msg, "A-MD008")
	})
}

func TestGetNotificationTitle(t *testing.T) {
	tests := []struct {
		notificationType string
		status           string
		expected         string
	}{
		{NotificationTypeApprovalRequest, approval.StatusPending, "Approval Request"},
		{NotificationTypeOutcome, approval.StatusApproved, "Approval Request Approved"},
		{NotificationTypeOutcome, approval.StatusDenied, "Approval Request Denied"},
		{NotificationTypeCancellation, approval.StatusCanceled, "Approval Request Canceled"},
		{NotificationTypeRequesterCancellation, approval.StatusCanceled, "Your Approval Request Was Canceled"}, // Story 10.6
		{NotificationTypeTimeout, approval.StatusTimeout, "Approval Request Timed Out"},
		{NotificationTypeVerification, approval.StatusApproved, "Approval Request Verified"},
		{"unknown", "", "Approval Notification"},
	}

	for _, tt := range tests {
		t.Run(tt.notificationType+"_"+tt.status, func(t *testing.T) {
			result := getNotificationTitle(tt.notificationType, tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateApproveAction_EmptyCode(t *testing.T) {
	// Edge case: empty code generates malformed URL
	action := CreateApproveAction("")
	assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/api/v1/approval//approve", action.Integration.URL)
	// Note: This is technically valid but API handler should reject empty code
}

func TestCreateDenyAction_EmptyCode(t *testing.T) {
	// Edge case: empty code generates malformed URL
	action := CreateDenyAction("")
	assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/api/v1/approval//deny", action.Integration.URL)
	// Note: This is technically valid but API handler should reject empty code
}

func TestIntegrationURLFormat(t *testing.T) {
	// Verify URL format matches AC2 requirement
	codes := []string{"A-ABC123", "A-XYZ789", "A-TEST01"}

	for _, code := range codes {
		approveAction := CreateApproveAction(code)
		denyAction := CreateDenyAction(code)

		// URLs should contain the plugin ID and API path
		assert.True(t, strings.HasPrefix(approveAction.Integration.URL, "/plugins/com.mattermost.plugin-approver2/api/v1/approval/"))
		assert.True(t, strings.HasSuffix(approveAction.Integration.URL, "/approve"))
		assert.Contains(t, approveAction.Integration.URL, code)

		assert.True(t, strings.HasPrefix(denyAction.Integration.URL, "/plugins/com.mattermost.plugin-approver2/api/v1/approval/"))
		assert.True(t, strings.HasSuffix(denyAction.Integration.URL, "/deny"))
		assert.Contains(t, denyAction.Integration.URL, code)
	}
}

func TestCreateInteractiveApprovalPost_NoButtonsForNonPending(t *testing.T) {
	// When status is not pending or notification type is not approval_request,
	// there should be no buttons in the post

	testCases := []struct {
		name             string
		status           string
		notificationType string
	}{
		{"approved outcome", approval.StatusApproved, NotificationTypeOutcome},
		{"denied outcome", approval.StatusDenied, NotificationTypeOutcome},
		{"cancellation", approval.StatusCanceled, NotificationTypeCancellation},
		{"timeout", approval.StatusTimeout, NotificationTypeTimeout},
		{"verification", approval.StatusApproved, NotificationTypeVerification},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record := &approval.ApprovalRecord{
				ID:                   "test-id",
				Code:                 "A-TEST01",
				Status:               tc.status,
				RequesterUsername:    "alice",
				RequesterDisplayName: "Alice",
				ApproverUsername:     "bob",
				ApproverDisplayName:  "Bob",
				Description:          "Test",
				CreatedAt:            1705600000000,
			}

			post := CreateInteractiveApprovalPost("bot", "channel", record, tc.notificationType)

			require.NotNil(t, post)

			// Verify attachments exist but have no actions (no buttons)
			attachments := post.GetProp("attachments")
			require.NotNil(t, attachments, "attachments should exist")

			attachmentSlice, ok := attachments.([]*model.SlackAttachment)
			require.True(t, ok, "attachments should be []*model.SlackAttachment")
			require.Len(t, attachmentSlice, 1, "should have exactly one attachment")

			// The attachment should have no actions (no buttons)
			assert.Empty(t, attachmentSlice[0].Actions, "non-pending posts should have no action buttons")
		})
	}
}
