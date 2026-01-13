# Story 4.6: Show Canceled Requests in List (Sorted to Bottom)

Status: done

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Priority:** Medium
**Estimate:** 3 points
**Assignee:** Dev Agent

## Story

As a **user (requester or approver)**,
I want **canceled requests to appear at the bottom of my list, sorted separately from active and decided requests**,
so that **I can focus on pending and decided approvals first, while still having visibility into canceled requests**.

## Context

Currently, `/approve list` displays all approval records in a single chronological list, regardless of status. This creates poor UX because canceled requests (which are no longer actionable) appear mixed with pending approvals (which require action) and decided approvals (which provide recent outcomes).

**Current Behavior (router.go:422-451):**
```
**Your Approval Records:**

**A-X7K9Q2** | 🚫 | Requested by: @wayne | Approver: @john | 2026-01-12 19:15
**A-ABC123** | 🕐 | Requested by: @wayne | Approver: @sarah | 2026-01-12 18:00
**A-DEF456** | ✅ | Requested by: @alice | Approver: @wayne | 2026-01-12 17:30
**A-GHI789** | 🚫 | Requested by: @bob | Approver: @wayne | 2026-01-12 16:00
**A-JKL012** | ❌ | Requested by: @wayne | Approver: @mike | 2026-01-12 15:00
```

**Problems:**
1. **Canceled requests interspersed** - Breaks visual scanning for actionable items
2. **No grouping by status** - Users can't quickly identify pending approvals
3. **Cancelled requests lack context** - No reason shown in list view (detail view shows it per Story 4.5)
4. **Poor prioritization** - Most important items (pending) don't stand out

**Required Change:**
Implement three-group sorting strategy with status-based sections:

```
**Your Approval Records:**

**Pending Approvals:**
**A-ABC123** | 🕐 | Requested by: @wayne | Approver: @sarah | 2026-01-12 18:00

**Decided Approvals:**
**A-DEF456** | ✅ | Requested by: @alice | Approver: @wayne | 2026-01-12 17:30
**A-JKL012** | ❌ | Requested by: @wayne | Approver: @mike | 2026-01-12 15:00

**Canceled Requests:**
**A-X7K9Q2** | 🚫 Canceled (No longer needed) | Requested by: @wayne | Approver: @john | 2026-01-12 19:15
**A-GHI789** | 🚫 Canceled (Changed requirements) | Requested by: @bob | Approver: @wayne | 2026-01-12 16:00
```

## Acceptance Criteria

- [x] AC1: List displays three distinct sections with headers: "Pending Approvals", "Decided Approvals", "Canceled Requests"
- [x] AC2: Pending section includes only records with Status == "pending", sorted by CreatedAt descending (newest first)
- [x] AC3: Decided section includes records with Status == "approved" OR "denied", sorted by DecidedAt descending (newest first)
- [x] AC4: Canceled section includes only records with Status == "canceled", sorted by CanceledAt descending (newest first, fallback to CreatedAt if CanceledAt == 0)
- [x] AC5: Canceled requests display abbreviated cancellation reason in list view: "🚫 Canceled ({reason})"
- [x] AC6: Cancellation reasons longer than 40 characters are truncated to 37 chars + "..." (e.g., "No longer needed - project was canceled by stakeholder" → "No longer needed - project was cancel...")
- [x] AC7: Old canceled records without CanceledReason show "🚫 Canceled" (no reason text)
- [x] AC8: Empty sections are omitted entirely (no header if no records in that section)
- [x] AC9: Pagination footer shows total count across all sections: "Showing 20 of 45 total records"
- [x] AC10: Each section respects the 20-record limit proportionally (if 50 total records, show mix from each section up to 20 total)

## Tasks / Subtasks

- [x] Task 1: Implement three-group sorting logic (AC: 1, 2, 3, 4, 8)
  - [x] Subtask 1.1: Create helper function `groupAndSortRecords` that takes slice of ApprovalRecord and returns three slices: pending, decided, canceled
  - [x] Subtask 1.2: Sort pending group by CreatedAt descending
  - [x] Subtask 1.3: Sort decided group (approved + denied) by DecidedAt descending
  - [x] Subtask 1.4: Sort canceled group by CanceledAt descending (fallback to CreatedAt if CanceledAt == 0)
  - [x] Subtask 1.5: Apply 20-record limit across all groups combined (take first 20 from concatenated result)

