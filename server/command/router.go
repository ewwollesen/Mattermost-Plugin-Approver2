package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Storer interface defines methods for accessing approval records
type Storer interface {
	GetAllApprovals() ([]*approval.ApprovalRecord, error)
	GetUserApprovals(userID string) ([]*approval.ApprovalRecord, error)
	GetApprovalByCode(code string) (*approval.ApprovalRecord, error)
}

// Router routes slash command invocations to appropriate handlers
type Router struct {
	api   plugin.API
	store Storer
}

// NewRouter creates a new command router
func NewRouter(api plugin.API, store Storer) *Router {
	return &Router{
		api:   api,
		store: store,
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
	case "list":
		return r.executeList(args)
	case "get":
		return r.executeGet(args)
	case "status":
		return r.executeStatus(args, split[2:])
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
* **/approve status** - View approval system statistics (admin only)
* **/approve help** - Display this help text

**Admin Commands:**

* **/approve status** - View overall approval statistics and notification health
* **/approve status --failed-notifications** - List specific approvals with failed notifications

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
	errorText := fmt.Sprintf("Unknown command: **%s**\n\nValid commands: `new`, `list`, `get`, `cancel`, `status`, `help`\n\nType `/approve help` for more information.", subcommand)

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

// executeStatus displays approval system statistics (admin only)
func (r *Router) executeStatus(args *model.CommandArgs, subargs []string) (*model.CommandResponse, error) {
	// Check if user is system admin
	user, appErr := r.api.GetUser(args.UserId)
	if appErr != nil {
		r.api.LogError("Failed to get user for status command", "user_id", args.UserId, "error", appErr.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Failed to verify permissions. Please try again.",
		}, nil
	}

	// Check if user has system admin role (exact match to prevent bypass)
	// Security: Split roles by space and check for exact "system_admin" match
	// to prevent bypass attacks like "fake_system_admin" or "not_system_admin"
	roles := strings.Fields(user.Roles)
	hasSystemAdmin := false
	for _, role := range roles {
		if role == "system_admin" {
			hasSystemAdmin = true
			break
		}
	}

	if !hasSystemAdmin {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Permission denied. Only system administrators can view approval statistics.",
		}, nil
	}

	// Retrieve all approval records
	records, err := r.store.GetAllApprovals()
	if err != nil {
		r.api.LogError("Failed to retrieve approval records for status command", "error", err.Error())
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to retrieve approval statistics. Please try again.",
		}, nil
	}

	// Check if --failed-notifications flag is present
	showFailedOnly := false
	for _, arg := range subargs {
		if arg == "--failed-notifications" {
			showFailedOnly = true
			break
		}
	}

	// Calculate statistics
	stats := calculateStatistics(records)

	// Format response based on flag
	var responseText string
	if showFailedOnly {
		responseText = formatFailedNotifications(records, stats)
	} else {
		responseText = formatStatusResponse(stats)
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         responseText,
	}, nil
}

// ApprovalStats holds statistics about the approval system
type ApprovalStats struct {
	TotalApprovals              int
	PendingApprovals            int
	ApprovedApprovals           int
	DeniedApprovals             int
	CanceledApprovals           int
	FailedApproverNotifications int
	FailedOutcomeNotifications  int
}

// calculateStatistics computes statistics from approval records
func calculateStatistics(records []*approval.ApprovalRecord) ApprovalStats {
	stats := ApprovalStats{
		TotalApprovals: len(records),
	}

	for _, record := range records {
		switch record.Status {
		case approval.StatusPending:
			stats.PendingApprovals++
			// Count failed approver notifications only for pending approvals
			if !record.NotificationSent {
				stats.FailedApproverNotifications++
			}
		case approval.StatusApproved:
			stats.ApprovedApprovals++
			// Count failed outcome notifications for completed approvals
			if !record.OutcomeNotified {
				stats.FailedOutcomeNotifications++
			}
		case approval.StatusDenied:
			stats.DeniedApprovals++
			// Count failed outcome notifications for completed approvals
			if !record.OutcomeNotified {
				stats.FailedOutcomeNotifications++
			}
		case approval.StatusCanceled:
			stats.CanceledApprovals++
		}
	}

	return stats
}

