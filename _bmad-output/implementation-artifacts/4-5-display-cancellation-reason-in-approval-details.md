# Story 4.5: Display Cancellation Reason in Approval Details

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Status:** To Do
**Priority:** Medium
**Estimate:** 2 points
**Assignee:** TBD

## User Story

**As a** user (requester or approver)
**I want** to see why a request was cancelled when viewing details
**So that** I have full context of what happened

## Context

When users run `/approve get [ID]` to view a specific approval, they should see cancellation information if the request was cancelled. This completes the audit trail by making cancellation reasons visible and queryable.

## Acceptance Criteria

- [ ] `/approve get [ID]` shows cancellation info for cancelled requests
- [ ] Display includes: reason, who cancelled, when cancelled
- [ ] Formatted clearly and consistently with other statuses
- [ ] Works for both requester and approver viewing the record
- [ ] Handles old records without cancellation reason gracefully ("No reason recorded")
- [ ] Status emoji shows 🚫 for cancelled requests

## Technical Implementation

### Files to Modify
- `server/command/get.go` - Update display formatting to include cancellation details

### Display Logic

```go
// In server/command/get.go

func (c *GetCommand) formatApprovalRequest(request *ApprovalRequest) string {
    var builder strings.Builder

    // Status header with appropriate emoji
    statusEmoji := getStatusEmoji(request.Status)
    builder.WriteString(fmt.Sprintf("## %s Approval Request (%s)\n\n", statusEmoji, strings.Title(request.Status)))

    // Basic info
    builder.WriteString(fmt.Sprintf("**Reference Code:** %s\n", request.ReferenceCode))
    builder.WriteString(fmt.Sprintf("**Status:** %s\n", strings.Title(request.Status)))
    builder.WriteString(fmt.Sprintf("**Requester:** @%s\n", request.RequesterUsername))
    builder.WriteString(fmt.Sprintf("**Approver:** @%s\n", request.ApproverUsername))
    builder.WriteString(fmt.Sprintf("**Created:** %s\n\n", formatTimestamp(request.CreatedAt)))

    // Description
    builder.WriteString("**Description:**\n")
    builder.WriteString(fmt.Sprintf("%s\n\n", request.Description))

    // Status-specific details
    switch request.Status {
    case "approved":
        builder.WriteString(fmt.Sprintf("**Decision:** Approved by @%s\n", request.ApproverUsername))
        builder.WriteString(fmt.Sprintf("**Decided:** %s\n", formatTimestamp(request.ApprovedAt)))

    case "denied":
        builder.WriteString(fmt.Sprintf("**Decision:** Denied by @%s\n", request.ApproverUsername))
        builder.WriteString(fmt.Sprintf("**Decided:** %s\n", formatTimestamp(request.DeniedAt)))

    case "canceled":
        builder.WriteString("---\n\n")
        builder.WriteString("**Cancellation:**\n")

        // Cancellation reason
        reason := request.CanceledReason
        if reason == "" {
            reason = "No reason recorded (cancelled before v0.2.0)"
        }
        builder.WriteString(fmt.Sprintf("**Reason:** %s\n", reason))

        // Who cancelled (always the requester in v0.2.0)
        builder.WriteString(fmt.Sprintf("**Canceled by:** @%s\n", request.RequesterUsername))

        // When cancelled
        if request.CanceledAt > 0 {
            builder.WriteString(fmt.Sprintf("**Canceled:** %s\n", formatTimestamp(request.CanceledAt)))
        }
    }

    return builder.String()
}

func getStatusEmoji(status string) string {
    switch status {
    case "pending":
        return "🕐"
    case "approved":
        return "✅"
    case "denied":
        return "❌"
    case "canceled":
        return "🚫"
    default:
        return "❓"
    }
}

func formatTimestamp(millis int64) string {
    if millis == 0 {
        return "Unknown"
    }
    t := time.Unix(millis/1000, 0)
    return t.Format("Jan 02, 2006 3:04 PM")
}
```

## Testing Requirements

### Unit Tests
- [ ] Test formatting for cancelled request with reason
- [ ] Test formatting for cancelled request without reason (old record)
- [ ] Test status emoji selection
- [ ] Test timestamp formatting
- [ ] Test with long cancellation reasons (truncation if needed)

### Integration Tests
- [ ] Create request → Cancel with reason → Get → Verify display
- [ ] Test each cancellation reason displays correctly
- [ ] Test "Other" reason with custom text displays
- [ ] Test old cancelled records (no reason) display gracefully

### Manual Testing
- [ ] Cancel request with each enumerated reason
- [ ] Run `/approve get [ID]` for each
- [ ] Verify formatting is clear and readable
- [ ] Verify emoji appears correctly
- [ ] Test with mobile Mattermost client

## UI/UX Examples

### Cancelled Request (with reason)

```
🚫 Approval Request (Canceled)

Reference Code: TUZ-2RK
Status: Canceled
Requester: @wayne
Approver: @john
Created: Jan 12, 2026 3:45 PM

Description:
Deploy hotfix to production database

---

Cancellation:
Reason: No longer needed
Canceled by: @wayne
Canceled: Jan 12, 2026 7:15 PM
```

### Cancelled Request ("Other" with custom text)

```
🚫 Approval Request (Canceled)

Reference Code: ABC-1XY
Status: Canceled
Requester: @alice
Approver: @bob
Created: Jan 11, 2026 9:00 AM

Description:
Approve budget increase for Q1 marketing

---

Cancellation:
Reason: Other: CFO already approved this via email, don't need plugin approval
Canceled by: @alice
Canceled: Jan 11, 2026 10:30 AM
```

### Old Cancelled Request (no reason)

```
🚫 Approval Request (Canceled)

Reference Code: OLD-2AB
Status: Canceled
Requester: @charlie
Approver: @dana
Created: Jan 5, 2026 2:00 PM

Description:
Emergency deployment approval

---

Cancellation:
Reason: No reason recorded (cancelled before v0.2.0)
Canceled by: @charlie
Canceled: Unknown
```

## Dependencies

- **Depends on:** Story 4.4 (Store reason) - needs data to display
- **Blocks:** None
- **Related to:** Story 4.6 (List display) - consistent status formatting

## Edge Cases

1. **Cancelled before v0.2.0**
   - CanceledReason is empty
   - CanceledAt might be 0
   - Display: "No reason recorded (cancelled before v0.2.0)"

2. **Very long custom reason**
   - "Other:" text could be 500 chars
   - Display: Show full text (it's in a modal view, not list)
   - Could add "..." truncation if needed in future

3. **Cancelled by deleted user**
   - Username might no longer exist
   - Display username from stored RequesterUsername field

4. **Timestamp is zero**
   - Old records might not have CanceledAt
   - Display: "Unknown"

## Definition of Done

- [ ] Code implemented and reviewed
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] Manual testing in real Mattermost
- [ ] All edge cases handled gracefully
- [ ] Formatting is consistent with existing displays
- [ ] Code merged to master

## Notes

**Why show "Canceled by @requester":**
- Even though only requesters can cancel in v0.2.0, it's explicit and clear
- Future versions might allow admins to cancel
- Audit trail clarity: "who took this action?"

**Why not truncate long custom reasons:**
- This is a detail view, not a list
- User specifically asked for full details (`/approve get`)
- Full context is more valuable than brevity here
- If it becomes a problem, can add truncation in v0.3.0

**Why different display from list:**
- List needs to be scannable (one line per item)
- Detail view can be verbose (full context)
- Different use cases: browsing vs investigating