- [x] Task 2: Update formatListResponse to display grouped sections (AC: 1, 5, 6, 7, 8)
  - [x] Subtask 2.1: Add section header "**Pending Approvals:**\n" if pending slice non-empty
  - [x] Subtask 2.2: Format pending records with existing format (no changes)
  - [x] Subtask 2.3: Add section header "**Decided Approvals:**\n" if decided slice non-empty
  - [x] Subtask 2.4: Format decided records with existing format (no changes)
  - [x] Subtask 2.5: Add section header "**Canceled Requests:**\n" if canceled slice non-empty
  - [x] Subtask 2.6: Format canceled records with new format: "🚫 Canceled ({reason})" instead of just "🚫"
  - [x] Subtask 2.7: Implement reason abbreviation: truncate to 37 chars + "..." if len > 40
  - [x] Subtask 2.8: Handle empty CanceledReason: display "🚫 Canceled" (no reason text)

- [x] Task 3: Update pagination footer logic (AC: 9, 10)
  - [x] Subtask 3.1: Count total records across all groups before limiting to 20
  - [x] Subtask 3.2: Update footer message to reflect total count: "Showing {displayed} of {total} total records"
  - [x] Subtask 3.3: If total <= 20, omit footer (all records shown)

- [x] Task 4: Write comprehensive tests (AC: all)
  - [x] Subtask 4.1: Test groupAndSortRecords with mixed statuses (pending, approved, denied, canceled)
  - [x] Subtask 4.2: Test sorting within each group (verify timestamp order)
  - [x] Subtask 4.3: Test canceled records with CanceledAt == 0 (should fall back to CreatedAt)
  - [x] Subtask 4.4: Test reason truncation (40 char boundary, 37 + "...")
  - [x] Subtask 4.5: Test empty CanceledReason (should show "🚫 Canceled" without reason text)
  - [x] Subtask 4.6: Test empty sections (should omit headers)
  - [x] Subtask 4.7: Test pagination with > 20 records across groups
  - [x] Subtask 4.8: Test list with all 4 cancellation reasons ("No longer needed", "Changed requirements", "Created by mistake", "Other: custom text")

## Dev Notes

### Architecture Compliance

**Display Pattern (from existing code analysis):**
- File: `server/command/router.go`
- Function: `formatListResponse` (lines 422-451)
- Pattern: strings.Builder with fmt.Sprintf
- Format: Markdown with bold labels (**Label:**)
- Timestamps: "2006-01-02 15:04" (time.Unix().UTC().Format())

**Current List Logic:**
```go
func formatListResponse(records []*approval.ApprovalRecord, total int) string {
    var output strings.Builder
    output.WriteString("**Your Approval Records:**\n\n")

    for _, record := range records {
        statusIcon := getStatusIcon(record.Status)
        createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
        formattedDate := createdTime.UTC().Format("2006-01-02 15:04")

        output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
            record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
    }

    // Pagination footer
    if total > 20 {
        output.WriteString(fmt.Sprintf("\n\n*Showing 20 of %d total records (most recent first).* ...", total))
    }

    return output.String()
}
```

**Required Changes:**
1. Add `groupAndSortRecords` helper function before `formatListResponse`
2. Modify `formatListResponse` to call grouping function and render sections
3. Update canceled record formatting to include reason
4. Preserve existing format for pending/decided records

### Current Implementation Context

**File to Modify:**
- `server/command/router.go` - `formatListResponse` function (lines 422-451)
  - Currently: Single chronological list, no grouping
  - Change: Three-section grouped list with status-based sorting

**Data Already Available (Stories 4.4 and 4.5):**
- `record.Status` - string: "pending" | "approved" | "denied" | "canceled"
- `record.CanceledReason` - string, may be empty for old records
- `record.CanceledAt` - int64 epoch millis, may be 0 for old records
- `record.DecidedAt` - int64 epoch millis, may be 0 for pending
- `record.CreatedAt` - int64 epoch millis, always set

**Story Dependencies (COMPLETED):**
- Story 3.1: List command implementation ✅
- Story 4.4: CanceledReason and CanceledAt fields added ✅
- Story 4.5: Detail view shows cancellation info ✅

**No Breaking Changes:**
- Existing `/approve list` command unchanged (same slash command)
- Same pagination limit (20 records)
- Backwards compatible with old canceled records
- No API changes, no database schema changes