// formatStatusResponse formats the statistics into a user-friendly message
func formatStatusResponse(stats ApprovalStats) string {
	if stats.TotalApprovals == 0 {
		return "📊 **Approval System Status**\n\nNo approvals in the system yet."
	}

	completedApprovals := stats.ApprovedApprovals + stats.DeniedApprovals + stats.CanceledApprovals

	// Calculate success percentages
	successfulApproverNotifications := stats.TotalApprovals - stats.FailedApproverNotifications
	approverNotifRate := 0.0
	if stats.TotalApprovals > 0 {
		approverNotifRate = (float64(successfulApproverNotifications) / float64(stats.TotalApprovals)) * 100
	}

	successfulOutcomeNotifications := completedApprovals - stats.FailedOutcomeNotifications
	outcomeNotifRate := 0.0
	if completedApprovals > 0 {
		outcomeNotifRate = (float64(successfulOutcomeNotifications) / float64(completedApprovals)) * 100
	}

	// Add warning if we're at or near the query limit
	limitWarning := ""
	if stats.TotalApprovals >= 9500 {
		limitWarning = "\n\n⚠️ **Note:** System has 9,500+ approvals. Statistics may be incomplete for very large deployments."
	}

	return fmt.Sprintf(`**📊 Approval System Status**

**Overall Statistics:**
- Total Approvals: %d
- Pending: %d
- Completed: %d (Approved: %d, Denied: %d, Canceled: %d)

**Notification Health:**
- ✅ Successful Approver Notifications: %d (%.1f%%)
- ❌ Failed Approver Notifications: %d
- ✅ Successful Outcome Notifications: %d (%.1f%%)
- ❌ Failed Outcome Notifications: %d

**Next Steps:**
- Use `+"`/approve status --failed-notifications`"+` to see specific records with failed notifications
- Failed notifications indicate user DM settings issues or deleted accounts
- Consider manual notification via direct message if critical%s`,
		stats.TotalApprovals,
		stats.PendingApprovals,
		completedApprovals, stats.ApprovedApprovals, stats.DeniedApprovals, stats.CanceledApprovals,
		successfulApproverNotifications, approverNotifRate,
		stats.FailedApproverNotifications,
		successfulOutcomeNotifications, outcomeNotifRate,
		stats.FailedOutcomeNotifications,
		limitWarning,
	)
}

// formatFailedNotifications formats records with failed notifications
func formatFailedNotifications(records []*approval.ApprovalRecord, stats ApprovalStats) string {
	if stats.FailedApproverNotifications == 0 && stats.FailedOutcomeNotifications == 0 {
		return "**📋 Approvals with Failed Notifications**\n\n✅ No failed notifications found. All notifications delivered successfully!"
	}

	var message strings.Builder
	message.WriteString("**📋 Approvals with Failed Notifications**\n\n")

	// Failed Approver Notifications
	if stats.FailedApproverNotifications > 0 {
		message.WriteString("**Failed Approver Notifications (NotificationSent=false):**\n")
		count := 0
		for _, record := range records {
			if record.Status == approval.StatusPending && !record.NotificationSent {
				count++
				if count > 50 {
					break // Limit to 50 records
				}
				createdAt := record.CreatedAt / 1000 // Convert to seconds
				message.WriteString(fmt.Sprintf("%d. `%s` | Requester: @%s | Approver: @%s | Status: %s | Created: <t:%d:f>\n",
					count, record.Code, record.RequesterUsername, record.ApproverUsername, record.Status, createdAt))
			}
		}
		message.WriteString("\n")
	}

	// Failed Outcome Notifications
	if stats.FailedOutcomeNotifications > 0 {
		message.WriteString("**Failed Outcome Notifications (OutcomeNotified=false):**\n")
		count := 0
		for _, record := range records {
			if (record.Status == approval.StatusApproved || record.Status == approval.StatusDenied) && !record.OutcomeNotified {
				count++
				if count > 50 {
					break // Limit to 50 records
				}
				decidedAt := record.DecidedAt / 1000 // Convert to seconds
				message.WriteString(fmt.Sprintf("%d. `%s` | Requester: @%s | Approver: @%s | Status: %s | Decided: <t:%d:f>\n",
					count, record.Code, record.RequesterUsername, record.ApproverUsername, record.Status, decidedAt))
			}
		}
		message.WriteString("\n")
	}

	message.WriteString(`**Guidance:**
- Check user DM settings (Settings > Advanced > Allow Direct Messages From)
- Verify users have not blocked the bot
- Consider sending manual DM to affected users if time-sensitive
- Review Mattermost server logs for specific error details`)

	return message.String()
}

// executeList displays all approval records for the authenticated user
func (r *Router) executeList(args *model.CommandArgs) (*model.CommandResponse, error) {
	// Security: args.UserId is authenticated by Mattermost Plugin API
	// The Mattermost server guarantees this ID matches the authenticated user session (NFR-S1)
	// Access control: GetUserApprovals only returns records where this user is requester or approver (NFR-S2, FR37)
	records, err := r.store.GetUserApprovals(args.UserId)
	if err != nil {
		r.api.LogError("Failed to retrieve approval records for list command",
			"user_id", args.UserId,
			"error", err.Error(),
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to retrieve approval records. Please try again.",
		}, nil
	}

	// Handle empty state
	if len(records) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "No approval records found. Use `/approve new` to create a request.",
		}, nil
	}

	// Apply pagination (limit to 20 records)
	total := len(records)
	displayRecords := records
	if total > 20 {
		displayRecords = records[:20]
	}

	// Format and return response
	responseText := formatListResponse(displayRecords, total)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         responseText,
	}, nil
}

