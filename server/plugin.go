package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/command"
	"github.com/mattermost/mattermost-plugin-approver2/server/notifications"
	"github.com/mattermost/mattermost-plugin-approver2/server/store"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// store provides persistence for approval records
	store *store.KVStore

	// service provides approval business logic
	service *approval.Service

	// botUserID is the ID of the bot user for sending notifications
	botUserID string
}

// OnActivate is called when the plugin is activated.
func (p *Plugin) OnActivate() error {
	p.API.LogInfo("Activating Mattermost Approval Workflow plugin")

	// Ensure bot user exists and get bot ID
	botID, appErr := p.API.EnsureBotUser(&model.Bot{
		Username:    "approvalbot",
		DisplayName: "Approval Bot",
		Description: "Bot for sending approval request notifications",
	})
	if appErr != nil {
		return fmt.Errorf("failed to ensure bot user: %s", appErr.Error())
	}
	p.botUserID = botID

	// Initialize store
	p.store = store.NewKVStore(p.API)

	// Initialize approval service
	p.service = approval.NewService(p.store, p.API, botID)

	// Register slash command
	if err := p.registerCommand(); err != nil {
		return fmt.Errorf("failed to register slash command: %w", err)
	}

	p.API.LogInfo("Mattermost Approval Workflow plugin activated successfully", "bot_user_id", botID)
	return nil
}

// registerCommand registers the /approve slash command
func (p *Plugin) registerCommand() error {
	cmd := &model.Command{
		Trigger:          "approve",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage approval requests",
		AutoCompleteHint: "[new|list|get|cancel|status|help]",
		DisplayName:      "Approval Request",
		Description:      "Create, manage, and view approval requests",
	}

	if err := p.API.RegisterCommand(cmd); err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand API.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Parse command into parts
	split := strings.Fields(args.Command)
	if len(split) < 2 {
		// Use router for help/empty command
		router := command.NewRouter(p.API, p.store)
		response, _ := router.Route(args)
		return response, nil
	}

	subcommand := split[1]

	// Handle cancel command directly (bypass router for direct commands)
	if subcommand == "cancel" {
		return p.handleCancelCommand(args, split), nil
	}

	// For other commands, use the router
	router := command.NewRouter(p.API, p.store)
	response, err := router.Route(args)
	if err != nil {
		p.API.LogError("Command execution failed", "error", err.Error(), "command", args.Command)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "An error occurred while processing your command. Please try again.",
		}, nil
	}

	return response, nil
}

// handleCancelCommand processes the /approve cancel <ID> command
func (p *Plugin) handleCancelCommand(args *model.CommandArgs, split []string) *model.CommandResponse {
	// Validate command format: /approve cancel <ID>
	if len(split) < 3 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Usage: /approve cancel <APPROVAL_ID>",
		}
	}

	// Reject extra arguments
	if len(split) > 3 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Usage: /approve cancel <APPROVAL_ID>\n\nError: Too many arguments provided.",
		}
	}

	approvalCode := split[2]
	requesterID := args.UserId

	// TODO (Story 4.3): Replace with modal to collect cancellation reason
	// For now, use default reason until modal is implemented
	reason := "No longer needed"

	// Get the requester user for post update
	requester, appErr := p.API.GetUser(requesterID)
	if appErr != nil {
		p.API.LogError("Failed to get requester user",
			"error", appErr.Error(),
			"user_id", requesterID,
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Failed to retrieve user information. Please try again.",
		}
	}

	// Call service to cancel approval
	err := p.service.CancelApproval(approvalCode, requesterID, reason)
	if err != nil {
		// Log error with context
		p.API.LogError("Failed to cancel approval",
			"error", err.Error(),
			"approval_code", approvalCode,
			"user_id", requesterID,
		)

		// Determine user-friendly error message
		errorMsg := p.formatCancelError(err, approvalCode)

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         errorMsg,
		}
	}

	// Update the approver's DM post to show cancellation (Story 4.1)
	// Get the updated record with cancellation data
	updatedRecord, err := p.store.GetByCode(approvalCode)
	if err != nil {
		p.API.LogWarn("Failed to retrieve updated record for post update",
			"error", err.Error(),
			"approval_code", approvalCode,
		)
		// Continue - don't fail the cancellation if post update fails
	} else {
		// Update the original post (Story 4.1)
		err = notifications.UpdateApprovalPostForCancellation(p.API, updatedRecord, requester.Username)
		if err != nil {
			p.API.LogWarn("Failed to update approver post",
				"error", err.Error(),
				"approval_code", approvalCode,
				"approver_post_id", updatedRecord.NotificationPostID,
			)
			// Continue - post update is best-effort
		}

		// Send cancellation notification DM (Story 4.2)
		_, err = notifications.SendCancellationNotificationDM(p.API, p.botUserID, updatedRecord, requester.Username)
		if err != nil {
			p.API.LogWarn("Failed to send cancellation notification, cancellation still successful",
				"error", err.Error(),
				"approval_id", updatedRecord.ID,
				"approver_id", updatedRecord.ApproverID,
			)
			// Continue - notification is best-effort
		}
	}

	// Success - send ephemeral confirmation
	successMsg := fmt.Sprintf("✅ Approval request `%s` has been canceled.", approvalCode)

	post := &model.Post{
		UserId:    "",
		ChannelId: args.ChannelId,
		Message:   successMsg,
	}

	ephemeralPost := p.API.SendEphemeralPost(requesterID, post)
	if ephemeralPost == nil {
		p.API.LogError("Failed to send ephemeral confirmation post",
			"user_id", requesterID,
			"approval_code", approvalCode,
		)
		// Fallback to regular post (AC3 pattern from Story 1.6)
		post.UserId = requesterID
		if _, appErr := p.API.CreatePost(post); appErr != nil {
			p.API.LogError("Failed to send fallback confirmation post",
				"user_id", requesterID,
				"approval_code", approvalCode,
				"error", appErr.Error(),
			)
		}
	}

	p.API.LogInfo("Approval canceled successfully",
		"approval_code", approvalCode,
		"user_id", requesterID,
	)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
	}
}

// formatCancelError converts service errors into user-friendly messages
func (p *Plugin) formatCancelError(err error, code string) string {
	errorStr := err.Error()

	// Check for specific error types
	switch {
	case strings.Contains(errorStr, "invalid approval code format"):
		return fmt.Sprintf("❌ Invalid approval code format: '%s'. Expected format like 'A-X7K9Q2'.", code)
	case strings.Contains(errorStr, "approval record not found"):
		return fmt.Sprintf("❌ Approval request '%s' not found. Use `/approve list` to see your requests.", code)
	case strings.Contains(errorStr, "permission denied"):
		return "❌ Permission denied. You can only cancel your own approval requests."
	case strings.Contains(errorStr, "cannot cancel"):
		// Extract status from error message if present
		switch {
		case strings.Contains(errorStr, "approved"):
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Status is already approved.", code)
		case strings.Contains(errorStr, "denied"):
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Status is already denied.", code)
		case strings.Contains(errorStr, "canceled"):
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Status is already canceled.", code)
		default:
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Request has already been decided.", code)
		}
	default:
		return "❌ Failed to cancel approval request. Please try again."
	}
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