### Implementation Specification

```go
// In server/command/router.go, add helper function before formatListResponse

// groupAndSortRecords separates records into three groups (pending, decided, canceled)
// and sorts each group by appropriate timestamp descending (newest first)
func groupAndSortRecords(records []*approval.ApprovalRecord) (pending, decided, canceled []*approval.ApprovalRecord) {
    // Separate into groups
    for _, record := range records {
        switch record.Status {
        case approval.StatusPending:
            pending = append(pending, record)
        case approval.StatusApproved, approval.StatusDenied:
            decided = append(decided, record)
        case approval.StatusCanceled:
            canceled = append(canceled, record)
        }
    }

    // Sort pending by CreatedAt descending
    sort.Slice(pending, func(i, j int) bool {
        return pending[i].CreatedAt > pending[j].CreatedAt
    })

    // Sort decided by DecidedAt descending
    sort.Slice(decided, func(i, j int) bool {
        return decided[i].DecidedAt > decided[j].DecidedAt
    })

    // Sort canceled by CanceledAt descending (fallback to CreatedAt if CanceledAt == 0)
    sort.Slice(canceled, func(i, j int) bool {
        iTime := canceled[i].CanceledAt
        if iTime == 0 {
            iTime = canceled[i].CreatedAt
        }
        jTime := canceled[j].CanceledAt
        if jTime == 0 {
            jTime = canceled[j].CreatedAt
        }
        return iTime > jTime
    })

    return pending, decided, canceled
}

// formatListResponse renders approval records in three grouped sections
func formatListResponse(records []*approval.ApprovalRecord, total int) string {
    var output strings.Builder
    output.WriteString("**Your Approval Records:**\n\n")

    // Group and sort records
    pending, decided, canceled := groupAndSortRecords(records)

    // Limit to 20 total records across all groups
    displayed := 0
    limit := 20

    // Render Pending section
    if len(pending) > 0 {
        output.WriteString("**Pending Approvals:**\n")
        for _, record := range pending {
            if displayed >= limit {
                break
            }
            statusIcon := getStatusIcon(record.Status)
            createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
            formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
            output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
                record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
            displayed++
        }
        output.WriteString("\n")
    }

    // Render Decided section
    if len(decided) > 0 && displayed < limit {
        output.WriteString("**Decided Approvals:**\n")
        for _, record := range decided {
            if displayed >= limit {
                break
            }
            statusIcon := getStatusIcon(record.Status)
            createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
            formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
            output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
                record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
            displayed++
        }
        output.WriteString("\n")
    }

    // Render Canceled section (with reason)
    if len(canceled) > 0 && displayed < limit {
        output.WriteString("**Canceled Requests:**\n")
        for _, record := range canceled {
            if displayed >= limit {
                break
            }

            // Format canceled status with reason
            var statusText string
            if record.CanceledReason != "" {
                reason := record.CanceledReason
                // Truncate if longer than 40 chars
                if len(reason) > 40 {
                    reason = reason[:37] + "..."
                }
                statusText = fmt.Sprintf("🚫 Canceled (%s)", reason)
            } else {
                statusText = "🚫 Canceled"
            }

            createdTime := time.Unix(0, record.CreatedAt*int64(time.Millisecond))
            formattedDate := createdTime.UTC().Format("2006-01-02 15:04")
            output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
                record.Code, statusText, record.RequesterUsername, record.ApproverUsername, formattedDate))
            displayed++
        }
        output.WriteString("\n")
    }

    // Pagination footer (if total > displayed)
    if total > displayed {
        output.WriteString(fmt.Sprintf("*Showing %d of %d total records.* Use `/approve get <ID>` to view specific requests.", displayed, total))
    }

    return output.String()
}
```

### Testing Requirements

**Unit Tests (router_test.go):**
1. **Test groupAndSortRecords:**
   - Input: Mixed records (pending, approved, denied, canceled)
   - Verify: Three separate slices with correct grouping
   - Verify: Pending sorted by CreatedAt desc
   - Verify: Decided sorted by DecidedAt desc
   - Verify: Canceled sorted by CanceledAt desc (or CreatedAt fallback)

2. **Test formatListResponse with three sections:**
   - Input: Records from all three groups
   - Verify: Section headers appear in correct order
   - Verify: Pending section first, Decided second, Canceled third
   - Verify: Empty sections omit headers

