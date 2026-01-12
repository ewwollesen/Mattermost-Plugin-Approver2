package approval

import (
	"fmt"
)

// ValidateApprovalRecord checks if an ApprovalRecord has all required fields
func ValidateApprovalRecord(record *ApprovalRecord) error {
	if record == nil {
		return fmt.Errorf("approval record cannot be nil")
	}

	if record.ID == "" {
		return fmt.Errorf("approval record ID is required")
	}

	if record.Code == "" {
		return fmt.Errorf("approval code is required")
	}

	if record.RequesterID == "" {
		return fmt.Errorf("requester ID is required")
	}

	if record.ApproverID == "" {
		return fmt.Errorf("approver ID is required")
	}

	if record.Description == "" {
		return fmt.Errorf("approval description is required")
	}

	if !IsValidStatus(record.Status) {
		return fmt.Errorf("invalid status: %s, must be pending|approved|denied|canceled", record.Status)
	}

	if record.CreatedAt <= 0 {
		return fmt.Errorf("created timestamp must be positive")
	}

	if record.SchemaVersion <= 0 {
		return fmt.Errorf("schema version must be positive")
	}

	return nil
}

// IsValidStatus checks if a status string is one of the valid values.
// Valid statuses are: pending, approved, denied, canceled.
func IsValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusApproved, StatusDenied, StatusCanceled:
		return true
	default:
		return false
	}
}
