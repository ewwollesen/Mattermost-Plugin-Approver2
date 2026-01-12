package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoute(t *testing.T) {
	api := &plugintest.API{}
	router := NewRouter(api)

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
		router := NewRouter(api)

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
		router := NewRouter(api)

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
		router := NewRouter(api)

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
		router := NewRouter(api)

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
		router := NewRouter(api)

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
