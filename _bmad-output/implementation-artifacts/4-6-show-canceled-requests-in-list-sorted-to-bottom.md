# Story 4.6: Show Canceled Requests in List (Sorted to Bottom)

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Status:** To Do
**Priority:** Medium
**Estimate:** 3 points
**Assignee:** TBD

## User Story

**As a** user
**I want** to see cancelled requests in my approval list
**So that** I have visibility into the full history without losing context

## Context

Currently `/approve list` shows all approval requests in reverse chronological order (newest first). With the introduction of cancelled status, we need to:
1. Continue showing cancelled requests (audit trail visibility)
2. Sort them to the bottom (de-prioritize non-actionable items)
3. Clearly indicate cancelled status with emoji and reason

This provides symmetry between requester and approver views, and maintains audit trail while keeping actionable items prominent.

## Acceptance Criteria

- [ ] `/approve list` includes cancelled requests for both requesters and approvers
- [ ] Cancelled requests sorted to bottom of list
- [ ] Clear status indicator (🚫 Canceled)
- [ ] Pending requests appear first, then Decided (approved/denied), then Cancelled
- [ ] Sort order within each status group: newest first
- [ ] Cancellation reason shown in list (abbreviated if needed)
- [ ] List remains readable and scannable

## Technical Implementation

### Files to Modify
- `server/command/list.go` - Update sorting and display logic

### Sorting Logic

```go
// In server/command/list.go

type ApprovalRequestGroup struct {
    Pending   []ApprovalRequest
    Decided   []ApprovalRequest // approved or denied
    Canceled  []ApprovalRequest
}

func (c *ListCommand) groupAndSortRequests(requests []ApprovalRequest) ApprovalRequestGroup {
    group := ApprovalRequestGroup{
        Pending:  make([]ApprovalRequest, 0),
        Decided:  make([]ApprovalRequest, 0),
        Canceled: make([]ApprovalRequest, 0),
    }

    // Group by status
    for _, req := range requests {
        switch req.Status {
        case "pending":
            group.Pending = append(group.Pending, req)
        case "approved", "denied":
            group.Decided = append(group.Decided, req)
        case "canceled":
            group.Canceled = append(group.Canceled, req)
        }
    }

    // Sort each group by timestamp (newest first)
    sort.Slice(group.Pending, func(i, j int) bool {
        return group.Pending[i].CreatedAt > group.Pending[j].CreatedAt
    })
    sort.Slice(group.Decided, func(i, j int) bool {
        return group.Decided[i].UpdatedAt > group.Decided[j].UpdatedAt
    })
    sort.Slice(group.Canceled, func(i, j int) bool {
        return group.Canceled[i].CanceledAt > group.Canceled[j].CanceledAt
    })

    return group
}
```

### Display Logic

```go
func (c *ListCommand) formatGroupedList(group ApprovalRequestGroup, userID string) string {
    var builder strings.Builder

    builder.WriteString("## Your Approval Requests\n\n")

    // Pending section
    if len(group.Pending) > 0 {
        builder.WriteString("### Pending:\n")
        for _, req := range group.Pending {
            builder.WriteString(c.formatPendingRequest(req, userID))
        }
        builder.WriteString("\n")
    }

    // Decided section
    if len(group.Decided) > 0 {
        builder.WriteString("### Decided:\n")
        for _, req := range group.Decided {
            builder.WriteString(c.formatDecidedRequest(req))
        }
        builder.WriteString("\n")
    }

    // Canceled section
    if len(group.Canceled) > 0 {
        builder.WriteString("### Canceled:\n")
        for _, req := range group.Canceled {
            builder.WriteString(c.formatCanceledRequest(req))
        }
    }

    if len(group.Pending) == 0 && len(group.Decided) == 0 && len(group.Canceled) == 0 {
        builder.WriteString("_No approval requests found._\n")
    }

    return builder.String()
}

func (c *ListCommand) formatCanceledRequest(req ApprovalRequest) string {
    timestamp := formatTimestamp(req.CanceledAt)

    // Abbreviate reason if too long
    reason := req.CanceledReason
    if reason == "" {
        reason = "No reason"
    } else if len(reason) > 40 {
        reason = reason[:37] + "..."
    }

    return fmt.Sprintf("🚫 **%s** - Canceled (%s) - %s\n",
        req.ReferenceCode,
        reason,
        timestamp,
    )
}

func (c *ListCommand) formatPendingRequest(req ApprovalRequest, userID string) string {
    // Determine perspective
    if req.RequesterID == userID {
        return fmt.Sprintf("🕐 **%s** - Awaiting approval from @%s - %s\n",
            req.ReferenceCode,
            req.ApproverUsername,
            formatTimestamp(req.CreatedAt),
        )
    } else {
        return fmt.Sprintf("🕐 **%s** - Awaiting your approval from @%s - %s\n",
            req.ReferenceCode,
            req.RequesterUsername,
            formatTimestamp(req.CreatedAt),
        )
    }
}

func (c *ListCommand) formatDecidedRequest(req ApprovalRequest) string {
    var statusEmoji, statusText string
    var timestamp int64

    if req.Status == "approved" {
        statusEmoji = "✅"
        statusText = fmt.Sprintf("Approved by @%s", req.ApproverUsername)
        timestamp = req.ApprovedAt
    } else {
        statusEmoji = "❌"
        statusText = fmt.Sprintf("Denied by @%s", req.ApproverUsername)
        timestamp = req.DeniedAt
    }

    return fmt.Sprintf("%s **%s** - %s - %s\n",
        statusEmoji,
        req.ReferenceCode,
        statusText,
        formatTimestamp(timestamp),
    )
}
```

