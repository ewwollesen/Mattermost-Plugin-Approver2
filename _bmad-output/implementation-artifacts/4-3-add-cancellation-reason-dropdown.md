# Story 4.3: Add Cancellation Reason Dropdown

Status: done

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Priority:** High
**Estimate:** 5 points
**Assignee:** Dev Agent

## Story

As a **requester**,
I want **to specify why I'm cancelling a request**,
so that **there's a clear record and the team can improve the process**.

## Context

Currently (as of Stories 4.1, 4.2, 4.4), `/approve cancel [ID]` immediately cancels the request with a hard-coded default reason of "No longer needed" (see plugin.go:138-140). This implementation was intentional - Story 4.4 (data model) was implemented before Story 4.3 (reason capture UI) to unblock Stories 4.1 and 4.2.

**Current Flow:**
1. User types `/approve cancel TUZ-2RK`
2. Plugin immediately cancels with reason = "No longer needed"
3. Record saved, post updated, DM sent

**Problems:**
1. **No safety confirmation** - accidental cancellations can happen
2. **No reason variety** - ALL cancellations show "No longer needed" regardless of actual reason
3. **Defeats data instrumentation goal** - Can't identify patterns (30% wrong approver, 18% sensitive info, etc.)
4. **Poor UX** - No chance to change mind before destructive action

**Required Change:**
Replace immediate cancellation with modal-based flow:
1. User types `/approve cancel TUZ-2RK`
2. Plugin opens modal with reason dropdown
3. User selects reason + confirms
4. Plugin cancels with user-selected reason

## Acceptance Criteria

- [x] AC1: `/approve cancel [ID]` opens a confirmation modal (does NOT cancel immediately)
- [x] AC2: Modal shows dropdown with 4 enumerated reasons (see UI mockup)
  - "No longer needed" (default selection)
  - "Wrong approver"
  - "Sensitive information"
  - "Other reason"
- [x] AC3: If "Other reason" selected, textarea field becomes required for explanation
- [x] AC4: Cancellation reason selection is required (cannot submit modal without choosing)
- [x] AC5: Modal has "Cancel Request" submit button and triggers confirmation on submit
- [x] AC6: After modal submission, request is cancelled with the user-selected reason
- [x] AC7: If modal is dismissed (user clicks outside or presses ESC), cancellation is aborted (request stays pending)
- [x] AC8: Validation error if "Other" selected without providing text explanation

## Tasks / Subtasks

- [x] Task 1: Update handleCancelCommand to open modal instead of immediate cancel (AC: 1, 7)
  - [x] Subtask 1.1: Remove immediate cancellation logic (lines 138-241 in plugin.go)
  - [x] Subtask 1.2: Extract approval code and validate it exists BEFORE opening modal
  - [x] Subtask 1.3: Call openCancellationModal with triggerID, approval code, record
  - [x] Subtask 1.4: Return command response after modal opens (no cancellation yet)

- [x] Task 2: Implement openCancellationModal function (AC: 1, 2, 3, 4, 5)
  - [x] Subtask 2.1: Create modal structure with callback ID format: "cancel_approval_{approvalID}"
  - [x] Subtask 2.2: Display reference code in IntroductionText (not editable field)
  - [x] Subtask 2.3: Add cancellation reason dropdown with 4 options
  - [x] Subtask 2.4: Add "other_reason_text" textarea field (optional=true initially)
  - [x] Subtask 2.5: Set "No longer needed" as default selection
  - [x] Subtask 2.6: Add help text: "Help us understand why requests are cancelled"
  - [x] Subtask 2.7: Call API.OpenInteractiveDialog with modal definition

- [x] Task 3: Implement handleCancelModalSubmission in api.go (AC: 3, 6, 8)
  - [x] Subtask 3.1: Register new route in ServeHTTP: "/dialog/submit" already exists, add routing logic
  - [x] Subtask 3.2: Update handleDialogSubmit to route cancel_approval_* callbacks
  - [x] Subtask 3.3: Parse callback ID to extract approval ID
  - [x] Subtask 3.4: Extract reason_code and other_reason_text from submission
  - [x] Subtask 3.5: Validate: if reason="other" AND other_reason_text is empty → return validation error
  - [x] Subtask 3.6: Map reason code to human-readable text (helper function)
  - [x] Subtask 3.7: Retrieve approval record and verify user is requester
  - [x] Subtask 3.8: Call service.CancelApproval with mapped reason text
  - [x] Subtask 3.9: Perform post-cancellation actions (post update, DM notification)
  - [x] Subtask 3.10: Return empty SubmitDialogResponse on success

