# Story 4.1: Update Approver DM Post on Cancellation

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Status:** Done
**Priority:** High
**Estimate:** 3 points
**Assignee:** Dev Agent (Claude)

## User Story

**As an** approver
**I want** the original approval request message to visually update when cancelled
**So that** I immediately see it's no longer actionable

## Context

Currently, when a requester cancels an approval request, the approver's DM still shows the original message with active approve/deny buttons. These buttons no longer work (ghost buttons), creating confusion and a poor user experience. The approver has no visual indication that the request has been cancelled.

## Acceptance Criteria

- [x] Original DM post shows 🚫 status banner with "Canceled" state
- [x] Request description shown in plain text (strikethrough removed per user feedback for cleaner UX)
- [x] Approve/Deny action buttons are completely removed from the post
- [x] Post shows "Canceled by @requester at [timestamp]"
- [x] Post update triggers immediately when cancel command executes
- [x] If post update fails (deleted channel, permissions, etc.), error is logged but process continues (Architecture Decision 2.2: Graceful Degradation)
- [x] Updated post is visually distinct from pending requests

## Technical Implementation

### Files to Modify
- `server/notifications/dm.go` - Add `UpdateApprovalPostForCancellation()` function
- `server/plugin.go` - Call post update in `handleCancelCommand()` after status change
- `server/approval/models.go` - NotificationPostID field already exists for storing approver post ID

### Key Functions

```go
// UpdateApprovalPostForCancellation updates the approver's DM post to show cancelled state
func (p *Plugin) UpdateApprovalPostForCancellation(request *ApprovalRequest, canceledBy string) error {
    if request.ApproverPostID == "" {
        p.API.LogWarn("Cannot update approver post: no post ID stored", "request_id", request.ID)
        return fmt.Errorf("no approver post ID found")
    }

    // Get the original post
    post, err := p.API.GetPost(request.ApproverPostID)
    if err != nil {
        p.API.LogError("Failed to get post for update", "post_id", request.ApproverPostID, "error", err.Error())
        return err
    }

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

    // Remove action buttons (props)
    post.Message = updatedMessage
    post.Props = map[string]interface{}{} // Clear all interactive elements

    // Update the post
    _, err = p.API.UpdatePost(post)
    if err != nil {
        p.API.LogError("Failed to update post", "post_id", request.ApproverPostID, "error", err.Error())
        return err
    }

    return nil
}
```

### Mattermost API Usage

```go
// Update existing post
updatedPost, err := p.API.UpdatePost(post)
```

**API Documentation:** https://api.mattermost.com/#tag/posts/operation/UpdatePost

### Error Handling

1. **Post doesn't exist** (deleted channel, archived, etc.)
   - Log warning with request ID and post ID
   - Continue with cancellation process
   - Send DM notification as backup

2. **Permission denied**
   - Log error
   - Continue with cancellation process
   - Bot should have permission if it created the post originally

3. **Network/API failure**
   - Log error with details
   - Retry once after 1 second delay
   - If still fails, continue (DM notification will still go through)

## Testing Requirements

### Unit Tests
- [x] Test post update with valid request
- [x] Test handling when post ID is empty
- [x] Test handling when post no longer exists
- [x] Test markdown formatting (plain text without strikethrough)
- [x] Test timestamp formatting
- [x] Test button removal (Props cleared)
- [x] Test who canceled display
- [x] Test UpdatePost failure handling
- [x] Test nil record validation
- [x] Test message format completeness

### Integration Tests
- [x] Create approval request → Cancel → Verify post updated (api_test.go)
- [x] Verify buttons removed from post props
- [x] Verify status banner appears
- [x] Test cancellation by different users (permission tests)

### Manual Testing
- [x] Create real approval in Mattermost
- [x] Cancel as requester
- [x] Verify approver sees updated post immediately
- [x] Verify buttons are gone
- [x] Verify formatting renders correctly (clean, non-header text approved by user)
- [x] User feedback: Strikethrough removed for cleaner appearance

## Dependencies

- **Depends on:** Story 4.4 (Store cancellation reason) - needs `CanceledAt` field
- **Blocks:** Story 4.7 (Fix ghost buttons) - same functionality, coordinate implementation
- **Related:** Story 4.2 (Send notification) - both fire on cancellation

## UI/UX Mockup

**Before Cancellation:**
```
┌────────────────────────────────────────────┐
│ 📋 New Approval Request                    │
│                                            │
│ Reference Code: TUZ-2RK                    │
│ From: @wayne                               │
│                                            │
│ Deploy hotfix to production database      │
│                                            │
│ [Approve] [Deny]                           │
└────────────────────────────────────────────┘
```

