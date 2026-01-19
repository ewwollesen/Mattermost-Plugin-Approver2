# Story 9.8: Server-Side Post Type Updates for Playbook Posts

Status: done

## Story

As a server,
I want to create approval posts with custom post type for playbook channels,
so that webapp renders them as rich components.

## Acceptance Criteria

**AC1: Update playbooks/client.go**
- Modify `PostMessageToPlaybookChannel()` to accept approval record
- Set `post.Type = "custom_approval"`
- Populate `post.Props` with approval data (see AC3)
- Set `post.Message` with markdown table as fallback

**AC2: Update playbooks/formatters.go**
- Keep existing formatters for markdown fallback messages
- Create new function: `FormatApprovalPropsForWebapp(record)` → map[string]interface{}
- Include all required fields with proper types
- Ensure timestamps are int64 (Unix millis), not strings

**AC3: Approval Props Schema**
- Props must match webapp expectations from Story 9.7:
```go
Props: map[string]interface{}{
    "approval_code": record.Code,                              // string
    "approval_status": record.Status,                          // string
    "requester_username": record.RequesterUsername,            // string
    "requester_display_name": record.RequesterDisplayName,     // string
    "approver_username": record.ApproverUsername,              // string
    "approver_display_name": record.ApproverDisplayName,       // string
    "description": record.Description,                         // string
    "created_at": record.CreatedAt,                            // int64 (Unix millis)
    "decided_at": record.DecidedAt,                            // int64 (Unix millis)
    "decision_comment": record.DecisionComment,                // string
    "note": record.DecisionComment,                            // string (for approved)
}
```

**AC4: Update Call Sites**
- `server/api.go`: Update approval creation to use custom post type
- `server/api.go`: Update approve/deny handlers to use UpdatePost with new props
- `server/timeout/checker.go`: Update timeout handler to use custom post type
- All updates maintain backward compatibility (markdown fallback)

**AC5: UpdatePost for Status Changes**
- When approval status changes (approved, denied, canceled, timeout)
- Call `UpdateMessageInPlaybookChannel()` with updated record
- Update `post.Props` with new status, timestamps, comments
- Webapp component re-renders automatically with new data

**AC6: Validation and Error Handling**
- Validate all props before setting (non-nil, correct types)
- Log errors if props are invalid
- Fall back to markdown-only post if webapp props fail
- Ensure existing v2.x behavior preserved if custom post type fails

**AC7: Unit Tests**
- Test props population from approval record
- Test custom post type set correctly
- Test markdown fallback message generation
- Test UpdatePost with status changes

## Tasks / Subtasks

- [x] Create helper function FormatApprovalPropsForWebapp() (AC2, AC3)
  - [x] Add to playbooks/formatters.go
  - [x] Accept approval.ApprovalRecord as input
  - [x] Return map[string]any with all required props
  - [x] Ensure timestamps are int64 (Unix millis)
  - [x] Handle nil/empty fields gracefully
  - [x] Add function comment documenting field mapping

- [x] Update PostMessageToPlaybookChannel() function (AC1)
  - [x] Modify signature to accept *approval.ApprovalRecord
  - [x] Create post with Type = "custom_approval"
  - [x] Call FormatApprovalPropsForWebapp() to populate Props
  - [x] Keep existing markdown formatter for Message field
  - [x] Update all callers to pass approval record

- [x] Update UpdateMessageInPlaybookChannel() function (AC5)
  - [x] Modify to update both Message and Props
  - [x] Call FormatApprovalPropsForWebapp() with updated record
  - [x] Use UpdatePost API to update existing post
  - [x] Ensure Type remains "custom_approval"
  - [x] Log errors but don't fail if update fails

- [x] Update approval creation in api.go (AC4)
  - [x] Find handleCreateApproval() or similar
  - [x] Pass approval record to PostMessageToPlaybookChannel()
  - [x] Verify custom post type created
  - [x] Test with real playbook channel

- [x] Update approval decision handlers in api.go (AC4, AC5)
  - [x] Find handleApprove() and handleDeny()
  - [x] After updating record, call UpdateMessageInPlaybookChannel()
  - [x] Pass updated record with decision timestamp
  - [x] Verify post updates in playbook channel

