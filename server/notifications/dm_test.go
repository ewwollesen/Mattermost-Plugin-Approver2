package notifications

import (
	"strings"
	"testing"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendApprovalRequestDM(t *testing.T) {
	t.Run("successful DM send to approver", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == botUserID &&
				post.ChannelId == dmChannelID &&
				strings.Contains(post.Message, "📋 **Approval Request**") &&
				strings.Contains(post.Message, "@alice (Alice Carter)") &&
				strings.Contains(post.Message, "Deploy hotfix to production") &&
				strings.Contains(post.Message, "A-X7K9Q2")
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Create test approval record
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix to production",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		// Execute
		_, err := SendApprovalRequestDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("message format matches AC2 exactly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy the hotfix to production environment",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify exact format
		assert.Contains(t, capturedMessage, "📋 **Approval Request**")
		assert.Contains(t, capturedMessage, "**From:** @alice (Alice Carter)")
		assert.Contains(t, capturedMessage, "**Requested:**")
		assert.Contains(t, capturedMessage, "**Description:**")
		assert.Contains(t, capturedMessage, "Deploy the hotfix to production environment")
		assert.Contains(t, capturedMessage, "**Request ID:** `A-X7K9Q2`")
	})

	t.Run("timestamp format is YYYY-MM-DD HH:MM:SS UTC", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Test description",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify timestamp format: YYYY-MM-DD HH:MM:SS UTC
		expectedTime := time.UnixMilli(1704988800000).UTC()
		expectedTimestamp := expectedTime.Format("2006-01-02 15:04:05 MST")
		assert.Contains(t, capturedMessage, expectedTimestamp)
	})

	t.Run("DM send failure handled gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "network error"})

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		// Execute - should return error
		_, err := SendApprovalRequestDM(api, botUserID, record)

		// Assert error is returned for caller to log
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send DM")
		api.AssertExpectations(t)
	})

	t.Run("GetDMChannelID handles disabled DMs gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"

		api.On("GetDirectChannel", botUserID, approverID).Return(nil, &model.AppError{Message: "DMs disabled"})

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		// Execute - should return error for DM channel creation failure
		_, err := SendApprovalRequestDM(api, botUserID, record)

		// Assert error is returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel")
		api.AssertExpectations(t)
	})

	t.Run("bot user ID not available", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           "approver456",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		// Execute - should return error for empty bot user ID
		_, err := SendApprovalRequestDM(api, "", record)

		// Assert error is returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot user ID not available")
		api.AssertExpectations(t)
	})

	t.Run("nil record validation", func(t *testing.T) {
		api := &plugintest.API{}

		// Execute - should return error for nil record
		_, err := SendApprovalRequestDM(api, "bot123", nil)

		// Assert error is returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record is nil")
		api.AssertExpectations(t)
	})

	t.Run("empty record ID validation", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:                   "", // Empty ID
			Code:                 "A-X7K9Q2",
			ApproverID:           "approver456",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		// Execute - should return error for empty record ID
		_, err := SendApprovalRequestDM(api, "bot123", record)

		// Assert error is returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record ID is empty")
		api.AssertExpectations(t)
	})
}

func TestGetDMChannelID(t *testing.T) {
	t.Run("successfully gets DM channel ID", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		expectedChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: expectedChannelID}, nil)

		channelID, err := GetDMChannelID(api, botUserID, approverID)

		assert.NoError(t, err)
		assert.Equal(t, expectedChannelID, channelID)
		api.AssertExpectations(t)
	})

	t.Run("handles GetDirectChannel failure", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"

		api.On("GetDirectChannel", botUserID, approverID).Return(nil, &model.AppError{Message: "user not found"})

		channelID, err := GetDMChannelID(api, botUserID, approverID)

		assert.Error(t, err)
		assert.Empty(t, channelID)
		assert.Contains(t, err.Error(), "failed to get DM channel")
		api.AssertExpectations(t)
	})
}

