# Story 7.5: Convert /approve list Output to Markdown Table

**Epic:** 7 - 1.0 Polish & UX Improvements
**Story ID:** 7.5
**Priority:** MEDIUM (Polish)
**Status:** review
**Created:** 2026-01-15
**Started:** 2026-01-15
**Code Review:** 2026-01-15 (HIGH issues fixed)
**Awaiting:** Manual testing on desktop/mobile clients

---

## User Story

**As a** user
**I want** `/approve list` output formatted as a markdown table
**So that** the output looks clean and professional

---

## Story Context

This story converts the `/approve list` command output from plain text formatting to professional markdown tables. The current implementation (from Epic 3, Epic 5) uses structured plain text with bold labels, icons, and pipe separators, but doesn't use proper markdown table syntax. This creates a less polished appearance compared to established Mattermost plugins like Jira and GitHub.

**Current Behavior (Plain Text with Formatting):**
```
## Your Approval Requests (5 pending)

**Pending Approvals:**
**A-X7K9Q2** | ⏳ Pending | Requested by: @wayne | Approver: @john | 2026-01-15 14:30
**A-B3M4N1** | ⏳ Pending | Requested by: @wayne | Approver: @sarah | 2026-01-15 13:45
...
```

**Desired Behavior (Markdown Table):**
```
## Your Approval Requests (5 pending)

**Pending Approvals:**
| Code | Status | Requestor | Approver | Created |
|------|--------|-----------|----------|---------|
| A-X7K9Q2 | ⏳ Pending | @wayne | @john | 2026-01-15 14:30 |
| A-B3M4N1 | ⏳ Pending | @wayne | @sarah | 2026-01-15 13:45 |
...
```

**Why This Matters:**
- **Professional Polish:** Tables look cleaner and more organized than plain text
- **Readability:** Columnar layout easier to scan, especially with many records
- **Consistency:** Matches quality bar of established Mattermost plugins (Jira, GitHub)
- **Mobile Friendly:** Mattermost markdown tables are responsive by default
- **Fast Win:** Format conversion only, no business logic changes

**Impact:**
- LOW technical risk (rendering change only)
- HIGH visual impact (immediate professional appearance)
- ZERO functional changes (all filtering, sorting, pagination preserved)

---

## Acceptance Criteria

### AC1: List Output Uses Markdown Table Format
**Given** a user runs `/approve list` with any filter (pending, approved, denied, canceled, all)
**When** the command returns results
**Then** the output renders as properly formatted markdown table
**And** table has appropriate columns: Code, Status, Requestor, Approver, Created
**And** table uses Mattermost markdown syntax: `| col1 | col2 |` with header separator `|-----|-----|`

### AC2: Table Formatting is Clean and Professional
**Given** markdown table output
**When** rendered in Mattermost client
**Then** columns are aligned and easy to read
**And** status icons (✅ ❌ ⏳ 🚫) render correctly in table cells
**And** usernames with @ mentions render as mentions
**And** dates are formatted consistently (YYYY-MM-DD HH:MM format)
**And** table header uses bold formatting

### AC3: Grouping and Sections Preserved
**Given** `/approve list all` returns mixed status records
**When** formatted as markdown
**Then** records are grouped into sections: Pending Approvals, Decided Approvals, Canceled Requests
**And** each section has a bold label above its table
**And** canceled requests appear at bottom (Epic 4.6 sort order)
**And** each section uses its own markdown table

### AC4: Filtering Logic Preserved
**Given** existing filter functionality from Epic 5
**When** running any filter (pending, approved, denied, canceled, all)
**Then** only matching records appear in table
**And** header shows correct count: "Your Approval Requests (N pending/approved/etc.)"
**And** empty states show filter-specific message: "No pending approval requests. Use `/approve list all` to see all requests."
**And** default filter is "pending" (Epic 5.2 behavior)

### AC5: Pagination and Limits Maintained
**Given** more than 20 records match filter
**When** rendering markdown table
**Then** only first 20 records displayed
**And** footer shows: "*Showing 20 of 50 total records.* Use `/approve get <ID>` to view specific requests."
**And** pagination applies within grouped sections (pending displayed first, then decided, then canceled)

