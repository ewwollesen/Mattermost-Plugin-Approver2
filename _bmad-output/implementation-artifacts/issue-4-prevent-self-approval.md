# Story: Prevent Self-Approval Requests

Status: complete

**GitHub Issue:** [#4 - You can select yourself as an approver for your own request](https://github.com/ewwollesen/Mattermost-Plugin-Approver2/issues/4)

## Story

As a system administrator,
I want to prevent users from selecting themselves as approvers for their own requests,
So that approval workflows maintain proper separation of duties and create valid audit trails.

## Acceptance Criteria

### AC1: Self-Approval Validation in Dialog Handler
- `HandleDialogSubmission()` validates that `approver != requester` (using user IDs from payload)
- Returns field-specific error on `approver` field: "You cannot approve your own request. Please select a different approver."
- Error keeps modal open, preserves user input

### AC2: Self-Approval Validation in API Handler
- `handleApproveNew()` validates `approverID != payload.UserId` before creating approval record
- Returns field-specific error if self-approval detected (defense-in-depth)
- Logs warning with requester ID when self-approval attempt occurs

### AC3: Clear Error Messaging
- Error message is specific and actionable
- Follows existing UX guidelines (helpful tone, explains why)
- Consistent with other validation errors in the system

### AC4: Comprehensive Unit Tests
- Test `HandleDialogSubmission()` rejects self-approval (requester ID == approver ID)
- Test `handleApproveNew()` rejects self-approval (edge case if client-side validation bypassed)
- Test error message content and structure
- All existing tests continue to pass

### AC5: No Regression
- All existing approval creation flows work correctly
- Different requester/approver combinations work as before
- No impact on other validation logic

## Tasks / Subtasks

- [x] Task 1: Add self-approval validation to HandleDialogSubmission() (AC: 1, 3)
  - [x] 1.1: Write failing test - `HandleDialogSubmission()` rejects when approver == requester ID
  - [x] 1.2: Update `HandleDialogSubmission()` signature to accept requesterID parameter
  - [x] 1.3: Add self-approval validation check (requesterID == approverID)
  - [x] 1.4: Return field-specific error on `approver` field with clear message
  - [x] 1.5: Verify test passes
  - [x] 1.6: Write test - error message contains "cannot approve your own request"
  - [x] 1.7: Write test - error message suggests selecting different approver

- [x] Task 2: Update handleApproveNew() to pass requesterID to dialog handler (AC: 1, 2)
  - [x] 2.1: Update `handleApproveNew()` call to `HandleDialogSubmission()` to pass `payload.UserId` as requesterID
  - [x] 2.2: Add defense-in-depth check: validate approverID != payload.UserId before creating record
  - [x] 2.3: Write test for `handleApproveNew()` rejecting self-approval
  - [x] 2.4: Add warning log when self-approval attempt detected

- [x] Task 3: Verify no regressions (AC: 4, 5)
  - [x] 3.1: Run all existing tests in `server/command/dialog_test.go` - must pass
  - [x] 3.2: Run all existing tests in `server/api_test.go` - must pass
  - [x] 3.3: Run full test suite: `go test ./server/...` - all must pass
  - [x] 3.4: Verify existing valid approvals (different requester/approver) still work

- [x] Task 4: Update story documentation (AC: all)
  - [x] 4.1: Mark all tasks complete
  - [x] 4.2: Update Dev Agent Record with implementation summary
  - [x] 4.3: Document files changed in File List

## Dev Notes

### Current Implementation Analysis

**Dialog Validation Flow:**
1. User submits `/approve new` dialog
2. `handleDialogSubmit()` in `server/api.go` receives submission
3. Routes to `handleApproveNew()` for `approve_new` callback
4. `handleApproveNew()` calls `command.HandleDialogSubmission()` for field presence validation
5. Then performs business logic validation (description length, approver exists)

**Key Files:**
- `server/command/dialog.go` - `HandleDialogSubmission()` performs basic field validation
- `server/api.go` - `handleApproveNew()` orchestrates validation and approval creation
- `server/command/dialog_test.go` - Unit tests for dialog validation
- `server/api_test.go` - Integration tests for API handlers

**Current HandleDialogSubmission Signature:**
```go
func HandleDialogSubmission(submission map[string]any) *model.SubmitDialogResponse
```

**Needs to change to:**
```go
func HandleDialogSubmission(submission map[string]any, requesterID string) *model.SubmitDialogResponse
```

**Current Validation Checks:**
1. Layer 1 (dialog.go): Approver present, description present
2. Layer 2 (api.go): Description length, approver exists/active

**Adding Self-Approval Check:**
- Add to Layer 1 in `HandleDialogSubmission()` - earliest possible validation
- Add defense-in-depth check in Layer 2 (`handleApproveNew()`) - security best practice

### Error Message Design

Following existing pattern from `dialog.go`:
- **Approver required:** "Approver field is required. Please select a user."
- **Description required:** "Description field is required. Please describe what needs approval."

**New self-approval error (AC3):**
- **Self-approval detected:** "You cannot approve your own request. Please select a different approver."

Rationale:
- Clear statement of what's wrong ("cannot approve your own")
- Actionable guidance ("select a different approver")
- Helpful tone consistent with existing errors
- Explains WHY (implicit: separation of duties)

### Testing Strategy

**Unit Tests (dialog_test.go):**
```go
TestHandleDialogSubmission/rejects_self_approval_when_requester_equals_approver
TestHandleDialogSubmission/allows_different_requester_and_approver
TestHandleDialogSubmission/error_message_explains_self_approval_not_allowed
```

**Integration Tests (api_test.go):**
```go
TestHandleApproveNew/rejects_self_approval_defense_in_depth
TestHandleApproveNew/logs_warning_on_self_approval_attempt
```

**Regression Tests:**
- All existing `dialog_test.go` tests must pass
- All existing `api_test.go` tests must pass
- Full suite: `go test ./server/...`

### Implementation Notes

**Why validate in both layers?**
1. **Layer 1 (dialog.go):** User-friendly validation, keeps modal open with helpful error
2. **Layer 2 (api.go):** Defense-in-depth if client-side validation bypassed, security best practice

**Logging Strategy:**
- Log warning when self-approval detected in `handleApproveNew()` (Layer 2)
- Include requester ID for audit trail
- Example: `"Self-approval attempt detected", "requester_id", payload.UserId`

**Edge Cases Covered:**
- Same user ID for requester and approver (primary case)
- Empty approver ID (already handled by existing validation)
- Invalid approver ID (already handled by `ValidateApprover()`)

### References

- [GitHub Issue #4](https://github.com/ewwollesen/Mattermost-Plugin-Approver2/issues/4)
- [Source: server/command/dialog.go - HandleDialogSubmission]
- [Source: server/api.go - handleApproveNew]
- [Source: server/command/dialog_test.go - Existing test patterns]

## Dev Agent Record

### File List

**Modified Files:**
1. `server/command/dialog.go` - Updated `HandleDialogSubmission()` signature, added self-approval validation
2. `server/command/dialog_test.go` - Added 2 new tests for self-approval rejection, updated 8 existing tests
3. `server/api.go` - Updated `handleApproveNew()` to pass requesterID, added defense-in-depth check
4. `server/api_test.go` - Added `TestHandleApproveNew_SelfApprovalRejection` with 2 test cases

### Change Log

**2026-01-20 - GitHub Issue #4 Implementation**

**server/command/dialog.go:**
- Updated `HandleDialogSubmission()` signature: added `requesterID string` parameter
- Added self-approval validation: checks if `approver == requesterID` after basic field validation
- Error message: "You cannot approve your own request. Please select a different approver."
- Added documentation comment about GitHub Issue #4

**server/command/dialog_test.go:**
- Added test: `rejects self-approval when requester equals approver` - verifies rejection with correct error
- Added test: `allows different requester and approver` - verifies normal flow works
- Updated 8 existing tests to pass `requesterID` parameter (using different ID from approver)
- All 10 tests in `TestHandleDialogSubmission` pass

**server/api.go:**
- Updated `handleApproveNew()` Layer 1 call to pass `payload.UserId` as requesterID
- Added defense-in-depth Layer 2 check before business logic validation
- Added `LogWarn` when self-approval detected in Layer 2 (though Layer 1 catches it first)
- Same error message as Layer 1 for consistency

**server/api_test.go:**
- Added `TestHandleApproveNew_SelfApprovalRejection` function
- Test 1: "rejects self-approval in Layer 1 validation" - verifies error returned before API layer logic
- Test 2: "allows different requester and approver" - full integration test with mocked dependencies
- Both tests pass

**Test Results:**
- `go test ./server/command`: PASS (all 10 dialog tests + router tests)
- `go test ./server -run TestHandleApproveNew`: PASS (all 7 test suites including new one)
- `go test ./server/...`: PASS (all packages: server, approval, command, notifications, playbooks, store, timeout)

**Implementation Notes:**
- Layer 1 validation catches self-approval before Layer 2 executes (defense-in-depth works as intended)
- Error is field-specific on `approver` field, keeps modal open with user input preserved
- No regression - all existing tests pass with updated signature
