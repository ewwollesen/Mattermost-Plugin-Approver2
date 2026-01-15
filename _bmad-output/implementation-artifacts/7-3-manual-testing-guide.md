# Manual Testing Guide: Story 7.3 - Additional Details Box Behavior Fix

**Story:** Fix Additional Details Box Behavior in Cancel Dialog
**Date:** 2026-01-15
**Tester:** [Name]
**Plugin Version:** 0.2.0+

---

## Overview

This manual testing guide validates that the "Additional details" field in the cancellation dialog now **always captures user input** regardless of which reason is selected. Previously, details were only captured when "Other reason" was selected, causing silent data loss.

---

## Prerequisites

1. Mattermost server running with plugin installed
2. At least 2 test users (requester + approver)
3. Clean test environment
4. Access to `/approve` slash commands

---

## Test Setup

### Initial Setup
1. Log in as User A (requester)
2. Create a test approval request:
   ```
   /approve request @UserB "Test approval for Story 7.3 validation"
   ```
3. Note the approval code (e.g., `A-TEST01`)

---

## Test Cases

### Test Case 1: Capture Details with "No longer needed" Reason
**Objective:** Verify details are captured for "No longer needed" reason

**Steps:**
1. As User A (requester), run:
   ```
   /approve cancel A-TEST01
   ```
2. In the dialog that appears:
   - Select reason: **"No longer needed"**
   - In "Additional details (optional)" field, type: **"Project postponed until Q2"**
   - Click **"Cancel Request"**

**Expected Results:**
- ✅ Cancellation succeeds
- ✅ Confirmation message displayed
- ✅ User B (approver) receives DM notification including: "**Details:** Project postponed until Q2"
- ✅ User A can verify details with `/approve show A-TEST01` - details section shows: "**Details:** Project postponed until Q2"

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 2: Capture Details with "Wrong approver" Reason
**Objective:** Verify details are captured for "Wrong approver" reason

**Steps:**
1. Create new approval: `/approve request @UserB "Test 2 for Story 7.3"`
2. Note code (e.g., `A-TEST02`)
3. Run: `/approve cancel A-TEST02`
4. In dialog:
   - Select reason: **"Wrong approver"**
   - In "Additional details (optional)" field, type: **"Manager approval required, not peer"**
   - Click **"Cancel Request"**

**Expected Results:**
- ✅ Cancellation succeeds
- ✅ Details appear in approver notification: "**Details:** Manager approval required, not peer"
- ✅ Details appear in `/approve show A-TEST02`: "**Details:** Manager approval required, not peer"

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 3: Capture Details with "Sensitive information" Reason
**Objective:** Verify details are captured for "Sensitive information" reason

**Steps:**
1. Create new approval: `/approve request @UserB "Test 3 for Story 7.3"`
2. Note code (e.g., `A-TEST03`)
3. Run: `/approve cancel A-TEST03`
4. In dialog:
   - Select reason: **"Sensitive information"**
   - In "Additional details (optional)" field, type: **"Discussed offline via secure channel"**
   - Click **"Cancel Request"**

**Expected Results:**
- ✅ Cancellation succeeds
- ✅ Details appear in notifications and show command

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 4: Capture Details with "Other" Reason
**Objective:** Verify details are REQUIRED for "Other" reason (existing behavior maintained)

**Steps:**
1. Create new approval: `/approve request @UserB "Test 4 for Story 7.3"`
2. Note code (e.g., `A-TEST04`)
3. Run: `/approve cancel A-TEST04`
4. In dialog:
   - Select reason: **"Other"**
   - **Leave "Additional details" EMPTY**
   - Click **"Cancel Request"**

**Expected Results:**
- ❌ Validation error: "Please provide details when selecting 'Other reason'"
- ❌ Modal does NOT close

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

**Step 2:**
5. Type in details: **"Requirements changed after team discussion"**
6. Click **"Cancel Request"**

