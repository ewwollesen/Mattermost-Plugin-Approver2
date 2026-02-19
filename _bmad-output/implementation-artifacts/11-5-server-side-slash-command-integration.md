# Story 11.5: Server-Side Slash Command Integration

Status: review

## Story

As a user,
I want `/approve new` to open the React modal,
so that I can create approvals with the new UI.

## Acceptance Criteria

### AC1: Update executeNew()
- Modify `server/command/router.go` `executeNew()` function
- Replace `OpenInteractiveDialog` call with ephemeral post that triggers React modal
- Send ephemeral post with `custom_approval_modal` type
- Include necessary context: `channel_id`, `team_id`, `trigger_user` (user_id)
- Modal opens via `ModalTriggerPost` component (Story 11.1)

### AC2: Deprecate Native Dialog (Conditional)
- Remove native `OpenInteractiveDialog` code path
- **No feature flag needed** - Epic 11 is POC validation, direct replacement
- Keep native dialog code as reference comments for rollback documentation
- Note: Stories 11.1-11.4 have already validated React modal works

### AC3: No Breaking Changes
- `/approve new` continues to work (just opens React modal instead)
- All other commands unchanged (`list`, `get`, `status`, `help`, etc.)
- Playbook context detection unchanged (Story 8.1)
- Error handling for missing trigger ID no longer applies (remove or adapt)

### AC4: Props Consistency
- Pass same props as `ModalTriggerPost` expects:
  - `modal_type`: `"approval_request"` (matches `ModalProvider` case)
  - `channel_id`: `args.ChannelId`
  - `team_id`: `args.TeamId`
  - `trigger_user`: `args.UserId`

### AC5: Test Coverage
- Update existing `executeNew` tests to verify ephemeral post creation
- Test ephemeral post has correct type and props
- Verify no regression in other router functions

## Tasks / Subtasks

- [x] Task 1: Modify `executeNew()` to send ephemeral modal trigger (AC: 1, 4)
  - [x] 1.1: Remove `OpenInteractiveDialog` call
  - [x] 1.2: Create ephemeral post with `Type: "custom_approval_modal"`
  - [x] 1.3: Set post props: `modal_type`, `channel_id`, `team_id`, `trigger_user`
  - [x] 1.4: Call `r.api.SendEphemeralPost(args.UserId, post)`
  - [x] 1.5: Return empty `CommandResponse{}` (modal handles UX)

- [x] Task 2: Clean up obsolete code (AC: 2, 3)
  - [x] 2.1: Remove `TriggerId` validation (not needed for ephemeral posts)
  - [x] 2.2: Remove `siteURL` and `callbackURL` logic (API endpoint handles submission)
  - [x] 2.3: Keep playbook context detection (still useful for logging)
  - [x] 2.4: Add comment documenting native dialog removal for rollback reference

- [x] Task 3: Update tests (AC: 5)
  - [x] 3.1: Update `TestExecuteNew` to verify ephemeral post is sent
  - [x] 3.2: Verify ephemeral post props match expected values
  - [x] 3.3: Remove tests for trigger ID validation (no longer applicable)
  - [x] 3.4: Remove tests for OpenInteractiveDialog (no longer called)
  - [x] 3.5: Ensure other router tests still pass

- [x] Task 4: Integration validation
  - [x] 4.1: Run full server test suite
  - [x] 4.2: Run full webapp test suite
  - [x] 4.3: Build plugin successfully

## Dev Notes

### Current executeNew() Implementation (To Be Modified)

