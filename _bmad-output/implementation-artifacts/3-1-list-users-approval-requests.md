# Story 3.1: list-users-approval-requests

Status: done

## Story

As a user,
I want to list all my approval requests and approvals,
So that I can review my approval history and find specific records.

## Acceptance Criteria

**AC1: Basic List Command**

**Given** a user types `/approve list` in any Mattermost context
**When** the system processes the command
**Then** the system retrieves all approval records where the authenticated user is either the requester or the approver
**And** the retrieval completes within 3 seconds (NFR-P4)
**And** the results are returned as an ephemeral message (visible only to the user)

**AC2: Empty State Handling**

**Given** the user has no approval records
**When** they execute `/approve list`
**Then** the system returns a message: "No approval records found. Use `/approve new` to create a request."
**And** the response completes within 1 second

**AC3: Multiple Records Display**

**Given** the user has multiple approval records
**When** the system retrieves the records
**Then** the records are sorted by CreatedAt timestamp (most recent first)
**And** the list includes records where the user is the requester
**And** the list includes records where the user is the approver
**And** each record shows: Code, Status, Requester, Approver, CreatedAt (formatted as date)

**AC4: Pagination**

**Given** the list contains more than 20 records
**When** the system displays the results
**Then** the first 20 records are shown
**And** a message indicates: "Showing 20 of {total} total records (most recent first). Use `/approve get <ID>` to view specific requests."
**And** the total count is displayed to inform the user
**And** the output remains readable and not truncated

**AC5: Structured Output Format**

**Given** the results are being formatted
**When** the system constructs the output
**Then** the output uses a structured table or list format:
```
**Your Approval Records:**

**A-X7K9Q2** | ✅ Approved | Requested by: @alex | Approver: @jordan | 2026-01-10 02:14
**A-Y3M5P1** | ⏳ Pending | Requested by: @alex | Approver: @morgan | 2026-01-09 14:23
**A-Z8N2K7** | ❌ Denied | Requested by: @chris | Approver: @alex | 2026-01-08 09:45
```
**And** each record is on a separate line for readability
**And** status is indicated with clear icons (✅ Approved, ❌ Denied, ⏳ Pending, 🚫 Canceled)

**AC6: Error Handling**

**Given** the KV store query fails
**When** the system attempts to retrieve records
**Then** an error message is returned: "❌ Failed to retrieve approval records. Please try again."
**And** the error is logged with context

**AC7: Access Control**

**Given** the authenticated user session is valid
**When** the list command is executed
**Then** the system uses the authenticated user's ID to filter records (NFR-S1)
**And** no records from other users are included (NFR-S2)

**Covers:** FR18 (retrieve own approval requests), FR19 (retrieve where user was approver), FR21 (records contain all context), FR23 (records include status), FR37 (access control), NFR-P4 (<3s retrieval), NFR-S1 (authenticated), NFR-S2 (users see only their own records)

## Tasks / Subtasks

### Task 1: Extend KV Store with GetUserApprovals method (AC: 1, 3, 7)
- [x] Add `GetUserApprovals(userID string) ([]*approval.ApprovalRecord, error)` method to KVStore in server/store/kvstore.go
- [x] Implement KVList query to find all records where user is requester OR approver
- [x] Use existing KVList pagination (MaxApprovalRecordsLimit = 10000)
- [x] Filter records to include only those matching userID in RequesterID or ApproverID fields
- [x] Sort results by CreatedAt timestamp descending (most recent first)
- [x] Return empty slice (not nil) when no records found
- [x] Log warning if hitting MaxApprovalRecordsLimit
- [x] Handle KV store errors with proper error wrapping

### Task 2: Update Storer interface in command/router.go (AC: 1)
- [x] Add `GetUserApprovals(userID string) ([]*approval.ApprovalRecord, error)` to Storer interface in server/command/router.go
- [x] This enables dependency injection for testing

