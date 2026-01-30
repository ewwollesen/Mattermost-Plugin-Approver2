# Epic 11: React Modal Framework (Approval Creation Migration)

**Version:** 1.0.0
**Status:** Draft
**Priority:** Medium
**Created:** 2026-01-22
**Related Issues:** Multi-Approval Foundation (Future)

## Overview

Migrate the approval creation modal from Mattermost's native `OpenInteractiveDialog` to a custom React modal component. This is a proof-of-concept migration with **zero new features** - maintaining exact feature parity with the current 1:1 approval workflow to validate the technical approach before adding multi-approval capabilities.

## Background

### Why This Epic Exists

The current approval creation uses Mattermost's native `model.OpenInteractiveDialog`, which provides:
- Text inputs, select dropdowns, text areas
- Built-in form validation and submission handling
- Automatic focus management and keyboard navigation

However, native dialogs have limitations:
- **No dynamic fields** - Cannot add/remove approvers on the fly
- **No complex grouping** - Cannot implement AND/OR approval logic
- **Limited customization** - Standard Mattermost styling only

To implement multi-approval workflows (Epic 12+), we need React-based modals. This epic proves the pattern works by migrating the existing 1:1 flow first.

### What This Epic Proves (POC Goals)

1. **Slash command can trigger React modal** - `/approve new` opens custom webapp modal
2. **Form state management works** - React handles approver selection + description
3. **Validation behaves correctly** - Field errors display, modal stays open
4. **Submit flow is equivalent** - API call, response handling, modal close
5. **UX parity maintained** - Same or better user experience as native dialog

### What This Epic Does NOT Do

- Add multi-approver selection (Epic 12)
- Add AND/OR approval logic (Epic 13+)
- Add sequential/parallel approval chains
- Any new features whatsoever

## Problem Statement

**Current State (v2.3.1):**
- `/approve new` triggers `OpenInteractiveDialog`
- Native Mattermost modal with user select + text area
- Works well for 1:1 approvals
- Cannot support multi-approval workflows

**Desired State (v2.4.0):**
- `/approve new` triggers custom React modal
- Identical UI/UX to native dialog (approver select + description)
- Same validation, same API flow, same result
- Foundation ready for multi-approval (Epic 12)

## Goals

### Primary Goals
1. **React Modal Infrastructure** - Build reusable modal framework in webapp
2. **Slash Command Integration** - `/approve new` opens React modal instead of native
3. **Feature Parity** - Identical user experience to current dialog
4. **Validation Parity** - Same error handling (required fields, self-approval check)
5. **API Parity** - Same backend flow, no server changes to approval creation logic

### Success Metrics
- `/approve new` opens React modal
- Modal looks and behaves identically to native dialog
- Form validation works (required fields, self-approval prevention)
- Successful submission creates approval record
- All existing tests continue to pass
- No regression in any approval functionality

## Technical Approach

### Option Analysis: How to Trigger React Modal from Slash Command

**Option A: WebSocket Custom Event (Recommended)**
```
Slash Command → Server → WebSocket Event → Webapp Listener → Open Modal
```
- Server sends custom WebSocket event with modal trigger
- Webapp listens for event and opens React modal
- Clean separation of concerns
- Similar to how Mattermost handles some plugin interactions

**Option B: Custom Post Type Trigger**
```
Slash Command → Server → Ephemeral Post (type: open_modal) → Webapp Renders Modal
```
- Server creates ephemeral post with `Type: "custom_approval_modal_trigger"`
- Webapp registers post type that renders as modal
- Leverages existing custom post type infrastructure
- Post auto-dismisses or contains modal UI

**Option C: Plugin Webapp Hook**
```
Slash Command → Server → HTTP Response with action → Plugin executes action
```
- Requires understanding Mattermost plugin webapp hooks
- More complex but possibly more "correct" approach

