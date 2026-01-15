# Epic 7: 1.0 Polish & UX Improvements

**Version:** 1.0
**Status:** Planned
**Priority:** High (Release Blocker)
**Created:** 2026-01-14

## Overview

Final polish pass for 1.0 release addressing critical feedback gaps, UX friction points, discoverability issues, and build stability. These improvements ensure a professional, confidence-inspiring user experience for the initial release.

## Problem Statement

**Current Issues:**
1. No feedback notification when approver cancels a request - requestor left in the dark
2. CI pipeline triggers multiple builds on tag push, two of which fail consistently
3. Additional details box in cancel dialog accepts input but ignores it unless "other" is selected
4. Slash command autocomplete only shows `/approve` without argument suggestions
5. `/approve list` output uses plain text formatting instead of cleaner markdown tables

**User Impact:**
- Broken feedback loops undermine trust in the system (Issue #1)
- Release process is unreliable and error-prone (Issue #2)
- Confusing UX violates user expectations (Issue #3)
- Poor discoverability forces users to memorize syntax or consult docs (Issue #4)
- List output looks unpolished compared to other Mattermost plugins (Issue #5)

## Goals

### Primary Goals
1. **Complete Feedback Loops:** Ensure all actions provide clear confirmation to all stakeholders
2. **Reliable Releases:** Fix CI/CD pipeline to enable confident deployments
3. **Clear UX Patterns:** Eliminate confusion in user interactions
4. **Self-Documenting CLI:** Enable command discovery through autocomplete
5. **Professional Polish:** Match quality bar of established Mattermost plugins

### Success Metrics
- Requestors receive notification when their request is canceled
- Tag push triggers clean build with zero failures
- Additional details behavior is obvious and consistent
- All slash command arguments autocomplete like Jira plugin
- `/approve list` output renders as clean markdown table

## User Stories

### Story 7.1: Fix Cancellation Notification to Requestor
**Priority:** CRITICAL (Broken feedback loop)

**As a** requestor
**I want** to receive notification when my approval request is canceled
**So that** I know the status changed and can take appropriate action

**Acceptance Criteria:**
- When approver cancels a request, requestor receives ephemeral or DM notification
- Notification includes: request ID, who canceled it, cancellation reason, timestamp
- Notification updates/replaces original request message if possible
- Requestor doesn't need to manually run `/request get` or `/request list canceled` to discover cancellation
- Notification tone is clear and professional
- Works for all cancellation reasons (approved/denied elsewhere, mistake, expired, duplicate, other)

**Technical Notes:**
- Determine notification mechanism (update ephemeral message vs. new DM)
- Consider Mattermost API limitations for updating user's ephemeral messages
- Ensure notification includes context for requestor to understand what happened
- Add logging for notification delivery success/failure

**Testing:**
- Cancel request with each reason type, verify requestor receives notification
- Test with requestor offline/online scenarios
- Verify notification content is complete and clear

---

### Story 7.2: Fix CI/CD Build Failures on Tag Push
**Priority:** CRITICAL (Release blocker)

**As a** maintainer
**I want** tag pushes to trigger clean builds without failures
**So that** I can release with confidence

**Acceptance Criteria:**
- Pushing a tag to master triggers only necessary builds
- All triggered builds complete successfully (zero failures)
- Redundant or duplicate build triggers eliminated
- Build logs are clean and actionable
- Release workflow is documented and repeatable

**Technical Notes:**
- Investigate why three builds trigger (examine CI config, webhooks, GitHub Actions)
- Identify root cause of two failing builds
- Determine if failures are flaky tests, environment issues, or configuration problems
- Consider if tag push should trigger fewer builds (e.g., only release build)
- Update CI configuration to deduplicate builds

**Testing:**
- Push test tag to verify single successful build
- Verify release artifacts are correctly generated
- Test on both initial tag and tag update scenarios

---

### Story 7.3: Fix Additional Details Box Behavior in Cancel Dialog
**Priority:** HIGH (Confusing UX)

**As an** approver
**I want** the additional details box to behave consistently
**So that** I understand when my input will be captured

**Acceptance Criteria:**
- **Option A (Conditional Display):** Additional details text box only appears when "Other" reason is selected
- **Option B (Always Capture):** Additional details text box is always visible and captured regardless of reason selected
- Chosen behavior is visually obvious to users
- No silent data loss (text entered is never ignored)
- Help text or placeholder clarifies purpose
- Captured details appear in audit trail and notifications

**Technical Notes:**
- Decide on Option A vs Option B based on UX philosophy
- Option A: Show/hide text box dynamically based on reason selection
- Option B: Always include details field in data model, append to reason
- Update cancel dialog component accordingly
- Ensure details are stored in request history
- Update cancellation notifications to include details if present

**Testing:**
- Select each cancellation reason and verify details box behavior matches expectations
- Enter text in details box with "Other" and non-"Other" reasons
- Verify captured details appear in audit trail and notifications
- Test with empty details (should be allowed)

---

### Story 7.4: Add Autocomplete for Slash Command Arguments
**Priority:** HIGH (Discoverability)

**As a** user
**I want** slash command arguments to autocomplete
**So that** I can discover available commands without consulting documentation

**Acceptance Criteria:**
- Typing `/approve` shows autocomplete suggestions for all subcommands (list, approve, deny, cancel, stats, help, etc.)
- Typing `/approve list` shows autocomplete for filter arguments (pending, approved, denied, canceled, all)
- Typing `/request` shows autocomplete for all subcommands (create, get, list, cancel, etc.)
- Autocomplete behavior matches Jira plugin pattern (help display with descriptions)
- All arguments have clear, concise descriptions in autocomplete
- Autocomplete is discoverable and intuitive

**Technical Notes:**
- Implement Mattermost autocomplete API (manifest.json commands structure)
- Define autocomplete suggestions for each command level
- Include parameter hints and descriptions
- Follow Mattermost plugin autocomplete best practices
- Reference Jira plugin implementation as model
- Test autocomplete performance with multiple plugins installed

**Testing:**
- Type each slash command and verify autocomplete appears
- Navigate autocomplete with keyboard
- Select autocomplete option and verify correct command is inserted
- Test with partial command typing (e.g., `/appr` should suggest `/approve`)

---

### Story 7.5: Convert /approve list Output to Markdown Table
**Priority:** MEDIUM (Polish)

**As a** user
**I want** `/approve list` output formatted as a markdown table
**So that** the output looks clean and professional

**Acceptance Criteria:**
- `/approve list` renders output as markdown table (not plain text)
- Table columns: Request ID, Requestor, Type, Status, Created, (optional: Link)
- Table formatting matches Mattermost markdown rendering standards
- Table is responsive/readable on desktop and mobile
- All existing filtering and sorting behavior preserved
- Header shows count as before: "## Your Approval Requests (N pending/approved/etc.)"
- Empty states remain clear and helpful

**Technical Notes:**
- Replace plain text output with markdown table syntax
- Ensure table column widths are reasonable
- Consider truncating long values (e.g., request details)
- Test table rendering in Mattermost web, desktop, and mobile apps
- Maintain Epic 5 filtering logic (pending/approved/denied/canceled/all)
- Keep sort order from Epic 4.6 (canceled at bottom for `all` view)

**Testing:**
- Run `/approve list` with each filter type (pending, approved, denied, canceled, all)
- Verify table formatting renders correctly
- Test with 0, 1, many requests
- Test with long requestor names, long request types
- View on web, desktop, mobile to verify responsiveness

---

## Technical Considerations

### Mattermost API Integration
- Story 7.1 may require investigating ephemeral message update capabilities
- Story 7.4 requires manifest.json command definitions with autocomplete
- Story 7.5 requires markdown table rendering support (should be native)

### CI/CD Pipeline
- Story 7.2 requires access to CI configuration (GitHub Actions, CircleCI, etc.)
- May need to audit webhook configurations
- Consider impact on development workflow vs. release workflow

### UX Consistency
- All changes should match existing plugin UX patterns
- Follow Mattermost plugin best practices
- Reference established plugins (Jira, GitHub) for patterns

### Breaking Changes
- None expected - all changes are additive or fixes

### Testing Strategy
- Unit tests for formatting changes (Story 7.5)
- Integration tests for notification delivery (Story 7.1)
- Manual testing for autocomplete (Story 7.4)
- Manual testing for CI pipeline (Story 7.2)
- Manual testing for dialog behavior (Story 7.3)

## Dependencies

- **Builds on:** All previous epics (Epic 1-6)
- **Blocks:** 1.0 Release
- **Blocked by:** None

## Out of Scope

- New features or functionality
- Performance optimization
- Extensive refactoring
- Migration or upgrade logic
- Additional cancellation reasons
- Advanced autocomplete (context-aware suggestions)

## Implementation Order

**Sequential by Priority:**
1. Story 7.2: Fix CI/CD (can't release without this)
2. Story 7.1: Fix cancellation notification (critical UX gap)
3. Story 7.3: Fix additional details behavior (high confusion factor)
4. Story 7.4: Add autocomplete (discoverability)
5. Story 7.5: Convert to markdown table (polish, fast)

Alternatively, 7.1, 7.3, 7.4, 7.5 can run in parallel after 7.2 is fixed.

## Success Validation

- [ ] Requestor receives notification when request is canceled (Story 7.1)
- [ ] Tag push triggers single successful build (Story 7.2)
- [ ] Additional details box behavior is clear and consistent (Story 7.3)
- [ ] All slash commands autocomplete arguments (Story 7.4)
- [ ] `/approve list` renders as clean markdown table (Story 7.5)
- [ ] All existing functionality preserved
- [ ] Manual testing on web, desktop, mobile clients
- [ ] Ready to ship 1.0 with confidence

## Notes

- This is the final epic before 1.0 release
- All items are polish and UX improvements, not new features
- Stories 7.1 and 7.2 are true blockers - cannot ship without them
- Stories 7.3, 7.4, 7.5 are polish but should be included for professional quality
- Wayne considers this feature-complete after Epic 6, now focusing on quality
- Story 7.5 is estimated to be quick/easy (format conversion only)