- [x] Task 4: Implement reason mapping helper function (AC: 2, 3)
  - [x] Subtask 4.1: Create mapCancellationReason(code, otherText string) string as Plugin method
  - [x] Subtask 4.2: Map "no_longer_needed" → "No longer needed"
  - [x] Subtask 4.3: Map "wrong_approver" → "Wrong approver"
  - [x] Subtask 4.4: Map "sensitive_info" → "Sensitive information"
  - [x] Subtask 4.5: Map "other" → "Other: {otherText}"
  - [x] Subtask 4.6: Handle unknown codes with default: "Unknown reason"

- [x] Task 5: Write comprehensive tests (AC: all)
  - [x] Subtask 5.1: Unit test: handleCancelCommand opens modal (no longer cancels immediately)
  - [x] Subtask 5.2: Unit test: Modal structure validation (dropdown options, fields)
  - [x] Subtask 5.3: Unit test: Reason code mapping for all 4 options (TestMapCancellationReason)
  - [x] Subtask 5.4: Unit test: Validation error when "other" selected without text (TestHandleCancelModalSubmission)
  - [x] Subtask 5.5: Unit test: Permission check (only requester can cancel) (TestHandleCancelModalSubmission)
  - [x] Subtask 5.6: Unit test: Successful cancellation with each reason type (TestHandleCancelModalSubmission)
  - [x] Subtask 5.7: Integration test: Full cancel flow with modal submission (TestHandleCancelCommand_Integration)
  - [x] Subtask 5.8: Test error scenarios: invalid approval ID, wrong user, non-pending status

## Dev Notes

### Architecture Compliance

**Mattermost Modal Patterns (from codebase analysis):**
- Modal opened via `p.API.OpenInteractiveDialog(model.OpenDialogRequest{...})`
- Submission handled via `/dialog/submit` endpoint registered in ServeHTTP
- Callback ID format should encode context: `"cancel_approval_{approvalID}"`
- Dialog response errors keep modal open, empty response closes modal

**Existing Modal Examples:**
- `openConfirmationModal` (api.go:351-400) - confirmation dialog pattern
- `handleApproveNew` (api.go:112-277) - field validation and submission pattern
- `handleDialogSubmit` (api.go:58-103) - routing based on callback ID prefix

### Current Implementation Context

**Files to Modify:**
1. **server/plugin.go** - `handleCancelCommand` function (lines 117-241)
   - Currently: immediate cancellation with hard-coded reason
   - Change to: open modal and return

2. **server/api.go** - Add modal submission handler
   - Add `handleCancelModalSubmission` function
   - Update `handleDialogSubmit` routing to handle `cancel_approval_*` callbacks
   - Perform actual cancellation logic moved from plugin.go

**Story 4.4 Integration:**
Story 4.4 (data model changes) is ALREADY COMPLETE:
- `ApprovalRecord.CanceledReason` field exists (models.go:35)
- `ApprovalRecord.CanceledAt` field exists (models.go:36)
- `service.CancelApproval(code, requesterID, reason string)` signature accepts reason parameter (service.go:48)
- Reason validation exists (service.go:65-68)

**Post-Cancellation Flow (already implemented in Stories 4.1, 4.2):**
After cancellation record is saved, these happen automatically:
- UpdateApprovalPostForCancellation (Story 4.1) - updates original DM post
- SendCancellationNotificationDM (Story 4.2) - sends separate notification DM
- Both are best-effort, logged as warnings if they fail

### Modal Definition Specification

