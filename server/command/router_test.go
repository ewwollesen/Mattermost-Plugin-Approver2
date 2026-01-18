package command

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/playbooks"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockStore is a mock implementation of the Storer interface
type mockStore struct {
	mock.Mock
}

func (m *mockStore) GetAllApprovals() ([]*approval.ApprovalRecord, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*approval.ApprovalRecord), args.Error(1)
}

func (m *mockStore) GetUserApprovals(userID string) ([]*approval.ApprovalRecord, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*approval.ApprovalRecord), args.Error(1)
}

func (m *mockStore) GetApprovalByCode(code string) (*approval.ApprovalRecord, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*approval.ApprovalRecord), args.Error(1)
}

func TestRoute(t *testing.T) {
	api := &plugintest.API{}
	store := &mockStore{}
	router := NewRouter(api, store, nil)

	tests := []struct {
		name             string
		command          string
		expectedContains string
	}{
		{
			name:             "empty command shows help",
			command:          "/approve",
			expectedContains: "Available Commands",
		},
		{
			name:             "help command shows help",
			command:          "/approve help",
			expectedContains: "Available Commands",
		},
		{
			name:             "unknown command shows error",
			command:          "/approve unknown",
			expectedContains: "Unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := &model.CommandArgs{
				Command: tt.command,
			}

			resp, err := router.Route(args)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Contains(t, resp.Text, tt.expectedContains)
		})
	}
}

func TestRouteNew(t *testing.T) {
	t.Run("new command opens modal dialog with correct structure", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetConfig to return site URL
		siteURL := "https://mattermost.example.com"
		config := &model.Config{}
		config.ServiceSettings.SiteURL = &siteURL
		api.On("GetConfig").Return(config)

		// Mock OpenInteractiveDialog with strict validation
		api.On("OpenInteractiveDialog", mock.MatchedBy(func(req model.OpenDialogRequest) bool {
			dialog := req.Dialog

			// Verify dialog title (AC1)
			if dialog.Title != "Create Approval Request" {
				return false
			}

			// Verify submit label (AC1)
			if dialog.SubmitLabel != "Submit Request" {
				return false
			}

			// Verify callback ID
			if dialog.CallbackId != "approve_new" {
				return false
			}

			// Verify exactly 2 fields (AC1)
			if len(dialog.Elements) != 2 {
				return false
			}

			// Verify user selector field (AC2)
			approverField := dialog.Elements[0]
			if approverField.DisplayName != "Select approver *" ||
				approverField.Name != "approver" ||
				approverField.Type != "select" ||
				approverField.DataSource != "users" {
				return false
			}

			// Verify textarea field (AC3)
			descField := dialog.Elements[1]
			if descField.DisplayName != "What needs approval? *" ||
				descField.Name != "description" ||
				descField.Type != "textarea" ||
				descField.Placeholder != "Describe the action requiring approval" ||
				descField.MaxLength != 1000 {
				return false
			}

			// Verify trigger ID is passed
			if req.TriggerId != "test-trigger-id-12345678901234567890" {
				return false
			}

			// Verify callback URL
			expectedURL := "https://mattermost.example.com/plugins/com.mattermost.plugin-approver2/dialog/submit"
			return req.URL == expectedURL
		})).Return(nil)

		args := &model.CommandArgs{
			Command:   "/approve new",
			TriggerId: "test-trigger-id-12345678901234567890",
			UserId:    "user-id-abcdefghijklmnopqrstuvwxyz",
			ChannelId: "channel-id-abcdefghijklmnopqrstuvwxyz",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)

		// Verify OpenInteractiveDialog was called with correct structure
		api.AssertExpectations(t)
	})

	t.Run("new command with missing trigger_id returns error", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		args := &model.CommandArgs{
			Command:   "/approve new",
			TriggerId: "", // Missing trigger_id
			UserId:    "user-id-abcdefghijklmnopqrstuvwxyz",
			ChannelId: "channel-id-abcdefghijklmnopqrstuvwxyz",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "trigger")
	})

	t.Run("new command with missing site URL returns error", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetConfig to return empty site URL
		config := &model.Config{}
		emptySiteURL := ""
		config.ServiceSettings.SiteURL = &emptySiteURL
		api.On("GetConfig").Return(config)

		// Mock LogError call
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command:   "/approve new",
			TriggerId: "test-trigger-id-12345678901234567890",
			UserId:    "user-id-abcdefghijklmnopqrstuvwxyz",
			ChannelId: "channel-id-abcdefghijklmnopqrstuvwxyz",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Plugin configuration error")
		assert.Contains(t, resp.Text, "system administrator")

		api.AssertExpectations(t)
	})

	t.Run("new command with nil site URL returns error", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetConfig to return nil site URL
		config := &model.Config{}
		config.ServiceSettings.SiteURL = nil
		api.On("GetConfig").Return(config)

		// Mock LogError call
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command:   "/approve new",
			TriggerId: "test-trigger-id-12345678901234567890",
			UserId:    "user-id-abcdefghijklmnopqrstuvwxyz",
			ChannelId: "channel-id-abcdefghijklmnopqrstuvwxyz",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Plugin configuration error")

		api.AssertExpectations(t)
	})

	t.Run("new command with OpenInteractiveDialog failure returns error", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetConfig to return site URL
		siteURL := "https://mattermost.example.com"
		config := &model.Config{}
		config.ServiceSettings.SiteURL = &siteURL
		api.On("GetConfig").Return(config)

		// Mock OpenInteractiveDialog to fail
		appErr := model.NewAppError("OpenInteractiveDialog", "app.plugin.open_dialog.error", nil, "test error", 500)
		api.On("OpenInteractiveDialog", mock.Anything).Return(appErr)

		// Mock LogError for the failure
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command:   "/approve new",
			TriggerId: "test-trigger-id-12345678901234567890",
			UserId:    "user-id-abcdefghijklmnopqrstuvwxyz",
			ChannelId: "channel-id-abcdefghijklmnopqrstuvwxyz",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Failed to open approval request form")
		assert.Contains(t, resp.Text, "try again")

		api.AssertExpectations(t)
	})
}

