# Story 8.5: Update Playbook Channel on Status Changes

**Epic:** 8 - Playbook Integration
**Status:** ready-for-dev
**Priority:** High
**Estimate:** 8 points
**Assignee:** TBD

## User Story

**As a** playbook team member
**I want** to see when approvals are approved, denied, canceled, or timed out
**So that** I know when blockers are resolved

## Context

After an approval request is created in a playbook channel (Story 8.3), the team needs real-time updates when the status changes. This completes the approval lifecycle visibility, ensuring the entire playbook team stays informed without manual status checks.

This story handles four status change scenarios:
1. **Approved** - Approver granted approval
2. **Denied** - Approver rejected request
3. **Canceled** - Requester canceled before decision
4. **Timed Out** - No response within timeout period

## Acceptance Criteria

- [ ] AC1: When approval approved, post "✅ **Approved:** [CODE] | [Details] | Approved by @approver at [time]"
- [ ] AC2: When approval denied, post "❌ **Denied:** [CODE] | [Details] | Denied by @approver | Reason: [reason]"
- [ ] AC3: When approval canceled, post "🚫 **Canceled:** [CODE] | [Details] | Reason: [cancellation reason]"
- [ ] AC4: When approval times out, post "⏱️ **Timeout:** [CODE] | [Details] | No response from @approver"
- [ ] AC5: All status messages include reference code for correlation
- [ ] AC6: Messages use consistent emoji and formatting
- [ ] AC7: Timestamps formatted as human-readable (e.g., "14:23" or "Jan 17, 14:23")
- [ ] AC8: Posting errors logged but don't block status updates
- [ ] AC9: Non-playbook approvals unchanged (no channel posts)

## Tasks / Subtasks

- [ ] Task 1: Implement status message formatting (AC: 1, 2, 3, 4, 5, 6, 7)
  - [ ] Subtask 1.1: Create formatApprovedStatusMessage helper
  - [ ] Subtask 1.2: Create formatDeniedStatusMessage helper
  - [ ] Subtask 1.3: Create formatCanceledStatusMessage helper
  - [ ] Subtask 1.4: Create formatTimedOutStatusMessage helper
  - [ ] Subtask 1.5: Implement consistent emoji/formatting across all messages
  - [ ] Subtask 1.6: Format timestamps using Mattermost date utilities
  - [ ] Subtask 1.7: Truncate long details/reasons to 100 characters
  - [ ] Subtask 1.8: Write unit tests for all formatters

- [ ] Task 2: Hook into approval decision flow (AC: 1, 2, 8, 9)
  - [ ] Subtask 2.1: Locate handleApprovalDecision function (approve/deny)
  - [ ] Subtask 2.2: After recording decision, check for playbook context
  - [ ] Subtask 2.3: Call appropriate status post based on decision
  - [ ] Subtask 2.4: Log errors but don't block decision processing
  - [ ] Subtask 2.5: Write integration tests for decision flow

- [ ] Task 3: Hook into cancellation flow (AC: 3, 8, 9)
  - [ ] Subtask 3.1: Locate handleCancellation function
  - [ ] Subtask 3.2: After recording cancellation, check for playbook context
  - [ ] Subtask 3.3: Post cancellation status with reason
  - [ ] Subtask 3.4: Handle case where cancellation reason is "Other" with details
  - [ ] Subtask 3.5: Write integration tests for cancellation flow

- [ ] Task 4: Hook into timeout flow (AC: 4, 8, 9)
  - [ ] Subtask 4.1: Locate timeout handler (if implemented in Epic 6)
  - [ ] Subtask 4.2: After marking as timed out, check for playbook context
  - [ ] Subtask 4.3: Post timeout status
  - [ ] Subtask 4.4: Write integration tests for timeout flow

- [ ] Task 5: Testing and validation (AC: 1-9)
  - [ ] Subtask 5.1: Unit tests for all message formatters
  - [ ] Subtask 5.2: Integration tests for all four status changes
  - [ ] Subtask 5.3: Manual testing in real playbook channel
  - [ ] Subtask 5.4: Verify non-playbook approvals unchanged
  - [ ] Subtask 5.5: Test error handling (API failures, permissions)

## Dev Notes

### Message Format Specifications

**Approved:**
```
✅ **Approved:** TUZ-2RK | Deploy v2.1.0 to production | Approved by @jane.doe at 14:23
```

**Denied (with reason):**
```
❌ **Denied:** A-X7K9Q2 | Emergency DB access | Denied by @security.manager | Reason: Insufficient justification for P3 incident
```

**Denied (no reason):**
```
❌ **Denied:** A-X7K9Q2 | Emergency DB access | Denied by @security.manager
```

**Canceled:**
```
🚫 **Canceled:** B-3M8PN | Purchase software license | Reason: Duplicate request
```

**Timed Out:**
```
⏱️ **Timeout:** C-4R7QT | Deploy hotfix to staging | No response from @lead.engineer
```