### Task 3: Implement executeList handler in command/router.go (AC: 1-7)
- [x] Add `executeList(args *model.CommandArgs) (*model.CommandResponse, error)` method to Router
- [x] Extract authenticated user ID from args.UserId
- [x] Call store.GetUserApprovals(args.UserId) to retrieve records
- [x] Handle empty state (AC2): return friendly message with suggestion to use `/approve new`
- [x] Handle KV store errors (AC6): return error message and log with context
- [x] Apply pagination limit of 20 records (AC4) from sorted results
- [x] Format results using formatListResponse helper (see Task 4)
- [x] Return CommandResponse with ephemeral type

### Task 4: Create formatListResponse helper function (AC: 3, 4, 5)
- [x] Add `formatListResponse(records []*approval.ApprovalRecord, total int) string` function
- [x] Build structured markdown output with header "**Your Approval Records:**"
- [x] For each record, format as: `**{Code}** | {StatusIcon} {Status} | Requested by: @{RequesterUsername} | Approver: @{ApproverUsername} | {FormattedDate}`
- [x] Use status icons: ✅ Approved, ❌ Denied, ⏳ Pending, 🚫 Canceled
- [x] Format timestamps as "YYYY-MM-DD HH:MM" (convert epoch millis to time.Time, format to UTC)
- [x] Add pagination footer if total > 20: "Showing 20 most recent records. Use `/approve get <ID>` for specific requests."
- [x] Ensure output is clean, readable, and follows UX specification

