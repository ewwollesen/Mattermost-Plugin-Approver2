package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestKVStore_SaveApproval(t *testing.T) {
	t.Run("successfully saves approval", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVGet for code uniqueness check (returns nil = code doesn't exist)
		// and for record existence check (returns nil = new record)
		api.On("KVGet", mock.Anything).Return(nil, nil)
		api.On("KVSet", mock.Anything, mock.Anything).Return(nil)

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)

		err = store.SaveApproval(record)
		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("returns error for nil record", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		err := store.SaveApproval(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil approval record")
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record := &approval.ApprovalRecord{
			ID: "",
		}

		err := store.SaveApproval(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")
	})

	t.Run("returns error when KV store fails", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVGet for code uniqueness check first
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return key[:14] == "approval_code:"
		})).Return(nil, nil)

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)

		// Mock KVGet for record existence check (new record)
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return key[:16] == "approval:record:"
		})).Return(nil, nil)

		// Mock KVSet to fail
		appErr := model.NewAppError("test", "test.error", nil, "", 500)
		api.On("KVSet", mock.Anything, mock.Anything).Return(appErr)

		err = store.SaveApproval(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save")
	})
}

func TestKVStore_GetApproval(t *testing.T) {
	t.Run("successfully retrieves approval", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVGet for code uniqueness check (returns nil = code doesn't exist)
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 14 && key[:14] == "approval_code:"
		})).Return(nil, nil)

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)

		// Mock KVGet to return serialized record for retrieval
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 16 && key[:16] == "approval:record:"
		})).Return(func(key string) []byte {
			data, _ := json.Marshal(record)
			return data
		}, nil)

		retrieved, err := store.GetApproval(record.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, record.ID, retrieved.ID)
		assert.Equal(t, record.Code, retrieved.Code)
		api.AssertExpectations(t)
	})

	t.Run("returns ErrRecordNotFound when record does not exist", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		api.On("KVGet", mock.Anything).Return(nil, nil)

		record, err := store.GetApproval("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.True(t, errors.Is(err, approval.ErrRecordNotFound))
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record, err := store.GetApproval("")
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "ID is required")
	})

	t.Run("returns error when KV store fails", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		appErr := model.NewAppError("test", "test.error", nil, "", 500)
		api.On("KVGet", mock.Anything).Return(nil, appErr)

		record, err := store.GetApproval("test123")
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "failed to get")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Return invalid JSON
		api.On("KVGet", mock.Anything).Return([]byte("invalid json"), nil)

		record, err := store.GetApproval("test123")
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})
}

func TestKVStore_DeleteApproval(t *testing.T) {
	t.Run("successfully deletes approval", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		api.On("KVDelete", mock.Anything).Return(nil)

		err := store.DeleteApproval("test123")
		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		err := store.DeleteApproval("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")
	})

	t.Run("returns error when KV store fails", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		appErr := model.NewAppError("test", "test.error", nil, "", 500)
		api.On("KVDelete", mock.Anything).Return(appErr)

		err := store.DeleteApproval("test123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete")
	})
}

func TestKVStore_SaveApproval_Immutability(t *testing.T) {
	t.Run("prevents overwriting finalized record", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVGet for code uniqueness check first
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 14 && key[:14] == "approval_code:"
		})).Return(nil, nil)

		// Create an approved record
		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)
		record.Status = approval.StatusApproved

		// Mock KVGet to return existing approved record for immutability check
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 16 && key[:16] == "approval:record:"
		})).Return(func(key string) []byte {
			data, _ := json.Marshal(record)
			return data
		}, nil)

		// Attempt to overwrite - should fail
		err = store.SaveApproval(record)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, approval.ErrRecordImmutable))
		assert.Contains(t, err.Error(), "cannot modify approval record")
	})

	t.Run("allows overwriting pending record", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVGet for code uniqueness check first
		codeKeyMatched := false
		api.On("KVGet", mock.Anything).Return(func(key string) []byte {
			if len(key) > 14 && key[:14] == "approval_code:" && !codeKeyMatched {
				codeKeyMatched = true
				return nil
			}
			// Return pending record for immutability check
			data, _ := json.Marshal(&approval.ApprovalRecord{Status: approval.StatusPending})
			return data
		}, nil)

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)
		record.Status = approval.StatusPending

		// Mock KVSet for save
		api.On("KVSet", mock.Anything, mock.Anything).Return(nil)

		// Should succeed for pending records
		err = store.SaveApproval(record)
		assert.NoError(t, err)
	})
}

