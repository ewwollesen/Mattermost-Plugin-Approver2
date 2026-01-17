# Story 8.2: Data Model Extension for Playbook Metadata

**Epic:** 8 - Playbook Integration
**Status:** ready-for-dev
**Priority:** High
**Estimate:** 3 points
**Assignee:** TBD

## User Story

**As a** plugin developer
**I want** to store playbook metadata with approval records
**So that** I can reference the playbook throughout the approval lifecycle

## Context

After detecting playbook context (Story 8.1), we need to persist this information with the approval record. This enables subsequent stories to post status updates to the playbook channel and add context to notifications.

The data model extension must be backward compatible with v1.0 approval records that don't have playbook fields.

## Acceptance Criteria

- [ ] AC1: Approval struct extended with optional playbook fields
- [ ] AC2: Fields stored in KV store when approval is created
- [ ] AC3: Fields retrieved correctly when reading approval from KV store
- [ ] AC4: v1.0 approval records (without playbook fields) load without errors
- [ ] AC5: `/approve get [CODE]` displays playbook context when present
- [ ] AC6: `/approve list` can display playbook name (optional enhancement)
- [ ] AC7: No data migration required for existing records
- [ ] AC8: JSON serialization uses omitempty tags for nil values

## Tasks / Subtasks

- [ ] Task 1: Extend Approval data structure (AC: 1, 8)
  - [ ] Subtask 1.1: Add PlaybookRunID string field with omitempty tag
  - [ ] Subtask 1.2: Add PlaybookName string field with omitempty tag
  - [ ] Subtask 1.3: Add PlaybookChannelID string field with omitempty tag
  - [ ] Subtask 1.4: Add PlaybookPostID string field with omitempty tag
  - [ ] Subtask 1.5: Update struct documentation with field descriptions

- [ ] Task 2: Update KV storage and retrieval (AC: 2, 3, 4, 7)
  - [ ] Subtask 2.1: Verify KV serialization handles new fields correctly
  - [ ] Subtask 2.2: Test loading v1.0 records (missing playbook fields)
  - [ ] Subtask 2.3: Test loading v2.0 records (with playbook fields)
  - [ ] Subtask 2.4: Ensure no migration script needed
  - [ ] Subtask 2.5: Write unit tests for backward compatibility

- [ ] Task 3: Update approval creation flow (AC: 2)
  - [ ] Subtask 3.1: Populate playbook fields when detection succeeds (Story 8.1)
  - [ ] Subtask 3.2: Leave playbook fields nil when no playbook detected
  - [ ] Subtask 3.3: Store approval with playbook metadata in KV store
  - [ ] Subtask 3.4: Write unit tests for both scenarios

- [ ] Task 4: Update display functions (AC: 5, 6)
  - [ ] Subtask 4.1: Modify `/approve get` formatting to show playbook context
  - [ ] Subtask 4.2: Add conditional section: "Playbook Context: [name] (link)"
  - [ ] Subtask 4.3: Optionally add playbook column to `/approve list` tables
  - [ ] Subtask 4.4: Update format tests to handle playbook fields
  - [ ] Subtask 4.5: Write integration tests for display functions

## Dev Notes

### Data Structure Extension

```go
// In server/store/approval.go
type Approval struct {
    // Existing v1.0 fields
    ID              string                 `json:"id"`
    ReferenceCode   string                 `json:"reference_code"`
    RequesterUserID string                 `json:"requester_user_id"`
    ApproverUserID  string                 `json:"approver_user_id"`
    RequestDetails  string                 `json:"request_details"`
    Status          ApprovalStatus         `json:"status"`
    Decision        string                 `json:"decision,omitempty"`
    DecisionReason  string                 `json:"decision_reason,omitempty"`
    CancelReason    string                 `json:"cancel_reason,omitempty"`
    CreatedAt       int64                  `json:"created_at"`
    DecidedAt       int64                  `json:"decided_at,omitempty"`
    CanceledAt      int64                  `json:"canceled_at,omitempty"`
    VerifiedAt      int64                  `json:"verified_at,omitempty"`
    ApproverPostID  string                 `json:"approver_post_id,omitempty"`

    // v2.0 Playbook Integration fields (optional)
    PlaybookRunID     string `json:"playbook_run_id,omitempty"`
    PlaybookName      string `json:"playbook_name,omitempty"`
    PlaybookChannelID string `json:"playbook_channel_id,omitempty"`
    PlaybookPostID    string `json:"playbook_post_id,omitempty"`
}
```

