# Epic 10: DM Notifications with Interactive Buttons (Matterpoll Pattern)

**Version:** 3.1.0
**Status:** Planned
**Priority:** High
**Created:** 2026-01-19
**Related Issues:** GitHub Issue #3 (Timezone), Matterpoll Pattern Analysis

## Overview

Implement improved DM notifications with interactive Approve/Deny buttons using the Matterpoll pattern. This epic applies the proven `model.ParseSlackAttachment()` approach that preserves Integration URLs with custom post types, enabling interactive buttons to work correctly.

## Background

### Why This Epic Exists

During Epic 9 implementation, the dev agent discovered that interactive buttons weren't working correctly when using direct `post.Props` assignment with custom post types. The existing API-based button implementation works but lacks the elegant webapp rendering.

After analyzing the Matterpoll plugin (see `matterpoll-interactive-buttons-analysis.md`), we discovered the critical pattern:
- **`model.ParseSlackAttachment(post, actions)`** preserves Integration URLs even with custom post types
- **`doPostAction(postId, actionId)`** from mattermost-redux handles button clicks automatically

### What Matterpoll Proves

1. Custom post types CAN have working Integration URLs
2. `ParseSlackAttachment()` is critical - don't manually set `props.attachments`
3. Both URL path params and Context maps work for passing state
4. The pattern is production-ready and widely used

## Problem Statement

**Current State (v2.2.0):**
- DM notifications use API-based posts with interactive buttons (working)
- Buttons work via `model.PostAction` with Integration URLs
- Posts render as markdown, not custom webapp components
- Timestamps display in UTC, not user's local timezone

**Desired State:**
- DM notifications render as custom webapp components
- Interactive Approve/Deny buttons work correctly
- Timestamps display in user's local timezone
- Consistent UX with playbook channel posts

## Goals

### Primary Goals
1. **Implement Matterpoll Pattern:** Use `ParseSlackAttachment` for interactive buttons
2. **Custom Post Type for DMs:** Register and render DM notifications as webapp components
3. **Timezone Support:** All DM timestamps in user's local timezone
4. **Interactive Buttons:** Approve/Deny buttons functional with custom post type
5. **Backward Compatibility:** Non-webapp clients see markdown fallback

### Success Metrics
- All DM notification types render as custom components
- Approve/Deny buttons work in DM custom posts
- Timestamps display in user's local timezone
- No functionality regression from v2.2.0

## Technical Approach

### The Matterpoll Pattern (Proven to Work)

**Server Side:**
```go
// 1. Create PostActions with Integration URLs
actions := []*model.PostAction{
    {
        Id:    "approve",
        Name:  "Approve",
        Type:  model.PostActionTypeButton,
        Style: "success",
        Integration: &model.PostActionIntegration{
            URL: fmt.Sprintf("/plugins/%s/api/v1/approval/%s/approve", pluginID, code),
        },
    },
    {
        Id:    "deny",
        Name:  "Deny",
        Type:  model.PostActionTypeButton,
        Style: "danger",
        Integration: &model.PostActionIntegration{
            URL: fmt.Sprintf("/plugins/%s/api/v1/approval/%s/deny", pluginID, code),
        },
    },
}

attachment := &model.SlackAttachment{
    Title:   "Approval Request",
    Text:    description,
    Actions: actions,
}

// 2. Create post with CUSTOM TYPE
post := &model.Post{
    UserId:    botUserID,
    ChannelId: channelID,
    Type:      "custom_approval_dm",
    Props: map[string]interface{}{
        "approval_code": code,
        // ... other approval data
    },
}

// 3. CRITICAL: Use ParseSlackAttachment (not direct Props assignment)
model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})

// 4. Create the post
createdPost, err := p.API.CreatePost(post)
```

**Client Side:**
```typescript
// 1. Register custom post type
registry.registerPostTypeComponent('custom_approval_dm', ApprovalDMPost);

// 2. Render component with buttons
class ApprovalDMPost extends React.PureComponent {
    render() {
        const {post} = this.props;
        const attachment = post.props.attachments?.[0];

        return (
            <div>
                <h3>{attachment.title}</h3>
                <p>{attachment.text}</p>
                {attachment.actions?.map(action => (
                    <button
                        onClick={() => this.handleAction(action.id)}
                        key={action.id}
                    >
                        {action.name}
                    </button>
                ))}
            </div>
        );
    }

    handleAction = (actionId) => {
        // Use mattermost-redux doPostAction
        this.props.actions.doPostAction(this.props.post.id, actionId);
    }
}
```

## Scope