func TestSendApprovalRequestDM_WithActionButtons(t *testing.T) {
	t.Run("includes approve and deny buttons in attachments", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Create test approval record
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix to production",
			CreatedAt:            1704988800000,
		}

		// Execute
		_, err := SendApprovalRequestDM(api, botUserID, record)

		// Assert no error
		assert.NoError(t, err)

		// Verify attachments exist
		attachments, ok := capturedPost.Props["attachments"].([]any)
		assert.True(t, ok, "Props.attachments should be array")
		assert.Len(t, attachments, 1, "should have 1 attachment")

		// Verify actions array exists
		attachment := attachments[0].(map[string]any)
		actions, ok := attachment["actions"].([]any)
		assert.True(t, ok, "attachment should have actions")
		assert.Len(t, actions, 2, "should have 2 buttons")
	})

	t.Run("approve button configured correctly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Extract approve button
		attachments := capturedPost.Props["attachments"].([]any)
		attachment := attachments[0].(map[string]any)
		actions := attachment["actions"].([]any)
		approveButton := actions[0].(map[string]any)

		// Verify approve button properties (no ID - using custom integration URL)
		assert.Equal(t, "Approve", approveButton["name"])
		assert.Equal(t, "primary", approveButton["style"])

		// Verify integration configuration
		integration := approveButton["integration"].(map[string]any)
		assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/action", integration["url"])

		// Verify context data
		context := integration["context"].(map[string]any)
		assert.Equal(t, "record123", context["approval_id"])
		assert.Equal(t, "approve", context["action"])
	})

	t.Run("deny button configured correctly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Extract deny button
		attachments := capturedPost.Props["attachments"].([]any)
		attachment := attachments[0].(map[string]any)
		actions := attachment["actions"].([]any)
		denyButton := actions[1].(map[string]any)

		// Verify deny button properties (no ID - using custom integration URL)
		assert.Equal(t, "Deny", denyButton["name"])
		assert.Equal(t, "danger", denyButton["style"])

		// Verify integration configuration
		integration := denyButton["integration"].(map[string]any)
		assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/action", integration["url"])

		// Verify context data
		context := integration["context"].(map[string]any)
		assert.Equal(t, "record123", context["approval_id"])
		assert.Equal(t, "deny", context["action"])
	})

	t.Run("message format remains intact with buttons", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy the hotfix to production",
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify message content is unchanged
		assert.Contains(t, capturedPost.Message, "📋 **Approval Request**")
		assert.Contains(t, capturedPost.Message, "**From:** @alice (Alice Carter)")
		assert.Contains(t, capturedPost.Message, "**Requested:**")
		assert.Contains(t, capturedPost.Message, "**Description:**")
		assert.Contains(t, capturedPost.Message, "Deploy the hotfix to production")
		assert.Contains(t, capturedPost.Message, "**Request ID:** `A-X7K9Q2`")

		// Verify buttons are in Props, not Message
		assert.NotContains(t, capturedPost.Message, "Approve")
		assert.NotContains(t, capturedPost.Message, "Deny")
	})

	t.Run("buttons work with long descriptions", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Create description close to 1000 chars
		longDescription := strings.Repeat("This is a very detailed approval request description that spans multiple lines and contains important information. ", 10)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          longDescription,
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify buttons still present with long description
		attachments := capturedPost.Props["attachments"].([]any)
		assert.Len(t, attachments, 1)

		attachment := attachments[0].(map[string]any)
		actions := attachment["actions"].([]any)
		assert.Len(t, actions, 2)

		// Verify full description preserved in message
		assert.Contains(t, capturedPost.Message, longDescription)
	})

	t.Run("approval context is unique per approval", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var post1, post2 *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)

		// Capture first post
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			if post1 == nil {
				post1 = post
			} else {
				post2 = post
			}
			return true
		})).Return(&model.Post{Id: "post_123"}, nil).Twice()

		// Create two different approval records
		record1 := &approval.ApprovalRecord{
			ID:                   "record111",
			Code:                 "A-ABC123",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "First approval",
			CreatedAt:            1704988800000,
		}

		record2 := &approval.ApprovalRecord{
			ID:                   "record222",
			Code:                 "A-XYZ789",
			ApproverID:           approverID,
			RequesterUsername:    "bob",
			RequesterDisplayName: "Bob Smith",
			Description:          "Second approval",
			CreatedAt:            1704988900000,
		}

		// Send both notifications
		_, err1 := SendApprovalRequestDM(api, botUserID, record1)
		_, err2 := SendApprovalRequestDM(api, botUserID, record2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Extract approval IDs from integration context in both posts
		attachments1 := post1.Props["attachments"].([]any)
		attachment1 := attachments1[0].(map[string]any)
		actions1 := attachment1["actions"].([]any)
		approveBtn1 := actions1[0].(map[string]any)
		denyBtn1 := actions1[1].(map[string]any)

		attachments2 := post2.Props["attachments"].([]any)
		attachment2 := attachments2[0].(map[string]any)
		actions2 := attachment2["actions"].([]any)
		approveBtn2 := actions2[0].(map[string]any)
		denyBtn2 := actions2[1].(map[string]any)

		// Extract integration context
		integration1Approve := approveBtn1["integration"].(map[string]any)
		context1Approve := integration1Approve["context"].(map[string]any)
		integration1Deny := denyBtn1["integration"].(map[string]any)
		context1Deny := integration1Deny["context"].(map[string]any)

		integration2Approve := approveBtn2["integration"].(map[string]any)
		context2Approve := integration2Approve["context"].(map[string]any)
		integration2Deny := denyBtn2["integration"].(map[string]any)
		context2Deny := integration2Deny["context"].(map[string]any)

		// Verify approval IDs in context are different between approvals
		assert.NotEqual(t, context1Approve["approval_id"], context2Approve["approval_id"], "Approval IDs should be unique per approval")
		assert.NotEqual(t, context1Deny["approval_id"], context2Deny["approval_id"], "Approval IDs should be unique per approval")

		// Verify correct approval IDs
		assert.Equal(t, "record111", context1Approve["approval_id"])
		assert.Equal(t, "record111", context1Deny["approval_id"])
		assert.Equal(t, "record222", context2Approve["approval_id"])
		assert.Equal(t, "record222", context2Deny["approval_id"])

		// Verify actions are correct
		assert.Equal(t, "approve", context1Approve["action"])
		assert.Equal(t, "deny", context1Deny["action"])
		assert.Equal(t, "approve", context2Approve["action"])
		assert.Equal(t, "deny", context2Deny["action"])
	})
}

