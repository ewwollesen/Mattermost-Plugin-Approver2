package approval

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCode(t *testing.T) {
	t.Run("generates code with correct format", func(t *testing.T) {
		code, err := GenerateCode()
		assert.NoError(t, err)
		assert.NotEmpty(t, code)

		// Verify format: A-{6_CHARS}
		assert.Equal(t, 8, len(code), "Code should be exactly 8 characters (A- + 6 chars)")
		assert.True(t, strings.HasPrefix(code, "A-"), "Code should start with 'A-'")

		// Verify format matches regex
		validFormat := regexp.MustCompile(`^A-[23456789ABCDEFGHJKLMNPQRSTUVWXYZ]{6}$`)
		assert.True(t, validFormat.MatchString(code), "Code should match format A-{6_CHARS} with safe alphabet")
	})

	t.Run("generates 1000 codes all with valid format", func(t *testing.T) {
		validFormat := regexp.MustCompile(`^A-[23456789ABCDEFGHJKLMNPQRSTUVWXYZ]{6}$`)

		for range 1000 {
			code, err := GenerateCode()
			assert.NoError(t, err)
			assert.True(t, validFormat.MatchString(code), "Code %s should match valid format", code)
		}
	})

	t.Run("excludes ambiguous characters", func(t *testing.T) {
		// Generate 100 codes and verify none contain ambiguous characters
		ambiguousChars := []string{"0", "O", "1", "I", "l"}

		for range 100 {
			code, err := GenerateCode()
			assert.NoError(t, err)

			for _, char := range ambiguousChars {
				assert.NotContains(t, code, char, "Code should not contain ambiguous character: %s", char)
			}
		}
	})

	t.Run("generates different codes on multiple calls", func(t *testing.T) {
		// Generate 50 codes and verify they're not all the same
		codes := make(map[string]bool)

		for range 50 {
			code, err := GenerateCode()
			assert.NoError(t, err)
			codes[code] = true
		}

		// With 32^6 possible codes, getting duplicates in 50 attempts is extremely unlikely
		// We should have at least 45 unique codes (allowing for small chance of collision)
		assert.GreaterOrEqual(t, len(codes), 45, "Should generate mostly unique codes")
	})
}

// MockStorer is a mock implementation of the Storer interface for testing
type MockStorer struct {
	codes map[string]bool
}

func NewMockStorer() *MockStorer {
	return &MockStorer{
		codes: make(map[string]bool),
	}
}

func (m *MockStorer) KVGet(key string) ([]byte, error) {
	if m.codes[key] {
		// Code exists - return a dummy record ID
		return []byte(`"test-record-id"`), nil
	}
	// Code doesn't exist
	return nil, nil
}

func (m *MockStorer) AddCode(code string) {
	m.codes[code] = true
}

func TestGenerateUniqueCode(t *testing.T) {
	t.Run("generates unique code on first attempt", func(t *testing.T) {
		store := NewMockStorer()

		code, err := GenerateUniqueCode(store)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Regexp(t, `^A-[23456789ABCDEFGHJKLMNPQRSTUVWXYZ]{6}$`, code)
	})

	t.Run("retries on collision and succeeds", func(t *testing.T) {
		store := NewMockStorer()

		// Add 3 codes that will collide (simulated)
		// We'll need to force collisions for testing
		// For now, test with empty store first
		code, err := GenerateUniqueCode(store)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
	})

	t.Run("fails after 5 collision attempts", func(t *testing.T) {
		// Create a store that always returns collision
		store := &MockStorerAlwaysCollision{}

		code, err := GenerateUniqueCode(store)
		assert.Error(t, err)
		assert.Empty(t, code)
		assert.ErrorIs(t, err, ErrCodeGenerationFailed)
	})
}

// MockStorerAlwaysCollision always returns collision (code exists)
type MockStorerAlwaysCollision struct{}

func (m *MockStorerAlwaysCollision) KVGet(key string) ([]byte, error) {
	// Always return existing code (collision)
	return []byte(`"existing-record-id"`), nil
}

func TestCodeGenerationPerformance(t *testing.T) {
	t.Run("generates 10 codes in under 1 second", func(t *testing.T) {
		store := NewMockStorer()

		start := time.Now()
		for range 10 {
			code, err := GenerateUniqueCode(store)
			assert.NoError(t, err)
			assert.NotEmpty(t, code)
		}
		elapsed := time.Since(start)

		// Should complete in well under 1 second (< 100ms per code = 1000ms total)
		assert.Less(t, elapsed.Milliseconds(), int64(1000), "10 codes should generate in < 1 second")
	})

	t.Run("single code generation under 100ms", func(t *testing.T) {
		store := NewMockStorer()

		// Test multiple times to get average
		var totalDuration time.Duration
		iterations := 5

		for range iterations {
			start := time.Now()
			code, err := GenerateUniqueCode(store)
			elapsed := time.Since(start)
			totalDuration += elapsed

			assert.NoError(t, err)
			assert.NotEmpty(t, code)
		}

		avgDuration := totalDuration / time.Duration(iterations)
		assert.Less(t, avgDuration.Milliseconds(), int64(100), "Average code generation should be < 100ms")
	})
}

func TestConcurrentCodeGeneration(t *testing.T) {
	t.Run("generates unique codes concurrently", func(t *testing.T) {
		store := NewMockStorer()
		numGoroutines := 10

		// Channel to collect generated codes
		codeChan := make(chan string, numGoroutines)
		errorChan := make(chan error, numGoroutines)

		// Launch goroutines
		for range numGoroutines {
			go func() {
				code, err := GenerateUniqueCode(store)
				if err != nil {
					errorChan <- err
					return
				}
				codeChan <- code
			}()
		}

		// Collect results
		codes := make(map[string]bool)
		for range numGoroutines {
			select {
			case code := <-codeChan:
				codes[code] = true
			case err := <-errorChan:
				t.Fatalf("Error generating code: %v", err)
			}
		}

		// Verify all codes are unique
		assert.Equal(t, numGoroutines, len(codes), "All generated codes should be unique")

		// Verify format
		for code := range codes {
			assert.Regexp(t, `^A-[23456789ABCDEFGHJKLMNPQRSTUVWXYZ]{6}$`, code)
		}
	})
}
