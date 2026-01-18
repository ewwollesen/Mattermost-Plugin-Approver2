// Package playbooks provides integration with the Mattermost Playbooks plugin,
// including circuit breaker protection, metrics tracking, and graceful error handling
// for production reliability (Story 8.6: Error Handling and Graceful Fallback).
//
// The package implements:
//   - Circuit breaker pattern to prevent repeated API failures
//   - Success/failure rate metrics and latency tracking
//   - Graceful degradation when Playbooks is unavailable
//   - User-context authentication for proper permission checking
package playbooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// PlaybookRun represents a playbook run from the Playbooks API
type PlaybookRun struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerUserID string `json:"owner_user_id"`
	TeamID      string `json:"team_id"`
	ChannelID   string `json:"channel_id"`
	CreateAt    int64  `json:"create_at"`
	EndAt       int64  `json:"end_at"`
	PlaybookID  string `json:"playbook_id"`
}

// API call timeouts - fast failure is better than blocking approval workflow
const (
	// PlaybooksAPITimeout is the maximum time to wait for Playbooks API responses
	// Set to 500ms for fast failure - read operations should be quick
	PlaybooksAPITimeout = 500 * time.Millisecond
)

// ClientInterface defines the interface for interacting with the Playbooks plugin
// This interface is implemented by Client and can be mocked for testing
type ClientInterface interface {
	GetPlaybookRunByChannel(channelID string, requesterUserID string) (*PlaybookRun, error)
	PostMessageToPlaybookChannel(channelID string, message string) (string, error)
	UpdateMessageInPlaybookChannel(channelID string, postID string, message string) error
	GetMetrics() Metrics
}

// Client handles communication with the Mattermost Playbooks plugin API
// Story 8.6: Enhanced with circuit breaker and metrics for production reliability
type Client struct {
	api            plugin.API
	siteURL        string
	botUserID      string
	circuitBreaker *CircuitBreaker
	metrics        *Metrics
}

// NewClient creates a new Playbooks API client
// Story 8.6: Circuit breaker prevents repeated failures (5 failures, 5-minute timeout)
// GitHub Issue #2: Added botUserID for posting messages as the bot
func NewClient(api plugin.API, siteURL string, botUserID string) *Client {
	cb := NewCircuitBreaker(5, 5*time.Minute)
	cb.SetLogger(api) // Enable circuit breaker state change logging

	return &Client{
		api:            api,
		siteURL:        siteURL,
		botUserID:      botUserID,
		circuitBreaker: cb,
		metrics:        NewMetrics(),
	}
}

// getUserToken retrieves or creates a user access token for Playbooks API calls
// Tokens are cached in KV store per user to avoid recreating on every call
func (c *Client) getUserToken(userID string) (string, error) {
	tokenKey := fmt.Sprintf("user_playbooks_token_%s", userID)

	// Check if token already exists in KV store
	tokenBytes, appErr := c.api.KVGet(tokenKey)
	if appErr != nil {
		return "", fmt.Errorf("failed to check for existing user token: %w", appErr)
	}

	// If token exists, reuse it
	if len(tokenBytes) > 0 {
		return string(tokenBytes), nil
	}

	// Create new personal access token for user
	token, appErr := c.api.CreateUserAccessToken(&model.UserAccessToken{
		UserId:      userID,
		Description: "Approver Plugin - Playbooks API Access",
	})
	if appErr != nil {
		return "", fmt.Errorf("failed to create user access token: %w", appErr)
	}

	// Store token in KV store for reuse
	if appErr := c.api.KVSet(tokenKey, []byte(token.Token)); appErr != nil {
		c.api.LogWarn("Failed to store user access token in KV store",
			"user_id", userID,
			"error", appErr.Error())
		// Continue - token is still valid for this request
	}

	return token.Token, nil
}

