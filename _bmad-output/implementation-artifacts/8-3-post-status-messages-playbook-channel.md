# Story 8.3: Post Status Messages to Playbook Channel

**Epic:** 8 - Playbook Integration
**Status:** ready-for-dev
**Priority:** High
**Estimate:** 5 points
**Assignee:** TBD

## User Story

**As a** playbook team member
**I want** to see approval status in the playbook channel
**So that** I know when approvals are blocking progress

## Context

When an approval request is created in a playbook channel, the entire team needs visibility into the approval status without checking DMs or running commands. This story posts the initial "pending" status message to the playbook channel immediately after approval creation.

The message should be clear, actionable, and include the reference code for correlation with DM notifications.

## Acceptance Criteria

- [ ] AC1: After approval creation in playbook channel, status message posted automatically
- [ ] AC2: Message format: "⏳ **Approval Pending:** [CODE] | [Details] | Waiting for @approver"
- [ ] AC3: Message uses markdown formatting for readability
- [ ] AC4: Message includes approval reference code
- [ ] AC5: Message mentions approver by @username
- [ ] AC6: Post ID returned by API is stored in approval.PlaybookPostID
- [ ] AC7: Posting error logged but doesn't block approval creation
- [ ] AC8: Non-playbook approvals unaffected (no channel post)
- [ ] AC9: Message styled appropriately for channel visibility

## Tasks / Subtasks

- [ ] Task 1: Implement Playbooks status posting (AC: 1, 2, 3, 6, 7)
  - [ ] Subtask 1.1: Add postPlaybookStatus method to PlaybooksClient
  - [ ] Subtask 1.2: Implement POST to `/runs/{id}/status` endpoint
  - [ ] Subtask 1.3: Format message with markdown and emoji
  - [ ] Subtask 1.4: Extract post ID from API response
  - [ ] Subtask 1.5: Handle API errors gracefully with logging

- [ ] Task 2: Format status messages (AC: 2, 3, 4, 5, 9)
  - [ ] Subtask 2.1: Create formatPendingStatusMessage helper function
  - [ ] Subtask 2.2: Include emoji (⏳) for visual indicator
  - [ ] Subtask 2.3: Bold "Approval Pending" for emphasis
  - [ ] Subtask 2.4: Include reference code in brackets
  - [ ] Subtask 2.5: Truncate details if > 100 characters
  - [ ] Subtask 2.6: Add @mention for approver (not notification, just display)

- [ ] Task 3: Integrate into approval creation flow (AC: 1, 6, 7, 8)
  - [ ] Subtask 3.1: Call postPlaybookStatus after successful approval creation
  - [ ] Subtask 3.2: Only call if approval.PlaybookRunID is set
  - [ ] Subtask 3.3: Store returned post ID in approval.PlaybookPostID
  - [ ] Subtask 3.4: Update approval record in KV store with post ID
  - [ ] Subtask 3.5: Ensure posting failure doesn't roll back approval creation

- [ ] Task 4: Testing and validation (AC: 7, 8, 9)
  - [ ] Subtask 4.1: Unit tests for message formatting
  - [ ] Subtask 4.2: Unit tests for API posting with mocks
  - [ ] Subtask 4.3: Integration test with mock Playbooks API
  - [ ] Subtask 4.4: Manual test in real playbook channel
  - [ ] Subtask 4.5: Verify non-playbook approvals unchanged

## Dev Notes

### Playbooks Status API

**Endpoint:** `POST /plugins/playbooks/api/v0/runs/{id}/status`

**Request Body:**
```json
{
  "message": "⏳ **Approval Pending:** TUZ-2RK | Deploy v2.1.0 to production | Waiting for @jane.doe"
}
```

**Response (200 OK):**
```json
{
  "id": "post123",
  "create_at": 1705507200000,
  "message": "..."
}
```

### Implementation

```go
// In server/playbooks_client.go
func (c *PlaybooksClient) PostPlaybookStatus(runID, message string) (string, error) {
    url := fmt.Sprintf("%s/plugins/playbooks/api/v0/runs/%s/status",
        c.siteURL, runID)

    body, _ := json.Marshal(map[string]string{"message": message})
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return "", err
    }

    req.Header.Set("Authorization", "Bearer "+c.botToken)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 1 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to post status: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var result struct {
        ID string `json:"id"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.ID, nil
}

// Message formatting helper
func formatPendingStatusMessage(approval *store.Approval, approverUsername string) string {
    details := approval.RequestDetails
    if len(details) > 100 {
        details = details[:97] + "..."
    }

    return fmt.Sprintf("⏳ **Approval Pending:** %s | %s | Waiting for @%s",
        approval.ReferenceCode,
        details,
        approverUsername)
}
```

### Integration into Approval Creation

```go
// In server/command/router.go - after creating approval
func (r *CommandRouter) handleApprovalSubmission(submission *model.SubmitDialogResponse) error {
    // ... create approval ...

    // Post to playbook channel if playbook-linked
    if approval.PlaybookRunID != "" {
        approverUser, _ := r.api.GetUser(approval.ApproverUserID)
        message := formatPendingStatusMessage(approval, approverUser.Username)

        postID, err := r.playbooksClient.PostPlaybookStatus(approval.PlaybookRunID, message)
        if err != nil {
            r.api.LogWarn("Failed to post playbook status", "error", err.Error())
        } else {
            // Store post ID for later updates (Story 8.5)
            approval.PlaybookPostID = postID
            r.store.UpdateApproval(approval)
        }
    }

    // Continue with DM notification...
    return nil
}
```

### Message Formatting Examples

**Standard Request:**
```
⏳ **Approval Pending:** TUZ-2RK | Deploy v2.1.0 to production | Waiting for @jane.doe
```

**Long Details (truncated):**
```
⏳ **Approval Pending:** A-X7K9Q2 | Emergency read access to prod DB to debug P0 incident affecting payment proc... | Waiting for @security.manager
```

### Files to Modify

**Modified Files:**
- `server/playbooks_client.go` - Add PostPlaybookStatus method
- `server/command/router.go` - Integrate posting into approval creation
- `server/playbooks_client_test.go` - Add tests for posting
- `server/command/router_test.go` - Add integration tests

## Definition of Done

- [ ] All acceptance criteria met
- [ ] PostPlaybookStatus method implemented and tested
- [ ] Message formatting follows design spec
- [ ] Integration with approval creation complete
- [ ] Error handling prevents approval creation failure
- [ ] Unit tests passing (100% coverage)
- [ ] Integration tests passing
- [ ] Manual testing in real playbook channel completed
- [ ] Code review approved
- [ ] Ready for Story 8.4 (DM context enhancement)

## Related Stories

- **Depends on:** Story 8.1 (playbook detection)
- **Depends on:** Story 8.2 (playbook fields in approval)
- **Blocks:** Story 8.5 (status updates reference this post)

## Technical Debt / Future Improvements

- Consider updating existing post instead of creating new ones (cleaner channel)
- Add ability to reply in thread for approval discussions
- Support rich formatting with attachments/cards
- Add reaction buttons to channel post for quick visibility