func TestExecuteStatus(t *testing.T) {
	t.Run("status command requires system admin", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetUser to return non-admin user
		user := &model.User{
			Id:    "user123",
			Roles: "system_user", // Not an admin
		}
		api.On("GetUser", "user123").Return(user, nil)

		args := &model.CommandArgs{
			Command: "/approve status",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Permission denied")
		assert.Contains(t, resp.Text, "system administrators")

		api.AssertExpectations(t)
	})

	t.Run("status command returns no approvals message when empty", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetUser to return admin user
		user := &model.User{
			Id:    "admin123",
			Roles: "system_user system_admin",
		}
		api.On("GetUser", "admin123").Return(user, nil)

		// Mock GetAllApprovals to return empty slice
		store.On("GetAllApprovals").Return([]*approval.ApprovalRecord{}, nil)

		args := &model.CommandArgs{
			Command: "/approve status",
			UserId:  "admin123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "No approvals in the system")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("status command displays statistics correctly", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetUser to return admin user
		user := &model.User{
			Id:    "admin123",
			Roles: "system_user system_admin",
		}
		api.On("GetUser", "admin123").Return(user, nil)

		// Mock GetAllApprovals to return test data
		records := []*approval.ApprovalRecord{
			// Pending with notification sent
			{
				ID:               "id1",
				Code:             "A-TEST1",
				Status:           approval.StatusPending,
				NotificationSent: true,
			},
			// Pending with notification failed
			{
				ID:               "id2",
				Code:             "A-TEST2",
				Status:           approval.StatusPending,
				NotificationSent: false,
			},
			// Approved with outcome notified
			{
				ID:              "id3",
				Code:            "A-TEST3",
				Status:          approval.StatusApproved,
				OutcomeNotified: true,
			},
			// Denied with outcome notification failed
			{
				ID:              "id4",
				Code:            "A-TEST4",
				Status:          approval.StatusDenied,
				OutcomeNotified: false,
			},
			// Canceled
			{
				ID:     "id5",
				Code:   "A-TEST5",
				Status: approval.StatusCanceled,
			},
		}
		store.On("GetAllApprovals").Return(records, nil)

		args := &model.CommandArgs{
			Command: "/approve status",
			UserId:  "admin123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify statistics in response
		assert.Contains(t, resp.Text, "Total Approvals: 5")
		assert.Contains(t, resp.Text, "Pending: 2")
		assert.Contains(t, resp.Text, "Approved: 1")
		assert.Contains(t, resp.Text, "Denied: 1")
		assert.Contains(t, resp.Text, "Canceled: 1")
		assert.Contains(t, resp.Text, "Failed Approver Notifications: 1")
		assert.Contains(t, resp.Text, "Failed Outcome Notifications: 1")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("status command with --failed-notifications flag", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetUser to return admin user
		user := &model.User{
			Id:    "admin123",
			Roles: "system_user system_admin",
		}
		api.On("GetUser", "admin123").Return(user, nil)

		// Mock GetAllApprovals to return test data
		records := []*approval.ApprovalRecord{
			// Pending with notification failed
			{
				ID:                "id1",
				Code:              "A-FAIL1",
				Status:            approval.StatusPending,
				NotificationSent:  false,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1641024000000, // 2022-01-01 12:00:00
			},
			// Approved with outcome notification failed
			{
				ID:                "id2",
				Code:              "A-FAIL2",
				Status:            approval.StatusApproved,
				OutcomeNotified:   false,
				RequesterUsername: "charlie",
				ApproverUsername:  "david",
				DecidedAt:         1641027600000, // 2022-01-01 13:00:00
			},
		}
		store.On("GetAllApprovals").Return(records, nil)

		args := &model.CommandArgs{
			Command: "/approve status --failed-notifications",
			UserId:  "admin123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify failed notifications are listed
		assert.Contains(t, resp.Text, "Failed Approver Notifications")
		assert.Contains(t, resp.Text, "A-FAIL1")
		assert.Contains(t, resp.Text, "@alice")
		assert.Contains(t, resp.Text, "@bob")

		assert.Contains(t, resp.Text, "Failed Outcome Notifications")
		assert.Contains(t, resp.Text, "A-FAIL2")
		assert.Contains(t, resp.Text, "@charlie")
		assert.Contains(t, resp.Text, "@david")

		assert.Contains(t, resp.Text, "Guidance")
		assert.Contains(t, resp.Text, "DM settings")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("status command with --failed-notifications when all notifications succeeded", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetUser to return admin user
		user := &model.User{
			Id:    "admin123",
			Roles: "system_user system_admin",
		}
		api.On("GetUser", "admin123").Return(user, nil)

		// Mock GetAllApprovals to return test data with all notifications sent
		records := []*approval.ApprovalRecord{
			{
				ID:               "id1",
				Code:             "A-OK1",
				Status:           approval.StatusPending,
				NotificationSent: true,
			},
			{
				ID:              "id2",
				Code:            "A-OK2",
				Status:          approval.StatusApproved,
				OutcomeNotified: true,
			},
		}
		store.On("GetAllApprovals").Return(records, nil)

		args := &model.CommandArgs{
			Command: "/approve status --failed-notifications",
			UserId:  "admin123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify success message
		assert.Contains(t, resp.Text, "No failed notifications found")
		assert.Contains(t, resp.Text, "delivered successfully")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("status command handles store error gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock GetUser to return admin user
		user := &model.User{
			Id:    "admin123",
			Roles: "system_user system_admin",
		}
		api.On("GetUser", "admin123").Return(user, nil)

		// Mock GetAllApprovals to return error
		store.On("GetAllApprovals").Return(nil, assert.AnError)

		// Mock LogError call
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command: "/approve status",
			UserId:  "admin123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Failed to retrieve approval statistics")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})
}

func TestExecuteList(t *testing.T) {
	t.Run("empty state - user with no records", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock store to return empty slice
		store.On("GetUserApprovals", "user123").Return([]*approval.ApprovalRecord{}, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Story 5.2: Filter-specific empty state message
		assert.Contains(t, capturedPost.Message, "No pending approval requests")
		assert.Contains(t, capturedPost.Message, "/approve list all")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("single record - verifies formatting and status icons", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-ABC123",
				Status:            approval.StatusPending,
				RequesterID:       "user123",
				RequesterUsername: "alice",
				ApproverID:        "user456",
				ApproverUsername:  "bob",
				CreatedAt:         1705000000000, // 2024-01-11 ~20:26 UTC
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Story 5.2: New header format with count
		assert.Contains(t, capturedPost.Message, "## Your Approval Requests (1 pending)")
		// Story 7.5: Markdown table format
		assert.Contains(t, capturedPost.Message, "| Code | Status | Requestor | Approver | Created |")
		assert.Contains(t, capturedPost.Message, "|------|--------|-----------|----------|----------|")
		assert.Contains(t, capturedPost.Message, "| A-ABC123 |")
		assert.Contains(t, capturedPost.Message, "⏳ Pending")
		assert.Contains(t, capturedPost.Message, "@alice")
		assert.Contains(t, capturedPost.Message, "@bob")
		assert.Contains(t, capturedPost.Message, "2024-01-11")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("multiple records - verifies sorting by most recent first", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Create records with different timestamps (store returns them sorted)
		records := []*approval.ApprovalRecord{
			{
				ID:                "record3",
				Code:              "A-NEW",
				Status:            approval.StatusApproved,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         3000, // Newest
			},
			{
				ID:                "record2",
				Code:              "A-MID",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "charlie",
				CreatedAt:         2000, // Middle
			},
			{
				ID:                "record1",
				Code:              "A-OLD",
				Status:            approval.StatusDenied,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1000, // Oldest
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list all", // Story 5.2: Use explicit 'all' filter to test sorting across all statuses
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Verify records appear in order (newest first)
		assert.Contains(t, capturedPost.Message, "A-NEW")
		assert.Contains(t, capturedPost.Message, "A-MID")
		assert.Contains(t, capturedPost.Message, "A-OLD")

		// Verify A-NEW appears before A-OLD in output
		newPos := strings.Index(capturedPost.Message, "A-NEW")
		oldPos := strings.Index(capturedPost.Message, "A-OLD")
		assert.Less(t, newPos, oldPos, "Newest record should appear before oldest")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("pagination - shows only first 20 records with footer", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Create 25 records
		records := make([]*approval.ApprovalRecord, 25)
		for i := range 25 {
			records[i] = &approval.ApprovalRecord{
				ID:                fmt.Sprintf("record%d", i),
				Code:              fmt.Sprintf("A-REC%02d", i),
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         int64(25-i) * 1000, // Descending timestamps
			}
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Should show first 20 records
		assert.Contains(t, capturedPost.Message, "A-REC00")
		assert.Contains(t, capturedPost.Message, "A-REC19")
		// Should NOT show 21st record
		assert.NotContains(t, capturedPost.Message, "A-REC20")
		assert.NotContains(t, capturedPost.Message, "A-REC24")

		// Should show pagination footer with new format (Story 4.6: updated message)
		assert.Contains(t, capturedPost.Message, "Showing 20 of 25 total records")
		assert.Contains(t, capturedPost.Message, "/approve get")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("access control - user sees only their records", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Records already filtered by store (GetUserApprovals does this)
		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-USER1",
				RequesterID:       "user123",
				ApproverID:        "user456",
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				Status:            approval.StatusPending,
				CreatedAt:         1000,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		assert.Contains(t, capturedPost.Message, "A-USER1")

		// Verify store was called with correct user ID
		store.AssertCalled(t, "GetUserApprovals", "user123")
		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("requester records - user sees records where they are requester", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-REQ1",
				RequesterID:       "user123", // User is requester
				ApproverID:        "user456",
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				Status:            approval.StatusApproved,
				CreatedAt:         1000,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list all", // Story 5.2: Use explicit 'all' filter to test access control across all statuses
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		assert.Contains(t, capturedPost.Message, "A-REQ1")
		assert.Contains(t, capturedPost.Message, "@alice") // Requester

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("approver records - user sees records where they are approver", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-APP1",
				RequesterID:       "user456",
				ApproverID:        "user123", // User is approver
				RequesterUsername: "bob",
				ApproverUsername:  "alice",
				Status:            approval.StatusPending,
				CreatedAt:         1000,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		assert.Contains(t, capturedPost.Message, "A-APP1")
		assert.Contains(t, capturedPost.Message, "@alice") // Approver

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("mixed records - user sees both requester and approver records", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-REQ",
				RequesterID:       "user123", // User is requester
				ApproverID:        "user456",
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				Status:            approval.StatusApproved,
				CreatedAt:         2000,
			},
			{
				ID:                "record2",
				Code:              "A-APP",
				RequesterID:       "user456",
				ApproverID:        "user123", // User is approver
				RequesterUsername: "bob",
				ApproverUsername:  "alice",
				Status:            approval.StatusPending,
				CreatedAt:         1000,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list all", // Story 5.2: Use explicit 'all' filter to test both requester and approver records across all statuses
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		assert.Contains(t, capturedPost.Message, "A-REQ")
		assert.Contains(t, capturedPost.Message, "A-APP")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("KV store error handling - returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Mock store to return error
		storeErr := fmt.Errorf("KV store connection failed")
		store.On("GetUserApprovals", "user123").Return(nil, storeErr)

		// Mock LogError call
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		assert.Contains(t, capturedPost.Message, "❌ Failed to retrieve approval records")
		assert.Contains(t, capturedPost.Message, "Please try again")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("all status types - verifies correct icons", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-APPR",
				Status:            approval.StatusApproved,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         4000,
			},
			{
				ID:                "record2",
				Code:              "A-DENY",
				Status:            approval.StatusDenied,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         3000,
			},
			{
				ID:                "record3",
				Code:              "A-PEND",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         2000,
			},
			{
				ID:                "record4",
				Code:              "A-CANC",
				Status:            approval.StatusCanceled,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1000,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list all", // Story 5.2: Use explicit 'all' filter to test icons for all status types
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)

		// Verify status icons
		assert.Contains(t, capturedPost.Message, "✅ Approved")
		assert.Contains(t, capturedPost.Message, "❌ Denied")
		assert.Contains(t, capturedPost.Message, "⏳ Pending")
		assert.Contains(t, capturedPost.Message, "🚫 Canceled")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("timestamp formatting - verifies date format", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		// Use a specific timestamp: Jan 10, 2024 14:30:00 UTC
		timestamp := int64(1704897000000) // milliseconds
		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-TIME",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         timestamp,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)

		// Verify timestamp format: YYYY-MM-DD HH:MM
		assert.Contains(t, capturedPost.Message, "2024-01-10 14:30")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	// Story 5.2: Tests for count display in header
	t.Run("header count - shows count for pending filter", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-PEND1",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1000,
			},
			{
				ID:                "record2",
				Code:              "A-PEND2",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         2000,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list", // Defaults to pending
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Story 5.2 AC2, AC3: Header should show count
		assert.Contains(t, capturedPost.Message, "## Your Approval Requests (2 pending)")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("header count - shows count for approved filter", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-APPR",
				Status:            approval.StatusApproved,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1000,
				DecidedAt:         1500,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list approved",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Story 5.2 AC3: Header should show count with filter type
		assert.Contains(t, capturedPost.Message, "## Your Approval Requests (1 approved)")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("header count - shows count for all filter", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{
			{
				ID:                "record1",
				Code:              "A-PEND",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1000,
			},
			{
				ID:                "record2",
				Code:              "A-APPR",
				Status:            approval.StatusApproved,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         2000,
				DecidedAt:         2500,
			},
			{
				ID:                "record3",
				Code:              "A-DENY",
				Status:            approval.StatusDenied,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         3000,
				DecidedAt:         3500,
			},
		}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list all",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Story 5.2 AC3: Header should show total count for "all" filter
		assert.Contains(t, capturedPost.Message, "## Your Approval Requests (3 all)")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("header count - shows zero count", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		records := []*approval.ApprovalRecord{}
		store.On("GetUserApprovals", "user123").Return(records, nil)

		// Story 7.5: Mock SendEphemeralPost to capture the message
		var capturedPost *model.Post
		api.On("SendEphemeralPost", mock.Anything, mock.MatchedBy(func(post *model.Post) bool {
			capturedPost = post
			return true
		})).Return(&model.Post{})

		args := &model.CommandArgs{
			Command:   "/approve list pending",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify the message was sent via ephemeral post
		assert.NotNil(t, capturedPost)
		// Story 5.2 AC4: Empty state message for zero count
		assert.Contains(t, capturedPost.Message, "No pending approval requests")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})
}

func TestExecuteGet(t *testing.T) {
	t.Run("successfully retrieves record by code", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                   "abc123def456ghi789jkl012",
			Code:                 "A-X7K9Q2",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			Description:          "Emergency rollback of payment-config-v2 deployment",
			Status:               approval.StatusApproved,
			DecisionComment:      "Approved. Rollback immediately.",
			CreatedAt:            1704897000000, // 2024-01-10 14:30:00 UTC
			DecidedAt:            1704897300000, // 2024-01-10 14:35:00 UTC
			RequestChannelID:     "channel123",
			TeamID:               "team456",
		}

		store.On("GetApprovalByCode", "A-X7K9Q2").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-X7K9Q2",
			UserId:  "user1", // Requester can view
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)

		// Verify complete record details are shown
		assert.Contains(t, resp.Text, "📋 Approval Record: A-X7K9Q2")
		assert.Contains(t, resp.Text, "✅ Approved")
		assert.Contains(t, resp.Text, "@alice")
		assert.Contains(t, resp.Text, "Alice Carter")
		assert.Contains(t, resp.Text, "@bob")
		assert.Contains(t, resp.Text, "Bob Smith")
		assert.Contains(t, resp.Text, "Emergency rollback of payment-config-v2 deployment")
		assert.Contains(t, resp.Text, "Approved. Rollback immediately.")
		assert.Contains(t, resp.Text, "2024-01-10 14:30:00 UTC")
		assert.Contains(t, resp.Text, "2024-01-10 14:35:00 UTC")
		assert.Contains(t, resp.Text, "immutable")

		store.AssertExpectations(t)
	})

	t.Run("successfully retrieves record by full 26-char ID", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		fullID := "abcdefghijklmnopqrstuvwxyz" // Exactly 26 chars
		record := &approval.ApprovalRecord{
			ID:                   fullID,
			Code:                 "A-X7K9Q2",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			Status:               approval.StatusPending,
			Description:          "Test approval",
			CreatedAt:            1704897000000,
		}

		store.On("GetApprovalByCode", fullID).Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get " + fullID,
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "📋 Approval Record: A-X7K9Q2")
		assert.Contains(t, resp.Text, "⏳ Pending")

		store.AssertExpectations(t)
	})

	t.Run("returns error when code not found", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		store.On("GetApprovalByCode", "A-NOTFND").Return(nil, fmt.Errorf("approval code A-NOTFND: %w", approval.ErrRecordNotFound))

		args := &model.CommandArgs{
			Command: "/approve get A-NOTFND",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "❌ Approval record 'A-NOTFND' not found")
		assert.Contains(t, resp.Text, "/approve list")

		store.AssertExpectations(t)
	})

	t.Run("returns usage help when code parameter missing", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		args := &model.CommandArgs{
			Command: "/approve get",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Usage: /approve get <APPROVAL_ID>")
		assert.Contains(t, resp.Text, "Example: /approve get A-X7K9Q2")
	})

	t.Run("requester can view their own request", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                "record1",
			Code:              "A-TEST1",
			RequesterID:       "user1", // User is requester
			RequesterUsername: "alice",
			ApproverID:        "user2",
			ApproverUsername:  "bob",
			Status:            approval.StatusPending,
			Description:       "Test",
			CreatedAt:         1000,
		}

		store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-TEST1",
			UserId:  "user1", // Requester's user ID
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "📋 Approval Record")
		assert.Contains(t, resp.Text, "A-TEST1")

		store.AssertExpectations(t)
	})

	t.Run("approver can view requests they approve", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                "record1",
			Code:              "A-TEST1",
			RequesterID:       "user1",
			RequesterUsername: "alice",
			ApproverID:        "user2", // User is approver
			ApproverUsername:  "bob",
			Status:            approval.StatusPending,
			Description:       "Test",
			CreatedAt:         1000,
		}

		store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-TEST1",
			UserId:  "user2", // Approver's user ID
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "📋 Approval Record")
		assert.Contains(t, resp.Text, "A-TEST1")

		store.AssertExpectations(t)
	})

	t.Run("unauthorized user receives permission denied", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:          "record1",
			Code:        "A-TEST1",
			RequesterID: "user1",
			ApproverID:  "user2",
			Status:      approval.StatusPending,
			Description: "Sensitive data",
			CreatedAt:   1000,
		}

		store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

		// Mock LogWarn for unauthorized access attempt
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command: "/approve get A-TEST1",
			UserId:  "user3", // Neither requester nor approver
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "❌ Permission denied")
		assert.Contains(t, resp.Text, "only view approval records where you are the requester or approver")
		// Should not leak record details
		assert.NotContains(t, resp.Text, "Sensitive data")
		assert.NotContains(t, resp.Text, "A-TEST1")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("displays all status types with correct icons", func(t *testing.T) {
		testCases := []struct {
			status       string
			expectedIcon string
		}{
			{approval.StatusApproved, "✅ Approved"},
			{approval.StatusDenied, "❌ Denied"},
			{approval.StatusPending, "⏳ Pending"},
			{approval.StatusCanceled, "🚫 Canceled"},
		}

		for _, tc := range testCases {
			t.Run(tc.status, func(t *testing.T) {
				api := &plugintest.API{}
				store := &mockStore{}
				router := NewRouter(api, store, nil)

				record := &approval.ApprovalRecord{
					ID:                "record1",
					Code:              "A-TEST1",
					RequesterID:       "user1",
					RequesterUsername: "alice",
					ApproverID:        "user2",
					Status:            tc.status,
					Description:       "Test",
					CreatedAt:         1000,
				}

				store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

				args := &model.CommandArgs{
					Command: "/approve get A-TEST1",
					UserId:  "user1",
				}

				resp, err := router.Route(args)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Contains(t, resp.Text, tc.expectedIcon)

				store.AssertExpectations(t)
			})
		}
	})

	t.Run("displays decision comment when present", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                "record1",
			Code:              "A-TEST1",
			RequesterID:       "user1",
			RequesterUsername: "alice",
			ApproverID:        "user2",
			Status:            approval.StatusApproved,
			Description:       "Test",
			DecisionComment:   "Looks good to me",
			CreatedAt:         1000,
			DecidedAt:         2000,
		}

		store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-TEST1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "Decision Comment")
		assert.Contains(t, resp.Text, "Looks good to me")

		store.AssertExpectations(t)
	})

	t.Run("omits decision comment when not present", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                "record1",
			Code:              "A-TEST1",
			RequesterID:       "user1",
			RequesterUsername: "alice",
			ApproverID:        "user2",
			Status:            approval.StatusPending,
			Description:       "Test",
			DecisionComment:   "", // No comment
			CreatedAt:         1000,
		}

		store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-TEST1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Should not show decision comment section
		assert.NotContains(t, resp.Text, "Decision Comment")

		store.AssertExpectations(t)
	})

	t.Run("handles store error gracefully", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		storeErr := fmt.Errorf("KV store connection failed")
		store.On("GetApprovalByCode", "A-TEST1").Return(nil, storeErr)

		// Mock LogError for the failure
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command: "/approve get A-TEST1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "❌ Failed to retrieve approval record")
		assert.Contains(t, resp.Text, "try again")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	// Story 4.5: Display cancellation reason in approval details
	t.Run("displays cancellation details for cancelled request with reason", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record1",
			Code:                 "A-CANC1",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			Description:          "Deploy to production",
			Status:               approval.StatusCanceled,
			CanceledReason:       "No longer needed",
			CanceledAt:           1704897300000, // 2024-01-10 14:35:00 UTC
			CreatedAt:            1704897000000, // 2024-01-10 14:30:00 UTC
			DecidedAt:            1704897300000, // 2024-01-10 14:35:00 UTC
		}

		store.On("GetApprovalByCode", "A-CANC1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-CANC1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify cancellation section is present (AC1, AC2)
		assert.Contains(t, resp.Text, "Cancellation:")
		assert.Contains(t, resp.Text, "**Reason:** No longer needed")
		assert.Contains(t, resp.Text, "**Canceled by:** @alice")
		assert.Contains(t, resp.Text, "**Canceled:** 2024-01-10 14:35:00 UTC")

		// Verify visual separator is present (AC3)
		assert.Contains(t, resp.Text, "---")

		// Verify cancellation section appears after description (AC3, Subtask 3.4)
		descPos := strings.Index(resp.Text, "**Description:**")
		cancelPos := strings.Index(resp.Text, "**Cancellation:**")
		assert.Greater(t, cancelPos, descPos, "Cancellation section should appear after Description")

		// Verify cancellation section appears before Context
		contextPos := strings.Index(resp.Text, "**Context:**")
		assert.Less(t, cancelPos, contextPos, "Cancellation section should appear before Context")

		// Code Review Fix: Verify "Decided:" is NOT shown for cancelled requests (Issue #3)
		assert.NotContains(t, resp.Text, "**Decided:**", "Cancelled requests should not show Decided timestamp")

		store.AssertExpectations(t)
	})

	t.Run("displays fallback text for cancelled request without reason (old record)", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record1",
			Code:                 "A-OLD1",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			Description:          "Old cancelled request",
			Status:               approval.StatusCanceled,
			CanceledReason:       "", // Empty - old record
			CanceledAt:           0,  // Zero value - old record
			CreatedAt:            1704897000000,
			DecidedAt:            1704897300000,
		}

		store.On("GetApprovalByCode", "A-OLD1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-OLD1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify fallback text for empty reason (AC5, Subtask 3.2)
		assert.Contains(t, resp.Text, "**Reason:** No reason recorded (canceled before v0.2.0)")

		// Verify fallback text for zero timestamp (AC6, Subtask 3.3)
		assert.Contains(t, resp.Text, "**Canceled:** Unknown")

		// Should still show who cancelled
		assert.Contains(t, resp.Text, "**Canceled by:** @alice")

		// Code Review Fix: Verify "Decided:" is NOT shown for cancelled requests (Issue #3)
		assert.NotContains(t, resp.Text, "**Decided:**", "Cancelled requests should not show Decided timestamp")

		store.AssertExpectations(t)
	})

	t.Run("displays cancellation with Other reason and custom text", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record1",
			Code:                 "A-OTHER1",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			Description:          "Complex deployment",
			Status:               approval.StatusCanceled,
			CanceledReason:       "Other: The deployment window was moved to next week due to unexpected infrastructure changes",
			CanceledAt:           1704897300000,
			CreatedAt:            1704897000000,
			DecidedAt:            1704897300000,
		}

		store.On("GetApprovalByCode", "A-OTHER1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-OTHER1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify full "Other" reason text is displayed (Subtask 3.6)
		assert.Contains(t, resp.Text, "**Reason:** Other: The deployment window was moved to next week due to unexpected infrastructure changes")

		store.AssertExpectations(t)
	})

	// Code Review Fix: Add tests for all cancellation reasons (Issue #4)
	t.Run("displays cancellation with 'Changed requirements' reason", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record1",
			Code:                 "A-CHG1",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			Description:          "Deploy feature X",
			Status:               approval.StatusCanceled,
			CanceledReason:       "Changed requirements",
			CanceledAt:           1704897300000,
			CreatedAt:            1704897000000,
		}

		store.On("GetApprovalByCode", "A-CHG1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-CHG1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify "Changed requirements" reason displays correctly
		assert.Contains(t, resp.Text, "**Reason:** Changed requirements")
		assert.NotContains(t, resp.Text, "**Decided:**", "Cancelled requests should not show Decided timestamp")

		store.AssertExpectations(t)
	})

	t.Run("displays cancellation with 'Created by mistake' reason", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil)

		record := &approval.ApprovalRecord{
			ID:                   "record1",
			Code:                 "A-MIST1",
			RequesterID:          "user1",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Carter",
			ApproverID:           "user2",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Smith",
			Description:          "Wrong approval",
			Status:               approval.StatusCanceled,
			CanceledReason:       "Created by mistake",
			CanceledAt:           1704897300000,
			CreatedAt:            1704897000000,
		}

		store.On("GetApprovalByCode", "A-MIST1").Return(record, nil)

		args := &model.CommandArgs{
			Command: "/approve get A-MIST1",
			UserId:  "user1",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify "Created by mistake" reason displays correctly
		assert.Contains(t, resp.Text, "**Reason:** Created by mistake")
		assert.NotContains(t, resp.Text, "**Decided:**", "Cancelled requests should not show Decided timestamp")

		store.AssertExpectations(t)
	})

	t.Run("does not display cancellation section for non-cancelled statuses", func(t *testing.T) {
		testCases := []struct {
			name   string
			status string
		}{
			{"pending", approval.StatusPending},
			{"approved", approval.StatusApproved},
			{"denied", approval.StatusDenied},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				api := &plugintest.API{}
				store := &mockStore{}
				router := NewRouter(api, store, nil)

				record := &approval.ApprovalRecord{
					ID:                   "record1",
					Code:                 "A-TEST1",
					RequesterID:          "user1",
					RequesterUsername:    "alice",
					RequesterDisplayName: "Alice Carter",
					ApproverID:           "user2",
					ApproverUsername:     "bob",
					ApproverDisplayName:  "Bob Smith",
					Description:          "Test request",
					Status:               tc.status,
					CreatedAt:            1704897000000,
				}

				store.On("GetApprovalByCode", "A-TEST1").Return(record, nil)

				args := &model.CommandArgs{
					Command: "/approve get A-TEST1",
					UserId:  "user1",
				}

				resp, err := router.Route(args)
				assert.NoError(t, err)
				assert.NotNil(t, resp)

				// Should NOT contain cancellation section
				assert.NotContains(t, resp.Text, "**Cancellation:**")
				assert.NotContains(t, resp.Text, "**Canceled by:**")

				store.AssertExpectations(t)
			})
		}
	})
}

