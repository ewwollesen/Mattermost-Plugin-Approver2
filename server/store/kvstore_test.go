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

		// Mock KVGet for code uniqueness check (returns nil = code doesn't exist)
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 14 && key[:14] == "approval_code:"
		})).Return(nil, nil).Once()

		record, err := approval.NewApprovalRecord(
			store,
			"req123", "alice", "Alice",
			"app456", "bob", "Bob",
			"Test approval",
			"channel789",
			"team012",
		)
		require.NoError(t, err)

		// Mock KVGet for code lookup during GetByCode
		codeKey := fmt.Sprintf("approval_code:%s", record.Code)
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

		// Mock KVGet to return nil (code not found)
		api.On("KVGet", "approval_code:A-NOTFND").Return(nil, nil)

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
	t.Run("writes both record and code lookup keys", func(t *testing.T) {
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

		// Mock KVGet to return not found (new record) for immutability check
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 16 && key[:16] == "approval:record:"
		})).Return(nil, nil)

		// Expect KVSet for both record key and code lookup key
		recordKey := fmt.Sprintf("approval:record:%s", record.ID)
		codeKey := fmt.Sprintf("approval_code:%s", record.Code)

		api.On("KVSet", recordKey, mock.Anything).Return(nil)
		api.On("KVSet", codeKey, mock.Anything).Return(nil)

		err = store.SaveApproval(record)
		assert.NoError(t, err)
		api.AssertExpectations(t)
	})
}

func TestMakeRecordKey(t *testing.T) {
	key := makeRecordKey("test123")
	assert.Equal(t, "approval:record:test123", key)
}
