# Story 6.2: Verification Step for Approved Requests

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a requester,
I want to verify that I completed the approved action,
so that I can close the approval loop and notify the approver that the action is done.

## Context

This is the second feature in Epic 6 "Request Timeout & Verification (Feature Complete for 1.0)". Epic 6 completes the single-approver workflow by:
1. **Auto-timeout (Story 6.1):** Handles abandoned requests by auto-canceling after 30 minutes
2. **Verification (this story):** Closes the approval loop when requesters confirm action completion

**Business Value:**
- Closes the approval loop with confirmation that action was completed
- Provides visibility to approvers that their approval led to action
- Creates complete audit trail from request → approval → execution → verification
- Enables "closed-loop accountability" for time-sensitive decisions

**Workflow:**
1. Requester creates approval request
2. Approver approves request
3. Requester executes the approved action
4. **NEW:** Requester verifies completion with `/approve verify`
5. Approver receives confirmation notification

## Acceptance Criteria

**AC1: Verify Command Basic Flow**
- **Given** a requester has an approved request
- **When** the requester types `/approve verify <CODE> [optional comment]`
- **Then** the system retrieves the approval record by code
- **And** verifies the authenticated user is the original requester (permission check)

**AC2: Update Record with Verification Metadata**
- **Given** the approval record has status "approved"
- **When** the verify command is executed by the requester
- **Then** the system updates the ApprovalRecord with new fields:
  - Verified: true (boolean)
  - VerifiedAt: current timestamp (epoch millis)
  - VerificationComment: {comment text or empty string}
- **And** the Status field remains "approved" (verification is separate metadata)
- **And** the update is persisted atomically to the KV store
- **And** the operation completes within 2 seconds

**AC3: Requester Confirmation**
- **Given** the verification is successful
- **When** the system confirms the verification
- **Then** an ephemeral message is posted to the requester:
  - "✅ Verification recorded for approval request {Code}."
  - "The approver will be notified that the action is complete."
- **And** the message appears within 1 second

**AC4: Approver Notification**
- **Given** a verification is recorded
- **When** the system processes the verification
- **Then** a DM notification is sent to the approver within 5 seconds
- **And** the DM includes:
  - Header: "✅ **Action Verified Complete**"
  - Request ID: "{Code}"
  - Requester info: "**Requester:** @{requesterUsername} ({requesterDisplayName})"
  - Verification time: "**Verified:** {timestamp in YYYY-MM-DD HH:MM:SS UTC format}"
  - Original request: "**Original Request:**\n> {description}" (quoted)
  - Verification comment (if provided): "**Verification Comment:**\n{verificationComment}"
  - Action statement: "**Status:** The requester has confirmed this action is complete."

**AC5: Optional Comment Handling**
- **Given** the verification command includes an optional comment
- **When** the comment is provided
- **Then** the comment is stored in VerificationComment field (max 500 characters)
- **And** the comment is included in the approver notification
- **And** the comment is validated before submission

**AC6: No Comment Provided**
- **Given** the verification command does not include a comment
- **When** the verify command is executed
- **Then** VerificationComment is stored as empty string
- **And** the "Verification Comment" section is omitted from the notification

**AC7: Help Text for Missing Arguments**
- **Given** a requester types `/approve verify` without providing a code
- **When** the command is processed
- **Then** the system returns help text:
  - "Usage: /approve verify <APPROVAL_ID> [optional comment]"
  - "Example: /approve verify A-X7K9Q2"
  - "Example: /approve verify A-X7K9Q2 Rollback completed successfully"

**AC8: Invalid Code Error**
- **Given** a requester types `/approve verify INVALID-CODE`
- **When** the command is processed
- **Then** the system returns an error: "❌ Approval request 'INVALID-CODE' not found. Use `/approve list` to see your requests."

**AC9: Permission Check (Non-Requester)**
- **Given** the authenticated user is NOT the original requester
- **When** they attempt to verify a request
- **Then** the system returns an error: "❌ Permission denied. Only the requester can verify completion of an approval request."
- **And** the approval record is not modified

