# Story 4.3: Add Cancellation Reason Dropdown

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Status:** To Do
**Priority:** High
**Estimate:** 5 points
**Assignee:** TBD

## User Story

**As a** requester
**I want** to specify why I'm cancelling a request
**So that** there's a clear record and the team can improve the process

## Context

Currently, `/approve cancel [ID]` immediately cancels the request with no confirmation or reason capture. This lacks:
1. Safety (no "are you sure?" confirmation)
2. Audit trail (no record of WHY it was cancelled)
3. Data for improvement (can't identify patterns like "30% cancelled due to wrong approver")

We need to add a modal workflow that captures structured cancellation reasons.

## Acceptance Criteria

- [ ] `/approve cancel [ID]` opens a confirmation modal (doesn't cancel immediately)
- [ ] Modal shows dropdown with enumerated reasons:
  - No longer needed (default)
  - Wrong approver
  - Sensitive information
  - Other reason
- [ ] If "Other reason" selected, text field appears for explanation
- [ ] Cancellation reason selection is required (cannot submit without choosing)
- [ ] Modal has "Cancel Request" button to confirm and "Go Back" button to abort
- [ ] After submission, request is cancelled with reason stored
- [ ] If user clicks "Go Back", cancellation is aborted (request stays pending)

## Technical Implementation

### Files to Modify
- `server/command/cancel.go` - Update to show modal instead of immediate cancel
- Create `server/command/cancel_modal.go` - Handle modal submission
- `server/approval_request.go` - Add reason parameter to cancel method

### Modal Definition

```go
func (c *CancelCommand) openCancellationModal(requestID, referenceCode string) error {
    modal := model.OpenDialogRequest{
        TriggerId: c.Args.TriggerId,
        URL:       fmt.Sprintf("/plugins/%s/api/cancel", manifest.Id),
        Dialog: model.Dialog{
            Title:       "Cancel Approval Request",
            IconURL:     "", // Optional: plugin icon
            CallbackId:  requestID, // Pass request ID through
            SubmitLabel: "Cancel Request",
            Elements: []model.DialogElement{
                {
                    DisplayName: "Reference Code",
                    Name:        "reference_code",
                    Type:        "text",
                    Default:     referenceCode,
                    Optional:    false,
                    ReadOnly:    true,
                    HelpText:    "",
                },
                {
                    DisplayName: "Reason for cancellation",
                    Name:        "cancellation_reason",
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
        },
    }

    return c.plugin.API.OpenInteractiveDialog(modal)
}
```

### Modal Submission Handler

```go
// In cancel_modal.go
func (p *Plugin) handleCancelModalSubmission(w http.ResponseWriter, r *http.Request) {
    var submission model.SubmitDialogRequest
    if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    requestID := submission.CallbackId
    reason := submission.Submission["cancellation_reason"].(string)
    otherText := submission.Submission["other_reason_text"].(string)

    // If "other" selected, require additional details
    if reason == "other" && strings.TrimSpace(otherText) == "" {
        response := model.SubmitDialogResponse{
            Errors: map[string]string{
                "other_reason_text": "Please provide details when selecting 'Other reason'",
            },
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
        return
    }

    // Map reason code to human-readable text
    reasonText := mapCancellationReason(reason, otherText)

    // Cancel the request with reason
    request, err := p.getApprovalRequest(requestID)
    if err != nil {
        respondWithError(w, "Request not found")
        return
    }

    // Verify user is the requester
    if request.RequesterID != submission.UserId {
        respondWithError(w, "Only the requester can cancel this approval")
        return
    }

    // Perform cancellation
    err = p.cancelApprovalRequest(request, reasonText, submission.UserId)
    if err != nil {
        respondWithError(w, "Failed to cancel request")
        return
    }

    // Success response
    w.WriteHeader(http.StatusOK)
}

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

### API Route Registration

```go
// In plugin.go
func (p *Plugin) initializeAPI() {
    p.router = p.API.NewRouter()
    p.router.HandleFunc("/cancel", p.handleCancelModalSubmission).Methods("POST")
    // ... other routes
}
```

## Testing Requirements

### Unit Tests
- [ ] Test modal structure and required fields
- [ ] Test reason code to text mapping
- [ ] Test validation when "Other" selected without text
- [ ] Test permission check (only requester can cancel)
- [ ] Test with each enumerated reason

### Integration Tests
- [ ] Execute `/approve cancel [ID]` → Modal appears
- [ ] Submit with "No longer needed" → Request cancelled
- [ ] Submit with "Other" + text → Text stored correctly
- [ ] Submit with "Other" + no text → Validation error
- [ ] Click "Go Back" → Modal closes, request unchanged

### Manual Testing
- [ ] Create approval request
- [ ] Run `/approve cancel [ID]`
- [ ] Verify modal appears with all options
- [ ] Test each dropdown option
- [ ] Test "Other" with and without text
- [ ] Verify confirmation button cancels request
- [ ] Verify "Go Back" button aborts

## UI/UX Mockup

```
┌─────────────────────────────────────────┐
│ Cancel Approval Request                 │
├─────────────────────────────────────────┤
│                                         │
│ Reference Code:                         │
│ [TUZ-2RK                     ] (readonly)│
│                                         │
│ Reason for cancellation: *              │
│ [ No longer needed          ▼ ]        │
│   - No longer needed                    │
│   - Wrong approver                      │
│   - Sensitive information               │
│   - Other reason                        │
│                                         │
│ Help us understand why requests are     │
│ cancelled                               │
│                                         │
│ Additional details (if Other):          │
│ ┌─────────────────────────────────────┐ │
│ │                                     │ │
│ │                                     │ │
│ └─────────────────────────────────────┘ │
│                                         │
│           [Cancel Request] [Go Back]    │
└─────────────────────────────────────────┘
```

**When "Other" selected:**
```
┌─────────────────────────────────────────┐
│ Cancel Approval Request                 │
├─────────────────────────────────────────┤
│                                         │
│ Reference Code:                         │
│ [TUZ-2RK                     ] (readonly)│
│                                         │
│ Reason for cancellation: *              │
│ [ Other reason              ▼ ]        │
│                                         │
│ Additional details (if Other): *        │
│ ┌─────────────────────────────────────┐ │
│ │ Approver is on vacation and I need  │ │
│ │ someone else to review this urgently│ │
│ └─────────────────────────────────────┘ │
│ * Required when "Other" is selected     │
│                                         │
│           [Cancel Request] [Go Back]    │
└─────────────────────────────────────────┘
```

## Dependencies

- **Blocks:** Story 4.4 (Store reason) - must capture before storing
- **Blocks:** Story 4.1, 4.2 (Post update, notification) - need reason for messages
- **No dependencies on other stories**

## Definition of Done

- [ ] Code implemented and reviewed
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] Manual testing in real Mattermost
- [ ] Modal validation working correctly
- [ ] Error messages are user-friendly
- [ ] Code merged to master

## Notes

**Why enumerated reasons instead of free text:**
- Structured data enables analysis ("30% wrong approver" is actionable)
- Faster for users (dropdown is quicker than typing)
- Reduces typos and inconsistent wording
- Still allows free text via "Other" for edge cases

**Why "No longer needed" is default:**
- Most common reason (based on user interview)
- Low friction: user can just confirm if it's the right reason
- Forces conscious choice if it's something else

**Why confirmation modal instead of immediate cancel:**
- Safety: prevents accidental cancellations
- Reason capture: can't cancel without providing context
- UX expectation: destructive actions should require confirmation