**From `server/command/router.go:124-220`:**
```go
func (r *Router) executeNew(args *model.CommandArgs) (*model.CommandResponse, error) {
    // Playbook context detection (KEEP)
    if r.playbooksClient != nil {
        run, err := r.playbooksClient.GetPlaybookRunByChannel(args.ChannelId, args.UserId)
        // ... logging ...
    }

    // TriggerId validation (REMOVE - not needed for ephemeral posts)
    if args.TriggerId == "" {
        // ...
    }

    // Site URL / callback URL (REMOVE - API endpoint handles submission)
    siteURL := r.api.GetConfig().ServiceSettings.SiteURL
    callbackURL := fmt.Sprintf("%s/plugins/.../dialog/submit", *siteURL)

    // Native dialog (REPLACE with ephemeral post)
    dialog := model.OpenDialogRequest{...}
    r.api.OpenInteractiveDialog(dialog)
}
```

### Target Implementation

**Replace with:**
```go
func (r *Router) executeNew(args *model.CommandArgs) (*model.CommandResponse, error) {
    // Playbook context detection (KEEP - useful for logging)
    if r.playbooksClient != nil {
        run, err := r.playbooksClient.GetPlaybookRunByChannel(args.ChannelId, args.UserId)
        if err != nil {
            // Only log if it's NOT a "not found" error
            errorMsg := err.Error()
            if !strings.Contains(errorMsg, "not found") && !strings.Contains(errorMsg, "404") {
                r.api.LogWarn("Failed to check for playbook context", ...)
            }
        } else if run != nil {
            r.api.LogDebug("Detected playbook context", ...)
        }
    }

    // Story 11.5: Create ephemeral post to trigger React modal
    // This replaces the native OpenInteractiveDialog approach
    // The ModalTriggerPost component (Story 11.1) handles opening the modal
    post := &model.Post{
        UserId:    args.UserId,
        ChannelId: args.ChannelId,
        Type:      "custom_approval_modal",
        Props: map[string]interface{}{
            "modal_type":   "approval_request",
            "channel_id":   args.ChannelId,
            "team_id":      args.TeamId,
            "trigger_user": args.UserId,
        },
    }

    r.api.SendEphemeralPost(args.UserId, post)

    // Return empty response - modal opens asynchronously via webapp
    return &model.CommandResponse{}, nil
}
```

### ModalTriggerPost Expected Props (Story 11.1)

**From `webapp/src/components/ModalTriggerPost.tsx:30-41`:**
```typescript
// Expected post.props format:
// {
//   modal_type: string;        // Type of modal to open
//   channel_id?: string;       // Channel context
//   team_id?: string;          // Team context
//   trigger_user?: string;     // User who triggered the modal
//   [key: string]: any;        // Additional props
// }
```

### ModalProvider Case (Story 11.1)

**From `webapp/src/context/ModalContext.tsx`:**
```typescript
// The ModalProvider renders modals based on modal_type
// 'approval_request' maps to ApprovalRequestModal
case 'approval_request':
    return <ApprovalRequestModal {...modalProps} />;
```

### API Endpoint Available (Story 11.4)

**The `ApprovalRequestModal` submits to:**
```
POST /plugins/com.mattermost.plugin-approver2/api/v1/approval/new
```

