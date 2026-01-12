package command

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

// HandleDialogSubmission validates a dialog submission and returns validation errors if any.
// Performs basic presence validation for required fields:
// - approver: Must be present and non-empty
// - description: Must be present and non-empty
//
// Returns field-specific errors that keep the modal open, preserving user input.
// Error messages follow UX guidelines: specific, actionable, helpful tone.
func HandleDialogSubmission(submission map[string]any) *model.SubmitDialogResponse {
	response := &model.SubmitDialogResponse{
		Errors: make(map[string]string),
	}

	// Validate approver field (AC1: Validate Missing Approver)
	approver, ok := submission["approver"].(string)
	if !ok || approver == "" {
		response.Errors["approver"] = "Approver field is required. Please select a user."
	}

	// Validate description field (AC2: Validate Missing Description)
	description, ok := submission["description"].(string)
	if !ok || description == "" {
		response.Errors["description"] = "Description field is required. Please describe what needs approval."
	}

	return response
}

// ParseDialogSubmissionPayload parses the dialog submission request from JSON
func ParseDialogSubmissionPayload(data []byte) (*model.SubmitDialogRequest, error) {
	var payload model.SubmitDialogRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse dialog submission payload: %w", err)
	}
	return &payload, nil
}