**AC10: Status Check (Non-Approved Requests)**
- **Given** the approval record status is "pending", "denied", or "canceled"
- **When** the verify command is executed
- **Then** the system returns an error: "❌ Cannot verify approval request {Code}. Only approved requests can be verified. (Current status: {Status})"
- **And** the approval record is not modified

**AC11: One-Time Verification (Already Verified)**
- **Given** the approval record is already verified (Verified == true)
- **When** the verify command is executed
- **Then** the system returns an error: "❌ Approval request {Code} has already been verified on {VerifiedAt timestamp}."
- **And** the existing verification is not modified (immutability - one-time verification only)

**AC12: Display in Detail View**
- **Given** a user executes `/approve get <CODE>` for a verified request
- **When** the detailed view is displayed
- **Then** the output includes a new "Verification" section:
```
**Verification:**
- ✅ Verified on: 2026-01-10 15:30:00 UTC
- Comment: Rollback completed successfully
```
- **And** the section appears after the "Decision Comment" section
- **And** if not verified, the "Verification" section is omitted

**AC13: List View (No Verification Display)**
- **Given** a user executes `/approve list`
- **When** the results include verified requests
- **Then** the status shows: "✅ Approved" (same as unverified approved requests)
- **And** verification status is NOT shown in the list view (only in detail view via `/approve get`)

**AC14: Notification Failure Handling**
- **Given** the verification notification fails to send
- **When** the DM delivery fails
- **Then** the verification record remains valid (data integrity prioritized)
- **And** the error is logged for debugging
- **And** the system does not retry automatically
- **And** the requester's confirmation still shows success

**AC15: Comment Length Validation**
- **Given** the verification comment exceeds 500 characters
- **When** the system validates the input
- **Then** an error message displays: "❌ Verification comment is too long (max 500 characters). Please shorten your comment."
- **And** the verification is not recorded

## Tasks / Subtasks

- [x] Task 1: Add verification fields to ApprovalRecord data model (AC: 2)
  - [x] Subtask 1.1: Add Verified bool field to ApprovalRecord struct in `server/approval/models.go`
  - [x] Subtask 1.2: Add VerifiedAt int64 field (epoch millis, default 0)
  - [x] Subtask 1.3: Add VerificationComment string field (default empty string)
  - [x] Subtask 1.4: Add JSON tags for all three fields (verified, verifiedAt, verificationComment)
  - [x] Subtask 1.5: Update struct documentation to describe verification fields
  - [x] Subtask 1.6: Schema version remains 1 (backward compatible additive fields)

- [x] Task 2: Implement verify command parsing in router (AC: 1, 7)
  - [x] Subtask 2.1: Add "verify" case to command router in `server/command/router.go`
  - [x] Subtask 2.2: Extract approval code from command arguments
  - [x] Subtask 2.3: Extract optional comment from remaining arguments (join with spaces)
  - [x] Subtask 2.4: Return help text if no code provided (AC7)
  - [x] Subtask 2.5: Route to new executeVerify function
  - [x] Subtask 2.6: Write unit tests for command parsing

- [x] Task 3: Implement executeVerify command handler (AC: 1, 8, 9, 10, 11, 15)
  - [x] Subtask 3.1: Create executeVerify function in `server/command/router.go`
  - [x] Subtask 3.2: Retrieve approval record by code (return error if not found - AC8)
  - [x] Subtask 3.3: Verify authenticated user is requester (AC9)
  - [x] Subtask 3.4: Verify status is "approved" (AC10)
  - [x] Subtask 3.5: Verify not already verified (AC11)
  - [x] Subtask 3.6: Validate comment length ≤ 500 characters (AC15)
  - [x] Subtask 3.7: Call ApprovalService.VerifyRequest() method
  - [x] Subtask 3.8: Return ephemeral confirmation to requester (AC3)
  - [x] Subtask 3.9: Write unit tests for all validation cases

