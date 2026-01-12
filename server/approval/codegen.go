package approval

import (
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	// codeChars is the character set for approval codes, excluding ambiguous characters.
	// Excludes: 0 (zero), O (letter), 1 (one), I (capital i), l (lowercase L)
	// This prevents confusion when codes are spoken or written.
	codeChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

	// maxRetries is the maximum number of attempts to generate a unique code
	maxRetries = 5
)

// ErrCodeGenerationFailed is returned when unable to generate a unique code after maximum retries
var ErrCodeGenerationFailed = errors.New("failed to generate unique approval code after maximum retries")

// GenerateCode generates a human-friendly approval reference code in format A-X7K9Q2.
// Uses cryptographically secure random generation (crypto/rand) and excludes
// ambiguous characters (0, O, 1, I, l) to prevent confusion.
// Returns an 8-character string: "A-" prefix + 6 random characters.
func GenerateCode() (string, error) {
	const codeLength = 6

	// Generate 6 random bytes
	randomBytes := make([]byte, codeLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Map each byte to a character from the safe alphabet
	code := make([]byte, codeLength)
	for i, b := range randomBytes {
		code[i] = codeChars[int(b)%len(codeChars)]
	}

	return fmt.Sprintf("A-%s", string(code)), nil
}

// Storer defines the interface for checking code uniqueness
type Storer interface {
	KVGet(key string) ([]byte, error)
}

// GenerateUniqueCode generates a unique approval code by checking against existing codes.
// Retries up to 5 times if collisions occur. Returns ErrCodeGenerationFailed if all retries fail.
// Note: Collisions are extremely rare (< 0.001% with 1M codes). In production, these should be
// logged at the plugin level using plugin.API.LogWarn() for monitoring.
func GenerateUniqueCode(store Storer) (string, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Generate a code
		code, err := GenerateCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate code at attempt %d: %w", attempt, err)
		}

		// Check if code already exists
		key := fmt.Sprintf("approval_code:%s", code)
		existing, err := store.KVGet(key)
		if err != nil {
			return "", fmt.Errorf("failed to check code uniqueness: %w", err)
		}

		// If code doesn't exist, we found a unique one
		if existing == nil {
			return code, nil
		}

		// Code exists - collision detected, retry
		// Note: Caller should log this at WARN level:
		// logger.LogWarn("Approval code collision detected", "code", code, "attempt", attempt, "max_retries", maxRetries)
	}

	// All retries exhausted
	return "", ErrCodeGenerationFailed
}