// Story 4.6: Tests for groupAndSortRecords and grouped list display

func TestGroupAndSortRecords(t *testing.T) {
	t.Run("separates records into three groups", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{ID: "1", Status: approval.StatusPending, CreatedAt: 1000},
			{ID: "2", Status: approval.StatusApproved, DecidedAt: 2000},
			{ID: "3", Status: approval.StatusCanceled, CanceledAt: 3000},
			{ID: "4", Status: approval.StatusDenied, DecidedAt: 4000},
			{ID: "5", Status: approval.StatusPending, CreatedAt: 5000},
		}

		pending, decided, canceled := groupAndSortRecords(records)

		// Verify group sizes
		assert.Equal(t, 2, len(pending), "Should have 2 pending records")
		assert.Equal(t, 2, len(decided), "Should have 2 decided records")
		assert.Equal(t, 1, len(canceled), "Should have 1 canceled record")

		// Verify pending group contains correct records
		assert.Contains(t, []string{"1", "5"}, pending[0].ID)
		assert.Contains(t, []string{"1", "5"}, pending[1].ID)

		// Verify decided group contains correct records
		assert.Contains(t, []string{"2", "4"}, decided[0].ID)
		assert.Contains(t, []string{"2", "4"}, decided[1].ID)

		// Verify canceled group
		assert.Equal(t, "3", canceled[0].ID)
	})

	t.Run("sorts pending by CreatedAt descending", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{ID: "1", Status: approval.StatusPending, CreatedAt: 1000},
			{ID: "2", Status: approval.StatusPending, CreatedAt: 3000},
			{ID: "3", Status: approval.StatusPending, CreatedAt: 2000},
		}

		pending, _, _ := groupAndSortRecords(records)

		// Newest first (descending)
		assert.Equal(t, "2", pending[0].ID, "First should be ID 2 with CreatedAt 3000")
		assert.Equal(t, "3", pending[1].ID, "Second should be ID 3 with CreatedAt 2000")
		assert.Equal(t, "1", pending[2].ID, "Third should be ID 1 with CreatedAt 1000")
	})

	t.Run("sorts decided by DecidedAt descending", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{ID: "1", Status: approval.StatusApproved, DecidedAt: 1000},
			{ID: "2", Status: approval.StatusDenied, DecidedAt: 3000},
			{ID: "3", Status: approval.StatusApproved, DecidedAt: 2000},
		}

		_, decided, _ := groupAndSortRecords(records)

		// Newest first (descending)
		assert.Equal(t, "2", decided[0].ID, "First should be ID 2 with DecidedAt 3000")
		assert.Equal(t, "3", decided[1].ID, "Second should be ID 3 with DecidedAt 2000")
		assert.Equal(t, "1", decided[2].ID, "Third should be ID 1 with DecidedAt 1000")
	})

	t.Run("sorts canceled by CanceledAt descending", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{ID: "1", Status: approval.StatusCanceled, CanceledAt: 1000, CreatedAt: 5000},
			{ID: "2", Status: approval.StatusCanceled, CanceledAt: 3000, CreatedAt: 6000},
			{ID: "3", Status: approval.StatusCanceled, CanceledAt: 2000, CreatedAt: 7000},
		}

		_, _, canceled := groupAndSortRecords(records)

		// Newest first (descending by CanceledAt)
		assert.Equal(t, "2", canceled[0].ID, "First should be ID 2 with CanceledAt 3000")
		assert.Equal(t, "3", canceled[1].ID, "Second should be ID 3 with CanceledAt 2000")
		assert.Equal(t, "1", canceled[2].ID, "Third should be ID 1 with CanceledAt 1000")
	})

	t.Run("falls back to CreatedAt when CanceledAt is zero", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{ID: "1", Status: approval.StatusCanceled, CanceledAt: 0, CreatedAt: 5000},
			{ID: "2", Status: approval.StatusCanceled, CanceledAt: 3000, CreatedAt: 1000},
			{ID: "3", Status: approval.StatusCanceled, CanceledAt: 0, CreatedAt: 7000},
		}

		_, _, canceled := groupAndSortRecords(records)

		// Should sort: ID 3 (CreatedAt 7000), ID 1 (CreatedAt 5000), ID 2 (CanceledAt 3000)
		assert.Equal(t, "3", canceled[0].ID, "First should be ID 3 with CreatedAt 7000")
		assert.Equal(t, "1", canceled[1].ID, "Second should be ID 1 with CreatedAt 5000")
		assert.Equal(t, "2", canceled[2].ID, "Third should be ID 2 with CanceledAt 3000")
	})

	t.Run("handles empty input", func(t *testing.T) {
		records := []*approval.ApprovalRecord{}

		pending, decided, canceled := groupAndSortRecords(records)

		assert.Equal(t, 0, len(pending))
		assert.Equal(t, 0, len(decided))
		assert.Equal(t, 0, len(canceled))
	})
}

