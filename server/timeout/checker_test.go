package timeout

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/store"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestNewChecker verifies TimeoutChecker is created with proper dependencies
func TestNewChecker(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockStore := store.NewKVStore(mockAPI)
	mockService := approval.NewService(mockStore, mockAPI, "bot123")
	botUserID := "bot123"

	checker := NewChecker(mockStore, mockService, mockAPI, botUserID)

	assert.NotNil(t, checker)
	assert.Equal(t, mockStore, checker.store)
	assert.Equal(t, mockService, checker.service)
	assert.Equal(t, mockAPI, checker.api)
	assert.Equal(t, botUserID, checker.botUserID)
	assert.NotNil(t, checker.done)
}

// TestStartStop verifies lifecycle management
func TestStartStop(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockStore := store.NewKVStore(mockAPI)
	mockService := approval.NewService(mockStore, mockAPI, "bot123")

	// Mock LogInfo for lifecycle messages
	mockAPI.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	checker := NewChecker(mockStore, mockService, mockAPI, "bot123")

	// Start checker
	checker.Start()
	assert.NotNil(t, checker.ctx)
	assert.NotNil(t, checker.cancel)

	// Stop checker (should complete without hanging)
	done := make(chan struct{})
	go func() {
		checker.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success - Stop completed
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not complete within 2 seconds")
	}

	mockAPI.AssertExpectations(t)
}

