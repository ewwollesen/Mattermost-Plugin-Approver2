package notifications

import (
	"strings"
)

// ClassifyDMError analyzes a DM notification error and returns both an error type
// classification and a user-friendly suggestion for resolution.
//
// This function implements AC6 from Story 2.6: providing specific error details and
// resolution steps for different types of notification failures.
//
// Error classifications:
//   - "user_dms_disabled": User has disabled DMs from bots in their settings
//   - "bot_blocked": User has blocked the plugin bot
//   - "user_not_found": User account doesn't exist or was deleted
//   - "api_error": Generic API error or unclassified failure
//
// Returns:
//   - errorType: classification string for logging and monitoring
//   - suggestion: actionable resolution steps for administrators
func ClassifyDMError(err error) (errorType, suggestion string) {
	// Handle nil error edge case
	if err == nil {
		return "api_error",
			"Generic API error. Check Mattermost server logs for details."
	}

	errMsg := err.Error()

	// Check for DMs disabled patterns
	// Pattern matching for various Mattermost error messages indicating DMs are disabled
	if strings.Contains(errMsg, "direct_messages_disabled") ||
		strings.Contains(errMsg, "DMs disabled") {
		return "user_dms_disabled",
			"User has DMs disabled. Ask user to enable DMs from bots in Settings > Advanced > Allow Direct Messages From."
	}

	// Check for bot blocked patterns
	// Pattern matching for cases where user has explicitly blocked the bot
	if strings.Contains(errMsg, "user_blocked_bot") ||
		strings.Contains(errMsg, "bot is blocked") {
		return "bot_blocked",
			"User has blocked the bot. User must unblock the bot to receive notifications."
	}

	// Check for user not found patterns
	// Pattern matching for deleted or non-existent user accounts
	// More specific 404 check to avoid false positives (e.g., "404 route not found")
	if strings.Contains(errMsg, "user_not_found") ||
		(strings.Contains(errMsg, "404") && (strings.Contains(errMsg, "user") || strings.Contains(errMsg, "User"))) {
		return "user_not_found",
			"User account not found. User may have been deleted."
	}

	// Default to generic API error for all other cases
	// This includes network errors, timeouts, permission issues, etc.
	return "api_error",
		"Generic API error. Check Mattermost server logs for details."
}
