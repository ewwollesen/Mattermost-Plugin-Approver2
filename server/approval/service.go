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
	GetByCode(code string) (*ApprovalRecord, error)
	SaveApproval(record *ApprovalRecord) error
}

// Service provides business logic for approval operations
type Service struct {
	store ApprovalStore
	api   plugin.API
}

// NewService creates a new approval service
func NewService(store ApprovalStore, api plugin.API) *Service {
	return &Service{
		store: store,
		api:   api,
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

// RecordDecision records an approval decision (approve or deny)
// Story 2.4 integration point - stub implementation for Story 2.3
func (s *Service) RecordDecision(approvalID, approverID, decision, comment string) error {
	// Validation
	if approvalID == "" {
		return fmt.Errorf("approval ID is required")
	}
	if approverID == "" {
		return fmt.Errorf("approver ID is required")
	}
	if decision != "approved" && decision != "denied" {
		return fmt.Errorf("invalid decision: must be 'approved' or 'denied'")
	}

	// TODO Story 2.4: Implement full decision recording logic
	// For now, return success to allow Story 2.3 to complete
	// Full implementation will:
	// - Retrieve approval record
	// - Verify approver matches
	// - Verify status is pending
	// - Update status, decision comment, decided timestamp
	// - Save record immutably

	return nil
}
