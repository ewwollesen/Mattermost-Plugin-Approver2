package command

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
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

func TestRoute(t *testing.T) {
	api := &plugintest.API{}
	store := &mockStore{}
	router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
				ID:                 "id1",
				Code:               "A-FAIL1",
				Status:             approval.StatusPending,
				NotificationSent:   false,
				RequesterUsername:  "alice",
				ApproverUsername:   "bob",
				CreatedAt:          1641024000000, // 2022-01-01 12:00:00
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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

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
		router := NewRouter(api, store)

		// Mock store to return empty slice
		store.On("GetUserApprovals", "user123").Return([]*approval.ApprovalRecord{}, nil)

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "No approval records found")
		assert.Contains(t, resp.Text, "/approve new")

		store.AssertExpectations(t)
	})

	t.Run("single record - verifies formatting and status icons", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "**Your Approval Records:**")
		assert.Contains(t, resp.Text, "**A-ABC123**")
		assert.Contains(t, resp.Text, "⏳ Pending")
		assert.Contains(t, resp.Text, "@alice")
		assert.Contains(t, resp.Text, "@bob")
		assert.Contains(t, resp.Text, "2024-01-11")

		store.AssertExpectations(t)
	})

	t.Run("multiple records - verifies sorting by most recent first", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Verify records appear in order (newest first)
		assert.Contains(t, resp.Text, "A-NEW")
		assert.Contains(t, resp.Text, "A-MID")
		assert.Contains(t, resp.Text, "A-OLD")

		// Verify A-NEW appears before A-OLD in output
		newPos := strings.Index(resp.Text, "A-NEW")
		oldPos := strings.Index(resp.Text, "A-OLD")
		assert.Less(t, newPos, oldPos, "Newest record should appear before oldest")

		store.AssertExpectations(t)
	})

	t.Run("pagination - shows only first 20 records with footer", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

		// Create 25 records
		records := make([]*approval.ApprovalRecord, 25)
		for i := 0; i < 25; i++ {
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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Should show first 20 records
		assert.Contains(t, resp.Text, "A-REC00")
		assert.Contains(t, resp.Text, "A-REC19")
		// Should NOT show 21st record
		assert.NotContains(t, resp.Text, "A-REC20")
		assert.NotContains(t, resp.Text, "A-REC24")

		// Should show pagination footer with new format
		assert.Contains(t, resp.Text, "Showing 20 of 25 total records (most recent first)")
		assert.Contains(t, resp.Text, "/approve get")

		store.AssertExpectations(t)
	})

	t.Run("access control - user sees only their records", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "A-USER1")

		// Verify store was called with correct user ID
		store.AssertCalled(t, "GetUserApprovals", "user123")
		store.AssertExpectations(t)
	})

	t.Run("requester records - user sees records where they are requester", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "A-REQ1")
		assert.Contains(t, resp.Text, "@alice") // Requester

		store.AssertExpectations(t)
	})

	t.Run("approver records - user sees records where they are approver", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "A-APP1")
		assert.Contains(t, resp.Text, "@alice") // Approver

		store.AssertExpectations(t)
	})

	t.Run("mixed records - user sees both requester and approver records", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "A-REQ")
		assert.Contains(t, resp.Text, "A-APP")

		store.AssertExpectations(t)
	})

	t.Run("KV store error handling - returns error message", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

		// Mock store to return error
		storeErr := fmt.Errorf("KV store connection failed")
		store.On("GetUserApprovals", "user123").Return(nil, storeErr)

		// Mock LogError call
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Failed to retrieve approval records")
		assert.Contains(t, resp.Text, "Please try again")

		api.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("all status types - verifies correct icons", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify status icons
		assert.Contains(t, resp.Text, "✅ Approved")
		assert.Contains(t, resp.Text, "❌ Denied")
		assert.Contains(t, resp.Text, "⏳ Pending")
		assert.Contains(t, resp.Text, "🚫 Canceled")

		store.AssertExpectations(t)
	})

	t.Run("timestamp formatting - verifies date format", func(t *testing.T) {
		api := &plugintest.API{}
		store := &mockStore{}
		router := NewRouter(api, store)

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

		args := &model.CommandArgs{
			Command: "/approve list",
			UserId:  "user123",
		}

		resp, err := router.Route(args)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify timestamp format: YYYY-MM-DD HH:MM
		assert.Contains(t, resp.Text, "2024-01-10 14:30")

		store.AssertExpectations(t)
	})
}