// GetPlaybookRunByChannel retrieves playbook run information for a given channel
// Uses the requester's user token for authentication - ensures proper permission checking
// Returns nil if no playbook is associated with the channel (404) or user lacks access
// Story 8.6: Enhanced with circuit breaker and metrics (AC2, AC5, AC6, AC7, AC9)
func (c *Client) GetPlaybookRunByChannel(channelID string, requesterUserID string) (*PlaybookRun, error) {
	startTime := time.Now()
	var run *PlaybookRun
	var callErr error

	// Wrap in circuit breaker to prevent repeated failures
	err := c.circuitBreaker.Call(func() error {
		run, callErr = c.getPlaybookRunByChannelInternal(channelID, requesterUserID)
		return callErr
	})

	// Record metrics
	latency := time.Since(startTime)
	success := err == nil && callErr == nil
	c.metrics.RecordDetection(success, latency)
	c.metrics.UpdateCircuitBreakerState(c.circuitBreaker.GetState())

	// Handle circuit breaker open state (AC7)
	if err != nil && err.Error() == "circuit breaker is open" {
		c.api.LogDebug("Playbooks circuit breaker is open, skipping detection call",
			"channel_id", channelID,
			"failure_count", c.circuitBreaker.GetFailureCount())
		return nil, nil // Fail gracefully (AC6)
	}

	// Handle API errors gracefully (AC2, AC5, AC6)
	if callErr != nil {
		c.api.LogWarn("Failed to get playbook run",
			"channel_id", channelID,
			"requester_user_id", requesterUserID,
			"error", callErr.Error(),
			"circuit_state", c.circuitBreaker.GetState().String())
		return nil, nil // Don't propagate error (AC6)
	}

	return run, nil
}

