package notifications

import (
	"net/http"
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
			Status:               approval.StatusPending,
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
			Status:               approval.StatusPending,
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
			Status:               approval.StatusPending,
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
			Status:               approval.StatusPending,
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
			Status:               approval.StatusPending,
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

// TestSendApprovalRequestDM_MatterpollPattern tests the Matterpoll pattern implementation (Story 10.3)
func TestSendApprovalRequestDM_MatterpollPattern(t *testing.T) {
	t.Run("uses custom_approval_dm post type", func(t *testing.T) {
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

		// Create test approval record with Status for button rendering
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
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

		// Story 10.3 AC1: Verify custom post type
		assert.Equal(t, CustomApprovalDMPostType, capturedPost.Type, "should use custom_approval_dm post type")

		// Story 10.3 AC1: Verify notification_type prop
		assert.Equal(t, NotificationTypeApprovalRequest, capturedPost.Props["notification_type"], "notification_type should be approval_request")

		// Verify is_dm prop
		assert.Equal(t, true, capturedPost.Props["is_dm"], "is_dm should be true")
	})

	t.Run("includes approve and deny buttons with new URL pattern", func(t *testing.T) {
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
			Status:               approval.StatusPending,
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

		// Verify attachments exist (from ParseSlackAttachment)
		attachmentsRaw, ok := capturedPost.Props["attachments"]
		assert.True(t, ok, "Props.attachments should exist")

		// ParseSlackAttachment returns []*model.SlackAttachment
		attachments, ok := attachmentsRaw.([]*model.SlackAttachment)
		assert.True(t, ok, "attachments should be []*model.SlackAttachment")
		assert.Len(t, attachments, 1, "should have 1 attachment")

		// Verify actions array exists with 2 buttons
		assert.Len(t, attachments[0].Actions, 2, "should have 2 buttons")
	})

	t.Run("approve button uses new URL pattern with code", func(t *testing.T) {
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
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Extract approve button
		attachments := capturedPost.Props["attachments"].([]*model.SlackAttachment)
		approveAction := attachments[0].Actions[0]

		// Verify approve button properties
		assert.Equal(t, "Approve", approveAction.Name)
		assert.Equal(t, "success", approveAction.Style)

		// Story 10.3: Verify new URL pattern with code in path (not context map)
		expectedURL := "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-X7K9Q2/approve"
		assert.Equal(t, expectedURL, approveAction.Integration.URL, "approve button should use new URL pattern")
	})

	t.Run("deny button uses new URL pattern with code", func(t *testing.T) {
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
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Extract deny button
		attachments := capturedPost.Props["attachments"].([]*model.SlackAttachment)
		denyAction := attachments[0].Actions[1]

		// Verify deny button properties
		assert.Equal(t, "Deny", denyAction.Name)
		assert.Equal(t, "danger", denyAction.Style)

		// Story 10.3: Verify new URL pattern with code in path
		expectedURL := "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-X7K9Q2/deny"
		assert.Equal(t, expectedURL, denyAction.Integration.URL, "deny button should use new URL pattern")
	})

	t.Run("includes approval props for webapp component", func(t *testing.T) {
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
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy hotfix",
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify approval props for webapp component (AC2)
		assert.Equal(t, "A-X7K9Q2", capturedPost.Props["approval_code"])
		assert.Equal(t, approval.StatusPending, capturedPost.Props["approval_status"])
		assert.Equal(t, "alice", capturedPost.Props["requester_username"])
		assert.Equal(t, "Alice Carter", capturedPost.Props["requester_display_name"])
		assert.Equal(t, "bob", capturedPost.Props["approver_username"])
		assert.Equal(t, "Bob Smith", capturedPost.Props["approver_display_name"])
		assert.Equal(t, "Deploy hotfix", capturedPost.Props["description"])
		assert.Equal(t, int64(1704988800000), capturedPost.Props["created_at"])
	})

	t.Run("markdown fallback in message for non-webapp clients", func(t *testing.T) {
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
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Deploy the hotfix to production",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Story 10.3 AC4: Verify markdown fallback for non-webapp clients
		assert.Contains(t, capturedPost.Message, "📋 **Approval Request**")
		assert.Contains(t, capturedPost.Message, "@alice (Alice Carter)")
		assert.Contains(t, capturedPost.Message, "Deploy the hotfix to production")
		assert.Contains(t, capturedPost.Message, "A-X7K9Q2")
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
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          longDescription,
			CreatedAt:            1704988800000,
		}

		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify buttons still present with long description (using new SlackAttachment format)
		attachments := capturedPost.Props["attachments"].([]*model.SlackAttachment)
		assert.Len(t, attachments, 1)
		assert.Len(t, attachments[0].Actions, 2, "should have 2 buttons with long description")

		// Verify full description preserved in message
		assert.Contains(t, capturedPost.Message, longDescription, "full description should be in message")
	})

	t.Run("approval code is unique per approval in button URLs", func(t *testing.T) {
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
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "First approval",
			CreatedAt:            1704988800000,
		}

		record2 := &approval.ApprovalRecord{
			ID:                   "record222",
			Code:                 "A-XYZ789",
			Status:               approval.StatusPending,
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

		// Story 10.3: Verify URLs use unique codes in path (not context maps)
		attachments1 := post1.Props["attachments"].([]*model.SlackAttachment)
		attachments2 := post2.Props["attachments"].([]*model.SlackAttachment)

		// Verify URLs are different (contain different codes)
		approve1URL := attachments1[0].Actions[0].Integration.URL
		deny1URL := attachments1[0].Actions[1].Integration.URL
		approve2URL := attachments2[0].Actions[0].Integration.URL
		deny2URL := attachments2[0].Actions[1].Integration.URL

		assert.NotEqual(t, approve1URL, approve2URL, "Approval URLs should be unique per approval")
		assert.NotEqual(t, deny1URL, deny2URL, "Deny URLs should be unique per approval")

		// Verify correct codes in URLs
		assert.Contains(t, approve1URL, "A-ABC123", "first approval URL should contain first code")
		assert.Contains(t, deny1URL, "A-ABC123", "first deny URL should contain first code")
		assert.Contains(t, approve2URL, "A-XYZ789", "second approval URL should contain second code")
		assert.Contains(t, deny2URL, "A-XYZ789", "second deny URL should contain second code")

		// Verify URL pattern
		assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-ABC123/approve", approve1URL)
		assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-ABC123/deny", deny1URL)
		assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-XYZ789/approve", approve2URL)
		assert.Equal(t, "/plugins/com.mattermost.plugin-approver2/api/v1/approval/A-XYZ789/deny", deny2URL)
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

		// Note: No mock expectations - validation fails before any API calls
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

	// Story 10.5: Matterpoll pattern tests
	t.Run("uses_custom_approval_dm_post_type_for_webapp_rendering", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-TEST01",
			RequesterID:         requesterID,
			ApproverUsername:    "approver",
			ApproverDisplayName: "Approver User",
			Description:         "Test outcome",
			Status:              approval.StatusApproved,
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC1: Verify custom post type for webapp rendering
		assert.Equal(t, "custom_approval_dm", capturedPost.Type)

		// AC1: Verify notification_type is "outcome"
		notifType, ok := capturedPost.Props["notification_type"].(string)
		assert.True(t, ok)
		assert.Equal(t, "outcome", notifType)

		// AC1: Verify decided_at timestamp prop
		decidedAt, ok := capturedPost.Props["decided_at"].(int64)
		assert.True(t, ok)
		assert.Equal(t, int64(1704988800000), decidedAt)

		// M1 Fix: Verify approval_status prop is correctly set
		approvalStatus, ok := capturedPost.Props["approval_status"].(string)
		assert.True(t, ok)
		assert.Equal(t, approval.StatusApproved, approvalStatus)

		// AC3: Verify no buttons (attachments should have empty actions)
		attachments, ok := capturedPost.Props["attachments"].([]*model.SlackAttachment)
		assert.True(t, ok)
		assert.Len(t, attachments, 1)
		assert.Empty(t, attachments[0].Actions) // No buttons for outcome

		api.AssertExpectations(t)
	})

	t.Run("includes_decision_comment_in_props_when_present", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester789"
		dmChannelID := "dm456"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-TEST01",
			RequesterID:         requesterID,
			ApproverUsername:    "approver",
			ApproverDisplayName: "Approver User",
			Description:         "Test outcome",
			Status:              approval.StatusDenied,
			DecisionComment:     "Need manager approval first",
			DecidedAt:           1704988800000,
		}

		_, err := SendOutcomeNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC1: Verify decision_comment prop
		comment, ok := capturedPost.Props["decision_comment"].(string)
		assert.True(t, ok)
		assert.Equal(t, "Need manager approval first", comment)

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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test description",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000, // 2024-01-11 12:00:00 UTC
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Deploy to production",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000, // 2024-01-11 12:00:00 UTC
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "", // Empty post ID
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Test",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "post_123",
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
			ID:                 "record123",
			Code:               "A-X7K9Q2",
			RequesterUsername:  "alice",
			Description:        "Deploy hotfix to production",
			Status:             approval.StatusCanceled,
			CanceledAt:         1704988800000,
			NotificationPostID: "post_123",
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

		// Verify markdown fallback includes all required fields (Story 10.6: FormatMarkdownFallback format)
		assert.Contains(t, capturedMessage, "🚫 **Approval Request Canceled**")
		assert.Contains(t, capturedMessage, "**Reference:** `TUZ-2RK`")
		assert.Contains(t, capturedMessage, "**Requester:** @alice")
		assert.Contains(t, capturedMessage, "**Reason:** No longer needed")
		assert.Contains(t, capturedMessage, "**Canceled:**")
		// Note: Story 10.6 uses FormatMarkdownFallback which doesn't include the trailing explanation
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

	// Story 10.6: Matterpoll pattern tests
	t.Run("uses Matterpoll pattern with custom post type", func(t *testing.T) {
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
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			Status:            approval.StatusCanceled,
			CanceledAt:        1736725200000,
			CanceledReason:    "No longer needed",
		}

		postID, err := SendCancellationNotificationDM(api, botUserID, record, "alice")
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify Matterpoll pattern: custom post type
		assert.Equal(t, CustomApprovalDMPostType, capturedPost.Type)

		// Verify notification_type is set correctly
		assert.Equal(t, NotificationTypeCancellation, capturedPost.Props["notification_type"])

		// Verify approval_status prop is set
		assert.Equal(t, approval.StatusCanceled, capturedPost.Props["approval_status"])
	})

	t.Run("includes canceled_at and canceled_reason in props", func(t *testing.T) {
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
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			Status:            approval.StatusCanceled,
			CanceledAt:        1736725200000,
			CanceledReason:    "No longer needed",
		}

		postID, err := SendCancellationNotificationDM(api, botUserID, record, "alice")
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify canceled_at and canceled_reason props for webapp component
		assert.Equal(t, int64(1736725200000), capturedPost.Props["canceled_at"])
		assert.Equal(t, "No longer needed", capturedPost.Props["canceled_reason"])
	})

	t.Run("does not include buttons for cancellation notifications", func(t *testing.T) {
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
			ID:                "approval123",
			Code:              "TUZ-2RK",
			ApproverID:        approverID,
			RequesterUsername: "alice",
			Status:            approval.StatusCanceled,
			CanceledAt:        1736725200000,
			CanceledReason:    "No longer needed",
		}

		_, err := SendCancellationNotificationDM(api, botUserID, record, "alice")
		assert.NoError(t, err)

		// Verify no buttons in attachments (cancellation is read-only)
		attachments, ok := capturedPost.Props["attachments"].([]*model.SlackAttachment)
		if ok && len(attachments) > 0 {
			assert.Empty(t, attachments[0].Actions, "Cancellation notifications should not have buttons")
		}
	})
}