func TestSendOutcomeNotificationDM(t *testing.T) {
	t.Run("successful approved notification", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == botUserID &&
				post.ChannelId == dmChannelID &&
				strings.Contains(post.Message, "✅ **Approval Request Approved**") &&
				strings.Contains(post.Message, "@jordan (Jordan Lee)") &&
				strings.Contains(post.Message, "You may proceed with this action")
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Create test approval record (approved)
		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Deploy hotfix to production",
			Status:              approval.StatusApproved,
			DecisionComment:     "Approved. Proceed immediately.",
			DecidedAt:           1704988800000, // 2024-01-11 12:00:00 UTC
		}

		// Execute
		postID, err := SendOutcomeNotificationDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "post_123", postID)
		api.AssertExpectations(t)
	})

	t.Run("successful denied notification with comment", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == botUserID &&
				post.ChannelId == dmChannelID &&
				strings.Contains(post.Message, "❌ **Approval Request Denied**") &&
				strings.Contains(post.Message, "@jordan (Jordan Lee)") &&
				strings.Contains(post.Message, "This request has been denied") &&
				strings.Contains(post.Message, "Need VP approval for production changes")
		})).Return(&model.Post{Id: "post_456"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record456",
			Code:                "A-ABC123",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Deploy hotfix to production",
			Status:              approval.StatusDenied,
			DecisionComment:     "Need VP approval for production changes",
			DecidedAt:           1704988800000,
		}

		postID, err := SendOutcomeNotificationDM(api, botUserID, record)

		assert.NoError(t, err)
		assert.Equal(t, "post_456", postID)
		api.AssertExpectations(t)
	})

	t.Run("approved message format matches AC2 exactly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Deploy hotfix to production",
			Status:              approval.StatusApproved,
			DecisionComment:     "Approved. Proceed immediately.",
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify exact format from AC2
		assert.Contains(t, capturedMessage, "✅ **Approval Request Approved**")
		assert.Contains(t, capturedMessage, "**Approver:** @jordan (Jordan Lee)")
		assert.Contains(t, capturedMessage, "**Decision Time:**")
		assert.Contains(t, capturedMessage, "**Request ID:** `A-X7K9Q2`")
		assert.Contains(t, capturedMessage, "**Original Request:**")
		assert.Contains(t, capturedMessage, "> Deploy hotfix to production")
		assert.Contains(t, capturedMessage, "**Comment:**")
		assert.Contains(t, capturedMessage, "Approved. Proceed immediately.")
		assert.Contains(t, capturedMessage, "**Status:** You may proceed with this action.")
	})

	t.Run("denied message format matches AC3 exactly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record456",
			Code:                "A-ABC123",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Deploy hotfix to production",
			Status:              approval.StatusDenied,
			DecisionComment:     "Need VP approval",
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify exact format from AC3
		assert.Contains(t, capturedMessage, "❌ **Approval Request Denied**")
		assert.Contains(t, capturedMessage, "**Approver:** @jordan (Jordan Lee)")
		assert.Contains(t, capturedMessage, "**Decision Time:**")
		assert.Contains(t, capturedMessage, "**Request ID:** `A-ABC123`")
		assert.Contains(t, capturedMessage, "**Original Request:**")
		assert.Contains(t, capturedMessage, "> Deploy hotfix to production")
		assert.Contains(t, capturedMessage, "**Comment:**")
		assert.Contains(t, capturedMessage, "Need VP approval")
		assert.Contains(t, capturedMessage, "**Status:** This request has been denied.")
	})

	t.Run("notification with empty comment omits comment section", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Deploy hotfix",
			Status:              approval.StatusApproved,
			DecisionComment:     "", // Empty comment
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify comment section is omitted
		assert.NotContains(t, capturedMessage, "**Comment:**")
		// But still contains status
		assert.Contains(t, capturedMessage, "**Status:** You may proceed")
	})

	t.Run("notification with long description formats correctly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		longDescription := strings.Repeat("This is a very detailed description with multiple lines. ", 20)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         longDescription,
			Status:              approval.StatusApproved,
			DecisionComment:     "Approved",
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify description is quoted with >
		assert.Contains(t, capturedMessage, "> "+longDescription)
	})

	t.Run("timestamp format is YYYY-MM-DD HH:MM:SS UTC", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Test",
			Status:              approval.StatusApproved,
			DecidedAt:           1704988800000, // 2024-01-11 12:00:00 UTC
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify timestamp format
		expectedTime := time.UnixMilli(1704988800000).UTC()
		expectedTimestamp := expectedTime.Format("2006-01-02 15:04:05 MST")
		assert.Contains(t, capturedMessage, expectedTimestamp)
	})

	t.Run("DM channel creation failure", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(nil, &model.AppError{Message: "DMs disabled"})

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Test",
			Status:              approval.StatusApproved,
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel for requester")
		api.AssertExpectations(t)
	})

	t.Run("CreatePost failure", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "network error"})

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Test",
			Status:              approval.StatusApproved,
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send outcome DM to requester")
		api.AssertExpectations(t)
	})

	t.Run("bot user ID not available", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         "requester789",
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Test",
			Status:              approval.StatusApproved,
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, "", record)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot user ID not available")
	})

	t.Run("nil record validation", func(t *testing.T) {
		api := &plugintest.API{}

		_, err := SendOutcomeNotificationDM(api, "bot123", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record is nil")
	})

	t.Run("empty record ID validation", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:                  "", // Empty ID
			Code:                "A-X7K9Q2",
			RequesterID:         "requester789",
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Test",
			Status:              approval.StatusApproved,
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, "bot123", record)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record ID is empty")
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterID:         requesterID,
			ApproverUsername:    "jordan",
			ApproverDisplayName: "Jordan Lee",
			Description:         "Test",
			Status:              approval.StatusPending, // Invalid for outcome notification
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status for outcome notification")
		api.AssertExpectations(t)
	})
}