### AC6: Canceled Reason Display Preserved
**Given** canceled requests with cancellation reason from Epic 4
**When** rendered in markdown table
**Then** Status column shows: "🚫 Canceled (reason)" for requests with reason
**And** reason is truncated if longer than 40 characters with "..." suffix
**And** reason-less cancellations show: "🚫 Canceled"

### AC7: Cross-Client Rendering Works
**Given** markdown table output
**When** viewed on different Mattermost clients
**Then** table renders correctly on web app
**And** table renders correctly on desktop app
**And** table is responsive and readable on mobile app
**And** table doesn't break layout or require horizontal scrolling on narrow screens

---

## Technical Requirements

### Implementation Location
**File:** `server/command/router.go`
**Function:** `formatListResponse()` (line 592-677)

### Current Implementation Analysis

**Current Format Pattern:**
```go
output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
    record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
```

**Required Change:**
Convert to proper markdown table syntax with header row and alignment separators.

### Markdown Table Syntax

**Mattermost Markdown Table Format:**
```markdown
| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |
```

**Key Requirements:**
- Header row with column names enclosed in `| ... |`
- Separator row with dashes: `|----------|----------|`
- Data rows with pipe-separated values
- Alignment: Left-align all columns (default)
- Spaces around pipes for readability (optional but improves source)

### Grouping Sections

**Epic 4.6 Sort Order Requirement:**
When filter is "all", records are grouped into sections with this priority:
1. **Pending Approvals** - Actionable items first
2. **Decided Approvals** - Approved/Denied combined
3. **Canceled Requests** - Historical items last

Each section should have:
- Bold section header (e.g., `**Pending Approvals:**`)
- Markdown table for that section's records
- Blank line between sections

### Column Design

**Recommended Columns:**
| Column | Content | Example | Notes |
|--------|---------|---------|-------|
| Code | Approval code | A-X7K9Q2 | Bold formatting optional |
| Status | Icon + label | ⏳ Pending | Includes emoji |
| Requestor | Username with @ | @wayne | Mention syntax |
| Approver | Username with @ | @john | Mention syntax |
| Created | Timestamp | 2026-01-15 14:30 | ISO format, UTC |

**Alternative: Combine Requestor/Approver for space:**
- Could use "Requestor → Approver" pattern to save horizontal space on mobile
- Example: `@wayne → @john`
- **Recommendation:** Keep separate columns for clarity (more readable)

---

## Architecture Compliance

### Mattermost Plugin API Patterns (AD 3.1)
✅ **Chat-Native Output:** Uses Mattermost markdown rendering (no custom UI)
✅ **Standard Markdown:** Follows CommonMark spec + Mattermost extensions
✅ **No External Dependencies:** Pure Go string building
✅ **Responsive:** Markdown tables adapt to client viewport

### Code Organization (AD 3.5)
✅ **Single Responsibility:** formatListResponse() handles formatting only
✅ **Pure Functions:** No side effects, returns formatted string
✅ **Existing Patterns:** Reuses groupAndSortRecords(), getStatusIcon() helpers
✅ **Testability:** Easy to unit test output string

### Performance Requirements (NFR-P1)
✅ **Sub-5 Second Response:** String building is fast (microseconds for 20 records)
✅ **No Network Calls:** Client-side markdown rendering
✅ **Efficient:** No regex or complex parsing, just string concatenation

### UX Consistency (NFR-U1)
✅ **Professional Quality:** Matches Jira/GitHub plugin table formatting
✅ **Clear Hierarchy:** Section headers + tables for visual grouping
✅ **Mobile Responsive:** Mattermost handles table rendering on small screens

---

## Implementation Plan

### Task Breakdown

#### Task 1: Update formatListResponse() to Generate Markdown Tables
**AC:** AC1, AC2
**File:** server/command/router.go (line 592-677)

**Subtasks:**
- [ ] Add table header function: `func formatTableHeader() string`
  - Returns: `"| Code | Status | Requestor | Approver | Created |\n|------|--------|-----------|----------|---------|"`
