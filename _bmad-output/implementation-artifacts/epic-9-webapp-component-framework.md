# Epic 9: Webapp Component Framework & Timezone-Aware Posts

**Version:** 3.0.0
**Status:** Completed (Stories 9.1-9.9)
**Priority:** Critical (User #1 Priority)
**Created:** 2026-01-18
**Completed:** 2026-01-19
**Related Issues:** GitHub Issue #3

> **Scope Adjustment:** Stories 9.10-9.11 (DM notification conversion) were **canceled** and moved to Epic 10. The dev agent discovered that implementing interactive buttons via the webapp framework required a different pattern (Matterpoll's `ParseSlackAttachment` + `doPostAction` approach). Epic 9's primary goal of improving Playbooks integration with timezone-aware posts was successfully achieved.

## Overview

Establish webapp infrastructure for the Approval Workflow Plugin to enable rich, timezone-aware post rendering with custom React components. This epic transitions the plugin from server-only markdown formatting to client-side rendering, unlocking better UX, proper timezone handling, and foundation for future enhancements (collapsible sections, mobile optimization, multiple approvers, etc.).

## Problem Statement

**Current Issues:**
1. **Timezone Problem:** Timestamps display in UTC instead of user's local timezone (GitHub Issue #3)
2. **Limited Formatting:** Markdown tables are functional but not optimal for screen real estate
3. **No Client-Side Interactivity:** Can't implement collapsible sections, tooltips, or dynamic updates
4. **Mobile Experience:** Markdown tables don't adapt well to narrow screens
5. **Future Limitations:** Multiple approvers, progress indicators, rich status badges blocked by lack of webapp

**User Impact:**
- Users in different timezones see confusing UTC timestamps
- Playbook channel posts take significant vertical space
- Poor mobile experience for incident responders
- Can't implement advanced UX features without webapp foundation

## Goals

### Primary Goals
1. **Webapp Infrastructure:** Set up React/TypeScript webapp with build pipeline
2. **Timezone Support:** Display all timestamps in user's local timezone
3. **Custom Post Components:** Create reusable approval post components
4. **Playbook Post Conversion:** Convert playbook channel posts to webapp components
5. **No Breaking Changes:** v1.0 and v2.x behavior preserved for non-webapp clients

### Secondary Goals (Enabled for Future)
- Collapsible approval details
- Responsive mobile layouts
- Rich status badges matching Playbooks style
- Client-side filtering/sorting
- Real-time status updates

### Success Metrics
- All timestamps display in user's local timezone configuration
- Playbook posts render as custom components with improved UX
- No functionality lost from markdown format
- Build pipeline successful and documented
- Foundation ready for future webapp enhancements

## Scope

### In Scope (Completed - Stories 9.1-9.9)
✅ Webapp infrastructure setup (React, TypeScript, webpack)
✅ plugin.json updates for webapp registration
✅ Timezone-aware timestamp component
✅ Custom post type for approval messages
✅ Playbook channel post conversion (pending, approved, denied, canceled, timeout)
✅ Mattermost theme integration (no custom styling)
✅ Build and deployment pipeline

### Canceled (Moved to Epic 10)
❌ DM notification conversion (ALL approval notifications) - *Requires Matterpoll pattern for interactive buttons*

### Out of Scope (Future Epics)
❌ Collapsible sections (future enhancement)
❌ Mobile-specific optimizations beyond responsive layout
❌ Real-time updates (future enhancement)
❌ Custom theming or branding

### Phasing Strategy

**Phase 1: Infrastructure (Stories 9.1-9.3)**
- Set up webapp directory structure
- Configure build pipeline
- Register webapp in plugin.json
- Create basic "hello world" component

**Phase 2: Component Framework (Stories 9.4-9.6)**
- Create ApprovalPost base component
- Implement timezone-aware Timestamp component
- Build reusable UI components (StatusBadge, UserMention, etc.)

**Phase 3: Playbook Post Conversion (Stories 9.7-9.9)**
- Register custom post type for approval posts
- Convert playbook status posts to webapp components
- Update server-side to use custom post type for playbook channels
- Ensure backward compatibility

**Phase 4: DM Notification Conversion (Stories 9.10-9.11) - CANCELED**
- ❌ Moved to Epic 10 - Requires Matterpoll pattern (`ParseSlackAttachment` + `doPostAction`)
- See Epic 10 for improved DM UI implementation

## Technical Decisions

### Architecture Decisions

**Decision 1: React + TypeScript**
- **Rationale:** Mattermost standard, type safety, better IDE support
- **Alternative Considered:** Preact (smaller bundle) - rejected for Mattermost compatibility
- **Impact:** Larger bundle size, but standard tooling and community support

**Decision 2: Webpack for Build**
- **Rationale:** Mattermost plugin standard, existing build/setup.mk support
- **Alternative Considered:** Vite/esbuild - rejected for consistency with Mattermost ecosystem
- **Impact:** Slower build times, but proven compatibility

**Decision 3: Use Mattermost Component Library**
- **Rationale:** Consistent UX, automatic theme support, no custom CSS needed
- **Alternative Considered:** Custom components - rejected per Wayne's requirement "stick to Mattermost theme"
- **Impact:** Limited design flexibility, but zero visual inconsistency

**Decision 4: Custom Post Type for Playbook Posts Only**
- **Rationale:** Phase 1 focuses on playbook channel posts (highest value, visible to teams)
- **Alternative Considered:** Convert all posts immediately - rejected for scope management
- **Impact:** DM posts remain markdown (acceptable for 1:1 communication)

**Decision 5: Timezone via mattermost-redux**
- **Rationale:** Use Mattermost's existing timezone selector and moment-timezone
- **Alternative Considered:** Browser timezone only - rejected for consistency with Mattermost settings
- **Impact:** Requires mattermost-redux dependency, but proper integration

**Decision 6: Store Timestamps as Unix Millis in Props**
- **Rationale:** Prevents server-side timezone formatting, enables client-side conversion
- **Alternative Considered:** Keep current string format - rejected because it's the root problem
- **Impact:** Requires server-side changes to store timestamps in post.props

**Decision 7: No Backward Compatibility for Old Posts**
- **Rationale:** Wayne confirmed "No we don't need to make it backwards compatible at all"
- **Alternative Considered:** Dual rendering logic - rejected to keep codebase simple
- **Impact:** Pre-v3.0.0 playbook posts display as markdown (acceptable)

### Technology Stack

**Frontend:**
- React 17+ (Mattermost compatibility)
- TypeScript 4.9+
- moment-timezone (for timezone conversion)
- mattermost-redux (for timezone selectors)
- @mattermost/client (for API types)

**Build Tools:**
- Webpack 5
- Babel (JSX/TS transpilation)
- ts-loader (TypeScript)
- css-loader / style-loader (if needed)

**Development:**
- ESLint (Mattermost rules)
- Prettier (code formatting)
- TypeScript strict mode

## User Stories

### Story 9.1: Webapp Infrastructure Setup

**As a** plugin developer
**I want** to set up the webapp directory structure and build pipeline
**So that** I can develop React components for the plugin

**Acceptance Criteria:**

**AC1: Directory Structure**
- Create `webapp/` directory at project root
- Create `webapp/src/` for source files
- Create `webapp/src/components/` for React components
- Create `webapp/src/types/` for TypeScript types
- Create `webapp/src/index.tsx` as entry point

**AC2: Package Dependencies**
- Create `webapp/package.json` with dependencies:
  - react, react-dom (version compatible with Mattermost)
  - typescript, @types/react, @types/react-dom
  - moment-timezone
  - mattermost-redux (for timezone selectors)
- Add devDependencies:
  - webpack, webpack-cli
  - babel, babel-loader, ts-loader
  - eslint, prettier

**AC3: Build Configuration**
- Create `webapp/webpack.config.js` for production builds
- Output: `webapp/dist/main.js` (plugin bundle)
- Configure TypeScript compilation with `webapp/tsconfig.json`
- Enable source maps for development

**AC4: Build Integration**
- Makefile automatically detects webapp (via HAS_WEBAPP)
- `make` builds both server and webapp
- `make deploy` includes webapp bundle in plugin archive
- Build succeeds without errors

**AC5: Basic Entry Point**
- Create `webapp/src/index.tsx` with plugin initialization
- Export `initialize()` function per Mattermost plugin API
- Register plugin with Mattermost registry
- Verify plugin loads in browser console (no errors)

**Technical Notes:**
- Follow mattermost-plugin-starter-template webapp structure
- Use Mattermost's recommended webpack configuration
- Ensure webapp/ is excluded from server build
- Update .gitignore for node_modules and dist/

---

### Story 9.2: Plugin Manifest Webapp Registration

**As a** plugin
**I want** to register the webapp bundle in plugin.json
**So that** Mattermost loads my React components

**Acceptance Criteria:**

**AC1: plugin.json Updates**
- Add `webapp` section to plugin.json:
```json
"webapp": {
    "bundle_path": "webapp/dist/main.js"
}
```
- Verify `build/bin/manifest has_webapp` returns "true"

**AC2: Plugin Loading**
- After `make` and `make deploy`, plugin loads successfully
- Browser console shows webapp initialization
- No JavaScript errors in console
- Webapp version logged for debugging

**AC3: Development Workflow**
- `make watch` rebuilds webapp on file changes (if supported)
- `make clean` removes webapp/dist/
- `make` builds both server and webapp sequentially

**Technical Notes:**
- The build/setup.mk automatically detects webapp via manifest tool
- Ensure bundle_path matches webpack output path
- Verify plugin archive includes webapp/dist/main.js

---

### Story 9.3: Hello World Component (Verification)

**As a** plugin developer
**I want** to create a simple test component
**So that** I can verify the webapp pipeline works end-to-end

**Acceptance Criteria:**

**AC1: Test Component**
- Create `webapp/src/components/HelloWorld.tsx`
- Component renders "Approval Plugin Webapp Loaded"
- Component displays current timestamp in user's timezone

**AC2: Component Registration**
- Register component via `registry.registerRootComponent()`
- Component appears somewhere in Mattermost UI (e.g., right-hand sidebar or channel header)
- Component updates every second (proves reactivity)

**AC3: Timezone Display**
- Use `getCurrentTimezone()` from mattermost-redux
- Display: "Current time: {formatted timestamp in user's timezone}"
- Verify timezone changes when user changes Mattermost timezone setting

**AC4: Verification**
- Build succeeds
- Plugin loads without errors
- Component visible in Mattermost UI
- Timestamp updates dynamically
- User can verify timezone by changing Mattermost setting

**Technical Notes:**
- This is a throwaway component for verification only
- Will be removed once real components are built
- Proves: build pipeline, React rendering, timezone integration, reactivity

---

### Story 9.4: Timestamp Component with Timezone Support

**As a** approval post component
**I want** a reusable Timestamp component that displays times in user's timezone
**So that** all approval timestamps respect user preferences

**Acceptance Criteria:**

**AC1: Timestamp Component**
- Create `webapp/src/components/Timestamp.tsx`
- Props interface:
```typescript
interface TimestampProps {
    unixMillis: number;       // Unix timestamp in milliseconds
    format?: string;          // moment format string (default: 'lll')
    relative?: boolean;       // Show relative time ("5 minutes ago")
}
```

**AC2: Timezone Conversion**
- Use `getCurrentTimezone()` from `mattermost-redux/selectors/entities/timezone`
- Convert Unix timestamp to user's timezone using moment-timezone
- Handle null/undefined timezone (fallback to browser timezone)
- Update when user changes timezone setting (React to selector changes)

**AC3: Format Options**
- Default format: "Jan 18, 2026 10:30 AM" (moment 'lll')
- Relative format: "5 minutes ago" (moment.fromNow())
- Custom format: Accept moment format strings
- Display timezone abbreviation on hover (tooltip)

**AC4: Edge Cases**
- Handle 0 or null timestamps (display "Not yet decided" or similar)
- Handle invalid timestamps gracefully
- Handle missing timezone data
- Performance: memo component to prevent unnecessary re-renders

**AC5: Unit Tests**
- Test timezone conversion accuracy
- Test format options
- Test edge cases (0, null, invalid)
- Test relative time updates

**Technical Notes:**
- Use React.memo() for performance
- Use useMemo() for timestamp calculations
- Consider using Mattermost's built-in timestamp utilities if available
- This component will be used throughout all approval posts

---

### Story 9.5: Approval Post UI Components Library

**As a** approval post developer
**I want** reusable UI components for common approval elements
**So that** I can build consistent approval posts quickly

**Acceptance Criteria:**

**AC1: StatusBadge Component**
- Create `webapp/src/components/StatusBadge.tsx`
- Props: `status: 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout'`
- Renders emoji + text:
  - Pending: ⏳ Approval Pending
  - Approved: ✅ Approval Approved
  - Denied: ❌ Approval Denied
  - Canceled: 🚫 Approval Canceled
  - Timeout: ⏱️ Approval Timed Out
- Uses Mattermost heading styles

**AC2: UserMention Component**
- Create `webapp/src/components/UserMention.tsx`
- Props: `username: string, displayName?: string`
- Renders: `@username` (clickable mention if possible)
- Falls back to plain text if mention system not available

**AC3: InfoRow Component**
- Create `webapp/src/components/InfoRow.tsx`
- Props: `label: string, value: string | ReactNode, icon?: string`
- Renders key-value pair with consistent styling
- Example: "Request ID: A-SSJEQZ"

**AC4: Component Library Organization**
- All components in `webapp/src/components/`
- Each component in its own file
- Export from `webapp/src/components/index.ts` for easy imports
- TypeScript interfaces for all props

**AC5: Mattermost Theme Integration**
- Use Mattermost CSS variables for colors
- No custom styling beyond layout
- Respect light/dark theme automatically
- Components render consistently across themes

**Technical Notes:**
- Keep components simple and focused (single responsibility)
- Use Mattermost's typography and spacing standards
- These components will be used in Story 9.7-9.9

---

### Story 9.6: ApprovalPost Base Component

**As a** plugin developer
**I want** a base component for all approval posts
**So that** I have a consistent structure for rendering approval data

**Acceptance Criteria:**

**AC1: ApprovalPost Component**
- Create `webapp/src/components/ApprovalPost.tsx`
- Props interface:
```typescript
interface ApprovalPostData {
    code: string;
    description: string;
    status: 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout';
    requesterUsername: string;
    requesterDisplayName: string;
    approverUsername: string;
    approverDisplayName: string;
    createdAt: number;           // Unix millis
    decidedAt?: number;          // Unix millis
    decisionComment?: string;
    note?: string;               // For approved posts
}
```

**AC2: Component Structure**
- StatusBadge at top
- InfoRow for Request ID
- InfoRow for Description (truncated to 80 chars, expandable in future)
- UserMention for Approver/Requester (depending on status)
- Timestamp(s) using Timestamp component
- Optional Note row (if present)

**AC3: Status-Specific Rendering**
- **Pending:** Show "Awaiting: @approver" + created timestamp
- **Approved:** Show "Approved By: @approver" + decided timestamp + optional note
- **Denied:** Show "Denied By: @approver" + decided timestamp + optional reason
- **Canceled:** Show cancellation reason
- **Timeout:** Show "Approver: @approver (no response)"

**AC4: Layout and Styling**
- Compact vertical layout (minimize screen real estate per Wayne's feedback)
- Responsive (adapts to narrow mobile screens)
- Uses Mattermost theme colors and spacing
- Consistent with Playbooks post styling

**AC5: Accessibility**
- Semantic HTML structure
- ARIA labels where appropriate
- Keyboard navigable (if interactive elements added later)
- Screen reader friendly

**Technical Notes:**
- This is the main component that will be registered as a custom post type
- Should be performant (memoized, no unnecessary re-renders)
- Extract common logic into hooks if needed

---

### Story 9.7: Register Custom Post Type

**As a** plugin
**I want** to register a custom post type for approval messages
**So that** Mattermost renders approval posts using my React component

**Acceptance Criteria:**

**AC1: Post Type Registration**
- In `webapp/src/index.tsx`, register custom post type:
```typescript
registry.registerPostTypeComponent('custom_approval', ApprovalPost);
```
- Post type ID: `custom_approval`
- Component: ApprovalPost from Story 9.6

**AC2: Post Type Detection**
- Webapp detects posts with `Type: "custom_approval"`
- ApprovalPost component receives post object
- Component extracts approval data from `post.props`

**AC3: Props Mapping**
- Server stores approval data in `post.props`:
```go
Props: map[string]interface{}{
    "approval_code": record.Code,
    "approval_status": record.Status,
    "requester_username": record.RequesterUsername,
    "requester_display_name": record.RequesterDisplayName,
    "approver_username": record.ApproverUsername,
    "approver_display_name": record.ApproverDisplayName,
    "description": record.Description,
    "created_at": record.CreatedAt,        // Unix millis
    "decided_at": record.DecidedAt,        // Unix millis
    "decision_comment": record.DecisionComment,
    "note": record.DecisionComment,        // For approved posts
}
```

**AC4: Fallback Rendering**
- If webapp not loaded, post displays message text (markdown fallback)
- Server sets `post.Message` with markdown table (v2.x format)
- Ensures non-webapp clients can still view approval info

**AC5: Integration Test**
- Create test approval in playbook channel
- Post renders using ApprovalPost component (not markdown)
- All data displays correctly
- Timestamps show in user's timezone
- Component updates when post props change

**Technical Notes:**
- registerPostTypeComponent is a Mattermost Plugin API method
- Post component receives full post object + theme + ... as props
- Server must set both Type and Props for custom post rendering

---

### Story 9.8: Server-Side Post Type Updates for Playbook Posts

**As a** server
**I want** to create approval posts with custom post type for playbook channels
**So that** webapp renders them as rich components

**Acceptance Criteria:**

**AC1: Update playbooks/client.go**
- Modify `PostMessageToPlaybookChannel()` to accept approval record
- Set `post.Type = "custom_approval"`
- Populate `post.Props` with approval data (see Story 9.7 AC3)
- Set `post.Message` with markdown table as fallback

**AC2: Update playbooks/formatters.go**
- Keep existing formatters for markdown fallback messages
- Create new function: `FormatApprovalPropsForWebapp(record)` → map[string]interface{}
- Include all required fields with proper types
- Ensure timestamps are int64 (Unix millis), not strings

**AC3: Update Call Sites**
- `server/api.go`: Update approval creation to use custom post type
- `server/api.go`: Update approve/deny handlers to use UpdatePost with new props
- `server/timeout/checker.go`: Update timeout handler to use custom post type
- All updates maintain backward compatibility (markdown fallback)

**AC4: UpdatePost for Status Changes**
- When approval status changes (approved, denied, canceled, timeout)
- Call `UpdateMessageInPlaybookChannel()` with updated record
- Update `post.Props` with new status, timestamps, comments
- Webapp component re-renders automatically with new data

**AC5: Validation and Error Handling**
- Validate all props before setting (non-nil, correct types)
- Log errors if props are invalid
- Fall back to markdown-only post if webapp props fail
- Ensure existing v2.x behavior preserved if custom post type fails

**AC6: Unit Tests**
- Test props population from approval record
- Test custom post type set correctly
- Test markdown fallback message generation
- Test UpdatePost with status changes

**Technical Notes:**
- This bridges server (Go) and webapp (React)
- Props must be JSON-serializable
- Timestamps MUST be int64 (not formatted strings)
- Maintain backward compatibility for non-webapp clients

---

### Story 9.9: End-to-End Testing and Validation

**As a** user
**I want** approval posts in playbook channels to display with proper timezones and formatting
**So that** I can see accurate local times without manual conversion

**Acceptance Criteria:**

**AC1: Approval Creation Flow**
- Run `/approve new` in playbook channel
- Select approver, provide description, submit
- Post appears in playbook channel as custom component (not markdown table)
- Post shows:
  - ⏳ Approval Pending header
  - Request ID
  - Description
  - Awaiting: @approver
  - Timestamp in user's local timezone

**AC2: Approval Decision Flow**
- Approver clicks Approve/Deny in DM
- Playbook channel post updates (not new post - same behavior as v2.x)
- Post shows:
  - ✅ Approval Approved (or ❌ Denied)
  - Approved By: @approver
  - Time: {local timezone}
  - Note: {approval comment if provided}

**AC3: Timeout Flow**
- Create approval, wait 30+ minutes (or manually trigger timeout)
- Playbook post updates to show timeout status
- Shows: ⏱️ Approval Timed Out
- Shows: Approver: @approver (no response)

**AC4: Timezone Verification**
- User A (PST timezone) sees timestamps in PST
- User B (EST timezone) sees same approval with timestamps in EST
- Timestamps are accurate (no off-by-one errors, DST handled)
- Hover shows timezone abbreviation

**AC5: Cross-Client Compatibility**
- Webapp client: Sees custom React components
- Mobile client (if no webapp support): Sees markdown fallback
- Desktop client: Sees custom components
- All clients show correct information (no data loss)

**AC6: Performance**
- Post rendering is fast (< 100ms)
- No memory leaks (component unmounts cleanly)
- Scrolling playbook channel is smooth
- Timezone calculation doesn't block UI

**AC7: Regression Testing**
- v1.0 behavior: All core approval functionality preserved
- v2.x behavior: Playbook detection still works
- GitHub Issue #2: No unwanted Playbooks API side effects
- All existing tests pass

**Technical Notes:**
- This story validates the playbook integration end-to-end
- Test with multiple users in different timezones
- Test on web, desktop, mobile (if possible)
- Verify no console errors, no visual glitches
- Note: This story validates playbook posts only; DM validation in Story 9.11

---

### Story 9.10: Convert DM Notifications to Custom Post Type - ⛔ CANCELED

> **CANCELED:** Moved to Epic 10. The dev agent discovered that implementing interactive buttons with custom post types requires the Matterpoll pattern (`model.ParseSlackAttachment` + `doPostAction`). Epic 10 will implement this properly.

**As a** user
**I want** all DM notifications to use webapp components with timezone-aware timestamps
**So that** I see consistent local times in all approval communications

**Acceptance Criteria:**

**AC1: Update notifications/dm.go Functions**
- Modify `SendApprovalRequestDM()` to use custom post type
- Modify `SendOutcomeNotificationDM()` to use custom post type
- Modify `SendCancellationNotificationDM()` to use custom post type
- Modify `SendTimeoutNotificationDM()` to use custom post type
- Modify `SendRequesterCancellationNotificationDM()` to use custom post type
- Modify `SendVerificationNotificationDM()` to use custom post type
- All functions set `post.Type = "custom_approval"`
- All functions populate `post.Props` with approval data (timestamps as Unix millis)
- All functions set `post.Message` with markdown fallback

**AC2: Props Schema for DM Posts**
- Same props structure as playbook posts (Story 9.7 AC3)
- Additional props for DM-specific context:
```go
Props: map[string]interface{}{
    // Standard approval fields
    "approval_code": record.Code,
    "approval_status": record.Status,
    "requester_username": record.RequesterUsername,
    "requester_display_name": record.RequesterDisplayName,
    "approver_username": record.ApproverUsername,
    "approver_display_name": record.ApproverDisplayName,
    "description": record.Description,
    "created_at": record.CreatedAt,        // Unix millis
    "decided_at": record.DecidedAt,        // Unix millis
    "decision_comment": record.DecisionComment,

    // DM-specific fields
    "notification_type": "approval_request" | "outcome" | "cancellation" | "timeout" | "verification",
    "is_dm": true,                         // Flag to differentiate from playbook posts
}
```

**AC3: Component Adaptation for DM Context**
- ApprovalPost component detects `is_dm: true` prop
- Adjusts layout for 1:1 DM context (more verbose than playbook posts)
- Shows additional context appropriate for DM (full description, not truncated)
- Maintains all existing DM notification content (no information loss)

**AC4: Interactive Buttons in DM**
- Approval request DMs still show Approve/Deny buttons (existing behavior)
- Buttons remain functional with custom post type
- Decision modals still work (server-side handlers unchanged)
- Outcome notifications have no buttons (read-only)

**AC5: Markdown Fallback**
- All DM notifications have readable markdown fallback in `post.Message`
- Non-webapp clients see markdown (mobile, etc.)
- Fallback includes all essential information
- Uses existing formatter functions from notifications/dm.go

**AC6: Backward Compatibility**
- Old approval request DM buttons still work (post ID references valid)
- UpdateApprovalPostForCancellation() still works with custom post type
- No breaking changes to existing approvals in flight

**AC7: Unit Tests**
- Test each DM notification function with custom post type
- Test props population for each notification type
- Test markdown fallback generation
- Test button functionality with custom post type

**Technical Notes:**
- Reuse ApprovalPost component from Story 9.6 (no new component needed)
- Server-side changes only in notifications/dm.go
- May need helper function: `formatApprovalPropsForDM(record, notificationType)` → map[string]interface{}
- Ensure timestamps stored as int64, not formatted strings

---

### Story 9.11: End-to-End DM Notification Validation - ⛔ CANCELED

> **CANCELED:** Moved to Epic 10 along with Story 9.10. Will be validated as part of Epic 10 implementation.

**As a** user
**I want** all approval DM notifications to display with proper timezones
**So that** I can trust the timestamps regardless of where approval communications appear

**Acceptance Criteria:**

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

**Technical Notes:**
- This story validates ALL DM notifications end-to-end
- Test all 6 notification types (approval request, outcome, cancellation, timeout, verification, approver cancellation)
- Verify timezone accuracy with multiple users in different zones
- Ensure button functionality preserved with custom post type
- Test markdown fallback on non-webapp clients

---

## Dependencies

### External Dependencies
- Mattermost Server v6.0+ (Plugin API support)
- Mattermost webapp with React support
- Node.js v16+ and npm v8+ (for webapp build)
- User must have timezone configured in Mattermost settings

### Internal Dependencies
- Epic 8 (Playbook Integration) must be complete
- v2.x codebase stable and deployed
- `server/playbooks/client.go` and `formatters.go` in place

### Blocking Issues
- None identified

## Risks and Mitigations

**Risk 1: Webapp Build Complexity**
- **Likelihood:** Medium
- **Impact:** High (could delay entire epic)
- **Mitigation:** Use mattermost-plugin-starter-template as reference, start with minimal config
- **Mitigation:** Claude Code has React/TypeScript experience (mitigates Wayne's lack of experience)

**Risk 2: Timezone Calculation Bugs**
- **Likelihood:** Medium
- **Impact:** Medium (incorrect times displayed)
- **Mitigation:** Use Mattermost's built-in timezone selectors (battle-tested)
- **Mitigation:** Extensive testing with multiple timezones
- **Mitigation:** Unit tests for timestamp conversion

**Risk 3: Bundle Size**
- **Likelihood:** Low
- **Impact:** Low (larger plugin file)
- **Mitigation:** Use webpack code splitting if needed
- **Mitigation:** Mattermost already loads React (no additional framework cost)

**Risk 4: Breaking Changes**
- **Likelihood:** Low (Wayne confirmed no backward compatibility needed)
- **Impact:** Low (old posts show as markdown)
- **Mitigation:** Clear release notes about v3.0 changes
- **Mitigation:** Markdown fallback ensures no data loss

**Risk 5: Mobile Client Compatibility**
- **Likelihood:** Medium
- **Impact:** Medium (mobile users see markdown instead of components)
- **Mitigation:** Ensure markdown fallback is readable
- **Mitigation:** Test on mobile to verify fallback works
- **Future:** Add mobile-specific components if needed

## Testing Strategy

### Unit Tests
- Timestamp component timezone conversion
- Props mapping (server data → webapp props)
- Component rendering with different statuses
- Edge cases (null timestamps, missing data)

### Integration Tests
- Custom post type registration
- Post rendering in playbook channels and DMs
- Post updates on status changes
- Fallback to markdown for non-webapp clients
- Interactive buttons in DM posts

### Manual Testing
- Create approval in playbook channel AND non-playbook channel (DM flow)
- Verify custom component renders in both contexts
- Change user timezone, verify timestamps update everywhere
- Approve/deny, verify post updates in playbook AND DM
- Test all 6 DM notification types
- Test on web, desktop, mobile

### Performance Testing
- Measure component render time
- Check for memory leaks
- Verify smooth scrolling with many approval posts

## Success Criteria

### Must Have (Epic Complete) - ✅ ACHIEVED
✅ Webapp build pipeline functional
✅ Custom post type registered and working
✅ Timestamps display in user's local timezone
✅ Playbook channel posts use webapp components
✅ All approval statuses render correctly (playbook posts)
✅ No breaking changes to existing functionality
✅ Markdown fallback works for non-webapp clients

### Deferred to Epic 10
❌ DM notifications use webapp components (requires Matterpoll pattern)

### Nice to Have (Future Enhancement)
- Collapsible approval details
- Mobile-optimized layout
- Real-time updates without page refresh

### Definition of Done - ✅ ACHIEVED (Stories 9.1-9.9)
- ~~All 11 stories completed~~ → Stories 9.1-9.9 completed; Stories 9.10-9.11 canceled (moved to Epic 10)
- All unit tests pass
- Integration tests pass
- Manual testing complete with no critical bugs
- Documentation updated (README, architecture docs)
- Build and deployment process documented
- Epic 9 released as v2.2.0

## Documentation Requirements

**For Developers:**
- Update `docs/` with webapp setup instructions
- Document custom post type registration
- Document props schema for approval posts
- Add troubleshooting section for webapp build issues

**For Users:**
- Update README with timezone feature
- Add note about webapp requirement for rich posts
- Document markdown fallback behavior

**For Admins:**
- Update installation guide with webapp build steps
- Document Node.js/npm requirements

## Future Enhancements (Post-Epic)

**Epic 10: Advanced Webapp Features** (Optional)
- Collapsible approval sections
- Mobile-optimized layouts
- Real-time status updates
- Rich status badges matching Playbooks
- Estimated: 5 stories, 2 weeks

**Epic 11: Multiple Approvers UI** (Depends on backend epic)
- Display multiple approver statuses
- Progress indicators
- Vote tallies
- Estimated: Part of larger Multiple Approvers epic

## Effort Estimate

**Story Effort (Developer Days):**
- Story 9.1: 1.5 days (infrastructure setup)
- Story 9.2: 0.5 days (manifest updates)
- Story 9.3: 0.5 days (hello world verification)
- Story 9.4: 1 day (timestamp component)
- Story 9.5: 1 day (UI component library)
- Story 9.6: 1.5 days (base approval post component)
- Story 9.7: 1 day (post type registration)
- Story 9.8: 2 days (server-side updates for playbook posts)
- Story 9.9: 2 days (end-to-end testing for playbook posts)
- Story 9.10: 1 day (DM notification conversion)
- Story 9.11: 1 day (end-to-end testing for DM notifications)

**Total: ~13 developer days (~3 weeks)**

**Note:** Estimate assumes Claude Code assistance and no major blockers. Wayne's lack of React/TS experience offset by AI assistance.

## Appendix: Technical Reference

### Mattermost Plugin Webapp API
- `registry.registerPostTypeComponent(type, component)`
- `registry.registerRootComponent(component)`
- `getCurrentTimezone()` from mattermost-redux

### Post Object Structure
```typescript
interface Post {
    id: string;
    create_at: number;
    update_at: number;
    user_id: string;
    channel_id: string;
    message: string;           // Markdown fallback
    type: string;              // 'custom_approval' for our posts
    props: {[key: string]: any}; // Our approval data
    // ... other fields
}
```

### Approval Props Schema
```typescript
interface ApprovalPostProps {
    approval_code: string;
    approval_status: 'pending' | 'approved' | 'denied' | 'canceled' | 'timeout';
    requester_username: string;
    requester_display_name: string;
    approver_username: string;
    approver_display_name: string;
    description: string;
    created_at: number;        // Unix milliseconds
    decided_at?: number;       // Unix milliseconds
    decision_comment?: string;
    note?: string;             // For approved posts
}
```

### Webpack Config Reference
```javascript
module.exports = {
    entry: './src/index.tsx',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'main.js',
    },
    resolve: {
        extensions: ['.ts', '.tsx', '.js', '.jsx'],
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
        ],
    },
    externals: {
        react: 'React',
        'react-dom': 'ReactDOM',
        redux: 'Redux',
        'react-redux': 'ReactRedux',
    },
};
```

---

**Epic Owner:** Mary (Analyst) / Amelia (Dev)
**Reviewer:** Wayne
**Status:** Ready for Review