## Testing Requirements

### Unit Tests
- [ ] Test grouping requests by status
- [ ] Test sorting within each group (newest first)
- [ ] Test with mix of all status types
- [ ] Test with only cancelled requests
- [ ] Test with empty list
- [ ] Test reason abbreviation (>40 chars)

### Integration Tests
- [ ] Create multiple requests with different statuses
- [ ] Verify they appear in correct groups
- [ ] Verify sort order is correct
- [ ] Cancel request → Verify it moves to Canceled section
- [ ] Test from requester and approver perspectives

### Manual Testing
- [ ] Create 3 requests: one pending, one approved, one cancelled
- [ ] Run `/approve list` as requester
- [ ] Verify sections appear in order: Pending, Decided, Canceled
- [ ] Verify cancelled requests show reason
- [ ] Run `/approve list` as approver
- [ ] Verify same grouping behavior

## UI/UX Example

### Requester View

```
## Your Approval Requests

### Pending:
🕐 **TUZ-3AB** - Awaiting approval from @alice - Jan 12 8:00 PM
🕐 **TUZ-4CD** - Awaiting approval from @bob - Jan 12 6:30 PM

### Decided:
✅ **TUZ-2EF** - Approved by @bob - Jan 12 5:30 PM
❌ **TUZ-1GH** - Denied by @charlie - Jan 11 2:00 PM

### Canceled:
🚫 **TUZ-2RK** - Canceled (No longer needed) - Jan 12 7:15 PM
🚫 **TUZ-5IJ** - Canceled (Sensitive information) - Jan 11 3:45 PM
```

### Approver View (same formatting, different perspective)

```
## Your Approval Requests

### Pending:
🕐 **TUZ-3AB** - Awaiting your approval from @wayne - Jan 12 8:00 PM

### Decided:
✅ **TUZ-2EF** - Approved by @bob - Jan 12 5:30 PM
❌ **TUZ-1GH** - Denied by @charlie - Jan 11 2:00 PM

### Canceled:
🚫 **TUZ-2RK** - Canceled (No longer needed) - Jan 12 7:15 PM
```

### Long Reason Example

```
### Canceled:
🚫 **ABC-1XY** - Canceled (Other: CFO already approved this ...) - Jan 11 10:30 AM
```

## Dependencies

- **Depends on:** Story 4.4 (Store reason) - needs CanceledReason field
- **Related to:** Story 4.5 (Display details) - consistent formatting
- **Blocks:** None

## Edge Cases

1. **All requests are cancelled**
   - Only show "Canceled:" section
   - No empty sections displayed

2. **No cancelled requests**
   - Don't show "Canceled:" section header
   - Only show Pending and/or Decided

3. **Reason is empty (old record)**
   - Display: "No reason"
   - Still sorted to Canceled section

4. **Very long custom reason**
   - Truncate to 40 chars + "..."
   - User can run `/approve get [ID]` for full details

5. **CanceledAt is 0 (old record)**
   - Sort to end of Canceled section
   - Display: "Unknown" timestamp

## Performance Considerations

- Grouping and sorting happens in-memory (not additional KV queries)
- Max expected items: ~100-200 per user
- Performance impact: negligible (sorting 200 items is instant)

## Future Enhancements (v0.3.0)

- Add filter dropdown: "Show: All | Pending | Decided | Canceled"
- Add pagination if list grows very long
- Add "Hide canceled" toggle for cleaner view

## Definition of Done

- [ ] Code implemented and reviewed
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] Manual testing in real Mattermost
- [ ] All status types display correctly
- [ ] Sorting is correct and consistent
- [ ] Code merged to master

## Notes

**Why group by status instead of pure chronological:**
- Actionability: Users need to see pending items first
- Mental model: "What do I need to act on?" vs "What happened chronologically?"
- Scannability: Grouped sections are easier to parse than mixed list

**Why show cancelled to both requester and approver:**
- Symmetry: Both roles see same history
- Audit trail: Approvers can reference "oh yeah, you cancelled that yesterday"
- Less surprising: Item doesn't just disappear from approver's view

**Why sort cancelled by CanceledAt not CreatedAt:**
- Most recent cancellation is most relevant
- "When did this become cancelled?" is more useful than "when was this created?"
- Consistent with Decided section (sorted by decision time)

**Why abbreviate reason in list:**
- List must be scannable (not wall of text)
- 40 chars is enough to convey gist
- Detail view (`/approve get`) has full text