- [ ] Update pending section rendering to use table format
  - Keep bold section header: `**Pending Approvals:**\n`
  - Add table header row after section header
  - Convert each record line from pipe-separated text to table row: `| A-X7K9Q2 | ⏳ Pending | @wayne | @john | 2026-01-15 14:30 |`
- [ ] Update decided section rendering to use table format
- [ ] Update canceled section rendering to use table format
- [ ] Preserve blank line spacing between sections

#### Task 2: Preserve Grouping and Sorting Logic
**AC:** AC3, AC4
**File:** server/command/router.go

**Subtasks:**
- [ ] Verify groupAndSortRecords() logic unchanged (line 715-744)
- [ ] Verify filterRecordsByStatus() logic unchanged (line 467-490)
- [ ] Verify sortRecordsByTimestamp() logic unchanged (line 529-570)
- [ ] Test all filter types produce correct table groupings

#### Task 3: Maintain Canceled Reason Display
**AC:** AC6
**File:** server/command/router.go (line 641-669)

**Subtasks:**
- [ ] Keep canceled reason truncation logic: `if len(runes) > 40 { reason = string(runes[:37]) + "..." }`
- [ ] Ensure Status column format: `"🚫 Canceled (reason)"` or `"🚫 Canceled"`
- [ ] Test reason display in table cell (verify no pipe escaping issues)

#### Task 4: Test Cross-Client Rendering
**AC:** AC7

**Subtasks:**
- [ ] Build plugin: `make dist`
- [ ] Deploy to test Mattermost instance
- [ ] Test `/approve list` with various filters on web app
- [ ] Test on desktop app (verify table rendering)
- [ ] Test on mobile app (verify responsiveness)
- [ ] Verify tables don't require horizontal scrolling
- [ ] Test with long usernames, long reasons (verify truncation/wrapping)

#### Task 5: Unit Test Updates
**File:** server/command/router_test.go

**Subtasks:**
- [ ] Update TestFormatListResponse test cases to expect markdown table format
- [ ] Verify test checks for `| Code | Status |` header pattern
- [ ] Verify test checks for `|------|--------|` separator pattern
- [ ] Add test case for empty sections (only headers, no rows)
- [ ] Run tests: `go test ./server/command -v`

---

## Testing Strategy

### Unit Tests

**Test Coverage Existing:**
- `TestFilterRecordsByStatus` (Epic 5.1) - Verify filtering unchanged
- `TestSortRecordsByTimestamp` (Epic 5.1) - Verify sorting unchanged
- `TestFormatListResponse` (Epic 3.3) - **UPDATE FOR MARKDOWN TABLES**

**New Test Cases Required:**
```go
func TestFormatListResponse_MarkdownTable(t *testing.T) {
    // Test: Markdown table header present
    // Test: Table rows formatted correctly
    // Test: Section headers before each table
    // Test: Blank lines between sections
}

func TestFormatListResponse_CanceledReasonInTable(t *testing.T) {
    // Test: Canceled reason appears in Status column
    // Test: Long reason truncates to 40 chars + "..."
    // Test: Empty reason shows "🚫 Canceled"
}
```

### Manual Testing Checklist

**Basic Table Rendering:**
- [ ] `/approve list` - Verify pending table renders
- [ ] `/approve list approved` - Verify approved table renders
- [ ] `/approve list denied` - Verify denied table renders
- [ ] `/approve list canceled` - Verify canceled table renders
- [ ] `/approve list all` - Verify all three sections (pending, decided, canceled) render

**Edge Cases:**
- [ ] 0 records - Verify empty message, no table
- [ ] 1 record - Verify table header + single row
- [ ] 20+ records - Verify pagination footer, limit to 20 rows
- [ ] Long username (20+ chars) - Verify no table breaking
- [ ] Long canceled reason (60 chars) - Verify truncation to 40 chars + "..."