func TestKVStore_GetByCode(t *testing.T) {
	t.Run("retrieves record by code successfully", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock ALL KVGet calls for code generation
		api.On("KVGet", mock.Anything).Return(nil, nil)

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)

		// Mock KVGet for code lookup during GetByCode (using new key format)
		codeKey := fmt.Sprintf("approval:code:%s", record.Code)
		recordIDJSON, _ := json.Marshal(record.ID)
		api.On("KVGet", codeKey).Return(recordIDJSON, nil)

		// Mock KVGet for record key to return full record
		recordKey := fmt.Sprintf("approval:record:%s", record.ID)
		recordJSON, _ := json.Marshal(record)
		api.On("KVGet", recordKey).Return(recordJSON, nil)

		retrieved, err := store.GetByCode(record.Code)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, record.ID, retrieved.ID)
		assert.Equal(t, record.Code, retrieved.Code)
		api.AssertExpectations(t)
	})

	t.Run("returns error for non-existent code", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVGet to return nil (code not found) - using new key format
		api.On("KVGet", "approval:code:A-NOTFND").Return(nil, nil)

		record, err := store.GetByCode("A-NOTFND")
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.ErrorIs(t, err, approval.ErrRecordNotFound)
		api.AssertExpectations(t)
	})

	t.Run("returns error for empty code", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record, err := store.GetByCode("")
		assert.Error(t, err)
		assert.Nil(t, record)
	})
}

func TestKVStore_SaveApproval_CodeLookupIndex(t *testing.T) {
	t.Run("writes all required keys: record, code, requester index, approver index", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock ALL KVGet calls (for code generation uniqueness checks and immutability checks)
		api.On("KVGet", mock.Anything).Return(nil, nil)

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)

		// Expect KVSet for all 4 keys:
		// 1. Primary record
		recordKey := fmt.Sprintf("approval:record:%s", record.ID)
		api.On("KVSet", recordKey, mock.Anything).Return(nil)

		// 2. Code lookup index
		codeKey := fmt.Sprintf("approval:code:%s", record.Code)
		api.On("KVSet", codeKey, mock.Anything).Return(nil)

		// 3. Requester index (timestamped)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 27 && key[:27] == "approval:index:requester:re"
		}), mock.Anything).Return(nil)

		// 4. Approver index (timestamped)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 26 && key[:26] == "approval:index:approver:ap"
		}), mock.Anything).Return(nil)

		err = store.SaveApproval(record)
		assert.NoError(t, err)
		api.AssertExpectations(t)
	})
}

func TestMakeRecordKey(t *testing.T) {
	key := makeRecordKey("test123")
	assert.Equal(t, "approval:record:test123", key)
}

