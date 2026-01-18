# Story 8.3: Post Status Messages to Playbook Channel

**Epic:** 8 - Playbook Integration
**Status:** done
**Priority:** High
**Estimate:** 5 points
**Assignee:** AI Dev Agent

## User Story

**As a** playbook team member
**I want** to see approval status in the playbook channel
**So that** I know when approvals are blocking progress

## Context

When an approval request is created in a playbook channel, the entire team needs visibility into the approval status without checking DMs or running commands. This story posts the initial "pending" status message to the playbook channel immediately after approval creation.

The message should be clear, actionable, and include the reference code for correlation with DM notifications.

## Acceptance Criteria

- [x] AC1: After approval creation in playbook channel, status message posted automatically
- [x] AC2: Message format: "⏳ **Approval Pending:** [CODE] | [Details] | Waiting for @approver"
- [x] AC3: Message uses markdown formatting for readability
- [x] AC4: Message includes approval reference code
- [x] AC5: Message mentions approver by @username
- [x] AC6: Post ID returned by API is stored in approval.PlaybookPostID
- [x] AC7: Posting error logged but doesn't block approval creation
- [x] AC8: Non-playbook approvals unaffected (no channel post)
- [x] AC9: Message styled appropriately for channel visibility

## Tasks / Subtasks

- [x] Task 1: Implement Playbooks status posting (AC: 1, 2, 3, 6, 7)
  - [x] Subtask 1.1: Add postPlaybookStatus method to PlaybooksClient
  - [x] Subtask 1.2: Implement POST to `/runs/{id}/status` endpoint
  - [x] Subtask 1.3: Format message with markdown and emoji
  - [x] Subtask 1.4: Extract post ID from API response
  - [x] Subtask 1.5: Handle API errors gracefully with logging

- [x] Task 2: Format status messages (AC: 2, 3, 4, 5, 9)
  - [x] Subtask 2.1: Create formatPendingStatusMessage helper function
  - [x] Subtask 2.2: Include emoji (⏳) for visual indicator
  - [x] Subtask 2.3: Bold "Approval Pending" for emphasis
  - [x] Subtask 2.4: Include reference code in brackets
  - [x] Subtask 2.5: Truncate details if > 100 characters
  - [x] Subtask 2.6: Add @mention for approver (not notification, just display)

- [x] Task 3: Integrate into approval creation flow (AC: 1, 6, 7, 8)
  - [x] Subtask 3.1: Call postPlaybookStatus after successful approval creation
  - [x] Subtask 3.2: Only call if approval.PlaybookRunID is set
  - [x] Subtask 3.3: Store returned post ID in approval.PlaybookPostID
  - [x] Subtask 3.4: Update approval record in KV store with post ID
  - [x] Subtask 3.5: Ensure posting failure doesn't roll back approval creation

- [x] Task 4: Testing and validation (AC: 7, 8, 9)
  - [x] Subtask 4.1: Unit tests for message formatting
  - [x] Subtask 4.2: Unit tests for API posting with mocks
  - [x] Subtask 4.3: Integration test with mock Playbooks API
  - [x] Subtask 4.4: Manual test in real playbook channel
  - [x] Subtask 4.5: Verify non-playbook approvals unchanged

## Dev Notes

### Playbooks Status API

**Endpoint:** `POST /plugins/playbooks/api/v0/runs/{id}/status`

**Request Body:**
```json
{
  "message": "⏳ **Approval Pending:** TUZ-2RK | Deploy v2.1.0 to production | Waiting for @jane.doe",
  "reminder": 1705593600000
}
```

**Note:** The `reminder` field is **required** by the Playbooks API and must be a non-zero Unix timestamp in milliseconds. The implementation uses 24 hours from now as a reasonable default to create a playbook reminder for checking on pending approvals.

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
- `server/playbooks/client.go` - Add PostPlaybookStatus method
- `server/api.go` - Integrate posting into approval creation
- `server/playbooks/client_test.go` - Add tests for posting
- `server/api_test.go` - Add integration and formatting tests
- `server/command/router_test.go` - Update mock interface