**Cross-Client Testing:**
- [ ] Mattermost Web App - Verify table alignment, readability
- [ ] Mattermost Desktop App - Verify identical rendering
- [ ] Mattermost Mobile App (iOS/Android) - Verify responsive layout, no horizontal scroll

**Filtering and Sorting:**
- [ ] Verify "pending" default filter still works (Epic 5.2)
- [ ] Verify canceled records sort to bottom for "all" filter (Epic 4.6)
- [ ] Verify chronological sorting within each section
- [ ] Verify header count matches filtered records

### Integration Testing

**No new integration tests required** - this is pure formatting change. Existing command tests validate functionality.

**Validation Steps:**
- [ ] Run full test suite: `make test`
- [ ] Verify 0 test failures
- [ ] Build succeeds: `make build`
- [ ] Lint passes: `make lint`

---

## Dev Notes

### Current Code Structure

**formatListResponse() Function Flow:**
1. Build header with count: `## Your Approval Requests (N filter)`
2. Group records: `pending, decided, canceled := groupAndSortRecords(records)`
3. Render Pending section (if any)
4. Render Decided section (if any)
5. Render Canceled section (if any)
6. Add pagination footer (if total > displayed)

**No Changes Required To:**
- `executeList()` - Command routing (line 392-465)
- `filterRecordsByStatus()` - Filtering logic (line 467-490)
- `sortRecordsByTimestamp()` - Sorting logic (line 529-570)
- `groupAndSortRecords()` - Grouping logic (line 715-744)
- `getStatusIcon()` - Status icon mapping (line 679-692)

**Change Required Only To:**
- `formatListResponse()` - String formatting (line 592-677)
  - Replace `fmt.Sprintf("**%s** | %s | ...")` with markdown table rows
  - Add table header before each section
  - Maintain section labels and blank line spacing

### Implementation Pattern

**Before (Current Plain Text):**
```go
output.WriteString("**Pending Approvals:**\n")
for _, record := range pending {
    output.WriteString(fmt.Sprintf("**%s** | %s | Requested by: @%s | Approver: @%s | %s\n",
        record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
}
output.WriteString("\n")
```

**After (Markdown Table):**
```go
output.WriteString("**Pending Approvals:**\n")
output.WriteString("| Code | Status | Requestor | Approver | Created |\n")
output.WriteString("|------|--------|-----------|----------|----------|\n")
for _, record := range pending {
    output.WriteString(fmt.Sprintf("| %s | %s | @%s | @%s | %s |\n",
        record.Code, statusIcon, record.RequesterUsername, record.ApproverUsername, formattedDate))
}
output.WriteString("\n")
```

**Helper Function (Optional):**
```go
func formatTableHeader() string {
    return "| Code | Status | Requestor | Approver | Created |\n|------|--------|-----------|----------|----------|\n"
}
```

### Markdown Rendering Notes

**Mattermost Markdown Support:**
- Tables: Full support (CommonMark + GFM tables)
- Mentions: `@username` renders as user mention
- Emojis: Unicode emojis render directly (✅ ❌ ⏳ 🚫)
- Bold: `**text**` works in table cells
- Line breaks: Use `\n` for rows, blank `\n\n` between sections

**Potential Issues:**
- Pipe character `|` in content: Mattermost escapes automatically in table context
- Very long usernames: Markdown tables auto-wrap text within cells
- Mobile viewport: Mattermost uses horizontal scroll for wide tables (acceptable)

### Previous Story Learnings

**From Story 7.4 (Autocomplete):**
- Small, focused changes are fast and low-risk
- Build + test early and often
- Manual testing on live Mattermost required for client-side features
- Documentation updates important (README if needed)

**From Epic 5 (Filtering):**
- Preserve existing behavior religiously (filterRecordsByStatus, sorting)
- Test all filter combinations (pending, approved, denied, canceled, all)
- Dynamic headers with counts improve UX
- Empty states need filter-specific messaging

**From Epic 4.6 (Canceled Sorting):**
- Canceled requests sort to bottom for "all" view (lower priority)
- Canceled reason display is important (users need context)
- Truncate long reasons to prevent UI clutter

---

## References

