# Story 10.2: Update API Handlers for Matterpoll Pattern

Status: done

## Story

As a plugin server,
I want API handlers to work with the Matterpoll button pattern,
so that Approve/Deny button clicks are processed correctly.

## Acceptance Criteria

### AC1: New API Routes for Matterpoll Pattern
- Add `POST /api/v1/approval/{code}/approve` route
- Add `POST /api/v1/approval/{code}/deny` route
- Extract approval `code` from URL path using `mux.Vars()`
- Receive `PostActionIntegrationRequest` from Mattermost
- Extract user ID from `request.UserId` for permission validation

### AC2: PostActionIntegrationResponse
- Return proper `PostActionIntegrationResponse` struct
- Include `EphemeralText` for user feedback on errors
- Success response opens confirmation modal (existing behavior)

### AC3: Handler Logic (Approve/Deny)
- Look up approval record by `code` (not ID)
- Validate user is the designated approver
- Validate approval status is `pending`
- Open confirmation modal for user to confirm/add comment
- Return appropriate response

### AC4: Handle Edge Cases
- Already decided: Return ephemeral message "This approval has already been decided"
- Invalid approval code: Return "Approval not found"
- Permission denied: Return "You are not the designated approver"
- Empty code: Return "Invalid approval code"

## Tasks / Subtasks

- [x] Task 1: Add new API routes to router (AC: 1)
  - [x] 1.1: In `server/api.go` `ServeHTTP()`, add route for `/api/v1/approval/{code}/approve`
  - [x] 1.2: Add route for `/api/v1/approval/{code}/deny`
  - [x] 1.3: Both routes map to same handler with action parameter OR two separate handlers
  - [x] 1.4: Ensure routes require authentication (`apiRouter` with `checkAuthentication`)

- [x] Task 2: Implement `handleApprovalAction()` handler (AC: 1, 2, 3, 4)
  - [x] 2.1: Create `handleApprovalAction(w, r, action string)` method on Plugin
  - [x] 2.2: Extract `code` from URL path using `mux.Vars(r)["code"]`
  - [x] 2.3: Validate code is not empty (AC4)
  - [x] 2.4: Decode `PostActionIntegrationRequest` from request body
  - [x] 2.5: Extract user ID from `request.UserId`

- [x] Task 3: Implement approval lookup and validation (AC: 3, 4)
  - [x] 3.1: Look up approval by code: `p.store.GetByCode(code)`
  - [x] 3.2: Handle not found (AC4)
  - [x] 3.3: Validate user is approver: `request.UserId == record.ApproverID`
  - [x] 3.4: Handle permission denied (AC4)
  - [x] 3.5: Validate status is pending: `record.Status == approval.StatusPending`
  - [x] 3.6: Handle already decided (AC4)

- [x] Task 4: Open confirmation modal (AC: 3)
  - [x] 4.1: Reuse existing `openConfirmationModal()` function
  - [x] 4.2: Pass approval record, action type, trigger ID from request
  - [x] 4.3: Return success response (empty `PostActionIntegrationResponse`)

- [x] Task 5: Add unit tests (AC: all)
  - [x] 5.1: Test route registration (verify routes exist)
  - [x] 5.2: Test valid approve request opens modal
  - [x] 5.3: Test valid deny request opens modal
  - [x] 5.4: Test invalid code returns error
  - [x] 5.5: Test non-approver returns permission denied
  - [x] 5.6: Test already decided returns appropriate error
  - [x] 5.7: Test missing PostActionIntegrationRequest returns error

## Dev Notes

### Critical Context: Story 10.1 Changed URL Pattern

Story 10.1 created `CreateInteractiveApprovalPost()` which generates Integration URLs in the format:
```
/plugins/com.mattermost.plugin-approver2/api/v1/approval/{code}/approve
/plugins/com.mattermost.plugin-approver2/api/v1/approval/{code}/deny
```

The **current** `/action` handler uses a Context map with `approval_id` and `action`. The **new** handlers must extract the approval `code` from the URL path.

### Current vs New Pattern

**Current (v2.2.0) - `/action` endpoint:**
```go
// URL: /plugins/com.mattermost.plugin-approver2/action
// Context: { "approval_id": "uuid", "action": "approve" }

contextData := request.Context
approvalID := contextData["approval_id"].(string)
action := contextData["action"].(string)
record := p.store.GetApproval(approvalID)  // Lookup by ID
```

**New (Story 10.2) - `/api/v1/approval/{code}/approve|deny`:**
```go
// URL: /plugins/com.mattermost.plugin-approver2/api/v1/approval/A-XYZ123/approve
// Code in URL path, not context

vars := mux.Vars(r)
code := vars["code"]                        // e.g., "A-XYZ123"
record := p.store.GetByCode(code)           // Lookup by CODE
```

### GetByCode() Implementation