// TestSendVerificationNotificationDM tests the Story 10.8 Matterpoll pattern implementation
// for verification notifications sent to approvers when the requester marks an approved request as verified.
func TestSendVerificationNotificationDM(t *testing.T) {
	t.Run("successful verification notification with Matterpoll pattern", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == botUserID &&
				post.ChannelId == dmChannelID &&
				post.Type == CustomApprovalDMPostType &&
				strings.Contains(post.Message, "✅ **Approval Request Verified**")
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		// Create test approval record
		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           1704931400000, // Jan 10, 2024
			VerificationComment:  "",
		}

		// Execute
		postID, err := SendVerificationNotificationDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "notification_post_123", postID)
		api.AssertExpectations(t)
	})

	t.Run("post type is custom_approval_dm (AC1)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           1704931400000,
		}

		_, err := SendVerificationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC1: Verify post type is custom_approval_dm
		assert.Equal(t, CustomApprovalDMPostType, capturedPost.Type)
		api.AssertExpectations(t)
	})

	t.Run("notification_type is verification (AC2)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           1704931400000,
		}

		_, err := SendVerificationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC2: Verify notification_type is "verification"
		assert.Equal(t, NotificationTypeVerification, capturedPost.GetProp("notification_type"))
		api.AssertExpectations(t)
	})

	t.Run("verified_at prop is populated (AC3)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"
		verifiedAt := int64(1704931400000) // Jan 10, 2024

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           verifiedAt,
		}

		_, err := SendVerificationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC3: Verify verified_at prop is populated with the timestamp
		assert.Equal(t, verifiedAt, capturedPost.GetProp("verified_at"))
		api.AssertExpectations(t)
	})

	t.Run("verification_comment prop when present (AC4)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           1704931400000,
			VerificationComment:  "Deployment completed successfully",
		}

		_, err := SendVerificationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC4: Verify verification_comment prop is populated
		assert.Equal(t, "Deployment completed successfully", capturedPost.GetProp("verification_comment"))
		// Also verify it's in the markdown fallback
		assert.Contains(t, capturedPost.Message, "Deployment completed successfully")
		api.AssertExpectations(t)
	})

	t.Run("no buttons in attachment (AC5)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           1704931400000,
			// VerificationComment intentionally empty
		}

		_, err := SendVerificationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC5: Verify no buttons in attachment - verification is read-only
		attachments := capturedPost.GetProp("attachments")
		if attachments != nil {
			attachmentSlice, ok := attachments.([]*model.SlackAttachment)
			if ok && len(attachmentSlice) > 0 {
				assert.Nil(t, attachmentSlice[0].Actions, "verification notifications should have no buttons")
			}
		}

		// M4 Fix: Verify empty verification_comment omits "Verification Note" from markdown fallback
		assert.NotContains(t, capturedPost.Message, "**Verification Note:**", "empty verification comment should not include Verification Note section")
		api.AssertExpectations(t)
	})

	t.Run("markdown fallback includes verification details (AC6)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "approval123",
			Code:                 "A-X7K9Q2",
			ApproverID:           approverID,
			RequesterID:          "requester123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusApproved,
			Verified:             true,
			VerifiedAt:           1704931400000,
			VerificationComment:  "Task completed",
		}

		_, err := SendVerificationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC6: Verify markdown fallback includes all verification details
		assert.Contains(t, capturedPost.Message, "✅ **Approval Request Verified**")
		assert.Contains(t, capturedPost.Message, "A-X7K9Q2")
		assert.Contains(t, capturedPost.Message, "@alice")
		assert.Contains(t, capturedPost.Message, "Deploy to production")
		assert.Contains(t, capturedPost.Message, "Task completed")
		api.AssertExpectations(t)
	})

	t.Run("empty bot user ID returns error", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			ApproverID: "approver456",
			Verified:   true,
		}

		// Execute with empty botUserID
		_, err := SendVerificationNotificationDM(api, "", record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot user ID not available")
	})

	t.Run("nil record returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		// Execute with nil record
		_, err := SendVerificationNotificationDM(api, botUserID, nil)

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
		_, err := SendVerificationNotificationDM(api, botUserID, record)

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
		_, err := SendVerificationNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approver ID is empty")
	})

	t.Run("get DM channel failure returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"

		// Mock GetDirectChannel to return error
		api.On("GetDirectChannel", botUserID, approverID).Return(nil, &model.AppError{Message: "channel error"})

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			ApproverID: approverID,
		}

		// Execute
		_, err := SendVerificationNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel")
		api.AssertExpectations(t)
	})

	t.Run("create post failure returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, &model.AppError{Message: "post error"})

		record := &approval.ApprovalRecord{
			ID:         "approval123",
			ApproverID: approverID,
			Verified:   true,
			VerifiedAt: 1704931400000,
		}

		// Execute
		_, err := SendVerificationNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send verification notification")
		api.AssertExpectations(t)
	})
}

