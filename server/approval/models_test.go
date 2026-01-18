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

// TestBackwardCompatibility_V1Record tests that v1.0 records (without playbook fields) deserialize correctly
func TestBackwardCompatibility_V1Record(t *testing.T) {
	// Simulate a v1.0 record JSON (no playbook fields)
	v1JSON := `{
		"id": "abcdefghijklmnopqrstuvwxyz",
		"code": "A-X7K9Q2",
		"requesterId": "req123",
		"requesterUsername": "alice",
		"requesterDisplayName": "Alice Smith",
		"approverId": "app456",
		"approverUsername": "bob",
		"approverDisplayName": "Bob Jones",
		"description": "Test approval",
		"status": "pending",
		"createdAt": 1704931200000,
		"decidedAt": 0,
		"requestChannelId": "channel789",
		"teamId": "team012",
		"notificationSent": false,
		"outcomeNotified": false,
		"verified": false,
		"verifiedAt": 0,
		"schemaVersion": 1
	}`

	var record ApprovalRecord
	err := json.Unmarshal([]byte(v1JSON), &record)

	require.NoError(t, err, "v1.0 record should deserialize without error")
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", record.ID)
	assert.Equal(t, "A-X7K9Q2", record.Code)
	assert.Equal(t, "pending", record.Status)

	// Playbook fields should be empty (not error)
	assert.Empty(t, record.PlaybookRunID, "v1.0 record should have empty PlaybookRunID")
	assert.Empty(t, record.PlaybookName, "v1.0 record should have empty PlaybookName")
	assert.Empty(t, record.PlaybookChannelID, "v1.0 record should have empty PlaybookChannelID")
	assert.Empty(t, record.PlaybookPostID, "v1.0 record should have empty PlaybookPostID")
}

// TestBackwardCompatibility_V2Record tests that v2.0 records (with playbook fields) serialize and deserialize correctly
func TestBackwardCompatibility_V2Record(t *testing.T) {
	// Simulate a v2.0 record JSON (with playbook fields)
	v2JSON := `{
		"id": "abcdefghijklmnopqrstuvwxyz",
		"code": "A-X7K9Q2",
		"requesterId": "req123",
		"requesterUsername": "alice",
		"requesterDisplayName": "Alice Smith",
		"approverId": "app456",
		"approverUsername": "bob",
		"approverDisplayName": "Bob Jones",
		"description": "Test approval",
		"status": "pending",
		"createdAt": 1704931200000,
		"decidedAt": 0,
		"requestChannelId": "channel789",
		"teamId": "team012",
		"notificationSent": false,
		"outcomeNotified": false,
		"verified": false,
		"verifiedAt": 0,
		"schemaVersion": 1,
		"playbookRunId": "playbook123",
		"playbookName": "Incident #47",
		"playbookChannelId": "pbchannel456",
		"playbookPostId": "post789"
	}`

	var record ApprovalRecord
	err := json.Unmarshal([]byte(v2JSON), &record)

	require.NoError(t, err, "v2.0 record should deserialize without error")
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", record.ID)
	assert.Equal(t, "A-X7K9Q2", record.Code)
	assert.Equal(t, "pending", record.Status)

	// Playbook fields should be populated
	assert.Equal(t, "playbook123", record.PlaybookRunID)
	assert.Equal(t, "Incident #47", record.PlaybookName)
	assert.Equal(t, "pbchannel456", record.PlaybookChannelID)
	assert.Equal(t, "post789", record.PlaybookPostID)
}

// TestPlaybookFields_OmitEmpty tests that empty playbook fields are omitted in JSON serialization
func TestPlaybookFields_OmitEmpty(t *testing.T) {
	record := &ApprovalRecord{
		ID:               "abcdefghijklmnopqrstuvwxyz",
		Code:             "A-X7K9Q2",
		RequesterID:      "req123",
		ApproverID:       "app456",
		Description:      "Test",
		Status:           StatusPending,
		CreatedAt:        1704931200000,
		RequestChannelID: "channel789",
		SchemaVersion:    1,
		// Playbook fields intentionally left empty
	}

	data, err := json.Marshal(record)
	require.NoError(t, err)

	// Verify playbook fields are NOT in JSON (omitempty working)
	jsonStr := string(data)
	assert.NotContains(t, jsonStr, "playbookRunId", "Empty PlaybookRunID should be omitted")
	assert.NotContains(t, jsonStr, "playbookName", "Empty PlaybookName should be omitted")
	assert.NotContains(t, jsonStr, "playbookChannelId", "Empty PlaybookChannelID should be omitted")
	assert.NotContains(t, jsonStr, "playbookPostId", "Empty PlaybookPostID should be omitted")
}

// TestPlaybookFields_SerializeWhenPresent tests that playbook fields are included when populated
func TestPlaybookFields_SerializeWhenPresent(t *testing.T) {
	record := &ApprovalRecord{
		ID:                "abcdefghijklmnopqrstuvwxyz",
		Code:              "A-X7K9Q2",
		RequesterID:       "req123",
		ApproverID:        "app456",
		Description:       "Test",
		Status:            StatusPending,
		CreatedAt:         1704931200000,
		RequestChannelID:  "channel789",
		SchemaVersion:     1,
		PlaybookRunID:     "playbook123",
		PlaybookName:      "Incident #47",
		PlaybookChannelID: "pbchannel456",
		PlaybookPostID:    "post789",
	}

	data, err := json.Marshal(record)
	require.NoError(t, err)

	// Unmarshal and verify fields persist
	var decoded ApprovalRecord
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "playbook123", decoded.PlaybookRunID)
	assert.Equal(t, "Incident #47", decoded.PlaybookName)
	assert.Equal(t, "pbchannel456", decoded.PlaybookChannelID)
	assert.Equal(t, "post789", decoded.PlaybookPostID)
}