// formatListResponse formats approval records into a readable list
func formatListResponse(records []*approval.ApprovalRecord, total int) string {
	var output strings.Builder

	output.WriteString("**Your Approval Records:**\n\n")

	for _, record := range records {
		// Get status icon
		statusIcon := getStatusIcon(record.Status)

		// Format timestamp
		createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
		formattedDate := createdTime.UTC().Format("2006-01-02 15:04")

		// Format record line
		output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
			record.Code,
			statusIcon,
			record.RequesterUsername,
			record.ApproverUsername,
			formattedDate,
		))
	}

	// Add pagination footer if needed
	if total > 20 {
		output.WriteString(fmt.Sprintf("\n\n*Showing 20 of %d total records (most recent first).* Use `/approve get <ID>` to view specific requests.", total))
	}

	return output.String()
}

// getStatusIcon returns the icon and label for an approval status
func getStatusIcon(status string) string {
	switch status {
	case "approved":
		return "✅ Approved"
	case "denied":
		return "❌ Denied"
	case "pending":
		return "⏳ Pending"
	case "canceled":
		return "🚫 Canceled"
	default:
		return status // Fallback
	}
}

// executeGet retrieves and displays a specific approval record by code or ID
func (r *Router) executeGet(args *model.CommandArgs) (*model.CommandResponse, error) {
	// Parse command arguments
	split := strings.Fields(args.Command)

	// Check if code/ID parameter is provided (AC6: usage help)
	if len(split) < 3 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Usage: /approve get <APPROVAL_ID>\n\nExample: /approve get A-X7K9Q2",
		}, nil
	}

	code := split[2]

	// Retrieve the approval record
	record, err := r.store.GetApprovalByCode(code)
	if err != nil {
		// Check if error is ErrRecordNotFound (AC5: not found error)
		if errors.Is(err, approval.ErrRecordNotFound) {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("❌ Approval record '%s' not found.\n\nUse `/approve list` to see your approval records.", code),
			}, nil
		}

		// Generic error handling
		r.api.LogError("Failed to retrieve approval record",
			"code", code,
			"user_id", args.UserId,
			"error", err.Error(),
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Failed to retrieve approval record. Please try again.",
		}, nil
	}

	// Access control check (AC4, AC7, AC8: verify user is requester or approver)
	// Security: Only show records where authenticated user (args.UserId) is requester or approver (NFR-S2, FR37)
	if record.RequesterID != args.UserId && record.ApproverID != args.UserId {
		r.api.LogWarn("Unauthorized approval access attempt",
			"user_id", args.UserId,
			"record_id", record.ID,
		)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "❌ Permission denied. You can only view approval records where you are the requester or approver.",
		}, nil
	}

	// Format and return complete record details (AC3: display complete record)
	responseText := formatRecordDetail(record)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         responseText,
	}, nil
}

// formatRecordDetail formats a complete approval record for display
func formatRecordDetail(record *approval.ApprovalRecord) string {
	var output strings.Builder

	// Header with emoji and code (AC3)
	output.WriteString(fmt.Sprintf("**📋 Approval Record: %s**\n\n", record.Code))

	// Status with icon (AC3)
	statusIcon := getStatusIcon(record.Status)
	output.WriteString(fmt.Sprintf("**Status:** %s\n\n", statusIcon))

	// Requester and Approver information (AC3)
	output.WriteString(fmt.Sprintf("**Requester:** @%s (%s)\n", record.RequesterUsername, record.RequesterDisplayName))
	output.WriteString(fmt.Sprintf("**Approver:** @%s (%s)\n\n", record.ApproverUsername, record.ApproverDisplayName))

	// Description (AC3)
	output.WriteString(fmt.Sprintf("**Description:**\n%s\n\n", record.Description))

	// Timestamps (AC3)
	// Format: YYYY-MM-DD HH:MM:SS UTC
	createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
	formattedCreated := createdTime.UTC().Format("2006-01-02 15:04:05 MST")
	output.WriteString(fmt.Sprintf("**Requested:** %s\n", formattedCreated))

	// Decided timestamp (only if decided)
	if record.DecidedAt > 0 {
		decidedTime := time.Unix(0, record.DecidedAt*int64(time.Millisecond))
		formattedDecided := decidedTime.UTC().Format("2006-01-02 15:04:05 MST")
		output.WriteString(fmt.Sprintf("**Decided:** %s\n", formattedDecided))
	} else {
		output.WriteString("**Decided:** Not yet decided\n")
	}

	// Decision comment (only if present) (AC3)
	if record.DecisionComment != "" {
		output.WriteString(fmt.Sprintf("\n**Decision Comment:**\n%s\n", record.DecisionComment))
	}

	// Context section (AC3)
	output.WriteString("\n**Context:**\n")
	output.WriteString(fmt.Sprintf("- Request ID: %s\n", record.Code))
	output.WriteString(fmt.Sprintf("- Full ID: %s\n", record.ID))
	if record.RequestChannelID != "" {
		output.WriteString(fmt.Sprintf("- Channel ID: %s\n", record.RequestChannelID))
	}
	if record.TeamID != "" {
		output.WriteString(fmt.Sprintf("- Team ID: %s\n", record.TeamID))
	}

	// Footer: immutability statement (AC3)
	output.WriteString("\n**This record is immutable and cannot be edited.**\n")

	return output.String()
}