func TestFormatListResponse_GroupedSections(t *testing.T) {
	t.Run("displays all three sections with headers", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{
				Code:              "A-PND1",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1704897000000,
			},
			{
				Code:              "A-APP1",
				Status:            approval.StatusApproved,
				RequesterUsername: "charlie",
				ApproverUsername:  "diane",
				CreatedAt:         1704897000000,
				DecidedAt:         1704897100000,
			},
			{
				Code:              "A-CAN1",
				Status:            approval.StatusCanceled,
				RequesterUsername: "eve",
				ApproverUsername:  "frank",
				CreatedAt:         1704897000000,
				CanceledAt:        1704897200000,
				CanceledReason:    "No longer needed",
			},
		}

		result := formatListResponse(records, 3, "all")

		// Verify section headers appear in correct order
		assert.Contains(t, result, "**Pending Approvals:**")
		assert.Contains(t, result, "**Decided Approvals:**")
		assert.Contains(t, result, "**Canceled Requests:**")

		// Verify markdown table format
		assert.Contains(t, result, "| Code | Status | Requestor | Approver | Created |")
		assert.Contains(t, result, "|------|--------|-----------|----------|----------|")

		// Verify order of sections (pending before decided before canceled)
		pendingIdx := strings.Index(result, "**Pending Approvals:**")
		decidedIdx := strings.Index(result, "**Decided Approvals:**")
		canceledIdx := strings.Index(result, "**Canceled Requests:**")

		assert.True(t, pendingIdx < decidedIdx, "Pending should come before Decided")
		assert.True(t, decidedIdx < canceledIdx, "Decided should come before Canceled")

		// Verify records appear in correct sections (in table row format)
		assert.Contains(t, result, "| A-PND1 |")
		assert.Contains(t, result, "| A-APP1 |")
		assert.Contains(t, result, "| A-CAN1 |")
	})

	t.Run("omits empty sections", func(t *testing.T) {
		// Only pending records
		records := []*approval.ApprovalRecord{
			{
				Code:              "A-PND1",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1704897000000,
			},
		}

		result := formatListResponse(records, 1, "all")

		assert.Contains(t, result, "**Pending Approvals:**")
		assert.NotContains(t, result, "**Decided Approvals:**")
		assert.NotContains(t, result, "**Canceled Requests:**")
	})

	t.Run("displays canceled reason in list view", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{
				Code:              "A-CAN1",
				Status:            approval.StatusCanceled,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1704897000000,
				CanceledAt:        1704897100000,
				CanceledReason:    "No longer needed",
			},
		}

		result := formatListResponse(records, 1, "all")

		assert.Contains(t, result, "🚫 Canceled (No longer needed)")
	})

	t.Run("truncates long cancellation reasons", func(t *testing.T) {
		longReason := "No longer needed - project was canceled by stakeholder team and requirements changed significantly"

		records := []*approval.ApprovalRecord{
			{
				Code:              "A-CAN1",
				Status:            approval.StatusCanceled,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1704897000000,
				CanceledAt:        1704897100000,
				CanceledReason:    longReason,
			},
		}

		result := formatListResponse(records, 1, "all")

		// Should truncate to 37 chars + "..." (exact first 37 characters)
		assert.Contains(t, result, "🚫 Canceled (No longer needed - project was cancel...)")
		assert.NotContains(t, result, longReason)
	})

	t.Run("displays canceled without reason for old records", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{
				Code:              "A-CAN1",
				Status:            approval.StatusCanceled,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1704897000000,
				CanceledAt:        0,
				CanceledReason:    "",
			},
		}

		result := formatListResponse(records, 1, "all")

		// Should show without reason text or parentheses
		assert.Contains(t, result, "🚫 Canceled")
		assert.NotContains(t, result, "🚫 Canceled ()")
	})

	t.Run("respects 20-record limit across all groups", func(t *testing.T) {
		var records []*approval.ApprovalRecord

		// Create 10 pending, 10 decided, 10 canceled (30 total)
		for i := range 10 {
			records = append(records, &approval.ApprovalRecord{
				Code:              fmt.Sprintf("A-PND%d", i),
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         int64(1704897000000 + i*1000),
			})
		}
		for i := range 10 {
			records = append(records, &approval.ApprovalRecord{
				Code:              fmt.Sprintf("A-APP%d", i),
				Status:            approval.StatusApproved,
				RequesterUsername: "charlie",
				ApproverUsername:  "diane",
				CreatedAt:         int64(1704897000000 + i*1000),
				DecidedAt:         int64(1704897100000 + i*1000),
			})
		}
		for i := range 10 {
			records = append(records, &approval.ApprovalRecord{
				Code:              fmt.Sprintf("A-CAN%d", i),
				Status:            approval.StatusCanceled,
				RequesterUsername: "eve",
				ApproverUsername:  "frank",
				CreatedAt:         int64(1704897000000 + i*1000),
				CanceledAt:        int64(1704897200000 + i*1000),
				CanceledReason:    "Reason",
			})
		}

		result := formatListResponse(records, 30, "all")

		// Count record codes in output (each appears once)
		recordCount := 0
		for i := range 10 {
			if strings.Contains(result, fmt.Sprintf("A-PND%d", i)) {
				recordCount++
			}
			if strings.Contains(result, fmt.Sprintf("A-APP%d", i)) {
				recordCount++
			}
			if strings.Contains(result, fmt.Sprintf("A-CAN%d", i)) {
				recordCount++
			}
		}

		assert.LessOrEqual(t, recordCount, 20, "Should display at most 20 records")

		// Should show pagination footer
		assert.Contains(t, result, "Showing")
		assert.Contains(t, result, "of 30 total records")
	})

	t.Run("updates pagination footer correctly", func(t *testing.T) {
		var records []*approval.ApprovalRecord
		for i := range 15 {
			records = append(records, &approval.ApprovalRecord{
				Code:              fmt.Sprintf("A-TST%d", i),
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         int64(1704897000000 + i*1000),
			})
		}

		result := formatListResponse(records, 25, "all")

		// Should show "Showing 15 of 25"
		assert.Contains(t, result, "Showing 15 of 25 total records")
	})

	t.Run("omits footer when all records shown", func(t *testing.T) {
		records := []*approval.ApprovalRecord{
			{
				Code:              "A-TST1",
				Status:            approval.StatusPending,
				RequesterUsername: "alice",
				ApproverUsername:  "bob",
				CreatedAt:         1704897000000,
			},
		}

		result := formatListResponse(records, 1, "all")

		// Should NOT show pagination footer
		assert.NotContains(t, result, "Showing")
		assert.NotContains(t, result, "total records")
	})

	t.Run("displays all four cancellation reasons correctly", func(t *testing.T) {
		reasons := []string{
			"No longer needed",
			"Changed requirements",
			"Created by mistake",
			"Other: Custom explanation text",
		}

		for _, reason := range reasons {
			records := []*approval.ApprovalRecord{
				{
					Code:              "A-CAN1",
					Status:            approval.StatusCanceled,
					RequesterUsername: "alice",
					ApproverUsername:  "bob",
					CreatedAt:         1704897000000,
					CanceledAt:        1704897100000,
					CanceledReason:    reason,
				},
			}

			result := formatListResponse(records, 1, "all")

			assert.Contains(t, result, fmt.Sprintf("🚫 Canceled (%s)", reason),
				"Should display reason: %s", reason)
		}
	})
}