- **Epic 7:** [Source: _bmad-output/implementation-artifacts/epic-7-polish-ux-improvements.md]
  - Story 7.5 requirements and acceptance criteria
  - Polish and UX improvements context
  - Professional quality bar goals

- **Story 5.1:** [Source: _bmad-output/implementation-artifacts/5-1-add-filter-parameter-to-list-command.md]
  - Filter implementation: pending, approved, denied, canceled, all
  - filterRecordsByStatus() logic
  - sortRecordsByTimestamp() logic

- **Story 5.2:** [Source: _bmad-output/implementation-artifacts/5-2-change-default-behavior-add-count-display.md]
  - Default filter changed to "pending"
  - Dynamic header with count: "Your Approval Requests (N filter)"
  - Filter-specific empty state messages

- **Story 4.6:** [Source: _bmad-output/implementation-artifacts/4-6-show-canceled-requests-in-list-sorted-to-bottom.md]
  - Canceled requests sort to bottom for "all" view
  - groupAndSortRecords() implementation
  - Section grouping: Pending → Decided → Canceled

- **Story 3.3:** [Source: _bmad-output/implementation-artifacts/3-3-display-approval-records-in-readable-format.md]
  - Original formatListResponse() implementation
  - Plain text formatting with pipe separators
  - Status icons: ✅ ❌ ⏳ 🚫

- **Architecture:** [Source: _bmad-output/planning-artifacts/architecture.md]
  - AD 3.1: Mattermost Plugin API patterns
  - AD 3.5: Code organization and modularity
  - NFR-P1: Performance requirements (sub-5 second)
  - NFR-U1: UX consistency requirements

- **Mattermost Documentation:**
  - Markdown Formatting: https://docs.mattermost.com/collaborate/format-messages.html
  - Plugin Best Practices: https://developers.mattermost.com/integrate/plugins/best-practices/
  - GitHub-Flavored Markdown Tables: https://github.github.com/gfm/#tables-extension-

---

## Definition of Done

- [x] formatListResponse() generates markdown table format
- [x] Table header row present: `| Code | Status | Requestor | Approver | Created |`
- [x] Table separator row present: `|------|--------|-----------|----------|----------|`
- [x] Each section (Pending, Decided, Canceled) has its own table
- [x] Section headers preserved: `**Pending Approvals:**`, `**Decided Approvals:**`, `**Canceled Requests:**`
- [x] Grouping and sorting logic unchanged (groupAndSortRecords, filterRecordsByStatus, sortRecordsByTimestamp)
- [x] Filtering works for all filter types (pending, approved, denied, canceled, all)
- [x] Pagination preserved (20 record limit, footer for more)
- [x] Canceled reason display preserved in Status column
- [x] Unit tests updated for markdown table format
- [x] All tests passing: `make test`
- [x] Plugin builds successfully: `make build`
- [x] Manual testing on web app completed (tables render correctly)
- [ ] Manual testing on desktop app - PENDING user testing
- [ ] Manual testing on mobile app (responsiveness) - PENDING user testing
- [x] Documentation updated (README.md)
- [x] Code review completed (HIGH issues fixed)
- [ ] Story marked as done in sprint-status.yaml - BLOCKED by pending manual tests

---

## Success Criteria

**Primary Metric:** `/approve list` output renders as professional markdown table indistinguishable from Jira/GitHub plugins.

**Validation:**
1. Run `/approve list` → See markdown table with clear columns
2. Run `/approve list all` → See three grouped tables (Pending, Decided, Canceled)
3. View on mobile → Table is responsive and readable
4. Compare with Jira plugin list output → Similar professional quality
5. Test with 20+ records → Pagination footer appears
6. Test with canceled reason → Reason appears in Status column

**Before this feature:**
- List output uses plain text with pipe separators (functional but less polished)
- Scanning multiple records requires careful reading of each line

**After this feature:**
- List output uses clean markdown tables (professional appearance)
- Columnar layout makes scanning easy and fast
- Matches quality bar of established Mattermost plugins

---

## Notes

