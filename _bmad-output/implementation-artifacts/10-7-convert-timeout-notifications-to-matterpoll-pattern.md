# Story 10.7: Convert Timeout Notifications to Matterpoll Pattern

Status: done

## Story

As a requester,
I want to receive timeout DMs as webapp components,
So that I know my request timed out with accurate timestamps in my timezone.

## Acceptance Criteria

### AC1: Update SendTimeoutNotificationDM()
- Use `CreateInteractiveApprovalPost()` helper
- Set `notification_type: "timeout"`
- Include timeout timestamp
- Send to the **requester** (not approver)

### AC2: Timeout Content
- Status header: "Approval Timed Out"
- Request ID and description
- Approver info (no response received)
- Timeout reason
- Created timestamp (local timezone via webapp)

### AC3: No Interactive Buttons
- Timeout notifications are read-only
- No actions array in SlackAttachment
- Just display information

### AC4: Backward Compatibility
- Markdown fallback includes all timeout details
- Works for non-webapp clients

## Tasks / Subtasks

- [x] Task 1: Update SendTimeoutNotificationDM() function (AC: 1, 2, 3)
  - [x] 1.1: Replace current implementation with `CreateInteractiveApprovalPost()` call
  - [x] 1.2: Pass `NotificationTypeTimeout` as notification type
  - [x] 1.3: Verify no buttons are rendered (CreateInteractiveApprovalPost handles this)
  - [x] 1.4: Ensure `created_at` prop is populated for timeout display

- [x] Task 2: Verify webapp component handles timeout type (AC: 2, 4)
  - [x] 2.1: Confirm ApprovalDMPost renders timeout notification type correctly
  - [x] 2.2: Verify "Approver" label renders with approver info
  - [x] 2.3: Verify status message displays "No response received (timed out)"
  - [x] 2.4: Verify "Requested" timestamp renders with Timestamp component

- [x] Task 3: Test backward compatibility (AC: 4)
  - [x] 3.1: Verify `FormatMarkdownFallback()` timeout format is correct
  - [x] 3.2: Run existing tests to ensure no regressions
  - [x] 3.3: Build plugin successfully: `make`

- [x] Task 4: Unit tests (AC: all)
  - [x] 4.1: Add test for custom_approval_dm post type
  - [x] 4.2: Add test for notification_type is "timeout"
  - [x] 4.3: Add test for created_at prop is populated
  - [x] 4.4: Add test for no buttons in attachment

## Dev Notes

### Infrastructure Already in Place

From Stories 10.1-10.6, all server infrastructure is ready:

1. **`CreateInteractiveApprovalPost()`** in `interactive_post.go`:
   - Already supports `NotificationTypeTimeout` (line 25)
   - Already omits buttons for non-approval_request types (line 78-84)

2. **`FormatApprovalPropsForDM()`** (lines 149-206):
   - Already includes `created_at` field

3. **`FormatMarkdownFallback()`** (lines 309-319):
   - Already has complete timeout formatting
   - Includes header, request ID, description, approver, reason, status

4. **`getNotificationTitle()`** (lines 131-132):
   - Returns "Approval Request Timed Out" for timeout type

5. **`ApprovalDMPost` webapp component** (lines 297-312):
   - Already handles `notification_type: "timeout"` case
   - Shows "Approver" with approver info
   - Shows "No response received (timed out)" status
   - Shows "Requested" timestamp via Timestamp component

### Current SendTimeoutNotificationDM() Analysis

**Location:** `server/notifications/dm.go` lines 232-279

**Current signature:**
```go
func SendTimeoutNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error)
```

**Current Implementation:**
- Creates standard `model.Post` with markdown message
- Manual formatting of timeout info
- Does NOT use custom post type or webapp rendering

**Key Characteristics:**
- Sends to `record.RequesterID` (requester)
- Uses `CreatedAt` timestamp (the original request time)
- Shows approver who didn't respond
- Status: "automatically canceled"

### Implementation Pattern (Same as Stories 10.5, 10.6)