func TestSendRequesterCancellationNotificationDM(t *testing.T) {
	t.Run("successful notification to requestor", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)

		var capturedMessage string
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			if post.UserId == botUserID && post.ChannelId == dmChannelID {
				capturedMessage = post.Message
				return strings.Contains(post.Message, "🚫 **Your Approval Request Was Canceled**") &&
					strings.Contains(post.Message, "A-X7K9Q2") &&
					strings.Contains(post.Message, "@janedoe") &&
					strings.Contains(post.Message, "No longer needed") &&
					strings.Contains(post.Message, "Deploy to production")
			}
			return false
		})).Return(&model.Post{Id: "notification_post_123"}, nil)

		// Create test approval record
		record := &approval.ApprovalRecord{
			ID:                "record123",
			Code:              "A-X7K9Q2",
			Description:       "Deploy to production",
			RequesterID:       requesterID,
			RequesterUsername: "johndoe",
			ApproverID:        "approver456",
			ApproverUsername:  "janedoe",
			Status:            approval.StatusCanceled,
			CanceledReason:    "No longer needed",
			CanceledAt:        1704931300000, // Jan 10, 2024 12:15 PM UTC
		}

		// Execute
		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "notification_post_123", postID)
		// Timestamp should be formatted in UTC with full format
		assert.Contains(t, capturedMessage, "Jan 11, 2024")
		assert.Contains(t, capturedMessage, "12:01 AM")
		api.AssertExpectations(t)
	})

	t.Run("DM channel creation fails", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"

		api.On("GetDirectChannel", botUserID, requesterID).Return(
			nil,
			model.NewAppError("GetDirectChannel", "api.channel.create_direct_channel.internal_error", nil, "", http.StatusInternalServerError),
		)

		record := &approval.ApprovalRecord{
			ID:          "record123",
			RequesterID: requesterID,
		}

		// Execute
		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, postID)
		assert.Contains(t, err.Error(), "failed to get DM channel")
		api.AssertExpectations(t)
	})

	t.Run("post creation fails", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.Anything).Return(
			nil,
			model.NewAppError("CreatePost", "api.post.create_post.internal_error", nil, "", http.StatusInternalServerError),
		)

		record := &approval.ApprovalRecord{
			ID:               "record123",
			Code:             "A-X7K9Q2",
			Description:      "Test request",
			RequesterID:      requesterID,
			ApproverUsername: "approver",
			CanceledReason:   "Reason",
			CanceledAt:       1704931300000,
		}

		// Execute
		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, postID)
		assert.Contains(t, err.Error(), "failed to send cancellation notification to requestor")
		api.AssertExpectations(t)
	})

	t.Run("empty bot user ID", func(t *testing.T) {
		api := &plugintest.API{}
		record := &approval.ApprovalRecord{
			ID:          "record123",
			RequesterID: "user123",
		}

		// Execute
		postID, err := SendRequesterCancellationNotificationDM(api, "", record)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, postID)
		assert.Contains(t, err.Error(), "bot user ID not available")
	})

	t.Run("nil record", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		// Execute
		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, nil)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, postID)
		assert.Contains(t, err.Error(), "approval record is nil")
	})

	t.Run("all cancellation reasons", func(t *testing.T) {
		reasons := []string{
			"Approved/denied elsewhere",
			"Wrong approver selected",
			"Mistake - no longer needed",
			"Sensitive information - needs private discussion",
			"Duplicate request",
			"Other: Changed requirements",
		}

		for _, reason := range reasons {
			t.Run(reason, func(t *testing.T) {
				api := &plugintest.API{}
				botUserID := "bot123"
				requesterID := "user123"
				dmChannelID := "dm789"

				api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
				api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
					return strings.Contains(post.Message, reason)
				})).Return(&model.Post{Id: "post123"}, nil)

				record := &approval.ApprovalRecord{
					ID:               "record123",
					Code:             "A-X7K9Q2",
					Description:      "Test",
					RequesterID:      requesterID,
					ApproverUsername: "approver",
					CanceledReason:   reason,
					CanceledAt:       1704931300000,
				}

				postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
				assert.NoError(t, err)
				assert.NotEmpty(t, postID)
				api.AssertExpectations(t)
			})
		}
	})

	t.Run("zero CanceledAt timestamp", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// Should format zero timestamp as Jan 01, 1970
			return strings.Contains(post.Message, "Jan 01, 1970")
		})).Return(&model.Post{Id: "post123"}, nil)

		record := &approval.ApprovalRecord{
			ID:               "record123",
			Code:             "A-X7K9Q2",
			Description:      "Test",
			RequesterID:      requesterID,
			ApproverUsername: "approver",
			CanceledReason:   "Reason",
			CanceledAt:       0, // Zero timestamp
		}

		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)
		api.AssertExpectations(t)
	})

	t.Run("empty record ID", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		record := &approval.ApprovalRecord{
			ID:          "", // Empty ID
			RequesterID: "user123",
		}

		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.Error(t, err)
		assert.Empty(t, postID)
		assert.Contains(t, err.Error(), "approval record ID is empty")
	})

	t.Run("empty requester ID", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		record := &approval.ApprovalRecord{
			ID:          "record123",
			RequesterID: "", // Empty requester ID
		}

		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.Error(t, err)
		assert.Empty(t, postID)
		assert.Contains(t, err.Error(), "requester ID is empty")
	})

	t.Run("empty optional fields still create valid message", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// Message should still be created even with empty optional fields
			return post.ChannelId == dmChannelID && post.UserId == botUserID
		})).Return(&model.Post{Id: "post123"}, nil)

		record := &approval.ApprovalRecord{
			ID:               "record123",
			Code:             "", // Empty optional
			Description:      "", // Empty optional
			RequesterID:      requesterID,
			ApproverUsername: "", // Empty optional
			CanceledReason:   "", // Empty optional
			CanceledAt:       1704931300000,
		}

		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)
		api.AssertExpectations(t)
	})

	// Story 10.6: Matterpoll pattern tests
	t.Run("uses Matterpoll pattern with custom post type", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                  "record123",
			Code:                "A-X7K9Q2",
			Description:         "Deploy to production",
			RequesterID:         requesterID,
			RequesterUsername:   "johndoe",
			ApproverID:          "approver456",
			ApproverUsername:    "janedoe",
			ApproverDisplayName: "Jane Doe",
			Status:              approval.StatusCanceled,
			CanceledReason:      "No longer needed",
			CanceledAt:          1704931300000,
		}

		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify Matterpoll pattern: custom post type
		assert.Equal(t, CustomApprovalDMPostType, capturedPost.Type)

		// Verify notification_type is set correctly
		assert.Equal(t, NotificationTypeRequesterCancellation, capturedPost.Props["notification_type"])

		// Verify approval_status prop is set
		assert.Equal(t, approval.StatusCanceled, capturedPost.Props["approval_status"])
	})

	t.Run("includes canceled_at and canceled_reason in props", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                "record123",
			Code:              "A-X7K9Q2",
			Description:       "Deploy to production",
			RequesterID:       requesterID,
			RequesterUsername: "johndoe",
			ApproverUsername:  "janedoe",
			Status:            approval.StatusCanceled,
			CanceledReason:    "No longer needed",
			CanceledAt:        1704931300000,
		}

		postID, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)
		assert.NotEmpty(t, postID)

		// Verify canceled_at and canceled_reason props for webapp component
		assert.Equal(t, int64(1704931300000), capturedPost.Props["canceled_at"])
		assert.Equal(t, "No longer needed", capturedPost.Props["canceled_reason"])
	})

	t.Run("does not include buttons for requester cancellation notifications", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "user123"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                "record123",
			Code:              "A-X7K9Q2",
			Description:       "Deploy to production",
			RequesterID:       requesterID,
			RequesterUsername: "johndoe",
			ApproverUsername:  "janedoe",
			Status:            approval.StatusCanceled,
			CanceledReason:    "No longer needed",
			CanceledAt:        1704931300000,
		}

		_, err := SendRequesterCancellationNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify no buttons in attachments (cancellation is read-only)
		attachments, ok := capturedPost.Props["attachments"].([]*model.SlackAttachment)
		if ok && len(attachments) > 0 {
			assert.Empty(t, attachments[0].Actions, "Requester cancellation notifications should not have buttons")
		}
	})
}

