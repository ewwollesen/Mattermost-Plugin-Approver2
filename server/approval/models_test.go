package approval

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApprovalRecord_JSONSerialization(t *testing.T) {
	record := &ApprovalRecord{
		ID:                   "abcdefghijklmnopqrstuvwxyz",
		Code:                 "A-X7K9Q2",
		RequesterID:          "requester123",
		RequesterUsername:    "alice",
		RequesterDisplayName: "Alice Smith",
		ApproverID:           "approver456",
		ApproverUsername:     "bob",
		ApproverDisplayName:  "Bob Jones",
		Description:          "Please approve deployment",
		Status:               StatusPending,
		DecisionComment:      "",
		CreatedAt:            1704931200000,
		DecidedAt:            0,
		RequestChannelID:     "channel789",
		TeamID:               "team012",
		NotificationSent:     false,
		OutcomeNotified:      false,
		Verified:             false,
		VerifiedAt:           0,
		VerificationComment:  "",
		SchemaVersion:        1,
	}

	// Marshal to JSON
	data, err := json.Marshal(record)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back to struct
	var decoded ApprovalRecord
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, record.ID, decoded.ID)
	assert.Equal(t, record.Code, decoded.Code)
	assert.Equal(t, record.RequesterID, decoded.RequesterID)
	assert.Equal(t, record.ApproverID, decoded.ApproverID)
	assert.Equal(t, record.Description, decoded.Description)
	assert.Equal(t, record.Status, decoded.Status)
	assert.Equal(t, record.CreatedAt, decoded.CreatedAt)
	assert.Equal(t, record.Verified, decoded.Verified)
	assert.Equal(t, record.VerifiedAt, decoded.VerifiedAt)
	assert.Equal(t, record.VerificationComment, decoded.VerificationComment)
	assert.Equal(t, record.SchemaVersion, decoded.SchemaVersion)
}

func TestNewApprovalRecord(t *testing.T) {
	// Use mock store that always returns nil (no collision)
	store := NewMockStorer()

	record, err := NewApprovalRecord(
		store,
		"req123", "alice", "Alice Smith",
		"app456", "bob", "Bob Jones",
		"Test approval",
		"channel789",
		"team012",
	)

	require.NoError(t, err)
	require.NotNil(t, record)
	assert.NotEmpty(t, record.ID)
	assert.NotEmpty(t, record.Code)
	assert.Equal(t, "req123", record.RequesterID)
	assert.Equal(t, "app456", record.ApproverID)
	assert.Equal(t, "Test approval", record.Description)
	assert.Equal(t, StatusPending, record.Status)
	assert.Greater(t, record.CreatedAt, int64(0))
	assert.Equal(t, int64(0), record.DecidedAt)
	assert.Equal(t, 1, record.SchemaVersion)
	assert.False(t, record.NotificationSent)
	assert.False(t, record.OutcomeNotified)
	assert.False(t, record.Verified)
	assert.Equal(t, int64(0), record.VerifiedAt)
	assert.Equal(t, "", record.VerificationComment)
}

func TestNewApprovalRecord_WithCollisions(t *testing.T) {
	t.Run("retries on collision and succeeds", func(t *testing.T) {
		store := NewMockStorer()
		// Simulate 2 collisions then success
		store.AddCode("A-TEST01")
		store.AddCode("A-TEST02")

		record, err := NewApprovalRecord(
			store,
			"req123", "alice", "Alice Smith",
			"app456", "bob", "Bob Jones",
			"Test approval",
			"channel789",
			"team012",
		)

		require.NoError(t, err)
		require.NotNil(t, record)
		assert.NotEmpty(t, record.Code)
		// Code should be different from the collision codes
		assert.NotEqual(t, "A-TEST01", record.Code)
		assert.NotEqual(t, "A-TEST02", record.Code)
	})

	t.Run("fails after 5 collisions", func(t *testing.T) {
		// Mock store that always returns collision
		store := &MockStorerAlwaysCollision{}

		record, err := NewApprovalRecord(
			store,
			"req123", "alice", "Alice Smith",
			"app456", "bob", "Bob Jones",
			"Test approval",
			"channel789",
			"team012",
		)

		assert.Error(t, err)
		assert.Nil(t, record)
		assert.ErrorIs(t, err, ErrCodeGenerationFailed)
	})
}

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	assert.Len(t, id, 26, "ID should be 26 characters")
	assert.NotEmpty(t, id)
}
