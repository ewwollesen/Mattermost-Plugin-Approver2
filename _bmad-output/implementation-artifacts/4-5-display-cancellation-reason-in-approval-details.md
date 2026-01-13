# Story 4.5: Display Cancellation Reason in Approval Details

Status: done

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Priority:** Medium
**Estimate:** 2 points
**Assignee:** Dev Agent

## Story

As a **user (requester or approver)**,
I want **to see why a request was cancelled when viewing details**,
so that **I have full context of what happened**.

## Context

When users run `/approve get [ID]` to view a specific approval, they currently see basic information but NO cancellation details if the request was cancelled. This creates an incomplete audit trail.

**Current Behavior (router.go:530-582):**
```
📋 Approval Record: A-X7K9Q2

Status: 🚫 Canceled

Requester: @wayne (Wayne)
Approver: @john (John)

Description:
Deploy hotfix to production database

Requested: 2026-01-12 15:45:00 UTC
Decided: 2026-01-12 19:15:00 UTC

Context:
- Request ID: A-X7K9Q2
- Full ID: abc123def456

This record is immutable and cannot be edited.
```

**Problems:**
1. **No cancellation reason shown** - Users can't see WHY it was cancelled
2. **No cancellation timestamp** - "Decided" field is ambiguous (approved, denied, or canceled?)
3. **Incomplete audit trail** - Missing critical information for process improvement
4. **Inconsistent with Stories 4.1-4.3** - Those stories capture/store/notify with reasons, but details view doesn't show it

**Required Change:**
Add cancellation details section when status is "canceled", showing:
- Cancellation reason (from `record.CanceledReason`)
- Who cancelled (always requester in v0.2.0, explicitly shown for clarity)
- When cancelled (from `record.CanceledAt`)
- Graceful handling for old records without these fields (backwards compatibility)

## Acceptance Criteria

- [x] AC1: `/approve get [ID]` shows cancellation section for cancelled requests
- [x] AC2: Cancellation section displays reason, who cancelled, and timestamp
- [x] AC3: Formatted clearly with visual separator (---) and consistent styling
- [x] AC4: Works for both requester and approver viewing the record
- [x] AC5: Old records without cancellation reason show "No reason recorded (cancelled before v0.2.0)"
- [x] AC6: Old records without CanceledAt timestamp show "Unknown"
- [x] AC7: Status emoji shows 🚫 for cancelled requests (already implemented in getStatusIcon)
- [x] AC8: Format matches existing display patterns (bold labels, Markdown structure)

## Tasks / Subtasks

- [x] Task 1: Update formatRecordDetail to display cancellation section (AC: 1, 2, 3, 8)
  - [x] Subtask 1.1: Add conditional block for status == "canceled"
  - [x] Subtask 1.2: Add visual separator (---) before cancellation section
  - [x] Subtask 1.3: Display cancellation reason (handle empty with fallback text)
  - [x] Subtask 1.4: Display "Canceled by: @{RequesterUsername}"
  - [x] Subtask 1.5: Display cancellation timestamp (handle zero value)
  - [x] Subtask 1.6: Position section after Description, before Context

- [x] Task 2: Handle backwards compatibility (AC: 4, 5, 6)
  - [x] Subtask 2.1: Check if CanceledReason is empty → show "No reason recorded (cancelled before v0.2.0)"
  - [x] Subtask 2.2: Check if CanceledAt is 0 → show "Unknown" instead of timestamp
  - [x] Subtask 2.3: Test with old cancelled records created before Story 4.4

- [x] Task 3: Write comprehensive tests (AC: all)
  - [x] Subtask 3.1: Unit test: Format cancelled request with reason
  - [x] Subtask 3.2: Unit test: Format cancelled request without reason (old record)
  - [x] Subtask 3.3: Unit test: Format cancelled request with CanceledAt = 0
  - [x] Subtask 3.4: Unit test: Verify cancellation section placement
  - [x] Subtask 3.5: Command layer tests: Mock cancelled records with all 4 reasons → Get → Verify display (Note: These are unit tests of the command router with mocked store, not true end-to-end integration tests that actually cancel requests. Full integration tests covered by existing service layer tests in approval/service_test.go)
  - [x] Subtask 3.6: Command layer test: "Other" reason with custom text displays correctly (mocked record)

## Dev Notes

### Architecture Compliance

**Display Pattern (from existing code analysis):**
- File: `server/command/router.go`
- Function: `formatRecordDetail` (lines 530-582)
- Pattern: strings.Builder with fmt.Sprintf
- Format: Markdown with bold labels (**Label:**)
- Timestamps: "2006-01-02 15:04:05 MST" (time.Unix().UTC().Format())

**Existing Status-Specific Logic:**
Currently, the function doesn't have status-specific sections. All records show the same fields regardless of status. We need to add a status-specific block for "canceled".

