package store

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// KVStore provides persistence for approval records using Mattermost KV store
type KVStore struct {
	api plugin.API
}

// NewKVStore creates a new KV store adapter
func NewKVStore(api plugin.API) *KVStore {
	return &KVStore{
		api: api,
	}
}

// SaveApproval persists an ApprovalRecord to the KV store
func (s *KVStore) SaveApproval(record *approval.ApprovalRecord) error {
	if record == nil {
		return fmt.Errorf("cannot save nil approval record")
	}

	if record.ID == "" {
		return fmt.Errorf("approval record ID is required")
	}

	// Enforce immutability: check if record exists and is finalized
	existing, err := s.GetApproval(record.ID)
	if err == nil {
		// Record exists - check if it's immutable
		if existing.Status != approval.StatusPending {
			return fmt.Errorf("cannot modify approval record %s: %w", record.ID, approval.ErrRecordImmutable)
		}
	}
	// If record doesn't exist (ErrRecordNotFound), proceed with save

	// Serialize record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal approval record: %w", err)
	}

	// Store with hierarchical key
	key := makeRecordKey(record.ID)
	appErr := s.api.KVSet(key, data)
	if appErr != nil {
		return fmt.Errorf("failed to save approval record %s: %w", record.ID, appErr)
	}

	// Create code lookup index: approval_code:{code} → recordID
	if record.Code != "" {
		codeKey := makeCodeKey(record.Code)
		recordIDJSON, err := json.Marshal(record.ID)
		if err != nil {
			return fmt.Errorf("failed to marshal record ID for code index: %w", err)
		}

		appErr = s.api.KVSet(codeKey, recordIDJSON)
		if appErr != nil {
			return fmt.Errorf("failed to save code lookup index for %s: %w", record.Code, appErr)
		}
	}

	return nil
}

// GetApproval retrieves an ApprovalRecord by ID
func (s *KVStore) GetApproval(id string) (*approval.ApprovalRecord, error) {
	if id == "" {
		return nil, fmt.Errorf("approval ID is required")
	}

	// Retrieve from KV store
	key := makeRecordKey(id)
	data, appErr := s.api.KVGet(key)
	if appErr != nil {
		return nil, fmt.Errorf("failed to get approval record %s: %w", id, appErr)
	}

	if data == nil {
		return nil, fmt.Errorf("approval record %s: %w", id, approval.ErrRecordNotFound)
	}

	// Deserialize JSON to struct
	var record approval.ApprovalRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal approval record %s: %w", id, err)
	}

	return &record, nil
}

// DeleteApproval removes an ApprovalRecord from the KV store (for testing/admin only)
func (s *KVStore) DeleteApproval(id string) error {
	if id == "" {
		return fmt.Errorf("approval ID is required")
	}

	key := makeRecordKey(id)
	appErr := s.api.KVDelete(key)
	if appErr != nil {
		return fmt.Errorf("failed to delete approval record %s: %w", id, appErr)
	}

	return nil
}

// GetByCode retrieves an ApprovalRecord by human-friendly code
func (s *KVStore) GetByCode(code string) (*approval.ApprovalRecord, error) {
	if code == "" {
		return nil, fmt.Errorf("approval code is required")
	}

	// Look up record ID from code index
	codeKey := makeCodeKey(code)
	data, appErr := s.api.KVGet(codeKey)
	if appErr != nil {
		return nil, fmt.Errorf("failed to lookup code %s: %w", code, appErr)
	}

	if data == nil {
		return nil, fmt.Errorf("approval code %s: %w", code, approval.ErrRecordNotFound)
	}

	// Unmarshal record ID
	var recordID string
	if err := json.Unmarshal(data, &recordID); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record ID for code %s: %w", code, err)
	}

	// Retrieve full record by ID
	return s.GetApproval(recordID)
}

// KVGet implements the Storer interface for code generation uniqueness checks
func (s *KVStore) KVGet(key string) ([]byte, error) {
	data, appErr := s.api.KVGet(key)
	if appErr != nil {
		return nil, appErr
	}
	return data, nil
}

// makeRecordKey generates the KV store key for an approval record
func makeRecordKey(id string) string {
	return fmt.Sprintf("approval:record:%s", id)
}

// makeCodeKey generates the KV store key for code lookup index
func makeCodeKey(code string) string {
	return fmt.Sprintf("approval_code:%s", code)
}