func TestFormatListResponse_UTF8Handling(t *testing.T) {
	testCases := []struct {
		name           string
		reason         string
		expectedOutput string
	}{
		{
			name:           "emoji in reason - truncated",
			reason:         "Changed requirements 🎯 using new approach",
			expectedOutput: "🚫 Canceled (Changed requirements 🎯 using new appr...)",
		},
		{
			name:           "emoji truncation at boundary",
			reason:         "No longer needed - project canceled 🚫🚫 with extra text that makes this very long",
			expectedOutput: "🚫 Canceled (No longer needed - project canceled 🚫...)",
		},
		{
			name:           "CJK characters",
			reason:         "要求が変更されました - 新しいアプローチを使用します",
			expectedOutput: "🚫 Canceled (要求が変更されました - 新しいアプローチを使用します)",
		},
		{
			name:           "accented characters",
			reason:         "Besoin supprimé - exigences modifiées",
			expectedOutput: "🚫 Canceled (Besoin supprimé - exigences modifiées)",
		},
		{
			name:           "CJK text exactly 40 characters - not truncated",
			reason:         "要求が変更されました新しいアプローチを使用しますこれはテストです長いテキストです",
			expectedOutput: "🚫 Canceled (要求が変更されました新しいアプローチを使用しますこれはテストです長いテキストです)",
		},
		{
			name:           "CJK text over 40 characters - truncated",
			reason:         "要求が変更されました新しいアプローチを使用しますこれはテストです長いテキストですさらに追加",
			expectedOutput: "🚫 Canceled (要求が変更されました新しいアプローチを使用しますこれはテストです長いテキス...)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			records := []*approval.ApprovalRecord{
				{
					Code:              "A-CAN1",
					Status:            approval.StatusCanceled,
					RequesterUsername: "alice",
					ApproverUsername:  "bob",
					CreatedAt:         1704897000000,
					CanceledAt:        1704897100000,
					CanceledReason:    tc.reason,
				},
			}

			result := formatListResponse(records, 1, "all")
			assert.Contains(t, result, tc.expectedOutput,
				"Should handle UTF-8 characters correctly")
		})
	}
}