### Populating Playbook Fields

```go
// In server/command/router.go - after modal submission
func (r *CommandRouter) handleApprovalSubmission(submission *model.SubmitDialogResponse, playbookRun *PlaybookRun) error {
    approval := &store.Approval{
        // ... existing v1.0 fields ...
    }

    // Populate playbook fields if context detected
    if playbookRun != nil {
        approval.PlaybookRunID = playbookRun.ID
        approval.PlaybookName = playbookRun.Name
        approval.PlaybookChannelID = playbookRun.ChannelID
        // PlaybookPostID will be set after posting (Story 8.3)
    }

    return r.store.CreateApproval(approval)
}
```

### Display Updates

```go
// In server/command/router.go - formatApprovalDetails
func (r *CommandRouter) formatApprovalDetails(approval *store.Approval) string {
    var output strings.Builder

    // ... existing v1.0 formatting ...

    // Add playbook context section if present
    if approval.PlaybookRunID != "" {
        output.WriteString("\n\n**Playbook Context:**\n")
        output.WriteString(fmt.Sprintf("- Name: %s\n", approval.PlaybookName))
        if approval.PlaybookChannelID != "" {
            channelLink := fmt.Sprintf("[View Playbook Channel](/_redirect/pl/%s)", approval.PlaybookChannelID)
            output.WriteString(fmt.Sprintf("- Channel: %s\n", channelLink))
        }
    }

    return output.String()
}
```

### Backward Compatibility Testing

```go
func TestLoadV1ApprovalRecord(t *testing.T) {
    // Create v1.0 JSON without playbook fields
    v1JSON := `{
        "id": "test123",
        "reference_code": "TUZ-2RK",
        "status": "pending"
    }`

    var approval store.Approval
    err := json.Unmarshal([]byte(v1JSON), &approval)

    require.NoError(t, err)
    assert.Empty(t, approval.PlaybookRunID) // Should be nil
    assert.Empty(t, approval.PlaybookName)
}

func TestLoadV2ApprovalRecord(t *testing.T) {
    // Create v2.0 JSON with playbook fields
    v2JSON := `{
        "id": "test123",
        "reference_code": "TUZ-2RK",
        "status": "pending",
        "playbook_run_id": "playbook123",
        "playbook_name": "Incident #47"
    }`

    var approval store.Approval
    err := json.Unmarshal([]byte(v2JSON), &approval)

    require.NoError(t, err)
    assert.Equal(t, "playbook123", approval.PlaybookRunID)
    assert.Equal(t, "Incident #47", approval.PlaybookName)
}
```

### Files to Create/Modify

**Modified Files:**
- `server/store/approval.go` - Add playbook fields to Approval struct
- `server/command/router.go` - Populate and display playbook fields
- `server/store/approval_test.go` - Add backward compatibility tests
- `server/command/router_test.go` - Update display format tests

## Definition of Done

- [ ] All acceptance criteria met and tested
- [ ] All tasks and subtasks completed
- [ ] Approval struct extended with playbook fields
- [ ] Backward compatibility with v1.0 records verified
- [ ] `/approve get` displays playbook context
- [ ] Unit tests passing (100% coverage for changes)
- [ ] Integration tests passing
- [ ] Code review completed
- [ ] Ready for Story 8.3 (channel status posts)

## Related Stories

- **Depends on:** Story 8.1 (provides playbook data to store)
- **Blocks:** Story 8.3 (needs fields to reference playbook)
- **Blocks:** Story 8.4 (needs playbook name for DM context)
- **Blocks:** Story 8.5 (needs run ID for status updates)

## Technical Debt / Future Improvements

- Consider adding indexed search by playbook run ID
- Add analytics: count approvals per playbook
- Add admin endpoint to query all approvals for a playbook
