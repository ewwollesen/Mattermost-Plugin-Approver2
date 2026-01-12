package command

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Router routes slash command invocations to appropriate handlers
type Router struct {
	api plugin.API
}

// NewRouter creates a new command router
func NewRouter(api plugin.API) *Router {
	return &Router{
		api: api,
	}
}

// Route determines which handler should process the command
func (r *Router) Route(args *model.CommandArgs) (*model.CommandResponse, error) {
	split := strings.Fields(args.Command)
	if len(split) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Invalid command format.",
		}, nil
	}

	// Extract subcommand (first arg after /approve)
	subcommand := ""
	if len(split) > 1 {
		subcommand = split[1]
	}

	// Route to appropriate handler
	switch subcommand {
	case "":
		return executeHelp(), nil
	case "help":
		return executeHelp(), nil
	case "new":
		return r.executeNew(args)
	default:
		return executeUnknown(subcommand), nil
	}
}

// executeHelp returns help text
func executeHelp() *model.CommandResponse {
	helpText := `### Mattermost Approval Workflow

**Available Commands:**

* **/approve new** - Create a new approval request
* **/approve list** - View your approval requests and decisions
* **/approve get [ID]** - View a specific approval by ID
* **/approve cancel <APPROVAL_ID>** - Cancel a pending approval request
* **/approve help** - Display this help text

**Example:**
` + "`/approve new`" + ` - Opens a modal to create an approval request

For more information, visit the plugin documentation.`

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         helpText,
	}
}

// executeUnknown returns error for unrecognized commands
func executeUnknown(subcommand string) *model.CommandResponse {
	errorText := fmt.Sprintf("Unknown command: **%s**\n\nValid commands: `new`, `list`, `get`, `cancel`, `help`\n\nType `/approve help` for more information.", subcommand)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         errorText,
	}
}

// executeNew opens an interactive dialog for creating a new approval request
func (r *Router) executeNew(args *model.CommandArgs) (*model.CommandResponse, error) {
	// Validate trigger ID is present
	if args.TriggerId == "" {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Missing trigger ID. Please try the command again.",
		}, nil
	}

	// Get plugin site URL for callback
	siteURL := r.api.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil || *siteURL == "" {
		r.api.LogError("Site URL not configured", "site_url", "empty or nil", "command", "approve new")
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Plugin configuration error. Please contact your system administrator.",
		}, nil
	}

	callbackURL := fmt.Sprintf("%s/plugins/com.mattermost.plugin-approver2/dialog/submit", *siteURL)

	// Define the dialog structure
	dialog := model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       callbackURL,
		Dialog: model.Dialog{
			Title:       "Create Approval Request",
			SubmitLabel: "Submit Request",
			CallbackId:  "approve_new",
			Elements: []model.DialogElement{
				{
					DisplayName: "Select approver *",
					Name:        "approver",
					Type:        "select",
					DataSource:  "users",
				},
				{
					DisplayName: "What needs approval? *",
					Name:        "description",
					Type:        "textarea",
					Placeholder: "Describe the action requiring approval",
					MaxLength:   1000,
				},
			},
		},
	}

	// Open the interactive dialog
	if appErr := r.api.OpenInteractiveDialog(dialog); appErr != nil {
		r.api.LogError("Failed to open interactive dialog", "error", appErr.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Failed to open approval request form. Please try again.",
		}, nil
	}

	// Return empty response - modal opens asynchronously
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
	}, nil
}
