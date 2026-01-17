# Story 8.6: Error Handling and Graceful Fallback

**Epic:** 8 - Playbook Integration
**Status:** ready-for-dev
**Priority:** High
**Estimate:** 5 points
**Assignee:** TBD

## User Story

**As a** user
**I want** approvals to work even when Playbooks is unavailable
**So that** my approval workflow is reliable

## Context

Playbook integration is an enhancement to the core approval workflow, not a dependency. If the Playbooks plugin is disabled, the API is unreachable, or permissions are insufficient, the approval workflow must continue functioning exactly as v1.0.

This story implements comprehensive error handling, circuit breakers, and logging to ensure production reliability. It's the final hardening step before Epic 8 is complete.

## Acceptance Criteria

- [ ] AC1: If Playbooks plugin disabled, approvals work as v1.0 (no playbook features)
- [ ] AC2: If API call fails (network, timeout, 5xx), approval creation continues
- [ ] AC3: If bot lacks channel permissions, approval continues with logged warning
- [ ] AC4: If playbook deleted mid-approval, status updates skip gracefully
- [ ] AC5: All errors logged with context for troubleshooting
- [ ] AC6: No user-visible errors when playbook integration fails
- [ ] AC7: Circuit breaker prevents repeated failed API calls
- [ ] AC8: Health check endpoint reports playbook integration status
- [ ] AC9: Metrics tracked for playbook integration success/failure rates

## Tasks / Subtasks

- [ ] Task 1: Implement comprehensive error handling (AC: 2, 3, 4, 5, 6)
  - [ ] Subtask 1.1: Wrap all Playbooks API calls in error handlers
  - [ ] Subtask 1.2: Implement proper error logging with context
  - [ ] Subtask 1.3: Ensure errors never propagate to user
  - [ ] Subtask 1.4: Add specific handling for 403 (permissions)
  - [ ] Subtask 1.5: Add specific handling for 404 (playbook deleted)
  - [ ] Subtask 1.6: Add specific handling for timeouts
  - [ ] Subtask 1.7: Write unit tests for all error scenarios

- [ ] Task 2: Implement circuit breaker pattern (AC: 7)
  - [ ] Subtask 2.1: Create CircuitBreaker struct with state management
  - [ ] Subtask 2.2: Track consecutive failures (threshold: 5)
  - [ ] Subtask 2.3: Open circuit for 5 minutes after threshold reached
  - [ ] Subtask 2.4: Allow single test request after circuit open duration
  - [ ] Subtask 2.5: Reset failure count on successful call
  - [ ] Subtask 2.6: Log circuit state changes
  - [ ] Subtask 2.7: Write unit tests for circuit breaker logic

- [ ] Task 3: Handle plugin disabled scenario (AC: 1)
  - [ ] Subtask 3.1: Check if Playbooks plugin active on approval plugin init
  - [ ] Subtask 3.2: If disabled, set playbooksClient to nil
  - [ ] Subtask 3.3: All playbook code checks for nil client before calling
  - [ ] Subtask 3.4: Log info message when Playbooks integration disabled
  - [ ] Subtask 3.5: Test approval workflow with Playbooks disabled

- [ ] Task 4: Add observability and metrics (AC: 8, 9)
  - [ ] Subtask 4.1: Create metrics struct for tracking calls
  - [ ] Subtask 4.2: Track success/failure counts per operation
  - [ ] Subtask 4.3: Track latency per operation
  - [ ] Subtask 4.4: Add health check endpoint exposing metrics
  - [ ] Subtask 4.5: Add admin command to view playbook integration status

- [ ] Task 5: Integration testing and validation (AC: 1-9)
  - [ ] Subtask 5.1: Test with Playbooks plugin disabled
  - [ ] Subtask 5.2: Test with network failures (mock)
  - [ ] Subtask 5.3: Test with permission errors (403)
  - [ ] Subtask 5.4: Test with deleted playbooks (404 mid-flow)
  - [ ] Subtask 5.5: Test circuit breaker activation and recovery
  - [ ] Subtask 5.6: Verify all errors logged appropriately
  - [ ] Subtask 5.7: Verify no user-visible errors in any scenario

## Dev Notes

### Circuit Breaker Implementation

```go
// In server/playbooks_client.go
type CircuitBreaker struct {
    mu               sync.RWMutex
    failureCount     int
    failureThreshold int  // Default: 5
    resetTimeout     time.Duration  // Default: 5 minutes
    openedAt         time.Time
    state            CircuitState
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota  // Normal operation
    CircuitOpen                         // Blocking calls
    CircuitHalfOpen                     // Testing recovery
)

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    // Check if circuit is open
    if cb.state == CircuitOpen {
        if time.Since(cb.openedAt) < cb.resetTimeout {
            return fmt.Errorf("circuit breaker is open")
        }
        // Try half-open (single test call)
        cb.state = CircuitHalfOpen
    }

    // Attempt the call
    err := fn()

    if err != nil {
        cb.failureCount++
        if cb.failureCount >= cb.failureThreshold {
            cb.state = CircuitOpen
            cb.openedAt = time.Now()
            // Log circuit opened
        }
        return err
    }

    // Success - reset
    if cb.state == CircuitHalfOpen {
        cb.state = CircuitClosed
        // Log circuit closed
    }
    cb.failureCount = 0
    return nil
}
```

### Error Handling Wrapper

