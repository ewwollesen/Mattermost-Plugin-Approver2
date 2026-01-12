package command

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
)

func TestHandleDialogSubmission(t *testing.T) {
	t.Run("successfully validates both fields present", func(t *testing.T) {
		submission := map[string]any{
			"approver":    "user-id-abcdefghijklmnopqrstuvwxyz",
			"description": "Test approval request description",
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.Empty(t, response.Error)
		assert.Empty(t, response.Errors)
	})

	t.Run("returns error for missing approver field", func(t *testing.T) {
		submission := map[string]any{
			"description": "Test approval request description",
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "approver")
		assert.Contains(t, response.Errors["approver"], "Approver field is required")
		assert.Contains(t, response.Errors["approver"], "Please select a user")
	})

	t.Run("returns error for empty approver field", func(t *testing.T) {
		submission := map[string]any{
			"approver":    "",
			"description": "Test approval request description",
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "approver")
		assert.Contains(t, response.Errors["approver"], "Approver field is required")
	})

	t.Run("returns error for missing description field", func(t *testing.T) {
		submission := map[string]any{
			"approver": "user-id-abcdefghijklmnopqrstuvwxyz",
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "description")
		assert.Contains(t, response.Errors["description"], "Description field is required")
		assert.Contains(t, response.Errors["description"], "Please describe what needs approval")
	})

	t.Run("returns error for empty description field", func(t *testing.T) {
		submission := map[string]any{
			"approver":    "user-id-abcdefghijklmnopqrstuvwxyz",
			"description": "",
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "description")
		assert.Contains(t, response.Errors["description"], "Description field is required")
	})

	t.Run("returns errors for both fields missing", func(t *testing.T) {
		submission := map[string]any{}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "approver")
		assert.Contains(t, response.Errors, "description")
	})

	t.Run("handles non-string approver gracefully", func(t *testing.T) {
		submission := map[string]any{
			"approver":    123, // Invalid type
			"description": "Test approval request description",
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "approver")
	})

	t.Run("handles non-string description gracefully", func(t *testing.T) {
		submission := map[string]any{
			"approver":    "user-id-abcdefghijklmnopqrstuvwxyz",
			"description": []string{"invalid", "type"}, // Invalid type
		}

		response := HandleDialogSubmission(submission)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
		assert.Contains(t, response.Errors, "description")
	})
}

func TestParseDialogSubmissionPayload(t *testing.T) {
	t.Run("successfully parses valid payload", func(t *testing.T) {
		payload := &model.SubmitDialogRequest{
			CallbackId: "approve_new",
			UserId:     "user-id-abcdefghijklmnopqrstuvwxyz",
			ChannelId:  "channel-id-abcdefghijklmnopqrstuvwxyz",
			TeamId:     "team-id-abcdefghijklmnopqrstuvwxyz",
			Submission: map[string]any{
				"approver":    "approver-id-abcdefghijklmnopqrstuvwxyz",
				"description": "Test approval description",
			},
		}

		data, err := json.Marshal(payload)
		assert.NoError(t, err)

		parsed, err := ParseDialogSubmissionPayload(data)
		assert.NoError(t, err)
		assert.NotNil(t, parsed)
		assert.Equal(t, "approve_new", parsed.CallbackId)
		assert.Equal(t, "approver-id-abcdefghijklmnopqrstuvwxyz", parsed.Submission["approver"])
		assert.Equal(t, "Test approval description", parsed.Submission["description"])
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		data := []byte("invalid json")

		parsed, err := ParseDialogSubmissionPayload(data)
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})
}
