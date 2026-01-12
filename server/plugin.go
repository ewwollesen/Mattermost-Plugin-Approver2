package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/command"
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
}

// OnActivate is called when the plugin is activated.
func (p *Plugin) OnActivate() error {
	p.API.LogInfo("Activating Mattermost Approval Workflow plugin")

	// Initialize store
	p.store = store.NewKVStore(p.API)

	// Initialize approval service
	p.service = approval.NewService(p.store, p.API)

	// Register slash command
	if err := p.registerCommand(); err != nil {
		return fmt.Errorf("failed to register slash command: %w", err)
	}

	p.API.LogInfo("Mattermost Approval Workflow plugin activated successfully")
	return nil
}

// registerCommand registers the /approve slash command
func (p *Plugin) registerCommand() error {
	cmd := &model.Command{
		Trigger:          "approve",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage approval requests",
		AutoCompleteHint: "[new|list|get|cancel|help]",
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
		router := command.NewRouter(p.API)
		response, _ := router.Route(args)
		return response, nil
	}

	subcommand := split[1]

	// Handle cancel command directly (bypass router for direct commands)
	if subcommand == "cancel" {
		return p.handleCancelCommand(args, split), nil
	}

	// For other commands, use the router
	router := command.NewRouter(p.API)
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

	// Call service to cancel approval
	err := p.service.CancelApproval(approvalCode, requesterID)
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
		if strings.Contains(errorStr, "approved") {
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Status is already approved.", code)
		} else if strings.Contains(errorStr, "denied") {
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Status is already denied.", code)
		} else if strings.Contains(errorStr, "canceled") {
			return fmt.Sprintf("❌ Cannot cancel approval request %s. Status is already canceled.", code)
		}
		return fmt.Sprintf("❌ Cannot cancel approval request %s. Request has already been decided.", code)
	default:
		return "❌ Failed to cancel approval request. Please try again."
	}
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