### In Scope
- Convert DM approval request notifications to Matterpoll pattern
- Interactive Approve/Deny buttons with custom post type
- Timezone-aware timestamps in DM posts
- DM outcome notifications (approved, denied)
- DM cancellation notifications
- DM timeout notifications
- DM verification notifications
- Post update when approval status changes
- Markdown fallback for non-webapp clients

### Out of Scope
- Playbook channel posts (already done in Epic 9)
- Real-time updates via WebSocket (future enhancement)
- Mobile-specific optimizations

## User Stories

### Story 10.1: Server-Side Matterpoll Pattern Implementation

**As a** plugin developer
**I want** to implement the Matterpoll pattern for creating interactive posts
**So that** interactive buttons work correctly with custom post types

**Acceptance Criteria:**

**AC1: Helper Function for Interactive Posts**
- Create `server/notifications/interactive_post.go`
- Implement `CreateInteractiveApprovalPost()` function
- Uses `model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})`
- Sets `post.Type = "custom_approval_dm"`
- Includes all approval data in `post.Props`
- Stores timestamps as Unix millis (int64)

**AC2: PostAction Structure**
- Create PostActions for Approve, Deny buttons
- Integration URLs: `/plugins/{pluginID}/api/v1/approval/{code}/approve|deny`
- Button styles: `success` for Approve, `danger` for Deny
- Include approval code in URL path (not Context map)

**AC3: SlackAttachment Structure**
- Title: Status header (e.g., "Approval Request")
- Text: Approval description and details
- Actions: Array of PostAction buttons
- No custom fields needed (use post.Props for data)

**AC4: Post Props Schema**
```go
Props: map[string]interface{}{
    "approval_code":           record.Code,
    "approval_status":         record.Status,
    "requester_username":      record.RequesterUsername,
    "requester_display_name":  record.RequesterDisplayName,
    "approver_username":       record.ApproverUsername,
    "approver_display_name":   record.ApproverDisplayName,
    "description":             record.Description,
    "created_at":              record.CreatedAt,      // Unix millis
    "decided_at":              record.DecidedAt,      // Unix millis
    "decision_comment":        record.DecisionComment,
    "notification_type":       "approval_request",    // or "outcome", "cancellation", etc.
    "is_dm":                   true,
}
```

**AC5: Markdown Fallback**
- Set `post.Message` with markdown content for non-webapp clients
- Include all essential information in markdown
- Reuse existing formatter functions from `notifications/dm.go`

**Technical Notes:**
- This is the foundation for all DM notifications
- The `ParseSlackAttachment` call is CRITICAL for button functionality
- Don't use `model.SlackAttachment` directly in Props - use the helper function

---

### Story 10.2: Update API Handlers for Matterpoll Pattern

**As a** plugin server
**I want** API handlers to work with the Matterpoll button pattern
**So that** Approve/Deny button clicks are processed correctly

**Acceptance Criteria:**

**AC1: PostActionIntegrationRequest Handling**
- API handlers already receive `PostActionIntegrationRequest` from Mattermost
- Verify existing `/api/v1/approval/{code}/approve` and `/api/v1/approval/{code}/deny` handlers work
- Extract user ID from request for permission validation
- No changes needed if handlers already exist

**AC2: PostActionIntegrationResponse**
- Return proper `PostActionIntegrationResponse` struct
- Include `EphemeralText` for user feedback
- Include `Update` field with updated post content
- Call `ParseSlackAttachment` on updated post

**AC3: Post Update on Decision**
- When approval is approved/denied, update the DM post
- Remove Approve/Deny buttons from updated post
- Show decision status and timestamp
- Use `p.API.UpdatePost()` with new props

**AC4: Handle Edge Cases**
- Already decided: Return ephemeral message "This approval has already been decided"
- Invalid approval code: Return appropriate error
- Permission denied: Return error if user is not the approver

**Technical Notes:**
- Handlers should already work - this story validates and adjusts if needed
- The key is ensuring `PostActionIntegrationResponse.Update` uses `ParseSlackAttachment`

---

### Story 10.3: Convert Approval Request DM to Matterpoll Pattern

**As an** approver
**I want** to receive approval request DMs as interactive webapp components
**So that** I can approve/deny with buttons and see timestamps in my timezone

**Acceptance Criteria:**

**AC1: Update SendApprovalRequestDM()**
- Modify `server/notifications/dm.go` `SendApprovalRequestDM()` function
- Use `CreateInteractiveApprovalPost()` helper from Story 10.1
- Include Approve and Deny buttons
- Set `notification_type: "approval_request"`