- Fifth and final story in Epic 7 (7.1, 7.2, 7.3, 7.4 completed, 7.5 in progress)
- MEDIUM priority - polish improvement, not a blocker
- Zero business logic changes - pure formatting conversion
- Fast implementation (estimated 30-60 minutes for code + tests)
- High visual impact for minimal effort
- No breaking changes or migrations needed
- Mobile testing important (verify responsive table rendering)

---

**Story Status:** ready-for-dev
**Ready for Development:** 2026-01-15
**Estimated Effort:** Small (formatting change only, ~50 lines)
**Risk Level:** Low (rendering change, no business logic impact)
**Testing Dependency:** Requires manual testing on live Mattermost instance

---

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### File List

**Modified Files:**
- `server/command/router.go` (lines 392-491, 635-703)
  - **CRITICAL FIX:** Added blank line (`\n\n`) after section headers (lines 635, 654, 673) - Required for Mattermost markdown parser to recognize tables
  - Updated formatListResponse() to generate markdown tables for all three sections (Pending, Decided, Canceled)
  - Added table header: `| Code | Status | Requestor | Approver | Created |`
  - Added separator row: `|------|--------|-----------|----------|----------|`
  - Changed row format from `**CODE** | status | ...` to `| CODE | status | @user | @user | date |`
  - Preserved canceled reason display logic (truncation, UTF-8 handling)
  - Modified executeList() to use SendEphemeralPost instead of CommandResponse.Text (lines 413-491)
  - Added fallback logic when SendEphemeralPost returns nil (consistent error handling)
  - **Reason for API change:** CommandResponse.Text doesn't render markdown tables; ephemeral posts do

- `server/command/router_test.go` (all TestExecuteList functions)
  - Updated all 16 TestExecuteList test functions to mock SendEphemeralPost
  - Changed assertions from `resp.Text` to `capturedPost.Message`
  - Added `ChannelId` to test CommandArgs
  - Updated TestFormatListResponse_GroupedSections to verify table headers and separators
  - All 443 tests passing

- `README.md` (lines 10-50)
  - Updated `/approve list` documentation to show filter syntax
  - Added autocomplete feature to feature list
  - Corrected command examples to use CODE instead of ID

### Completion Notes

**Implementation Summary:**
Successfully converted `/approve list` output from plain text to professional markdown tables. Key changes:
- Added 3 lines per section (blank line + table header + separator row)
- Modified format string from pipe-separated plain text to table row format
- **CRITICAL DISCOVERY:** Blank line required after section headers for Mattermost markdown parser to recognize tables
- Changed delivery mechanism from CommandResponse.Text to API.SendEphemeralPost() (enables markdown rendering)
- Added consistent error handling with fallback logic across all code paths
- No changes to business logic (grouping, sorting, filtering, pagination)
- Zero breaking changes

**Debugging Notes:**
Initial implementation didn't render tables. After investigation, discovered Mattermost requires blank line before markdown tables. Changed `"**Section:**\n"` to `"**Section:**\n\n"` which fixed rendering. This is NOT documented in Mattermost markdown spec but is critical for table parsing.

**Testing:**
- All 443 unit tests pass
- Plugin builds successfully across all platforms (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64)
- Existing tests for grouping, sorting, filtering, pagination, UTF-8 handling all still pass
- Updated tests now verify markdown table format
- Manual testing on web app confirmed tables render correctly

**Files Changed:** 3 files, ~100 lines modified (including test updates)
**Actual Time:** ~90 minutes (implementation + debugging + testing + fixing rendering issue)

**Outstanding Items:**
- ✅ Manual testing on web app completed - tables render correctly
- ⏳ Manual testing on desktop app - pending user testing
- ⏳ Manual testing on mobile app - pending user testing
- ✅ Documentation updated (README.md)
- ⏳ Code review - in progress

**Risk Assessment:** LOW
- Formatting change with delivery mechanism update
- All automated tests passing
- Error handling improved (consistent fallback logic)
- Markdown tables are standard Mattermost feature (well-supported across all clients)
- Easy to rollback if issues found during manual testing
- Blank line requirement documented for future maintainers
