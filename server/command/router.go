package command

import (
	"errors"
	"fmt"
	"slices"
	"sort"
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
* **/approve list [filter]** - View your approval requests and decisions
  * No filter: shows pending requests (default)
  * **pending** - pending approval requests
  * **approved** - approved requests
  * **denied** - denied requests
  * **canceled** - canceled requests
  * **all** - all requests (pending, approved, denied, canceled)
* **/approve get [ID]** - View a specific approval by ID
* **/approve cancel <APPROVAL_ID>** - Cancel a pending approval request
* **/approve verify <APPROVAL_CODE> [comment]** - Mark an approved request as verified/complete
* **/approve status** - View approval system statistics (admin only)
* **/approve help** - Display this help text

**Admin Commands:**

* **/approve status** - View overall approval statistics and notification health
* **/approve status --failed-notifications** - List specific approvals with failed notifications

**Examples:**
` + "`/approve new`" + ` - Opens a modal to create an approval request
` + "`/approve list`" + ` - Shows pending approval requests
` + "`/approve list approved`" + ` - Shows approved requests
` + "`/approve list all`" + ` - Shows all requests

For more information, visit the plugin documentation.`

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         helpText,
	}
}

// executeUnknown returns error for unrecognized commands
func executeUnknown(subcommand string) *model.CommandResponse {
	errorText := fmt.Sprintf("Unknown command: **%s**\n\nValid commands: `new`, `list`, `get`, `cancel`, `verify`, `status`, `help`\n\nType `/approve help` for more information.", subcommand)

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
	if !slices.Contains(roles, "system_admin") {
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
	showFailedOnly := slices.Contains(subargs, "--failed-notifications")

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
	// Parse filter parameter from command (Story 5.1)
	// Example: "/approve list pending" -> filter = "pending"
	// Example: "/approve list" -> filter = "pending" (default, Story 5.2: changed to focus on actionable items)
	parts := strings.Fields(args.Command)
	filter := "pending" // Story 5.2: Changed default to focus on actionable items (was "all" in Story 5.1)

	if len(parts) > 2 {
		// Extract filter argument (parts[2] is the filter after "/approve list")
		filter = strings.ToLower(strings.TrimSpace(parts[2]))

		// Validate filter
		validFilters := map[string]bool{
			"pending":  true,
			"approved": true,
			"denied":   true,
			"canceled": true,
			"all":      true,
		}

		if !validFilters[filter] {
			errorMsg := fmt.Sprintf("Invalid filter '%s'. Valid filters: pending, approved, denied, canceled, all", parts[2])
			post := &model.Post{
				UserId:    args.UserId,
				ChannelId: args.ChannelId,
				Message:   errorMsg,
			}
			ephemeralPost := r.api.SendEphemeralPost(args.UserId, post)
			if ephemeralPost == nil {
				// Fallback to CommandResponse if ephemeral post fails
				return &model.CommandResponse{
					ResponseType: model.CommandResponseTypeEphemeral,
					Text:         errorMsg,
				}, nil
			}
			return &model.CommandResponse{}, nil
		}
	}

	// Security: args.UserId is authenticated by Mattermost Plugin API
	// The Mattermost server guarantees this ID matches the authenticated user session (NFR-S1)
	// Access control: GetUserApprovals only returns records where this user is requester or approver (NFR-S2, FR37)
	records, err := r.store.GetUserApprovals(args.UserId)
	if err != nil {
		r.api.LogError("Failed to retrieve approval records for list command",
			"user_id", args.UserId,
			"error", err.Error(),
		)
		errorMsg := "❌ Failed to retrieve approval records. Please try again."
		post := &model.Post{
			UserId:    args.UserId,
			ChannelId: args.ChannelId,
			Message:   errorMsg,
		}
		ephemeralPost := r.api.SendEphemeralPost(args.UserId, post)
		if ephemeralPost == nil {
			// Fallback to CommandResponse if ephemeral post fails
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         errorMsg,
			}, nil
		}
		return &model.CommandResponse{}, nil
	}

	// Apply status filter (Story 5.1)
	filteredRecords := filterRecordsByStatus(records, filter)

	// Handle empty state (Story 5.2: Filter-specific message)
	if len(filteredRecords) == 0 {
		emptyMessage := fmt.Sprintf("No %s approval requests. Use `/approve list all` to see all requests.", filter)
		post := &model.Post{
			UserId:    args.UserId,
			ChannelId: args.ChannelId,
			Message:   emptyMessage,
		}
		ephemeralPost := r.api.SendEphemeralPost(args.UserId, post)
		if ephemeralPost == nil {
			// Fallback to CommandResponse if ephemeral post fails
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         emptyMessage,
			}, nil
		}
		return &model.CommandResponse{}, nil
	}

	// Apply chronological sorting for specific status filters (Story 5.1, Subtask 2.5)
	// For "all" filter, groupAndSortRecords will handle sorting in formatListResponse
	sortRecordsByTimestamp(filteredRecords, filter)

	// Apply pagination (limit to 20 records)
	total := len(filteredRecords)
	displayRecords := filteredRecords
	if total > 20 {
		displayRecords = filteredRecords[:20]
	}

	// Format response (Story 5.2: Pass filter for dynamic header)
	responseText := formatListResponse(displayRecords, total, filter)

	// Story 7.5: Send as ephemeral post instead of CommandResponse to enable markdown table rendering
	// CommandResponse.Text doesn't properly render markdown tables in Mattermost
	post := &model.Post{
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   responseText,
	}

	ephemeralPost := r.api.SendEphemeralPost(args.UserId, post)
	if ephemeralPost == nil {
		r.api.LogError("Failed to send ephemeral list response", "user_id", args.UserId)
		// Fallback to CommandResponse if ephemeral post fails
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         responseText,
		}, nil
	}

	// Return empty response since we already sent the ephemeral post
	return &model.CommandResponse{}, nil
}

// filterRecordsByStatus filters approval records by status type (Story 5.1)
// Supported filters: "pending", "approved", "denied", "canceled", "all"
// Returns filtered slice or original slice if filter is "all"
func filterRecordsByStatus(records []*approval.ApprovalRecord, filter string) []*approval.ApprovalRecord {
	// "all" filter returns all records unchanged
	if filter == "all" {
		return records
	}

	// Filter records by matching status
	// Initialize as empty slice (not nil) for consistent behavior
	filtered := make([]*approval.ApprovalRecord, 0)
	for _, record := range records {
		switch filter {
		case "pending":
			if record.Status == approval.StatusPending {
				filtered = append(filtered, record)
			}
		case "approved":
			if record.Status == approval.StatusApproved {
				filtered = append(filtered, record)
			}
		case "denied":
			if record.Status == approval.StatusDenied {
				filtered = append(filtered, record)
			}
		case "canceled":
			if record.Status == approval.StatusCanceled {
				filtered = append(filtered, record)
			}
		}
	}

	return filtered
}

// sortRecordsByTimestamp sorts approval records chronologically by appropriate timestamp
// for specific status filters (Story 5.1, Subtask 2.5).
// For "all" filter, this is a no-op (groupAndSortRecords handles it in formatListResponse).
// Sorts in-place, descending (newest first).
func sortRecordsByTimestamp(records []*approval.ApprovalRecord, filter string) {
	// No-op for "all" filter (groupAndSortRecords handles this)
	if filter == "all" {
		return
	}

	// Sort by appropriate timestamp based on filter
	sort.Slice(records, func(i, j int) bool {
		var iTime, jTime int64

		switch filter {
		case "pending":
			// Pending records: sort by CreatedAt
			iTime = records[i].CreatedAt
			jTime = records[j].CreatedAt

		case "approved", "denied":
			// Decided records: sort by DecidedAt
			iTime = records[i].DecidedAt
			jTime = records[j].DecidedAt

		case "canceled":
			// Canceled records: sort by CanceledAt (fallback to CreatedAt if not set)
			iTime = records[i].CanceledAt
			if iTime == 0 {
				iTime = records[i].CreatedAt
			}
			jTime = records[j].CanceledAt
			if jTime == 0 {
				jTime = records[j].CreatedAt
			}

		default:
			// Unknown filter: no sorting
			return false
		}

		// Sort descending (newest first)
		return iTime > jTime
	})
}

// groupAndSortRecords separates records into three groups (pending, decided, canceled)
// and sorts each group by appropriate timestamp descending (newest first)
func groupAndSortRecords(records []*approval.ApprovalRecord) (pending, decided, canceled []*approval.ApprovalRecord) {
	// Separate into groups
	for _, record := range records {
		switch record.Status {
		case approval.StatusPending:
			pending = append(pending, record)
		case approval.StatusApproved, approval.StatusDenied:
			decided = append(decided, record)
		case approval.StatusCanceled:
			canceled = append(canceled, record)
		}
	}

	// Sort pending by CreatedAt descending
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].CreatedAt > pending[j].CreatedAt
	})

	// Sort decided by DecidedAt descending
	sort.Slice(decided, func(i, j int) bool {
		return decided[i].DecidedAt > decided[j].DecidedAt
	})

	// Sort canceled by CanceledAt descending (fallback to CreatedAt if CanceledAt == 0)
	sort.Slice(canceled, func(i, j int) bool {
		iTime := canceled[i].CanceledAt
		if iTime == 0 {
			iTime = canceled[i].CreatedAt
		}
		jTime := canceled[j].CanceledAt
		if jTime == 0 {
			jTime = canceled[j].CreatedAt
		}
		return iTime > jTime
	})

	return pending, decided, canceled
}

// formatListResponse formats approval records into a readable list with grouped sections
// Story 5.2: Added filter parameter for dynamic header count
func formatListResponse(records []*approval.ApprovalRecord, total int, filter string) string {
	var output strings.Builder

	// Story 5.2: Dynamic header with count (AC2, AC3)
	header := fmt.Sprintf("## Your Approval Requests (%d %s)\n\n", total, filter)
	output.WriteString(header)

	// Group and sort records
	pending, decided, canceled := groupAndSortRecords(records)

	// Limit to 20 total records across all groups
	displayed := 0
	limit := 20

	// Render Pending section
	if len(pending) > 0 {
		output.WriteString("**Pending Approvals:**\n\n")
		output.WriteString("| Code | Status | Requestor | Approver | Created |\n")
		output.WriteString("|------|--------|-----------|----------|----------|\n")
		for _, record := range pending {
			if displayed >= limit {
				break
			}
			statusIcon := getStatusIcon(record.Status)
			createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
			formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
			output.WriteString(fmt.Sprintf("| %s | %s | @%s | @%s | %s |\n",
				record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
			displayed++
		}
		output.WriteString("\n")
	}

	// Render Decided section
	if len(decided) > 0 && displayed < limit {
		output.WriteString("**Decided Approvals:**\n\n")
		output.WriteString("| Code | Status | Requestor | Approver | Created |\n")
		output.WriteString("|------|--------|-----------|----------|----------|\n")
		for _, record := range decided {
			if displayed >= limit {
				break
			}
			statusIcon := getStatusIcon(record.Status)
			createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
			formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
			output.WriteString(fmt.Sprintf("| %s | %s | @%s | @%s | %s |\n",
				record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
			displayed++
		}
		output.WriteString("\n")
	}

	// Render Canceled section (with reason)
	if len(canceled) > 0 && displayed < limit {
		output.WriteString("**Canceled Requests:**\n\n")
		output.WriteString("| Code | Status | Requestor | Approver | Created |\n")
		output.WriteString("|------|--------|-----------|----------|----------|\n")
		for _, record := range canceled {
			if displayed >= limit {
				break
			}

			// Format canceled status with reason
			var statusText string
			if record.CanceledReason != "" {
				reason := record.CanceledReason
				// Truncate if longer than 40 characters (use rune count for proper UTF-8 handling)
				runes := []rune(reason)
				if len(runes) > 40 {
					reason = string(runes[:37]) + "..."
				}
				statusText = fmt.Sprintf("🚫 Canceled (%s)", reason)
			} else {
				statusText = "🚫 Canceled"
			}

			createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
			formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
			output.WriteString(fmt.Sprintf("| %s | %s | @%s | @%s | %s |\n",
				record.Code, statusText, record.RequesterUsername, record.ApproverUsername, formattedDate))
			displayed++
		}
		output.WriteString("\n")
	}

	// Pagination footer (if total > displayed)
	if total > displayed {
		output.WriteString(fmt.Sprintf("*Showing %d of %d total records.* Use `/approve get <ID>` to view specific requests.", displayed, total))
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

	// Cancellation details (Story 7.3: display reason, details, and timestamp)
	if record.Status == approval.StatusCanceled {
		output.WriteString("---\n\n")
		output.WriteString("**Cancellation:**\n")

		// Cancellation reason (handle empty for old records)
		reason := record.CanceledReason
		if reason == "" {
			reason = "No reason recorded (canceled before v0.2.0)"
		}
		output.WriteString(fmt.Sprintf("**Reason:** %s\n", reason))

		// Additional details (Story 7.3: display if present)
		if record.CanceledDetails != "" {
			output.WriteString(fmt.Sprintf("**Details:** %s\n", record.CanceledDetails))
		}

		// Who canceled (only requester can cancel in v0.2.0)
		output.WriteString(fmt.Sprintf("**Canceled by:** @%s\n", record.RequesterUsername))

		// Cancellation timestamp (handle zero value for old records)
		if record.CanceledAt > 0 {
			canceledTime := time.Unix(0, record.CanceledAt*int64(time.Millisecond))
			formattedCanceled := canceledTime.UTC().Format("2006-01-02 15:04:05 MST")
			output.WriteString(fmt.Sprintf("**Canceled:** %s\n", formattedCanceled))
		} else {
			output.WriteString("**Canceled:** Unknown\n")
		}

		output.WriteString("\n") // Spacing before next section
	}

	// Timestamps (AC3)
	// Format: YYYY-MM-DD HH:MM:SS UTC
	createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
	formattedCreated := createdTime.UTC().Format("2006-01-02 15:04:05 MST")
	output.WriteString(fmt.Sprintf("**Requested:** %s\n", formattedCreated))

	// Decided timestamp (only if decided and not canceled)
	if record.Status != approval.StatusCanceled {
		if record.DecidedAt > 0 {
			decidedTime := time.Unix(0, record.DecidedAt*int64(time.Millisecond))
			formattedDecided := decidedTime.UTC().Format("2006-01-02 15:04:05 MST")
			output.WriteString(fmt.Sprintf("**Decided:** %s\n", formattedDecided))
		} else {
			output.WriteString("**Decided:** Not yet decided\n")
		}
	}

	// Decision comment (only if present) (AC3)
	if record.DecisionComment != "" {
		output.WriteString(fmt.Sprintf("\n**Decision Comment:**\n%s\n", record.DecisionComment))
	}

	// Verification details (Story 6.2: display verification info for verified approved requests)
	if record.Status == approval.StatusApproved && record.Verified {
		output.WriteString("\n---\n\n")
		output.WriteString("**✅ Verification:**\n")

		// Verification timestamp
		if record.VerifiedAt > 0 {
			verifiedTime := time.Unix(0, record.VerifiedAt*int64(time.Millisecond))
			formattedVerified := verifiedTime.UTC().Format("2006-01-02 15:04:05 MST")
			output.WriteString(fmt.Sprintf("**Verified:** %s\n", formattedVerified))
		}

		// Verified by (always the requester in Story 6.2)
		output.WriteString(fmt.Sprintf("**Verified by:** @%s\n", record.RequesterUsername))

		// Verification comment (optional)
		if record.VerificationComment != "" {
			output.WriteString(fmt.Sprintf("**Note:** %s\n", record.VerificationComment))
		}

		output.WriteString("\n") // Spacing before next section
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
