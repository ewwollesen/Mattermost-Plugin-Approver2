package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost-plugin-approver2/server/notifications"
	"github.com/mattermost/mattermost-plugin-approver2/server/store"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// TODO: Future enhancement - make timeout configurable
// Configuration options:
// - Default timeout duration (e.g., 15min, 30min, 1hr, 24hr)
// - Per-channel timeout overrides
// - Disable timeout feature entirely
const DefaultTimeoutDuration = 30 * time.Minute // Hardcoded for MVP

// TimeoutChecker periodically scans for timed-out pending approval requests
// and automatically cancels them with notification to the requester.
type TimeoutChecker struct {
	store     *store.KVStore
	service   *approval.Service
	api       plugin.API
	botUserID string
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewChecker creates a new TimeoutChecker instance.
func NewChecker(store *store.KVStore, service *approval.Service, api plugin.API, botUserID string) *TimeoutChecker {
	return &TimeoutChecker{
		store:     store,
		service:   service,
		api:       api,
		botUserID: botUserID,
		done:      make(chan struct{}),
	}
}

// Start launches the background goroutine that checks for timed-out requests.
// It uses a 5-minute ticker to periodically scan for requests older than 30 minutes.
func (tc *TimeoutChecker) Start() {
	tc.ctx, tc.cancel = context.WithCancel(context.Background())

	go tc.run()

	tc.api.LogInfo("Timeout checker started", "check_interval", "5m", "timeout_duration", "30m")
}

// Stop gracefully shuts down the timeout checker goroutine.
func (tc *TimeoutChecker) Stop() {
	if tc.cancel != nil {
		tc.cancel()
	}
	// Wait for goroutine to exit
	<-tc.done

	tc.api.LogInfo("Timeout checker stopped")
}

// run is the main loop that periodically checks for timed-out requests.
func (tc *TimeoutChecker) run() {
	defer close(tc.done)
	defer func() {
		if r := recover(); r != nil {
			tc.api.LogError("Timeout checker panic recovered", "panic", r)
		}
	}()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-tc.ctx.Done():
			return
		case <-ticker.C:
			if err := tc.checkTimeouts(); err != nil {
				tc.api.LogError("Timeout check failed", "error", err.Error())
			}
		}
	}
}

// checkTimeouts queries for pending requests older than the timeout duration
// and processes them for auto-cancellation.
func (tc *TimeoutChecker) checkTimeouts() error {
	records, err := tc.store.GetPendingRequestsOlderThan(DefaultTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to query timed-out requests: %w", err)
	}

	if len(records) == 0 {
		tc.api.LogDebug("Completed timeout scan",
			"eligible_count", 0,
			"canceled_count", 0,
			"failed_count", 0)
		return nil
	}

	tc.api.LogDebug("Processing timed-out requests",
		"count", len(records),
		"timeout_duration", DefaultTimeoutDuration.String())

	// Track metrics for aggregate logging
	canceledCount := 0
	failedCount := 0

	// Process each timed-out request with graceful degradation:
	// Critical path (cancellation) must succeed; notifications are best-effort
	// Architecture Decision 2.2: Notification failures don't block state changes
	for _, record := range records {
		// Auto-cancel with timeout reason (critical path - must succeed)
		if err := tc.service.CancelApprovalByID(record.ID, record.RequesterID, true); err != nil {
			tc.api.LogError("Failed to auto-cancel timed-out request",
				"approval_id", record.ID,
				"approval_code", record.Code,
				"error", err.Error())
			failedCount++
			continue
		}

		canceledCount++

		tc.api.LogInfo("Auto-canceled timed-out request",
			"approval_id", record.ID,
			"approval_code", record.Code,
			"requester_id", record.RequesterID)

		// Reload record to get updated cancel information
		updatedRecord, err := tc.store.GetApproval(record.ID)
		if err != nil {
			tc.api.LogError("Failed to reload canceled record for notification",
				"approval_id", record.ID,
				"error", err.Error())
			continue
		}

		// Send timeout notification to requester (best-effort, graceful degradation)
		if _, err := notifications.SendTimeoutNotificationDM(tc.api, tc.botUserID, updatedRecord); err != nil {
			tc.api.LogWarn("Failed to send timeout notification",
				"approval_id", record.ID,
				"approval_code", record.Code,
				"requester_id", record.RequesterID,
				"error", err.Error())
			// Continue - notification failure doesn't affect cancellation
		}

		// Update approver's original notification to disable buttons (best-effort)
		// For auto-timeout, show "System" as the canceler
		if err := notifications.UpdateApprovalPostForCancellation(tc.api, updatedRecord, "System"); err != nil {
			tc.api.LogWarn("Failed to update approver post after timeout",
				"approval_id", record.ID,
				"approval_code", record.Code,
				"post_id", updatedRecord.NotificationPostID,
				"error", err.Error())
			// Continue - post update failure doesn't affect cancellation
		}
	}

	// Log completion summary with aggregate metrics
	tc.api.LogDebug("Completed timeout scan",
		"eligible_count", len(records),
		"canceled_count", canceledCount,
		"failed_count", failedCount)

	return nil
}