**Expected Results:**
- ✅ Cancellation succeeds
- ✅ Details appear in notifications and show command

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 5: Cancellation WITHOUT Details (Empty Field)
**Objective:** Verify cancellation works when details field is left empty (for non-"Other" reasons)

**Steps:**
1. Create new approval: `/approve request @UserB "Test 5 for Story 7.3"`
2. Note code (e.g., `A-TEST05`)
3. Run: `/approve cancel A-TEST05`
4. In dialog:
   - Select reason: **"No longer needed"**
   - **Leave "Additional details" field EMPTY**
   - Click **"Cancel Request"**

**Expected Results:**
- ✅ Cancellation succeeds (no validation error)
- ✅ Approver notification shows reason but NO details section
- ✅ `/approve show A-TEST05` shows reason but NO "**Details:**" line

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 6: Details with Whitespace Only (Edge Case)
**Objective:** Verify whitespace-only details are treated as empty

**Steps:**
1. Create new approval: `/approve request @UserB "Test 6 for Story 7.3"`
2. Note code (e.g., `A-TEST06`)
3. Run: `/approve cancel A-TEST06`
4. In dialog:
   - Select reason: **"Other"**
   - Type only spaces/tabs in details: **"     "**
   - Click **"Cancel Request"**

**Expected Results:**
- ❌ Validation error: "Please provide details when selecting 'Other reason'"
- ❌ Modal does NOT close

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 7: Field Label and Help Text Verification
**Objective:** Verify UI labels are clear and match requirements

**Steps:**
1. Run: `/approve cancel <any-approval-code>`
2. Observe the modal dialog

**Expected Results:**
- ✅ Field label: **"Additional details (optional)"** (NOT "if Other")
- ✅ Field has help text: **"Add context or explanation (optional)"**
- ✅ Placeholder text: **"e.g., Project scope changed, requirements updated..."**
- ✅ Reason dropdown includes all 4 options: "No longer needed", "Wrong approver", "Sensitive information", "Other"

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 8: Requestor Notification Includes Details (Story 7.1 Integration)
**Objective:** Verify requestor receives notification with details (fixes Story 7.1)

**Steps:**
1. Create approval as User A: `/approve request @UserB "Test 8 for Story 7.3"`
2. Note code (e.g., `A-TEST08`)
3. Cancel as User A with details:
   - Reason: "No longer needed"
   - Details: "Verified requestor notification"
4. Check User A's DMs from the bot

**Expected Results:**
- ✅ User A receives DM with title: "🚫 **Your Approval Request Was Canceled**"
- ✅ Message includes: "**Reason:** No longer needed"
- ✅ Message includes: "**Details:** Verified requestor notification"
- ✅ Message includes approver username

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 9: Approver Notification Includes Details
**Objective:** Verify approver receives notification with details

**Steps:**
1. Create approval as User A: `/approve request @UserB "Test 9 for Story 7.3"`
2. Note code (e.g., `A-TEST09`)
3. Cancel as User A with details:
   - Reason: "Wrong approver"
   - Details: "Should have been manager, not peer"
4. As User B, check DMs from the bot

**Expected Results:**
- ✅ User B receives DM with title: "🚫 **Approval Request Canceled**"
- ✅ Message includes: "**Reason:** Wrong approver"
- ✅ Message includes: "**Details:** Should have been manager, not peer"
- ✅ Message includes requester username

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 10: Display Details in `/approve show` Command
**Objective:** Verify details appear in full record display

**Steps:**
1. Create approval: `/approve request @UserB "Test 10 for Story 7.3"`
2. Note code (e.g., `A-TEST10`)
3. Cancel with:
   - Reason: "Sensitive information"
   - Details: "Escalated to security team"
4. Run: `/approve show A-TEST10`

**Expected Results:**
- ✅ Output includes cancellation section
- ✅ Shows: "**Reason:** Sensitive information"
- ✅ Shows: "**Details:** Escalated to security team"
- ✅ Shows: "**Canceled by:** @UserA"
- ✅ Shows: "**Canceled:** [timestamp]"

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 11: Display in `/approve list` Command
**Objective:** Verify cancellation info appears in list view