// TestFormatRecordDetail_Verification tests that verification information is displayed correctly
func TestFormatRecordDetail_Verification(t *testing.T) {
	t.Run("shows verification section for verified approved request without comment", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusApproved,
			RequesterID:          "user123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverID:           "approver456",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1704931200000,
			DecidedAt:            1704931300000,
			Verified:             true,
			VerifiedAt:           1704931400000, // Jan 10, 2024
			VerificationComment:  "",
		}

		result := formatRecordDetail(record)

		// Verify section header
		assert.Contains(t, result, "**✅ Verification:**")

		// Verify timestamp is shown
		assert.Contains(t, result, "**Verified:** 2024-01-11")

		// Verify "Verified by" shows requester
		assert.Contains(t, result, "**Verified by:** @alice")

		// Verify no comment section since it's empty
		assert.NotContains(t, result, "**Note:**")
	})

	t.Run("shows verification section for verified approved request with comment", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusApproved,
			RequesterID:          "user123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverID:           "approver456",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1704931200000,
			DecidedAt:            1704931300000,
			Verified:             true,
			VerifiedAt:           1704931400000,
			VerificationComment:  "Deployment completed successfully",
		}

		result := formatRecordDetail(record)

		// Verify section header
		assert.Contains(t, result, "**✅ Verification:**")

		// Verify timestamp is shown
		assert.Contains(t, result, "**Verified:** 2024-01-11")

		// Verify "Verified by" shows requester
		assert.Contains(t, result, "**Verified by:** @alice")

		// Verify comment is shown
		assert.Contains(t, result, "**Note:** Deployment completed successfully")
	})

	t.Run("no verification section for approved request not verified", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			ID:                   "record123",
			Code:                 "A-X7K9Q2",
			Status:               approval.StatusApproved,
			RequesterID:          "user123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverID:           "approver456",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Deploy to production",
			CreatedAt:            1704931200000,
			DecidedAt:            1704931300000,
			Verified:             false,
		}

		result := formatRecordDetail(record)

		// Verification section should NOT be shown
		assert.NotContains(t, result, "**✅ Verification:**")
		assert.NotContains(t, result, "**Verified by:**")
	})

	t.Run("no verification section for non-approved statuses", func(t *testing.T) {
		statuses := []string{
			approval.StatusPending,
			approval.StatusDenied,
			approval.StatusCanceled,
		}

		for _, status := range statuses {
			record := &approval.ApprovalRecord{
				ID:                   "record123",
				Code:                 "A-X7K9Q2",
				Status:               status,
				RequesterID:          "user123",
				RequesterUsername:    "alice",
				RequesterDisplayName: "Alice Smith",
				ApproverID:           "approver456",
				ApproverUsername:     "bob",
				ApproverDisplayName:  "Bob Jones",
				Description:          "Deploy to production",
				CreatedAt:            1704931200000,
				Verified:             true, // Even if verified flag is true
				VerifiedAt:           1704931400000,
			}

			result := formatRecordDetail(record)

			// Verification section should NOT be shown for non-approved status
			assert.NotContains(t, result, "**✅ Verification:**",
				"Status %s should not show verification section", status)
		}
	})
}