- [x] Update timeout handler in timeout/checker.go (AC4, AC5)
  - [x] Find timeout detection logic
  - [x] Call UpdateMessageInPlaybookChannel() with timeout status
  - [x] Pass updated record with timeout timestamp
  - [x] Verify post updates correctly

- [x] Add validation and error handling (AC6)
  - [x] Validate approval record before creating props
  - [x] Check required fields: Code, Status, usernames
  - [x] Log warnings for missing optional fields
  - [x] Fall back to markdown-only if props fail
  - [x] Don't crash if webapp props can't be created

- [x] Create unit tests for new functions (AC7)
  - [x] Test FormatApprovalPropsForWebapp() with all statuses
  - [x] Test pending status (no decidedAt)
  - [x] Test approved status (with note)
  - [x] Test denied status (with reason)
  - [x] Test canceled status
  - [x] Test timeout status
  - [x] Verify timestamp types (int64, not string)
  - [x] Test error handling (nil record, missing fields)

- [x] Integration testing
  - [x] Create approval in playbook channel
  - [x] Verify post.Type = "custom_approval"
  - [x] Verify post.Props contains all fields
  - [x] Approve/deny approval
  - [x] Verify post updates (not new post)
  - [x] Check webapp renders correctly
  - [x] Test markdown fallback (disable webapp)

## Dev Notes

### Architecture Requirements

**Server-Webapp Integration Pattern:**
This story bridges server (Go) and webapp (React):
- **Story 9.7**: Webapp registered custom post type "custom_approval"
- **Story 9.8**: Server creates posts with Type="custom_approval" and populates Props ← YOU ARE HERE
- **Story 9.9**: End-to-end validation of the integration

**Data Flow:**
```
1. Server: Approval record created/updated
2. Server: Format props with FormatApprovalPropsForWebapp(record)
3. Server: Create/update post with Type="custom_approval" and Props
4. Server: Set Message with markdown table (fallback)
5. Mattermost: Deliver post to clients
6. Webapp: Detect Type="custom_approval"
7. Webapp: Render ApprovalPost component with post.props
8. Non-webapp: Render post.Message (markdown table)
```

**Backward Compatibility:**
- Old markdown-only behavior preserved if webapp not loaded
- post.Message always contains markdown table
- Mobile clients, API clients see markdown
- Webapp clients see rich React components

### Component Implementation Details

**playbooks/formatters.go - New Function:**

```go
// FormatApprovalPropsForWebapp formats approval record data for webapp custom post type
// Returns map suitable for post.Props that matches webapp ApprovalPost component expectations
// Story 9.8: Server-side support for custom approval posts
func FormatApprovalPropsForWebapp(record *approval.ApprovalRecord) map[string]interface{} {
    if record == nil {
        return make(map[string]interface{})
    }

    props := map[string]interface{}{
        "approval_code":           record.Code,
        "approval_status":         record.Status,
        "requester_username":      record.RequesterUsername,
        "requester_display_name":  record.RequesterDisplayName,
        "approver_username":       record.ApproverUsername,
        "approver_display_name":   record.ApproverDisplayName,
        "description":             record.Description,
        "created_at":              record.CreatedAt,        // int64 Unix millis
        "decided_at":              record.DecidedAt,        // int64 Unix millis
        "decision_comment":        record.DecisionComment,
        "note":                    record.DecisionComment,  // For approved posts
    }

    return props
}
```

**playbooks/client.go - Updated PostMessageToPlaybookChannel:**

```go
// PostMessageToPlaybookChannel posts a message to the playbook channel
// Story 9.8: Updated to support custom post type with webapp props
func (c *Client) PostMessageToPlaybookChannel(channelID string, record *approval.ApprovalRecord) (string, error) {
    // GitHub Issue #2: Check circuit breaker before API calls
    if err := c.circuitBreaker.Call(func() error {
        return nil // Just a check, actual call below
    }); err != nil {
        return "", fmt.Errorf("circuit breaker open: %w", err)
    }

    // Create custom post with webapp props
    post := &model.Post{
        ChannelId: channelID,
        UserId:    c.botUserID,
        Type:      "custom_approval",                      // Story 9.8: Custom post type
        Props:     FormatApprovalPropsForWebapp(record),   // Story 9.8: Webapp props
        Message:   FormatPendingStatusMessage(record),     // Markdown fallback
    }

    // Create post via API
    createdPost, appErr := c.api.CreatePost(post)
    if appErr != nil {
        c.metrics.RecordFailure()
        return "", fmt.Errorf("failed to create post: %w", appErr)
    }

    c.metrics.RecordSuccess()
    return createdPost.Id, nil
}
```

