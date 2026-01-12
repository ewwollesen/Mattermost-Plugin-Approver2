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
		})).Return(&model.Post{}, nil)

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
		err := SendApprovalRequestDM(api, botUserID, record)

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
		})).Return(&model.Post{}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy the hotfix to production environment",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		err := SendApprovalRequestDM(api, botUserID, record)
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
		})).Return(&model.Post{}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Test description",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		err := SendApprovalRequestDM(api, botUserID, record)
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
		err := SendApprovalRequestDM(api, botUserID, record)

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
		err := SendApprovalRequestDM(api, botUserID, record)

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
		err := SendApprovalRequestDM(api, "", record)

		// Assert error is returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot user ID not available")
		api.AssertExpectations(t)
	})

	t.Run("nil record validation", func(t *testing.T) {
		api := &plugintest.API{}

		// Execute - should return error for nil record
		err := SendApprovalRequestDM(api, "bot123", nil)

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
		err := SendApprovalRequestDM(api, "bot123", record)

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
		})).Return(&model.Post{}, nil)

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
		err := SendApprovalRequestDM(api, botUserID, record)

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
		})).Return(&model.Post{}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Extract approve button
		attachments := capturedPost.Props["attachments"].([]any)
		attachment := attachments[0].(map[string]any)
		actions := attachment["actions"].([]any)
		approveButton := actions[0].(map[string]any)

		// Verify approve button properties
		assert.Equal(t, "approve_record123", approveButton["id"])
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
		})).Return(&model.Post{}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Extract deny button
		attachments := capturedPost.Props["attachments"].([]any)
		attachment := attachments[0].(map[string]any)
		actions := attachment["actions"].([]any)
		denyButton := actions[1].(map[string]any)

		// Verify deny button properties
		assert.Equal(t, "deny_record123", denyButton["id"])
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
		})).Return(&model.Post{}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy the hotfix to production",
			CreatedAt:            1704988800000,
		}

		err := SendApprovalRequestDM(api, botUserID, record)
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
		})).Return(&model.Post{}, nil)

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

		err := SendApprovalRequestDM(api, botUserID, record)
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

	t.Run("button IDs are unique per approval", func(t *testing.T) {
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
		})).Return(&model.Post{}, nil).Twice()

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
		err1 := SendApprovalRequestDM(api, botUserID, record1)
		err2 := SendApprovalRequestDM(api, botUserID, record2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Extract button IDs from both posts
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

		// Verify button IDs are different between approvals
		assert.NotEqual(t, approveBtn1["id"], approveBtn2["id"], "Approve button IDs should be unique per approval")
		assert.NotEqual(t, denyBtn1["id"], denyBtn2["id"], "Deny button IDs should be unique per approval")

		// Verify correct format
		assert.Equal(t, "approve_record111", approveBtn1["id"])
		assert.Equal(t, "deny_record111", denyBtn1["id"])
		assert.Equal(t, "approve_record222", approveBtn2["id"])
		assert.Equal(t, "deny_record222", denyBtn2["id"])
	})
}

// Helper function to verify the plugin.API interface is satisfied
var _ plugin.API = (*plugintest.API)(nil)
