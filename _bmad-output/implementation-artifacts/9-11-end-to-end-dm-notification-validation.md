# Story 9.11: End-to-End DM Notification Validation

Status: ready-for-dev

## Story

As a user,
I want all approval DM notifications to display with proper timezones,
so that I can trust the timestamps regardless of where approval communications appear.

## Acceptance Criteria

**AC1: Approval Request DM Flow**
- Create approval via `/approve new` (any channel)
- Approver receives DM as custom component (not markdown)
- DM shows:
  - 📋 Approval Request header
  - Requester info with mentions
  - Description (full text)
  - Requested timestamp in approver's local timezone
  - Request ID
  - Approve/Deny buttons (functional)

**AC2: Outcome Notification DM Flow**
- Approver clicks Approve or Deny
- Requester receives outcome DM as custom component
- DM shows:
  - ✅ Approved or ❌ Denied header
  - Approver info
  - Decision timestamp in requester's local timezone
  - Original request (quoted)
  - Decision comment (if provided)
  - Status statement

**AC3: Cancellation Notification DM Flow**
- Requester cancels pending approval via `/approve cancel`
- Approver receives cancellation DM as custom component
- DM shows:
  - 🚫 Approval Canceled header
  - Request ID and description
  - Cancellation reason
  - Canceled timestamp in approver's local timezone
  - Requester info

**AC4: Timeout Notification DM Flow**
- Approval request times out (30+ minutes)
- Requester receives timeout DM as custom component
- DM shows:
  - ⏱️ Approval Timed Out header
  - Request ID and original description
  - Approver info
  - Timeout reason
  - Auto-canceled timestamp in requester's local timezone

**AC5: Verification Notification DM Flow**
- Requester runs `/approve verify <CODE>` with optional comment
- Approver receives verification DM as custom component
- DM shows:
  - ✅ Action Verified Complete header
  - Request ID
  - Requester info
  - Verified timestamp in approver's local timezone
  - Verification comment (if provided)

**AC6: Approver Cancellation DM Flow**
- Approver cancels approval request (Story 7.1 feature)
- Requester receives cancellation DM as custom component
- Shows cancellation by approver with timestamp in requester's timezone

**AC7: UpdatePost for Canceled Approvals**
- Approver's original DM post updates when request canceled
- Post shows 🚫 Approval Request (Canceled)
- Buttons disabled
- Canceled timestamp shown in approver's local timezone
- Uses same custom post type (updated props)

**AC8: Cross-Timezone Testing**
- User A (PST) and User B (EST) exchange approval
- All timestamps accurate in respective timezones
- No off-by-one errors, DST handled correctly
- Hover tooltips show timezone abbreviation

**AC9: Cross-Client Compatibility**
- Webapp client: Sees custom components for all DMs
- Mobile client: Sees markdown fallback
- Desktop client: Sees custom components
- No data loss across clients

**AC10: Regression Testing**
- All v1.0/v2.x DM notification behavior preserved
- Approve/Deny buttons still functional
- Post updates still work (cancellation)
- No breaking changes to existing approvals
- All existing unit tests pass
- All existing integration tests pass

**AC11: Performance**
- DM rendering fast (< 100ms)
- No memory leaks
- Multiple DMs in quick succession don't cause issues

## Tasks / Subtasks

- [ ] Prepare test environment (AC1-AC11)
  - [ ] Deploy plugin with Story 9.10 implementation to test Mattermost instance
  - [ ] Verify webapp bundle loaded in browser console
  - [ ] Verify custom post type registered for DMs
  - [ ] Create 2 test user accounts with different timezones (PST, EST)
  - [ ] Configure timezone settings for both users
  - [ ] Clear DM history to start fresh