// Story 8.4: Tests for formatPlaybookContext
func TestFormatPlaybookContext(t *testing.T) {
	t.Run("formats playbook context with all fields", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channel456").Return(&model.Channel{
			Id:   "channel456",
			Name: "deploy-prod-v2-1-0",
		}, nil)

		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run123",
			PlaybookName:      "Deploy - Production Release v2.1.0",
			PlaybookChannelID: "channel456",
		}

		result := formatPlaybookContext(api, record)

		// Verify all components present (AC2, AC3)
		assert.Contains(t, result, "**Playbook Context:**")
		assert.Contains(t, result, "- Playbook: Deploy - Production Release v2.1.0")
		assert.Contains(t, result, "- Channel: ~deploy-prod-v2-1-0")
		api.AssertExpectations(t)
	})

	t.Run("truncates long playbook names", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channel456").Return(&model.Channel{
			Id:   "channel456",
			Name: "test-channel",
		}, nil)

		longName := "This is a very long playbook name that exceeds the fifty character limit and needs truncation"
		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run123",
			PlaybookName:      longName,
			PlaybookChannelID: "channel456",
		}

		result := formatPlaybookContext(api, record)

		// Verify truncation (AC8)
		assert.Contains(t, result, "**Playbook Context:**")
		assert.NotContains(t, result, longName, "Should not contain full long name")
		assert.Contains(t, result, "...", "Should end with ellipsis")

		// Verify the playbook name part is truncated to ≤50 chars
		// The format is "- Playbook: NAME\n", so find and extract the name
		startMarker := "- Playbook: "
		startIdx := strings.Index(result, startMarker)
		assert.NotEqual(t, -1, startIdx, "Should contain playbook marker")

		// Extract from marker to next newline
		nameStart := startIdx + len(startMarker)
		endIdx := strings.Index(result[nameStart:], "\n")
		if endIdx != -1 {
			playbookName := result[nameStart : nameStart+endIdx]
			assert.LessOrEqual(t, len(playbookName), 50, "Truncated name should be ≤50 chars")
		}
	})

	t.Run("handles exactly 50 character playbook name", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channel456").Return(&model.Channel{
			Id:   "channel456",
			Name: "test-channel",
		}, nil)

		exactName := strings.Repeat("a", 50)
		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run123",
			PlaybookName:      exactName,
			PlaybookChannelID: "channel456",
		}

		result := formatPlaybookContext(api, record)

		// Should not be truncated at exactly 50 chars
		assert.Contains(t, result, exactName)
		assert.NotContains(t, result, "...")
		api.AssertExpectations(t)
	})

	t.Run("returns empty string when no playbook run ID", func(t *testing.T) {
		api := &plugintest.API{}
		record := &approval.ApprovalRecord{
			PlaybookRunID:     "",
			PlaybookName:      "Some Playbook",
			PlaybookChannelID: "channel456",
		}

		result := formatPlaybookContext(api, record)

		// Should return empty string (AC4)
		assert.Empty(t, result)
	})

	t.Run("returns empty string when no playbook name", func(t *testing.T) {
		api := &plugintest.API{}
		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run123",
			PlaybookName:      "",
			PlaybookChannelID: "channel456",
		}

		result := formatPlaybookContext(api, record)

		// Should return empty string (AC4)
		assert.Empty(t, result)
	})

	t.Run("formats channel link with channel name", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channelid123").Return(&model.Channel{
			Id:   "channelid123",
			Name: "incident-47",
		}, nil)

		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run123",
			PlaybookName:      "Test Playbook",
			PlaybookChannelID: "channelid123",
		}

		result := formatPlaybookContext(api, record)

		// Verify channel link format uses channel name, not ID (AC3)
		assert.Contains(t, result, "~incident-47")
		assert.NotContains(t, result, "channelid123")
		api.AssertExpectations(t)
	})

	t.Run("falls back to channel ID when GetChannel fails", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channel456").Return(nil, &model.AppError{Message: "not found"})
		api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run123",
			PlaybookName:      "Test Playbook",
			PlaybookChannelID: "channel456",
		}

		result := formatPlaybookContext(api, record)

		// Should fallback to ID if channel lookup fails
		assert.Contains(t, result, "~channel456")
		api.AssertExpectations(t)
	})

	t.Run("truncates UTF-8 multibyte characters correctly", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channel789").Return(&model.Channel{
			Id:   "channel789",
			Name: "deploy-prod",
		}, nil)

		// 51 runes total - emoji 🚀 takes 4 bytes but counts as 1 rune
		// Should truncate to 47 runes + "..." = 50 character limit
		longNameWithEmoji := "Deploy 🚀 Production Release v2.1.0 for Customer ABC"

		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run789",
			PlaybookName:      longNameWithEmoji,
			PlaybookChannelID: "channel789",
		}

		result := formatPlaybookContext(api, record)

		// Should truncate at 47 runes + "..." without corrupting UTF-8
		assert.Contains(t, result, "Deploy 🚀 Production Release v2.1.0 for Customer...")
		// Verify the emoji is intact (not corrupted)
		assert.Contains(t, result, "🚀")
		api.AssertExpectations(t)
	})

	t.Run("falls back to channel ID when channel name is empty", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetChannel", "channel999").Return(&model.Channel{
			Id:   "channel999",
			Name: "", // Empty name
		}, nil)
		api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		record := &approval.ApprovalRecord{
			PlaybookRunID:     "run999",
			PlaybookName:      "Test Playbook",
			PlaybookChannelID: "channel999",
		}

		result := formatPlaybookContext(api, record)

		// Should fallback to ID when channel name is empty
		assert.Contains(t, result, "~channel999")
		api.AssertExpectations(t)
	})
}