### Implementation

```go
// Message formatters
func formatApprovedStatusMessage(approval *store.Approval, approverUsername string) string {
    details := truncateString(approval.RequestDetails, 100)
    timeStr := formatTimestamp(approval.DecidedAt)

    return fmt.Sprintf("✅ **Approved:** %s | %s | Approved by @%s at %s",
        approval.ReferenceCode,
        details,
        approverUsername,
        timeStr)
}

func formatDeniedStatusMessage(approval *store.Approval, approverUsername string) string {
    details := truncateString(approval.RequestDetails, 100)
    timeStr := formatTimestamp(approval.DecidedAt)

    message := fmt.Sprintf("❌ **Denied:** %s | %s | Denied by @%s",
        approval.ReferenceCode,
        details,
        approverUsername)

    if approval.DecisionReason != "" {
        reason := truncateString(approval.DecisionReason, 100)
        message += fmt.Sprintf(" | Reason: %s", reason)
    }

    return message
}

func formatCanceledStatusMessage(approval *store.Approval) string {
    details := truncateString(approval.RequestDetails, 100)
    reason := approval.CancelReason
    if reason == "" {
        reason = "Not specified"
    }

    return fmt.Sprintf("🚫 **Canceled:** %s | %s | Reason: %s",
        approval.ReferenceCode,
        details,
        reason)
}

func formatTimedOutStatusMessage(approval *store.Approval, approverUsername string) string {
    details := truncateString(approval.RequestDetails, 100)

    return fmt.Sprintf("⏱️ **Timeout:** %s | %s | No response from @%s",
        approval.ReferenceCode,
        details,
        approverUsername)
}

func formatTimestamp(unixMillis int64) string {
    t := time.Unix(unixMillis/1000, 0)
    return t.Format("15:04") // "14:23"
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

### Integration Points

```go
// In approval decision handler
func (r *CommandRouter) handleApprovalDecision(approval *store.Approval, decision string, reason string) error {
    // ... record decision ...

    // Post to playbook channel
    if approval.PlaybookRunID != "" {
        approverUser, _ := r.api.GetUser(approval.ApproverUserID)

        var message string
        if decision == "approved" {
            message = formatApprovedStatusMessage(approval, approverUser.Username)
        } else {
            message = formatDeniedStatusMessage(approval, approverUser.Username)
        }

        _, err := r.playbooksClient.PostPlaybookStatus(approval.PlaybookRunID, message)
        if err != nil {
            r.api.LogWarn("Failed to post playbook status update", "error", err.Error())
        }
    }

    // ... continue with notifications ...
    return nil
}

// In cancellation handler
func (r *CommandRouter) handleCancellation(approval *store.Approval) error {
    // ... record cancellation ...

    // Post to playbook channel
    if approval.PlaybookRunID != "" {
        message := formatCanceledStatusMessage(approval)
        _, err := r.playbooksClient.PostPlaybookStatus(approval.PlaybookRunID, message)
        if err != nil {
            r.api.LogWarn("Failed to post cancellation status", "error", err.Error())
        }
    }

    // ... continue with notifications ...
    return nil
}

// In timeout handler (Epic 6.1)
func (r *CommandRouter) handleTimeout(approval *store.Approval) error {
    // ... mark as timed out ...

    // Post to playbook channel
    if approval.PlaybookRunID != "" {
        approverUser, _ := r.api.GetUser(approval.ApproverUserID)
        message := formatTimedOutStatusMessage(approval, approverUser.Username)
        _, err := r.playbooksClient.PostPlaybookStatus(approval.PlaybookRunID, message)
        if err != nil {
            r.api.LogWarn("Failed to post timeout status", "error", err.Error())
        }
    }

    // ... continue with notifications ...
    return nil
}
```

### Files to Modify

**Modified Files:**
- `server/command/router.go` - Add status posting to all status change handlers
- `server/playbooks_client.go` - Message formatting helpers (or separate formatter file)
- `server/command/router_test.go` - Add tests for all status change scenarios

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All four status changes post to playbook channel
- [ ] Message formatting consistent and clear
- [ ] Error handling prevents blocking
- [ ] Non-playbook approvals unchanged
- [ ] Unit tests passing (100% coverage for formatters)
- [ ] Integration tests passing (all four scenarios)
- [ ] Manual testing in real playbook completed
- [ ] Code review approved
- [ ] Ready for Story 8.6 (error handling hardening)

## Related Stories

- **Depends on:** Story 8.1 (playbook detection)
- **Depends on:** Story 8.2 (playbook fields in approval)
- **Depends on:** Story 8.3 (status posting method)
- **Blocks:** Epic 8 completion

## Technical Debt / Future Improvements

- Update original status post instead of creating new ones (thread replies?)
- Add reactions to status posts for quick acknowledgment
- Support @mentions in status posts to alert specific team members
- Add "View Approval Details" button in status posts
- Consider webhook notifications for external systems
