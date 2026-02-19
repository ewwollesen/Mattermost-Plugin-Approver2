# Story 11.4: API Endpoint for React Modal Submission

Status: done

## Story

As a server,
I want an API endpoint that handles React modal submissions,
so that approvals are created the same way as native dialog.

## Acceptance Criteria

### AC1: New API Endpoint
- `POST /plugins/com.mattermost.plugin-approver2/api/v1/approval/new`
- Accepts JSON body: `{channel_id, team_id, approver_id, description}`
- Returns approval record on success or errors on failure
- Endpoint protected by `MattermostAuthorizationRequired` middleware
- Request body size validation (prevent DoS)

### AC2: Validation Logic
- Reuse existing validation patterns from `handleApproveNew()`
- Self-approval check (GitHub Issue #4) - compare authenticated user with approver_id
- Description length validation (max 1000 chars via `approval.ValidateDescription()`)
- Approver exists and active check via `approval.ValidateApprover()`
- Empty field validation for approver_id and description

### AC3: Response Format
```json
// Success (201 Created)
{
  "success": true,
  "approval": {
    "code": "ABC-123",
    "id": "uuid",
    "status": "pending"
  }
}

// Validation Error (400 Bad Request)
{
  "success": false,
  "errors": {
    "approver_id": "Error message",
    "description": "Error message"
  }
}

// Server Error (500 Internal Server Error)
{
  "success": false,
  "error": "Generic error message"
}
```

### AC4: Authorization
- Verify user is authenticated via `Mattermost-User-ID` header (existing middleware)
- Extract requester ID from authenticated user (NOT from request body)
- Channel access validation NOT required (approval creation works cross-channel)

### AC5: Backward Compatibility
- Existing `/dialog/submit` endpoint unchanged
- Both endpoints can create approvals
- Same approval record structure used

## Tasks / Subtasks

- [x] Task 1: Create API request/response structs (AC: 1, 3)
  - [x] 1.1: Define `ApprovalNewRequest` struct with `ChannelID`, `TeamID`, `ApproverID`, `Description` fields
  - [x] 1.2: Define `ApprovalNewResponse` struct with `Success`, `Approval`, `Errors`, `Error` fields
  - [x] 1.3: Add JSON struct tags matching snake_case convention (consistent with Epic 10 API patterns)
  - [x] 1.4: Add validation struct tags if using go-validator (optional) - N/A, using manual validation

- [x] Task 2: Register new endpoint in `ServeHTTP` (AC: 1, 4)
  - [x] 2.1: Add route `apiRouter.HandleFunc("/approval/new", p.handleApprovalNew).Methods(http.MethodPost)`
  - [x] 2.2: Route under `/api/v1` prefix for auth middleware protection
  - [x] 2.3: Verify route doesn't conflict with existing `/approval/{code}/approve` and `/approval/{code}/deny` routes

- [x] Task 3: Implement `handleApprovalNew()` handler (AC: 1, 2, 3)
  - [x] 3.1: Parse JSON request body with size limit (16KB max)
  - [x] 3.2: Extract authenticated user ID from `r.Header.Get("Mattermost-User-ID")`
  - [x] 3.3: Perform validation (see Task 4)
  - [x] 3.4: Create approval record using existing flow from `handleApproveNew()`
  - [x] 3.5: Return JSON response with appropriate status code (201, 400, or 500)

- [x] Task 4: Implement validation (AC: 2)
  - [x] 4.1: Empty field validation for `approver_id` (required, non-empty)
  - [x] 4.2: Empty field validation for `description` (required, non-empty after trim)
  - [x] 4.3: Self-approval check: `approver_id != authenticated_user_id` (GitHub Issue #4)
  - [x] 4.4: Description length validation via `approval.ValidateDescription()`
  - [x] 4.5: Approver validation via `approval.ValidateApprover()` (exists, active)
  - [x] 4.6: Return field-specific errors in response (same format as client validation)

- [x] Task 5: Reuse existing approval creation logic (AC: 5)
  - [x] 5.1: Extract approval creation logic from `handleApproveNew()` into reusable function if needed - N/A, duplicated logic
  - [x] 5.2: Or duplicate relevant logic (if extraction is too invasive) - Implemented
  - [x] 5.3: Include DM notification sending (best effort, graceful degradation)
  - [x] 5.4: Include playbook detection and posting (if playbook-linked channel)
  - [x] 5.5: Include ephemeral confirmation message (skip in playbook channels)

- [x] Task 6: Create comprehensive unit tests (AC: 1, 2, 3, 4, 5)
  - [x] 6.1: Test successful approval creation returns 201 with approval data
  - [x] 6.2: Test missing `approver_id` returns 400 with field error
  - [x] 6.3: Test missing `description` returns 400 with field error
  - [x] 6.4: Test empty/whitespace `description` returns 400 with field error
  - [x] 6.5: Test self-approval returns 400 with field error
  - [x] 6.6: Test description > 1000 chars returns 400 with field error
  - [x] 6.7: Test invalid/deleted approver returns 400 with field error
  - [x] 6.8: Test unauthorized request (no `Mattermost-User-ID`) returns 401
  - [x] 6.9: Test KV store error returns 500
  - [x] 6.10: Test existing `/dialog/submit` endpoint still works

- [x] Task 7: Update webapp `ApprovalRequestModal` to call new endpoint (AC: 1)
  - [x] 7.1: Update `handleSubmit()` in `ApprovalRequestModal.tsx` to call real API
  - [x] 7.2: Replace simulated success with actual fetch call
  - [x] 7.3: Handle server-side validation errors (display in form)
  - [x] 7.4: Handle success response (close modal)
  - [x] 7.5: Handle network/server errors (show generic error)

- [x] Task 8: Integration validation (AC: 5)
  - [x] 8.1: Test full flow: modal → API → approval created → DM sent - verified via unit tests
  - [x] 8.2: Verify existing dialog submission still works - verified via TestHandleApprovalNew/existing_dialog/submit
  - [x] 8.3: Run full test suite (server + webapp) - 675 server tests, 234 webapp tests pass
  - [x] 8.4: Build successful - dist/com.mattermost.plugin-approver2-2.3.1+aba2cb4.tar.gz

## Dev Notes

### Story 11.3 Implementation Available

**From Story 11.3 (Approval Request Modal):**
- `ApprovalRequestModal.tsx` - Form component with validation
- ~~Currently simulates success with `setTimeout(() => onClose(), 500)`~~ → Now calls real API
- ~~Has TODO comment: `// TODO: Story 11.4 - API call`~~ → Implemented in this story
- Client-side validation already in place (matches server-side)
- `handleSubmit()` now calls `/api/v1/approval/new` endpoint

### Existing API Patterns

**From `server/api.go`:**
```go
// Matterpoll pattern endpoints (Story 10.2)
apiRouter.HandleFunc("/approval/{code}/approve", p.handleApprovalApprove).Methods(http.MethodPost)
apiRouter.HandleFunc("/approval/{code}/deny", p.handleApprovalDeny).Methods(http.MethodPost)

// Auth middleware already exists
apiRouter.Use(p.MattermostAuthorizationRequired)

// Get user ID from header
userID := r.Header.Get("Mattermost-User-ID")
```

**Response patterns from existing handlers:**
```go
// Success response (JSON)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(response)

// Error response
http.Error(w, "error message", http.StatusBadRequest)
```

### Existing Validation Logic (Reuse)

**From `handleApproveNew()` in `server/api.go`:**
```go
// Layer 1: Field presence validation
response := command.HandleDialogSubmission(payload.Submission, payload.UserId)
if len(response.Errors) > 0 {
    return response
}

// Layer 2: Business logic validation
// Self-approval check (GitHub Issue #4)
if approverID == payload.UserId {
    return &model.SubmitDialogResponse{
        Errors: map[string]string{
            "approver": "You cannot approve your own request.",
        },
    }
}

// Description validation
if err := approval.ValidateDescription(description); err != nil {
    return &model.SubmitDialogResponse{
        Errors: map[string]string{"description": err.Error()},
    }
}

// Approver validation
approver, err := approval.ValidateApprover(approverID, p.API)
if err != nil {
    return &model.SubmitDialogResponse{
        Errors: map[string]string{"approver": err.Error()},
    }
}
```

### Request/Response Struct Design

```go
// server/api.go - Request struct
type ApprovalNewRequest struct {
    ChannelID   string `json:"channel_id"`
    TeamID      string `json:"team_id"`
    ApproverID  string `json:"approver_id"`
    Description string `json:"description"`
}

// Response struct
type ApprovalNewResponse struct {
    Success  bool              `json:"success"`
    Approval *ApprovalData     `json:"approval,omitempty"`
    Errors   map[string]string `json:"errors,omitempty"`
    Error    string            `json:"error,omitempty"`
}

type ApprovalData struct {
    ID     string `json:"id"`
    Code   string `json:"code"`
    Status string `json:"status"`
}
```

### Handler Implementation Pattern

```go
// handleApprovalNew creates a new approval from React modal submission
// POST /plugins/com.mattermost.plugin-approver2/api/v1/approval/new
func (p *Plugin) handleApprovalNew(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Get authenticated user
    requesterID := r.Header.Get("Mattermost-User-ID")

    // Parse request body with size limit
    r.Body = http.MaxBytesReader(w, r.Body, 16*1024) // 16KB limit
    var req ApprovalNewRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Handle parse error
    }

    // Validation
    errors := make(map[string]string)

    if req.ApproverID == "" {
        errors["approver_id"] = "Approver field is required."
    }
    if strings.TrimSpace(req.Description) == "" {
        errors["description"] = "Description field is required."
    }
    if req.ApproverID == requesterID {
        errors["approver_id"] = "You cannot approve your own request."
    }

    if len(errors) > 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ApprovalNewResponse{
            Success: false,
            Errors:  errors,
        })
        return
    }

    // Continue with validation and approval creation...
    // (Reuse logic from handleApproveNew)
}
```

### Webapp Integration

**Update `ApprovalRequestModal.tsx`:**
```typescript
const handleSubmit = async () => {
    if (!validate()) return;

    setForm(f => ({...f, submitting: true}));

    try {
        const response = await fetch(`/plugins/com.mattermost.plugin-approver2/api/v1/approval/new`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                channel_id: channelId,
                team_id: teamId,
                approver_id: form.approverId,
                description: form.description,
            }),
        });

        const data = await response.json();

        if (data.success) {
            onClose();
        } else if (data.errors) {
            // Field-specific errors
            setForm(f => ({
                ...f,
                submitting: false,
                errors: {
                    approver: data.errors.approver_id,
                    description: data.errors.description,
                },
            }));
        } else {
            // Generic server error
            setForm(f => ({
                ...f,
                submitting: false,
                errors: {
                    description: data.error || 'An error occurred. Please try again.',
                },
            }));
        }
    } catch (err) {
        // Network error
        setForm(f => ({
            ...f,
            submitting: false,
            errors: {
                description: 'Network error. Please try again.',
            },
        }));
    }
};
```

### Testing Strategy

**Server Tests (`server/api_test.go`):**
1. Test handler isolation with mock Plugin API
2. Test validation errors individually
3. Test success path creates approval record
4. Test existing dialog endpoint unchanged

**Webapp Tests (`ApprovalRequestModal.test.tsx`):**
1. Mock fetch for API calls
2. Test successful submission closes modal
3. Test server validation errors display in form
4. Test network error handling

### Dependencies on Previous Stories

- **Story 11.1**: Modal infrastructure (not directly used by API)
- **Story 11.2**: UserSelector (not directly used by API)
- **Story 11.3**: ApprovalRequestModal (will call this API)

### What's Deferred to Story 11.5

- Slash command modification (`/approve new` trigger)
- Ephemeral post for modal trigger
- Integration with ModalTriggerPost

### Error Message Consistency

Use same messages as client-side validation in Story 11.3:
```typescript
// Client messages (ApprovalRequestModal.tsx)
'Please select an approver'     → Server: 'Approver field is required.'
'Please describe what needs...' → Server: 'Description field is required.'
```

Server messages should be consistent but can be slightly more formal (API context).

### References

- [Source: server/api.go#handleApproveNew - Existing dialog handler]
- [Source: server/api.go#handleApprovalApprove - Matterpoll pattern handler]
- [Source: server/command/dialog.go#HandleDialogSubmission - Validation logic]
- [Source: webapp/src/components/ApprovalRequestModal.tsx - Modal component]
- [Source: _bmad-output/implementation-artifacts/epic-11-react-modal-framework.md#story-114 - Epic requirements]
- [GitHub Issue #4: Self-approval prevention]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- **Task 1-4**: Implemented `handleApprovalNew()` handler with:
  - `ApprovalNewRequest`, `ApprovalNewResponse`, `ApprovalData` structs
  - 16KB request body size limit (DoS prevention)
  - Two-layer validation (field presence + business logic)
  - Self-approval prevention (GitHub Issue #4)
  - Description length validation (1000 char max)
  - Approver existence/active check

- **Task 5**: Duplicated approval creation logic from `handleApproveNew()` (extraction too invasive):
  - DM notification sending (graceful degradation)
  - Playbook detection and posting
  - Ephemeral confirmation (skipped in playbook channels)

- **Task 6**: Added 10 server tests in `TestHandleApprovalNew`:
  - Success (201), validation errors (400), auth (401), KV error (500)
  - Backward compatibility test for `/dialog/submit`

- **Task 7**: Updated `ApprovalRequestModal.tsx`:
  - Real fetch call to `/api/v1/approval/new`
  - Server validation error display
  - Network error handling
  - isMountedRef check after async operations

- **Task 8**: Validation complete:
  - 675 server tests pass
  - 234 webapp tests pass (including 4 new API integration tests)
  - Build successful

### File List

**Server (Modified):**
- server/api.go - Added structs, route registration, handleApprovalNew handler
- server/api_test.go - Added TestHandleApprovalNew with 11 test cases + helper functions

**Webapp (Modified):**
- webapp/src/components/ApprovalRequestModal.tsx - Real API integration
- webapp/src/components/ApprovalRequestModal.test.tsx - Added setupMockFetch, 4 API tests
- webapp/src/components/index.ts - Export ApprovalRequestModal component
- webapp/src/index.tsx - Register modal components

**Sprint Tracking (Modified):**
- _bmad-output/implementation-artifacts/sprint-status.yaml - Story status updates