```go
func (p *Plugin) openCancellationModal(triggerID string, approvalCode string, record *approval.ApprovalRecord) error {
    dialog := model.OpenDialogRequest{
        TriggerId: triggerID,
        URL:       "/plugins/com.mattermost.plugin-approver2/dialog/submit",
        Dialog: model.Dialog{
            CallbackId:       fmt.Sprintf("cancel_approval_%s", record.ID),
            Title:            "Cancel Approval Request",
            IntroductionText: fmt.Sprintf("You are about to cancel approval request `%s`\n\n**This action cannot be undone.**", approvalCode),
            Elements: []model.DialogElement{
                {
                    DisplayName: "Reference Code",
                    Name:        "reference_code",
                    Type:        "text",
                    Default:     approvalCode,
                    Optional:    false,
                    HelpText:    "",
                    Placeholder: "",
                },
                {
                    DisplayName: "Reason for cancellation",
                    Name:        "reason_code",
                    Type:        "select",
                    Placeholder: "Select a reason",
                    Options: []*model.PostActionOptions{
                        {Text: "No longer needed", Value: "no_longer_needed"},
                        {Text: "Wrong approver", Value: "wrong_approver"},
                        {Text: "Sensitive information", Value: "sensitive_info"},
                        {Text: "Other reason", Value: "other"},
                    },
                    Default:  "no_longer_needed",
                    Optional: false,
                    HelpText: "Help us understand why requests are cancelled",
                },
                {
                    DisplayName: "Additional details (if Other)",
                    Name:        "other_reason_text",
                    Type:        "textarea",
                    Placeholder: "Please explain...",
                    Optional:    true,
                    MaxLength:   500,
                },
            },
            SubmitLabel: "Cancel Request",
        },
    }

    if appErr := p.API.OpenInteractiveDialog(dialog); appErr != nil {
        return fmt.Errorf("failed to open cancellation modal: %w", appErr)
    }

    return nil
}
```

### Modal Submission Handler Specification

```go
// handleCancelModalSubmission processes cancellation modal submissions
func (p *Plugin) handleCancelModalSubmission(payload *model.SubmitDialogRequest) *model.SubmitDialogResponse {
    // Parse callback ID: "cancel_approval_{approvalID}"
    parts := strings.Split(payload.CallbackId, "_")
    if len(parts) != 3 {
        p.API.LogError("Invalid callback ID format", "callback_id", payload.CallbackId)
        return &model.SubmitDialogResponse{
            Error: "Invalid request format",
        }
    }

    approvalID := parts[2]

    // Extract form data
    reasonCode, ok := payload.Submission["reason_code"].(string)
    if !ok || reasonCode == "" {
        return &model.SubmitDialogResponse{
            Errors: map[string]string{
                "reason_code": "Reason is required",
            },
        }
    }

    otherText := ""
    if otherVal, ok := payload.Submission["other_reason_text"]; ok {
        if otherStr, ok := otherVal.(string); ok {
            otherText = strings.TrimSpace(otherStr)
        }
    }

    // Validate "Other" reason requires text
    if reasonCode == "other" && otherText == "" {
        return &model.SubmitDialogResponse{
            Errors: map[string]string{
                "other_reason_text": "Please provide details when selecting 'Other reason'",
            },
        }
    }

    // Map reason code to human-readable text
    reasonText := mapCancellationReason(reasonCode, otherText)

    // Get approval record
    record, err := p.store.GetApproval(approvalID)
    if err != nil {
        p.API.LogError("Failed to get approval record in cancel modal",
            "approval_id", approvalID,
            "error", err.Error(),
        )
        return &model.SubmitDialogResponse{
            Error: "Approval request not found",
        }
    }

    // Verify user is the requester
    if record.RequesterID != payload.UserId {
        p.API.LogError("Unauthorized cancellation attempt",
            "approval_id", approvalID,
            "authenticated_user", payload.UserId,
            "requester_id", record.RequesterID,
        )
        return &model.SubmitDialogResponse{
            Error: "Only the requester can cancel this approval",
        }
    }

    // Get requester user for post update
    requester, appErr := p.API.GetUser(payload.UserId)
    if appErr != nil {
        p.API.LogError("Failed to get requester user",
            "error", appErr.Error(),
            "user_id", payload.UserId,
        )
        return &model.SubmitDialogResponse{
            Error: "Failed to retrieve user information. Please try again.",
        }
    }

    // Call service to cancel approval (Story 4.4 already implemented)
    err = p.service.CancelApproval(record.Code, payload.UserId, reasonText)
    if err != nil {
        p.API.LogError("Failed to cancel approval",
            "error", err.Error(),
            "approval_code", record.Code,
            "user_id", payload.UserId,
        )
        return &model.SubmitDialogResponse{
            Error: p.formatCancelError(err, record.Code),
        }
    }

    // Post-cancellation actions (Stories 4.1, 4.2)
    updatedRecord, err := p.store.GetByCode(record.Code)
    if err != nil {
        p.API.LogWarn("Failed to retrieve updated record for post update",
            "error", err.Error(),
            "approval_code", record.Code,
        )
    } else {
        // Update the original post (Story 4.1)
        err = notifications.UpdateApprovalPostForCancellation(p.API, updatedRecord, requester.Username)
        if err != nil {
            p.API.LogWarn("Failed to update approver post",
                "error", err.Error(),
                "approval_code", record.Code,
                "approver_post_id", updatedRecord.NotificationPostID,
            )
        }

        // Send cancellation notification DM (Story 4.2)
        _, err = notifications.SendCancellationNotificationDM(p.API, p.botUserID, updatedRecord, requester.Username)
        if err != nil {
            p.API.LogWarn("Failed to send cancellation notification",
                "error", err.Error(),
                "approval_id", updatedRecord.ID,
                "approver_id", updatedRecord.ApproverID,
            )
        }
    }

    p.API.LogInfo("Approval canceled successfully via modal",
        "approval_code", record.Code,
        "user_id", payload.UserId,
        "reason", reasonText,
    )

    // Success - return empty response (modal will close)
    return &model.SubmitDialogResponse{}
}

// mapCancellationReason maps reason codes to human-readable text
func mapCancellationReason(code, otherText string) string {
    switch code {
    case "no_longer_needed":
        return "No longer needed"
    case "wrong_approver":
        return "Wrong approver"
    case "sensitive_info":
        return "Sensitive information"
    case "other":
        return fmt.Sprintf("Other: %s", otherText)
    default:
        return "Unknown reason"
    }
}
```