- [x] Task 4: Implement ApprovalService.VerifyRequest() method (AC: 2, 5, 6)
  - [x] Subtask 4.1: Create VerifyRequest(id, comment string) error method in `server/approval/service.go`
  - [x] Subtask 4.2: Load approval record from store
  - [x] Subtask 4.3: Set Verified = true
  - [x] Subtask 4.4: Set VerifiedAt = current timestamp (epoch millis)
  - [x] Subtask 4.5: Set VerificationComment = comment (or empty string if no comment - AC6)
  - [x] Subtask 4.6: Save updated record atomically to KV store (Status remains "approved" - AC2)
  - [x] Subtask 4.7: Write unit tests for verification logic

- [x] Task 5: Implement approver verification notification (AC: 4, 14)
  - [x] Subtask 5.1: Create formatVerificationNotification function in `server/notifications/dm.go`
  - [x] Subtask 5.2: Format DM with all required fields (header, code, requester, timestamp, description, optional comment, action statement)
  - [x] Subtask 5.3: Handle optional comment display (omit section if empty - AC6)
  - [x] Subtask 5.4: Send DM to approver after verification recorded
  - [x] Subtask 5.5: Handle notification failures gracefully (log error, don't rollback - AC14)
  - [x] Subtask 5.6: Write unit tests for notification formatting

- [x] Task 6: Update detail view to display verification (AC: 12)
  - [x] Subtask 6.1: Modify formatRecordDetails function in `server/command/router.go`
  - [x] Subtask 6.2: Add conditional "Verification" section after "Decision Comment"
  - [x] Subtask 6.3: Display verification timestamp in YYYY-MM-DD HH:MM:SS UTC format
  - [x] Subtask 6.4: Display verification comment if present
  - [x] Subtask 6.5: Omit entire section if Verified == false (AC12)
  - [x] Subtask 6.6: Write unit tests for verified and unverified display

- [x] Task 7: Update help documentation (AC: 7)
  - [x] Subtask 7.1: Add verify command to help text in `server/command/router.go`
  - [x] Subtask 7.2: Add verify to AutoCompleteHint in plugin.go command registration
  - [x] Subtask 7.3: Document verify command with examples
  - [x] Subtask 7.4: Update README.md with verify command usage

- [x] Task 8: Comprehensive testing (All ACs)
  - [x] Subtask 8.1: Unit tests for data model (new fields serialize correctly)
  - [x] Subtask 8.2: Unit tests for command parsing (with/without comment)
  - [x] Subtask 8.3: Unit tests for all validation cases (AC9, AC10, AC11, AC15)
  - [x] Subtask 8.4: Unit tests for verification notification formatting
  - [x] Subtask 8.5: Unit tests for detail view display
  - [x] Subtask 8.6: Integration test for full verification flow (create → approve → verify → notify)
  - [x] Subtask 8.7: Integration test for error cases (wrong user, wrong status, already verified)
  - [x] Subtask 8.8: Manual testing with real Mattermost instance

## Dev Notes

### Architecture Overview

**New Components:**
- **Verify Command Handler:** New slash command `/approve verify <code> [comment]`
- **Verification Service Method:** ApprovalService.VerifyRequest() for business logic
- **Verification Notification Formatter:** New DM message format for verification confirmations
- **Data Model Extension:** Three new fields on ApprovalRecord (Verified, VerifiedAt, VerificationComment)

**Integration Points:**
- Command router for verify command parsing
- ApprovalService for verification logic
- Notification service for approver DM
- KV store for atomic record updates
- Detail view formatter for verification display

**Similarities to Existing Features:**
- Command pattern matches `/approve cancel` (Story 1.7)
- Permission check pattern matches cancel command (requester-only)
- Status check pattern matches decision logic (only certain statuses)
- One-time operation matches decision immutability
- Notification pattern matches existing DM notifications

### Architecture Changes (Story 6.2)

**Critical Change: Two-Tier Immutability Model**

**Problem:**
Verification feature requires updating decided (approved) records, but the original immutability enforcement blocked ALL modifications to decided records. This caused a production bug where `/approve verify` commands failed with "approval record is immutable" error.

**Original Behavior:**
```go
// server/store/kvstore.go (before Story 6.2)
if existing.Status != approval.StatusPending {
    // BLOCKS ALL CHANGES to decided records
    return fmt.Errorf("cannot modify approval record %s: %w", record.ID, approval.ErrRecordImmutable)
}
```

**New Behavior - Relaxed Immutability with Validation:**
```go
// server/store/kvstore.go (Story 6.2)
if existing.Status != approval.StatusPending {
    // Allow verification updates on approved records ONLY
    if !isValidVerificationUpdate(existing, record) {
        return fmt.Errorf("cannot modify approval record %s: %w", record.ID, approval.ErrRecordImmutable)
    }
}
```

**Validation Rules (isValidVerificationUpdate):**
1. **Status Check:** Only approved records can be verified (not denied/canceled)
2. **Field-Level Immutability:** Core decision fields MUST remain unchanged:
   - ID, Code, Status, DecidedAt, DecisionComment
   - RequesterID, ApproverID, Description
   - CreatedAt, CanceledReason, CanceledAt
   - All notification tracking fields
3. **One-Time Operation:** Already verified records cannot be re-verified
4. **Required Fields:** Verified=true and VerifiedAt>0 must be set
5. **Allowed Changes:** ONLY Verified, VerifiedAt, VerificationComment fields

**Implementation:**
- **File:** `server/store/kvstore.go`
- **Function:** `isValidVerificationUpdate(existing, updated *ApprovalRecord) bool` (lines 26-77)
- **Test Coverage:** 4 comprehensive test cases in `server/store/kvstore_test.go`
  - Allows verification update on approved record
  - Rejects verification on already verified record
  - Rejects verification on denied/canceled records
  - Rejects changes to core fields during verification

**Impact:**
- Maintains data integrity while enabling verification workflow
- Prevents tampering with decision history
- Enforces one-time verification semantics
- No breaking changes to existing records

**Why This Wasn't in Original Story:**
This architecture change emerged during implementation when the immutability check blocked the verify command. The fix was discovered and implemented during development, representing an emergent architecture decision that maintains system integrity while enabling the feature.

### Critical Architecture Patterns (From Architecture.md)

**1. Data Model Extension (Backward Compatible)**
```go
// From Architecture: "Schema versioning for future migrations"
// NEW FIELDS (additive - backward compatible):
type ApprovalRecord struct {
    // ... existing fields ...

    // Verification metadata (Story 6.2)
    Verified             bool   `json:"verified"`              // Default: false
    VerifiedAt           int64  `json:"verifiedAt"`            // Default: 0
    VerificationComment  string `json:"verificationComment"`   // Default: ""

    SchemaVersion        int    `json:"schemaVersion"`         // Remains 1
}

// ✅ Backward compatible: Old records without these fields will deserialize with defaults
// ✅ No migration needed: Additive fields with safe defaults
```

**2. Command Parsing Pattern (From Story 1.7 - Cancel)**
```go
// Pattern: /approve verify <CODE> [optional comment]
// Similar to: /approve cancel <CODE> (Story 1.7)

func (r *Router) Route(args *model.CommandArgs) (*model.CommandResponse, error) {
    split := strings.Fields(args.Command)
    if len(split) < 2 {
        return r.executeHelp(args), nil
    }

    subcommand := split[1]
    switch subcommand {
    case "verify":
        return r.executeVerify(args, split), nil
    // ...
    }
}

func (r *Router) executeVerify(args *model.CommandArgs, split []string) *model.CommandResponse {
    // Validate: /approve verify <CODE> [comment]
    if len(split) < 3 {
        return helpResponse("Usage: /approve verify <APPROVAL_ID> [optional comment]")
    }

    code := split[2]
    comment := "" // Optional
    if len(split) > 3 {
        comment = strings.Join(split[3:], " ") // Join remaining as comment
    }

    // Process verification...
}
```

**3. Permission Check Pattern (From Story 1.7)**
```go
// From Story 1.7 (Cancel): Only requester can cancel
// Same pattern for Story 6.2: Only requester can verify

func (r *Router) executeVerify(args *model.CommandArgs, split []string) *model.CommandResponse {
    record, err := r.store.GetByCode(code)
    if err != nil {
        return errorResponse("❌ Approval request '%s' not found.", code)
    }

    // Permission check: authenticated user must be requester
    if record.RequesterID != args.UserId {
        return errorResponse("❌ Permission denied. Only the requester can verify completion.")
    }

    // Status check: only approved requests can be verified
    if record.Status != "approved" {
        return errorResponse("❌ Cannot verify. Only approved requests can be verified. (Current status: %s)", record.Status)
    }

    // One-time check: prevent double-verification
    if record.Verified {
        timestamp := formatTimestamp(record.VerifiedAt)
        return errorResponse("❌ Already verified on %s.", timestamp)
    }

    // Process verification...
}
```

**4. Graceful Degradation Pattern**
```go
// From Architecture: "Critical path first, notifications best-effort"
// Order:
// 1. Update verification record (MUST succeed)
// 2. Send approver notification (best effort, log if fails)
// 3. Return requester confirmation (always succeeds)

// ✅ Correct pattern:
if err := s.service.VerifyRequest(code, comment); err != nil {
    return errorResponse("Failed to record verification: %s", err.Error())
}

// Best effort notification
if err := s.notificationService.SendVerificationDM(record); err != nil {
    s.logger.LogError("verification notification failed",
        "approval_id", record.ID,
        "error", err.Error())
    // Continue - verification is still recorded
}

// Always return success to requester
return ephemeralResponse("✅ Verification recorded for approval request %s.\nThe approver will be notified.", code)
```

**5. Notification Formatting Pattern**
```go
// From Story 2.5 (Outcome Notification) and Story 4.2 (Cancellation Notification)
// Follow existing structured format

func formatVerificationNotification(record *ApprovalRecord) string {
    msg := "✅ **Action Verified Complete**\n\n"
    msg += fmt.Sprintf("**Request ID:** %s\n\n", record.Code)
    msg += fmt.Sprintf("**Requester:** @%s (%s)\n", record.RequesterUsername, record.RequesterDisplayName)
    msg += fmt.Sprintf("**Verified:** %s\n\n", formatTimestampUTC(record.VerifiedAt))
    msg += fmt.Sprintf("**Original Request:**\n> %s\n\n", record.Description)

    // Optional comment (omit section if empty)
    if record.VerificationComment != "" {
        msg += fmt.Sprintf("**Verification Comment:**\n%s\n\n", record.VerificationComment)
    }

    msg += "**Status:** The requester has confirmed this action is complete."
    return msg
}
```

**6. Atomic Update Pattern**
```go
// From Architecture: "Atomic persistence (all-or-nothing)"
// From Story 2.4: Decision recording pattern

func (s *ApprovalService) VerifyRequest(id string, comment string) error {
    record, err := s.store.Get(id)
    if err != nil {
        return fmt.Errorf("failed to get approval record: %w", err)
    }

    // Update verification fields
    record.Verified = true
    record.VerifiedAt = time.Now().Unix() * 1000 // Epoch millis
    record.VerificationComment = comment
    // Note: Status remains "approved"

    // Atomic save
    if err := s.store.Save(record); err != nil {
        return fmt.Errorf("failed to save verification: %w", err)
    }

    return nil
}
```

### File Structure (From Architecture.md)

**Modified Files:**
```
server/
  approval/
    models.go               # Add 3 fields to ApprovalRecord
    models_test.go          # Test field serialization
    service.go              # Add VerifyRequest() method
    service_test.go         # Test verification logic

  command/
    router.go               # Add verify command handler
    router_test.go          # Test command parsing and validation

  notifications/
    dm.go                   # Add formatVerificationNotification()
    dm_test.go              # Test notification formatting

  plugin.go                 # Update AutoCompleteHint for verify command
```

**No New Files:**
- All changes are extensions to existing files
- No new packages needed (unlike Story 6.1 which added `timeout/`)

### Naming Conventions (From Architecture.md)

**✅ Follow Mattermost Conventions:**
- CamelCase variables: `verificationComment`, `verifiedAt`, `isVerified`
- Proper initialisms: `DMNotification`, `IDValidation`
- snake_case logging keys: `"verification_comment"`, `"verified_at"`, `"approval_id"`
- Method names: `VerifyRequest()`, `formatVerificationNotification()`

**❌ Avoid:**
- snake_case variables: `verification_comment`, `verified_at`
- Generic receivers: `(me *ApprovalService)`, `(self *Router)`

### Previous Story Learnings

**From Story 6.1 (Auto-Timeout):**
- **New package creation:** Story 6.1 created `timeout/` package for background service
- **Lifecycle management:** Background goroutines need OnActivate/OnDeactivate handling
- **KV store queries:** Efficient index scanning for pending requests
- **Graceful degradation:** Critical path (record update) must succeed, notifications are best-effort

**Apply to Story 6.2:**
- **No new packages needed:** Verification is a simple command + service method
- **No background tasks:** Verification is user-initiated, not periodic
- **Simple KV operations:** Just Get + Update (no complex queries)
- **Same graceful degradation:** Record verification first, notify approver second

**From Story 5.2 (List Filtering):**
- **Minimal focused changes:** Change only what's needed
- **Comprehensive testing:** Unit tests for all cases + integration tests
- **Clear error messages:** User-friendly validation errors

**Apply to Story 6.2:**
- **Surgical changes:** Add 3 fields, 1 command, 1 service method, 1 notification format
- **Test all validation paths:** Wrong user, wrong status, already verified, comment too long
- **Clear help text:** Explain usage with examples

### Command Registration Update

**Update AutoCompleteHint (plugin.go):**
```go
// Current (Story 5.2):
AutoCompleteHint: "[new|list [pending|approved|denied|canceled|all]|get|cancel|status|help]"

// NEW (Story 6.2):
AutoCompleteHint: "[new|list [pending|approved|denied|canceled|all]|get|cancel|verify|status|help]"
//                                                                          ^^^^^^^
//                                                                          ADD THIS
```

### Display Pattern (Detail View)

**Conditional Section Display:**
```go
// From Story 4.5: CanceledReason is only displayed if status is "canceled"
// Same pattern for Story 6.2: Verification only displayed if Verified == true

func formatRecordDetails(record *ApprovalRecord) string {
    // ... existing fields ...

    // Decision Comment (existing)
    if record.DecisionComment != "" {
        msg += fmt.Sprintf("\n**Decision Comment:**\n%s\n", record.DecisionComment)
    }

    // Verification (NEW - only if verified)
    if record.Verified {
        msg += "\n**Verification:**\n"
        msg += fmt.Sprintf("- ✅ Verified on: %s\n", formatTimestampUTC(record.VerifiedAt))
        if record.VerificationComment != "" {
            msg += fmt.Sprintf("- Comment: %s\n", record.VerificationComment)
        }
    }

    // ... rest of display ...
}
```

### Validation Rules Summary

**All Validation Checks (in order):**
1. ✅ Code provided (not empty)
2. ✅ Code exists in system
3. ✅ Authenticated user is requester (permission)
4. ✅ Status is "approved" (state check)
5. ✅ Not already verified (one-time check)
6. ✅ Comment ≤ 500 characters (length validation)

**Error Messages:**
- Missing code: "Usage: /approve verify <APPROVAL_ID> [optional comment]"
- Invalid code: "❌ Approval request 'INVALID-CODE' not found."
- Wrong user: "❌ Permission denied. Only the requester can verify completion."
- Wrong status: "❌ Cannot verify. Only approved requests can be verified. (Current status: {Status})"
- Already verified: "❌ Already verified on {timestamp}."
- Comment too long: "❌ Verification comment is too long (max 500 characters)."

### Testing Strategy

**Unit Tests:**
- Data model serialization (3 new fields)
- Command parsing (with/without comment)
- All 6 validation checks
- Notification formatting (with/without comment)
- Detail view display (verified/unverified)

**Integration Tests:**
- Full flow: create → approve → verify → notify
- Error cases: wrong user, wrong status, already verified
- Comment handling: empty, present, too long
- Notification failure (graceful degradation)

**Manual Testing:**
- Test with real Mattermost instance
- Verify ephemeral messages display correctly
- Verify DM notifications arrive
- Verify detail view shows verification section
- Verify list view does NOT show verification

### References

- **Architecture:** `_bmad-output/planning-artifacts/architecture.md`
  - Section: "Go Code Structure" (lines 619-698) - Synchronous by default
  - Section: "Error Handling Patterns" (lines 360-378) - Error wrapping with %w
  - Section: "Logging Patterns" (lines 576-617) - snake_case keys
  - Section: "Data Format Conventions" (lines 740-765) - JSON field naming
  - Section: "Error Handling & Resilience" (lines 359-395) - Graceful degradation

- **Epics:** `_bmad-output/planning-artifacts/epics.md`
  - Section: Epic 6 Story 6.2 (lines 1242-1360) - Full acceptance criteria

- **PRD:** `_bmad-output/planning-artifacts/prd.md`
  - Section: FR37 (line 877) - Access control
  - Section: FR36 (line 874) - Mattermost authentication
  - Section: NFR-P2 (line 887) - Response time <2s
  - Section: NFR-P3 (line 890) - Notification <5s

- **Previous Story:** `_bmad-output/implementation-artifacts/6-1-auto-timeout-for-pending-approval-requests.md`
  - Section: Dev Notes - Graceful degradation pattern
  - Section: Tasks - Comprehensive testing approach

- **Story 1.7:** `_bmad-output/implementation-artifacts/1-7-cancel-pending-approval-requests.md`
  - Command parsing pattern for requester-only operations
  - Permission check implementation

- **Story 4.5:** `_bmad-output/implementation-artifacts/4-5-display-cancellation-reason-in-approval-details.md`
  - Conditional display pattern for optional fields in detail view

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5

### Completion Notes List

**Story created by:** Scrum Master Agent (Bob)
**Creation date:** 2026-01-13
**Story context analysis:** Ultimate BMad Method context engine completed
**Developer readiness:** Story is fully prepared with comprehensive context, architecture patterns, and implementation guidance

**Key Developer Guardrails:**
1. ✅ Add 3 backward-compatible fields to ApprovalRecord (Verified, VerifiedAt, VerificationComment)
2. ✅ Follow command parsing pattern from Story 1.7 (cancel command)
3. ✅ Implement 6 validation checks in order (code, exists, requester, approved, not-verified, comment-length)
4. ✅ Follow graceful degradation: update record first, notify second
5. ✅ Use conditional display in detail view (only if Verified == true)
6. ✅ DO NOT display verification in list view (detail view only)
7. ✅ Comprehensive testing: all validation paths + integration flow

**Relationship to Story 6.1:**
- Story 6.1: Background service (complex, async, periodic)
- Story 6.2: Slash command (simple, sync, user-initiated)
- Both: Complete Epic 6 "Feature Complete for 1.0"

### File List

**Files to create:**
- None (all changes are modifications to existing files)

**Files to modify:**
- `server/approval/models.go` - Add 3 fields to ApprovalRecord
- `server/approval/models_test.go` - Test field serialization
- `server/approval/service.go` - Add VerifyRequest() method
- `server/approval/service_test.go` - Test verification logic
- `server/command/router.go` - Add verify command handler + detail view display
- `server/command/router_test.go` - Test command parsing and validation
- `server/notifications/dm.go` - Add SendVerificationNotificationDM()
- `server/notifications/dm_test.go` - Test notification formatting
- `server/plugin.go` - Update AutoCompleteHint to include "verify", add handleVerifyCommand()
- `server/plugin_test.go` - Test verify command handler (TestHandleVerifyCommand with 6 test cases)
- `server/store/kvstore.go` - Add isValidVerificationUpdate() for relaxed immutability on verification updates (CRITICAL architecture change)
- `server/store/kvstore_test.go` - Test verification update immutability rules (4 test cases)