func TestKVStore_GetUserApprovals(t *testing.T) {
	t.Run("retrieves records where user is requester", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Create test records
		record1 := &approval.ApprovalRecord{
			ID:                "record1",
			Code:              "A-ABC123",
			RequesterID:       "user1",
			RequesterUsername: "alice",
			ApproverID:        "user2",
			ApproverUsername:  "bob",
			Status:            approval.StatusPending,
			CreatedAt:         1000,
		}
		record1JSON, _ := json.Marshal(record1)
		record1IDJSON, _ := json.Marshal("record1")

		// Mock KVList to return index keys (2 calls: requester and approver)
		indexKey := "approval:index:requester:user1:9999999998999:record1"
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return([]string{indexKey}, nil).Once()
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return([]string{}, nil).Once()

		// Mock KVGet for index key to return record ID
		api.On("KVGet", indexKey).Return(record1IDJSON, nil)

		// Mock KVGet to return full record
		api.On("KVGet", "approval:record:record1").Return(record1JSON, nil)

		records, err := store.GetUserApprovals("user1")
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, "A-ABC123", records[0].Code)
		api.AssertExpectations(t)
	})

	t.Run("retrieves records where user is approver", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record1 := &approval.ApprovalRecord{
			ID:                "record1",
			Code:              "A-XYZ789",
			RequesterID:       "user1",
			RequesterUsername: "alice",
			ApproverID:        "user2",
			ApproverUsername:  "bob",
			Status:            approval.StatusApproved,
			CreatedAt:         2000,
		}
		record1JSON, _ := json.Marshal(record1)
		record1IDJSON, _ := json.Marshal("record1")

		// Mock KVList - requester index returns empty, approver index returns record
		indexKey := "approval:index:approver:user2:9999999997999:record1"
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return([]string{}, nil).Once()       // requester index (empty)
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return([]string{indexKey}, nil).Once() // approver index

		// Mock KVGet for index key
		api.On("KVGet", indexKey).Return(record1IDJSON, nil)

		// Mock KVGet for full record
		api.On("KVGet", "approval:record:record1").Return(record1JSON, nil)

		records, err := store.GetUserApprovals("user2")
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, "A-XYZ789", records[0].Code)
		api.AssertExpectations(t)
	})

	t.Run("retrieves records where user is both requester and approver", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record1 := &approval.ApprovalRecord{
			ID:          "record1",
			Code:        "A-ABC123",
			RequesterID: "user1",
			ApproverID:  "user2",
			Status:      approval.StatusPending,
			CreatedAt:   1000,
		}
		record2 := &approval.ApprovalRecord{
			ID:          "record2",
			Code:        "A-XYZ789",
			RequesterID: "user2",
			ApproverID:  "user1",
			Status:      approval.StatusApproved,
			CreatedAt:   2000,
		}
		record1JSON, _ := json.Marshal(record1)
		record2JSON, _ := json.Marshal(record2)

		keys := []string{"approval:record:record1", "approval:record:record2"}
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return(keys, nil)
		api.On("KVGet", "approval:record:record1").Return(record1JSON, nil)
		api.On("KVGet", "approval:record:record2").Return(record2JSON, nil)

		records, err := store.GetUserApprovals("user1")
		require.NoError(t, err)
		assert.Len(t, records, 2)
		api.AssertExpectations(t)
	})

	t.Run("sorts records by CreatedAt descending", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record1 := &approval.ApprovalRecord{
			ID:          "record1",
			Code:        "A-OLD",
			RequesterID: "user1",
			ApproverID:  "user2",
			CreatedAt:   1000, // Older
		}
		record2 := &approval.ApprovalRecord{
			ID:          "record2",
			Code:        "A-NEW",
			RequesterID: "user1",
			ApproverID:  "user2",
			CreatedAt:   3000, // Newer
		}
		record3 := &approval.ApprovalRecord{
			ID:          "record3",
			Code:        "A-MID",
			RequesterID: "user1",
			ApproverID:  "user2",
			CreatedAt:   2000, // Middle
		}
		record1JSON, _ := json.Marshal(record1)
		record2JSON, _ := json.Marshal(record2)
		record3JSON, _ := json.Marshal(record3)

		keys := []string{"approval:record:record1", "approval:record:record2", "approval:record:record3"}
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return(keys, nil)
		api.On("KVGet", "approval:record:record1").Return(record1JSON, nil)
		api.On("KVGet", "approval:record:record2").Return(record2JSON, nil)
		api.On("KVGet", "approval:record:record3").Return(record3JSON, nil)

		records, err := store.GetUserApprovals("user1")
		require.NoError(t, err)
		assert.Len(t, records, 3)
		// Should be sorted newest first
		assert.Equal(t, "A-NEW", records[0].Code)
		assert.Equal(t, "A-MID", records[1].Code)
		assert.Equal(t, "A-OLD", records[2].Code)
		api.AssertExpectations(t)
	})

	t.Run("returns empty slice when user has no records", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		// Mock KVList to return empty for both requester and approver indexes
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return([]string{}, nil).Twice()

		// Query for user3 who has no records
		records, err := store.GetUserApprovals("user3")
		require.NoError(t, err)
		assert.NotNil(t, records)
		assert.Len(t, records, 0)
		api.AssertExpectations(t)
	})

	t.Run("filters out records from other users", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record1 := &approval.ApprovalRecord{
			ID:          "record1",
			RequesterID: "user1",
			ApproverID:  "user2",
			Code:        "A-USER1",
		}
		record2 := &approval.ApprovalRecord{
			ID:          "record2",
			RequesterID: "user2",
			ApproverID:  "user3",
			Code:        "A-USER2",
		}
		record1JSON, _ := json.Marshal(record1)
		record2JSON, _ := json.Marshal(record2)

		keys := []string{"approval:record:record1", "approval:record:record2"}
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return(keys, nil)
		api.On("KVGet", "approval:record:record1").Return(record1JSON, nil)
		api.On("KVGet", "approval:record:record2").Return(record2JSON, nil)

		records, err := store.GetUserApprovals("user1")
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, "A-USER1", records[0].Code)
		api.AssertExpectations(t)
	})

	t.Run("returns error when KVList fails", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		appErr := model.NewAppError("test", "test.error", nil, "", 500)
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return(nil, appErr).Once()

		records, err := store.GetUserApprovals("user1")
		assert.Error(t, err)
		assert.Nil(t, records)
		assert.Contains(t, err.Error(), "failed to list")
		api.AssertExpectations(t)
	})

	t.Run("continues with partial results when individual record retrieval fails", func(t *testing.T) {
		api := &plugintest.API{}
		store := NewKVStore(api)

		record1 := &approval.ApprovalRecord{
			ID:          "record1",
			Code:        "A-GOOD",
			RequesterID: "user1",
			ApproverID:  "user2",
		}
		record1JSON, _ := json.Marshal(record1)

		keys := []string{"approval:record:record1", "approval:record:record2"}
		api.On("KVList", 0, MaxApprovalRecordsLimit).Return(keys, nil)
		api.On("KVGet", "approval:record:record1").Return(record1JSON, nil)
		// record2 retrieval fails
		appErr := model.NewAppError("test", "test.error", nil, "", 500)
		api.On("KVGet", "approval:record:record2").Return(nil, appErr)
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		records, err := store.GetUserApprovals("user1")
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, "A-GOOD", records[0].Code)
		api.AssertExpectations(t)
	})
}