The store already has `GetByCode(code string)` method - verify it exists:
```go
// server/store/approval_store.go
func (s *ApprovalStore) GetByCode(code string) (*approval.ApprovalRecord, error)
```

### Modal Flow Preserved

The confirmation modal flow remains unchanged:
1. User clicks Approve/Deny button
2. Handler opens confirmation modal
3. User confirms (optionally adds comment)
4. Modal submission goes to `/dialog/submit`
5. `handleConfirmDecision()` records the decision
6. Post is updated to remove buttons

### Router Registration Pattern

Follow existing pattern in `api.go`:
```go
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
    router := mux.NewRouter()

    // Existing routes
    router.HandleFunc("/action", p.handleAction).Methods(http.MethodPost)

    // API routes (authenticated)
    apiRouter := router.PathPrefix("/api/v1").Subrouter()
    apiRouter.Use(p.checkAuthentication)

    // NEW: Matterpoll pattern routes
    apiRouter.HandleFunc("/approval/{code}/approve", p.handleApprovalApprove).Methods(http.MethodPost)
    apiRouter.HandleFunc("/approval/{code}/deny", p.handleApprovalDeny).Methods(http.MethodPost)
```

### Handler Signature Options

**Option A: Two separate handlers**
```go
func (p *Plugin) handleApprovalApprove(w http.ResponseWriter, r *http.Request) {
    p.handleApprovalAction(w, r, "approve")
}

func (p *Plugin) handleApprovalDeny(w http.ResponseWriter, r *http.Request) {
    p.handleApprovalAction(w, r, "deny")
}
```

**Option B: Single handler with path inspection**
```go
// Not recommended - harder to read
```

Use Option A for clarity.

### Response Format

Use existing helper functions:
```go
// Success - opens modal, no update needed yet
p.writeActionSuccess(w)

// Error - returns ephemeral text
p.writeActionError(w, "This approval has already been decided")
```

### Existing Code to Reuse

- `openConfirmationModal()` - Opens the decision confirmation modal
- `writeActionSuccess()` - Returns empty `PostActionIntegrationResponse`
- `writeActionError()` - Returns `PostActionIntegrationResponse` with `EphemeralText`
- Validation logic from `handleAction()` - Permission checks, status checks

### Testing Strategy

Mock the store to return test records:
```go
mockStore.EXPECT().GetByCode("A-TEST01").Return(&approval.ApprovalRecord{
    Code:       "A-TEST01",
    Status:     approval.StatusPending,
    ApproverID: "user123",
}, nil)
```

### References

- [Source: server/api.go#lines 21-37 - Current route registration]
- [Source: server/api.go#lines 395-463 - Current /action handler]
- [Source: server/notifications/interactive_post.go#lines 31-55 - New URL format]
- [Source: epic-10-dm-interactive-buttons.md#Story 10.2]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- Build: `go build ./server/...` - PASS
- Tests: `go test ./server/... -run "TestHandleApprovalAction_MatterpollPattern|TestMatterpollRouteRegistration"` - PASS (10 new tests)
- Full suite: `go test ./server/...` - PASS

### Completion Notes List

1. Added two new routes to `apiRouter` in `ServeHTTP()`:
   - `POST /api/v1/approval/{code}/approve` → `handleApprovalApprove`
   - `POST /api/v1/approval/{code}/deny` → `handleApprovalDeny`
2. Routes use `MattermostAuthorizationRequired` middleware (require auth header)
3. Created `handleApprovalApprove()` and `handleApprovalDeny()` wrapper handlers
4. Created `handleApprovalAction()` shared handler:
   - Extracts code from URL path using `mux.Vars(r)`
   - Validates code is not empty
   - Decodes `PostActionIntegrationRequest` from body
   - Validates user ID is present
   - Looks up approval by code using `p.store.GetByCode(code)`
   - Validates user is designated approver
   - Validates status is pending
   - Opens confirmation modal using existing `openConfirmationModal()`
   - Returns appropriate error messages per AC4
5. Added 13 comprehensive unit tests:
   - 10 tests for `TestHandleApprovalAction_MatterpollPattern` (approve/deny happy paths, permission denied, approved/denied/canceled statuses, not found, invalid JSON, missing user ID, modal failure)
   - 3 tests for `TestMatterpollRouteRegistration` (approve route, deny route, auth required)

### Code Review Fixes Applied

- **H1 FIXED:** Added tests for `denied` and `canceled` statuses to ensure all terminal states are handled
- **M1 FIXED:** Changed error message from "You are not the designated approver" to "Permission denied" to avoid information disclosure
- **M2 FIXED:** Added test for `OpenInteractiveDialog` failure case
- **M3 FIXED:** Added `defer r.Body.Close()` for proper resource cleanup, consistent with other handlers

### File List

- `server/api.go` (MODIFIED - added routes at lines 36-40 and handlers at lines 467-563)
- `server/plugin_test.go` (MODIFIED - added 13 tests for Matterpoll pattern handlers)
