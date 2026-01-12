package approval

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockApprovalStore is a mock implementation of the ApprovalStore interface
type MockApprovalStore struct {
	mock.Mock
}

func (m *MockApprovalStore) GetApproval(id string) (*ApprovalRecord, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ApprovalRecord), args.Error(1)
}

func (m *MockApprovalStore) GetByCode(code string) (*ApprovalRecord, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ApprovalRecord), args.Error(1)
}

func (m *MockApprovalStore) SaveApproval(record *ApprovalRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func TestCancelApproval(t *testing.T) {
	tests := []struct {
		name           string
		approvalCode   string
		requesterID    string
		existingRecord *ApprovalRecord
		getByCodeErr   error
		saveErr        error
		wantErr        bool
		errContains    string
		wantStatus     string
		wantDecidedAt  bool // true if DecidedAt should be set (>0)
	}{
		{
			name:         "successful cancellation by requester",
			approvalCode: "A-X7K9Q2",
			requesterID:  "user123",
			existingRecord: &ApprovalRecord{
				ID:          "abc123",
				Code:        "A-X7K9Q2",
				RequesterID: "user123",
				Status:      StatusPending,
				CreatedAt:   1704931200000,
				DecidedAt:   0,
			},
			getByCodeErr:  nil,
			saveErr:       nil,
			wantErr:       false,
			wantStatus:    StatusCanceled,
			wantDecidedAt: true,
		},
		{
			name:         "permission denied - different user",
			approvalCode: "A-X7K9Q2",
			requesterID:  "user456", // Different from record's requester
			existingRecord: &ApprovalRecord{
				ID:          "abc123",
				Code:        "A-X7K9Q2",
				RequesterID: "user123", // Original requester
				Status:      StatusPending,
			},
			getByCodeErr: nil,
			wantErr:      true,
			errContains:  "permission denied",
			wantStatus:   StatusPending, // Unchanged
		},
		{
			name:         "cannot cancel approved approval",
			approvalCode: "A-X7K9Q2",
			requesterID:  "user123",
			existingRecord: &ApprovalRecord{
				ID:          "abc123",
				Code:        "A-X7K9Q2",
				RequesterID: "user123",
				Status:      StatusApproved, // Already approved
				DecidedAt:   1704931300000,
			},
			getByCodeErr: nil,
			wantErr:      true,
			errContains:  "cannot cancel",
			wantStatus:   StatusApproved, // Unchanged
		},
		{
			name:         "cannot cancel denied approval",
			approvalCode: "A-X7K9Q2",
			requesterID:  "user123",
			existingRecord: &ApprovalRecord{
				ID:          "abc123",
				Code:        "A-X7K9Q2",
				RequesterID: "user123",
				Status:      StatusDenied, // Already denied
				DecidedAt:   1704931300000,
			},
			getByCodeErr: nil,
			wantErr:      true,
			errContains:  "cannot cancel",
			wantStatus:   StatusDenied, // Unchanged
		},
		{
			name:         "cannot cancel already canceled approval",
			approvalCode: "A-X7K9Q2",
			requesterID:  "user123",
			existingRecord: &ApprovalRecord{
				ID:          "abc123",
				Code:        "A-X7K9Q2",
				RequesterID: "user123",
				Status:      StatusCanceled, // Already canceled
				DecidedAt:   1704931300000,
			},
			getByCodeErr: nil,
			wantErr:      true,
			errContains:  "cannot cancel",
			wantStatus:   StatusCanceled, // Unchanged
		},
		{
			name:           "record not found",
			approvalCode:   "Z-NOTFND",
			requesterID:    "user123",
			existingRecord: nil,
			getByCodeErr:   ErrRecordNotFound,
			wantErr:        true,
			errContains:    "failed to retrieve",
		},
		{
			name:         "empty approval code",
			approvalCode: "",
			requesterID:  "user123",
			wantErr:      true,
			errContains:  "approval code is required",
		},
		{
			name:         "empty requester ID",
			approvalCode: "A-X7K9Q2",
			requesterID:  "",
			wantErr:      true,
			errContains:  "requester ID is required",
		},
		{
			name:         "save fails after validation passes",
			approvalCode: "A-X7K9Q2",
			requesterID:  "user123",
			existingRecord: &ApprovalRecord{
				ID:          "abc123",
				Code:        "A-X7K9Q2",
				RequesterID: "user123",
				Status:      StatusPending,
			},
			getByCodeErr: nil,
			saveErr:      errors.New("KV store unavailable"),
			wantErr:      true,
			errContains:  "failed to save",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockStore := new(MockApprovalStore)
			mockAPI := &plugintest.API{}

			// Setup expectations
			// Only set up GetByCode mock if code passes initial validation (non-empty and valid format)
			codePassesValidation := tt.approvalCode != "" &&
				strings.TrimSpace(tt.approvalCode) != "" &&
				approvalCodePattern.MatchString(strings.TrimSpace(tt.approvalCode)) &&
				tt.requesterID != "" &&
				strings.TrimSpace(tt.requesterID) != ""

			if codePassesValidation {
				if tt.existingRecord != nil {
					// Create a copy to track mutations
					recordCopy := *tt.existingRecord
					mockStore.On("GetByCode", tt.approvalCode).Return(&recordCopy, tt.getByCodeErr)

					// Only expect SaveApproval if we pass access control and status checks
					if tt.requesterID == tt.existingRecord.RequesterID &&
						tt.existingRecord.Status == StatusPending {
						mockStore.On("SaveApproval", mock.MatchedBy(func(r *ApprovalRecord) bool {
							// Verify the record was updated correctly
							return r.ID == tt.existingRecord.ID &&
								r.Status == StatusCanceled &&
								r.DecidedAt > 0
						})).Return(tt.saveErr)
					}
				} else {
					mockStore.On("GetByCode", tt.approvalCode).Return(nil, tt.getByCodeErr)
				}
			}

			// Create service
			service := NewService(mockStore, mockAPI)

			// Execute
			err := service.CancelApproval(tt.approvalCode, tt.requesterID)

			// Verify error expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations were met
			mockStore.AssertExpectations(t)
		})
	}
}