**playbooks/client.go - Updated UpdateMessageInPlaybookChannel:**

```go
// UpdateMessageInPlaybookChannel updates an existing playbook channel post
// Story 9.8: Updated to update both Message and Props for webapp support
func (c *Client) UpdateMessageInPlaybookChannel(channelID string, postID string, record *approval.ApprovalRecord) error {
    // GitHub Issue #2: Check circuit breaker
    if err := c.circuitBreaker.Call(func() error {
        return nil
    }); err != nil {
        return fmt.Errorf("circuit breaker open: %w", err)
    }

    // Get existing post
    post, appErr := c.api.GetPost(postID)
    if appErr != nil {
        c.metrics.RecordFailure()
        return fmt.Errorf("failed to get post: %w", appErr)
    }

    // Update post with new status
    post.Type = "custom_approval"                      // Ensure type is set
    post.Props = FormatApprovalPropsForWebapp(record)  // Update props
    post.Message = formatStatusMessage(record)         // Update markdown fallback

    // Update post via API
    _, appErr = c.api.UpdatePost(post)
    if appErr != nil {
        c.metrics.RecordFailure()
        return fmt.Errorf("failed to update post: %w", appErr)
    }

    c.metrics.RecordSuccess()
    return nil
}

// formatStatusMessage returns appropriate markdown based on status
func formatStatusMessage(record *approval.ApprovalRecord) string {
    switch record.Status {
    case "approved":
        return FormatApprovedStatusMessage(record)
    case "denied":
        return FormatDeniedStatusMessage(record)
    case "canceled":
        return FormatCanceledStatusMessage(record)
    case "timeout":
        return FormatTimedOutStatusMessage(record)
    default:
        return FormatPendingStatusMessage(record)
    }
}
```

**Key Implementation Notes:**

1. **Timestamp Types Critical:**
   - MUST be int64 (Unix milliseconds)
   - NOT formatted strings like "2026-01-18 10:30:00 UTC"
   - Webapp Timestamp component expects numeric timestamps

2. **Props Schema Matches Webapp:**
   - Field names use snake_case: `approval_code`, `created_at`, etc.
   - Webapp ApprovalPost extracts with same names
   - See Story 9.7 AC3 for webapp expectations

3. **Decision Comment vs Note:**
   - `decision_comment`: Generic field for denied/canceled reasons
   - `note`: Specific field for approved post comments
   - Server sets both to `record.DecisionComment` for compatibility

4. **Markdown Fallback Always Present:**
   - post.Message contains markdown table
   - Non-webapp clients, mobile apps see markdown
   - API tools, webhooks see markdown
   - Webapp ignores Message, uses Props

5. **UpdatePost Behavior:**
   - Updates existing post (same post ID)
   - v2.x behavior: post updated, not replaced
   - Webapp re-renders ApprovalPost component automatically
   - Props change triggers React re-render

### Library & Framework Requirements

**Dependencies Already Present:**
- github.com/mattermost/mattermost/server/public/model (Post types)
- github.com/mattermost/mattermost/server/public/plugin (API)
- github.com/mattermost/mattermost-plugin-approver2/server/approval (ApprovalRecord)

**No New Dependencies Required:**
All server-side code uses existing Mattermost Plugin API.

### File Structure Requirements

**Files to Modify:**
- `server/playbooks/formatters.go` - Add FormatApprovalPropsForWebapp()
- `server/playbooks/client.go` - Update PostMessageToPlaybookChannel(), UpdateMessageInPlaybookChannel()
- `server/api.go` - Update approval handlers to pass record to playbook client
- `server/timeout/checker.go` - Update timeout handler

