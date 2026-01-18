package playbooks

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Story 8.6: Unit tests for circuit breaker logic

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)

	assert.Equal(t, CircuitClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

func TestCircuitBreaker_SuccessfulCall(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)

	err := cb.Call(func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, CircuitClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

func TestCircuitBreaker_SingleFailure(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)

	err := cb.Call(func() error {
		return fmt.Errorf("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, CircuitClosed, cb.GetState())
	assert.Equal(t, 1, cb.GetFailureCount())
}

func TestCircuitBreaker_OpensAfterThresholdReached(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Minute)

	// Simulate 3 failures
	for i := 0; i < 3; i++ {
		err := cb.Call(func() error {
			return fmt.Errorf("test error %d", i)
		})
		assert.Error(t, err)
	}

	// Circuit should be open
	assert.Equal(t, CircuitOpen, cb.GetState())
	assert.Equal(t, 3, cb.GetFailureCount())
}

func TestCircuitBreaker_BlocksCallsWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 5*time.Minute)

	// Open the circuit with failures
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return fmt.Errorf("test error")
		})
	}

	assert.Equal(t, CircuitOpen, cb.GetState())

	// Next call should fail immediately without executing function
	callExecuted := false
	err := cb.Call(func() error {
		callExecuted = true
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
	assert.False(t, callExecuted, "Function should not be called when circuit is open")
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond) // Short timeout for testing

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return fmt.Errorf("test error")
		})
	}

	assert.Equal(t, CircuitOpen, cb.GetState())

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Next call should attempt (half-open state)
	callExecuted := false
	err := cb.Call(func() error {
		callExecuted = true
		return nil // Success
	})

	assert.NoError(t, err)
	assert.True(t, callExecuted, "Function should be called in half-open state")
	assert.Equal(t, CircuitClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

func TestCircuitBreaker_HalfOpenFailureReopensCircuit(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return fmt.Errorf("test error")
		})
	}

	assert.Equal(t, CircuitOpen, cb.GetState())

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Fail in half-open state
	err := cb.Call(func() error {
		return fmt.Errorf("still failing")
	})

	assert.Error(t, err)
	assert.Equal(t, CircuitOpen, cb.GetState())
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)

	// Simulate 2 failures
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return fmt.Errorf("test error")
		})
	}

	assert.Equal(t, 2, cb.GetFailureCount())

	// Success resets counter
	err := cb.Call(func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, CircuitClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(2, 5*time.Minute)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return fmt.Errorf("test error")
		})
	}

	assert.Equal(t, CircuitOpen, cb.GetState())

	// Manual reset
	cb.Reset()

	assert.Equal(t, CircuitClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

// Story 8.6 AC7: Circuit breaker prevents repeated API failures
func TestCircuitBreaker_PreventRepeatedFailures(t *testing.T) {
	cb := NewCircuitBreaker(5, 1*time.Minute)

	// Simulate 10 failures
	errorCount := 0
	for i := 0; i < 10; i++ {
		err := cb.Call(func() error {
			errorCount++
			return fmt.Errorf("API failure")
		})
		if err != nil && err.Error() == "circuit breaker is open" {
			break
		}
	}

	// Circuit should open after 5 failures, preventing remaining calls
	assert.Equal(t, CircuitOpen, cb.GetState())
	assert.Equal(t, 5, errorCount, "Circuit breaker should prevent calls after threshold")
}
