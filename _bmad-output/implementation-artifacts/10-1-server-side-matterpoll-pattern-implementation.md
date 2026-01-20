# Story 10.1: Server-Side Matterpoll Pattern Implementation

Status: done

## Story

As a plugin developer,
I want to implement the Matterpoll pattern for creating interactive posts,
so that interactive buttons work correctly with custom post types.

## Acceptance Criteria

### AC1: Helper Function for Interactive Posts
- Create `server/notifications/interactive_post.go`
- Implement `CreateInteractiveApprovalPost()` function
- Uses `model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})`
- Sets `post.Type = "custom_approval_dm"`
- Includes all approval data in `post.Props`
- Stores timestamps as Unix millis (int64)

### AC2: PostAction Structure
- Create PostActions for Approve, Deny buttons
- Integration URLs: `/plugins/com.mattermost.plugin-approver2/api/v1/approval/{code}/approve|deny`
- Button styles: `success` for Approve, `danger` for Deny
- Include approval code in URL path (not Context map)

### AC3: SlackAttachment Structure
- Title: Status header (e.g., "Approval Request")
- Text: Approval description and details
- Actions: Array of PostAction buttons
- No custom fields needed (use post.Props for data)

### AC4: Post Props Schema
```go
Props: map[string]interface{}{
    "approval_code":           record.Code,
    "approval_status":         record.Status,
    "requester_username":      record.RequesterUsername,
    "requester_display_name":  record.RequesterDisplayName,
    "approver_username":       record.ApproverUsername,
    "approver_display_name":   record.ApproverDisplayName,
    "description":             record.Description,
    "created_at":              record.CreatedAt,      // Unix millis
    "decided_at":              record.DecidedAt,      // Unix millis (0 if pending)
    "decision_comment":        record.DecisionComment,
    "notification_type":       "approval_request",    // or "outcome", "cancellation", etc.
    "is_dm":                   true,
}
```

### AC5: Markdown Fallback
- Set `post.Message` with markdown content for non-webapp clients
- Include all essential information in markdown
- Reuse existing formatter functions from `notifications/dm.go`

## Tasks / Subtasks

- [x] Task 1: Create `server/notifications/interactive_post.go` (AC: 1, 2, 3)
  - [x] 1.1: Add imports (model, approval, plugin)
  - [x] 1.2: Define `PluginID` constant (`com.mattermost.plugin-approver2`)
  - [x] 1.3: Define custom post type constant `CustomApprovalDMPostType = "custom_approval_dm"`
  - [x] 1.4: Implement `CreateApproveAction()` helper - creates approve PostAction with Integration URL
  - [x] 1.5: Implement `CreateDenyAction()` helper - creates deny PostAction with Integration URL
  - [x] 1.6: Implement `CreateInteractiveApprovalPost()` main function

- [x] Task 2: Implement `CreateInteractiveApprovalPost()` function (AC: 1, 2, 3, 4)
  - [x] 2.1: Accept parameters: `botUserID string, channelID string, record *approval.ApprovalRecord, notificationType string`
  - [x] 2.2: Build PostAction slice with Approve/Deny buttons (conditionally for pending status)
  - [x] 2.3: Create SlackAttachment with title, text, and actions
  - [x] 2.4: Create Post with custom type `custom_approval_dm`
  - [x] 2.5: Populate post.Props with approval data (AC4 schema)
  - [x] 2.6: **CRITICAL**: Call `model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})`
  - [x] 2.7: Return the prepared post (caller will create via API)

- [x] Task 3: Implement `FormatApprovalPropsForDM()` helper (AC: 4)
  - [x] 3.1: Accept `record *approval.ApprovalRecord, notificationType string`
  - [x] 3.2: Return `map[string]interface{}` with all required fields
  - [x] 3.3: Handle optional fields (decided_at, decision_comment) - only include if non-zero/non-empty
  - [x] 3.4: Include playbook context fields if available

- [x] Task 4: Implement markdown fallback message builder (AC: 5)
  - [x] 4.1: Create `FormatMarkdownFallback()` function
  - [x] 4.2: Accept `record *approval.ApprovalRecord, notificationType string`
  - [x] 4.3: Return markdown string suitable for `post.Message`
  - [x] 4.4: Reuse existing message formatting patterns from dm.go

- [x] Task 5: Add unit tests (AC: all)
  - [x] 5.1: Test `CreateApproveAction()` generates correct Integration URL
  - [x] 5.2: Test `CreateDenyAction()` generates correct Integration URL
  - [x] 5.3: Test `CreateInteractiveApprovalPost()` sets correct Type and Props
  - [x] 5.4: Test `FormatApprovalPropsForDM()` produces correct schema
  - [x] 5.5: Verify `ParseSlackAttachment` is called (can check post.Props["attachments"])

## Dev Notes

### Critical Pattern: Matterpoll's Secret Sauce

