package approval

import (
	"errors"
)

// ApprovalRecord represents a complete approval request and its decision history
type ApprovalRecord struct {
	// Identity
	ID   string `json:"id"`   // 26-char Mattermost ID (model.NewId())
	Code string `json:"code"` // Human-friendly: "A-X7K9Q2"

	// Requester (snapshot at creation time)
	RequesterID          string `json:"requesterId"`
	RequesterUsername    string `json:"requesterUsername"`
	RequesterDisplayName string `json:"requesterDisplayName"`

	// Approver (snapshot at creation time)
	ApproverID          string `json:"approverId"`
	ApproverUsername    string `json:"approverUsername"`
	ApproverDisplayName string `json:"approverDisplayName"`

	// Request details
	Description string `json:"description"`

	// State
	Status          string `json:"status"` // "pending" | "approved" | "denied" | "canceled"
	DecisionComment string `json:"decisionComment,omitempty"`

	// Timestamps (UTC epoch milliseconds)
	CreatedAt int64 `json:"createdAt"`
	DecidedAt int64 `json:"decidedAt"` // 0 if pending

	// Cancellation fields (v0.2.0+)
	CanceledReason  string `json:"canceledReason,omitempty"`  // Reason for cancellation
	CanceledDetails string `json:"canceledDetails,omitempty"` // Additional context (optional, Story 7.3)
	CanceledAt      int64  `json:"canceledAt,omitempty"`      // Timestamp when canceled

	// Verification fields (v0.3.0+) - requester confirms action completion
	Verified            bool   `json:"verified"`                      // Whether requester verified completion
	VerifiedAt          int64  `json:"verifiedAt,omitempty"`          // Timestamp when verified (0 if not verified)
	VerificationComment string `json:"verificationComment,omitempty"` // Optional comment from requester

	// Context
	RequestChannelID string `json:"requestChannelId"`
	TeamID           string `json:"teamId,omitempty"`

	// Delivery tracking flags
	NotificationSent   bool   `json:"notificationSent"`
	NotificationPostID string `json:"notificationPostId,omitempty"` // Post ID of the DM notification with buttons
	OutcomeNotified    bool   `json:"outcomeNotified"`

	// Playbook Integration fields (v2.0 - Epic 8)
	PlaybookRunID     string `json:"playbookRunId,omitempty"`     // ID of the associated playbook run
	PlaybookName      string `json:"playbookName,omitempty"`      // Display name of the playbook run
	PlaybookChannelID string `json:"playbookChannelId,omitempty"` // Channel where playbook is running
	PlaybookPostID    string `json:"playbookPostId,omitempty"`    // Post ID in playbook channel where status was posted

	// Schema versioning
	SchemaVersion int `json:"schemaVersion"`
}

// Status constants for ApprovalRecord
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusDenied   = "denied"
	StatusCanceled = "canceled"
	StatusTimeout  = "timeout" // Story 9.8: Timeout status for expired approvals
)

// Schema version constant
const CurrentSchemaVersion = 1

// Sentinel errors for common approval record operations
var (
	// ErrRecordNotFound is returned when an approval record does not exist
	ErrRecordNotFound = errors.New("approval record not found")

	// ErrRecordImmutable is returned when attempting to modify an immutable record
	ErrRecordImmutable = errors.New("approval record is immutable")

	// ErrInvalidStatus is returned when an invalid status transition is attempted
	ErrInvalidStatus = errors.New("invalid status transition")
)
