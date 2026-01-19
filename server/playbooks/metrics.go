package playbooks

import (
	"sync"
	"time"
)

// Metrics tracks success/failure rates and latency for Playbooks API calls
// Story 8.6 AC9: Track playbook integration success/failure rates
type Metrics struct {
	mu sync.RWMutex

	// Detection metrics (GetPlaybookRunByChannel calls)
	DetectionCalls   int64
	DetectionSuccess int64
	DetectionFailed  int64
	DetectionLatency time.Duration

	// Status post metrics (PostMessageToPlaybookChannel / UpdateMessageInPlaybookChannel calls)
	StatusPostCalls   int64
	StatusPostSuccess int64
	StatusPostFailed  int64
	StatusPostLatency time.Duration

	// Circuit breaker state
	CircuitBreakerState CircuitState
	CircuitBreakerOpens int64 // Count of times circuit breaker opened
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		CircuitBreakerState: CircuitClosed,
	}
}

// RecordDetection records a playbook detection call result
func (m *Metrics) RecordDetection(success bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DetectionCalls++
	if success {
		m.DetectionSuccess++
	} else {
		m.DetectionFailed++
	}

	// Update average latency using clearer logic
	oldCount := m.DetectionCalls - 1
	if oldCount == 0 {
		m.DetectionLatency = latency
	} else {
		// Running average: weighted sum of old average and new sample
		m.DetectionLatency = (m.DetectionLatency*time.Duration(oldCount) + latency) / time.Duration(m.DetectionCalls)
	}
}

// RecordStatusPost records a playbook status post call result
func (m *Metrics) RecordStatusPost(success bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StatusPostCalls++
	if success {
		m.StatusPostSuccess++
	} else {
		m.StatusPostFailed++
	}

	// Update average latency using clearer logic
	oldCount := m.StatusPostCalls - 1
	if oldCount == 0 {
		m.StatusPostLatency = latency
	} else {
		// Running average: weighted sum of old average and new sample
		m.StatusPostLatency = (m.StatusPostLatency*time.Duration(oldCount) + latency) / time.Duration(m.StatusPostCalls)
	}
}

// UpdateCircuitBreakerState updates the circuit breaker state metric
func (m *Metrics) UpdateCircuitBreakerState(state CircuitState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.CircuitBreakerState
	m.CircuitBreakerState = state

	// Track circuit breaker opens
	if oldState != CircuitOpen && state == CircuitOpen {
		m.CircuitBreakerOpens++
	}
}

// GetSnapshot returns a copy of current metrics (thread-safe)
// Note: Creates a new Metrics struct without copying the mutex
func (m *Metrics) GetSnapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy fields explicitly to avoid copying the mutex
	return Metrics{
		DetectionCalls:      m.DetectionCalls,
		DetectionSuccess:    m.DetectionSuccess,
		DetectionFailed:     m.DetectionFailed,
		DetectionLatency:    m.DetectionLatency,
		StatusPostCalls:     m.StatusPostCalls,
		StatusPostSuccess:   m.StatusPostSuccess,
		StatusPostFailed:    m.StatusPostFailed,
		StatusPostLatency:   m.StatusPostLatency,
		CircuitBreakerState: m.CircuitBreakerState,
		CircuitBreakerOpens: m.CircuitBreakerOpens,
	}
}

// GetDetectionSuccessRate returns the detection success rate as percentage (0-100)
func (m *Metrics) GetDetectionSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.DetectionCalls == 0 {
		return 100.0 // No calls yet = 100% success
	}
	return (float64(m.DetectionSuccess) / float64(m.DetectionCalls)) * 100.0
}

// GetStatusPostSuccessRate returns the status post success rate as percentage (0-100)
func (m *Metrics) GetStatusPostSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.StatusPostCalls == 0 {
		return 100.0 // No calls yet = 100% success
	}
	return (float64(m.StatusPostSuccess) / float64(m.StatusPostCalls)) * 100.0
}