// Story 8.4: Integration tests for SendApprovalRequestDM with playbook context
func TestSendApprovalRequestDM_PlaybookContext(t *testing.T) {
	t.Run("includes playbook context when present", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("GetChannel", "incident-channel-123").Return(&model.Channel{
			Id:   "incident-channel-123",
			Name: "incident-47-deploy",
		}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// Verify standard fields
			if post.UserId != botUserID || post.ChannelId != dmChannelID {
				return false
			}

			// Verify playbook context is included (AC1, AC5)
			return strings.Contains(post.Message, "**Playbook Context:**") &&
				strings.Contains(post.Message, "Incident #47") &&
				strings.Contains(post.Message, "~incident-47-deploy") &&
				strings.Contains(post.Message, "📋 **Approval Request**") &&
				strings.Contains(post.Message, "A-TEST01")
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Create test approval record with playbook context
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-TEST01",
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			Description:          "Emergency DB access",
			CreatedAt:            1704988800000,
			PlaybookRunID:        "run456",
			PlaybookName:         "Incident #47",
			PlaybookChannelID:    "incident-channel-123",
		}

		// Execute
		postID, err := SendApprovalRequestDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "post_123", postID)
		api.AssertExpectations(t)
	})

	t.Run("excludes playbook context when not present", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// Verify playbook context is NOT included (AC4)
			return !strings.Contains(post.Message, "**Playbook Context:**") &&
				strings.Contains(post.Message, "📋 **Approval Request**") &&
				strings.Contains(post.Message, "A-TEST02")
		})).Return(&model.Post{Id: "post_456"}, nil)

		// Create test approval record WITHOUT playbook context
		record := &approval.ApprovalRecord{
			ID:                   "record456",
			Code:                 "A-TEST02",
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "bob",
			RequesterDisplayName: "Bob Smith",
			Description:          "Regular approval request",
			CreatedAt:            1704988800000,
			// No playbook fields
		}

		// Execute
		postID, err := SendApprovalRequestDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "post_456", postID)
		api.AssertExpectations(t)
	})

	t.Run("playbook context appears before buttons", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		approverID := "approver456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, approverID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("GetChannel", "deploy-channel").Return(&model.Channel{
			Id:   "deploy-channel",
			Name: "deploy-v3-0",
		}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_789"}, nil)

		// Create test approval record with playbook context
		record := &approval.ApprovalRecord{
			ID:                   "record789",
			Code:                 "A-TEST03",
			Status:               approval.StatusPending,
			ApproverID:           approverID,
			RequesterUsername:    "charlie",
			RequesterDisplayName: "Charlie Davis",
			Description:          "Test approval",
			CreatedAt:            1704988800000,
			PlaybookRunID:        "run789",
			PlaybookName:         "Deploy v3.0",
			PlaybookChannelID:    "deploy-channel",
		}

		// Execute
		_, err := SendApprovalRequestDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify playbook context appears after Request ID (AC5)
		requestIDIndex := strings.Index(capturedMessage, "**Request ID:**")
		playbookContextIndex := strings.Index(capturedMessage, "**Playbook Context:**")

		assert.NotEqual(t, -1, requestIDIndex, "Request ID should be present")
		assert.NotEqual(t, -1, playbookContextIndex, "Playbook context should be present")
		assert.Greater(t, playbookContextIndex, requestIDIndex, "Playbook context should appear after Request ID")

		api.AssertExpectations(t)
	})
}

