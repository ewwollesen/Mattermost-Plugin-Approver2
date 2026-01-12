package notifications

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// SendApprovalRequestDM sends a DM notification to the approver when a new approval request is created.
// The message includes complete context: requester info, timestamp, description, and request ID.
// Returns the post ID and error. Error returned if DM send fails (caller should log and handle gracefully).
func SendApprovalRequestDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
	// Validate inputs
	if botUserID == "" {
		return "", fmt.Errorf("bot user ID not available")
	}
	if record == nil {
		return "", fmt.Errorf("approval record is nil")
	}
	if record.ID == "" {
		return "", fmt.Errorf("approval record ID is empty")
	}

	// Get or create DM channel between bot and approver
	channelID, err := GetDMChannelID(api, botUserID, record.ApproverID)
	if err != nil {
		return "", fmt.Errorf("failed to get DM channel for approver %s: %w", record.ApproverID, err)
	}

	// Format timestamp as YYYY-MM-DD HH:MM:SS UTC (AC2 requirement)
	timestamp := time.UnixMilli(record.CreatedAt).UTC()
	timestampStr := timestamp.Format("2006-01-02 15:04:05 MST")

	// Construct DM message with exact format from AC2
	message := fmt.Sprintf("📋 **Approval Request**\n\n"+
		"**From:** @%s (%s)\n"+
		"**Requested:** %s\n"+
		"**Description:**\n%s\n\n"+
		"**Request ID:** `%s`",
		record.RequesterUsername,
		record.RequesterDisplayName,
		timestampStr,
		record.Description,
		record.Code)

	// Create post with interactive action buttons
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: channelID,
		Message:   message,
		Props: model.StringInterface{
			"attachments": []any{
				map[string]any{
					"actions": []any{
						// Approve button (green/primary style)
						map[string]any{
							"name": "Approve",
							"integration": map[string]any{
								"url": "/plugins/com.mattermost.plugin-approver2/action",
								"context": map[string]any{
									"approval_id": record.ID,
									"action":      "approve",
								},
							},
							"style": "primary",
						},
						// Deny button (red/danger style)
						map[string]any{
							"name": "Deny",
							"integration": map[string]any{
								"url": "/plugins/com.mattermost.plugin-approver2/action",
								"context": map[string]any{
									"approval_id": record.ID,
									"action":      "deny",
								},
							},
							"style": "danger",
						},
					},
				},
			},
		},
	}

	// Send DM via CreatePost (persistent message, not ephemeral)
	createdPost, appErr := api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to send DM to approver %s: %w", record.ApproverID, appErr)
	}

	return createdPost.Id, nil
}

// GetDMChannelID gets or creates a DM channel between the bot and the target user.
// Returns the channel ID if successful, or an error if the channel cannot be created.
func GetDMChannelID(api plugin.API, botUserID, targetUserID string) (string, error) {
	// Get or create DM channel (creates if doesn't exist)
	channel, appErr := api.GetDirectChannel(botUserID, targetUserID)
	if appErr != nil {
		return "", fmt.Errorf("failed to get DM channel for user %s: %w", targetUserID, appErr)
	}

	if channel == nil {
		return "", fmt.Errorf("DM channel is nil for user %s", targetUserID)
	}

	return channel.Id, nil
}
