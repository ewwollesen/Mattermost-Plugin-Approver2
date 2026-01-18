package playbooks

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the current state of the circuit breaker
type CircuitState int

const (
	// CircuitClosed - Normal operation, calls allowed
	CircuitClosed CircuitState = iota
	// CircuitOpen - Blocking all calls due to failures
	CircuitOpen
	// CircuitHalfOpen - Testing if service has recovered
	CircuitHalfOpen
)

// String returns string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Logger interface for circuit breaker logging
type Logger interface {
	LogWarn(message string, keyValuePairs ...interface{})
	LogInfo(message string, keyValuePairs ...interface{})
}

// CircuitBreaker implements the circuit breaker pattern to prevent repeated
// failed calls to Playbooks API (Story 8.6 AC7)
type CircuitBreaker struct {
	mu               sync.RWMutex
	failureCount     int
	failureThreshold int           // Default: 5
	resetTimeout     time.Duration // Default: 5 minutes
	openedAt         time.Time
	state            CircuitState
	logger           Logger // Optional logger for state changes
}

// NewCircuitBreaker creates a new circuit breaker with specified threshold and timeout
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            CircuitClosed,
	}
}

// SetLogger sets the logger for circuit breaker state change logging
func (cb *CircuitBreaker) SetLogger(logger Logger) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.logger = logger
}

// Call executes the given function with circuit breaker protection
// Returns error if circuit is open or if the function fails
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit is open
	if cb.state == CircuitOpen {
		// Check if enough time has passed to try again
		if time.Since(cb.openedAt) < cb.resetTimeout {
			return fmt.Errorf("circuit breaker is open")
		}
		// Try half-open (single test call)
		cb.state = CircuitHalfOpen
		if cb.logger != nil {
			cb.logger.LogInfo("Circuit breaker entering half-open state (testing recovery)",
				"failure_threshold", cb.failureThreshold,
				"reset_timeout", cb.resetTimeout.String())
		}
	}

	// Attempt the call
	err := fn()

	if err != nil {
		cb.failureCount++

		// If we're in half-open state, reopen immediately on failure
		if cb.state == CircuitHalfOpen {
			cb.state = CircuitOpen
			cb.openedAt = time.Now()
			if cb.logger != nil {
				cb.logger.LogWarn("Circuit breaker reopened - recovery test failed",
					"reset_timeout", cb.resetTimeout.String())
			}
			return err
		}

		// Check if we've reached failure threshold
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitOpen
			cb.openedAt = time.Now()
			if cb.logger != nil {
				cb.logger.LogWarn("Circuit breaker opened - too many failures",
					"failure_count", cb.failureCount,
					"failure_threshold", cb.failureThreshold,
					"reset_timeout", cb.resetTimeout.String())
			}
		}
		return err
	}

	// Success - reset circuit breaker
	if cb.state == CircuitHalfOpen {
		// Log recovery when transitioning from half-open to closed
		if cb.logger != nil {
			cb.logger.LogInfo("Circuit breaker closed - service recovered",
				"previous_failures", cb.failureCount)
		}
		cb.state = CircuitClosed
		cb.failureCount = 0
	} else if cb.state == CircuitClosed {
		cb.failureCount = 0
	}

	return nil
}

// GetState returns the current circuit breaker state (thread-safe)
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count (thread-safe)
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failureCount = 0
}
