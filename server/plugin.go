package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/command"
	"github.com/mattermost/mattermost-plugin-approver2/server/notifications"
	"github.com/mattermost/mattermost-plugin-approver2/server/store"
	"github.com/mattermost/mattermost-plugin-approver2/server/timeout"
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

	// timeoutChecker periodically scans for timed-out pending requests
	timeoutChecker *timeout.TimeoutChecker

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

	// Initialize and start timeout checker (Story 6.1)
	p.timeoutChecker = timeout.NewChecker(p.store, p.service, p.API, botID)
	p.timeoutChecker.Start()

	// Register slash command
	if err := p.registerCommand(); err != nil {
		return fmt.Errorf("failed to register slash command: %w", err)
	}

	p.API.LogInfo("Mattermost Approval Workflow plugin activated successfully", "bot_user_id", botID)
	return nil
}

// OnDeactivate is called when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	p.API.LogInfo("Deactivating Mattermost Approval Workflow plugin")

	// Stop timeout checker gracefully (Story 6.1)
	if p.timeoutChecker != nil {
		p.timeoutChecker.Stop()
	}

	p.API.LogInfo("Mattermost Approval Workflow plugin deactivated successfully")
	return nil
}

// registerCommand registers the /approve slash command
func (p *Plugin) registerCommand() error {
	cmd := &model.Command{
		Trigger:          "approve",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage approval requests",
		AutoCompleteHint: "[new|list [pending|approved|denied|canceled|all]|get|cancel|verify|status|help]",
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

	// Handle verify command directly (Story 6.2)
	if subcommand == "verify" {
		return p.handleVerifyCommand(args, split), nil
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
// Story 4.3: Opens a modal to collect cancellation reason instead of immediate cancel
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

	// Validate approval exists and get record BEFORE opening modal
	record, err := p.store.GetByCode(approvalCode)
	if err != nil {
		// Log error with context
		p.API.LogError("Failed to get approval record for cancellation",
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

	// Verify user is the requester (permission check)
	if record.RequesterID != requesterID {
		p.API.LogError("Unauthorized cancellation attempt",
			"approval_code", approvalCode,
			"authenticated_user", requesterID,
			"requester_id", record.RequesterID,
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Permission denied. You can only cancel your own approval requests.",
		}
	}

	// Open cancellation modal (Story 4.3)
	if err := p.openCancellationModal(args.TriggerId, approvalCode, record); err != nil {
		p.API.LogError("Failed to open cancellation modal",
			"error", err.Error(),
			"approval_code", approvalCode,
			"user_id", requesterID,
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Failed to open cancellation dialog. Please try again.",
		}
	}

	// Success - modal opened (actual cancellation happens in modal submission)
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
	}
}

// openCancellationModal opens an interactive dialog for cancellation reason selection
// Story 4.3: Collect structured cancellation reasons for data instrumentation
func (p *Plugin) openCancellationModal(triggerID string, approvalCode string, record *approval.ApprovalRecord) error {
	// Create interactive dialog
	dialog := model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       "/plugins/com.mattermost.plugin-approver2/dialog/submit",
		Dialog: model.Dialog{
			CallbackId:       fmt.Sprintf("cancel_approval_%s", record.ID),
			Title:            "Cancel Approval Request",
			IntroductionText: fmt.Sprintf("You are about to cancel approval request **%s**\n\n**This action cannot be undone.**\n\nPlease select a reason for cancellation:", approvalCode),
			Elements: []model.DialogElement{
				{
					DisplayName: "Reason for cancellation",
					Name:        "reason_code",
					Type:        "select",
					Placeholder: "Select a reason",
					Options: []*model.PostActionOptions{
						{Text: "No longer needed", Value: "no_longer_needed"},
						{Text: "Wrong approver", Value: "wrong_approver"},
						{Text: "Sensitive information", Value: "sensitive_info"},
						{Text: "Other reason", Value: "other"},
					},
					Default:  "no_longer_needed",
					Optional: false,
					HelpText: "Help us understand why requests are canceled",
				},
				{
					DisplayName: "Additional details (if Other)",
					Name:        "other_reason_text",
					Type:        "textarea",
					Placeholder: "Please explain...",
					Optional:    true,
					MaxLength:   500,
				},
			},
			SubmitLabel: "Cancel Request",
		},
	}

	if appErr := p.API.OpenInteractiveDialog(dialog); appErr != nil {
		return fmt.Errorf("failed to open cancellation modal: %w", appErr)
	}

	return nil
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

// handleVerifyCommand processes the /approve verify <CODE> [comment] command
// Story 6.2: Allows requester to mark approved requests as verified
func (p *Plugin) handleVerifyCommand(args *model.CommandArgs, split []string) *model.CommandResponse {
	// Validate command format: /approve verify <CODE> [comment]
	if len(split) < 3 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Usage: /approve verify <APPROVAL_CODE> [comment]\n\nExample: `/approve verify A-X7K9Q2` or `/approve verify A-X7K9Q2 Deployment completed successfully`",
		}
	}

	approvalCode := split[2]
	requesterID := args.UserId

	// Extract optional comment (everything after the code)
	comment := ""
	if len(split) > 3 {
		comment = strings.Join(split[3:], " ")
	}

	// Validate comment length (AC7: max 500 characters)
	if len(comment) > 500 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("❌ Verification comment too long (%d characters). Maximum length is 500 characters.", len(comment)),
		}
	}

	// Validate approval exists and get record
	record, err := p.store.GetByCode(approvalCode)
	if err != nil {
		p.API.LogError("Failed to get approval record for verification",
			"error", err.Error(),
			"approval_code", approvalCode,
			"user_id", requesterID,
		)

		errorMsg := p.formatVerifyError(err, approvalCode)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         errorMsg,
		}
	}

	// Verify user is the requester (AC4: permission check)
	if record.RequesterID != requesterID {
		p.API.LogError("Unauthorized verification attempt",
			"approval_code", approvalCode,
			"authenticated_user", requesterID,
			"requester_id", record.RequesterID,
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Permission denied. Only the requester can verify their approval requests.",
		}
	}

	// Verify status is approved (AC5)
	if record.Status != approval.StatusApproved {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("❌ Cannot verify approval request %s. Only approved requests can be verified (current status: %s).", approvalCode, record.Status),
		}
	}

	// Verify not already verified (AC6)
	if record.Verified {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("❌ Approval request %s has already been verified.", approvalCode),
		}
	}

	// Call service to verify request (updates record)
	err = p.service.VerifyRequest(approvalCode, requesterID, comment)
	if err != nil {
		p.API.LogError("Failed to verify approval request",
			"error", err.Error(),
			"approval_code", approvalCode,
			"user_id", requesterID,
		)

		errorMsg := p.formatVerifyError(err, approvalCode)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         errorMsg,
		}
	}

	// Reload record to get updated verification fields
	updatedRecord, err := p.store.GetByCode(approvalCode)
	if err != nil {
		p.API.LogError("Failed to reload verified record for notification",
			"approval_code", approvalCode,
			"error", err.Error(),
		)
		// Continue - notification failure doesn't block success response
	} else {
		// Send verification notification to approver (best-effort, graceful degradation)
		if _, err := notifications.SendVerificationNotificationDM(p.API, p.botUserID, updatedRecord); err != nil {
			p.API.LogWarn("Failed to send verification notification",
				"approval_code", approvalCode,
				"approver_id", record.ApproverID,
				"error", err.Error(),
			)
			// Continue - notification failure doesn't affect verification
		}
	}

	// Success response
	successMsg := fmt.Sprintf("✅ Approval request **%s** marked as verified.", approvalCode)
	if comment != "" {
		successMsg += fmt.Sprintf("\n\n**Verification note:** %s", comment)
	}
	successMsg += fmt.Sprintf("\n\n%s has been notified.", record.ApproverDisplayName)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         successMsg,
	}
}

// formatVerifyError converts service errors into user-friendly messages for verify command
func (p *Plugin) formatVerifyError(err error, code string) string {
	errorStr := err.Error()

	// Check for specific error types
	switch {
	case strings.Contains(errorStr, "invalid approval code format"):
		return fmt.Sprintf("❌ Invalid approval code format: '%s'. Expected format like 'A-X7K9Q2'.", code)
	case strings.Contains(errorStr, "approval record not found"):
		return fmt.Sprintf("❌ Approval request '%s' not found. Use `/approve list` to see your requests.", code)
	case strings.Contains(errorStr, "permission denied"):
		return "❌ Permission denied. Only the requester can verify their approval requests."
	case strings.Contains(errorStr, "not approved"):
		return fmt.Sprintf("❌ Cannot verify approval request %s. Only approved requests can be verified.", code)
	case strings.Contains(errorStr, "already verified"):
		return fmt.Sprintf("❌ Approval request %s has already been verified.", code)
	default:
		return "❌ Failed to verify approval request. Please try again."
	}
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