**After Cancellation (Final Implementation):**
```
┌────────────────────────────────────────────┐
│ 🚫 Approval Request (Canceled)             │
│                                            │
│ From: @wayne                               │
│ Request ID: TUZ-2RK                        │
│                                            │
│ Description:                               │
│ Deploy hotfix to production database      │
│                                            │
│ ─────────────────────────────────────────  │
│ Canceled by @wayne at Jan 12, 2026 7:15 PM│
└────────────────────────────────────────────┘
```

**Design Note:** Original spec included markdown header (`##`) and strikethrough (`~~text~~`) but these were removed after user testing. Header made text too large compared to other requests. Strikethrough was deemed unnecessary—removal of buttons + status indicator + cancellation footer provide sufficient visual distinction.

## Definition of Done

- [x] Code implemented and reviewed (adversarial code review completed)
- [x] Unit tests written and passing (10 test cases, 299/299 total tests passing)
- [x] Integration tests passing (cancel flow end-to-end validated)
- [x] Manual testing completed in real Mattermost instance (user approved final UX)
- [x] Error handling verified for edge cases (graceful degradation pattern applied)
- [ ] Code merged to master (currently uncommitted, ready for commit)
- [ ] Documented in commit message (pending commit)

## Dev Agent Record

### Implementation Summary

**Story:** 4.1 - Update Approver DM Post on Cancellation
**Completed:** January 12, 2026
**Agent:** Dev Agent (Claude)
**Test Results:** 299/299 tests passing (10 new tests added)

### Files Modified

1. **server/notifications/dm.go** (lines 177-230)
   - Added `UpdateApprovalPostForCancellation()` function
   - Validates record and post ID before attempting update
   - Formats cancellation message with plain text description
   - Clears post Props to remove action buttons (fixes ghost buttons)
   - Implements graceful degradation per Architecture Decision 2.2

2. **server/notifications/dm_test.go** (10 new test cases)
   - TestUpdateApprovalPostForCancellation
   - TestUpdateApprovalPostForCancellation_NoStrikethrough
   - TestUpdateApprovalPostForCancellation_ButtonsRemoved
   - TestUpdateApprovalPostForCancellation_TimestampFormat
   - TestUpdateApprovalPostForCancellation_WhoCanceled
   - TestUpdateApprovalPostForCancellation_EmptyPostID
   - TestUpdateApprovalPostForCancellation_PostNotFound
   - TestUpdateApprovalPostForCancellation_UpdatePostFails
   - TestUpdateApprovalPostForCancellation_NilRecord
   - TestUpdateApprovalPostForCancellation_MessageFormat

3. **server/plugin.go** (cancel command integration)
   - Integrated UpdateApprovalPostForCancellation into cancel flow
   - Added graceful error handling for post update failures

4. **server/plugin_test.go** (integration tests)
   - Added cancellation flow tests with post update verification

5. **server/api_test.go** (end-to-end tests)
   - Added API-level cancellation tests verifying complete flow

### Change Log

**Initial Implementation:**
- Implemented post update with strikethrough description: `~~%s~~`
- Used markdown header for title: `## 🚫 Approval Request (Canceled)`

**User Feedback Iteration 1:**
> "the title of the post is in a header format...I'd like to just keep it regular sized"

- Changed from markdown header `##` to bold text: `🚫 **Approval Request (Canceled)**`

**User Feedback Iteration 2:**
> "I'm not sold on the strikethrough now that I see it...getting rid of the buttons and adding the cancel text...and also marking it canceled is enough"

- Removed strikethrough from description text
- Changed format from `~~%s~~` to plain `%s`
- Updated comment documentation to reflect "plain description text"

**Final Implementation:**
- Clean visual design with bold title (not header)
- Plain text description (no strikethrough)
- Complete button removal via Props clearing
- Clear cancellation footer with username and timestamp

### Architecture Compliance

- **Architecture Decision 2.2 (Graceful Degradation):** Post update failures are logged but don't fail the cancellation operation
- **Testing Standards:** Comprehensive unit and integration test coverage
- **Error Handling:** Validates inputs, handles missing posts, logs all failures appropriately

### User Acceptance

User explicitly approved the final UX design after manual testing:
- Cleaner appearance without header formatting
- Sufficient visual distinction without strikethrough
- Button removal + status indicator + cancellation footer provides clear state

## Notes

**Why strikethrough instead of delete:**
Preserves audit trail. Approver can still see what the request was about, which is important for:
- Understanding context if referenced later
- Compliance/audit requirements
- Debugging if something goes wrong

**Why update instead of new message:**
- Keeps DM thread clean (no duplicate messages)
- Clearer UX: "this request changed state" vs "here's a new thing"
- Mattermost notification behavior: updates don't always trigger new notifications (that's why Story 4.2 sends a separate DM)
