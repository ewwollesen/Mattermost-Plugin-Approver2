package approval

import (
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

// GenerateID creates a new 26-character Mattermost ID
func GenerateID() string {
	return model.NewId()
}

// NewApprovalRecord creates a new ApprovalRecord with generated ID and unique Code
func NewApprovalRecord(
	store Storer,
	requesterID, requesterUsername, requesterDisplayName string,
	approverID, approverUsername, approverDisplayName string,
	description, channelID, teamID string,
) (*ApprovalRecord, error) {
	now := time.Now().UnixMilli()

	// Generate unique code with collision checking
	code, err := GenerateUniqueCode(store)
	if err != nil {
		return nil, err
	}

	return &ApprovalRecord{
		ID:   GenerateID(),
		Code: code,

		RequesterID:          requesterID,
		RequesterUsername:    requesterUsername,
		RequesterDisplayName: requesterDisplayName,

		ApproverID:          approverID,
		ApproverUsername:    approverUsername,
		ApproverDisplayName: approverDisplayName,

		Description: description,

		Status:          StatusPending,
		DecisionComment: "",

		CreatedAt: now,
		DecidedAt: 0,

		RequestChannelID: channelID,
		TeamID:           teamID,

		NotificationSent: false,
		OutcomeNotified:  false,

		SchemaVersion: CurrentSchemaVersion,
	}, nil
}
