package notifications

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyDMError(t *testing.T) {
	tests := []struct {
		name               string
		err                error
		expectedErrorType  string
		expectedSuggestion string
	}{
		{
			name:               "DMs disabled - pattern 1",
			err:                fmt.Errorf("direct_messages_disabled: user has disabled direct messages"),
			expectedErrorType:  "user_dms_disabled",
			expectedSuggestion: "User has DMs disabled. Ask user to enable DMs from bots in Settings > Advanced > Allow Direct Messages From.",
		},
		{
			name:               "DMs disabled - pattern 2",
			err:                fmt.Errorf("failed to create channel: DMs disabled by user"),
			expectedErrorType:  "user_dms_disabled",
			expectedSuggestion: "User has DMs disabled. Ask user to enable DMs from bots in Settings > Advanced > Allow Direct Messages From.",
		},
		{
			name:               "Bot blocked - pattern 1",
			err:                fmt.Errorf("user_blocked_bot: user has blocked the bot"),
			expectedErrorType:  "bot_blocked",
			expectedSuggestion: "User has blocked the bot. User must unblock the bot to receive notifications.",
		},
		{
			name:               "Bot blocked - pattern 2",
			err:                fmt.Errorf("cannot send message: bot is blocked by user"),
			expectedErrorType:  "bot_blocked",
			expectedSuggestion: "User has blocked the bot. User must unblock the bot to receive notifications.",
		},
		{
			name:               "User not found - pattern 1",
			err:                fmt.Errorf("user_not_found: user ID does not exist"),
			expectedErrorType:  "user_not_found",
			expectedSuggestion: "User account not found. User may have been deleted.",
		},
		{
			name:               "User not found - pattern 2 (404)",
			err:                fmt.Errorf("GetUser failed: 404 not found"),
			expectedErrorType:  "user_not_found",
			expectedSuggestion: "User account not found. User may have been deleted.",
		},
		{
			name:               "Generic API error",
			err:                fmt.Errorf("connection timeout: failed to reach server"),
			expectedErrorType:  "api_error",
			expectedSuggestion: "Generic API error. Check Mattermost server logs for details.",
		},
		{
			name:               "Network error",
			err:                fmt.Errorf("network unreachable"),
			expectedErrorType:  "api_error",
			expectedSuggestion: "Generic API error. Check Mattermost server logs for details.",
		},
		{
			name:               "Channel creation failed - generic",
			err:                fmt.Errorf("failed to create DM channel"),
			expectedErrorType:  "api_error",
			expectedSuggestion: "Generic API error. Check Mattermost server logs for details.",
		},
		{
			name:               "404 without user context should be generic",
			err:                fmt.Errorf("404 route not found in API gateway"),
			expectedErrorType:  "api_error",
			expectedSuggestion: "Generic API error. Check Mattermost server logs for details.",
		},
		{
			name:               "404 with user context should be user_not_found",
			err:                fmt.Errorf("GetUser failed: 404 user not found"),
			expectedErrorType:  "user_not_found",
			expectedSuggestion: "User account not found. User may have been deleted.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorType, suggestion := ClassifyDMError(tt.err)

			assert.Equal(t, tt.expectedErrorType, errorType, "Error type mismatch")
			assert.Equal(t, tt.expectedSuggestion, suggestion, "Suggestion mismatch")
		})
	}
}

func TestClassifyDMError_NilError(t *testing.T) {
	// Edge case: nil error should return generic api_error
	errorType, suggestion := ClassifyDMError(nil)

	assert.Equal(t, "api_error", errorType)
	assert.Equal(t, "Generic API error. Check Mattermost server logs for details.", suggestion)
}

func TestClassifyDMError_EmptyError(t *testing.T) {
	// Edge case: empty error message should return generic api_error
	errorType, suggestion := ClassifyDMError(fmt.Errorf(""))

	assert.Equal(t, "api_error", errorType)
	assert.Equal(t, "Generic API error. Check Mattermost server logs for details.", suggestion)
}
