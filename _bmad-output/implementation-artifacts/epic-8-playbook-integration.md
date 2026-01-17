# Epic 8: Playbook Integration

**Version:** 2.0.0
**Status:** Planned
**Priority:** High
**Created:** 2026-01-17

## Overview

Integrate the Approval Workflow Plugin with Mattermost Playbooks to provide contextual approval workflows within operational processes. When users create approval requests in playbook channels, the plugin automatically detects the playbook context, links the approval to the run, and provides status updates to the playbook team - keeping everyone informed without switching context.

## Problem Statement

**Current Issues:**
1. Approvals created in playbook channels send DMs but don't update the playbook team
2. No visibility into approval status for playbook participants
3. Approvers lack context about which playbook generated the approval request
4. Approval bottlenecks aren't visible in the playbook channel
5. No connection between approval workflow and playbook workflow

**User Impact:**
- Playbook team members unaware of approval blockers
- Context switching required to check approval status
- Approvers don't understand urgency or playbook context
- Poor user experience for incident response, deploy workflows, change management
- Manual coordination needed to track approval dependencies

## Goals

### Primary Goals
1. **Automatic Playbook Detection:** Plugin detects when command is run in playbook channel
2. **Channel Visibility:** Approval status posted in playbook channel for team awareness
3. **Contextual DMs:** Approver notifications include playbook name and context
4. **Zero Configuration:** Works automatically without user parameters or setup
5. **Graceful Fallback:** Non-playbook channels work exactly as v1.0

### Success Metrics
- Approval requests auto-link to playbook runs when created in playbook channels
- Playbook channel receives status updates for all approval state changes
- Approver DMs include playbook context
- v1.0 behavior preserved for non-playbook channels
- Zero breaking changes to existing approval workflow

## User Stories

### Story 8.1: Playbook Context Detection
**As a** plugin developer
**I want** to detect when an approval is created in a playbook channel
**So that** I can automatically link the approval to the playbook run

**Acceptance Criteria:**
- Plugin calls Playbooks API to check if channel is a playbook channel
- If playbook exists, retrieves run ID, name, and metadata
- If no playbook (404), proceeds with normal v1.0 flow
- Detection happens transparently without user input
- API call uses bot token authentication
- Error handling prevents approval creation from failing

**Technical Notes:**
- Use `GET /plugins/playbooks/api/v0/runs/channel/{channel_id}` endpoint
- Call synchronously during approval creation
- Store result for subsequent operations
- Log errors but don't block approval creation

---

### Story 8.2: Data Model Extension for Playbook Metadata
**As a** plugin developer
**I want** to store playbook metadata with approval records
**So that** I can reference the playbook throughout the approval lifecycle

**Acceptance Criteria:**
- Approval struct extended with playbook fields (run ID, name, channel ID, post ID)
- Fields stored in KV store with approval record
- Fields retrievable via `/approve get` command
- Backward compatible with v1.0 records (nil playbook fields)
- No data migration required

**Technical Notes:**
- Add optional fields to Approval struct
- Update KV serialization to handle new fields
- Ensure omitempty JSON tags for backward compatibility
- Display playbook context in approval detail view

---

### Story 8.3: Post Status Messages to Playbook Channel
**As a** playbook team member
**I want** to see approval status in the playbook channel
**So that** I know when approvals are blocking progress

**Acceptance Criteria:**
- When approval created in playbook channel, post status message
- Message format: "⏳ Approval pending: [CODE] - [Details] | Waiting for @approver"
- Message posted using Playbooks status API
- Post ID stored in approval record for later updates
- Error posting doesn't break approval creation
- Message styled appropriately for channel visibility

**Technical Notes:**
- Use `POST /plugins/playbooks/api/v0/runs/{id}/status` endpoint
- Format message with markdown for readability
- Include approval reference code for correlation
- Handle API errors gracefully with logging

---

### Story 8.4: Add Playbook Context to Approver DM Notifications
**As an** approver
**I want** to see which playbook the approval request came from
**So that** I understand the urgency and context

**Acceptance Criteria:**
- Approver DM includes playbook name when approval is playbook-linked
- Format: "From playbook: [Playbook Name]"
- Link back to playbook channel included
- Works for all notification types (new request, reminders)
- Non-playbook approvals unchanged (v1.0 behavior)

**Technical Notes:**
- Modify DM notification templates
- Check if approval has playbook metadata
- Conditionally add playbook context section
- Include channel link for easy navigation

---

### Story 8.5: Update Playbook Channel on Status Changes
**As a** playbook team member
**I want** to see when approvals are approved, denied, canceled, or timed out
**So that** I know when blockers are resolved

**Acceptance Criteria:**
- Approved: Post "✅ Approved: [CODE] - [Details] | Approved by @approver at [time]"
- Denied: Post "❌ Denied: [CODE] - [Details] | Denied by @approver at [time]"
- Canceled: Post "🚫 Canceled: [CODE] - [Reason]"
- Timed out: Post "⏱️ Timeout: [CODE] - No response from @approver"
- All status posts include reference code for correlation
- Messages styled consistently with initial status post

**Technical Notes:**
- Hook into existing status change handlers
- Check for playbook metadata before posting
- Use same status API endpoint
- Format messages for clarity and scannability
- Log errors but don't block status updates