**Recommendation:** Start with Option B (Custom Post Type Trigger) since we already have custom post type infrastructure from Epic 9-10. If that doesn't work well, fall back to Option A.

### Server-Side Changes

```go
// executeNew triggers React modal instead of native dialog
func (r *Router) executeNew(args *model.CommandArgs, subargs []string) (*model.CommandResponse, error) {
    // Create ephemeral post that triggers webapp modal
    post := &model.Post{
        UserId:    r.botUserID,
        ChannelId: args.ChannelId,
        Type:      "custom_approval_modal",
        Props: map[string]interface{}{
            "trigger_user_id": args.UserId,
            "channel_id":      args.ChannelId,
            "team_id":         args.TeamId,
            "trigger_id":      args.TriggerId,  // May need for modal APIs
        },
    }

    // Send as ephemeral post only visible to user
    r.api.SendEphemeralPost(args.UserId, post)

    return &model.CommandResponse{}, nil
}
```

### Client-Side Modal Architecture

```typescript
// webapp/src/components/ApprovalRequestModal.tsx
interface ApprovalRequestModalProps {
    visible: boolean;
    onClose: () => void;
    channelId: string;
    teamId: string;
}

interface FormState {
    approverId: string;
    description: string;
    errors: {
        approver?: string;
        description?: string;
    };
    submitting: boolean;
}

const ApprovalRequestModal: React.FC<ApprovalRequestModalProps> = ({
    visible,
    onClose,
    channelId,
    teamId,
}) => {
    const [form, setForm] = useState<FormState>({...});

    // Client-side validation (mirrors server-side)
    const validate = (): boolean => {
        const errors: FormState['errors'] = {};
        if (!form.approverId) {
            errors.approver = "Approver field is required. Please select a user.";
        }
        if (!form.description.trim()) {
            errors.description = "Description field is required. Please describe what needs approval.";
        }
        // Self-approval check
        if (form.approverId === currentUserId) {
            errors.approver = "You cannot approve your own request. Please select a different approver.";
        }
        setForm(f => ({...f, errors}));
        return Object.keys(errors).length === 0;
    };

    const handleSubmit = async () => {
        if (!validate()) return;

        setForm(f => ({...f, submitting: true}));

        // Call plugin API endpoint
        const response = await fetch(`/plugins/${pluginId}/api/v1/approval/new`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                channel_id: channelId,
                team_id: teamId,
                approver_id: form.approverId,
                description: form.description,
            }),
        });

        if (response.ok) {
            onClose();
        } else {
            const data = await response.json();
            // Handle server-side validation errors
            if (data.errors) {
                setForm(f => ({...f, errors: data.errors, submitting: false}));
            }
        }
    };

    return (
        <Modal visible={visible} onClose={onClose} title="Request Approval">
            <UserSelector
                value={form.approverId}
                onChange={(id) => setForm(f => ({...f, approverId: id}))}
                error={form.errors.approver}
                label="Select Approver *"
            />
            <TextArea
                value={form.description}
                onChange={(text) => setForm(f => ({...f, description: text}))}
                error={form.errors.description}
                label="What needs approval? *"
                placeholder="Describe the action requiring approval"
                maxLength={1000}
            />
            <Button onClick={handleSubmit} loading={form.submitting}>
                Submit
            </Button>
        </Modal>
    );
};
```

### Key Technical Challenges

**Challenge 1: User Selector Component**
- Need to replicate Mattermost's `DataSource: "users"` functionality
- Options: Use Mattermost's internal user search API, or build autocomplete

**Challenge 2: Modal State Management**
- Where does "modal open" state live?
- Need global or context-based state to trigger from post type

**Challenge 3: Styling Consistency**
- Modal should look native to Mattermost
- Use Mattermost's CSS variables and component patterns

## Scope

### In Scope
- React modal component for approval request creation
- User selector component (approver selection)
- Text area component (description input)
- Client-side validation (required fields, self-approval)
- Server-side API endpoint for React modal submission
- Trigger mechanism from `/approve new` slash command
- Feature parity with native dialog
- Error handling and user feedback