**Placement Strategy:**
Insert cancellation section AFTER description, BEFORE context section:
1. Header (code, status)
2. Requester/Approver info
3. Description
4. **→ CANCELLATION SECTION (NEW)** ← Insert here
5. Timestamps
6. Decision comment (if present)
7. Context
8. Footer

### Current Implementation Context

**File to Modify:**
- `server/command/router.go` - `formatRecordDetail` function (lines 530-582)
  - Currently: No status-specific formatting
  - Change: Add conditional block for status == "canceled"

**Data Already Available (Story 4.4):**
- `record.CanceledReason` - string field, may be empty for old records
- `record.CanceledAt` - int64 epoch millis, may be 0 for old records
- `record.RequesterUsername` - who cancelled (only requesters can cancel in v0.2.0)

**Story Dependencies (COMPLETED):**
- Story 4.3: Modal captures reason from user ✅
- Story 4.4: Fields added to ApprovalRecord struct ✅
- Story 4.4: Service.CancelApproval stores reason ✅

**No Breaking Changes:**
- Existing `/approve get` behavior unchanged for non-cancelled requests
- Backwards compatible with old cancelled records
- No API changes, no database schema changes

### Implementation Specification

```go
// In server/command/router.go, formatRecordDetail function
// Insert after Description section, before Timestamps section

// Cancellation details (Story 4.5: display reason/timestamp for canceled requests)
if record.Status == approval.StatusCanceled {
	output.WriteString("---\n\n")
	output.WriteString("**Cancellation:**\n")

	// Cancellation reason (handle empty for old records)
	reason := record.CanceledReason
	if reason == "" {
		reason = "No reason recorded (cancelled before v0.2.0)"
	}
	output.WriteString(fmt.Sprintf("**Reason:** %s\n", reason))

	// Who cancelled (always requester in v0.2.0, but explicit for audit clarity)
	output.WriteString(fmt.Sprintf("**Canceled by:** @%s\n", record.RequesterUsername))

	// When cancelled (handle zero value for old records)
	if record.CanceledAt > 0 {
		canceledTime := time.Unix(0, record.CanceledAt*int64(time.Millisecond))
		formattedCanceled := canceledTime.UTC().Format("2006-01-02 15:04:05 MST")
		output.WriteString(fmt.Sprintf("**Canceled:** %s\n", formattedCanceled))
	} else {
		output.WriteString("**Canceled:** Unknown\n")
	}

	output.WriteString("\n") // Spacing before next section
}
```

### Testing Requirements

**Unit Tests (router_test.go):**
1. Test formatRecordDetail with cancelled status + reason
2. Test formatRecordDetail with cancelled status but empty reason (old record)
3. Test formatRecordDetail with cancelled status but CanceledAt = 0
4. Test formatRecordDetail with "Other: custom text" reason displays correctly
5. Test non-cancelled statuses unchanged (pending, approved, denied)

**Integration Tests (router_test.go or api_test.go):**
1. Create request → Cancel with "No longer needed" → Get → Verify display
2. Create request → Cancel with "Other: detailed explanation" → Get → Verify display
3. Mock old cancelled record (empty reason) → Get → Verify "No reason recorded" appears

**Expected Output Examples:**

*Cancelled with reason:*
```
📋 Approval Record: A-X7K9Q2

Status: 🚫 Canceled

Requester: @wayne (Wayne)
Approver: @john (John)

Description:
Deploy hotfix to production database

---

Cancellation:
**Reason:** No longer needed
**Canceled by:** @wayne
**Canceled:** 2026-01-12 19:15:00 UTC

Requested: 2026-01-12 15:45:00 UTC
Decided: 2026-01-12 19:15:00 UTC

Context:
- Request ID: A-X7K9Q2
...
```

*Cancelled before v0.2.0 (old record):*
```
...

---

Cancellation:
**Reason:** No reason recorded (cancelled before v0.2.0)
**Canceled by:** @wayne
**Canceled:** Unknown

...
```

### Edge Cases

1. **Old cancelled records (before Story 4.4)**
   - CanceledReason = "" (empty string)
   - CanceledAt = 0
   - Display fallback text, don't crash

2. **"Other" reason with long custom text**
   - Could be up to 500 chars
   - Display full text (detail view, not list)
   - Markdown formatting preserved

3. **Cancelled by deleted user**
   - Username stored at cancel time (snapshot)
   - Display stored RequesterUsername even if user deleted
   - No API lookup needed

4. **Decision comment present on cancelled request**
   - Should NOT display decision comment for cancelled status
   - Decision comment only relevant for approved/denied
   - Current code already handles this correctly (checks if comment != "")

