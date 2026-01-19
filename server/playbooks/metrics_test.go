package playbooks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Story 8.6 AC9: Metrics tracked for playbook integration success/failure rates

func TestMetrics_InitialState(t *testing.T) {
	m := NewMetrics()

	assert.Equal(t, int64(0), m.DetectionCalls)
	assert.Equal(t, int64(0), m.DetectionSuccess)
	assert.Equal(t, int64(0), m.DetectionFailed)
	assert.Equal(t, int64(0), m.StatusPostCalls)
	assert.Equal(t, int64(0), m.StatusPostSuccess)
	assert.Equal(t, int64(0), m.StatusPostFailed)
	assert.Equal(t, CircuitClosed, m.CircuitBreakerState)
	assert.Equal(t, int64(0), m.CircuitBreakerOpens)
}

func TestMetrics_RecordDetectionSuccess(t *testing.T) {
	m := NewMetrics()

	m.RecordDetection(true, 100*time.Millisecond)

	assert.Equal(t, int64(1), m.DetectionCalls)
	assert.Equal(t, int64(1), m.DetectionSuccess)
	assert.Equal(t, int64(0), m.DetectionFailed)
	assert.Equal(t, 100*time.Millisecond, m.DetectionLatency)
}

func TestMetrics_RecordDetectionFailure(t *testing.T) {
	m := NewMetrics()

	m.RecordDetection(false, 50*time.Millisecond)

	assert.Equal(t, int64(1), m.DetectionCalls)
	assert.Equal(t, int64(0), m.DetectionSuccess)
	assert.Equal(t, int64(1), m.DetectionFailed)
	assert.Equal(t, 50*time.Millisecond, m.DetectionLatency)
}

func TestMetrics_RecordStatusPostSuccess(t *testing.T) {
	m := NewMetrics()

	m.RecordStatusPost(true, 200*time.Millisecond)

	assert.Equal(t, int64(1), m.StatusPostCalls)
	assert.Equal(t, int64(1), m.StatusPostSuccess)
	assert.Equal(t, int64(0), m.StatusPostFailed)
	assert.Equal(t, 200*time.Millisecond, m.StatusPostLatency)
}

func TestMetrics_RecordStatusPostFailure(t *testing.T) {
	m := NewMetrics()

	m.RecordStatusPost(false, 150*time.Millisecond)

	assert.Equal(t, int64(1), m.StatusPostCalls)
	assert.Equal(t, int64(0), m.StatusPostSuccess)
	assert.Equal(t, int64(1), m.StatusPostFailed)
	assert.Equal(t, 150*time.Millisecond, m.StatusPostLatency)
}

func TestMetrics_AverageLatency(t *testing.T) {
	m := NewMetrics()

	// Record 3 detections with different latencies
	m.RecordDetection(true, 100*time.Millisecond)
	m.RecordDetection(true, 200*time.Millisecond)
	m.RecordDetection(true, 300*time.Millisecond)

	// Average should be 200ms
	assert.Equal(t, int64(3), m.DetectionCalls)
	assert.Equal(t, 200*time.Millisecond, m.DetectionLatency)
}

func TestMetrics_GetDetectionSuccessRate(t *testing.T) {
	tests := []struct {
		name         string
		successes    int
		failures     int
		expectedRate float64
	}{
		{"no calls", 0, 0, 100.0},
		{"all successful", 10, 0, 100.0},
		{"all failed", 0, 10, 0.0},
		{"half successful", 5, 5, 50.0},
		{"80% successful", 8, 2, 80.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()
			for i := 0; i < tt.successes; i++ {
				m.RecordDetection(true, 10*time.Millisecond)
			}
			for i := 0; i < tt.failures; i++ {
				m.RecordDetection(false, 10*time.Millisecond)
			}

			rate := m.GetDetectionSuccessRate()
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

func TestMetrics_GetStatusPostSuccessRate(t *testing.T) {
	tests := []struct {
		name         string
		successes    int
		failures     int
		expectedRate float64
	}{
		{"no calls", 0, 0, 100.0},
		{"all successful", 10, 0, 100.0},
		{"all failed", 0, 10, 0.0},
		{"half successful", 5, 5, 50.0},
		{"80% successful", 8, 2, 80.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()
			for i := 0; i < tt.successes; i++ {
				m.RecordStatusPost(true, 10*time.Millisecond)
			}
			for i := 0; i < tt.failures; i++ {
				m.RecordStatusPost(false, 10*time.Millisecond)
			}

			rate := m.GetStatusPostSuccessRate()
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

func TestMetrics_UpdateCircuitBreakerState(t *testing.T) {
	m := NewMetrics()

	// Closed -> Open (first open)
	m.UpdateCircuitBreakerState(CircuitOpen)
	assert.Equal(t, CircuitOpen, m.CircuitBreakerState)
	assert.Equal(t, int64(1), m.CircuitBreakerOpens)

	// Open -> Closed
	m.UpdateCircuitBreakerState(CircuitClosed)
	assert.Equal(t, CircuitClosed, m.CircuitBreakerState)
	assert.Equal(t, int64(1), m.CircuitBreakerOpens)

	// Closed -> Open (second open)
	m.UpdateCircuitBreakerState(CircuitOpen)
	assert.Equal(t, int64(2), m.CircuitBreakerOpens)

	// Open -> Half-Open
	m.UpdateCircuitBreakerState(CircuitHalfOpen)
	assert.Equal(t, CircuitHalfOpen, m.CircuitBreakerState)
	assert.Equal(t, int64(2), m.CircuitBreakerOpens)

	// Half-Open -> Open (reopening)
	m.UpdateCircuitBreakerState(CircuitOpen)
	assert.Equal(t, int64(3), m.CircuitBreakerOpens)
}

func TestMetrics_GetSnapshot(t *testing.T) {
	m := NewMetrics()

	// Record some metrics
	m.RecordDetection(true, 100*time.Millisecond)
	m.RecordStatusPost(false, 200*time.Millisecond)
	m.UpdateCircuitBreakerState(CircuitOpen)

	// Get snapshot
	snapshot := m.GetSnapshot()

	// Verify snapshot matches current state
	assert.Equal(t, int64(1), snapshot.DetectionCalls)
	assert.Equal(t, int64(1), snapshot.DetectionSuccess)
	assert.Equal(t, int64(1), snapshot.StatusPostCalls)
	assert.Equal(t, int64(1), snapshot.StatusPostFailed)
	assert.Equal(t, CircuitOpen, snapshot.CircuitBreakerState)
	assert.Equal(t, int64(1), snapshot.CircuitBreakerOpens)

	// Modify original - snapshot should be unaffected
	m.RecordDetection(true, 100*time.Millisecond)
	assert.Equal(t, int64(1), snapshot.DetectionCalls)
	assert.Equal(t, int64(2), m.DetectionCalls)
}

func TestMetrics_ThreadSafety(t *testing.T) {
	m := NewMetrics()

	// Simulate concurrent metric recording
	done := make(chan bool)

	// Spawn 10 goroutines recording detections
	for range 10 {
		go func() {
			for range 100 {
				m.RecordDetection(true, 10*time.Millisecond)
			}
			done <- true
		}()
	}

	// Spawn 10 goroutines recording status posts
	for range 10 {
		go func() {
			for range 100 {
				m.RecordStatusPost(true, 10*time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 20 {
		<-done
	}

	// Verify counts
	assert.Equal(t, int64(1000), m.DetectionCalls)
	assert.Equal(t, int64(1000), m.StatusPostCalls)
}