The **key insight** from analyzing Matterpoll is that `model.ParseSlackAttachment()` must be used instead of directly setting `post.Props["attachments"]`. This function properly processes the SlackAttachment structure and preserves Integration URLs even with custom post types.

**WRONG (current dm.go pattern):**
```go
post.Props = model.StringInterface{
    "attachments": []any{
        map[string]any{
            "actions": []any{...}
        },
    },
}
```

**CORRECT (Matterpoll pattern):**
```go
attachment := &model.SlackAttachment{
    Title:   "Approval Request",
    Text:    description,
    Actions: []*model.PostAction{approveAction, denyAction},
}
model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
```

### Plugin ID Pattern

The plugin ID is hardcoded throughout the codebase as `com.mattermost.plugin-approver2`. Define a constant in the new file:

```go
const PluginID = "com.mattermost.plugin-approver2"
```

### Integration URL Pattern

Matterpoll uses URL path parameters (not Context maps). For this implementation:

```go
Integration: &model.PostActionIntegration{
    URL: fmt.Sprintf("/plugins/%s/api/v1/approval/%s/approve", PluginID, record.Code),
}
```

**Note:** The existing handler at `/plugins/com.mattermost.plugin-approver2/action` uses Context maps. Story 10.2 will add new API handlers at the `/api/v1/approval/{code}/approve|deny` paths OR adapt the existing handler.

### Notification Types

The `notification_type` field differentiates DM post rendering:
- `"approval_request"` - Shows buttons (pending)
- `"outcome"` - No buttons (approved/denied)
- `"cancellation"` - No buttons (canceled)
- `"timeout"` - No buttons (timed out)
- `"verification"` - No buttons (verified)

### Existing Code Reference

See `server/notifications/dm.go` lines 520-571 for the **commented out** `FormatApprovalPropsForDM()` function. This was stubbed for Story 9.10 but never activated. Uncomment and adapt it.

### Project Structure Notes

- **New file:** `server/notifications/interactive_post.go`
- **Existing patterns:** Follow the same validation, error handling, and logging patterns as `dm.go`
- **No changes to existing functions yet:** This story creates the helper; Story 10.3 will call it

### Architecture Compliance

- **Graceful Degradation (AD 2.2):** Post creation failures should not break approval workflow
- **KV Store:** No KV store changes in this story (approval record structure unchanged)
- **Error Handling:** Return errors to caller; caller decides to log and continue

### Testing Standards

- Unit tests in `server/notifications/interactive_post_test.go`
- Test that `ParseSlackAttachment` properly populates `post.Props["attachments"]`
- Test that Integration URLs follow expected format
- Test edge cases: nil record, empty fields, different notification types

### References

- [Source: matterpoll-interactive-buttons-analysis.md#10 - THE CRITICAL PATTERN (PROVEN TO WORK)]
- [Source: epic-10-dm-interactive-buttons.md#Story 10.1]
- [Source: server/notifications/dm.go#lines 520-571 - commented FormatApprovalPropsForDM]
- [Source: server/notifications/dm.go#lines 63-97 - current button implementation]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- Build: `go build ./server/...` - PASS
- Tests: `go test ./server/notifications/...` - PASS (all 12 new tests + existing tests)
- Full suite: `go test ./server/...` - PASS

### Completion Notes List

1. Created `server/notifications/interactive_post.go` with Matterpoll pattern implementation
2. Implemented `CreateApproveAction()` and `CreateDenyAction()` with URL path parameters (not Context maps)
3. Implemented `CreateInteractiveApprovalPost()` using `model.ParseSlackAttachment()` - THE CRITICAL PATTERN
4. Implemented `FormatApprovalPropsForDM()` with AC4 schema including all optional fields
5. Implemented `FormatMarkdownFallback()` for all 5 notification types
6. Added 12 comprehensive unit tests covering all functions and edge cases
7. All tests pass (existing + new)

**Key Decision:** Removed `api plugin.API` from `CreateInteractiveApprovalPost()` signature - not needed since caller creates post. This aligns with separation of concerns.

### Code Review Fixes Applied

- **H1 FIXED:** Added input validation for empty `botUserID`/`channelID` in `CreateInteractiveApprovalPost()`
- **M1 FIXED:** `TestCreateInteractiveApprovalPost_NoButtonsForNonPending` now properly verifies attachment has no actions
- **M2 FIXED:** Added edge case tests for empty code (`TestCreateApproveAction_EmptyCode`, `TestCreateDenyAction_EmptyCode`)
- **M3 NOTED:** Timestamp format inconsistency matches existing dm.go patterns (intentional)
- Added tests: `TestCreateInteractiveApprovalPost_EmptyBotUserID`, `TestCreateInteractiveApprovalPost_EmptyChannelID`

### File List

- `server/notifications/interactive_post.go` (NEW - 268 lines)
- `server/notifications/interactive_post_test.go` (NEW - 298 lines)