```go
func SendTimeoutNotificationDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
    // Validate inputs (keep existing validation)
    if botUserID == "" {
        return "", fmt.Errorf("bot user ID not available")
    }
    if record == nil {
        return "", fmt.Errorf("approval record is nil")
    }
    if record.ID == "" {
        return "", fmt.Errorf("approval record ID is empty")
    }
    if record.RequesterID == "" {
        return "", fmt.Errorf("requester ID is empty")
    }

    // Get or create DM channel between bot and REQUESTER
    channelID, err := GetDMChannelID(api, botUserID, record.RequesterID)
    if err != nil {
        return "", fmt.Errorf("failed to get DM channel for requester %s: %w", record.RequesterID, err)
    }

    // Story 10.7: Use Matterpoll pattern helper
    post := CreateInteractiveApprovalPost(botUserID, channelID, record, NotificationTypeTimeout)
    if post == nil {
        return "", fmt.Errorf("failed to create interactive approval post")
    }

    // Send DM via CreatePost
    createdPost, appErr := api.CreatePost(post)
    if appErr != nil {
        return "", fmt.Errorf("failed to send timeout notification to requester %s: %w", record.RequesterID, appErr)
    }

    return createdPost.Id, nil
}
```

### Files to Modify

1. **`server/notifications/dm.go`** - Replace `SendTimeoutNotificationDM()` implementation
2. **`server/notifications/dm_test.go`** - Add tests for Matterpoll pattern

### Webapp Component Already Handles Timeout

The `ApprovalDMPost.tsx` already has the timeout case implemented (from Story 10.4):

```typescript
// AC3: timeout - Show timeout notice
case 'timeout':
    return (
        <>
            <InfoRow
                label="Approver"
                value={<UserMention username={data.approverUsername} displayName={data.approverDisplayName} />}
            />
            <InfoRow label="Status" value="No response received (timed out)" />
            {data.createdAt > 0 && (
                <InfoRow
                    label="Requested"
                    value={<Timestamp unixMillis={data.createdAt} />}
                />
            )}
        </>
    );
```

**Existing webapp tests pass** - Test exists at `ApprovalDMPost.test.tsx` lines 340-363.

### Testing Strategy

**Unit Tests (Server):**
- No existing tests for `SendTimeoutNotificationDM` in dm_test.go
- Add comprehensive tests for Matterpoll pattern:
  - Test post.Type = "custom_approval_dm"
  - Test notification_type = "timeout" in props
  - Test created_at prop is populated
  - Test no buttons in attachment

**Webapp Tests:**
- Existing test at ApprovalDMPost.test.tsx:340-363 covers timeout rendering
- Verify test passes with new server implementation

**Manual Testing:**
1. Create approval: `/approve new @approver "test timeout"`
2. Wait for timeout (or mock timeout flow)
3. Verify requester receives DM with:
   - Custom component rendering (not markdown)
   - Status badge showing "timeout"
   - Approver info displayed
   - Requested timestamp in local timezone

## Dev Agent Record

### File List

| File | Change Type | Description |
|------|-------------|-------------|
| `server/notifications/dm.go` | Modified | Replaced `SendTimeoutNotificationDM()` implementation to use Matterpoll pattern with `CreateInteractiveApprovalPost()` helper |
| `server/notifications/dm_test.go` | Modified | Added 15 comprehensive unit tests for `TestSendTimeoutNotificationDM` covering AC validation, error handling, and prop verification |

### Change Log

- **2026-01-19**: Story 10.7 implementation complete
  - Replaced markdown-only timeout notification with Matterpoll pattern
  - Uses `CreateInteractiveApprovalPost()` with `NotificationTypeTimeout`
  - No interactive buttons rendered (read-only notification)
  - Webapp component already supports timeout type from Story 10.4
  - All tests pass (18 server tests + 4 webapp timeout tests)
- **2026-01-19**: Code review fixes applied
  - Added Dev Agent Record / File List section
  - Strengthened "no_buttons_in_attachment" test with proper assertions
  - Added test for approval_code prop
  - Added test for description prop
  - Added test for requester info props

### References

- [Source: server/notifications/dm.go - SendTimeoutNotificationDM lines 236-274]
- [Source: server/notifications/interactive_post.go - NotificationTypeTimeout line 25]
- [Source: server/notifications/interactive_post.go - FormatMarkdownFallback timeout case lines 309-319]
- [Source: webapp/src/components/ApprovalDMPost.tsx - timeout case lines 297-312]
- [Source: webapp/src/components/ApprovalDMPost.test.tsx - timeout test lines 340-363]
- [Source: epic-10-dm-interactive-buttons.md#Story 10.7]
- [Source: 10-6-convert-cancellation-notifications-to-matterpoll-pattern.md - Pattern reference]