### Out of Scope
- Multi-approver selection (Epic 12)
- Approval chain configuration
- Any new features or functionality
- Migration of other dialogs (approve/deny decision modals)
- Mobile-specific handling

## User Stories

### Story 11.1: Modal Infrastructure and Trigger Mechanism

**As a** plugin developer
**I want** infrastructure to open React modals from slash commands
**So that** future features can use custom React UI

**Acceptance Criteria:**

**AC1: Custom Post Type for Modal Trigger**
- Register `custom_approval_modal` post type in webapp
- Post type renders as modal trigger (or invisible)
- Ephemeral post auto-cleans up after modal opens

**AC2: Modal Container Component**
- Create `webapp/src/components/Modal.tsx` base component
- Handles overlay, close on escape, focus trap
- Consistent styling with Mattermost patterns

**AC3: Global Modal State**
- Create modal context or Redux slice
- Track which modal is open and with what props
- Support multiple modal types for future use

**AC4: Trigger Flow**
- `/approve new` → Server creates ephemeral post → Webapp opens modal
- Verify modal opens consistently
- No flicker or double-open issues

---

### Story 11.2: User Selector Component

**As a** user creating an approval request
**I want** to select an approver from a searchable list
**So that** I can easily find the right person

**Acceptance Criteria:**

**AC1: User Search Functionality**
- Autocomplete/typeahead search for users
- Use Mattermost API: `GET /api/v4/users/autocomplete`
- Show display name and username

**AC2: Selection Display**
- Show selected user with avatar and name
- Clear selection button

**AC3: Styling**
- Match Mattermost's native user selector appearance
- Consistent with existing UI patterns

**AC4: Error State**
- Show error message below selector when validation fails
- Red border or highlight on error

---

### Story 11.3: Approval Request Modal Component

**As a** user
**I want** to create approval requests via a React modal
**So that** I have the same experience as the native dialog

**Acceptance Criteria:**

**AC1: Modal Layout**
- Title: "Request Approval"
- Approver selector field (required)
- Description text area field (required)
- Cancel and Submit buttons

**AC2: Field Parity**
- Approver: User selector, required
- Description: Text area, max 1000 chars, required
- Same placeholder text as native dialog

