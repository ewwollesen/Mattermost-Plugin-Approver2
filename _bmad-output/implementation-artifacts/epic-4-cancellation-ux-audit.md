# Epic 4: Improved Cancellation UX + Audit Trail

**Version:** 0.2.0
**Status:** Planned
**Priority:** High
**Created:** 2026-01-12

## Overview

Improve the cancellation experience for both requesters and approvers, ensuring proper notification, UI state management, and audit trail preservation. Currently, when a request is cancelled, the approver is not notified and the approve/deny buttons remain visible but non-functional, creating a poor user experience.

## Problem Statement

**Current Issues:**
1. Approvers are not notified when a request is cancelled
2. Interactive buttons remain visible in approver DMs but don't work (ghost buttons)
3. No audit trail for WHY requests were cancelled
4. Canceled requests may clutter lists without clear status indication
5. No visibility into cancellation patterns for process improvement

**User Impact:**
- Approvers waste time clicking non-functional buttons
- Confusion about request status ("Did they cancel? Is it still pending?")
- Loss of trust in the plugin's reliability
- No data to understand if cancellations indicate UX problems (wrong approver selection, etc.)

## Goals

### Primary Goals
1. **Clear Communication:** Approvers immediately understand when a request is cancelled
2. **Honest UI:** No interactive elements that don't work
3. **Audit Trail:** Preserve full history of cancellations with reasons
4. **Data Instrumentation:** Collect structured cancellation reasons for process improvement

### Success Metrics
- Zero ghost button interactions (buttons removed on cancel)
- 100% of cancelled requests have reasons recorded
- Approvers can see cancellation reason in approval record
- List views clearly show cancelled status

## User Stories

### Story 4.1: Update Approver DM Post on Cancellation
**As an** approver
**I want** the original approval request message to visually update when cancelled
**So that** I immediately see it's no longer actionable

**Acceptance Criteria:**
- Original DM post shows 🚫 status banner with "Canceled" state
- Request description is struck through (preserves original text)
- Approve/Deny buttons are removed
- Shows "Canceled by @requester at [timestamp]"
- Post update triggers immediately when cancel command executes

**Technical Notes:**
- Use Mattermost's post update API
- Store original post ID in approval record for updates
- Handle case where post no longer exists (deleted channel, etc.)

---

### Story 4.2: Send Cancellation Notification to Approver
**As an** approver
**I want** to receive a DM notification when a request I received is cancelled
**So that** I'm actively informed rather than discovering it passively

**Acceptance Criteria:**
- Approver receives DM when request is cancelled
- Notification includes: request reference code, requester name, cancellation reason
- Notification is sent even if post update succeeds (belt + suspenders)
- Error handling if DM delivery fails

**Message Format:**
```
🚫 Approval Request Canceled

Reference: TUZ-2RK
Requester: @wayne
Reason: No longer needed
Canceled: 2026-01-12 7:15 PM

The approval request you received has been canceled by the requester.
```

---

### Story 4.3: Add Cancellation Reason Dropdown
**As a** requester
**I want** to specify why I'm cancelling a request
**So that** there's a clear record and the team can improve the process

**Acceptance Criteria:**
- `/approve cancel [ID]` opens a modal (not immediate cancel)
- Modal shows dropdown with enumerated reasons:
  - No longer needed (default)
  - Wrong approver
  - Sensitive information
  - Other reason
- If "Other reason" selected, text field appears for explanation
- Cancellation requires reason selection (not optional)
- Modal has "Cancel Request" button to confirm

**UI Mock:**
```
┌─────────────────────────────────────┐
│ Cancel Approval Request             │
├─────────────────────────────────────┤
│ Reference: TUZ-2RK                  │
│                                     │
│ Reason for cancellation:            │
│ [ No longer needed        ▼ ]      │
│                                     │
│ [Cancel Request] [Go Back]          │
└─────────────────────────────────────┘
```

---

### Story 4.4: Store Cancellation Reason in Approval Record
**As a** system
**I want** to persist the cancellation reason with the approval record
**So that** it's available for audit and reporting

**Acceptance Criteria:**
- Add `CanceledReason` field to ApprovalRequest struct
- Add `CanceledAt` timestamp field
- Update KV storage to persist both fields
- Fields remain immutable once set (no editing post-cancel)
- Historical records without these fields handle gracefully (empty/null)

**Data Model Update:**
```go
type ApprovalRequest struct {
    // ... existing fields ...
    CanceledReason string    `json:"canceled_reason,omitempty"`
    CanceledAt     int64     `json:"canceled_at,omitempty"`
}
```

---

### Story 4.5: Display Cancellation Reason in Approval Details
**As a** user (requester or approver)
**I want** to see why a request was cancelled when viewing details
**So that** I have full context of what happened

**Acceptance Criteria:**
- `/approve get [ID]` shows cancellation info for cancelled requests
- Display includes: reason, who cancelled, when cancelled
- Formatted clearly in the response message
- Works for both requester and approver viewing the record

**Display Format:**
```
🚫 Approval Request (Canceled)

Reference Code: TUZ-2RK
Status: Canceled
Requester: @wayne
Approver: @john
Created: Jan 12, 2026 3:45 PM
Canceled: Jan 12, 2026 7:15 PM

Description:
[Original request description]

Cancellation:
Reason: No longer needed
Canceled by: @wayne
```