func TestUpdateApprovalPostForCancellation(t *testing.T) {
	t.Run("successful post update with cancellation", func(t *testing.T) {
		api := &plugintest.API{}

		// Create original post with buttons
		originalPost := &model.Post{
			Id:        "post_123",
			ChannelId: "dm_channel",
			Message:   "Original approval request message",
			Props: model.StringInterface{
				"attachments": []any{
					map[string]any{
						"actions": []any{
							map[string]any{"name": "Approve"},
							map[string]any{"name": "Deny"},
						},
					},
				},
			},
		}

		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.Id == "post_123" &&
				strings.Contains(post.Message, "🚫 **Approval Request (Canceled)**") &&
				strings.Contains(post.Message, "A-X7K9Q2") &&
				strings.Contains(post.Message, "Test description") &&
				strings.Contains(post.Message, "alice") &&
				len(post.Props) == 0 // Props cleared
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test description",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000, // 2024-01-11 12:00:00 UTC
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("post update shows description without strikethrough", func(t *testing.T) {
		api := &plugintest.API{}

		originalPost := &model.Post{
			Id:      "post_123",
			Message: "Original message",
			Props:   model.StringInterface{"attachments": []any{}},
		}

		var capturedMessage string
		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Deploy to production",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")
		assert.NoError(t, err)

		// Verify description shown without strikethrough
		assert.Contains(t, capturedMessage, "Deploy to production")
		assert.NotContains(t, capturedMessage, "~~Deploy to production~~")
		assert.Contains(t, capturedMessage, "**Description:**")
	})

	t.Run("post update removes action buttons", func(t *testing.T) {
		api := &plugintest.API{}

		// Original post with interactive buttons
		originalPost := &model.Post{
			Id:      "post_123",
			Message: "Original message",
			Props: model.StringInterface{
				"attachments": []any{
					map[string]any{
						"actions": []any{
							map[string]any{"name": "Approve", "style": "primary"},
							map[string]any{"name": "Deny", "style": "danger"},
						},
					},
				},
			},
		}

		var updatedPost *model.Post
		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
			updatedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")
		assert.NoError(t, err)

		// Verify Props are cleared (no buttons)
		assert.Empty(t, updatedPost.Props, "Props should be empty to remove buttons")
	})

	t.Run("timestamp formatted correctly", func(t *testing.T) {
		api := &plugintest.API{}

		originalPost := &model.Post{
			Id:      "post_123",
			Message: "Original message",
			Props:   model.StringInterface{},
		}

		var capturedMessage string
		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000, // 2024-01-11 12:00:00 UTC
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")
		assert.NoError(t, err)

		// Verify timestamp format: Jan 02, 2006 3:04 PM
		expectedTime := time.UnixMilli(1704988800000).UTC()
		expectedTimestamp := expectedTime.Format("Jan 02, 2006 3:04 PM")
		assert.Contains(t, capturedMessage, expectedTimestamp)
	})

	t.Run("shows who canceled the request", func(t *testing.T) {
		api := &plugintest.API{}

		originalPost := &model.Post{
			Id:      "post_123",
			Message: "Original message",
			Props:   model.StringInterface{},
		}

		var capturedMessage string
		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "bob")
		assert.NoError(t, err)

		// Verify canceler is shown
		assert.Contains(t, capturedMessage, "Canceled by @bob")
	})

	t.Run("empty post ID returns error", func(t *testing.T) {
		api := &plugintest.API{}

		api.On("LogWarn", "Cannot update approver post: no post ID stored", "request_id", "record123").Return()

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "", // Empty post ID
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no approver post ID found")
		api.AssertExpectations(t)
	})

	t.Run("post no longer exists", func(t *testing.T) {
		api := &plugintest.API{}

		api.On("GetPost", "post_123").Return(nil, &model.AppError{Message: "post not found"})
		api.On("LogError", "Failed to get post for update", "post_id", "post_123", "error", "post not found").Return()

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get post")
		api.AssertExpectations(t)
	})

	t.Run("UpdatePost fails", func(t *testing.T) {
		api := &plugintest.API{}

		originalPost := &model.Post{
			Id:      "post_123",
			Message: "Original message",
			Props:   model.StringInterface{},
		}

		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.Anything).Return(nil, &model.AppError{Message: "network error"})
		api.On("LogError", "Failed to update post", "post_id", "post_123", "error", "network error").Return()

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Test",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update post")
		api.AssertExpectations(t)
	})

	t.Run("nil record validation", func(t *testing.T) {
		api := &plugintest.API{}

		err := UpdateApprovalPostForCancellation(api, nil, "alice")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record is nil")
	})

	t.Run("message format includes all required fields", func(t *testing.T) {
		api := &plugintest.API{}

		originalPost := &model.Post{
			Id:      "post_123",
			Message: "Original message",
			Props:   model.StringInterface{},
		}

		var capturedMessage string
		api.On("GetPost", "post_123").Return(originalPost, nil)
		api.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			RequesterUsername:   "alice",
			Description:         "Deploy hotfix to production",
			Status:              approval.StatusCanceled,
			CanceledAt:          1704988800000,
			NotificationPostID:  "post_123",
		}

		err := UpdateApprovalPostForCancellation(api, record, "alice")
		assert.NoError(t, err)

		// Verify all required fields are present
		assert.Contains(t, capturedMessage, "🚫 **Approval Request (Canceled)**")
		assert.Contains(t, capturedMessage, "**Request ID:** `A-X7K9Q2`")
		assert.Contains(t, capturedMessage, "**From:** @alice")
		assert.Contains(t, capturedMessage, "**Description:**")
		assert.Contains(t, capturedMessage, "Deploy hotfix to production")
		assert.NotContains(t, capturedMessage, "~~Deploy hotfix to production~~")
		assert.Contains(t, capturedMessage, "---")
		assert.Contains(t, capturedMessage, "_Canceled by @alice at")
	})
}