3. **Test canceled reason display:**
   - Input: Canceled record with reason "No longer needed"
   - Verify: Displays "🚫 Canceled (No longer needed)"
   - Input: Canceled record with 50-char reason
   - Verify: Truncates to 37 + "..."
   - Input: Canceled record with empty reason
   - Verify: Displays "🚫 Canceled" (no parentheses)

4. **Test pagination with grouped records:**
   - Input: 30 total records (10 pending, 10 decided, 10 canceled)
   - Verify: Displays 20 total (prioritizes pending, then decided, then canceled)
   - Verify: Footer shows "Showing 20 of 30 total records"

5. **Test backwards compatibility:**
   - Input: Old canceled record (CanceledAt == 0, CanceledReason == "")
   - Verify: Sorts by CreatedAt (fallback)
   - Verify: Displays "🚫 Canceled" (no reason)

**Expected Output Examples:**

*Mixed list with all sections:*
```
**Your Approval Records:**

**Pending Approvals:**
**A-ABC123** | 🕐 | Requested by: @wayne | Approver: @sarah | 2026-01-12 18:00
**A-DEF456** | 🕐 | Requested by: @alice | Approver: @wayne | 2026-01-12 17:00

**Decided Approvals:**
**A-GHI789** | ✅ | Requested by: @wayne | Approver: @john | 2026-01-12 16:30
**A-JKL012** | ❌ | Requested by: @bob | Approver: @wayne | 2026-01-12 15:00

**Canceled Requests:**
**A-MNO345** | 🚫 Canceled (No longer needed) | Requested by: @wayne | Approver: @mike | 2026-01-12 14:00
**A-PQR678** | 🚫 Canceled (Changed requirements) | Requested by: @sarah | Approver: @wayne | 2026-01-12 13:00
```

*Only pending requests (other sections omitted):*
```
**Your Approval Records:**

**Pending Approvals:**
**A-ABC123** | 🕐 | Requested by: @wayne | Approver: @sarah | 2026-01-12 18:00
```

*Canceled record with long reason (truncated):*
```
**A-XYZ999** | 🚫 Canceled (No longer needed - project was ca...) | Requested by: @wayne | Approver: @john | 2026-01-12 12:00
```

*Old canceled record (no reason):*
```
**A-OLD123** | 🚫 Canceled | Requested by: @alice | Approver: @bob | 2026-01-10 10:00
```

### Edge Cases

1. **Empty List**
   - All three groups empty
   - Display: "**Your Approval Records:**\n\n" (no section headers)
   - Handle gracefully, no crashes

2. **Only Canceled Records**
   - Pending and Decided groups empty
   - Display: Only "**Canceled Requests:**" section
   - No headers for empty sections

3. **Pagination with Imbalanced Groups**
   - Example: 1 pending, 1 decided, 50 canceled
   - Limit: 20 total records
   - Result: 1 pending + 1 decided + 18 canceled displayed
   - Footer: "Showing 20 of 52 total records"

4. **Old Canceled Records (before Story 4.4)**
   - CanceledReason = "" (empty string)
   - CanceledAt = 0
   - Sorting: Falls back to CreatedAt
   - Display: "🚫 Canceled" (no reason text, no parentheses)

5. **Cancellation Reason Exactly 40 Characters**
   - Example: "No longer needed - stakeholder change"
   - Length: 40 chars
   - Should NOT truncate (truncate only if > 40)
   - Display: "🚫 Canceled (No longer needed - stakeholder change)"

6. **Cancellation Reason 41+ Characters**
   - Example: "No longer needed - project was canceled by stakeholder team"
   - Length: 60 chars
   - Truncate to 37 + "..." = 40 chars total
   - Display: "🚫 Canceled (No longer needed - project was ca...)"

7. **"Other" Reason with Long Custom Text**
   - Example: "Other: Detailed explanation of why this approval is no longer relevant because requirements changed"
   - Should truncate to 37 + "..."
   - Display: "🚫 Canceled (Other: Detailed explanation of why...)"

8. **All Records from Same Group**
   - Example: 25 pending records, 0 decided, 0 canceled
   - Display: Only "**Pending Approvals:**" section with 20 records
   - Footer: "Showing 20 of 25 total records"