// TestCheckTimeoutsNoIndexKeys verifies behavior when no index keys are returned from KVList
func TestCheckTimeoutsNoIndexKeys(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockStore := store.NewKVStore(mockAPI)
	mockService := approval.NewService(mockStore, mockAPI, "bot123")

	// Mock KVList to return empty results
	mockAPI.On("KVList", 0, store.MaxApprovalRecordsLimit).Return([]string{}, nil)
	// Mock LogDebug with all expected arguments (variadic keyvals)
	mockAPI.On("LogDebug", "Completed timeout scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	checker := NewChecker(mockStore, mockService, mockAPI, "bot123")

	err := checker.checkTimeouts()

	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

// TestCheckTimeoutsSkipsNonPendingRequests verifies only pending requests are processed
func TestCheckTimeoutsSkipsNonPendingRequests(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockStore := store.NewKVStore(mockAPI)
	mockService := approval.NewService(mockStore, mockAPI, "bot123")

	// Create an approved record (should be skipped)
	approvedRecord := &approval.ApprovalRecord{
		ID:          "record123",
		Code:        "A-X7K9Q2",
		Status:      approval.StatusApproved,
		RequesterID: "user123",
		CreatedAt:   time.Now().Add(-1*time.Hour).Unix() * 1000, // Old enough
	}

	// Mock KVList to return approver index key
	indexKey := "approval:index:approver:approver123:0000000000001:record123"
	mockAPI.On("KVList", 0, store.MaxApprovalRecordsLimit).Return([]string{indexKey}, nil)

	// Mock KVGet for index
	recordIDJSON := []byte(`"record123"`)
	mockAPI.On("KVGet", indexKey).Return(recordIDJSON, nil)

	// Mock KVGet for record
	recordJSON := mustMarshalJSON(t, approvedRecord)
	mockAPI.On("KVGet", "approval:record:record123").Return(recordJSON, nil)

	// Mock LogDebug with all expected arguments (variadic keyvals)
	mockAPI.On("LogDebug", "Completed timeout scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	checker := NewChecker(mockStore, mockService, mockAPI, "bot123")

	err := checker.checkTimeouts()

	assert.NoError(t, err)
	// Verify CancelApprovalByID was NOT called (record skipped)
	mockAPI.AssertExpectations(t)
}

// TestCheckTimeoutsSkipsNewRequests verifies recent requests are not timed out
func TestCheckTimeoutsSkipsNewRequests(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockStore := store.NewKVStore(mockAPI)
	mockService := approval.NewService(mockStore, mockAPI, "bot123")

	// Create a pending record that's too new (created 5 minutes ago)
	newRecord := &approval.ApprovalRecord{
		ID:          "record123",
		Code:        "A-X7K9Q2",
		Status:      approval.StatusPending,
		RequesterID: "user123",
		CreatedAt:   time.Now().Add(-5*time.Minute).Unix() * 1000, // Only 5 minutes old
	}

	// Mock KVList to return approver index key
	indexKey := "approval:index:approver:approver123:0000000000001:record123"
	mockAPI.On("KVList", 0, store.MaxApprovalRecordsLimit).Return([]string{indexKey}, nil)

	// Mock KVGet for index
	recordIDJSON := []byte(`"record123"`)
	mockAPI.On("KVGet", indexKey).Return(recordIDJSON, nil)

	// Mock KVGet for record
	recordJSON := mustMarshalJSON(t, newRecord)
	mockAPI.On("KVGet", "approval:record:record123").Return(recordJSON, nil)

	// Mock LogDebug with all expected arguments (variadic keyvals)
	mockAPI.On("LogDebug", "Completed timeout scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	checker := NewChecker(mockStore, mockService, mockAPI, "bot123")

	err := checker.checkTimeouts()

	assert.NoError(t, err)
	// Verify CancelApprovalByID was NOT called (record too new)
	mockAPI.AssertExpectations(t)
}

// TestCheckTimeoutsHappyPath verifies the complete timeout and cancellation flow
func TestCheckTimeoutsHappyPath(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockStore := store.NewKVStore(mockAPI)
	mockService := approval.NewService(mockStore, mockAPI, "bot123")

	// Create a timed-out pending record (created 31 minutes ago)
	oldTime := time.Now().Add(-31*time.Minute).Unix() * 1000
	timedOutRecord := &approval.ApprovalRecord{
		ID:                  "record123",
		Code:                "A-X7K9Q2",
		Status:              approval.StatusPending,
		RequesterID:         "user123",
		RequesterUsername:   "johndoe",
		ApproverID:          "approver123",
		ApproverUsername:    "janedoe",
		ApproverDisplayName: "Jane Doe",
		Description:         "Test approval request",
		CreatedAt:           oldTime,
		NotificationPostID:  "post123",
	}

	// Mock KVList to return approver index key
	indexKey := "approval:index:approver:approver123:0000000000001:record123"
	mockAPI.On("KVList", 0, store.MaxApprovalRecordsLimit).Return([]string{indexKey}, nil)

	// Mock KVGet for index
	recordIDJSON := []byte(`"record123"`)
	mockAPI.On("KVGet", indexKey).Return(recordIDJSON, nil)

	// Mock KVGet for record - return pending record 3 times:
	// 1. GetPendingRequestsOlderThan loads the record
	// 2. CancelApprovalByID → GetApproval loads it
	// 3. SaveApproval → GetApproval (immutability check) loads it
	recordJSON := mustMarshalJSON(t, timedOutRecord)
	mockAPI.On("KVGet", "approval:record:record123").Return(recordJSON, nil).Times(3)

	// Mock SaveApproval KVSet calls (CancelApprovalByID triggers SaveApproval)
	// SaveApproval will KVSet the record, code index, requester index, and approver index
	mockAPI.On("KVSet", "approval:record:record123", mock.Anything).Return(nil).Once()
	mockAPI.On("KVSet", "approval:code:A-X7K9Q2", mock.Anything).Return(nil).Once()
	mockAPI.On("KVSet", mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "approval:index:requester:")
	}), mock.Anything).Return(nil).Once()
	mockAPI.On("KVSet", mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "approval:index:approver:")
	}), mock.Anything).Return(nil).Once()

	// Mock LogInfo for cancellation (service.go logs this)
	mockAPI.On("LogInfo", "Approval canceled", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// Mock LogInfo for timeout completion (checker.go logs this)
	mockAPI.On("LogInfo", "Auto-canceled timed-out request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// Mock KVGet for record reload - now returns the canceled record
	canceledRecord := *timedOutRecord
	canceledRecord.Status = approval.StatusCanceled
	canceledRecord.CanceledReason = "Auto-canceled: No response within 30 minutes"
	canceledRecord.CanceledAt = time.Now().Unix() * 1000
	canceledRecord.DecidedAt = canceledRecord.CanceledAt
	canceledRecordJSON := mustMarshalJSON(t, &canceledRecord)
	mockAPI.On("KVGet", "approval:record:record123").Return(canceledRecordJSON, nil).Once()

	// Mock GetDirectChannel for timeout notification
	mockAPI.On("GetDirectChannel", "bot123", "user123").Return(&model.Channel{Id: "dm_channel_123"}, nil)

	// Mock CreatePost for timeout notification
	mockAPI.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
		return post.ChannelId == "dm_channel_123" && post.UserId == "bot123"
	})).Return(&model.Post{Id: "notification_post_123"}, nil)

	// Mock GetPost for approver post update
	originalPost := &model.Post{
		Id:      "post123",
		Message: "Original approval message",
		Props:   model.StringInterface{},
	}
	mockAPI.On("GetPost", "post123").Return(originalPost, nil)

	// Mock UpdatePost for disabling buttons
	mockAPI.On("UpdatePost", mock.MatchedBy(func(post *model.Post) bool {
		return post.Id == "post123" && len(post.Props) == 0 // Props cleared
	})).Return(&model.Post{Id: "post123"}, nil)

	// Mock LogDebug for processing and completion
	mockAPI.On("LogDebug", "Processing timed-out requests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	mockAPI.On("LogDebug", "Completed timeout scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	checker := NewChecker(mockStore, mockService, mockAPI, "bot123")

	err := checker.checkTimeouts()

	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

// TODO: Add integration tests for error paths and graceful degradation scenarios
// These require complex mocking of the full SaveApproval flow including KVSet calls
// Deferred to integration testing phase with actual Mattermost instance
//
// Scenarios to test:
// - Notification failure doesn't block cancellation (AC8)
// - Post update failure doesn't block cancellation
// - GetPendingRequestsOlderThan error handling
// - CancelApprovalByID failure with graceful degradation
// - Record reload failure after cancellation

// Helper function to marshal JSON for tests
func mustMarshalJSON(t *testing.T, v any) []byte {
	data, err := json.Marshal(v)
	assert.NoError(t, err)
	return data
}
