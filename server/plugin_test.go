package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOnActivate(t *testing.T) {
	t.Run("successfully registers command and initializes store", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.NoError(t, err)
		assert.NotNil(t, p.store, "store should be initialized")
		assert.Equal(t, "bot123", p.botUserID, "bot user ID should be initialized")
		api.AssertExpectations(t)
	})

	t.Run("handles registration failure", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(model.NewAppError("test", "test.error", nil, "", 500))
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register slash command")
	})

	t.Run("handles bot user creation failure", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("", &model.AppError{Message: "bot creation failed"})
		api.On("LogInfo", mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ensure bot user")
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

	t.Run("cancel command opens modal (Story 4.3)", func(t *testing.T) {
		// Story 4.3: Cancel command now opens modal instead of immediate cancellation
		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations for GetByCode
		api.On("KVGet", "approval:code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

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

		// Mock OpenInteractiveDialog - modal should open
		api.On("OpenInteractiveDialog", mock.MatchedBy(func(req model.OpenDialogRequest) bool {
			return req.Dialog.CallbackId == "cancel_approval_record123" &&
				req.Dialog.Title == "Cancel Approval Request"
		})).Return(nil)

		// Mock logging
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2",
			UserId:    "user123",
			ChannelId: "channel123",
			TriggerId: "trigger123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)

		api.AssertExpectations(t)
	})

	t.Run("permission denied for different user", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations
		api.On("KVGet", "approval:code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

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

	t.Run("modal opens for approved approval (validation on submit)", func(t *testing.T) {
		// Story 4.3: Modal opens even for approved approvals
		// Validation happens when modal is submitted
		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations
		api.On("KVGet", "approval:code:A-X7K9Q2").Return([]byte(`"record123"`), nil)

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

		// Mock OpenInteractiveDialog - modal still opens
		api.On("OpenInteractiveDialog", mock.AnythingOfType("model.OpenDialogRequest")).Return(nil)

		// Mock logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-X7K9Q2",
			UserId:    "user123",
			ChannelId: "channel123",
			TriggerId: "trigger123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)

		api.AssertExpectations(t)
	})

	t.Run("approval not found shows error", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV store operations - code not found
		api.On("KVGet", "approval:code:Z-NOTFND").Return(nil, nil)

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

	// Story 4.3: Removed "ephemeral post fallback works" test - no longer applicable
	// since cancel command now opens a modal instead of immediately canceling.
	// Fallback behavior (if needed) would be in modal submission handler, not command handler.

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

	// Story 4.3: Updated test - format validation moved to modal submission
	// Command handler only validates approval exists before opening modal
	t.Run("non-existent code shows not found error", func(t *testing.T) {
		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Mock KV lookup for non-existent code (returns nil = not found)
		api.On("KVGet", "approval:code:A-NOTFND").Return(nil, nil)

		// Mock logging
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-NOTFND",
			UserId:    "user123",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Approval request 'A-NOTFND' not found")
		assert.Contains(t, resp.Text, "Use `/approve list` to see your requests")

		api.AssertExpectations(t)
	})
}

func TestServeHTTP(t *testing.T) {
	t.Run("routes /action to handleAction", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

		p := &Plugin{}
		p.SetAPI(api)

		// Create HTTP request for /action endpoint
		req := httptest.NewRequest("POST", "/action", strings.NewReader(`{}`))
		w := httptest.NewRecorder()

		p.ServeHTTP(nil, w, req)

		// Should return 200 or 400 (not 404)
		assert.NotEqual(t, http.StatusNotFound, w.Code, "should route /action endpoint")
	})

	t.Run("returns 404 for unknown path", func(t *testing.T) {
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)

		req := httptest.NewRequest("POST", "/unknown", strings.NewReader(`{}`))
		w := httptest.NewRecorder()

		p.ServeHTTP(nil, w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleAction(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		approvalID     string
		action         string
		userID         string
		approverID     string
		recordStatus   string
		setupMocks     func(*plugintest.API, *approval.Service)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "approve button opens modal for pending approval",
			requestBody: `{
				"user_id": "approver456",
				"trigger_id": "trigger123",
				"context": {
					"approval_id": "record123",
					"action": "approve"
				}
			}`,
			approvalID:   "record123",
			action:       "approve",
			userID:       "approver456",
			approverID:   "approver456",
			recordStatus: "pending",
			setupMocks: func(api *plugintest.API, svc *approval.Service) {
				api.On("OpenInteractiveDialog", mock.MatchedBy(func(req model.OpenDialogRequest) bool {
					return req.Dialog.Title == "Confirm Approval" &&
						strings.Contains(req.Dialog.IntroductionText, "Confirm you are approving:")
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "deny button opens modal for pending approval",
			requestBody: `{
				"user_id": "approver456",
				"trigger_id": "trigger123",
				"context": {
					"approval_id": "record123",
					"action": "deny"
				}
			}`,
			approvalID:   "record123",
			action:       "deny",
			userID:       "approver456",
			approverID:   "approver456",
			recordStatus: "pending",
			setupMocks: func(api *plugintest.API, svc *approval.Service) {
				api.On("OpenInteractiveDialog", mock.MatchedBy(func(req model.OpenDialogRequest) bool {
					return req.Dialog.Title == "Confirm Denial" &&
						strings.Contains(req.Dialog.IntroductionText, "Confirm you are denying:")
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-approver rejected with permission denied",
			requestBody: `{
				"user_id": "unauthorized789",
				"trigger_id": "trigger123",
				"context": {
					"approval_id": "record123",
					"action": "approve"
				}
			}`,
			approvalID:     "record123",
			action:         "approve",
			userID:         "unauthorized789",
			approverID:     "approver456",
			recordStatus:   "pending",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Permission denied",
		},
		{
			name: "already approved request rejected",
			requestBody: `{
				"user_id": "approver456",
				"trigger_id": "trigger123",
				"context": {
					"approval_id": "record123",
					"action": "approve"
				}
			}`,
			approvalID:     "record123",
			action:         "approve",
			userID:         "approver456",
			approverID:     "approver456",
			recordStatus:   "approved",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Decision already recorded",
		},
		{
			name: "canceled request rejected",
			requestBody: `{
				"user_id": "approver456",
				"trigger_id": "trigger123",
				"context": {
					"approval_id": "record123",
					"action": "approve"
				}
			}`,
			approvalID:     "record123",
			action:         "approve",
			userID:         "approver456",
			approverID:     "approver456",
			recordStatus:   "canceled",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Decision already recorded",
		},
		{
			name:           "invalid JSON returns error",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request",
		},
		{
			name: "approval not found returns error",
			requestBody: `{
				"user_id": "approver456",
				"trigger_id": "trigger123",
				"context": {
					"approval_id": "notfound",
					"action": "approve"
				}
			}`,
			approvalID:     "notfound",
			userID:         "approver456",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Approval not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
			api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)
			api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			// Setup service mocks
			if tt.approvalID != "" && tt.approvalID != "notfound" {
				// Mock GetByID for service
				recordJSON := fmt.Sprintf(`{
					"id": "%s",
					"code": "A-X7K9Q2",
					"requesterId": "requester123",
					"requesterUsername": "requester",
					"requesterDisplayName": "Requester User",
					"approverId": "%s",
					"approverUsername": "approver",
					"approverDisplayName": "Approver User",
					"description": "Test approval request",
					"status": "%s",
					"createdAt": 1704931200000,
					"decidedAt": 0,
					"schemaVersion": 1
				}`, tt.approvalID, tt.approverID, tt.recordStatus)
				api.On("KVGet", fmt.Sprintf("approval:record:%s", tt.approvalID)).Return([]byte(recordJSON), nil)
			} else if tt.approvalID == "notfound" {
				api.On("KVGet", fmt.Sprintf("approval:record:%s", tt.approvalID)).Return(nil, nil)
			}

			// Setup additional mocks if provided
			if tt.setupMocks != nil {
				tt.setupMocks(api, nil)
			}

			p := &Plugin{}
			p.SetAPI(api)
			err := p.OnActivate()
			assert.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/action", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			p.ServeHTTP(nil, w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response model.PostActionIntegrationResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response.EphemeralText, tt.expectedError)
			}

			api.AssertExpectations(t)
		})
	}
}