9. **Multiple Canceled Records with Same Timestamp**
   - CanceledAt identical (unlikely but possible)
   - Sorting: Go's sort.Slice is stable, maintains original order for ties
   - Display: Same timestamp order preserved

10. **Decided Records with Same DecidedAt**
    - Multiple approvals processed simultaneously
    - Sorting: Stable sort maintains original order
    - Display: Preserves creation order for ties

### Performance Considerations

- **Sorting Complexity**: O(n log n) for each group, acceptable for typical volumes
- **String Truncation**: O(1) operation per record
- **Memory**: Three slices created temporarily (GC collects after response sent)
- **Target**: Sub-3 second retrieval (architecture requirement NFR-P4)
- **Typical Volume**: 20-100 records per user, well within performance budget

### References

- [Source: server/command/router.go:422-451] - formatListResponse implementation
- [Source: server/approval/models.go:33-36] - Status constants and CanceledReason/CanceledAt fields
- [Source: _bmad-output/planning-artifacts/architecture.md:1149-1177] - Requirements coverage and performance targets
- [Source: _bmad-output/planning-artifacts/epics.md] - Epic 4 Story 6 requirements
- [Source: Story 4.4] - CanceledReason and CanceledAt data model additions
- [Source: Story 4.5] - Detail view cancellation display pattern

### Project Structure Notes

**Alignment with Architecture:**
- Uses established display pattern (strings.Builder, fmt.Sprintf, Markdown formatting)
- Follows Go naming conventions (CamelCase: `groupAndSortRecords`, `statusText`)
- Co-located tests in `router_test.go` (Mattermost pattern)
- No new packages created (modification to existing `server/command/router.go`)
- Timestamps use established format: "2006-01-02 15:04"
- Error handling not required (display logic, no external calls)

**No Conflicts Detected:**
- Change is isolated to list display logic
- No API contract changes (same `/approve list` command)
- No data model changes (uses existing fields)
- No performance concerns (sorting 20-100 records is trivial)
- Backwards compatible with old canceled records

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

N/A - Story file created via create-story workflow

### Completion Notes List

**Implementation Completed - Story 4.6:**

1. **Three-Group Sorting Logic** (Task 1) - Implemented `groupAndSortRecords` helper function in router.go:421-460
   - Separates records into pending, decided, and canceled slices
   - Sorts pending by CreatedAt descending (newest first)
   - Sorts decided (approved + denied) by DecidedAt descending
   - Sorts canceled by CanceledAt descending with CreatedAt fallback for old records
   - All sorting tests passing (6 test scenarios)

2. **Grouped Section Display** (Task 2) - Updated `formatListResponse` function in router.go:462-545
   - Added section headers: "Pending Approvals:", "Decided Approvals:", "Canceled Requests:"
   - Omits empty sections (AC8)
   - Canceled records display reason: "🚫 Canceled ({reason})"
   - Truncates reasons > 40 chars to exactly 37 + "..."
   - Handles empty CanceledReason: displays "🚫 Canceled" without parentheses
   - All display tests passing (10 test scenarios)

3. **Pagination Logic** (Task 3) - Updated pagination to work across all groups
   - 20-record limit applies to total displayed (not per-section)
   - Footer shows: "Showing {displayed} of {total} total records"
   - Omits footer when all records shown
   - Removed "(most recent first)" text since grouping makes ordering explicit

4. **Comprehensive Test Coverage** (Task 4) - Added 361 lines of tests in router_test.go:1614-1967
   - 6 tests for groupAndSortRecords function
   - 10 tests for formatListResponse grouped sections
   - Covers all 10 acceptance criteria
   - Tests all 4 cancellation reasons
   - Tests backwards compatibility (old records without CanceledAt/CanceledReason)
   - Tests pagination, truncation, empty sections
   - All 16 new tests passing + existing tests maintained (no regressions)

5. **Code Quality** - Modernized loop syntax (Go 1.22+ range over int)
   - Fixed 1 existing test that expected old pagination message
   - Added `sort` package import
   - All code follows Mattermost Go style guide

**Test Results:**
- Total new tests: 16 test functions with 30+ test cases
- All command package tests: PASS (including 333+ existing tests)
- All server package tests: PASS
- No regressions introduced