**AC2: DM Content**
- Status header with emoji
- Requester information with @mention
- Request description (full text)
- Request ID
- Created timestamp (will be converted to local timezone by webapp)

**AC3: Button Functionality**
- Approve button triggers `/api/v1/approval/{code}/approve`
- Deny button triggers `/api/v1/approval/{code}/deny`
- Both buttons open decision modal (existing behavior)
- Modal submission completes the decision

**AC4: Backward Compatibility**
- Markdown fallback in `post.Message` includes all info
- Non-webapp clients can still see approval details
- Links to approve/deny still work (if present)

**AC5: Testing**
- Create approval, verify DM renders as custom component
- Approve button works, opens modal
- Deny button works, opens modal
- Post updates after decision

---

### Story 10.4: Webapp Component for DM Notifications

**As a** webapp
**I want** to render DM approval notifications as custom components
**So that** users see timezone-aware timestamps and interactive buttons

**Acceptance Criteria:**

**AC1: Register DM Post Type**
- In `webapp/src/index.tsx`, register:
```typescript
registry.registerPostTypeComponent('custom_approval_dm', ApprovalDMPost);
```

**AC2: ApprovalDMPost Component**
- Create `webapp/src/components/ApprovalDMPost.tsx`
- Extract attachment from `post.props.attachments[0]`
- Extract approval data from `post.props`
- Render based on `notification_type` prop

**AC3: Notification Type Rendering**
- `approval_request`: Show request details + Approve/Deny buttons
- `outcome`: Show decision details (approved/denied)
- `cancellation`: Show cancellation notice
- `timeout`: Show timeout notice
- `verification`: Show verification confirmation

**AC4: Button Rendering**
- Extract buttons from `attachment.actions`
- Render with appropriate styles (success/danger)
- On click, call `doPostAction(postId, actionId)`
- Hide buttons for non-pending statuses

**AC5: Timestamp Rendering**
- Use `Timestamp` component from Epic 9
- Extract `created_at`, `decided_at` from `post.props`
- Display in user's local timezone

**AC6: Connect to Redux**
```typescript
const mapDispatchToProps = (dispatch) => ({
    actions: {
        doPostAction: (postId, actionId) =>
            dispatch(doPostAction(postId, actionId)),
    },
});
```

**Technical Notes:**
- Reuse `Timestamp` component from Epic 9
- Use `doPostAction` from `mattermost-redux/actions/posts`
- Component handles all DM notification types

---

### Story 10.5: Convert Outcome Notifications to Matterpoll Pattern

**As a** requester
**I want** to receive approval outcome DMs as webapp components
**So that** I see the decision with timestamps in my local timezone

**Acceptance Criteria:**

**AC1: Update SendOutcomeNotificationDM()**
- Modify to use `CreateInteractiveApprovalPost()` helper
- Set `notification_type: "outcome"`
- Include decision status (approved/denied)
- Include decision timestamp and comment

**AC2: Outcome Content**
- Status header: "Approval Approved" or "Approval Denied"
- Approver information
- Decision timestamp (local timezone via webapp)
- Original request reference
- Decision comment (if provided)

**AC3: No Interactive Buttons**
- Outcome notifications are read-only
- No actions array in SlackAttachment
- Just display information

**AC4: Backward Compatibility**
- Markdown fallback includes all outcome details
- Works for non-webapp clients

---

### Story 10.6: Convert Cancellation Notifications to Matterpoll Pattern

**As an** approver
**I want** to receive cancellation DMs as webapp components
**So that** I know the request was canceled with accurate timestamps

**Acceptance Criteria:**

**AC1: Update SendCancellationNotificationDM()**
- Use `CreateInteractiveApprovalPost()` helper
- Set `notification_type: "cancellation"`
- Include cancellation reason and timestamp

**AC2: Cancellation Content**
- Status header: "Approval Canceled"
- Request ID and description
- Cancellation reason
- Canceled timestamp (local timezone)
- Requester info

**AC3: Update Original DM Post**
- Call `UpdateApprovalPostForCancellation()`
- Update original DM to show "Request Canceled"
- Remove Approve/Deny buttons
- Show cancellation timestamp

---

### Story 10.7: Convert Timeout Notifications to Matterpoll Pattern

**As a** requester
**I want** to receive timeout DMs as webapp components
**So that** I know my request timed out with accurate timestamps

**Acceptance Criteria:**

**AC1: Update SendTimeoutNotificationDM()**
- Use `CreateInteractiveApprovalPost()` helper
- Set `notification_type: "timeout"`
- Include timeout timestamp

**AC2: Timeout Content**
- Status header: "Approval Timed Out"
- Request ID and description
- Approver info (no response received)
- Timeout reason
- Auto-canceled timestamp (local timezone)

