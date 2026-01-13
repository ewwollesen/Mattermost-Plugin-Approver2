# Epic 5: List Filtering and Default View Optimization

**Version:** 0.3.0
**Status:** Planned
**Priority:** Medium
**Created:** 2026-01-13

## Overview

Add filtering capabilities to `/approve list` command and optimize the default view to focus on actionable items. Currently, all approval requests (pending, approved, denied, canceled) are shown by default, creating noise and making it harder to identify what needs attention.

## Problem Statement

**Current Issues:**
1. `/approve list` shows all requests regardless of status
2. Users must scan through completed/canceled requests to find pending items
3. No way to view only approved, denied, or canceled requests for audit purposes
4. List can become cluttered, especially after Epic 4 added cancellation tracking

**User Impact:**
- Time wasted scanning irrelevant completed requests
- Pending approvals buried in noise
- No efficient way to audit historical decisions by status
- Poor user experience when dogfooding the plugin

## Goals

### Primary Goals
1. **Focus on Action:** Default view shows only pending requests (what needs attention)
2. **Flexible Filtering:** Users can explicitly request approved/denied/canceled/all views
3. **Clear Feedback:** Count display and empty state messaging guide users
4. **Backward Compatibility:** All current functionality preserved via `all` filter

### Success Metrics
- Default `/approve list` shows pending only
- All 5 filters work correctly (pending, approved, denied, canceled, all)
- Header shows accurate count for filtered view
- Empty states provide clear guidance

## User Stories

### Story 5.1: Add Filter Parameter to List Command
**As a** user
**I want** to filter approval requests by status
**So that** I can view specific subsets (approved, denied, canceled) for audit or reference

**Acceptance Criteria:**
- `/approve list <filter>` accepts positional filter argument
- Supported filters: `pending`, `approved`, `denied`, `canceled`, `all`
- Filters are case-insensitive
- Invalid filter shows helpful error message with valid options
- Filtering logic correctly matches request status field
- All existing list functionality preserved (sorting, formatting)

**Technical Notes:**
- Add filter parsing to list command handler
- Validate filter values before query
- Filter at data layer (not display layer)
- Maintain Epic 4.6 sorting (canceled at bottom for `all` view)

---

### Story 5.2: Change Default Behavior and Add Count Display
**As a** user
**I want** the default list view to show only pending requests with a count
**So that** I immediately see what needs my attention

**Acceptance Criteria:**
- `/approve list` (no arguments) shows pending requests only
- Header format: "## Your Approval Requests (N pending/approved/denied/canceled/all)"
- Empty state message: "No [status] approval requests. Use `/approve list all` to see all requests."
- Count accurately reflects filtered results
- Breaking change documented in changelog

**Technical Notes:**
- Default filter parameter = "pending"
- Count logic based on filtered result set
- Empty state varies by filter type
- Update help text to document new behavior

---

## Technical Considerations

### Command Parsing
- Positional argument after `list` subcommand
- Lowercase comparison for case-insensitivity
- Graceful handling of extra arguments (ignore or error?)

### Data Layer
- Existing `GetUserApprovals` may need filter parameter
- Or filter in command layer from full result set
- Consider performance: query-level filtering vs. post-query filtering

### Breaking Changes
- Default behavior changes from "all" to "pending"
- Users accustomed to seeing everything will see filtered view
- Mitigated by clear empty state messaging

### Testing
- Unit tests for filter parsing
- Unit tests for each filter type
- Integration tests for count accuracy
- Edge cases: empty results, invalid filters, no arguments

## Dependencies

- **Builds on:** Epic 1 (list command exists), Epic 4 (canceled status exists)
- **Blocks:** None
- **Blocked by:** None

## Out of Scope

- Multi-status filtering (e.g., `pending,approved`)
- Date range filtering
- Search/text filtering
- Pagination (if list gets very long)
- Sort order changes (Epic 4.6 sorting preserved)

## Implementation Order

1. Story 5.1: Add filtering (keeps current default "all")
2. Story 5.2: Change default + add count (requires 5.1 filtering logic)

Sequential implementation ensures testing at each stage.

## Success Validation

- `/approve list` shows pending only ✓
- `/approve list pending` shows pending ✓
- `/approve list approved` shows approved ✓
- `/approve list denied` shows denied ✓
- `/approve list canceled` shows canceled ✓
- `/approve list all` shows everything ✓
- `/approve list invalid` shows error ✓
- Header count matches filtered results ✓
- Empty state messaging clear ✓

## Notes

- This is Wayne dogfooding his own plugin
- Small, focused epic - should be fast to implement
- No new data model changes required
- Good candidate for quick-dev workflow if desired