**Key Implementation Details:**
- Used Go 1.22+ `range over int` syntax for modern loop patterns
- Exact 37-character truncation preserves readability ("cancel" not "ca")
- Stable sort maintains order for records with identical timestamps
- Proportional pagination fills pending first, then decided, then canceled up to 20 total

### File List

**Files Modified (Story 4.6):**
- `server/command/router.go` (lines 3-7, 421-545, 643-685)
  - Added `sort` package import
  - Added `groupAndSortRecords` helper function (lines 421-460) - separates records into Pending/Decided/Canceled groups and sorts each by appropriate timestamp
  - Completely rewrote `formatListResponse` function (lines 462-545) - implements grouped sections with headers, cancellation reason display with UTF-8-safe truncation, proportional pagination
  - **Code Review Fix:** Changed truncation from byte slicing to rune slicing for proper UTF-8 handling (lines 522-524)
  - **Code Review Fix:** Re-added Story 4.5 cancellation display code in `formatRecordDetail` (lines 643-685) that was accidentally reverted

- `server/command/router_test.go` (lines 701, 1614-2021)
  - Fixed existing pagination test (line 701) to match new footer message format
  - Added 16 comprehensive tests (361 lines total):
    - 6 tests for `groupAndSortRecords` function covering all grouping/sorting scenarios
    - 10 tests for `formatListResponse` covering display, truncation, pagination, and edge cases
  - **Code Review Fix:** Added `TestFormatListResponse_UTF8Handling` with 6 test cases for emoji, CJK, and accented characters (lines 1968-2021)
  - Modernized loop syntax using Go 1.22+ `range over int`

- `server/api_test.go` (line 1029)
  - **Code Review Fix:** Ran gofmt to fix pre-existing formatting issue from Story 4.3

- `_bmad-output/implementation-artifacts/4-6-show-canceled-requests-in-list-sorted-to-bottom.md`
  - Marked all 10 acceptance criteria complete [x]
  - Marked all 4 tasks and 23 subtasks complete [x]
  - Updated Completion Notes with detailed implementation summary
  - Updated File List section (this section)
  - Status updated to "review"

- `_bmad-output/implementation-artifacts/sprint-status.yaml` (line 80)
  - Updated status: ready-for-dev → in-progress → review

**No New Files Created:**
All changes are modifications to existing files

**No Breaking Changes:**
- Same `/approve list` command
- Same 20-record pagination limit
- Backwards compatible with old canceled records (CanceledAt==0 falls back to CreatedAt, empty CanceledReason displays without parentheses)
- No API or data model changes

**Uncommitted Dependencies (Code Review Finding):**
Story 4.6 was implemented on top of uncommitted changes from Stories 4.1-4.5. The following files show uncommitted changes from previous stories:
- `server/api.go` - Story 4.3 cancellation modal handling
- `server/plugin.go` - Story 4.3 cancel command updates
- `server/plugin_test.go` - Various story tests
- `server/notifications/dm_test.go` - Test formatting updates
- Story documentation files (4-3, 4-5) - Completion notes

These files are not part of Story 4.6's scope but are required dependencies. All previous stories (4.1-4.5) show status "done" in sprint-status.yaml but have never been committed to git.

### Change Log

**2026-01-13: Story 4.6 Implementation Complete**
- Implemented three-group sorting strategy (Pending → Decided → Canceled)
- Each group sorted by appropriate timestamp with descending order
- Added cancellation reason display in list view with 37-char truncation for long reasons
- Updated pagination to work proportionally across all three groups (20 total records)
- Added comprehensive test coverage: 16 new tests (361 lines) covering all grouping, sorting, display, and edge cases
- All tests passing with no regressions
- Code follows existing patterns and quality standards
- Backwards compatible with old records missing CanceledAt/CanceledReason fields
- Ready for code review

**2026-01-13: Code Review Fixes Applied**
- Fixed UTF-8 truncation bug: Changed from byte slicing `reason[:37]` to rune slicing `[]rune(reason)[:37]` to properly handle emojis, accented characters, and CJK text
- Added comprehensive UTF-8 test coverage: 6 new test cases covering emoji, CJK, and accented character handling
- Added Story 4.5 cancellation display code to formatRecordDetail (was accidentally reverted during code review)
- Fixed pre-existing formatting issue in api_test.go (Story 4.3 code)
- All 357 tests passing
- Build successful (make dist completed)
- Issues found and fixed: 2 HIGH, 3 MEDIUM, 1 LOW