This endpoint was created in Story 11.4 and handles:
- JSON body: `{channel_id, team_id, approver_id, description}`
- Two-layer validation (field + business)
- Self-approval prevention (GitHub Issue #4)
- DM notifications, playbook posting, ephemeral confirmation

### Test Updates Required

**Current tests in `server/command/router_test.go`:**
- `TestExecuteNew_MissingTriggerId` - REMOVE (no longer applies)
- `TestExecuteNew_MissingSiteURL` - REMOVE (no longer applies)
- `TestExecuteNew_DialogOpenFails` - REMOVE (no longer applies)
- `TestExecuteNew_Success` - UPDATE to verify ephemeral post

**New test approach:**
```go
func TestExecuteNew_SendsEphemeralPost(t *testing.T) {
    api := &plugintest.API{}
    router := NewRouter(api, nil, nil)

    args := &model.CommandArgs{
        UserId:    "user123",
        ChannelId: "channel456",
        TeamId:    "team789",
    }

    // Expect SendEphemeralPost to be called
    api.On("SendEphemeralPost", "user123", mock.MatchedBy(func(post *model.Post) bool {
        return post.Type == "custom_approval_modal" &&
            post.Props["modal_type"] == "approval_request" &&
            post.Props["channel_id"] == "channel456" &&
            post.Props["team_id"] == "team789" &&
            post.Props["trigger_user"] == "user123"
    })).Return(&model.Post{Id: "ephemeral123"})

    response, err := router.executeNew(args)

    assert.NoError(t, err)
    assert.Equal(t, &model.CommandResponse{}, response)
    api.AssertExpectations(t)
}
```

### Dependencies on Previous Stories

- **Story 11.1**: `ModalTriggerPost` component registered as `custom_approval_modal` post type
- **Story 11.2**: `UserSelector` component for approver selection
- **Story 11.3**: `ApprovalRequestModal` component with form and validation
- **Story 11.4**: `/api/v1/approval/new` API endpoint for submission

### What This Story Completes

After this story:
1. `/approve new` → Server sends ephemeral post
2. Ephemeral post renders `ModalTriggerPost`
3. `ModalTriggerPost` calls `openModal('approval_request', props)`
4. `ModalProvider` renders `ApprovalRequestModal`
5. User fills form and submits
6. `ApprovalRequestModal` calls `/api/v1/approval/new`
7. API creates approval, sends DM, returns success
8. Modal closes

This is the final integration point connecting all React modal components.

### What's Deferred to Story 11.6

- End-to-end manual validation
- UX comparison with native dialog
- Regression testing for all approval workflows
- Keyboard navigation verification

### References

- [Source: server/command/router.go#executeNew - Current implementation]
- [Source: webapp/src/components/ModalTriggerPost.tsx - Modal trigger component]
- [Source: webapp/src/context/ModalContext.tsx - Modal state management]
- [Source: webapp/src/components/ApprovalRequestModal.tsx - Modal form component]
- [Source: server/api.go#handleApprovalNew - API endpoint]
- [Source: _bmad-output/implementation-artifacts/11-4-api-endpoint-for-react-modal-submission.md - Story 11.4]
- [Source: _bmad-output/implementation-artifacts/epic-11-react-modal-framework.md#story-115 - Epic requirements]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A

### Completion Notes List

- **Task 1-2**: Modified `executeNew()` in `server/command/router.go`:
  - Replaced ~90 lines of native dialog code with ~30 lines of ephemeral post trigger
  - Removed `OpenInteractiveDialog`, `TriggerId` validation, `siteURL`/`callbackURL` logic
  - Added `SendEphemeralPost()` with `custom_approval_modal` type and required props
  - Kept playbook context detection for logging (Story 8.1 compatibility)
  - Added docstring documenting native dialog removal for rollback reference

- **Task 3**: Updated `TestRouteNew` and `TestExecuteNew_PlaybookIntegration` in `server/command/router_test.go`:
  - Replaced 5 old tests (dialog structure, trigger ID, site URL, nil site URL, dialog failure)
  - Added 3 new tests verifying ephemeral post props, no trigger ID requirement, no site URL requirement
  - Updated 4 playbook integration tests to use `SendEphemeralPost` instead of `OpenInteractiveDialog`
  - All 7 new/updated tests pass

- **Task 4**: Integration validation:
  - Server tests: All pass (9 packages)
  - Webapp tests: 234 pass (15 suites)
  - Build: dist/com.mattermost.plugin-approver2-2.3.1+aba2cb4.tar.gz

### File List

**Server (Modified):**
- server/command/router.go - Replaced `executeNew()` implementation

**Server Tests (Modified):**
- server/command/router_test.go - Updated `TestRouteNew`, `TestExecuteNew_PlaybookIntegration`

**Sprint Tracking (Modified):**
- _bmad-output/implementation-artifacts/sprint-status.yaml - Story status updates