---

### Story 4.6: Show Canceled Requests in List (Sorted to Bottom)
**As a** user
**I want** to see cancelled requests in my approval list
**So that** I have visibility into the full history without losing context

**Acceptance Criteria:**
- `/approve list` includes cancelled requests for both requesters and approvers
- Cancelled requests sorted to bottom of list
- Clear status indicator (🚫 Canceled)
- Pending/Decided requests appear before Cancelled requests
- Sort order within each status group: newest first

**List Output Example:**
```
Your Approval Requests:

Pending:
🕐 TUZ-3AB - Awaiting approval from @alice - Jan 12 8:00 PM

Decided:
✅ TUZ-2CD - Approved by @bob - Jan 12 5:30 PM
❌ TUZ-1EF - Denied by @charlie - Jan 11 2:00 PM

Canceled:
🚫 TUZ-2RK - Canceled (No longer needed) - Jan 12 7:15 PM
```

---

### Story 4.7: Fix Ghost Buttons Bug
**As an** approver
**I want** non-functional buttons to be removed
**So that** the interface doesn't lie to me

**Acceptance Criteria:**
- When approval decision is made (approve/deny), buttons removed from post
- When request is cancelled, buttons removed from post
- No clickable buttons remain for completed/cancelled requests
- If button removal fails, log error but don't fail the operation

**Technical Implementation:**
- Update post using Mattermost API to remove action attachments
- Test edge cases: deleted posts, permission changes, etc.
- Add integration test for button removal

---

## Technical Design

### Data Flow: Cancellation Process

```
1. User executes: /approve cancel TUZ-2RK
2. Plugin opens confirmation modal with reason dropdown
3. User selects reason + clicks "Cancel Request"
4. Plugin validates:
   - Request exists
   - Request is in "pending" state
   - User is the requester
5. Plugin updates ApprovalRequest:
   - Status = "canceled"
   - CanceledReason = selected reason
   - CanceledAt = current timestamp
6. Plugin persists to KV store
7. Plugin updates approver DM post:
   - Add 🚫 banner
   - Strike through description
   - Remove buttons
   - Add "Canceled by @user at [time]"
8. Plugin sends cancellation DM to approver
9. Plugin responds to requester with confirmation
```

### Dependencies
- Story 4.3 must complete before 4.4 (reason capture before storage)
- Story 4.4 must complete before 4.5 (storage before display)
- Story 4.1 and 4.7 can be developed in parallel (both update posts)
- Story 4.6 depends on 4.4 (needs status in data model)

### Testing Requirements
- Unit tests for each story
- Integration test for full cancellation flow
- Manual testing in real Mattermost instance
- Test error scenarios: network failures, deleted posts, etc.

## Non-Goals (Out of Scope for v0.2.0)

- **Edit functionality:** Not adding ability to edit requests
- **Delegation:** Not adding "approve on behalf of" functionality
- **Filtering:** Not adding filter UI to `/approve list` (future v0.3.0)
- **Rerouting:** Not adding ability to change approver mid-flight
- **Bulk cancellation:** Not adding multi-select cancel
- **Analytics dashboard:** Data collection only, no reporting UI yet

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Post update fails (deleted channel) | Approver sees ghost buttons | Medium | Log error, send DM anyway, show warning to requester |
| Modal UX is clunky | Users skip selecting reason properly | Low | Make reason required, test with real users |
| Performance with large lists | Slow list rendering | Low | Already sorted, cancelled items at bottom |
| Data migration for old records | Existing records lack new fields | Low | Handle null/empty gracefully in display |

## Release Criteria

**v0.2.0 ships when:**
- [ ] All 7 stories complete and tested
- [ ] Integration tests pass
- [ ] Manual QA in real Mattermost instance
- [ ] CHANGELOG.md updated
- [ ] README.md updated (if needed)
- [ ] GitHub issue #1 closed
- [ ] Tag created: `v0.2.0`

## Future Enhancements (Post v0.2.0)

**v0.3.0 Candidates:**
- Filter controls in `/approve list` (Show: All | Pending | Decided | Canceled)
- Export cancellation reason data to CSV for analysis
- Cancellation reason analytics dashboard
- Bulk operations (cancel multiple pending requests)

**Later Versions:**
- Edit request functionality (reduces wrong-approver cancellations)
- Reroute to different approver (reduces cancel+recreate pattern)
- Approver delegation/OOO handling

## References

- GitHub Issue: [#1 - Cancelling request does not alert approver or remove dialog](https://github.com/ewwollesen/Mattermost-Plugin-Approver2/issues/1)
- Related Code: `server/command/cancel.go`
- Related Code: `server/notifications.go`
- Mattermost API Docs: [Update Post](https://api.mattermost.com/#tag/posts/operation/UpdatePost)

## Notes

**Why enumerated reasons matter:**
This isn't just UX polish. Six months from now, we can run queries:
- 43% "No longer needed" = normal usage
- 31% "Wrong approver" = approver selection UX is broken
- 18% "Sensitive info" = users don't understand what's safe to write

That's actionable product data that drives real improvements.

**Why show cancelled in lists:**
Audit trail matters. Approvers especially need to see "what happened to that request I saw this morning?" Hiding cancelled requests creates confusion. Sorting to bottom balances visibility with focus on actionable items.
