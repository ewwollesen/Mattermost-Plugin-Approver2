package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

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

// isValidVerificationUpdate checks if an update to a decided record is a valid verification operation.
// Story 6.2: Allows adding verification fields to approved records while keeping core fields immutable.
// Returns true if:
// - Status is approved (not denied/canceled)
// - Only verification fields changed (Verified, VerifiedAt, VerificationComment)
// - Core decision fields remain unchanged (Status, DecidedAt, DecisionComment, etc.)
func isValidVerificationUpdate(existing, updated *approval.ApprovalRecord) bool {
	// Only allow verification updates on approved records
	if existing.Status != approval.StatusApproved {
		return false
	}

	// Verify core immutable fields haven't changed
	if existing.ID != updated.ID ||
		existing.Code != updated.Code ||
		existing.Status != updated.Status ||
		existing.RequesterID != updated.RequesterID ||
		existing.RequesterUsername != updated.RequesterUsername ||
		existing.RequesterDisplayName != updated.RequesterDisplayName ||
		existing.ApproverID != updated.ApproverID ||
		existing.ApproverUsername != updated.ApproverUsername ||
		existing.ApproverDisplayName != updated.ApproverDisplayName ||
		existing.Description != updated.Description ||
		existing.DecisionComment != updated.DecisionComment ||
		existing.CreatedAt != updated.CreatedAt ||
		existing.DecidedAt != updated.DecidedAt ||
		existing.CanceledReason != updated.CanceledReason ||
		existing.CanceledAt != updated.CanceledAt ||
		existing.RequestChannelID != updated.RequestChannelID ||
		existing.TeamID != updated.TeamID ||
		existing.NotificationSent != updated.NotificationSent ||
		existing.NotificationPostID != updated.NotificationPostID ||
		existing.OutcomeNotified != updated.OutcomeNotified ||
		existing.SchemaVersion != updated.SchemaVersion {
		return false
	}

	// Allow changes only to verification fields
	// This is valid if: !existing.Verified && updated.Verified (first-time verification)
	if existing.Verified {
		// Already verified - no further changes allowed
		return false
	}

	// New verification: Verified must be set to true, VerifiedAt must be > 0
	if !updated.Verified || updated.VerifiedAt == 0 {
		return false
	}

	// Valid verification update
	return true
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
		// Record exists - check if modifications violate immutability
		if existing.Status != approval.StatusPending {
			// Decided records are generally immutable, but allow verification updates (Story 6.2)
			if !isValidVerificationUpdate(existing, record) {
				return fmt.Errorf("cannot modify approval record %s: %w", record.ID, approval.ErrRecordImmutable)
			}
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

	// Create code lookup index: approval:code:{code} → recordID
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

	// Create requester index: approval:index:requester:{userID}:{invertedTimestamp}:{recordID} → recordID
	// This enables efficient queries for "approvals I requested"
	if record.RequesterID != "" && record.CreatedAt > 0 {
		requesterKey := makeRequesterIndexKey(record.RequesterID, record.CreatedAt, record.ID)
		recordIDJSON, err := json.Marshal(record.ID)
		if err != nil {
			return fmt.Errorf("failed to marshal record ID for requester index: %w", err)
		}

		appErr = s.api.KVSet(requesterKey, recordIDJSON)
		if appErr != nil {
			return fmt.Errorf("failed to save requester index for %s: %w", record.ID, appErr)
		}
	}

	// Create approver index: approval:index:approver:{userID}:{invertedTimestamp}:{recordID} → recordID
	// This enables efficient queries for "approvals I need to decide"
	if record.ApproverID != "" && record.CreatedAt > 0 {
		approverKey := makeApproverIndexKey(record.ApproverID, record.CreatedAt, record.ID)
		recordIDJSON, err := json.Marshal(record.ID)
		if err != nil {
			return fmt.Errorf("failed to marshal record ID for approver index: %w", err)
		}

		appErr = s.api.KVSet(approverKey, recordIDJSON)
		if appErr != nil {
			return fmt.Errorf("failed to save approver index for %s: %w", record.ID, appErr)
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

// GetApprovalByCode retrieves an ApprovalRecord by either human-friendly code or full 26-char ID
// This method supports both code formats (e.g., "A-X7K9Q2") and full Mattermost IDs (26 chars)
func (s *KVStore) GetApprovalByCode(codeOrID string) (*approval.ApprovalRecord, error) {
	if codeOrID == "" {
		return nil, fmt.Errorf("approval code or ID is required")
	}

	// Detect if input is a full 26-char ID (no dashes, exactly 26 characters)
	// Mattermost IDs are 26 characters of alphanumeric characters
	if len(codeOrID) == 26 && !strings.Contains(codeOrID, "-") {
		// Direct lookup by full ID
		return s.GetApproval(codeOrID)
	}

	// Otherwise, treat as human-friendly code and do code lookup
	// Look up record ID from code index
	codeKey := makeCodeKey(codeOrID)
	data, appErr := s.api.KVGet(codeKey)
	if appErr != nil {
		return nil, fmt.Errorf("failed to lookup code %s: %w", codeOrID, appErr)
	}

	if data == nil {
		return nil, fmt.Errorf("approval code %s: %w", codeOrID, approval.ErrRecordNotFound)
	}

	// Unmarshal record ID
	var recordID string
	if err := json.Unmarshal(data, &recordID); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record ID for code %s: %w", codeOrID, err)
	}

	// Retrieve full record by ID
	return s.GetApproval(recordID)
}

const (
	// MaxApprovalRecordsLimit is the maximum number of approval records that can be retrieved at once
	// This limit prevents memory exhaustion and ensures reasonable query performance
	MaxApprovalRecordsLimit = 10000
)

// GetAllApprovals retrieves all approval records from the KV store
// This method is used for admin operations like the status command to calculate statistics
//
// Performance: This method loads up to MaxApprovalRecordsLimit (10,000) records into memory.
// For deployments with more than 10,000 approval records, only the first 10,000 will be returned
// and a warning will be logged. Consider implementing pagination for very large deployments.
func (s *KVStore) GetAllApprovals() ([]*approval.ApprovalRecord, error) {
	// List all keys with the approval record prefix
	keys, appErr := s.api.KVList(0, MaxApprovalRecordsLimit)
	if appErr != nil {
		return nil, fmt.Errorf("failed to list approval records: %w", appErr)
	}

	// Warn if we hit the limit (may indicate truncated results)
	if len(keys) >= MaxApprovalRecordsLimit {
		s.api.LogWarn("GetAllApprovals hit maximum record limit - results may be incomplete",
			"limit", MaxApprovalRecordsLimit,
			"keys_returned", len(keys),
			"recommendation", "Consider implementing pagination for large deployments",
		)
	}

	records := make([]*approval.ApprovalRecord, 0)
	recordPrefix := "approval:record:"

	// Iterate through keys and retrieve records
	for _, key := range keys {
		// Filter for approval record keys only (exclude code indexes and other keys)
		if len(key) < len(recordPrefix) || key[:len(recordPrefix)] != recordPrefix {
			continue
		}

		// Extract record ID from key
		recordID := key[len(recordPrefix):]

		// Retrieve the record
		record, err := s.GetApproval(recordID)
		if err != nil {
			// Log error but continue with other records
			// This ensures partial failures don't prevent status reporting
			s.api.LogWarn("Failed to retrieve approval record during GetAllApprovals",
				"record_id", recordID,
				"error", err.Error(),
			)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

// GetUserApprovals retrieves all approval records where the specified user is either requester or approver
// This method is used for the /approve list command to show a user their approval history
//
// Performance: Uses efficient prefix-based index queries instead of scanning all records.
// Queries two indexes:
//   - approval:index:requester:{userID}:* for records where user is requester
//   - approval:index:approver:{userID}:* for records where user is approver
//
// Records are naturally sorted by timestamp descending (most recent first) due to inverted
// timestamp keys, eliminating the need for in-memory sorting.
func (s *KVStore) GetUserApprovals(userID string) ([]*approval.ApprovalRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Track unique record IDs to avoid duplicates (user could be both requester and approver)
	seenRecords := make(map[string]bool)
	records := make([]*approval.ApprovalRecord, 0)

	// Query 1: Get records where user is the requester
	requesterPrefix := fmt.Sprintf("approval:index:requester:%s:", userID)
	requesterKeys, appErr := s.api.KVList(0, MaxApprovalRecordsLimit)
	if appErr != nil {
		return nil, fmt.Errorf("failed to list requester index keys: %w", appErr)
	}

	for _, key := range requesterKeys {
		// Filter for this user's requester index keys only
		if len(key) < len(requesterPrefix) || key[:len(requesterPrefix)] != requesterPrefix {
			continue
		}

		// Get the record ID from the index
		recordIDData, getErr := s.api.KVGet(key)
		if getErr != nil {
			s.api.LogWarn("Failed to get record ID from requester index",
				"key", key,
				"user_id", userID,
				"error", getErr.Error(),
			)
			continue
		}

		if recordIDData == nil {
			continue
		}

		var recordID string
		if err := json.Unmarshal(recordIDData, &recordID); err != nil {
			s.api.LogWarn("Failed to unmarshal record ID from requester index",
				"key", key,
				"user_id", userID,
				"error", err.Error(),
			)
			continue
		}

		// Skip if already seen (handles duplicate case)
		if seenRecords[recordID] {
			continue
		}

		// Retrieve the full record
		record, err := s.GetApproval(recordID)
		if err != nil {
			s.api.LogWarn("Failed to retrieve approval record from requester index",
				"record_id", recordID,
				"user_id", userID,
				"error", err.Error(),
			)
			continue
		}

		records = append(records, record)
		seenRecords[recordID] = true
	}

	// Query 2: Get records where user is the approver
	approverPrefix := fmt.Sprintf("approval:index:approver:%s:", userID)
	approverKeys, appErr := s.api.KVList(0, MaxApprovalRecordsLimit)
	if appErr != nil {
		return nil, fmt.Errorf("failed to list approver index keys: %w", appErr)
	}

	for _, key := range approverKeys {
		// Filter for this user's approver index keys only
		if len(key) < len(approverPrefix) || key[:len(approverPrefix)] != approverPrefix {
			continue
		}

		// Get the record ID from the index
		recordIDData, appErr := s.api.KVGet(key)
		if appErr != nil {
			s.api.LogWarn("Failed to get record ID from approver index",
				"key", key,
				"user_id", userID,
				"error", appErr.Error(),
			)
			continue
		}

		if recordIDData == nil {
			continue
		}

		var recordID string
		if err := json.Unmarshal(recordIDData, &recordID); err != nil {
			s.api.LogWarn("Failed to unmarshal record ID from approver index",
				"key", key,
				"user_id", userID,
				"error", err.Error(),
			)
			continue
		}

		// Skip if already seen (handles case where user is both requester and approver)
		if seenRecords[recordID] {
			continue
		}

		// Retrieve the full record
		record, err := s.GetApproval(recordID)
		if err != nil {
			s.api.LogWarn("Failed to retrieve approval record from approver index",
				"record_id", recordID,
				"user_id", userID,
				"error", err.Error(),
			)
			continue
		}

		records = append(records, record)
		seenRecords[recordID] = true
	}

	// Records are naturally sorted by timestamp descending due to inverted timestamp keys
	// However, when combining requester and approver indexes, we need to sort the merged result
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})

	return records, nil
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
	return fmt.Sprintf("approval:code:%s", code)
}

// makeRequesterIndexKey generates timestamped index key for requester queries
// Format: approval:index:requester:{userID}:{timestamp}:{recordID}
// Timestamp is inverted (9999999999999 - timestamp) to achieve descending order
func makeRequesterIndexKey(userID string, timestamp int64, recordID string) string {
	// Invert timestamp for descending order (most recent first)
	invertedTimestamp := 9999999999999 - timestamp
	return fmt.Sprintf("approval:index:requester:%s:%013d:%s", userID, invertedTimestamp, recordID)
}

// makeApproverIndexKey generates timestamped index key for approver queries
// Format: approval:index:approver:{userID}:{timestamp}:{recordID}
// Timestamp is inverted (9999999999999 - timestamp) to achieve descending order
func makeApproverIndexKey(userID string, timestamp int64, recordID string) string {
	// Invert timestamp for descending order (most recent first)
	invertedTimestamp := 9999999999999 - timestamp
	return fmt.Sprintf("approval:index:approver:%s:%013d:%s", userID, invertedTimestamp, recordID)
}

// GetPendingRequestsOlderThan retrieves all pending approval requests older than the specified duration
// This method is used by the timeout checker to find abandoned requests for auto-cancellation.
//
// Performance: Scans all approver index keys and filters by:
// 1. Status == "pending" (only pending requests can time out)
// 2. (CurrentTime - CreatedAt) > cutoffDuration
//
// Returns records in arbitrary order (timeout processing doesn't require ordering).
//
// TODO: Add pagination support for systems with thousands of pending requests.
// Current implementation scans all index keys without pagination (using MaxApprovalRecordsLimit),
// which is acceptable for MVP but may impact performance at scale.
func (s *KVStore) GetPendingRequestsOlderThan(cutoffDuration time.Duration) ([]*approval.ApprovalRecord, error) {
	// Calculate cutoff timestamp (requests created before this are considered old)
	cutoffTime := time.Now().Add(-cutoffDuration).Unix() * 1000 // Convert to epoch millis

	// List all keys in KV store
	keys, appErr := s.api.KVList(0, MaxApprovalRecordsLimit)
	if appErr != nil {
		return nil, fmt.Errorf("failed to list KV store keys: %w", appErr)
	}

	// Track unique pending old records
	seenRecords := make(map[string]bool)
	records := make([]*approval.ApprovalRecord, 0)

	// Scan all approver index keys to find old pending requests
	// Format: approval:index:approver:{userID}:{invertedTimestamp}:{recordID}
	approverIndexPrefix := "approval:index:approver:"

	for _, key := range keys {
		// Filter for approver index keys only
		if len(key) < len(approverIndexPrefix) || key[:len(approverIndexPrefix)] != approverIndexPrefix {
			continue
		}

		// Get the record ID from the index
		recordIDData, getErr := s.api.KVGet(key)
		if getErr != nil {
			s.api.LogWarn("Failed to get record ID from approver index during timeout scan",
				"key", key,
				"error", getErr.Error(),
			)
			continue
		}

		if recordIDData == nil {
			continue
		}

		var recordID string
		if err := json.Unmarshal(recordIDData, &recordID); err != nil {
			s.api.LogWarn("Failed to unmarshal record ID from approver index during timeout scan",
				"key", key,
				"error", err.Error(),
			)
			continue
		}

		// Skip if already seen (avoid duplicates)
		if seenRecords[recordID] {
			continue
		}
		seenRecords[recordID] = true

		// Retrieve the full record
		record, err := s.GetApproval(recordID)
		if err != nil {
			s.api.LogWarn("Failed to retrieve approval record during timeout scan",
				"record_id", recordID,
				"error", err.Error(),
			)
			continue
		}

		// Filter 1: Only pending requests can time out
		if record.Status != approval.StatusPending {
			continue
		}

		// Filter 2: Check if request is older than cutoff
		if record.CreatedAt > cutoffTime {
			continue // Request is too new
		}

		// This is a pending request older than cutoff - add to results
		records = append(records, record)
	}

	s.api.LogDebug("Completed timeout scan",
		"cutoff_duration", cutoffDuration.String(),
		"cutoff_timestamp", cutoffTime,
		"pending_old_requests", len(records),
	)

	return records, nil
}