### Task 5: Wire up list command in Route method (AC: 1)
- [x] Add case "list" to switch statement in Route() method
- [x] Call `return r.executeList(args)` for list command
- [x] Update executeHelp() if list is not already documented (it is, verify)
- [x] Update executeUnknown() to include "list" in valid commands message (verify it's there)

### Task 6: Write comprehensive tests for list functionality (AC: ALL)
- [x] Create TestExecuteList in server/command/router_test.go
- [x] Test empty state: user with no records gets friendly message
- [x] Test single record: verify correct formatting and status icons
- [x] Test multiple records: verify sorting (most recent first)
- [x] Test pagination: create >20 records, verify only first 20 shown with pagination footer
- [x] Test access control: verify only records for authenticated user are returned
- [x] Test requester records: user sees records where they are requester
- [x] Test approver records: user sees records where they are approver
- [x] Test mixed records: user sees both requester and approver records
- [x] Test KV store error handling: mock store error, verify error message returned
- [x] Test all status types: pending, approved, denied, canceled with correct icons
- [x] Test timestamp formatting: verify date format is correct
- [x] Update mockStore if needed to support GetUserApprovals

### Task 7: Write unit tests for KVStore.GetUserApprovals (AC: 1, 3, 7)
- [x] Create TestGetUserApprovals in server/store/kvstore_test.go
- [x] Test retrieval when user is requester only
- [x] Test retrieval when user is approver only
- [x] Test retrieval when user is both requester and approver in different records
- [x] Test sorting by CreatedAt descending
- [x] Test empty result when user has no records
- [x] Test filtering ensures other users' records are not included
- [x] Test error handling when KVList fails

## Dev Notes

### Architecture Alignment

**From Architecture Decision Document:**

- **Data Model**: ApprovalRecord struct with all fields in server/approval/models.go (already implemented)
- **KV Store Pattern**: Use hierarchical key structure `approval:record:{id}` for records
- **Error Handling**: Use error wrapping with `fmt.Errorf("context: %w", err)` pattern
- **Graceful Degradation**: If KV store fails, return user-friendly error message (not technical stack trace)
- **Performance**: Must complete within 3 seconds (NFR-P4), use existing MaxApprovalRecordsLimit (10000) for safety
- **Access Control**: Filter by authenticated user ID (args.UserId), enforce NFR-S1 and NFR-S2

**KV Store Key Structure (from Architecture):**
```
approval:record:{id}  → ApprovalRecord JSON (already implemented)
```

**Testing Standards:**
- Use standard Go testing with testify/assert
- Table-driven tests for multiple scenarios
- Mock the Plugin API for command handler tests
- Integration tests for KV store operations

### Previous Story Intelligence

**From Story 2.6 (handle-notification-delivery-failures):**

**Key Learnings:**
1. **Command Router Pattern**: All slash commands go through server/command/router.go Route() method with switch/case routing
2. **Storer Interface**: Used for dependency injection and testing - add methods to interface first, then implement in KVStore
3. **Error Logging**: Use r.api.LogError() with structured fields (key-value pairs) for debugging
4. **KVStore Pattern**: Existing GetAllApprovals() method shows how to:
   - Use s.api.KVList(0, MaxApprovalRecordsLimit) to list keys
   - Filter by prefix "approval:record:"
   - Handle partial failures gracefully (log warning, continue with other records)
   - Return empty slice when no records found (not error)
5. **Testing Mock**: mockStore in server/command/router_test.go implements Storer interface for testing
6. **Markdown Formatting**: Use bold labels (`**Label:**`), structured output with clear sections
7. **Status Icons**: Story 2.6 used ✅/❌ for success/failure - continue this pattern for status indicators

**Files Modified in Story 2.6:**
- server/command/router.go: Added executeStatus method with admin check and statistics calculation
- server/store/kvstore.go: Added GetAllApprovals() method with KVList and filtering
- server/command/router_test.go: Added tests with mockStore extension

**Code Patterns to Follow:**
```go
// Error logging pattern from Story 2.6
r.api.LogError("Failed to retrieve approval records",
    "user_id", args.UserId,
    "error", err.Error())

// KV Store query pattern from Story 2.6
keys, appErr := s.api.KVList(0, MaxApprovalRecordsLimit)
if appErr != nil {
    return nil, fmt.Errorf("failed to list approval records: %w", appErr)
}

// Response format pattern
return &model.CommandResponse{
    ResponseType: model.CommandResponseTypeEphemeral,
    Text:         formattedOutput,
}, nil
```

### Git Intelligence from Recent Commits

**Recent Commit Patterns:**
- Commit format: "Story X.Y: Brief description of change"
- Code review fixes applied as separate commits after initial implementation
- Testing files follow pattern: {file}_test.go alongside {file}.go
- All tests must pass before story completion

**Established Codebase Patterns:**
1. Package structure: server/{feature}/{file}.go
2. Command handlers in server/command/router.go
3. Data access in server/store/kvstore.go
4. Data models in server/approval/models.go
5. Testing with testify/assert library
6. Mock generation for interfaces
7. Table-driven tests for multiple scenarios

### Implementation Sequence

**RED-GREEN-REFACTOR Approach:**

1. **RED Phase - Write Failing Tests First:**
   - Task 6: Write comprehensive tests in router_test.go (will fail)
   - Task 7: Write unit tests for KVStore (will fail)

2. **GREEN Phase - Minimal Implementation:**
   - Task 1: Implement GetUserApprovals in kvstore.go
   - Task 2: Update Storer interface
   - Task 3: Implement executeList handler
   - Task 4: Create formatListResponse helper
   - Task 5: Wire up in Route method
   - Run tests - all should pass

3. **REFACTOR Phase:**
   - Review code for clarity and simplicity
   - Extract any repeated logic into helpers
   - Ensure error messages are user-friendly
   - Verify markdown formatting renders correctly
   - Check performance with large record counts

### Critical Implementation Details

**Status Icon Mapping:**
```go
func getStatusIcon(status string) string {
    switch status {
    case "approved":
        return "✅ Approved"
    case "denied":
        return "❌ Denied"
    case "pending":
        return "⏳ Pending"
    case "canceled":
        return "🚫 Canceled"
    default:
        return status // Fallback
    }
}
```

**Timestamp Formatting:**
```go
// Convert epoch millis to formatted date
createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
```

**Sorting Pattern:**
```go
// Sort by CreatedAt descending (most recent first)
sort.Slice(records, func(i, j int) bool {
    return records[i].CreatedAt > records[j].CreatedAt
})
```

**Pagination Logic:**
```go
displayLimit := 20
total := len(records)
if total > displayLimit {
    records = records[:displayLimit]
}
// Add footer if paginated
if total > displayLimit {
    output += fmt.Sprintf("\n\nShowing %d most recent records. Use `/approve get <ID>` for specific requests.", displayLimit)
}
```

### Security Considerations

**Access Control (NFR-S1, NFR-S2):**
- CRITICAL: Only show records where user is requester OR approver
- User ID comes from args.UserId (authenticated session)
- Filter in KVStore.GetUserApprovals, not in command handler (defense in depth)
- Never log sensitive data (descriptions) in error messages
- Use structured logging with user_id, record count, not individual records

**Error Message Safety:**
- Don't leak information about other users' records
- Generic error messages for failures ("Failed to retrieve records")
- Specific guidance only for user's own issues ("You have no approval records")

### Performance Considerations

**NFR-P4 Compliance (<3 seconds):**
- KVList is bounded by MaxApprovalRecordsLimit (10000)
- Filtering and sorting happen in-memory (fast for <10k records)
- Pagination limits output size for large result sets
- No external API calls or network requests

**Optimization Notes:**
- Consider caching if performance becomes issue (post-MVP)
- Current approach is simple and sufficient for MVP scale
- Warn in logs if approaching MaxApprovalRecordsLimit

### Testing Strategy

**Unit Tests (server/store/kvstore_test.go):**
- Test GetUserApprovals with various scenarios
- Mock Plugin API KVList responses
- Verify filtering logic is correct
- Test sorting behavior
- Test error handling

**Integration Tests (server/command/router_test.go):**
- End-to-end command handling
- Mock store with predefined data
- Verify response formatting
- Test all acceptance criteria
- Table-driven tests for different scenarios

**Test Data Patterns:**
```go
// Create test records with predictable data
record1 := &approval.ApprovalRecord{
    ID: "record1",
    Code: "A-ABC123",
    RequesterID: "user1",
    RequesterUsername: "alice",
    ApproverID: "user2",
    ApproverUsername: "bob",
    Status: "pending",
    CreatedAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
}
```

### UX Specification Compliance

**From UX Design Document:**
- Structured Markdown with bold labels
- Status indicators with icons (visual clarity)
- User mentions with @username format
- Explicit timestamps (not relative: "2 hours ago")
- Ephemeral messages (only visible to user)
- Clear actionable guidance ("Use `/approve new`")
- Mobile-friendly formatting (no wide tables)

**Output Format Example:**
```
**Your Approval Records:**

**A-X7K9Q2** | ✅ Approved | Requested by: @alice | Approver: @bob | 2026-01-10 14:23
**A-Y8D4K1** | ⏳ Pending | Requested by: @alice | Approver: @charlie | 2026-01-09 09:15
**A-Z3M9P5** | ❌ Denied | Requested by: @bob | Approver: @alice | 2026-01-08 16:42

Showing 3 records.
```

### Change Log

**Code Review Architectural Enhancement (commit 93581b0):**

During Story 3.1 code review, a critical performance issue was discovered and fixed:

**BEFORE:**
- GetUserApprovals used O(n) full-system scan
- KVList returned all records, then filtered in memory
- Performance degrades with total system record count

**AFTER (Story 3.4 Index Strategy Implemented):**
- SaveApproval creates 4 keys instead of 2:
  - `approval:record:{id}` (primary record)
  - `approval:code:{code}` (code lookup) - **format changed from approval_code:**
  - `approval:index:requester:{userID}:{invertedTimestamp}:{recordID}` (NEW)
  - `approval:index:approver:{userID}:{invertedTimestamp}:{recordID}` (NEW)
- GetUserApprovals uses O(k) prefix queries on index keys
- Inverted timestamp enables natural descending order
- Performance independent of total system records

**Impact:** Story 3.4 "Implement Index Strategy for Fast Retrieval" was completed during this code review.

**Test Updates (commit 20e5eac):**
- Updated all test files to use new key format (approval:code: instead of approval_code:)
- Updated mocks to include approval:index: key patterns
- All 19 tests passing

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic 3: Story 3.1]
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 3: Story 3.4] - **Implemented during Story 3.1 code review**
- [Source: _bmad-output/planning-artifacts/architecture.md#1.3 Approval Record Data Model]
- [Source: _bmad-output/planning-artifacts/architecture.md#1.5 Index Strategy] - **Fully implemented**
- [Source: _bmad-output/planning-artifacts/architecture.md#2.2 Graceful Degradation Strategy]
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#Message Formatting Patterns]
- [Source: _bmad-output/implementation-artifacts/2-6-handle-notification-delivery-failures.md#Tasks]

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-5-20250929

### Debug Log References

**Architectural Decision: Index Strategy Implementation**

During code review (commit 93581b0), a critical performance issue was identified: GetUserApprovals was using O(n) full-system scans. The fix implemented Story 3.4's complete index strategy:

1. **Problem:** Original implementation scanned all records and filtered in memory
2. **Solution:** Added requester/approver index keys with inverted timestamps
3. **Impact:** Performance improved from O(n) to O(k) where k = user's record count
4. **Side Effect:** Story 3.4 "Implement Index Strategy for Fast Retrieval" was completed during Story 3.1 code review

**Key Format Standardization**

Changed from `approval_code:` to `approval:code:` for consistency:
- All keys now use colon separator uniformly
- Updated in SaveApproval, GetByCode, and all tests
- Rationale: Hierarchical namespace clarity

### Completion Notes

**⚠️ IMPORTANT: This story implemented Story 3.4's index strategy during code review**

**Tasks 3, 4, 5, 6 Complete:**
- Implemented executeList() handler in server/command/router.go
- Handles empty state with friendly message
- Implements pagination (20 records max with footer)
- Logs errors with structured context
- Created formatListResponse() helper with markdown formatting
- Status icons: ✅ Approved, ❌ Denied, ⏳ Pending, 🚫 Canceled
- Timestamp format: YYYY-MM-DD HH:MM (UTC)
- Added getStatusIcon() helper function
- Wired up "list" case in Route() method
- Added 11 comprehensive integration tests (all pass)
- Updated mockStore to support GetUserApprovals

**Task 2 Complete:**
- Added GetUserApprovals to Storer interface in server/command/router.go

**Tasks 1 & 7 Complete:**
- Implemented GetUserApprovals() method in server/store/kvstore.go
- **CODE REVIEW ENHANCEMENT:** Changed from O(n) full scan to O(k) index-based queries
- **Implemented Story 3.4 index strategy:** Added requester/approver index key creation in SaveApproval
- Method retrieves all records where user is requester OR approver using prefix queries
- Filtering implemented at data access layer (defense in depth)
- Sorts by CreatedAt descending using inverted timestamp keys
- Returns empty slice (not nil) for no results
- Handles partial failures gracefully with LogWarn
- Logs warning if hitting MaxApprovalRecordsLimit (10,000)
- Added 8 comprehensive unit tests covering all scenarios
- All tests pass (8/8)

### File List

**Primary Implementation Files:**
- server/store/kvstore.go (modified - added GetUserApprovals method with index queries, makeRequesterIndexKey, makeApproverIndexKey functions, updated SaveApproval to create 4 keys instead of 2, **235 lines added**)
- server/store/kvstore_test.go (modified - added TestKVStore_GetUserApprovals with 8 test scenarios testing index-based queries, 233 lines)
- server/command/router.go (modified - added GetUserApprovals to Storer interface, executeList handler, formatListResponse helper, getStatusIcon helper, wired up "list" case, **291 lines added**)
- server/command/router_test.go (modified - updated mockStore, added TestExecuteList with 11 test scenarios, 424 lines added)

**Supporting Files Updated for Index Strategy:**
- server/api_test.go (modified - updated KVGet/KVSet mocks to include approval:index: key patterns)
- server/approval/codegen.go (modified - changed key format from approval_code: to approval:code: for consistency)
- server/plugin_test.go (modified - updated test mocks for new key format and index keys)