---

### Story 10.8: Convert Verification Notifications to Matterpoll Pattern

**As an** approver
**I want** to receive verification DMs as webapp components
**So that** I know the action was verified with accurate timestamps

**Acceptance Criteria:**

**AC1: Update SendVerificationNotificationDM()**
- Use `CreateInteractiveApprovalPost()` helper
- Set `notification_type: "verification"`
- Include verification timestamp and comment

**AC2: Verification Content**
- Status header: "Action Verified Complete"
- Request ID
- Requester info
- Verified timestamp (local timezone)
- Verification comment (if provided)

---

### Story 10.9: End-to-End DM Flow Validation

**As a** user
**I want** all DM notifications to work correctly with the new pattern
**So that** I have a consistent, timezone-aware experience

**Acceptance Criteria:**

**AC1: Approval Request Flow**
- Create approval via `/approve new`
- Approver receives DM as custom component
- Timestamps in approver's local timezone
- Approve/Deny buttons functional
- Decision modal works
- Requester receives outcome DM as custom component

**AC2: Cancellation Flow**
- Requester cancels via `/approve cancel`
- Approver receives cancellation DM as custom component
- Original DM post updated (buttons removed)
- Timestamps accurate

**AC3: Timeout Flow**
- Approval times out
- Requester receives timeout DM as custom component
- Original approver DM updated
- Timestamps accurate

**AC4: Verification Flow**
- Requester verifies via `/approve verify`
- Approver receives verification DM as custom component
- Timestamp in approver's timezone

**AC5: Cross-Timezone Testing**
- User A (PST) creates approval for User B (EST)
- All timestamps display correctly in respective timezones
- No off-by-one or DST errors

**AC6: Backward Compatibility**
- Webapp clients: Custom components with interactive buttons
- Non-webapp clients: Markdown fallback with all information
- All existing approval functionality preserved

**AC7: Regression Testing**
- All v2.2.0 functionality works
- No breaking changes
- Existing tests pass

---

## Dependencies

### From Epic 9
- Webapp infrastructure (Stories 9.1-9.3)
- `Timestamp` component (Story 9.4)
- UI component library (Story 9.5)
- Build pipeline configured

### External
- Mattermost Server v6.0+
- `mattermost-redux` for `doPostAction`

## Risks and Mitigations

**Risk 1: Button Click Not Reaching Handler**
- **Likelihood:** Low (Matterpoll pattern proven)
- **Mitigation:** Exact replication of Matterpoll's approach

**Risk 2: Post Update Race Conditions**
- **Likelihood:** Medium
- **Mitigation:** Use optimistic locking with `KVCompareAndSet` if needed

**Risk 3: Webapp Not Loading**
- **Likelihood:** Low
- **Mitigation:** Markdown fallback ensures functionality

## Testing Strategy

### Unit Tests
- `CreateInteractiveApprovalPost()` generates correct structure
- `ParseSlackAttachment` called correctly
- Props populated with all required fields
- Timestamps stored as int64

### Integration Tests
- Button clicks reach API handlers
- `PostActionIntegrationResponse` updates post
- Custom component renders correctly
- `doPostAction` triggers Integration URL

### Manual Testing
- Full approval flow with buttons
- All notification types render correctly
- Cross-timezone validation
- Mobile fallback verification

## Effort Estimate

| Story | Effort |
|-------|--------|
| 10.1: Server-Side Helper | 1 day |
| 10.2: API Handler Validation | 0.5 day |
| 10.3: Approval Request DM | 1 day |
| 10.4: Webapp Component | 1.5 days |
| 10.5: Outcome Notifications | 0.5 day |
| 10.6: Cancellation Notifications | 0.5 day |
| 10.7: Timeout Notifications | 0.5 day |
| 10.8: Verification Notifications | 0.5 day |
| 10.9: End-to-End Validation | 1 day |

**Total: ~7 developer days**

## Appendix: Matterpoll Pattern Reference

See `matterpoll-interactive-buttons-analysis.md` for the complete analysis of how Matterpoll implements interactive buttons with custom post types.

**Key Takeaways:**
1. `model.ParseSlackAttachment(post, attachments)` is CRITICAL
2. Integration URLs must be in format `/plugins/{pluginID}/api/v1/...`
3. `doPostAction(postId, actionId)` from mattermost-redux handles button clicks
4. Return `PostActionIntegrationResponse` with `Update` field for post updates
5. Custom post types DO work with interactive buttons when using this pattern

---

**Epic Owner:** Wayne
**Status:** Ready for Development
