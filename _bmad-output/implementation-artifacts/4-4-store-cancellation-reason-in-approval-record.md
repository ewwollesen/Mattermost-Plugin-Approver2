# Story 4.4: Store Cancellation Reason in Approval Record

**Epic:** 4 - Improved Cancellation UX + Audit Trail
**Status:** To Do
**Priority:** High
**Estimate:** 2 points
**Assignee:** TBD

## User Story

**As a** system
**I want** to persist the cancellation reason with the approval record
**So that** it's available for audit and reporting

## Context

Cancellation reasons captured in Story 4.3 need to be stored permanently with the approval record. This data enables:
1. Audit trail - "Why was this cancelled?"
2. Reporting - "What % of requests are cancelled for each reason?"
3. Process improvement - "Are we seeing patterns that indicate UX problems?"

The data must be immutable once set (no editing after cancellation).

## Acceptance Criteria

- [ ] Add `CanceledReason` field to `ApprovalRequest` struct
- [ ] Add `CanceledAt` timestamp field to `ApprovalRequest` struct
- [ ] Update KV storage to persist both fields
- [ ] Fields remain immutable once set (no editing post-cancel)
- [ ] Historical records without these fields handle gracefully (backwards compatible)
- [ ] JSON marshaling/unmarshaling works correctly with new fields

## Technical Implementation

### Data Model Update

```go
// In server/approval_request.go

type ApprovalRequest struct {
    ID                 string `json:"id"`
    ReferenceCode      string `json:"reference_code"`
    RequesterID        string `json:"requester_id"`
    RequesterUsername  string `json:"requester_username"`
    ApproverID         string `json:"approver_id"`
    ApproverUsername   string `json:"approver_username"`
    Description        string `json:"description"`
    Status             string `json:"status"` // "pending", "approved", "denied", "canceled"
    CreatedAt          int64  `json:"created_at"`
    UpdatedAt          int64  `json:"updated_at"`
    ApprovedAt         int64  `json:"approved_at,omitempty"`
    DeniedAt           int64  `json:"denied_at,omitempty"`

    // New fields for cancellation
    CanceledReason     string `json:"canceled_reason,omitempty"`
    CanceledAt         int64  `json:"canceled_at,omitempty"`

    ApproverPostID     string `json:"approver_post_id,omitempty"`
}
```

### Update Cancel Method

```go
// In server/approval_request.go

func (p *Plugin) cancelApprovalRequest(request *ApprovalRequest, reason string, canceledBy string) error {
    // Validate state
    if request.Status != StatusPending {
        return fmt.Errorf("cannot cancel request with status: %s", request.Status)
    }

    // Update request fields (immutable once set)
    request.Status = StatusCanceled
    request.CanceledReason = reason
    request.CanceledAt = time.Now().UnixMilli()
    request.UpdatedAt = time.Now().UnixMilli()

    // Persist to KV store
    err := p.storeApprovalRequest(request)
    if err != nil {
        p.API.LogError("Failed to store cancelled request", "request_id", request.ID, "error", err.Error())
        return err
    }

    // Update indexes (if list sorting changes)
    err = p.updateIndexesForCancellation(request)
    if err != nil {
        p.API.LogWarn("Failed to update indexes", "request_id", request.ID, "error", err.Error())
        // Don't fail - request is already cancelled
    }

    p.API.LogInfo("Request cancelled",
        "request_id", request.ID,
        "reason", reason,
        "canceled_by", canceledBy)

    return nil
}
```

### Backwards Compatibility

```go
// When loading old records
func (p *Plugin) getApprovalRequest(requestID string) (*ApprovalRequest, error) {
    data, err := p.API.KVGet(fmt.Sprintf("approval_request:v1:%s", requestID))
    if err != nil {
        return nil, err
    }
    if data == nil {
        return nil, fmt.Errorf("request not found: %s", requestID)
    }

    var request ApprovalRequest
    err = json.Unmarshal(data, &request)
    if err != nil {
        return nil, err
    }

    // Old records won't have these fields - they'll be empty strings/zero values
    // This is fine - empty CanceledReason means "no reason captured"

    return &request, nil
}
```

### Validation

```go
// Prevent modification after cancellation
func (p *Plugin) validateCancellationImmutability(request *ApprovalRequest) error {
    if request.Status == StatusCanceled {
        if request.CanceledReason == "" {
            return fmt.Errorf("canceled requests must have a reason")
        }
        if request.CanceledAt == 0 {
            return fmt.Errorf("canceled requests must have a timestamp")
        }
    }
    return nil
}
```

## Testing Requirements

### Unit Tests
- [ ] Test cancellation updates fields correctly
- [ ] Test cancelled request cannot be re-cancelled
- [ ] Test reason and timestamp are stored
- [ ] Test loading old records without new fields (backwards compatibility)
- [ ] Test JSON marshaling with new fields
- [ ] Test immutability validation

### Integration Tests
- [ ] Create request → Cancel with reason → Load request → Verify fields
- [ ] Test that cancelled requests persist after plugin restart
- [ ] Test cancellation reason appears in subsequent operations
- [ ] Test old records without cancellation fields load correctly

### Data Migration Tests
- [ ] Create approval using v0.1.0 code (no cancel fields)
- [ ] Upgrade to v0.2.0 and load that approval
- [ ] Verify it displays correctly with empty cancellation reason

## Dependencies

- **Depends on:** Story 4.3 (Capture reason) - provides reason data
- **Blocks:** Story 4.1, 4.2, 4.5, 4.6 - all need stored reason data
- **No other dependencies**

## Database Schema

**KV Store Key Pattern:**
```
approval_request:v1:{requestID}
```

**Example Stored JSON:**
```json
{
  "id": "abc123",
  "reference_code": "TUZ-2RK",
  "requester_id": "user123",
  "requester_username": "wayne",
  "approver_id": "user456",
  "approver_username": "john",
  "description": "Deploy hotfix to production",
  "status": "canceled",
  "created_at": 1705089600000,
  "updated_at": 1705093200000,
  "canceled_reason": "No longer needed",
  "canceled_at": 1705093200000,
  "approver_post_id": "post789"
}
```

## Migration Strategy

**No explicit migration needed** - fields use `omitempty` JSON tag:
- Old records: Fields will be absent in JSON, loaded as empty string/zero
- New records: Fields will be present and populated
- Display logic handles both cases gracefully

**If cancellation reason is empty:**
- Display: "No reason provided" or "Cancelled before v0.2.0"
- Queries: Can filter for `CanceledReason != ""`  to get only v0.2.0+ cancellations

## Definition of Done

- [ ] Data model updated with new fields
- [ ] Cancel method stores reason and timestamp
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] Backwards compatibility verified
- [ ] JSON serialization tested
- [ ] Code merged to master

## Notes

**Why immutable:**
- Audit trail must be tamper-proof
- If reason changes, creates confusion ("did they lie initially?")
- If recancellation needed, create a new approval request

**Why milliseconds for CanceledAt:**
- Consistent with existing timestamp fields (CreatedAt, etc.)
- Higher precision for sorting/ordering if needed
- JavaScript-friendly (Date constructor accepts millis)

**Why omitempty:**
- Keeps old records compact (don't add fields to historical data)
- Clear signal: "this field didn't exist when this record was created"
- Backwards compatible with v0.1.0 records
