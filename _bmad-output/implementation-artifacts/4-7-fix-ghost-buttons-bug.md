# Story 4.7: Fix Ghost Buttons Bug

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Status:** To Do
**Priority:** High
**Estimate:** 2 points
**Assignee:** TBD

## User Story

**As an** approver
**I want** non-functional buttons to be removed
**So that** the interface doesn't lie to me

## Context

**Current Bug (Issue #1):**
When a request is cancelled or decided (approved/denied), the approve/deny buttons remain visible in the approver's DM but no longer work. Clicking them does nothing, creating a confusing and frustrating user experience.

This is a fundamental trust issue: the UI presents interactive elements that don't function, making users question whether their actions are being processed or if the plugin is broken.

## Acceptance Criteria

- [ ] When approval decision is made (approve/deny), buttons removed from post
- [ ] When request is cancelled, buttons removed from post
- [ ] No clickable buttons remain for completed/cancelled requests
- [ ] If button removal fails, log error but don't fail the operation
- [ ] Button removal happens immediately (not delayed)
- [ ] Works for both cancellation and decision workflows

## Technical Implementation

### Root Cause

Mattermost posts with interactive buttons use `Props` field with action attachments. When we update a post's message but don't clear the Props, the buttons remain even though the backend no longer handles those actions.

### Files to Modify
- `server/notifications.go` - Update post functions to clear Props
- `server/command/approve.go` - Clear buttons after approval decision
- `server/command/deny.go` - Clear buttons after deny decision
- `server/command/cancel.go` - Clear buttons after cancellation (already covered in 4.1)

### Fix for Cancellation (extends Story 4.1)

```go
// In server/notifications.go - UpdateApprovalPostForCancellation()

func (p *Plugin) UpdateApprovalPostForCancellation(request *ApprovalRequest, canceledBy string) error {
    // ... existing code to get post ...

    // Build updated message
    canceledAt := time.Unix(request.CanceledAt/1000, 0).Format("Jan 02, 2006 3:04 PM")
    updatedMessage := fmt.Sprintf(
        "## 🚫 Approval Request (Canceled)\n\n" +
        "**Reference Code:** %s\n" +
        "**Status:** Canceled\n" +
        "**Requester:** @%s\n\n" +
        "**Description:**\n~~%s~~\n\n" +
        "---\n" +
        "_Canceled by @%s at %s_",
        request.ReferenceCode,
        request.RequesterUsername,
        request.Description,
        canceledBy,
        canceledAt,
    )

    // CRITICAL: Clear props to remove buttons
    post.Message = updatedMessage
    post.Props = map[string]interface{}{} // This removes all interactive elements

    // Update the post
    _, err = p.API.UpdatePost(post)
    if err != nil {
        p.API.LogError("Failed to update post", "post_id", request.ApproverPostID, "error", err.Error())
        return err
    }

    return nil
}
```

### Fix for Approval Decision

```go
// In server/command/approve.go

func (c *ApproveCommand) handleApprovalConfirmation(request *ApprovalRequest) error {
    // ... existing code to record decision ...

    // Update the approver's original post to show approved state
    err := c.plugin.UpdateApprovalPostForDecision(request, "approved")
    if err != nil {
        c.plugin.API.LogError("Failed to update approver post", "error", err.Error())
        // Don't fail - decision is already recorded
    }

    // ... existing code to notify requester ...
}

// In server/notifications.go

func (p *Plugin) UpdateApprovalPostForDecision(request *ApprovalRequest, decision string) error {
    if request.ApproverPostID == "" {
        p.API.LogWarn("Cannot update approver post: no post ID stored", "request_id", request.ID)
        return fmt.Errorf("no approver post ID found")
    }

    post, err := p.API.GetPost(request.ApproverPostID)
    if err != nil {
        p.API.LogError("Failed to get post for update", "post_id", request.ApproverPostID, "error", err.Error())
        return err
    }

    // Build updated message based on decision
    var statusEmoji, statusText string
    var timestamp int64

    if decision == "approved" {
        statusEmoji = "✅"
        statusText = "Approved"
        timestamp = request.ApprovedAt
    } else {
        statusEmoji = "❌"
        statusText = "Denied"
        timestamp = request.DeniedAt
    }

    decidedAt := time.Unix(timestamp/1000, 0).Format("Jan 02, 2006 3:04 PM")

    updatedMessage := fmt.Sprintf(
        "## %s Approval Request (%s)\n\n" +
        "**Reference Code:** %s\n" +
        "**Status:** %s\n" +
        "**Requester:** @%s\n\n" +
        "**Description:**\n%s\n\n" +
        "---\n" +
        "_Decision recorded at %s_",
        statusEmoji,
        statusText,
        request.ReferenceCode,
        statusText,
        request.RequesterUsername,
        request.Description,
        decidedAt,
    )

    // CRITICAL: Clear props to remove buttons
    post.Message = updatedMessage
    post.Props = map[string]interface{}{}

    _, err = p.API.UpdatePost(post)
    if err != nil {
        p.API.LogError("Failed to update post", "post_id", request.ApproverPostID, "error", err.Error())
        return err
    }

    return nil
}
```

### Same fix for Deny

```go
// In server/command/deny.go - mirror approve.go logic

func (c *DenyCommand) handleDenialConfirmation(request *ApprovalRequest) error {
    // ... existing code to record decision ...

    err := c.plugin.UpdateApprovalPostForDecision(request, "denied")
    if err != nil {
        c.plugin.API.LogError("Failed to update approver post", "error", err.Error())
    }

    // ... existing code to notify requester ...
}
```

## Testing Requirements

### Unit Tests
- [ ] Test props are cleared on cancellation
- [ ] Test props are cleared on approval
- [ ] Test props are cleared on denial
- [ ] Test handling when post doesn't exist
- [ ] Test handling when Props is already empty

### Integration Tests
- [ ] Create request → Cancel → Verify buttons gone
- [ ] Create request → Approve → Verify buttons gone
- [ ] Create request → Deny → Verify buttons gone
- [ ] Click button after removal → Verify no action taken

### Manual Testing (Critical)
- [ ] Create approval request in real Mattermost
- [ ] **Before fix:** Click buttons after cancel → nothing happens (ghost buttons)
- [ ] **After fix:** Buttons should be completely gone
- [ ] Test on mobile client as well
- [ ] Test with slow network (ensure update completes)

### Regression Testing
- [ ] Verify approve flow still works
- [ ] Verify deny flow still works
- [ ] Verify cancel flow still works
- [ ] Verify notifications still sent

## Dependencies

- **Related to:** Story 4.1 (Update post on cancel) - same post update logic
- **Blocks:** None
- **Blocked by:** None (can implement independently)

## UI Comparison

### Before Fix (Bug)

```
┌────────────────────────────────────────────┐
│ 🚫 Approval Request (Canceled)             │
│                                            │
│ Reference Code: TUZ-2RK                    │
│ Status: Canceled                           │
│ Requester: @wayne                          │
│                                            │
│ Description:                               │
│ ~~Deploy hotfix to production database~~  │
│                                            │
│ [Approve] [Deny]  ← GHOST BUTTONS (broken) │
│                                            │
│ ─────────────────────────────────────────  │
│ Canceled by @wayne at Jan 12, 2026 7:15 PM│
└────────────────────────────────────────────┘
```

### After Fix

```
┌────────────────────────────────────────────┐
│ 🚫 Approval Request (Canceled)             │
│                                            │
│ Reference Code: TUZ-2RK                    │
│ Status: Canceled                           │
│ Requester: @wayne                          │
│                                            │
│ Description:                               │
│ ~~Deploy hotfix to production database~~  │
│                                            │
│ ─────────────────────────────────────────  │
│ Canceled by @wayne at Jan 12, 2026 7:15 PM│
└────────────────────────────────────────────┘
```

No buttons = honest UI.

## Error Handling

1. **Post doesn't exist** (deleted channel, archived)
   - Log warning
   - Continue (can't update what doesn't exist)
   - Decision/cancellation is already recorded

2. **Permission denied** (bot lost permissions)
   - Log error
   - Continue (decision/cancellation already recorded)
   - Rare edge case

3. **Network timeout**
   - Log error
   - Retry once after 1 second
   - If still fails, continue (don't block user action)

4. **Props is null/undefined**
   - Handle gracefully (set to empty map)
   - Don't crash on edge case

## Performance Impact

- Minimal: One additional API call per decision/cancellation
- API call is async (doesn't block user response)
- Post updates are fast (<100ms typically)

## Definition of Done

- [ ] Code implemented and reviewed
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] Manual testing confirms buttons are removed
- [ ] Tested on desktop and mobile clients
- [ ] GitHub Issue #1 closed
- [ ] Code merged to master

## Notes

**Why this is high priority:**
- It's the original bug report (Issue #1)
- Breaks user trust in the plugin
- Makes plugin feel broken/unreliable
- Simple fix with high user impact

**Why clear Props instead of modifying:**
- Props can contain complex nested data
- Easier to clear everything than selectively remove action buttons
- No side effects (we only put action buttons in Props)
- More maintainable

**Why don't fail if post update fails:**
- Decision/cancellation is already recorded (most important)
- Post update is UX polish, not core functionality
- Failing would block user from completing action
- Logging error is sufficient for debugging

**Testing the fix:**
The key test is: "Can I click buttons after cancellation/decision?"
- Before fix: Yes (but nothing happens) ❌
- After fix: No (buttons are gone) ✅