**AC3: Client-Side Validation**
- Required field validation
- Self-approval prevention (GitHub Issue #4)
- Show field-specific errors
- Keep modal open on validation failure

**AC4: Close Behavior**
- Close on Cancel button
- Close on Escape key
- Close on successful submission
- Prompt on close if form has data (optional)

---

### Story 11.4: API Endpoint for React Modal Submission

**As a** server
**I want** an API endpoint that handles React modal submissions
**So that** approvals are created the same way as native dialog

**Acceptance Criteria:**

**AC1: New API Endpoint**
- `POST /plugins/{pluginId}/api/v1/approval/new`
- Accepts JSON body: `{channel_id, team_id, approver_id, description}`
- Returns approval record on success or errors on failure

**AC2: Validation Logic**
- Reuse existing `HandleDialogSubmission()` validation
- Self-approval check (GitHub Issue #4)
- Description length validation
- Approver exists and active check

**AC3: Response Format**
```json
// Success
{ "success": true, "approval": { "code": "ABC-123", ... } }

// Validation Error
{ "success": false, "errors": { "approver": "Error message" } }
```

**AC4: Authorization**
- Verify user is authenticated
- Verify user has access to channel

**AC5: Backward Compatibility**
- Existing dialog submission endpoint still works
- No breaking changes to API

---

### Story 11.5: Server-Side Slash Command Integration

**As a** user
**I want** `/approve new` to open the React modal
**So that** I can create approvals with the new UI

**Acceptance Criteria:**

**AC1: Update executeNew()**
- Modify `server/command/router.go` `executeNew()` function
- Send ephemeral post with `custom_approval_modal` type
- Include necessary context (channel_id, team_id, user_id)

**AC2: Deprecate Native Dialog (Conditional)**
- Keep native dialog code but behind feature flag
- Allow rollback if React modal has issues
- Flag: `EnableReactModal` in plugin settings (optional)

**AC3: No Breaking Changes**
- `/approve new` continues to work
- Just opens React modal instead of native dialog
- All other commands unchanged

---

### Story 11.6: End-to-End Feature Parity Validation

**As a** user
**I want** the React modal to work identically to the native dialog
**So that** my workflow isn't disrupted

**Acceptance Criteria:**

**AC1: Happy Path**
- `/approve new` opens React modal
- Select approver, enter description
- Submit creates approval record
- Approver receives DM notification
- Modal closes on success

**AC2: Validation Flow**
- Empty approver shows error, modal stays open
- Empty description shows error, modal stays open
- Self-approval shows error, modal stays open
- Fix error and resubmit works

**AC3: Server Validation**
- Invalid approver (deleted user) shows error
- Description too long shows error
- All server errors display correctly

**AC4: UX Comparison**
- Modal opens as fast as native dialog
- Keyboard navigation works (Tab, Enter, Escape)
- Looks professional and native to Mattermost

**AC5: Regression Testing**
- All existing approval tests pass
- All DM notifications work correctly
- Playbook integration unchanged
- List, get, cancel commands unchanged

---

## Dependencies

### From Epic 9-10
- Webapp infrastructure (TypeScript/React framework)
- Custom post type registration pattern
- Styling patterns and conventions

### External
- Mattermost Server v6.0+
- Mattermost user autocomplete API
- React 17+, TypeScript

## Risks and Mitigations

**Risk 1: Modal Trigger Mechanism Doesn't Work**
- **Likelihood:** Medium
- **Impact:** High - Epic blocked
- **Mitigation:** Prototype trigger mechanism in Story 11.1 first; have Option A (WebSocket) as fallback

**Risk 2: User Selector Hard to Replicate**
- **Likelihood:** Medium
- **Impact:** Medium - Core UX affected
- **Mitigation:** Research Mattermost's internal user selector; may be able to reuse existing components

**Risk 3: Modal Feels Non-Native**
- **Likelihood:** Low
- **Impact:** Medium - UX degradation
- **Mitigation:** Use Mattermost CSS variables; test with users; iterate on styling

**Risk 4: Performance Degradation**
- **Likelihood:** Low
- **Impact:** Medium
- **Mitigation:** Profile bundle size; lazy load modal component

## Testing Strategy

### Unit Tests
- Modal component renders correctly
- User selector search and selection works
- Validation logic covers all cases
- API endpoint validation

### Integration Tests
- Slash command triggers modal
- Form submission creates approval
- Error responses display correctly

### Manual Testing
- UX comparison with native dialog
- Keyboard navigation
- Different screen sizes
- Error scenarios

### Regression Testing
- Full test suite passes
- All existing approval workflows work
- No changes to approval behavior

## Effort Estimate

| Story | Effort |
|-------|--------|
| 11.1: Modal Infrastructure | 1 day |
| 11.2: User Selector Component | 1.5 days |
| 11.3: Approval Request Modal | 1 day |
| 11.4: API Endpoint | 0.5 day |
| 11.5: Slash Command Integration | 0.5 day |
| 11.6: E2E Validation | 1 day |

**Total: ~5.5 developer days**

## Future Epics Enabled

Once Epic 11 is complete, these become possible:

- **Epic 12: Sequential Multi-Approval** - Add 2-3 approver slots
- **Epic 13: Approval Chain Templates** - Reusable approval patterns
- **Epic 14: AND/OR Approval Logic** - Complex approval chains
- **Epic 15: Delegation and Forwarding** - Approver can delegate

---

**Epic Owner:** Wayne
**Status:** Draft - Awaiting Review