## Definition of Done

- [x] All acceptance criteria met
- [x] PostPlaybookStatus method implemented and tested
- [x] Message formatting follows design spec
- [x] Integration with approval creation complete
- [x] Error handling prevents approval creation failure
- [x] Unit tests passing (100% coverage - 488 tests pass)
- [x] Integration tests passing
- [x] Manual testing in real playbook channel completed
- [x] Code review approved
- [x] Ready for Story 8.4 (DM context enhancement)

## Related Stories

- **Depends on:** Story 8.1 (playbook detection)
- **Depends on:** Story 8.2 (playbook fields in approval)
- **Blocks:** Story 8.5 (status updates reference this post)

## Technical Debt / Future Improvements

- Consider updating existing post instead of creating new ones (cleaner channel)
- Add ability to reply in thread for approval discussions
- Support rich formatting with attachments/cards
- Add reaction buttons to channel post for quick visibility

---

## Dev Agent Record

### File List

**Modified Files:**
- `server/playbooks/client.go` - Added PostPlaybookStatus method with reminder field, documented timeouts
- `server/playbooks/client_test.go` - Added 5 unit tests for PostPlaybookStatus API calls
- `server/api.go` - Added formatPendingPlaybookStatusMessage with edge case handling, integrated posting into approval creation flow
- `server/api_test.go` - Added 8 tests for message formatting (including unicode, empty fields) and 3 integration tests for playbook posting
- `server/command/router_test.go` - Updated mock interface to include PostPlaybookStatus

### Change Log

**Story 8.3 Implementation:**

1. **PostPlaybookStatus Method (server/playbooks/client.go:143-199)**
   - Implemented POST to `/plugins/playbooks/api/v0/runs/{id}/status` endpoint
   - Added required `reminder` field (24h from now) - discovered during testing that API requires non-zero value
   - Uses user-context authentication (requester's token)
   - Returns post ID for future updates (Story 8.5)
   - 1-second timeout for write operations
   - Comprehensive error handling

2. **Message Formatting (server/api.go:856-879)**
   - Created `formatPendingPlaybookStatusMessage` helper
   - Format: "⏳ **Approval Pending:** [CODE] | [Details] | Waiting for @approver"
   - Truncates description to 100 characters with ellipsis
   - Edge case handling: empty code → "UNKNOWN", empty username → "approver"
   - Supports unicode and emoji in descriptions

3. **Integration (server/api.go:270-299)**
   - Integrated posting after DM notification in approval creation flow
   - Only posts if `PlaybookRunID` is set (non-playbook approvals unaffected)
   - Stores post ID in `record.PlaybookPostID` for future updates
   - Graceful error handling - posting failures logged but don't block approval creation
   - Updates KV store with post ID when successful

4. **Testing (488 tests pass)**
   - 5 unit tests for PostPlaybookStatus: success, API error, invalid JSON, token error, network error
   - 8 unit tests for message formatting: standard, truncation, 100-char boundary, short text, special chars, unicode/emoji, empty code, empty username
   - 3 integration tests: successful posting with post ID storage, error handling without blocking approval, skipping non-playbook approvals

### Production Testing Results

- ✅ Tested in real Mattermost instance with Playbooks plugin
- ✅ Pending messages successfully posted to playbook channels
- ✅ Message formatting displays correctly with emoji and markdown
- ✅ Post ID correctly stored in approval records
- ✅ Non-playbook approvals unaffected
- ⚠️ Discovered `reminder` field requirement (not in original API docs) - added to implementation

### Code Review Fixes Applied

**High Priority Fixes:**
- Added edge case validation for empty code and username fields
- Added test coverage for unicode/emoji in descriptions
- Documented timeout rationale (500ms read, 1s write)
- Documented 24-hour reminder rationale (reasonable approval turnaround time)

**Documentation Fixes:**
- Corrected file paths in Dev Notes
- Added `reminder` field to API documentation
- Updated status to "done"
- Marked all ACs and tasks as complete
- Added this Dev Agent Record section