// Story 8.1: Integration tests for playbook context detection
// These tests verify that playbook detection integrates correctly with the /approve new flow

// mockPlaybooksClient is a mock implementation of playbooks.ClientInterface
type mockPlaybooksClient struct {
	mock.Mock
}

func (m *mockPlaybooksClient) GetPlaybookRunByChannel(channelID string, requesterUserID string) (*playbooks.PlaybookRun, error) {
	args := m.Called(channelID, requesterUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*playbooks.PlaybookRun), args.Error(1)
}

func (m *mockPlaybooksClient) PostPlaybookStatus(runID string, message string, requesterUserID string) (string, error) {
	args := m.Called(runID, message, requesterUserID)
	return args.String(0), args.Error(1)
}

func (m *mockPlaybooksClient) GetMetrics() playbooks.Metrics {
	args := m.Called()
	return args.Get(0).(playbooks.Metrics)
}

func TestExecuteNew_PlaybookIntegration(t *testing.T) {
	t.Run("playbook detection succeeds - logs context and continues normally", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		playbooksClient := &mockPlaybooksClient{}
		router := NewRouter(api, store, playbooksClient)

		// Mock playbooks client returning a run
		run := &playbooks.PlaybookRun{
			ID:          "run123",
			Name:        "Incident #47 - Database Down",
			Description: "Production database outage",
			OwnerUserID: "user456",
			TeamID:      "team789",
			ChannelID:   "channel012",
			CreateAt:    1705507200000,
			EndAt:       0,
			PlaybookID:  "playbook345",
		}
		playbooksClient.On("GetPlaybookRunByChannel", "channel012", "user123").Return(run, nil)

		// Mock GetConfig to return site URL
		siteURL := "https://mattermost.example.com"
		config := &model.Config{}
		config.ServiceSettings.SiteURL = &siteURL
		api.On("GetConfig").Return(config)

		// Mock OpenInteractiveDialog - approval creation should continue
		api.On("OpenInteractiveDialog", mock.Anything).Return(nil)

		// Mock LogDebug - should log playbook context detection
		var loggedRunID, loggedRunName string
		api.On("LogDebug", "Detected playbook context", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			// Capture logged values
			for i := 1; i < len(args); i += 2 {
				key := args.String(i)
				if key == "run_id" {
					loggedRunID = args.String(i + 1)
				}
				if key == "run_name" {
					loggedRunName = args.String(i + 1)
				}
			}
		}).Return()

		args := &model.CommandArgs{
			Command:   "/approve new",
			UserId:    "user123",
			ChannelId: "channel012",
			TriggerId: "trigger123",
		}

		resp, err := router.Route(args)

		// Verify approval creation succeeds
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify playbook context was logged
		assert.Equal(t, "run123", loggedRunID)
		assert.Equal(t, "Incident #47 - Database Down", loggedRunName)

		api.AssertExpectations(t)
		playbooksClient.AssertExpectations(t)
	})

	t.Run("playbook detection returns 404 - no playbook in channel, continues normally", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		playbooksClient := &mockPlaybooksClient{}
		router := NewRouter(api, store, playbooksClient)

		// Mock playbooks client returning nil (404 - no playbook)
		playbooksClient.On("GetPlaybookRunByChannel", "regular-channel", "user123").Return(nil, nil)

		// Mock GetConfig to return site URL
		siteURL := "https://mattermost.example.com"
		config := &model.Config{}
		config.ServiceSettings.SiteURL = &siteURL
		api.On("GetConfig").Return(config)

		// Mock OpenInteractiveDialog - approval creation should continue
		api.On("OpenInteractiveDialog", mock.Anything).Return(nil)

		args := &model.CommandArgs{
			Command:   "/approve new",
			UserId:    "user123",
			ChannelId: "regular-channel",
			TriggerId: "trigger123",
		}

		resp, err := router.Route(args)

		// Verify approval creation succeeds even without playbook
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		api.AssertExpectations(t)
		playbooksClient.AssertExpectations(t)
	})

	t.Run("playbook detection fails - logs error but approval creation continues (graceful degradation)", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		playbooksClient := &mockPlaybooksClient{}
		router := NewRouter(api, store, playbooksClient)

		// Mock playbooks client returning error
		playbooksClient.On("GetPlaybookRunByChannel", "channel123", "user123").Return(nil, fmt.Errorf("playbooks API returned status 500"))

		// Mock LogWarn - should log the error (now includes user_id)
		var loggedError string
		api.On("LogWarn", "Failed to check for playbook context", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			for i := 1; i < len(args); i += 2 {
				key := args.String(i)
				if key == "error" {
					loggedError = args.String(i + 1)
				}
			}
		}).Return()

		// Mock GetConfig to return site URL
		siteURL := "https://mattermost.example.com"
		config := &model.Config{}
		config.ServiceSettings.SiteURL = &siteURL
		api.On("GetConfig").Return(config)

		// Mock OpenInteractiveDialog - approval creation should STILL continue
		api.On("OpenInteractiveDialog", mock.Anything).Return(nil)

		args := &model.CommandArgs{
			Command:   "/approve new",
			UserId:    "user123",
			ChannelId: "channel123",
			TriggerId: "trigger123",
		}

		resp, err := router.Route(args)

		// Verify approval creation succeeds despite detection error
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify error was logged
		assert.Contains(t, loggedError, "playbooks API returned status 500")

		api.AssertExpectations(t)
		playbooksClient.AssertExpectations(t)
	})

	t.Run("playbooks client is nil - skips detection entirely, approval creation continues", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store, nil) // No playbooks client

		// Mock GetConfig to return site URL
		siteURL := "https://mattermost.example.com"
		config := &model.Config{}
		config.ServiceSettings.SiteURL = &siteURL
		api.On("GetConfig").Return(config)

		// Mock OpenInteractiveDialog - approval creation should continue
		api.On("OpenInteractiveDialog", mock.Anything).Return(nil)

		args := &model.CommandArgs{
			Command:   "/approve new",
			UserId:    "user123",
			ChannelId: "channel123",
			TriggerId: "trigger123",
		}

		resp, err := router.Route(args)

		// Verify approval creation succeeds without playbook detection
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		api.AssertExpectations(t)
		// No playbooks client assertions - was never called
	})
}