```go
func (c *PlaybooksClient) GetPlaybookRunByChannelSafe(channelID string) (*PlaybookRun, error) {
    if c == nil {
        return nil, nil // Client disabled
    }

    var run *PlaybookRun
    var apiErr error

    err := c.circuitBreaker.Call(func() error {
        run, apiErr = c.getPlaybookRunByChannelInternal(channelID)
        return apiErr
    })

    if err != nil {
        if err.Error() == "circuit breaker is open" {
            c.api.LogDebug("Playbooks circuit breaker is open, skipping call")
            return nil, nil // Fail gracefully
        }

        c.api.LogWarn("Failed to get playbook run",
            "channel_id", channelID,
            "error", err.Error())
        return nil, nil // Don't propagate error
    }

    return run, nil
}

func (c *PlaybooksClient) PostPlaybookStatusSafe(runID, message string) error {
    if c == nil {
        return nil // Client disabled
    }

    err := c.circuitBreaker.Call(func() error {
        _, err := c.postPlaybookStatusInternal(runID, message)
        return err
    })

    if err != nil {
        c.api.LogWarn("Failed to post playbook status",
            "run_id", runID,
            "error", err.Error(),
            "circuit_state", c.circuitBreaker.state)
        // Don't return error - fail silently
    }

    return nil
}
```

### Detecting Plugin Status

```go
// In server/plugin.go - OnActivate
func (p *Plugin) OnActivate() error {
    // ... existing initialization ...

    // Check if Playbooks plugin is active
    plugins, appErr := p.API.GetPlugins()
    if appErr != nil {
        p.API.LogWarn("Could not check for Playbooks plugin", "error", appErr.Error())
    } else {
        playbooksActive := false
        for _, plugin := range plugins {
            if plugin.Manifest.Id == "playbooks" && plugin.Manifest.Active {
                playbooksActive = true
                break
            }
        }

        if playbooksActive {
            p.playbooksClient = NewPlaybooksClient(p.API, siteURL, botToken)
            p.API.LogInfo("Playbooks integration enabled")
        } else {
            p.playbooksClient = nil
            p.API.LogInfo("Playbooks plugin not active, integration disabled")
        }
    }

    return nil
}
```

### Metrics and Observability

```go
type PlaybookMetrics struct {
    mu                sync.RWMutex
    DetectionCalls    int64
    DetectionSuccess  int64
    DetectionFailed   int64
    StatusPostCalls   int64
    StatusPostSuccess int64
    StatusPostFailed  int64
    AverageLatencyMs  int64
}

func (m *PlaybookMetrics) RecordDetection(success bool, latency time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.DetectionCalls++
    if success {
        m.DetectionSuccess++
    } else {
        m.DetectionFailed++
    }
    m.updateAvgLatency(latency)
}

// Expose via plugin API or admin command
func (p *Plugin) GetPlaybookMetrics() *PlaybookMetrics {
    if p.playbooksClient == nil {
        return nil
    }
    return p.playbooksClient.metrics
}
```

### Error Logging Standards

**Log Levels:**
- **Debug:** Successful operations, circuit breaker state
- **Warn:** API failures, permission errors, 404s
- **Error:** Unexpected failures, circuit breaker open

**Log Format:**
```go
p.API.LogWarn("Failed to post playbook status",
    "run_id", runID,
    "approval_code", approval.ReferenceCode,
    "error", err.Error(),
    "http_status", resp.StatusCode)
```

### Testing Error Scenarios

```go
func TestApprovalCreation_PlaybooksDisabled(t *testing.T) {
    // Set playbooksClient to nil
    router.playbooksClient = nil

    // Create approval - should work exactly as v1.0
    resp, err := router.executeNew(args)

    require.NoError(t, err)
    // Verify approval created
    // Verify no playbook fields set
}

func TestApprovalCreation_PlaybooksAPIFailure(t *testing.T) {
    // Mock Playbooks API to return 500 error
    mockClient := &MockPlaybooksClient{
        GetPlaybookRunByChannelFunc: func(id string) (*PlaybookRun, error) {
            return nil, fmt.Errorf("API error: 500")
        },
    }
    router.playbooksClient = mockClient

    // Create approval - should succeed despite API failure
    resp, err := router.executeNew(args)

    require.NoError(t, err)
    // Verify approval created without playbook fields
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
    cb := NewCircuitBreaker(3, 1*time.Minute)

    // Simulate 3 failures
    for i := 0; i < 3; i++ {
        cb.Call(func() error {
            return fmt.Errorf("failure")
        })
    }

    // Circuit should be open
    assert.Equal(t, CircuitOpen, cb.state)

    // Next call should fail immediately
    err := cb.Call(func() error {
        panic("should not be called")
    })
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circuit breaker is open")
}
```

### Files to Create/Modify

**New Files:**
- `server/playbooks_circuit_breaker.go` - Circuit breaker implementation
- `server/playbooks_metrics.go` - Metrics tracking

**Modified Files:**
- `server/playbooks_client.go` - Add error handling wrappers
- `server/plugin.go` - Add plugin detection logic
- `server/command/router.go` - Use safe methods with error handling
- All test files - Add error scenario tests

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Circuit breaker implemented and tested
- [ ] Playbooks disabled scenario tested
- [ ] All error scenarios tested (network, permissions, 404, timeout)
- [ ] Metrics and observability implemented
- [ ] Error logging comprehensive and actionable
- [ ] No user-visible errors in any failure scenario
- [ ] Unit tests passing (100% coverage for error paths)
- [ ] Integration tests passing (all error scenarios)
- [ ] Load testing completed (circuit breaker under stress)
- [ ] Code review approved
- [ ] Epic 8 ready for production deployment

## Related Stories

- **Depends on:** All previous stories in Epic 8
- **Completes:** Epic 8 - Playbook Integration

## Technical Debt / Future Improvements

- Add automatic retry with exponential backoff
- Implement distributed circuit breaker (multi-server coordination)
- Add Prometheus/Grafana metrics export
- Add alerting for high failure rates
- Implement adaptive timeout based on latency trends
- Add playbook integration health dashboard
