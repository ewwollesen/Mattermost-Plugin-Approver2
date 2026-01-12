package approval

import (
	"testing"

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