**Files to Create:**
- `server/playbooks/formatters_test.go` - Unit tests for FormatApprovalPropsForWebapp() (if doesn't exist)
- Or add tests to existing test file

**Current Server Structure:**
```
server/
├── approval/
│   └── record.go                  # ApprovalRecord model
├── playbooks/
│   ├── client.go                  # MODIFY: PostMessageToPlaybookChannel, UpdateMessageInPlaybookChannel
│   ├── formatters.go              # MODIFY: Add FormatApprovalPropsForWebapp
│   ├── circuit_breaker.go         # Used by client (Story 8.6)
│   ├── metrics.go                 # Used by client (Story 8.6)
│   └── client_test.go             # ADD: Tests for new functions
├── api.go                         # MODIFY: Approval handlers
├── timeout/
│   └── checker.go                 # MODIFY: Timeout handler
└── manifest/
    └── plugin.json                # No changes (webapp already registered)
```

### Previous Story Intelligence (Story 9.7 Learnings)

**Critical Discoveries from Story 9.7:**

1. **Props Schema Defined:**
   - Story 9.7 AC3 specifies exact props schema
   - Field names: snake_case (approval_code, requester_username, etc.)
   - Timestamp fields: created_at, decided_at (int64)
   - Optional fields: decided_at, decision_comment, note
   - **For Story 9.8**: Server must match this schema exactly

2. **Custom Post Type ID:**
   - Post type: "custom_approval"
   - Webapp registered this in index.tsx
   - **For Story 9.8**: Server must use identical string

3. **Fallback Rendering:**
   - post.Message contains markdown table
   - Non-webapp clients display Message field
   - **For Story 9.8**: Keep existing markdown formatters

4. **ApprovalPost Component Expectations:**
   - Extracts data from post.props with defensive defaults
   - Handles missing fields gracefully
   - **For Story 9.8**: Server should populate all fields, but missing fields won't crash webapp

5. **Integration Test Plan:**
   - Story 9.7 AC5 requires creating test approval in playbook channel
   - Verify custom component renders
   - **For Story 9.8**: Server changes enable this test

### Git Intelligence Summary

**Recent Commits (Last 5):**

1. **bf000fe: Fix: GitHub Issue #2 - Replace Playbooks API with markdown tables**
   - Removed Playbooks API integration
   - Server posts markdown tables to playbook channels
   - Modified: server/playbooks/client.go, server/playbooks/formatters.go
   - **Relevance**: This story builds on this commit, adding custom post type to existing markdown posting logic

2. **53c03a3: Story 8.6: Error Handling and Graceful Fallback**
   - Circuit breaker pattern for Playbooks integration
   - Defensive coding with fallbacks
   - **Relevance**: Maintain circuit breaker checks, add fallback if webapp props fail

3. **48954c5: Story 8.2: Data Model Extension for Playbook Metadata**
   - Extended approval record with playbook metadata
   - **Relevance**: All approval record fields available for props population

**Key Patterns Identified:**
- Playbooks client uses circuit breaker before API calls
- Formatters generate markdown tables (now also webapp props)
- Server always has fallback behavior
- Timestamps stored as int64 (Unix millis) in ApprovalRecord

### Project Structure Context

**Approval Record Model (server/approval/record.go):**
```go
type ApprovalRecord struct {
    ID                    string  // Mattermost ID (26 chars)
    Code                  string  // A-XXXXXX format
    Status                string  // pending, approved, denied, canceled, timeout
    RequesterUserID       string
    RequesterUsername     string
    RequesterDisplayName  string
    ApproverUserID        string
    ApproverUsername      string
    ApproverDisplayName   string
    Description           string
    CreatedAt             int64   // Unix millis
    DecidedAt             int64   // Unix millis (0 if pending)
    DecisionComment       string  // Optional
    ChannelID             string
    PlaybookRunID         string  // From Story 8.2
    PlaybookID            string  // From Story 8.2
    // ... other fields
}
```

**All fields needed for webapp props are already in ApprovalRecord.**

### References

- [Source: Epic 9 - Story 9.8 Acceptance Criteria] - Server-side specifications
- [Source: Story 9.7 AC3] - Props schema expectations from webapp
- [Source: Story 9.6 Dev Notes] - ApprovalPost component props extraction
- [Source: server/playbooks/client.go] - Existing PostMessageToPlaybookChannel implementation
- [Source: server/playbooks/formatters.go] - Existing markdown formatters
- [Source: GitHub Issue #2 Resolution] - Markdown table pattern without Playbooks API
- [Source: Story 8.6 Dev Notes] - Circuit breaker pattern

### Critical Gotchas

**AVOID THESE MISTAKES:**

1. **Don't Use Wrong Timestamp Format:**
   - MUST be int64 (Unix milliseconds)
   - NOT strings like "2026-01-18T10:30:00Z"
   - ApprovalRecord.CreatedAt is already int64 ✅
   - **Impact**: Webapp Timestamp component will crash or show "Invalid Date"

2. **Don't Use Wrong Post Type String:**
   - MUST be "custom_approval" (matches Story 9.7)
   - NOT "approval", "custom-approval", etc.
   - **Impact**: Webapp won't detect post, falls back to markdown

3. **Don't Forget Markdown Fallback:**
   - post.Message MUST contain readable markdown
   - Mobile clients, API tools rely on this
   - **Impact**: Non-webapp clients see empty/broken posts

4. **Don't Use camelCase for Props:**
   - Webapp expects snake_case: `approval_code`, `created_at`
   - NOT camelCase: `approvalCode`, `createdAt`
   - Go structs use PascalCase internally, props use snake_case externally
   - **Impact**: Webapp extracts wrong fields, shows "UNKNOWN" defaults

5. **Don't Forget Circuit Breaker:**
   - Story 8.6 added circuit breaker to playbooks client
   - All API calls must check circuit breaker first
   - **Impact**: Repeated failures if circuit breaker ignored

6. **Don't Update Signature Without Updating Callers:**
   - PostMessageToPlaybookChannel signature changed (added record param)
   - Find all callers: api.go, timeout/checker.go, etc.
   - **Impact**: Compilation errors or wrong data passed

7. **Don't Skip Validation:**
   - Validate record before creating props
   - Check Code, Status, usernames not empty
   - **Impact**: Invalid props passed to webapp, component shows defaults

**Common Errors to Watch For:**
- "cannot use formatters.FormatApprovalPropsForWebapp(record) (value of type map[string]interface{}) as model.StringInterface": Use model.StringInterface type alias
- "post.Props["created_at"] is string, not int64": Timestamp formatting error
- "webapp shows 'UNKNOWN' for all fields": Props field names don't match expectations
- "UpdatePost overwrites Type": Ensure Type="custom_approval" set on update

**Testing Gotchas:**
- Can't test webapp rendering without deploying plugin
- Use browser DevTools to inspect post.Props
- Check post.Type === "custom_approval"
- Verify timestamps are numbers, not strings in JSON

### Implementation Order

**Recommended Implementation Sequence:**
1. Add FormatApprovalPropsForWebapp() to formatters.go
2. Write unit tests for FormatApprovalPropsForWebapp()
3. Update PostMessageToPlaybookChannel() in client.go
4. Update UpdateMessageInPlaybookChannel() in client.go
5. Find approval creation handler in api.go
6. Update handler to pass record to PostMessageToPlaybookChannel()
7. Find approve/deny handlers in api.go
8. Update handlers to call UpdateMessageInPlaybookChannel() after decision
9. Find timeout handler in timeout/checker.go
10. Update handler to call UpdateMessageInPlaybookChannel() on timeout
11. Compile and fix any build errors
12. Run unit tests
13. Deploy to test server
14. Create test approval in playbook channel
15. Verify post.Type and post.Props in browser DevTools
16. Test approval, denial, cancellation, timeout flows

**Why This Order:**
- FormatApprovalPropsForWebapp first: Core function needed by all others
- Unit tests early: Catch props formatting bugs before integration
- Client functions before handlers: Bottom-up dependency order
- Callers after functions: Avoid compilation errors
- Integration testing last: Requires deployed plugin

### Performance Considerations

**Server-Side Impact:**
- Props creation: ~1-2ms (map allocation + field copying)
- No database queries added
- No external API calls
- **Total overhead: Negligible**

**Network Impact:**
- post.Props adds ~500 bytes per post (JSON data)
- post.Message already ~800 bytes (markdown table)
- **Total post size: ~1.3KB (acceptable, below Mattermost limits)**

**Backward Compatibility:**
- Old clients ignore post.Props, use post.Message
- No breaking changes
- Graceful degradation

### Architecture Compliance

**Aligns with Epic 9 Decisions:**
- ✅ Custom Post Type for Playbook Posts (Decision 4)
- ✅ Store Timestamps as Unix Millis (Decision 6)
- ✅ No Backward Compatibility for Old Posts (Decision 7) - new posts use custom type

**Aligns with Project Structure:**
- ✅ Playbooks package handles playbook channel interactions
- ✅ Formatters provide consistent message formatting
- ✅ Circuit breaker protection (Story 8.6)
- ✅ Defensive coding with fallbacks

**Prepares for Story 9.9:**
Story 9.9 (End-to-End Testing) will:
1. Test full approval flow in playbook channel
2. Verify webapp renders ApprovalPost component
3. Verify timezone display accuracy
4. Verify status updates (pending → approved/denied)
5. Regression testing (v1.0, v2.x behavior preserved)

### Data Contract (Server → Webapp)

**Server Creates (Go):**
```go
post := &model.Post{
    Type:      "custom_approval",
    Message:   FormatPendingStatusMessage(record),  // Markdown fallback
    Props:     FormatApprovalPropsForWebapp(record),
}
```

**Webapp Receives (TypeScript):**
```typescript
interface Post {
    Type: "custom_approval";
    message: string;  // Markdown fallback (ignored by webapp)
    props: {
        approval_code: string;
        approval_status: string;
        // ... all other fields from FormatApprovalPropsForWebapp
    };
}
```

**Field Mapping (Go → JSON → TypeScript):**
```
Go:         record.Code             → JSON: "approval_code"    → TS: props.approval_code
Go:         record.Status           → JSON: "approval_status"  → TS: props.approval_status
Go:         record.CreatedAt (int64)→ JSON: "created_at" (num) → TS: props.created_at (number)
```

**Type Safety:**
- Go struct fields are strongly typed
- JSON preserves types (int64 → number, string → string)
- TypeScript receives correctly typed values
- **No type conversions needed** (just field name mapping)

### Wayne's Feedback Integration

**Critical User Requirements:**
1. **"Stick to Mattermost theme"** - Server sets Type, webapp handles rendering
2. **"Minimize screen real estate"** - Webapp component handles layout (Story 9.6)
3. **"No backward compatibility needed"** - New posts use custom type, old posts stay markdown
4. **Timezone issue (GitHub Issue #3)** - Server sends Unix millis, webapp converts to user timezone

**Design Philosophy:**
- Server provides data, webapp renders UI
- Separation of concerns: backend = data, frontend = presentation
- Fallback ensures no data loss for non-webapp clients

### Type Definitions

**Mattermost Post Type (Go):**
```go
// From github.com/mattermost/mattermost/server/public/model
type Post struct {
    Id              string
    CreateAt        int64
    UpdateAt        int64
    UserId          string
    ChannelId       string
    Message         string           // Markdown fallback
    Type            string           // "custom_approval"
    Props           model.StringInterface  // map[string]interface{}
    Hashtags        string
    FileIds         StringArray
    PendingPostId   string
    Metadata        *PostMetadata
}

type StringInterface map[string]interface{}
```

**ApprovalRecord Type (Go):**
```go
// From server/approval/record.go
type ApprovalRecord struct {
    ID                    string
    Code                  string
    Status                string
    RequesterUserID       string
    RequesterUsername     string
    RequesterDisplayName  string
    ApproverUserID        string
    ApproverUsername      string
    ApproverDisplayName   string
    Description           string
    CreatedAt             int64  // Unix milliseconds
    DecidedAt             int64  // Unix milliseconds
    DecisionComment       string
    ChannelID             string
    PlaybookRunID         string
    PlaybookID            string
    CanceledReason        string
    CanceledDetails       string
    CanceledByUsername    string
    // ... other fields
}
```

### DM vs Playbook Context (Future Story 9.10)

**Current Scope (Story 9.8):**
- Server uses custom post type for playbook channels only
- DM notifications remain markdown (unchanged from v2.x)
- Story 9.10 will extend custom post type to DMs

**Story 9.10 (Future):**
- Modify notifications/dm.go functions
- Add same custom post type + props pattern
- May add `is_dm: true` to props for context detection

**Not Implemented in This Story:**
DM notification updates deferred to Story 9.10. This story focuses on playbook channel posts only.

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

- Code Review Session: Adversarial review identified 3 CRITICAL, 2 MEDIUM, 1 LOW issues
- Test Compilation Errors: Fixed mock signature mismatches across 4 test files
- Build Validation: All 568 Go tests passing, all 59 webapp tests passing
- Integration Testing: User verified custom post type rendering in browser console

### Completion Notes List

1. **Created FormatApprovalPropsForWebapp()** - New function in server/playbooks/formatters.go (line 153) formats approval record data as map[string]any props for webapp consumption with snake_case field names and int64 timestamps

2. **Updated PostMessageToPlaybookChannel()** - Modified signature to accept *approval.ApprovalRecord (line 226), creates posts with Type="custom_approval", Props from FormatApprovalPropsForWebapp(), and Message as markdown fallback

3. **Updated UpdateMessageInPlaybookChannel()** - Modified signature to accept *approval.ApprovalRecord (line 327), updates both post.Props and post.Message, ensures Type remains "custom_approval"

4. **Added Validation & Error Handling** - Comprehensive nil checks and required field validation (lines 267-301, 375-407 in client.go) with graceful fallback to markdown-only posts

5. **Updated All Call Sites** - Modified 3 locations in server/api.go (lines 317, 647, 924) and 2 locations in server/timeout/checker.go (lines 173, 190) to pass approval record instead of formatted string

6. **Created Comprehensive Unit Tests** - Added 9 test cases in server/playbooks/formatters_test.go (lines 358-555) covering all statuses, nil handling, timestamp precision, and snake_case validation. All tests passing.

7. **Fixed Test Mock Signatures** - Updated mock implementations in client_test.go, api_test.go, timeout/checker_test.go, command/router_test.go to match new signatures

8. **Removed Obsolete Test** - Deleted TestFormatPendingPlaybookStatusMessage from api_test.go (functionality moved to playbooks package)

9. **Added StatusTimeout Constant** - Added approval.StatusTimeout = "timeout" to server/approval/models.go (line 69) to replace magic string

10. **Build & Deploy Success** - Plugin compiled successfully, bundle created at dist/com.mattermost.plugin-approver2-2.1.0.tar.gz

### File List

**Modified Files:**
- `server/approval/models.go` - Added StatusTimeout constant
- `server/playbooks/formatters.go` - Added FormatApprovalPropsForWebapp() function
- `server/playbooks/formatters_test.go` - Added 9 comprehensive unit tests for props formatter
- `server/playbooks/client.go` - Updated PostMessageToPlaybookChannel() and UpdateMessageInPlaybookChannel() signatures, added validation
- `server/playbooks/client_test.go` - Updated mock signatures and test cases
- `server/api.go` - Updated 3 call sites to pass approval record, removed formatPendingPlaybookStatusMessage()
- `server/api_test.go` - Updated mock signatures, removed obsolete test function
- `server/timeout/checker.go` - Updated 2 call sites to pass approval record
- `server/timeout/checker_test.go` - Updated mock signatures
- `server/command/router_test.go` - Updated mock signatures
- `server/playbooks/circuit_breaker.go` - No functional changes (formatting)
- `server/playbooks/circuit_breaker_test.go` - No functional changes (formatting)
- `server/playbooks/metrics.go` - No functional changes (formatting)
- `server/playbooks/metrics_test.go` - No functional changes (formatting)
- `plugin.json` - No functional changes
- `_bmad-output/implementation-artifacts/sprint-status.yaml` - Updated story status to done
- `.gitignore` - No functional changes

**New Files:**
- `webapp/` - Entire webapp directory (created in previous stories 9.1-9.7)