### Routing Update in handleDialogSubmit

```go
// In api.go handleDialogSubmit function (lines 84-94)
switch {
case strings.HasPrefix(payload.CallbackId, "confirm_"):
    response = p.handleConfirmDecision(payload)
case payload.CallbackId == "approve_new":
    response = p.handleApproveNew(payload)
case strings.HasPrefix(payload.CallbackId, "cancel_approval_"):  // NEW
    response = p.handleCancelModalSubmission(payload)           // NEW
default:
    p.API.LogWarn("Unknown dialog callback ID", "callback_id", payload.CallbackId)
    response = &model.SubmitDialogResponse{
        Error: "Unknown dialog type",
    }
}
```

### Testing Strategy

**Unit Tests (server/plugin_test.go, server/api_test.go):**

1. **TestHandleCancelCommand_OpensModal** - Verify modal opens instead of immediate cancel
2. **TestOpenCancellationModal_Structure** - Verify modal has all 4 dropdown options
3. **TestMapCancellationReason** - Test all 4 reason mappings + default
4. **TestHandleCancelModalSubmission_Success** - Each reason type cancels successfully
5. **TestHandleCancelModalSubmission_OtherRequiresText** - Validation error when other+empty
6. **TestHandleCancelModalSubmission_PermissionDenied** - Non-requester cannot cancel
7. **TestHandleCancelModalSubmission_RecordNotFound** - Handle missing approval gracefully
8. **TestHandleCancelModalSubmission_NonPendingStatus** - Cannot cancel decided approvals

**Integration Test Pattern:**
```go
// Execute `/approve cancel TUZ-2RK`
// Mock OpenInteractiveDialog to capture modal definition
// Verify modal structure (dropdown options, fields, defaults)
// Simulate modal submission with each reason type
// Verify cancellation succeeds with correct reason text
// Verify post update and DM notification triggered
```

### Migration Notes

**Backward Compatibility:**
- No data migration needed - Story 4.4 already added CanceledReason/CanceledAt fields
- Existing tests that cancel approvals may need updates if they rely on command-line cancellation
- Old behavior (immediate cancel) is completely replaced - this is intentional UX improvement

**Testing Existing Cancel Functionality:**
After this story, ALL cancellations must go through modal flow. Update any tests that call `/approve cancel` directly to either:
1. Mock the modal opening and submission flow
2. Call `service.CancelApproval` directly (bypasses command layer)

### References