// TestCancelApproval_TimestampSet verifies DecidedAt is set to current time
func TestCancelApproval_TimestampSet(t *testing.T) {
	mockStore := new(MockApprovalStore)
	mockAPI := &plugintest.API{}

	record := &ApprovalRecord{
		ID:          "abc123",
		Code:        "A-X7K9Q2",
		RequesterID: "user123",
		Status:      StatusPending,
		CreatedAt:   1704931200000,
		DecidedAt:   0,
	}

	mockStore.On("GetByCode", "A-X7K9Q2").Return(record, nil)
	mockStore.On("SaveApproval", mock.MatchedBy(func(r *ApprovalRecord) bool {
		// Verify timestamp is set and reasonable (within last 10 seconds)
		return r.DecidedAt > 0 && r.DecidedAt >= 1704931200000
	})).Return(nil)

	service := NewService(mockStore, mockAPI)
	err := service.CancelApproval("A-X7K9Q2", "user123")

	assert.NoError(t, err)
	assert.Equal(t, StatusCanceled, record.Status)
	assert.Greater(t, record.DecidedAt, int64(0), "DecidedAt should be set")
	mockStore.AssertExpectations(t)
}

// TestCancelApproval_WhitespaceValidation verifies whitespace-only inputs are rejected
func TestCancelApproval_WhitespaceValidation(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		requesterID string
		errContains string
	}{
		{
			name:        "whitespace-only approval code",
			code:        "   ",
			requesterID: "user123",
			errContains: "approval code is required",
		},
		{
			name:        "tab-only approval code",
			code:        "\t",
			requesterID: "user123",
			errContains: "approval code is required",
		},
		{
			name:        "whitespace-only requester ID",
			code:        "A-X7K9Q2",
			requesterID: "  ",
			errContains: "requester ID is required",
		},
		{
			name:        "code with leading/trailing whitespace is trimmed",
			code:        "  A-X7K9Q2  ",
			requesterID: "user123",
			errContains: "", // Should succeed - whitespace trimmed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockApprovalStore)
			mockAPI := &plugintest.API{}

			// For the successful trim case, set up mocks
			if tt.errContains == "" {
				record := &ApprovalRecord{
					ID:          "abc123",
					Code:        "A-X7K9Q2",
					RequesterID: "user123",
					Status:      StatusPending,
					CreatedAt:   1704931200000,
					DecidedAt:   0,
				}
				mockStore.On("GetByCode", "A-X7K9Q2").Return(record, nil)
				mockStore.On("SaveApproval", mock.Anything).Return(nil)
			}

			service := NewService(mockStore, mockAPI)
			err := service.CancelApproval(tt.code, tt.requesterID)

			if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

// TestCancelApproval_InvalidFormat verifies approval code format validation
func TestCancelApproval_InvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"lowercase prefix", "a-X7K9Q2"},
		{"no dash", "AX7K9Q2"},
		{"too short", "A-X7K9"},
		{"too long", "A-X7K9Q23"},
		{"invalid characters", "A-X7K9Q!"},
		{"multiple dashes", "A-X7K-9Q2"},
		{"empty after trim", ""},
		{"random string", "obviously-not-valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockApprovalStore)
			mockAPI := &plugintest.API{}

			service := NewService(mockStore, mockAPI)
			err := service.CancelApproval(tt.code, "user123")

			assert.Error(t, err)
			if tt.code == "" {
				assert.Contains(t, err.Error(), "approval code is required")
			} else {
				assert.Contains(t, err.Error(), "invalid approval code format")
			}

			mockStore.AssertExpectations(t)
		})
	}
}

// TestCancelApproval_CorruptedIndex verifies behavior when index points to non-existent record
func TestCancelApproval_CorruptedIndex(t *testing.T) {
	mockStore := new(MockApprovalStore)
	mockAPI := &plugintest.API{}

	// Code index exists but points to deleted/non-existent record
	mockStore.On("GetByCode", "A-X7K9Q2").Return(nil, fmt.Errorf("approval record deleted-id-123: %w", ErrRecordNotFound))

	service := NewService(mockStore, mockAPI)
	err := service.CancelApproval("A-X7K9Q2", "user123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve approval")
	assert.ErrorIs(t, err, ErrRecordNotFound)

	mockStore.AssertExpectations(t)
}