- [ ] Test approval request DM flow (AC1)
  - [ ] User A (PST) runs `/approve new` in any channel
  - [ ] Select User B (EST) as approver, add description
  - [ ] Submit approval request
  - [ ] Switch to User B account
  - [ ] Open DMs, locate approval request from bot
  - [ ] Verify DM renders as custom component (not markdown table)
  - [ ] Inspect post in DevTools: Type="custom_approval", Props populated
  - [ ] Check notification_type="approval_request", is_dm=true in props
  - [ ] Verify 📋 Approval Request header visible
  - [ ] Verify requester info shows @UserA with display name
  - [ ] Verify full description displayed (not truncated)
  - [ ] Verify timestamp shows in EST (User B's timezone)
  - [ ] Verify Request ID displayed
  - [ ] Verify Approve/Deny buttons present and styled correctly
  - [ ] Hover over timestamp, verify timezone abbreviation tooltip

- [ ] Test outcome notification DM flow (AC2)
  - [ ] User B (EST) clicks "Approve" button in DM
  - [ ] Add optional approval note in modal
  - [ ] Confirm approval
  - [ ] Switch to User A (PST) account
  - [ ] Open DMs, locate outcome notification from bot
  - [ ] Verify DM renders as custom component
  - [ ] Inspect post: Type="custom_approval", notification_type="outcome"
  - [ ] Verify ✅ Approval Approved header visible
  - [ ] Verify approver info shows @UserB with display name
  - [ ] Verify decision timestamp shows in PST (User A's timezone)
  - [ ] Verify original request description quoted/referenced
  - [ ] Verify approval note displayed (if provided)
  - [ ] Verify status statement ("You may proceed" or similar)
  - [ ] No interactive buttons (read-only)

- [ ] Test denial outcome DM flow (AC2)
  - [ ] Create new approval: User A → User B
  - [ ] User B clicks "Deny" button
  - [ ] Add denial reason in modal
  - [ ] Confirm denial
  - [ ] User A receives outcome DM
  - [ ] Verify ❌ Approval Denied header
  - [ ] Verify denial reason displayed
  - [ ] Verify decision timestamp in PST (User A's timezone)

- [ ] Test cancellation notification DM flow (AC3)
  - [ ] Create approval: User A → User B
  - [ ] User A runs `/approve cancel <CODE>`
  - [ ] Provide cancellation reason
  - [ ] User B receives cancellation DM
  - [ ] Verify DM renders as custom component
  - [ ] Inspect post: notification_type="cancellation"
  - [ ] Verify 🚫 Approval Canceled header
  - [ ] Verify Request ID and description shown
  - [ ] Verify cancellation reason displayed
  - [ ] Verify canceled timestamp in EST (User B's timezone)
  - [ ] Verify requester info (@UserA)

- [ ] Test timeout notification DM flow (AC4)
  - [ ] Create approval with short timeout (or manually trigger via admin command)
  - [ ] Wait for timeout (30 minutes or trigger immediately)
  - [ ] User A receives timeout DM
  - [ ] Verify DM renders as custom component
  - [ ] Inspect post: notification_type="timeout"
  - [ ] Verify ⏱️ Approval Timed Out header
  - [ ] Verify Request ID and original description
  - [ ] Verify approver info (@UserB)
  - [ ] Verify timeout reason/explanation
  - [ ] Verify auto-canceled timestamp in PST (User A's timezone)

- [ ] Test verification notification DM flow (AC5)
  - [ ] Create approval: User A → User B
  - [ ] User B approves
  - [ ] User A runs `/approve verify <CODE>` with comment
  - [ ] User B receives verification DM
  - [ ] Verify DM renders as custom component
  - [ ] Inspect post: notification_type="verification"
  - [ ] Verify ✅ Action Verified Complete header
  - [ ] Verify Request ID shown
  - [ ] Verify requester info (@UserA)
  - [ ] Verify verified timestamp in EST (User B's timezone)
  - [ ] Verify verification comment displayed

- [ ] Test approver cancellation DM flow (AC6, Story 7.1)
  - [ ] Create approval: User A → User B
  - [ ] User B cancels the approval request (approver-initiated cancellation)
  - [ ] User A receives cancellation notification DM
  - [ ] Verify DM renders as custom component
  - [ ] Verify shows cancellation by approver (@UserB)
  - [ ] Verify timestamp in PST (User A's timezone)
  - [ ] Verify cancellation reason (if provided)

- [ ] Test UpdatePost for canceled approvals (AC7)
  - [ ] Create approval: User A → User B
  - [ ] User B receives approval request DM
  - [ ] Note the post ID of User B's approval request DM
  - [ ] User A cancels the approval via `/approve cancel`
  - [ ] Switch to User B account
  - [ ] Verify original DM post UPDATED (same post ID, not new post)
  - [ ] Verify post shows 🚫 Approval Request (Canceled)
  - [ ] Verify Approve/Deny buttons disabled or removed
  - [ ] Verify canceled timestamp shown in EST (User B's timezone)
  - [ ] Inspect props: status updated to "canceled"

- [ ] Test cross-timezone accuracy (AC8)
  - [ ] User A (PST): Create approval at specific time (e.g., 3:00 PM PST)
  - [ ] User A: Note timestamp in DM (should show "3:00 PM PST")
  - [ ] User B (EST): View same approval request
  - [ ] User B: Note timestamp in DM (should show "6:00 PM EST")
  - [ ] Verify timestamps represent same moment (3 hour difference)
  - [ ] Hover over both timestamps, verify timezone abbreviations
  - [ ] User B approves at specific time (e.g., 6:30 PM EST)
  - [ ] User A views outcome DM
  - [ ] Verify decision timestamp shows "3:30 PM PST" (same moment)
  - [ ] Test during DST transition if possible (verify no off-by-one errors)

- [ ] Test cross-client compatibility (AC9)
  - [ ] Webapp client (Chrome/Firefox): Verify custom components render for all DM types
  - [ ] Desktop client (Mattermost Desktop): Verify custom components render
  - [ ] Mobile client (iOS/Android if available): Check if markdown fallback shown
  - [ ] API client (curl/Postman): Verify post.Message has readable markdown
  - [ ] For each DM notification type (6 total):
    - [ ] Webapp: Custom component renders
    - [ ] Mobile: Markdown fallback visible
    - [ ] No data loss: All information present in both views

- [ ] Test performance (AC11)
  - [ ] Create 10+ approval requests quickly (User A → User B)
  - [ ] User B opens DM thread, scroll through all requests
  - [ ] Measure DM rendering time (Chrome DevTools Performance tab)
  - [ ] Verify < 100ms per DM post
  - [ ] Check for jank/lag during scrolling
  - [ ] Open/close DM thread multiple times
  - [ ] Check Chrome DevTools Memory tab for leaks
  - [ ] Verify component cleanup on unmount
  - [ ] Create multiple approvals in quick succession (stress test)
  - [ ] Verify no UI freezing or errors

- [ ] Regression testing (AC10)
  - [ ] Run all existing approval plugin tests (make test)
  - [ ] Verify all unit tests pass (server + webapp)
  - [ ] Test v1.0 behavior: Create approval in non-playbook channel
  - [ ] Verify DM notifications still work with custom components
  - [ ] Test Approve/Deny button functionality
  - [ ] Verify approval record created correctly
  - [ ] Test `/approve list` command shows all approvals
  - [ ] Test `/approve show <CODE>` displays correct info
  - [ ] Verify no console errors or warnings in browser
  - [ ] Check server logs for errors during testing

- [ ] Browser compatibility testing (optional but recommended)
  - [ ] Chrome/Chromium: Verify custom DM components render
  - [ ] Firefox: Verify custom DM components render
  - [ ] Safari: Verify custom DM components render
  - [ ] Edge: Verify custom DM components render

- [ ] Document test results
  - [ ] Create comprehensive test report with screenshots
  - [ ] Document all 6 DM notification types tested
  - [ ] Document timezone accuracy results
  - [ ] Document browser/client compatibility matrix
  - [ ] Document any issues found and resolutions
  - [ ] Add to Dev Agent Record section
  - [ ] Note Epic 9 completion status

- [ ] Final validation and sign-off
  - [ ] Confirm all 11 acceptance criteria passed
  - [ ] Verify GitHub Issue #3 fully resolved (timezone display working)
  - [ ] Confirm Epic 9 success metrics achieved
  - [ ] Document any known limitations or future enhancements
  - [ ] Prepare for Epic 9 retrospective (optional)

## Dev Notes

### Architecture Requirements

**Epic 9 Final Validation Story**
This is the **final story of Epic 9** - comprehensive end-to-end validation of ALL DM notification types with custom post types and timezone-aware timestamps. This story validates that Story 9.10 implementation works correctly across all approval flows.

**Story Purpose:**
- Validate all 6 DM notification types render as custom components
- Verify timezone accuracy for all timestamps in DMs
- Ensure interactive buttons (Approve/Deny) still functional
- Confirm markdown fallback works for non-webapp clients
- Validate performance and cross-client compatibility
- **Complete GitHub Issue #3 resolution** (timezone display in all approval communications)

**Why This Story Matters:**
- **Epic 9 Completion**: Final validation before closing Epic 9
- **Production Readiness**: Ensures v3.0.0 ready for deployment
- **User Confidence**: All approval timestamps now display correctly
- **Quality Assurance**: Comprehensive testing prevents production issues

**Testing Scope (6 DM Notification Types):**
1. **Approval Request DM** (SendApprovalRequestDM) - Most critical: has interactive buttons
2. **Outcome Notification DM** (SendOutcomeNotificationDM) - Most frequent: requester receives decision
3. **Cancellation Notification DM** (SendCancellationNotificationDM) - Approver notified of requester cancellation
4. **Timeout Notification DM** (SendTimeoutNotificationDM) - Requester notified of timeout
5. **Verification Notification DM** (SendVerificationNotificationDM) - Approver notified of verification
6. **Approver Cancellation DM** (SendRequesterCancellationNotificationDM) - Requester notified of approver cancellation

### Component Implementation Details

**Testing Strategy Overview:**

This is a **manual end-to-end validation story**, similar to Story 9.9 but focused on DM notifications instead of playbook posts.

**Testing Flow:**
```
1. Deploy Plugin (with Story 9.10 implementation)
   └─ make
   └─ make deploy
   └─ Restart Mattermost
   └─ Verify webapp loaded in browser console

2. Verify Custom Post Type for DMs
   └─ Open browser DevTools console
   └─ Look for: "Registered custom post type: custom_approval"
   └─ Create approval, check DM post.Type in DevTools

3. Test All 6 DM Notification Types
   └─ Approval Request (AC1)
   └─ Outcome Notification (AC2)
   └─ Cancellation Notification (AC3)
   └─ Timeout Notification (AC4)
   └─ Verification Notification (AC5)
   └─ Approver Cancellation (AC6)

4. Test UpdatePost for Cancellation (AC7)
   └─ Original DM post updates when approval canceled
   └─ Buttons disabled
   └─ Props updated with new status

5. Test Cross-Timezone Accuracy (AC8)
   └─ User A (PST) creates approval
   └─ User B (EST) views request
   └─ Compare timestamps (3 hour difference)
   └─ Verify both represent same moment

6. Test Cross-Client Compatibility (AC9)
   └─ Webapp: Custom components
   └─ Mobile: Markdown fallback
   └─ Desktop: Custom components
   └─ API: Markdown in post.Message

7. Test Performance (AC11)
   └─ Create 10+ DMs
   └─ Measure render time
   └─ Check for memory leaks

8. Run Regression Tests (AC10)
   └─ make test (all tests pass)
   └─ Verify v1.0/v2.x behavior preserved
```

**DevTools Inspection for DMs:**

```javascript
// In browser console, inspect DM post:
const postElement = document.querySelector('[data-post-id="DM_POST_ID"]');

// Use React DevTools to inspect props:
// 1. Install React DevTools browser extension
// 2. Find Post component in React tree
// 3. Inspect props:
//    - post.Type === "custom_approval"
//    - post.Props.notification_type === "approval_request" (or other type)
//    - post.Props.is_dm === true
//    - post.Props.created_at === 1705593000000 (number, not string)

// Or use Mattermost's Redux store:
window.store.getState().entities.posts.posts['DM_POST_ID']
```

**Key Validation Points:**

1. **Custom Component Rendering:**
   - All 6 DM types render as ApprovalPost component (not markdown)
   - is_dm prop detected correctly
   - Layout adjusted for DM context (more verbose than playbook)

2. **Timezone Accuracy:**
   - Timestamps convert to user's timezone
   - PST vs EST difference verified (3 hours)
   - Hover shows timezone abbreviation
   - No off-by-one errors

3. **Interactive Elements:**
   - Approve/Deny buttons visible and functional in approval request DMs
   - Buttons trigger correct modals
   - Decision recorded correctly
   - Other DM types have no buttons (read-only)

4. **Post Updates:**
   - Cancellation updates original DM post (same post ID)
   - Props updated with new status
   - Buttons disabled after cancellation
   - Timestamp shows canceled time in user's timezone

5. **Markdown Fallback:**
   - post.Message contains readable markdown for all DM types
   - Mobile clients show markdown fallback
   - No data loss in fallback

6. **Performance:**
   - DM rendering < 100ms
   - No memory leaks
   - Smooth scrolling with multiple DMs
   - No UI freezing

### Library & Framework Requirements

**No New Dependencies:**
This is a validation story, no code changes required (assuming Story 9.10 implemented correctly).

**Testing Tools Required:**
- Browser DevTools (Chrome/Firefox)
- React DevTools browser extension
- Mattermost test instance (local or staging)
- 2+ test user accounts with different timezones

### File Structure Requirements

**No Files to Create or Modify:**
This is a validation story, not implementation.

**Optional: Test Report Document:**
- Create `_bmad-output/testing/epic-9-dm-test-report.md`
- Include screenshots of all 6 DM notification types
- Document timezone accuracy results
- Document cross-client compatibility
- Document performance metrics
- Note any issues found

### Previous Story Intelligence

**Critical Discoveries from Story 9.9 (Playbook Post Validation):**

1. **Custom Post Type Works Perfectly:**
   - Story 9.9 validated custom post type for playbook channel posts
   - All tests passed (568 server, 59 webapp)
   - User feedback: "looks so much better!"
   - **For Story 9.11**: Apply same validation approach to DM notifications

2. **Timezone Display Accurate:**
   - GitHub Issue #3 resolved for playbook posts
   - Timestamps convert to user's timezone correctly
   - Hover shows timezone abbreviation
   - **For Story 9.11**: Verify same accuracy for DM timestamps

3. **Markdown Fallback Validated:**
   - Mobile clients see markdown tables
   - No data loss across client types
   - **For Story 9.11**: Verify markdown fallback for DMs

4. **Performance Acceptable:**
   - Post rendering < 100ms
   - No memory leaks
   - Smooth scrolling
   - **For Story 9.11**: Verify same performance for DMs

5. **Testing Methodology:**
   - Manual testing with 2 users, different timezones
   - DevTools inspection of post.Type and post.Props
   - Cross-browser testing (Chrome, Firefox, Safari, Edge)
   - **For Story 9.11**: Use identical methodology for DM testing

**Implementation Patterns from Story 9.10 (DM Conversion):**

Story 9.10 implemented the DM notification conversion:

```go
// From notifications/dm.go (Story 9.10)
func SendApprovalRequestDM(api plugin.API, botUserID string, record *approval.ApprovalRecord) (string, error) {
    // Format markdown message (fallback)
    message := formatApprovalRequestMessage(record)

    // Format props for custom post type
    props := FormatApprovalPropsForDM(record, "approval_request")

    post := &model.Post{
        UserId:    botUserID,
        ChannelId: channelID,
        Message:   message, // Markdown fallback
        Type:      "custom_approval", // Custom post type
        Props:     props, // Approval data + is_dm=true + notification_type
    }

    // Preserve interactive buttons (if approval_request)
    if record.Status == "pending" {
        post.Props["attachments"] = []any{
            // Approve/Deny buttons
        }
    }

    createdPost, err := api.CreatePost(post)
    return createdPost.Id, err
}
```

**For Story 9.11, verify this implementation works correctly:**
1. Post.Type = "custom_approval" ✓
2. Props populated with is_dm=true and notification_type ✓
3. Timestamps are int64 (Unix millis) ✓
4. Markdown fallback in post.Message ✓
5. Interactive buttons preserved (if applicable) ✓

### Git Intelligence Summary

**Recent Commits (Last 5):**

1. **bf000fe: Fix: GitHub Issue #2 - Replace Playbooks API with markdown tables**
   - **Relevance**: Markdown formatting for fallback messages
   - **For Story 9.11**: Verify markdown fallback in DMs

2. **53c03a3: Story 8.6: Error Handling and Graceful Fallback**
   - **Relevance**: Graceful degradation when notifications fail
   - **For Story 9.11**: Verify DM notifications robust

3. **48954c5: Story 8.2: Data Model Extension for Playbook Metadata**
   - **Relevance**: Playbook context in approval records
   - **For Story 9.11**: Verify playbook metadata displayed in DMs (if applicable)

**Testing Patterns from Previous Stories:**
- Story 9.9: Manual end-to-end validation with 2 users
- Story 9.9: DevTools inspection of post.Type and post.Props
- Story 9.9: Cross-timezone testing (PST vs EST)
- **For Story 9.11**: Apply identical patterns to DM validation

### Project Structure Context

**Test Scenario Organization:**

```
Test User Setup:
├── User A (PST timezone)
│   ├── Creates approval requests
│   ├── Receives outcome notifications
│   ├── Receives timeout notifications
│   └── Runs /approve verify
├── User B (EST timezone)
│   ├── Receives approval request DMs
│   ├── Approves/denies requests
│   ├── Receives cancellation notifications
│   └── Receives verification notifications

Test Flows (6 DM Types):
1. Approval Request: A → B (B receives request DM)
2. Outcome: B approves → A (A receives outcome DM)
3. Cancellation: A cancels → B (B receives cancellation DM)
4. Timeout: Request times out → A (A receives timeout DM)
5. Verification: A verifies → B (B receives verification DM)
6. Approver Cancellation: B cancels → A (A receives cancellation DM)
```

**Testing Environment Requirements:**

```
Mattermost Test Instance:
├── Plugin deployed with Story 9.10 implementation
├── Webapp bundle loaded (verify in console)
├── Custom post type registered (verify in console)
├── 2 test users configured:
│   ├── User A: Timezone set to PST (US/Pacific)
│   └── User B: Timezone set to EST (US/Eastern)
└── Clear DM history for clean testing

Browser Setup:
├── Chrome/Firefox with DevTools
├── React DevTools extension installed
├── Console open to verify post.Type and props
└── Network tab to monitor API calls (optional)
```

### References

- [Source: Epic 9 - Story 9.11 Acceptance Criteria] - AC1-AC11 requirements
- [Source: Story 9.9 Dev Notes] - End-to-end testing methodology for playbook posts
- [Source: Story 9.10 Dev Notes] - DM notification conversion implementation details
- [Source: Story 9.10 - AC2 Props Schema] - DM props structure with is_dm and notification_type
- [Source: server/notifications/dm.go:1-100] - DM send function implementations (Story 9.10)
- [Mattermost Developer Docs] - Plugin testing guide
- [React DevTools Documentation] - Component inspection

### Critical Gotchas

**AVOID THESE TESTING MISTAKES:**

1. **DON'T Skip Timezone Configuration:**
   - ❌ WRONG: Test with both users in same timezone
   - ✅ CORRECT: Configure User A (PST), User B (EST)
   - **Impact**: Can't verify timezone conversion working

2. **DON'T Test with Single User:**
   - ❌ WRONG: Create approval with same user as approver
   - ✅ CORRECT: Use 2 separate user accounts
   - **Impact**: Can't test DM delivery and cross-timezone display

3. **DON'T Assume Webapp Loaded:**
   - ❌ WRONG: Start testing without checking console
   - ✅ CORRECT: Verify "Approval Plugin Webapp loaded" in console
   - **Impact**: Tests fail silently if webapp not loaded

4. **DON'T Skip DevTools Inspection:**
   - ❌ WRONG: Only verify visual rendering
   - ✅ CORRECT: Inspect post.Type, post.Props in DevTools
   - **Impact**: Miss incorrect props structure (timestamps as strings, etc.)

5. **DON'T Test Only Approval Request DM:**
   - ❌ WRONG: Only test one DM type
   - ✅ CORRECT: Test all 6 DM notification types
   - **Impact**: Miss issues in outcome, cancellation, timeout, etc.

6. **DON'T Forget to Test Post Updates:**
   - ❌ WRONG: Skip cancellation flow (AC7)
   - ✅ CORRECT: Verify original DM post updates when approval canceled
   - **Impact**: Miss critical regression (duplicate posts, buttons not disabled)

7. **DON'T Skip Performance Testing:**
   - ❌ WRONG: Only test with 1-2 DMs
   - ✅ CORRECT: Create 10+ DMs, measure render time
   - **Impact**: Miss memory leaks and performance degradation

8. **DON'T Test Only Happy Path:**
   - ❌ WRONG: Only test approval flow
   - ✅ CORRECT: Test denial, cancellation, timeout flows
   - **Impact**: Miss edge cases and error scenarios

**Common Testing Errors:**
- "DM shows markdown, not custom component": Webapp not loaded or Story 9.10 not implemented
- "Timestamp shows UTC": User timezone not configured in Mattermost settings
- "Buttons not working": Story 9.10 implementation didn't preserve attachments
- "Post duplicated on cancellation": UpdatePost using CreatePost instead of UpdatePost
- "Props show undefined fields": Story 9.10 implementation incomplete

### Implementation Order

**Recommended Testing Sequence:**

**Phase 1: Environment Setup (30 minutes)**
1. Deploy plugin with Story 9.10 implementation
2. Restart Mattermost server
3. Verify webapp loaded (browser console)
4. Create/configure 2 test users (PST, EST)
5. Clear DM history

**Phase 2: Basic DM Flows (1 hour)**
6. Test approval request DM (AC1) - Most critical
7. Test outcome notification DM (AC2) - Most frequent
8. Test denial outcome DM (AC2 variant)
9. Verify DevTools inspection for all 3

**Phase 3: Advanced DM Flows (1 hour)**
10. Test cancellation notification DM (AC3)
11. Test timeout notification DM (AC4)
12. Test verification notification DM (AC5)
13. Test approver cancellation DM (AC6)
14. Verify DevTools inspection for all 4

**Phase 4: Post Update Testing (30 minutes)**
15. Test UpdatePost for canceled approvals (AC7)
16. Verify original post ID preserved
17. Verify buttons disabled
18. Verify props updated

**Phase 5: Cross-Timezone Validation (30 minutes)**
19. Test PST vs EST timestamp display (AC8)
20. Verify 3-hour difference
21. Verify hover tooltips show timezone
22. Test decision timestamp accuracy

**Phase 6: Cross-Client Compatibility (1 hour)**
23. Test webapp client (Chrome, Firefox, Safari, Edge)
24. Test desktop client
25. Test mobile client (if available)
26. Verify markdown fallback via API

**Phase 7: Performance Testing (30 minutes)**
27. Create 10+ approvals rapidly
28. Measure DM render time
29. Check for memory leaks
30. Test multiple DMs in quick succession

**Phase 8: Regression Testing (30 minutes)**
31. Run `make test` (all tests pass)
32. Test v1.0 approval flows
33. Test v2.x playbook detection
34. Verify no console errors

**Phase 9: Documentation (30 minutes)**
35. Create test report with screenshots
36. Document all 6 DM types tested
37. Document timezone accuracy results
38. Document any issues found

**Total Estimated Time: 6-7 hours (1 day)** (per Epic 9 estimate)

**Why This Order:**
1. **Setup First**: Must have working environment
2. **Basic Flows First**: Most critical (approval request, outcome)
3. **Advanced Flows Second**: Less common (cancellation, timeout, verification)
4. **Post Update Third**: Tests integration with existing code
5. **Cross-Timezone Fourth**: Validates primary Epic 9 goal
6. **Cross-Client Fifth**: Ensures broad compatibility
7. **Performance Sixth**: Validates production readiness
8. **Regression Last**: Ensures nothing broken
9. **Documentation Final**: Captures all results

### Performance Considerations

**Performance Targets:**
- DM render time: < 100ms per DM (same as playbook posts, Story 9.9)
- No memory leaks (heap size stable)
- Smooth scrolling through DM thread
- No UI freezing with multiple DMs

**How to Measure:**

```javascript
// Chrome DevTools Performance tab:
// 1. Open DM thread with 10+ approval posts
// 2. Start recording (Performance tab)
// 3. Scroll through DM thread
// 4. Stop recording
// 5. Look for "ApprovalPost" in flame graph
// 6. Measure render time per component (should be < 100ms)

// Memory tab:
// 1. Take heap snapshot
// 2. Create 10 approvals (DMs to User B)
// 3. Open/close DM thread 10 times
// 4. Take another heap snapshot
// 5. Compare heap sizes (should be similar, no growth)

// Visual lag test:
// 1. Create 20+ approvals rapidly
// 2. Open DM thread, scroll up/down quickly
// 3. Watch for jank/lag (should be smooth 60 FPS)
```

**Performance Validation:**
- Story 9.9 validated playbook post performance (< 100ms) ✓
- Story 9.11 validates DM post performance (< 100ms) ✓
- Same ApprovalPost component used for both ✓
- No performance degradation expected ✓

### Architecture Compliance

**Aligns with Epic 9 Goals:**
- ✅ Timezone Support (Goal 1.2): All DM timestamps display in user's local timezone
- ✅ Custom Post Components (Goal 1.3): All approval posts (playbook + DM) use webapp components
- ✅ DM Notification Conversion (Goal 1.6): Phase 4 validated
- ✅ No Breaking Changes (Goal 1.5): Markdown fallback ensures backward compatibility

**Validates Epic 9 Success Metrics:**
- ✅ All timestamps display in user's local timezone (playbook + DM validated)
- ✅ Playbook posts render as custom components (Story 9.9 validated)
- ✅ DM posts render as custom components (Story 9.11 validates)
- ✅ No functionality lost from markdown format (fallback validated)
- ✅ Foundation ready for future enhancements (component architecture proven)

**Epic 9 Completion Criteria:**
This story completes Epic 9. After Story 9.11 passes:
- ✅ All 11 stories in Epic 9 complete
- ✅ GitHub Issue #3 fully resolved (timezone display working everywhere)
- ✅ v3.0.0 ready for release
- ✅ Optional: Run Epic 9 retrospective

### Data Contract Validation

**Server → Webapp Contract for DMs (Story 9.10):**

```
Server (Go) → Mattermost → Webapp (TypeScript)

post.Type: "custom_approval"
post.Props: {
  // Standard approval fields
  approval_code: string
  approval_status: string
  requester_username: string
  requester_display_name: string
  approver_username: string
  approver_display_name: string
  description: string
  created_at: number              ← MUST be int64 Unix millis
  decided_at?: number             ← Optional, int64 Unix millis
  decision_comment?: string       ← Optional

  // DM-specific fields (NEW in Story 9.10)
  notification_type: string       ← "approval_request" | "outcome" | "cancellation" | "timeout" | "verification"
  is_dm: boolean                  ← true for DMs

  // Interactive buttons (approval_request only)
  attachments?: Array<{
    actions: Array<Button>
  }>
}
post.Message: string ← Markdown fallback (REQUIRED)
```

**Validation Checklist for Each DM Type:**

For each of the 6 DM notification types, verify:
- [ ] post.Type === "custom_approval"
- [ ] post.Props.is_dm === true
- [ ] post.Props.notification_type is correct ("approval_request", etc.)
- [ ] post.Props.created_at is number (int64), not string
- [ ] post.Props.decided_at is number or undefined (not 0 or "")
- [ ] post.Message contains readable markdown
- [ ] Webapp renders as ApprovalPost component (not markdown)
- [ ] Timestamp displays in user's timezone
- [ ] Hover shows timezone abbreviation

**DevTools Validation Example:**

```javascript
// Inspect DM post in DevTools
const post = window.store.getState().entities.posts.posts['POST_ID'];

// Verify contract
console.assert(post.type === 'custom_approval', 'Type should be custom_approval');
console.assert(post.props.is_dm === true, 'is_dm should be true');
console.assert(typeof post.props.created_at === 'number', 'created_at should be number');
console.assert(['approval_request', 'outcome', 'cancellation', 'timeout', 'verification'].includes(post.props.notification_type), 'notification_type should be valid');
console.assert(post.message.length > 0, 'message should have markdown fallback');

// If approval_request, verify buttons
if (post.props.notification_type === 'approval_request') {
    console.assert(post.props.attachments && post.props.attachments.length > 0, 'Approval request should have buttons');
}
```

### Wayne's Feedback Integration

**Critical User Requirements:**
1. **"Stick to Mattermost theme"** - Maintained: ApprovalPost uses Mattermost styles ✓
2. **"Minimize screen real estate"** - DMs can be more verbose (1:1 context appropriate) ✓
3. **"No backward compatibility needed"** - Old approvals stay markdown (acceptable) ✓
4. **GitHub Issue #3: Timezone display** - **PRIMARY VALIDATION**: DM timestamps in local timezone

**GitHub Issue #3 Final Resolution Validation:**

This story is the **final validation** that GitHub Issue #3 is fully resolved:

- Story 9.9: ✅ Playbook channel posts show local timezone
- Story 9.10: ✅ DM notifications converted to custom post type
- Story 9.11: ✅ **VALIDATE DM timestamps show local timezone** ← YOU ARE HERE
- After Story 9.11 passes: **GitHub Issue #3 FULLY RESOLVED** 🎯

**Testing Checklist for GitHub Issue #3:**
- [ ] User A (PST): Create approval, note timestamp
- [ ] User B (EST): View same approval request, note timestamp
- [ ] Verify 3-hour difference (PST → EST)
- [ ] User B approves, note decision timestamp
- [ ] User A views outcome DM, verify decision timestamp in PST
- [ ] Verify hover shows timezone abbreviation (PST, EST)
- [ ] No off-by-one errors
- [ ] DST transitions handled correctly (if testable)
- **If all pass**: GitHub Issue #3 RESOLVED ✅

### Type Definitions

**TypeScript Types (Webapp - from Story 9.6):**

```typescript
// From webapp/src/types/approval.ts
interface ApprovalPostProps {
    post: Post; // Mattermost Post object
}

interface ApprovalPostData {
    // Standard fields
    approval_code: string;
    approval_status: 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout';
    requester_username: string;
    requester_display_name: string;
    approver_username: string;
    approver_display_name: string;
    description: string;
    created_at: number;        // Unix millis
    decided_at?: number;       // Unix millis
    decision_comment?: string;

    // DM-specific (Story 9.10)
    notification_type?: 'approval_request' | 'outcome' | 'cancellation' | 'timeout' | 'verification';
    is_dm?: boolean;

    // Playbook context (optional)
    playbook_id?: string;
    playbook_title?: string;
    playbook_channel_id?: string;
}
```

**For Story 9.11, validate webapp receives correct types:**
- created_at is number (not string)
- decided_at is number or undefined (not 0 or "")
- notification_type is one of 5 valid values
- is_dm is true for all DM posts

### DM vs Playbook Context

**Differentiation Validation:**

| Aspect | Playbook Channel Posts (9.9) | DM Notifications (9.11) |
|--------|------------------------------|-------------------------|
| **Purpose** | Team visibility, audit trail | Private 1:1 communication |
| **Tested in** | Story 9.9 ✅ | Story 9.11 ← YOU ARE HERE |
| **Layout** | Compact (minimize screen space) | More verbose (full context) |
| **Description** | Truncated to 80 chars | Full description (no truncation) |
| **Props Flag** | is_dm: undefined or false | is_dm: true |
| **notification_type** | Not used | "approval_request", "outcome", etc. |
| **Interactive Buttons** | No buttons (read-only) | Approve/Deny buttons (approval_request only) |
| **Post Updates** | UpdatePost on status change | Original post + separate outcome notification |
| **Validation Status** | ✅ Passed (Story 9.9) | 🔍 Testing (Story 9.11) |

**Component Rendering Validation:**

```typescript
// webapp/src/components/ApprovalPost.tsx
// Verify this logic works correctly in Story 9.11

const ApprovalPost: React.FC<ApprovalPostProps> = ({post}) => {
    const isDM = post.props.is_dm === true;
    const notificationType = post.props.notification_type;

    if (isDM) {
        // DM Layout: More verbose, full context
        // Validate in Story 9.11:
        // - Full description displayed (not truncated)
        // - notification_type-specific rendering
        // - Appropriate headers (📋, ✅, ❌, 🚫, ⏱️)
        return <DMApprovalPostLayout />;
    }

    // Playbook Channel Layout: Compact
    // Validated in Story 9.9 ✅
    return <PlaybookApprovalPostLayout />;
};
```

**Testing Differentiation:**
- Story 9.9: ✅ Validated playbook channel posts (compact layout)
- Story 9.11: 🔍 Validating DM notifications (verbose layout)
- After Story 9.11: Both contexts validated, Epic 9 complete

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

(To be populated during testing)

### Completion Notes List

(To be populated during testing)

**Testing Checklist (to be completed during Story 9.11 execution):**

#### Environment Setup
- [ ] Plugin deployed with Story 9.10 implementation
- [ ] Webapp loaded (verified in console)
- [ ] Custom post type registered (verified in console)
- [ ] User A configured (PST timezone)
- [ ] User B configured (EST timezone)
- [ ] DM history cleared

#### AC1: Approval Request DM Flow
- [ ] DM renders as custom component (not markdown)
- [ ] Type="custom_approval", notification_type="approval_request", is_dm=true
- [ ] 📋 Approval Request header visible
- [ ] Requester info with @mention
- [ ] Full description displayed
- [ ] Timestamp in EST (User B's timezone)
- [ ] Request ID displayed
- [ ] Approve/Deny buttons present and functional

#### AC2: Outcome Notification DM Flow
- [ ] Approval outcome DM renders as custom component
- [ ] ✅ Approved header (or ❌ Denied)
- [ ] Approver info displayed
- [ ] Decision timestamp in PST (User A's timezone)
- [ ] Original request referenced
- [ ] Decision comment displayed (if provided)
- [ ] No interactive buttons (read-only)

#### AC3: Cancellation Notification DM Flow
- [ ] Cancellation DM renders as custom component
- [ ] 🚫 Approval Canceled header
- [ ] Request ID and description shown
- [ ] Cancellation reason displayed
- [ ] Canceled timestamp in EST (User B's timezone)
- [ ] Requester info displayed

#### AC4: Timeout Notification DM Flow
- [ ] Timeout DM renders as custom component
- [ ] ⏱️ Approval Timed Out header
- [ ] Request ID and description shown
- [ ] Approver info displayed
- [ ] Timeout reason shown
- [ ] Auto-canceled timestamp in PST (User A's timezone)

#### AC5: Verification Notification DM Flow
- [ ] Verification DM renders as custom component
- [ ] ✅ Action Verified Complete header
- [ ] Request ID shown
- [ ] Requester info displayed
- [ ] Verified timestamp in EST (User B's timezone)
- [ ] Verification comment displayed (if provided)

#### AC6: Approver Cancellation DM Flow
- [ ] Approver cancellation DM renders as custom component
- [ ] Cancellation by approver indicated
- [ ] Timestamp in PST (User A's timezone)
- [ ] Cancellation reason displayed

#### AC7: UpdatePost for Canceled Approvals
- [ ] Original DM post updates (same post ID)
- [ ] 🚫 Approval Request (Canceled) shown
- [ ] Approve/Deny buttons disabled/removed
- [ ] Canceled timestamp shown in EST (User B's timezone)
- [ ] Props updated with status="canceled"

#### AC8: Cross-Timezone Testing
- [ ] PST timestamp verified (User A)
- [ ] EST timestamp verified (User B)
- [ ] 3-hour difference confirmed
- [ ] Hover tooltips show timezone abbreviation
- [ ] No off-by-one errors
- [ ] DST handled correctly (if testable)

#### AC9: Cross-Client Compatibility
- [ ] Webapp: Custom components render (all 6 DM types)
- [ ] Desktop: Custom components render
- [ ] Mobile: Markdown fallback visible
- [ ] API: post.Message has readable markdown
- [ ] No data loss across clients

#### AC10: Regression Testing
- [ ] make test: All tests pass
- [ ] v1.0 behavior preserved
- [ ] Approve/Deny buttons functional
- [ ] Post updates work (cancellation)
- [ ] No breaking changes
- [ ] No console errors

#### AC11: Performance
- [ ] DM rendering < 100ms (measured)
- [ ] No memory leaks (heap snapshot stable)
- [ ] Smooth scrolling with 10+ DMs
- [ ] No UI freezing

#### Browser Compatibility
- [ ] Chrome: Custom components render
- [ ] Firefox: Custom components render
- [ ] Safari: Custom components render
- [ ] Edge: Custom components render

#### Final Validation
- [ ] All 11 acceptance criteria passed
- [ ] GitHub Issue #3 fully resolved
- [ ] Epic 9 success metrics achieved
- [ ] Test report documented with screenshots
- [ ] Epic 9 ready for completion

### File List

**No Files to Modify:**
This is a validation story with no code changes (assuming Story 9.10 implemented correctly).

**Optional Files to Create:**
- `_bmad-output/testing/epic-9-dm-test-report.md` (test results documentation)
- Screenshots of all 6 DM notification types
- Performance metrics (render times, memory usage)
- Cross-client compatibility matrix
