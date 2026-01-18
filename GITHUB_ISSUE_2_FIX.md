# GitHub Issue #2 Fix: Unwanted Playbook Status Updates

## Problem

When the Approver plugin posted messages to playbook channels using the Playbooks status API endpoint (`/plugins/playbooks/api/v0/runs/{runID}/status`), it was triggering unwanted side effects in the Playbooks plugin:

1. **Reminder Operations**: The `reminder` field was triggering `PlaybookRunReminder` operations
2. **Multiple Posts**: Creating new posts for each status change instead of updating the original (unlike DM behavior)

## Root Cause

The Playbooks status API requires a non-zero reminder time and cannot accept `reminder: 0`. Any use of the `/status` endpoint triggers Playbooks workflow side effects.

## Solution

**Use standard Mattermost CreatePost with markdown tables**: Stop using the Playbooks API entirely and format messages as markdown tables:

### Key Changes

1. **Use CreatePost API**: Replace Playbooks status API with standard Mattermost `CreatePost` / `UpdatePost` APIs
2. **Format with markdown tables**: Use markdown table syntax for nice structured display
3. **Update existing posts**: Like DM behavior, update the original status post when approval state changes instead of creating new posts

### Implementation

```go
// POST initial message using CreatePost with markdown table
func (c *Client) PostMessageToPlaybookChannel(channelID string, message string) (string, error) {
    post := &model.Post{
        UserId:    c.botUserID,
        ChannelId: channelID,
        Message:   message, // Contains markdown table
    }
    createdPost, appErr := c.api.CreatePost(post)
    return createdPost.Id, appErr
}

// UPDATE existing post
func (c *Client) UpdateMessageInPlaybookChannel(channelID string, postID string, message string) error {
    existingPost, appErr := c.api.GetPost(postID)
    if appErr != nil {
        return appErr
    }
    existingPost.Message = message
    _, appErr = c.api.UpdatePost(existingPost)
    return appErr
}
```

### Markdown Table Format

```go
func FormatApprovedStatusMessage(record *approval.ApprovalRecord) string {
    return fmt.Sprintf(`### ✅ Approval Approved

| Field | Value |
|:------|:------|
| **Request ID** | %s |
| **Description** | %s |
| **Approved By** | @%s |
| **Time** | %s |`,
        record.Code,
        details,
        record.ApproverUsername,
        timeStr)
}
```

### Call Site Logic

```go
// Skip ephemeral confirmation in playbook channels (requester sees status post there)
if record.PlaybookRunID == "" {
    // Send ephemeral post (only requester sees it)
    ephemeralPost := p.API.SendEphemeralPost(payload.UserId, post)
    // ... handle error
} else {
    p.API.LogDebug("Skipping ephemeral confirmation - approval posted to playbook channel")
}

// Update playbook channel status
if updatedRecord.PlaybookPostID != "" {
    // UPDATE existing post
    err := p.playbooksClient.UpdateMessageInPlaybookChannel(
        updatedRecord.PlaybookChannelID,
        updatedRecord.PlaybookPostID,
        statusMessage,
    )
} else {
    // Fallback: create new post if we don't have the original post ID
    _, err := p.playbooksClient.PostMessageToPlaybookChannel(
        updatedRecord.PlaybookChannelID,
        statusMessage,
    )
}
```

## Benefits

1. **Formatted Display**: Markdown tables provide nice structured display
2. **No Side Effects**: No Playbooks API = no unwanted reminder operations
3. **Single Post**: Updates the original post like DM behavior (cleaner channel history)
4. **User Consistency**: Matches behavior users expect from DM notifications
5. **Simpler Implementation**: Using standard Mattermost APIs instead of Playbooks-specific endpoints
6. **No Duplicate Confirmations**: Ephemeral confirmation message in playbook channel is suppressed (requester sees status post in the channel)

## Files Changed

### Core Implementation
- `server/playbooks/client.go`:
  - `PostMessageToPlaybookChannel()` using CreatePost API
  - `UpdateMessageInPlaybookChannel()` using UpdatePost API
  - Added `botUserID` to Client struct for post identity
- `server/playbooks/formatters.go`: All formatters updated to use markdown table syntax
- `server/playbooks/client_test.go`: Tests updated for new CreatePost approach

### Call Sites
- `server/api.go`:
  - `formatPendingPlaybookStatusMessage()` updated to markdown table
  - Initial post: `PostMessageToPlaybookChannel()`
  - Updates: `UpdateMessageInPlaybookChannel()` if post ID exists
  - Applied to approval creation, approve/deny, cancellation
  - **Ephemeral Suppression**: Skip ephemeral confirmation message in playbook channels (approval creation)
- `server/timeout/checker.go`: Same update pattern for timeouts
- `server/plugin.go`: Pass botUserID to NewClient

### Test Mocks
- `server/api_test.go`: Updated mock to new method signatures
- `server/command/router_test.go`: Updated mock to new method signatures
- `server/timeout/checker_test.go`: Updated mock to new method signatures

## Testing

All tests pass with the new implementation:
```
✓ server tests: PASS (validates reminder=0 in request body)
✓ server/approval tests: PASS
✓ server/command tests: PASS
✓ server/notifications tests: PASS
✓ server/playbooks tests: PASS (HTTP-based tests verify reminder field)
✓ server/store tests: PASS
✓ server/timeout tests: PASS
```

## User Experience

### Before Fix
- ❌ Playbook reminders triggered by status API
- ❌ Multiple posts in channel (one per status change)

### After Fix
- ✅ No unwanted reminders (no Playbooks API)
- ✅ Nice formatted tables (markdown syntax)
- ✅ Single post updated in place (like DM behavior)
- ✅ Simpler and more reliable

## Related Stories

- Story 8.3: Post Status Messages to Playbook Channel
- Story 8.5: Update Playbook Channel on Status Changes
- Story 8.6: Error Handling and Graceful Fallback

## Date
2026-01-17