- [Epic 4 Requirements](epic-4-cancellation-ux-audit.md#story-43-add-cancellation-reason-dropdown)
- [Story 4.1 Implementation](4-1-update-approver-dm-post-on-cancellation.md) - Post update logic
- [Story 4.2 Implementation](4-2-send-cancellation-notification-to-approver.md) - Notification DM logic
- [Story 4.4 Data Model](4-4-store-cancellation-reason-in-approval-record.md) - CanceledReason/CanceledAt fields
- [Architecture: Graceful Degradation](_bmad-output/planning-artifacts/architecture.md#22-graceful-degradation-strategy)
- [Mattermost Dialog API](https://developers.mattermost.com/integrate/admin-guide/admin-plugins-beta/interactive-dialogs/)

## UI/UX Mockup

**Modal Display:**
```
┌─────────────────────────────────────────┐
│ Cancel Approval Request                 │
├─────────────────────────────────────────┤
│ You are about to cancel approval        │
│ request `TUZ-2RK`                        │
│                                         │
│ **This action cannot be undone.**       │
│                                         │
│ Reference Code:                         │
│ [TUZ-2RK                     ] (readonly)│
│                                         │
│ Reason for cancellation: *              │
│ [ No longer needed          ▼ ]        │
│ Help us understand why requests are     │
│ cancelled                               │
│                                         │
│ Additional details (if Other):          │
│ ┌─────────────────────────────────────┐ │
│ │                                     │ │
│ │                                     │ │
│ └─────────────────────────────────────┘ │
│                                         │
│               [Cancel Request]          │
└─────────────────────────────────────────┘
```

**When "Other" selected and no text provided:**
```
┌─────────────────────────────────────────┐
│ Cancel Approval Request                 │
├─────────────────────────────────────────┤
│ ...                                     │
│ Reason for cancellation: *              │
│ [ Other reason              ▼ ]        │
│                                         │
│ Additional details (if Other): *        │
│ ┌─────────────────────────────────────┐ │
│ │                                     │ │ ← ERROR: Please provide
│ │                                     │ │   details when selecting
│ └─────────────────────────────────────┘ │   'Other reason'
│                                         │
│               [Cancel Request]          │
└─────────────────────────────────────────┘
```

## Definition of Done

- [ ] handleCancelCommand opens modal instead of immediate cancel
- [ ] Modal displays with all 4 reason options + default selection
- [ ] "Other" reason validation works (requires text explanation)
- [ ] Modal submission handler processes cancellation with selected reason
- [ ] Permission validation (only requester can cancel)
- [ ] Post-cancellation actions triggered (post update, DM notification)
- [ ] All unit tests passing (8+ test cases)
- [ ] Integration test verifies full modal flow
- [ ] Manual testing confirms modal UX works in Mattermost
- [ ] Code follows Mattermost naming conventions
- [ ] Error messages are user-friendly
- [ ] No regressions in existing cancel functionality

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

No debug logs required - implementation completed without blocking issues

### Completion Notes List

**Implementation Completed:**
1. Replaced immediate cancellation with modal-based flow
2. Modal displays 2 elements (reference code moved to IntroductionText for display-only)
3. All 8 Acceptance Criteria fully implemented and verified
4. Comprehensive test coverage: 12 new test cases added (5 for reason mapping, 7 for modal submission)
5. All 300+ tests passing

**Code Review Fixes Applied:**
1. Removed editable reference_code field - moved to IntroductionText (display-only)
2. Changed mapCancellationReason from standalone function to Plugin method for consistency
3. Added TestMapCancellationReason with 5 test cases
4. Added TestHandleCancelModalSubmission with 7 test cases
5. Updated integration tests to reflect new modal structure

**Key Implementation Details:**
- Modal has 2 elements (not 3): reason dropdown + optional "other" textarea
- Reference code displayed in IntroductionText with confirmation message
- Validation ensures "other" reason requires text explanation
- Permission check happens both before modal opens and on submission
- Post-cancellation actions (Stories 4.1, 4.2) integrated as best-effort operations

### File List

**Modified Files:**
- `server/plugin.go` - Updated handleCancelCommand to open modal, added openCancellationModal function
- `server/api.go` - Added handleCancelModalSubmission, mapCancellationReason method, routing for cancel_approval_* callbacks
- `server/api_test.go` - Added TestMapCancellationReason (5 cases), TestHandleCancelModalSubmission (7 cases), updated integration tests
- `server/plugin_test.go` - Updated cancel command tests to expect modal opening instead of immediate cancellation
- `_bmad-output/implementation-artifacts/sprint-status.yaml` - Updated story status to "done"

**Lines Changed:**
- plugin.go:116-243 (128 lines) - Modal-based cancellation flow
- api.go:89-90, 635-781 (149 lines) - Modal submission handling and reason mapping
- api_test.go:3-14, 932-1179 (260 lines) - Comprehensive test coverage
- plugin_test.go:318-320 (removed obsolete fallback test, updated format validation test)