func TestSendCancellationNotificationDM(t *testing.T) {
	t.Run("successful cancellation notification", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == botUserID &&
				post.ChannelId == dmChannelID &&
				strings.Contains(post.Message, "🚫 **Approval Request Canceled**") &&
				strings.Contains(post.Message, "@alice") &&
				strings.Contains(post.Message, "TUZ-2RK") &&
				strings.Contains(post.Message, "No longer needed")
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		// Create test approval record
		record := &approval.ApprovalRecord{
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			CanceledAt:        1736725200000, // Jan 12, 2026 7:15 PM UTC
			CanceledReason:    "No longer needed",
		}

		// Execute
		postID, err := SendCancellationNotificationDM(api, botUserID, record, "alice")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "notification_post_123", postID)
		api.AssertExpectations(t)
	})

	t.Run("empty bot user ID returns error", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			ApproverID: "approver456",
		}

		// Execute with empty botUserID
		_, err := SendCancellationNotificationDM(api, "", record, "alice")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot user ID not available")
	})

	t.Run("nil record returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		// Execute with nil record
		_, err := SendCancellationNotificationDM(api, botUserID, nil, "alice")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record is nil")
	})

	t.Run("empty approval ID returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		record := &approval.ApprovalRecord{
			ID:         "", // Empty ID
			ApproverID: "approver456",
		}

		// Execute with empty ID
		_, err := SendCancellationNotificationDM(api, botUserID, record, "alice")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record ID is empty")
	})

	t.Run("empty approver ID returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			ApproverID: "", // Empty approver ID
		}

		// Execute with empty approver ID
		_, err := SendCancellationNotificationDM(api, botUserID, record, "alice")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approver ID is empty")
	})

	t.Run("DM channel creation failure handled gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"

		// Mock GetDirectChannel to return error
		api.On("GetDirectChannel", botUserID, approverID).Return(nil, &model.AppError{Message: "DM channel disabled"})

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			Code:       "TUZ-2RK",
			ApproverID: approverID,
			CanceledAt: 1736725200000,
		}

		// Execute - should return error
		_, err := SendCancellationNotificationDM(api, botUserID, record, "alice")

		// Assert error is returned for caller to log
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel")
		api.AssertExpectations(t)
	})

	t.Run("CreatePost failure handled gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "network error"})

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			Code:       "TUZ-2RK",
			ApproverID: approverID,
			CanceledAt: 1736725200000,
		}

		// Execute - should return error
		_, err := SendCancellationNotificationDM(api, botUserID, record, "alice")

		// Assert error is returned for caller to log
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send cancellation notification")
		api.AssertExpectations(t)
	})

	t.Run("message format includes all required fields", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			CanceledAt:        1736725200000,
			CanceledReason:    "No longer needed",
		}

		postID, err := SendCancellationNotificationDM(api, botUserID, record, "alice")
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify all required fields are present
		assert.Contains(t, capturedMessage, "🚫 **Approval Request Canceled**")
		assert.Contains(t, capturedMessage, "**Reference:** `TUZ-2RK`")
		assert.Contains(t, capturedMessage, "**Requester:** @alice")
		assert.Contains(t, capturedMessage, "**Reason:** No longer needed")
		assert.Contains(t, capturedMessage, "**Canceled:**")
		assert.Contains(t, capturedMessage, "The approval request you received has been canceled by the requester.")
	})

	t.Run("timestamp formatted as Jan 02, 2006 3:04 PM", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Use specific timestamp: Jan 12, 2026 7:15 PM UTC
		record := &approval.ApprovalRecord{
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			CanceledAt:        1736714100000, // Jan 12, 2026 7:15 PM UTC
		}

		postID, err := SendCancellationNotificationDM(api, botUserID, record, "alice")
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify timestamp format: "Jan 02, 2006 3:04 PM"
		expectedTime := time.UnixMilli(1736714100000).UTC()
		expectedTimestamp := expectedTime.Format("Jan 02, 2006 3:04 PM")
		assert.Contains(t, capturedMessage, expectedTimestamp)
	})

	t.Run("cancellation reason handles empty string gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			CanceledAt:        1736725200000,
			CanceledReason:    "", // Empty reason
		}

		postID, err := SendCancellationNotificationDM(api, botUserID, record, "alice")
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify "Not specified" is used when reason is empty
		assert.Contains(t, capturedMessage, "**Reason:** Not specified")
	})

	t.Run("username display includes @ symbol", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "bob.jones",
			CanceledAt:        1736725200000,
			CanceledReason:    "Changed plans",
		}

		postID, err := SendCancellationNotificationDM(api, botUserID, record, "bob.jones")
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify @ symbol is included with username
		assert.Contains(t, capturedMessage, "**Requester:** @bob.jones")
	})
}

// Helper function to verify the plugin.API interface is satisfied
var _ plugin.API = (*plugintest.API)(nil)