// Story 8.2: Tests for formatRecordDetail with playbook context
func TestFormatRecordDetail_PlaybookContext(t *testing.T) {
	t.Run("displays playbook context when present", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			ID:                   "abc123",
			Code:                 "A-TEST1",
			RequesterID:          "req123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverID:           "app456",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test approval",
			Status:               approval.StatusPending,
			CreatedAt:            1704931200000,
			RequestChannelID:     "channel123",
			TeamID:               "team456",
			// Playbook fields
			PlaybookRunID:     "playbook_run_789",
			PlaybookName:      "Incident #47",
			PlaybookChannelID: "pb_channel_123",
			PlaybookPostID:    "post_456",
		}

		result := formatRecordDetail(record)

		// Verify playbook context section exists
		assert.Contains(t, result, "**🎯 Playbook Context:**")
		assert.Contains(t, result, "- Name: Incident #47")
		assert.Contains(t, result, "- Channel ID: pb_channel_123")
		assert.Contains(t, result, "- Status Post ID: post_456")
		assert.Contains(t, result, "- Playbook Run ID: playbook_run_789")
	})

	t.Run("omits playbook context when not present", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			ID:                   "abc123",
			Code:                 "A-TEST2",
			RequesterID:          "req123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverID:           "app456",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test approval",
			Status:               approval.StatusPending,
			CreatedAt:            1704931200000,
			RequestChannelID:     "channel123",
			TeamID:               "team456",
			// Playbook fields empty
		}

		result := formatRecordDetail(record)

		// Verify playbook context section does NOT exist
		assert.NotContains(t, result, "**🎯 Playbook Context:**")
		assert.NotContains(t, result, "Playbook Run ID:")
	})

	t.Run("handles partial playbook context gracefully", func(t *testing.T) {
		record := &approval.ApprovalRecord{
			ID:                   "abc123",
			Code:                 "A-TEST3",
			RequesterID:          "req123",
			RequesterUsername:    "alice",
			RequesterDisplayName: "Alice Smith",
			ApproverID:           "app456",
			ApproverUsername:     "bob",
			ApproverDisplayName:  "Bob Jones",
			Description:          "Test approval",
			Status:               approval.StatusPending,
			CreatedAt:            1704931200000,
			RequestChannelID:     "channel123",
			TeamID:               "team456",
			// Partial playbook fields (no post ID yet - Story 8.3 not done)
			PlaybookRunID:     "playbook_run_789",
			PlaybookName:      "Incident #47",
			PlaybookChannelID: "pb_channel_123",
		}

		result := formatRecordDetail(record)

		// Verify playbook context section exists
		assert.Contains(t, result, "**🎯 Playbook Context:**")
		assert.Contains(t, result, "- Name: Incident #47")
		assert.Contains(t, result, "- Channel ID: pb_channel_123")
		assert.NotContains(t, result, "- Status Post ID:") // Should not appear when empty
		assert.Contains(t, result, "- Playbook Run ID: playbook_run_789")
	})
}
