package approval

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// approvalCodePattern validates the format: A-X7K9Q2 (letter-dash-6 alphanumeric chars)
var approvalCodePattern = regexp.MustCompile(`^[A-Z]-[A-Z0-9]{6}$`)

// ApprovalStore defines the interface for approval record persistence
type ApprovalStore interface {
	GetApproval(id string) (*ApprovalRecord, error)
	GetByCode(code string) (*ApprovalRecord, error)
	SaveApproval(record *ApprovalRecord) error
}

// Service provides business logic for approval operations
type Service struct {
	store     ApprovalStore
	api       plugin.API
	botUserID string
}

// NewService creates a new approval service
func NewService(store ApprovalStore, api plugin.API, botUserID string) *Service {
	return &Service{
		store:     store,
		api:       api,
		botUserID: botUserID,
	}
}

// CancelApproval cancels a pending approval request
// Returns:
// - ErrRecordNotFound if approval doesn't exist
// - ErrRecordImmutable if approval is not pending
// - error with "permission denied" if requester doesn't match
func (s *Service) CancelApproval(approvalCode, requesterID string) error {
	// Validation: code and requester ID required (trim whitespace)
	approvalCode = strings.TrimSpace(approvalCode)
	if approvalCode == "" {
		return fmt.Errorf("approval code is required")
	}

	// Validate approval code format (A-X7K9Q2)
	if !approvalCodePattern.MatchString(approvalCode) {
		return fmt.Errorf("invalid approval code format: expected format like 'A-X7K9Q2'")
	}

	requesterID = strings.TrimSpace(requesterID)
	if requesterID == "" {
		return fmt.Errorf("requester ID is required")
	}

	// Retrieve approval record by code
	record, err := s.store.GetByCode(approvalCode)
	if err != nil {
		// Wrap ErrRecordNotFound with context
		return fmt.Errorf("failed to retrieve approval %s: %w", approvalCode, err)
	}

	// Access control: verify requester is the owner
	if record.RequesterID != requesterID {
		return fmt.Errorf("permission denied: only requester can cancel approval")
	}

	// Immutability check: only pending approvals can be canceled
	if record.Status != StatusPending {
		return fmt.Errorf("cannot cancel approval with status %s: %w", record.Status, ErrRecordImmutable)
	}

	// Update record status and timestamp
	record.Status = StatusCanceled
	record.DecidedAt = model.GetMillis()

	// Persist updated record
	if err := s.store.SaveApproval(record); err != nil {
		return fmt.Errorf("failed to save canceled approval %s: %w", approvalCode, err)
	}

	return nil
}

// RecordDecision records an approval decision (approve or deny) with immutability guarantees.
// This method enforces:
// - Authorization: Only the designated approver can record a decision
// - Immutability: Decisions can only be recorded on pending approvals
// - Atomicity: All field updates happen atomically via KV store
// - Concurrency Safety: Uses optimistic locking via KVStore to prevent race conditions
//
// Performance: Completes within 2 seconds (NFR-P2). Timing is measured and logged.
//
// Returns:
// - On success: (updated ApprovalRecord, nil) - caller should use record to send outcome notification
// - On failure: (nil, error) where error is:
//   - ErrRecordNotFound if approval doesn't exist
//   - ErrRecordImmutable if approval is not pending
//   - error with "permission denied" if approver doesn't match
//   - error for validation failures (empty IDs, invalid decision value)
func (s *Service) RecordDecision(approvalID, approverID, decision, comment string) (*ApprovalRecord, error) {
	// Performance tracking (NFR-P2: must complete within 2 seconds)
	startTime := model.GetMillis()

	// Validation: trim whitespace and check required fields
	approvalID = strings.TrimSpace(approvalID)
	if approvalID == "" {
		return nil, fmt.Errorf("approval ID is required")
	}

	approverID = strings.TrimSpace(approverID)
	if approverID == "" {
		return nil, fmt.Errorf("approver ID is required")
	}

	// Validate decision value
	if decision != "approved" && decision != "denied" {
		return nil, fmt.Errorf("invalid decision: must be 'approved' or 'denied'")
	}

	// Trim comment (can be empty string)
	comment = strings.TrimSpace(comment)

	// Retrieve approval record by ID
	record, err := s.store.GetApproval(approvalID)
	if err != nil {
		// Wrap error with context (preserves ErrRecordNotFound if present)
		return nil, fmt.Errorf("failed to retrieve approval %s: %w", approvalID, err)
	}

	// Defensive nil check (should not happen with current KVStore implementation)
	if record == nil {
		return nil, fmt.Errorf("approval record %s is nil after retrieval", approvalID)
	}

	// Authorization check: verify authenticated user is the designated approver
	if record.ApproverID != approverID {
		s.api.LogError("Unauthorized decision attempt",
			"approval_id", approvalID,
			"authenticated_user", approverID,
			"designated_approver", record.ApproverID,
		)
		return nil, fmt.Errorf("permission denied: only the designated approver can make this decision")
	}

	// Immutability check: only pending approvals can be decided
	if record.Status != StatusPending {
		s.api.LogError("Attempted to modify finalized approval",
			"approval_id", approvalID,
			"current_status", record.Status,
			"attempted_action", decision,
		)
		return nil, fmt.Errorf("cannot modify approval with status %s: %w", record.Status, ErrRecordImmutable)
	}

	// Map decision string to status constant
	var newStatus string
	if decision == "approved" {
		newStatus = StatusApproved
	} else {
		newStatus = StatusDenied
	}

	// Update record fields atomically (in-memory, persisted atomically in SaveApproval)
	record.Status = newStatus
	record.DecisionComment = comment
	record.DecidedAt = model.GetMillis()

	// Persist updated record with defense-in-depth immutability check
	// KVStore re-checks status != pending before write (kvstore.go:33-40),
	// providing protection against race conditions via optimistic locking
	if err := s.store.SaveApproval(record); err != nil {
		return nil, fmt.Errorf("failed to save decision for approval %s: %w", approvalID, err)
	}

	// Calculate operation duration for performance monitoring (NFR-P2)
	duration := model.GetMillis() - startTime

	// Log successful decision recording
	s.api.LogInfo("Approval decision recorded",
		"approval_id", approvalID,
		"code", record.Code,
		"decision", decision,
		"approver_id", approverID,
	)

	// Separately log performance metrics for monitoring
	s.api.LogDebug("RecordDecision performance",
		"approval_id", approvalID,
		"duration_ms", duration,
	)

	// Warn if operation exceeds performance budget (2000ms = NFR-P2)
	if duration > 2000 {
		s.api.LogWarn("RecordDecision exceeded performance budget",
			"approval_id", approvalID,
			"duration_ms", duration,
			"budget_ms", 2000,
		)
	}

	// Return updated record for caller to send outcome notification
	return record, nil
}
