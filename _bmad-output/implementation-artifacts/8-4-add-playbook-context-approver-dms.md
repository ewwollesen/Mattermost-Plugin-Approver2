# Story 8.4: Add Playbook Context to Approver DM Notifications

**Epic:** 8 - Playbook Integration
**Status:** ready-for-dev
**Priority:** Medium
**Estimate:** 3 points
**Assignee:** TBD

## User Story

**As an** approver
**I want** to see which playbook the approval request came from
**So that** I understand the urgency and context

## Context

When approvers receive DM notifications for approval requests originating from playbook channels, they need context about which playbook (incident, deploy, change management) the request is associated with. This helps approvers:

- Understand urgency (incident vs routine)
- Navigate back to playbook channel for more context
- Prioritize multiple pending approvals

This story enhances the existing DM notification to include playbook name and channel link when available.

## Acceptance Criteria

- [ ] AC1: Approver DM includes playbook name when approval is playbook-linked
- [ ] AC2: Format: Section header "**Playbook Context:**" followed by playbook name
- [ ] AC3: Link to playbook channel included for easy navigation
- [ ] AC4: Non-playbook approvals show no playbook section (v1.0 behavior)
- [ ] AC5: Playbook context appears before Approve/Deny buttons
- [ ] AC6: Context uses consistent formatting with rest of DM
- [ ] AC7: Works for initial notification and reminder notifications
- [ ] AC8: Playbook name truncated if > 50 characters

## Tasks / Subtasks

- [ ] Task 1: Update DM notification template (AC: 1, 2, 3, 5, 6)
  - [ ] Subtask 1.1: Locate existing DM notification formatting code
  - [ ] Subtask 1.2: Add conditional playbook context section
  - [ ] Subtask 1.3: Format playbook name with bold header
  - [ ] Subtask 1.4: Generate channel link using playbook channel ID
  - [ ] Subtask 1.5: Position section above action buttons

- [ ] Task 2: Implement context formatting (AC: 2, 3, 6, 8)
  - [ ] Subtask 2.1: Create formatPlaybookContext helper function
  - [ ] Subtask 2.2: Check if approval has playbook metadata
  - [ ] Subtask 2.3: Format section with header and playbook name
  - [ ] Subtask 2.4: Add clickable channel link
  - [ ] Subtask 2.5: Truncate long playbook names with ellipsis
  - [ ] Subtask 2.6: Write unit tests for formatting

- [ ] Task 3: Integrate into notification flow (AC: 1, 4, 7)
  - [ ] Subtask 3.1: Update sendApproverNotification function
  - [ ] Subtask 3.2: Pass approval object with playbook fields
  - [ ] Subtask 3.3: Conditionally append playbook context
  - [ ] Subtask 3.4: Ensure non-playbook approvals unchanged
  - [ ] Subtask 3.5: Apply to reminder notifications (if implemented)

- [ ] Task 4: Testing and validation (AC: 4, 7, 8)
  - [ ] Subtask 4.1: Unit tests for playbook context formatting
  - [ ] Subtask 4.2: Integration tests for DM with/without playbook
  - [ ] Subtask 4.3: Manual test in real playbook channel
  - [ ] Subtask 4.4: Verify non-playbook DMs unchanged
  - [ ] Subtask 4.5: Test with long playbook names

## Dev Notes

### Current DM Notification Structure

```go
// In server/command/router.go (approximate)
func (r *CommandRouter) sendApproverNotification(approval *store.Approval) error {
    message := fmt.Sprintf("**Approval Request:** %s\n\n", approval.ReferenceCode)
    message += fmt.Sprintf("**From:** @%s\n", requesterUsername)
    message += fmt.Sprintf("**Details:** %s\n\n", approval.RequestDetails)
    // [Existing approve/deny buttons]

    post := &model.Post{
        UserId:    r.botUserID,
        ChannelId: dmChannel.Id,
        Message:   message,
        Props:     map[string]interface{}{...},
    }

    return r.api.CreatePost(post)
}
```

### Enhanced DM with Playbook Context

```go
func (r *CommandRouter) sendApproverNotification(approval *store.Approval) error {
    message := fmt.Sprintf("**Approval Request:** %s\n\n", approval.ReferenceCode)
    message += fmt.Sprintf("**From:** @%s\n", requesterUsername)
    message += fmt.Sprintf("**Details:** %s\n\n", approval.RequestDetails)

    // Add playbook context if present
    if approval.PlaybookRunID != "" {
        message += formatPlaybookContext(approval)
    }

    message += "\n**Action Required:** Please review and respond below.\n"

    // [Existing approve/deny buttons]
    return r.api.CreatePost(post)
}

func formatPlaybookContext(approval *store.Approval) string {
    playbookName := approval.PlaybookName
    if len(playbookName) > 50 {
        playbookName = playbookName[:47] + "..."
    }

    channelLink := fmt.Sprintf("~%s", approval.PlaybookChannelID)

    return fmt.Sprintf("**Playbook Context:**\n"+
        "- Playbook: %s\n"+
        "- Channel: %s\n\n",
        playbookName,
        channelLink)
}
```

### Example DM Messages

**With Playbook Context:**
```
**Approval Request:** TUZ-2RK

**From:** @wayne
**Details:** Deploy v2.1.0 to production

**Playbook Context:**
- Playbook: Deploy - Production Release v2.1.0
- Channel: ~deploy-prod-v2-1-0

**Action Required:** Please review and respond below.

[Approve] [Deny]
```

**Without Playbook Context (v1.0):**
```
**Approval Request:** A-X7K9Q2

**From:** @wayne
**Details:** Emergency access to production database

**Action Required:** Please review and respond below.

[Approve] [Deny]
```

### Files to Modify

**Modified Files:**
- `server/command/router.go` - Update sendApproverNotification function
- `server/command/router_test.go` - Add tests for playbook context in DMs

## Definition of Done

- [ ] All acceptance criteria met
- [ ] DM notification includes playbook context when present
- [ ] Formatting matches design spec
- [ ] Non-playbook notifications unchanged
- [ ] Unit tests passing (100% coverage)
- [ ] Integration tests passing
- [ ] Manual testing in real playbook completed
- [ ] Code review approved
- [ ] Ready for Story 8.5 (status change updates)

## Related Stories

- **Depends on:** Story 8.2 (playbook fields in approval)
- **Related to:** Story 8.3 (channel posts provide complementary visibility)

## Technical Debt / Future Improvements

- Add "View Playbook Run" button that deep-links to playbook
- Include playbook owner information
- Add urgency indicator based on playbook type
- Support custom context messages per playbook template
