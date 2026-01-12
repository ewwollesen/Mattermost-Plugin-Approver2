package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOnActivate(t *testing.T) {
	t.Run("successfully registers command and initializes store", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.NoError(t, err)
		assert.NotNil(t, p.store, "store should be initialized")
		api.AssertExpectations(t)
	})

	t.Run("handles registration failure", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(model.NewAppError("test", "test.error", nil, "", 500))
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register slash command")
	})
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		expectedContains string
		expectedType     string
	}{
		{
			name:             "no subcommand shows help",
			command:          "/approve",
			expectedContains: "Available Commands",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
		{
			name:             "help subcommand shows help",
			command:          "/approve help",
			expectedContains: "Available Commands",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
		{
			name:             "unknown subcommand shows error",
			command:          "/approve invalid",
			expectedContains: "Unknown command",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
		{
			name:             "unknown subcommand suggests valid commands",
			command:          "/approve foo",
			expectedContains: "Valid commands",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{}
			p.SetAPI(api)

			args := &model.CommandArgs{
				Command: tt.command,
			}

			resp, appErr := p.ExecuteCommand(nil, args)
			assert.Nil(t, appErr)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedType, resp.ResponseType)
			assert.Contains(t, resp.Text, tt.expectedContains)
		})
	}
}

func TestHandleCancelCommand(t *testing.T) {
	t.Run("missing approval code shows usage", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/approve cancel",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "Usage: /approve cancel <APPROVAL_ID>")
	})

	t.Run("successful cancellation shows confirmation", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin registration
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations for GetByCode
		api.On("KVGet", "approval_code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

		// Mock record retrieval
		recordJSON := `{
			"id": "record123",
			"code": "A-X7K9Q2",
			"requesterId": "user123",
			"status": "pending",
			"createdAt": 1704931200000,
			"decidedAt": 0,
			"schemaVersion": 1
		}`
		api.On("KVGet", "approval:record:record123").Return([]byte(recordJSON), nil)

		// Mock save approval (will update status to canceled)
		api.On("KVSet", "approval:record:record123", mock.Anything).Return(nil)
		api.On("KVSet", "approval_code:A-X7K9Q2", mock.Anything).Return(nil)

		// Mock ephemeral post
		api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
			return post.ChannelId == "channel123" &&
				post.Message == "✅ Approval request `A-X7K9Q2` has been canceled."
		})).Return(&model.Post{})

		// Mock logging
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)

		api.AssertExpectations(t)
	})

	t.Run("permission denied for different user", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin registration
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations
		api.On("KVGet", "approval_code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

		recordJSON := `{
			"id": "record123",
			"code": "A-X7K9Q2",
			"requesterId": "user123",
			"status": "pending",
			"createdAt": 1704931200000,
			"decidedAt": 0,
			"schemaVersion": 1
		}`
		api.On("KVGet", "approval:record:record123").Return([]byte(recordJSON), nil)

		// Mock error logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2",
			UserId:    "user456", // Different user
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Permission denied")
		assert.Contains(t, resp.Text, "only cancel your own approval requests")

		api.AssertExpectations(t)
	})

	t.Run("cannot cancel approved approval", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin registration
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations
		api.On("KVGet", "approval_code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

		recordJSON := `{
			"id": "record123",
			"code": "A-X7K9Q2",
			"requesterId": "user123",
			"status": "approved",
			"createdAt": 1704931200000,
			"decidedAt": 1704931300000,
			"schemaVersion": 1
		}`
		api.On("KVGet", "approval:record:record123").Return([]byte(recordJSON), nil)

		// Mock logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Cannot cancel")
		assert.Contains(t, resp.Text, "Status is already approved")

		api.AssertExpectations(t)
	})

	t.Run("approval not found shows error", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin registration
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations - code not found
		api.On("KVGet", "approval_code:Z-NOTFND").Return(nil, nil)

		// Mock logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel Z-NOTFND",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Approval request 'Z-NOTFND' not found")
		assert.Contains(t, resp.Text, "Use `/approve list` to see your requests")

		api.AssertExpectations(t)
	})

	t.Run("ephemeral post fallback works", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin registration
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations
		api.On("KVGet", "approval_code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

		recordJSON := `{
			"id": "record123",
			"code": "A-X7K9Q2",
			"requesterId": "user123",
			"status": "pending",
			"createdAt": 1704931200000,
			"decidedAt": 0,
			"schemaVersion": 1
		}`
		api.On("KVGet", "approval:record:record123").Return([]byte(recordJSON), nil)

		// Mock save operations
		api.On("KVSet", "approval:record:record123", mock.Anything).Return(nil)
		api.On("KVSet", "approval_code:A-X7K9Q2", mock.Anything).Return(nil)

		// Mock ephemeral post failure (returns nil)
		api.On("SendEphemeralPost", "user123", mock.Anything).Return(nil)

		// Mock fallback to CreatePost
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "user123" &&
				post.ChannelId == "channel123" &&
				post.Message == "✅ Approval request `A-X7K9Q2` has been canceled."
		})).Return(&model.Post{}, nil)

		// Mock logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)

		api.AssertExpectations(t)
	})

	t.Run("extra arguments rejected", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2 extra-arg",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "Usage: /approve cancel <APPROVAL_ID>")
		assert.Contains(t, resp.Text, "Too many arguments")
	})

	t.Run("invalid code format shows error", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin registration
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel invalid-format",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Invalid approval code format")
		assert.Contains(t, resp.Text, "Expected format like 'A-X7K9Q2'")

		api.AssertExpectations(t)
	})
}
