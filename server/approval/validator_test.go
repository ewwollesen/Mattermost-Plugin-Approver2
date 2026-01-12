package approval

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
)

func TestValidateApprovalRecord(t *testing.T) {
	tests := []struct {
		name    string
		record  *ApprovalRecord
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid record",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz", // 26 chars like real Mattermost ID
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345", // 26 chars
				ApproverID:    "a1234567890123456789012345", // 26 chars
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: false,
		},
		{
			name:    "nil record",
			record:  nil,
			wantErr: true,
			errMsg:  "cannot be nil",
		},
		{
			name: "missing ID",
			record: &ApprovalRecord{
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345",
				ApproverID:    "a1234567890123456789012345",
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "ID is required",
		},
		{
			name: "missing code",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				RequesterID:   "r1234567890123456789012345",
				ApproverID:    "a1234567890123456789012345",
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "missing requester ID",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				Code:          "A-X7K9Q2",
				ApproverID:    "a1234567890123456789012345",
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "requester ID is required",
		},
		{
			name: "missing approver ID",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345",
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "approver ID is required",
		},
		{
			name: "missing description",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345",
				ApproverID:    "a1234567890123456789012345",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "invalid status",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345",
				ApproverID:    "a1234567890123456789012345",
				Description:   "Test",
				Status:        "invalid",
				CreatedAt:     1704931200000,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "negative timestamp",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345",
				ApproverID:    "a1234567890123456789012345",
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     -1,
				SchemaVersion: 1,
			},
			wantErr: true,
			errMsg:  "timestamp must be positive",
		},
		{
			name: "invalid schema version",
			record: &ApprovalRecord{
				ID:            "abcdefghijklmnopqrstuvwxyz",
				Code:          "A-X7K9Q2",
				RequesterID:   "r1234567890123456789012345",
				ApproverID:    "a1234567890123456789012345",
				Description:   "Test",
				Status:        StatusPending,
				CreatedAt:     1704931200000,
				SchemaVersion: 0,
			},
			wantErr: true,
			errMsg:  "schema version must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateApprovalRecord(tt.record)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{StatusPending, true},
		{StatusApproved, true},
		{StatusDenied, true},
		{StatusCanceled, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid description",
			description: "Please approve this deployment to production",
			expectError: false,
		},
		{
			name:        "exactly 1000 characters",
			description: string(make([]byte, 1000)),
			expectError: false,
		},
		{
			name:        "1001 characters fails",
			description: string(make([]byte, 1001)),
			expectError: true,
			errorMsg:    "max 1000",
		},
		{
			name:        "empty description fails",
			description: "",
			expectError: true,
			errorMsg:    "required",
		},
		{
			name:        "whitespace only description fails",
			description: "   \t\n   ",
			expectError: true,
			errorMsg:    "required",
		},
		{
			name:        "Unicode characters within limit",
			description: "请批准此部署 🚀",
			expectError: false,
		},
		{
			name:        "error includes character count",
			description: string(make([]byte, 1500)),
			expectError: true,
			errorMsg:    "1500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.description)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateApprover(t *testing.T) {
	t.Run("valid user passes validation", func(t *testing.T) {
		api := &plugintest.API{}
		expectedUser := &model.User{
			Id:       "user123",
			Username: "alice",
			DeleteAt: 0, // Active user
		}
		api.On("GetUser", "user123").Return(expectedUser, nil)

		user, err := ValidateApprover("user123", api)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, user)
		api.AssertExpectations(t)
	})

	t.Run("non-existent user returns error", func(t *testing.T) {
		api := &plugintest.API{}
		appErr := model.NewAppError("GetUser", "api.user.get.app_error", nil, "", 404)
		api.On("GetUser", "invalid").Return(nil, appErr)

		user, err := ValidateApprover("invalid", api)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to validate approver")
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("deleted user returns error", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("GetUser", "deleted").Return(&model.User{
			Id:       "deleted",
			Username: "deactivated",
			DeleteAt: 1234567890000, // Deleted user (non-zero DeleteAt)
		}, nil)

		user, err := ValidateApprover("deleted", api)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "not a valid user")
		assert.Contains(t, err.Error(), "active")
	})

	t.Run("API error is properly wrapped", func(t *testing.T) {
		api := &plugintest.API{}
		appErr := model.NewAppError("GetUser", "api.user.get.app_error", nil, "database connection failed", 500)
		api.On("GetUser", "user456").Return(nil, appErr)

		user, err := ValidateApprover("user456", api)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to validate approver")
		assert.Contains(t, err.Error(), "user456")
	})
}