---

### Story 8.6: Error Handling and Graceful Fallback
**As a** user
**I want** approvals to work even when Playbooks is unavailable
**So that** my approval workflow is reliable

**Acceptance Criteria:**
- If Playbooks plugin disabled, approvals work as v1.0 (no playbook integration)
- If API call fails, approval creation continues
- If bot lacks channel permissions, approval continues with logged warning
- If playbook deleted mid-approval, status updates skip gracefully
- All errors logged for debugging
- No user-visible errors when playbook integration fails

**Technical Notes:**
- Wrap all Playbooks API calls in error handlers
- Use circuit breaker pattern to avoid repeated failures
- Log errors with context for troubleshooting
- Ensure approval core workflow never blocked by playbook failures
- Test with Playbooks plugin disabled
- Test with invalid permissions

---

## Technical Considerations

### API Integration
- **Playbooks API Base:** `/plugins/playbooks/api/v0/`
- **Authentication:** Bot token via plugin API
- **Key Endpoints:**
  - `GET /runs/channel/{channel_id}` - Detect playbook context
  - `POST /runs/{id}/status` - Post status updates
  - `GET /runs/{id}` - Retrieve run details
  - `GET /runs/{id}/metadata` - Get channel/team info

### Data Model Changes
```go
type Approval struct {
    // ... existing v1.0 fields ...

    // v2.0 Playbook Integration (optional fields)
    PlaybookRunID     string `json:"playbook_run_id,omitempty"`
    PlaybookName      string `json:"playbook_name,omitempty"`
    PlaybookChannelID string `json:"playbook_channel_id,omitempty"`
    PlaybookPostID    string `json:"playbook_post_id,omitempty"`
}
```

### Performance Impact
- One additional API call per approval creation (synchronous)
- Minimal latency (< 100ms expected)
- Status updates asynchronous (don't block approval workflow)
- No database schema changes

### Error Handling Strategy
**Principle:** Playbook integration is **additive, not required**
- All Playbooks API calls wrapped in try-catch equivalent
- Errors logged but never block core approval workflow
- Silent degradation to v1.0 behavior on failure
- Circuit breaker prevents repeated failing API calls

### Testing Strategy
- Unit tests for playbook detection logic
- Unit tests for status message formatting
- Integration tests with mock Playbooks API
- Manual testing with real Playbooks plugin
- Error injection testing (API failures, permissions, etc.)
- Backward compatibility testing (v1.0 records)

## Dependencies

- **Builds on:** All Epic 1-7 functionality
- **Requires:** Mattermost Playbooks plugin installed (optional, graceful degradation)
- **Blocks:** None
- **Blocked by:** None

## Out of Scope for v2.0

- Checklist item integration (update tasks automatically)
- Multi-approver workflows in playbooks
- Approval templates tied to playbook types
- Playbook webhook triggers creating approvals
- Analytics on approval delays by playbook
- Playbook-specific configuration options

These features deferred to future versions (2.1+) after validating core integration.

## Implementation Order

1. **Story 8.1:** Playbook context detection (foundation)
2. **Story 8.2:** Data model extension (enables storage)
3. **Story 8.3:** Channel status posts (first visible integration)
4. **Story 8.4:** DM context enhancement (improves approver experience)
5. **Story 8.5:** Status change integration (complete lifecycle)
6. **Story 8.6:** Error handling (production hardening)

Sequential implementation ensures each piece builds on previous work.

## Success Validation

**Functional Tests:**
- [ ] Approval in playbook channel auto-detects playbook context
- [ ] Playbook channel receives initial pending status post
- [ ] Approver DM includes playbook name and link
- [ ] Approval decision updates playbook channel
- [ ] Cancellation updates playbook channel
- [ ] Timeout updates playbook channel
- [ ] Non-playbook channels work exactly as v1.0
- [ ] Error scenarios degrade gracefully

**User Experience Tests:**
- [ ] Playbook team members can see approval blockers without DMs
- [ ] Approvers understand playbook context from notification
- [ ] No additional commands or parameters required
- [ ] Status messages are clear and actionable
- [ ] Reference codes allow correlation between channel posts and DMs

**Production Readiness:**
- [ ] All tests passing (unit, integration, manual)
- [ ] Error handling tested with injected failures
- [ ] Performance acceptable (< 100ms overhead)
- [ ] Documentation updated (README, CHANGELOG)
- [ ] Backward compatibility with v1.0 records verified

## Release Strategy

**Version:** 2.0.0 (minor version bump for new feature)
**Breaking Changes:** None
**Migration Required:** None

**Upgrade Path:**
- Users on v1.0 can upgrade directly to v2.0
- Existing approval records remain accessible
- Playbook integration activates automatically for new approvals
- No configuration changes required

**Rollback Plan:**
- If issues arise, users can downgrade to v1.0
- v1.0 records compatible with both versions
- No data loss on rollback

## Notes

- This epic originated from user research conversation on 2026-01-17
- Analyst (Mary) validated Playbooks API capabilities via documentation review
- Key insight: `GET /runs/channel/{channel_id}` enables auto-detection without user input
- Design prioritizes zero-configuration user experience
- Strong emphasis on graceful degradation and error resilience
- Feature is additive - v1.0 functionality completely preserved