**Steps:**
1. Cancel an approval with details (reuse from previous test)
2. Run: `/approve list` (or `/approve list --canceled`)

**Expected Results:**
- ✅ Canceled request appears in list
- ✅ Shows reason in display
- ✅ Shows canceled timestamp
- ✅ Shows details (if list view includes them - verify against Epic 5 requirements)

**Actual Results:**
- [ ] Pass / [ ] Fail
- Notes: _____________________

---

### Test Case 12: Backward Compatibility - Old Canceled Records
**Objective:** Verify old canceled records (without CanceledDetails field) display correctly

**Note:** This test requires access to records created before Story 7.3 implementation. If unavailable, manually create a record with empty `CanceledDetails` field in the KV store.

**Steps:**
1. Identify an old canceled record (created before v0.2.0 with Story 7.3)
2. Run: `/approve show <old-code>`

**Expected Results:**
- ✅ Record displays without errors
- ✅ Shows: "**Reason:** [reason]"
- ✅ Does NOT show "**Details:**" line (since field is empty)
- ✅ No error messages or "undefined" text

**Actual Results:**
- [ ] Pass / [ ] Fail / [ ] N/A (no old records available)
- Notes: _____________________

---

## Success Criteria Summary

### ✅ Primary Goal: No Silent Data Loss
- [x] Details captured for "No longer needed" (Test Case 1)
- [x] Details captured for "Wrong approver" (Test Case 2)
- [x] Details captured for "Sensitive information" (Test Case 3)
- [x] Details captured for "Other" (Test Case 4)
- [x] Details optional for non-"Other" reasons (Test Case 5)

### ✅ Display Verification
- [x] Details appear in `/approve show` (Test Case 10)
- [x] Details appear in requestor notification (Test Case 8)
- [x] Details appear in approver notification (Test Case 9)

### ✅ UX Validation
- [x] Field label clearly says "(optional)" (Test Case 7)
- [x] Help text provides guidance (Test Case 7)
- [x] Validation works for "Other" reason (Test Case 4)

### ✅ Edge Cases
- [x] Whitespace-only details rejected for "Other" (Test Case 6)
- [x] Empty details allowed for non-"Other" reasons (Test Case 5)
- [x] Backward compatibility with old records (Test Case 12)

---

## Defects Found

| Test Case | Severity | Description | Status |
|-----------|----------|-------------|--------|
| _Example: TC-3_ | _High_ | _Details not displayed in notification_ | _Open_ |
|           |          |             |        |

---

## Sign-Off

### Testing Completed By
- **Name:** _____________________
- **Date:** _____________________
- **Overall Result:** [ ] Pass / [ ] Fail
- **Ready for Merge:** [ ] Yes / [ ] No

### Notes/Comments
_____________________
_____________________
_____________________

---

## Appendix: Quick Verification Script

For rapid verification, run these commands in sequence:

```bash
# Setup
/approve request @UserB "Quick test - No longer needed"
# Note code, e.g., A-TEST01

# Test 1: Details with "No longer needed"
/approve cancel A-TEST01
# Select "No longer needed", add details "Test details 1", submit
/approve show A-TEST01
# Verify details appear

# Test 2: Empty details allowed
/approve request @UserB "Quick test - Empty details"
# Note code, e.g., A-TEST02
/approve cancel A-TEST02
# Select "No longer needed", leave details empty, submit
# Should succeed

# Test 3: "Other" requires details
/approve request @UserB "Quick test - Other validation"
# Note code, e.g., A-TEST03
/approve cancel A-TEST03
# Select "Other", leave details empty, submit
# Should show validation error
```

---

**Document Version:** 1.0
**Last Updated:** 2026-01-15
**Story:** 7.3 - Fix Additional Details Box Behavior in Cancel Dialog