// getPlaybookRunByChannelInternal is the internal implementation that performs the actual API call
func (c *Client) getPlaybookRunByChannelInternal(channelID string, requesterUserID string) (*PlaybookRun, error) {
	url := fmt.Sprintf("%s/plugins/playbooks/api/v0/runs/channel/%s",
		c.siteURL, channelID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get user token for authentication (requester's context)
	userToken, err := c.getUserToken(requesterUserID)
	if err != nil {
		// If we can't get user token, return gracefully (playbook detection is best effort)
		return nil, fmt.Errorf("failed to get user token: %w", err)
	}

	// Use requester's token for authentication
	req.Header.Set("Authorization", "Bearer "+userToken)

	// 500ms timeout - read operations should be fast, fail quickly if playbooks unavailable
	client := &http.Client{Timeout: PlaybooksAPITimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Playbooks API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 404 means no playbook for this channel OR user lacks access - normal case, not an error (AC4)
	// Playbooks API returns 404 for both "not found" and "no permission" scenarios
	if resp.StatusCode == 404 {
		return nil, nil
	}

	// 401 means authentication failed - graceful degradation if user token creation failed
	if resp.StatusCode == 401 {
		return nil, nil
	}

	// 403 means permission denied - log warning and continue (AC3)
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("permission denied (403)")
	}

	// Any other non-200 status is an error
	if resp.StatusCode != 200 {
		// Read response body for debugging context
		body, _ := io.ReadAll(resp.Body)
		bodyPreview := string(body)
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..."
		}
		return nil, fmt.Errorf("playbooks API returned status %d for %s: %s", resp.StatusCode, url, bodyPreview)
	}

	var run PlaybookRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &run, nil
}

// PostMessageToPlaybookChannel posts a message to a playbook channel using standard CreatePost
// GitHub Issue #2: Using CreatePost with markdown tables for nice formatting without Playbooks API side effects
// Returns the post ID for future updates
func (c *Client) PostMessageToPlaybookChannel(channelID string, message string) (string, error) {
	startTime := time.Now()
	var postID string
	var callErr error

	// Wrap in circuit breaker to prevent repeated failures
	err := c.circuitBreaker.Call(func() error {
		postID, callErr = c.postMessageToPlaybookChannelInternal(channelID, message)
		return callErr
	})

	// Record metrics
	latency := time.Since(startTime)
	success := err == nil && callErr == nil
	c.metrics.RecordStatusPost(success, latency)
	c.metrics.UpdateCircuitBreakerState(c.circuitBreaker.GetState())

	// Handle circuit breaker open state (AC7)
	if err != nil && err.Error() == "circuit breaker is open" {
		c.api.LogDebug("Playbooks circuit breaker is open, skipping channel post",
			"channel_id", channelID,
			"failure_count", c.circuitBreaker.GetFailureCount())
		return "", nil // Fail gracefully (AC6)
	}

	// Handle API errors gracefully (AC2, AC3, AC4, AC5, AC6)
	if callErr != nil {
		c.api.LogWarn("Failed to post message to playbook channel",
			"channel_id", channelID,
			"error", callErr.Error(),
			"circuit_state", c.circuitBreaker.GetState().String())
		return "", nil // Don't propagate error (AC6)
	}

	return postID, nil
}

// postMessageToPlaybookChannelInternal is the internal implementation that posts a message
// GitHub Issue #2: Using CreatePost with markdown tables for nice formatting without Playbooks API side effects
func (c *Client) postMessageToPlaybookChannelInternal(channelID string, message string) (string, error) {
	post := &model.Post{
		UserId:    c.botUserID,
		ChannelId: channelID,
		Message:   message, // Contains markdown table
	}

	createdPost, appErr := c.api.CreatePost(post)
	if appErr != nil {
		return "", fmt.Errorf("failed to create post: %w", appErr)
	}

	return createdPost.Id, nil
}

// UpdateMessageInPlaybookChannel updates an existing post in the playbook channel
// Used to update the original status post when approval state changes (like DM behavior)
// GitHub Issue #2: Using UpdatePost instead of Playbooks API
func (c *Client) UpdateMessageInPlaybookChannel(channelID string, postID string, message string) error {
	startTime := time.Now()
	var callErr error

	// Wrap in circuit breaker to prevent repeated failures
	err := c.circuitBreaker.Call(func() error {
		callErr = c.updateMessageInPlaybookChannelInternal(channelID, postID, message)
		return callErr
	})

	// Record metrics
	latency := time.Since(startTime)
	success := err == nil && callErr == nil
	c.metrics.RecordStatusPost(success, latency)
	c.metrics.UpdateCircuitBreakerState(c.circuitBreaker.GetState())

	// Handle circuit breaker open state (AC7)
	if err != nil && err.Error() == "circuit breaker is open" {
		c.api.LogDebug("Playbooks circuit breaker is open, skipping post update",
			"channel_id", channelID,
			"post_id", postID,
			"failure_count", c.circuitBreaker.GetFailureCount())
		return nil // Fail gracefully (AC6)
	}

	// Handle API errors gracefully (AC2, AC3, AC4, AC5, AC6)
	if callErr != nil {
		c.api.LogWarn("Failed to update message in playbook channel",
			"channel_id", channelID,
			"post_id", postID,
			"error", callErr.Error(),
			"circuit_state", c.circuitBreaker.GetState().String())
		return nil // Don't propagate error (AC6)
	}

	return nil
}

// updateMessageInPlaybookChannelInternal updates an existing post
// GitHub Issue #2: Using UpdatePost instead of Playbooks API
func (c *Client) updateMessageInPlaybookChannelInternal(channelID string, postID string, message string) error {
	// Get the existing post
	existingPost, appErr := c.api.GetPost(postID)
	if appErr != nil {
		return fmt.Errorf("failed to get post: %w", appErr)
	}

	// Update the message content
	existingPost.Message = message

	// Update the post
	_, appErr = c.api.UpdatePost(existingPost)
	if appErr != nil {
		return fmt.Errorf("failed to update post: %w", appErr)
	}

	return nil
}

// GetMetrics returns a snapshot of current metrics
func (c *Client) GetMetrics() Metrics {
	if c.metrics == nil {
		return Metrics{}
	}
	return c.metrics.GetSnapshot()
}

// GetCircuitBreakerState returns the current circuit breaker state
func (c *Client) GetCircuitBreakerState() CircuitState {
	if c.circuitBreaker == nil {
		return CircuitClosed
	}
	return c.circuitBreaker.GetState()
}