// TestSendTimeoutNotificationDM tests the Story 10.7 Matterpoll pattern implementation
// for timeout notifications sent to requesters when their approval request times out.
func TestSendTimeoutNotificationDM(t *testing.T) {
	t.Run("successful timeout notification with Matterpoll pattern", func(t *testing.T) {
		// Setup mock API
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == botUserID &&
				post.ChannelId == dmChannelID &&
				post.Type == CustomApprovalDMPostType &&
				strings.Contains(post.Message, "⏱️ **Approval Request Timed Out**")
		})).Return(&model.Post{Id: "post_123"}, nil)

		// Create test approval record
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy hotfix to production",
			CreatedAt:            1704988800000, // 2024-01-11 12:00:00 UTC
		}

		// Execute
		postID, err := SendTimeoutNotificationDM(api, botUserID, record)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "post_123", postID)
		api.AssertExpectations(t)
	})

	t.Run("post type is custom_approval_dm (AC1)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC1: Verify post type is custom_approval_dm
		assert.Equal(t, CustomApprovalDMPostType, capturedPost.Type)
		api.AssertExpectations(t)
	})

	t.Run("notification_type is timeout (AC2)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC2: Verify notification_type prop is "timeout"
		assert.Equal(t, NotificationTypeTimeout, capturedPost.GetProp("notification_type"))
		api.AssertExpectations(t)
	})

	t.Run("created_at prop is populated (AC3)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		expectedCreatedAt := int64(1704988800000) // 2024-01-11 12:00:00 UTC

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            expectedCreatedAt,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC3: Verify created_at prop is populated with Unix milliseconds
		assert.Equal(t, expectedCreatedAt, capturedPost.GetProp("created_at"))
		api.AssertExpectations(t)
	})

	t.Run("no buttons in attachment (AC4)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC4: Verify no buttons are rendered for timeout notifications
		// Attachments are stored by ParseSlackAttachment - verify they exist and have no actions
		attachmentsProp := capturedPost.GetProp("attachments")
		assert.NotNil(t, attachmentsProp, "Attachments prop should exist")

		// ParseSlackAttachment stores as []*model.SlackAttachment
		attachments, ok := attachmentsProp.([]*model.SlackAttachment)
		assert.True(t, ok, "Attachments should be []*model.SlackAttachment type")
		assert.NotEmpty(t, attachments, "Should have at least one attachment")
		assert.Empty(t, attachments[0].Actions, "Timeout notifications should have no interactive buttons")
		api.AssertExpectations(t)
	})

	t.Run("empty bot user ID returns error", func(t *testing.T) {
		api := &plugintest.API{}

		record := &approval.ApprovalRecord{
			ID:          "record123",
			RequesterID: "requester456",
		}

		// Execute with empty botUserID
		_, err := SendTimeoutNotificationDM(api, "", record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot user ID not available")
	})

	t.Run("nil record returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		// Execute with nil record
		_, err := SendTimeoutNotificationDM(api, botUserID, nil)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record is nil")
	})

	t.Run("empty approval ID returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		record := &approval.ApprovalRecord{
			ID:          "", // Empty ID
			RequesterID: "requester456",
		}

		// Execute with empty ID
		_, err := SendTimeoutNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approval record ID is empty")
	})

	t.Run("empty requester ID returns error", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"

		record := &approval.ApprovalRecord{
			ID:          "record123",
			RequesterID: "", // Empty requester ID
		}

		// Execute with empty requester ID
		_, err := SendTimeoutNotificationDM(api, botUserID, record)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requester ID is empty")
	})

	t.Run("DM channel creation failure handled gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"

		// Mock GetDirectChannel to return error
		api.On("GetDirectChannel", botUserID, requesterID).Return(nil, &model.AppError{Message: "DM channel disabled"})

		record := &approval.ApprovalRecord{
			ID:          "record123",
			Code:        "A-X7K9Q2",
			RequesterID: requesterID,
			CreatedAt:   1704988800000,
		}

		// Execute - should return error
		_, err := SendTimeoutNotificationDM(api, botUserID, record)

		// Assert error is returned for caller to log
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get DM channel")
		api.AssertExpectations(t)
	})

	t.Run("CreatePost failure handled gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "network error"})

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		// Execute - should return error
		_, err := SendTimeoutNotificationDM(api, botUserID, record)

		// Assert error is returned for caller to log
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send timeout notification")
		api.AssertExpectations(t)
	})

	t.Run("markdown fallback includes all timeout details (AC5)", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedMessage string
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedMessage = post.Message
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy hotfix to production",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// AC5: Verify markdown fallback includes all required fields
		assert.Contains(t, capturedMessage, "⏱️ **Approval Request Timed Out**")
		assert.Contains(t, capturedMessage, "**Request ID:** `A-X7K9Q2`")
		assert.Contains(t, capturedMessage, "Deploy hotfix to production")
		assert.Contains(t, capturedMessage, "@bob (Bob Jones)")
		assert.Contains(t, capturedMessage, "No response within 30 minutes")
		assert.Contains(t, capturedMessage, "automatically canceled")
		api.AssertExpectations(t)
	})

	t.Run("sends to requester not approver", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		approverID := "approver789"
		dmChannelID := "dm_requester"

		// Should call GetDirectChannel with requester ID, not approver ID
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == dmChannelID
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			ApproverID:           approverID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify GetDirectChannel was called with requester (not approver)
		api.AssertCalled(t, "GetDirectChannel", botUserID, requesterID)
		api.AssertNotCalled(t, "GetDirectChannel", botUserID, approverID)
		api.AssertExpectations(t)
	})

	t.Run("is_dm prop is set to true", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify is_dm prop is set
		assert.Equal(t, true, capturedPost.GetProp("is_dm"))
		api.AssertExpectations(t)
	})

	t.Run("approver info in props for webapp display", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify approver info is included in props for webapp rendering
		assert.Equal(t, "bob", capturedPost.GetProp("approver_username"))
		assert.Equal(t, "Bob Jones", capturedPost.GetProp("approver_display_name"))
		api.AssertExpectations(t)
	})

	t.Run("approval_code prop is set correctly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-TIMEOUT7",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test timeout request",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify approval_code prop matches record.Code
		assert.Equal(t, "A-TIMEOUT7", capturedPost.GetProp("approval_code"))
		api.AssertExpectations(t)
	})

	t.Run("description prop is set correctly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		expectedDescription := "Deploy critical hotfix to production servers"
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          expectedDescription,
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify description prop matches record.Description
		assert.Equal(t, expectedDescription, capturedPost.GetProp("description"))
		api.AssertExpectations(t)
	})

	t.Run("requester info props are set correctly", func(t *testing.T) {
		api := &plugintest.API{}
		botUserID := "bot123"
		requesterID := "requester456"
		dmChannelID := "dm789"

		var capturedPost *model.Post
		api.On("GetDirectChannel", botUserID, requesterID).Return(&model.Channel{Id: dmChannelID}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{Id: "post_123"}, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusPending,
			RequesterID:          requesterID,
			RequesterUsername:    "alice.smith",
			RequesterDisplayName: "Alice Smith",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test description",
			CreatedAt:            1704988800000,
		}

		_, err := SendTimeoutNotificationDM(api, botUserID, record)
		assert.NoError(t, err)

		// Verify requester info props are included for completeness
		assert.Equal(t, "alice.smith", capturedPost.GetProp("requester_username"))
		assert.Equal(t, "Alice Smith", capturedPost.GetProp("requester_display_name"))
		api.AssertExpectations(t)
	})
}

// Helper function to verify the plugin.API interface is satisfied
var _ plugin.API = (*plugintest.API)(nil)
