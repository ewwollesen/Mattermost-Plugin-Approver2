package approval

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
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

// Sentinel errors for validation failures
var (
	ErrDescriptionRequired = errors.New("description is required")
	ErrDescriptionTooLong  = errors.New("description exceeds maximum length")
	ErrApproverRequired    = errors.New("approver is required")
	ErrInvalidApprover     = errors.New("invalid approver")
)

const (
	MaxDescriptionLength = 1000
)

// ValidateDescription validates the description field for approval requests.
// Returns an error if:
// - Description is empty or whitespace-only
// - Description exceeds 1000 characters
//
// Error messages include character count for length violations to help users fix the issue.
func ValidateDescription(description string) error {
	// Trim whitespace for empty check
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return fmt.Errorf("description field is required: please describe what needs approval")
	}

	// Check length against max
	length := len(description)
	if length > MaxDescriptionLength {
		return fmt.Errorf("description is %d characters (max %d): please shorten your request", length, MaxDescriptionLength)
	}

	return nil
}

// ValidateApprover validates the approver user ID for approval requests.
// Returns the user object and an error if:
// - User does not exist in Mattermost
// - User is deleted/deactivated (DeleteAt > 0)
//
// Returns the validated user object to avoid redundant API calls.
// API errors are propagated with proper error wrapping (%w) to preserve error chain.
// Note: Empty string validation is handled by HandleDialogSubmission (Layer 1).
func ValidateApprover(approverID string, api plugin.API) (*model.User, error) {
	// Get user from Mattermost API
	user, appErr := api.GetUser(approverID)
	if appErr != nil {
		return nil, fmt.Errorf("failed to validate approver %s: %w", approverID, appErr)
	}

	// Check if user is deleted/deactivated
	if user.DeleteAt > 0 {
		return nil, fmt.Errorf("selected approver is not a valid user: please select an active user")
	}

	return user, nil
}