5. **Markdown formatting in CanceledReason (Code Review Note)**
   - CanceledReason is displayed without escaping: `fmt.Sprintf("**Reason:** %s\n", reason)`
   - This is INTENTIONAL: Mattermost uses markdown for DM formatting
   - Users can include markdown (bold, links) in "Other" reason custom text
   - Risk: User could inject formatting, but this is acceptable in private DMs
   - Example: "Other: **URGENT** - see [ticket](http://jira.com/ABC-123)" renders correctly
   - Not a security issue: Users can only see their own approval details or approvals they're involved in

### References

- [Source: server/command/router.go:530-582] - formatRecordDetail implementation
- [Source: server/approval/models.go:35-36] - CanceledReason and CanceledAt fields
- [Source: Story 4.3] - Cancellation reason modal and enumerated values
- [Source: Story 4.4] - Data model additions and storage

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

N/A - No issues encountered during implementation

### Completion Notes List

**Implementation Completed:**

1. **Cancellation Display Logic** (Task 1 - All Subtasks)
   - Added conditional block in formatRecordDetail (router.go:547-572) to check for StatusCanceled
   - Visual separator (---) added before cancellation section
   - Cancellation reason displayed with fallback for empty string: "No reason recorded (cancelled before v0.2.0)"
   - "Canceled by" shows RequesterUsername (only requesters can cancel in v0.2.0)
   - Cancellation timestamp formatted as "2006-01-02 15:04:05 MST" with fallback "Unknown" for zero value
   - Section positioned after Description, before Timestamps

2. **Backwards Compatibility** (Task 2 - All Subtasks)
   - Empty CanceledReason → displays "No reason recorded (cancelled before v0.2.0)"
   - CanceledAt = 0 → displays "Unknown"
   - All edge cases handled gracefully without crashes

3. **Comprehensive Test Coverage** (Task 3 - All Subtasks)
   - Added 4 new test functions in router_test.go (lines 1339-1527)
   - Test 1: Cancelled request with reason - verifies all fields display correctly and section placement
   - Test 2: Old record without reason/timestamp - verifies fallback text
   - Test 3: "Other" reason with long custom text - verifies full text display
   - Test 4: Non-cancelled statuses - verifies cancellation section NOT displayed
   - All 331 tests passing (includes existing tests + new tests)

4. **All Acceptance Criteria Met:**
   - AC1-AC8: All verified through comprehensive unit tests
   - Formatting matches existing patterns (bold labels, Markdown)
   - Works for both requester and approver (access control unchanged)
   - Status icon 🚫 for cancelled requests (pre-existing in getStatusIcon)

### File List

**Modified Files (Story 4.5):**
- server/command/router.go (lines 547-572) - Added cancellation display logic to formatRecordDetail
- server/command/router_test.go (lines 1339-1527) - Added 4 comprehensive test functions
- _bmad-output/implementation-artifacts/4-5-display-cancellation-reason-in-approval-details.md - Updated status and marked all tasks complete
- _bmad-output/implementation-artifacts/sprint-status.yaml (line 79) - Updated status: ready-for-dev → in-progress → review

**Additional Uncommitted Files (Story 4.3 Code Review Fixes):**
Note: The following files show uncommitted changes from Story 4.3 code review fixes, not Story 4.5:
- server/api.go - Story 4.3 code review fixes (mapCancellationReason method signature change)
- server/api_test.go - Story 4.3 code review fixes (added TestMapCancellationReason, TestHandleCancelModalSubmission)
- server/plugin.go - Story 4.3 code review fixes (reference code moved to IntroductionText)
- server/plugin_test.go - Story 4.3 code review fixes (updated modal tests)
- _bmad-output/implementation-artifacts/4-3-add-cancellation-reason-dropdown.md - Story 4.3 marked done

**No Files Created:**
All changes were modifications to existing files

**Implementation Summary:**
- Lines of code added: ~60 (31 lines implementation + 240 lines tests)
- Test coverage: 6 test functions with 9 test cases covering all edge cases including all 4 cancellation reasons
- All tests passing: 333 tests (0 failures)
- TDD approach: Tests written first (RED), then implementation (GREEN)

**Code Review Fixes Applied:**
1. **Issue #1 (HIGH) - Incomplete File List:** Updated File List to document all 9 files changed, clarifying which were Story 4.3 code review fixes vs Story 4.5
2. **Issue #2 (HIGH) - Integration Test Clarification:** Updated Subtask 3.5 description to accurately reflect command layer unit tests vs true integration tests
3. **Issue #3 (MEDIUM) - Confusing Decided Timestamp:** Hidden "Decided:" timestamp for cancelled requests (router.go:582-590) to avoid UX confusion
4. **Issue #4 (MEDIUM) - Missing Test Coverage:** Added 2 new tests for "Changed requirements" and "Created by mistake" reasons (router_test.go:1485-1561)
5. **Issue #5 (LOW) - Markdown Escaping:** Documented markdown handling as intentional behavior (Edge Case #5)
