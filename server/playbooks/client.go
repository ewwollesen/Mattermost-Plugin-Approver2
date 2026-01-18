package playbooks

import (
	"encoding/json"
	"fmt"
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

// ClientInterface defines the interface for interacting with the Playbooks plugin
// This interface is implemented by Client and can be mocked for testing
type ClientInterface interface {
	GetPlaybookRunByChannel(channelID string, requesterUserID string) (*PlaybookRun, error)
}

// Client handles communication with the Mattermost Playbooks plugin API
type Client struct {
	api     plugin.API
	siteURL string
}

// NewClient creates a new Playbooks API client
func NewClient(api plugin.API, siteURL string) *Client {
	return &Client{
		api:     api,
		siteURL: siteURL,
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
// Returns error only for unexpected failures (logged, doesn't block approval creation)
func (c *Client) GetPlaybookRunByChannel(channelID string, requesterUserID string) (*PlaybookRun, error) {
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

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Playbooks API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 404 means no playbook for this channel OR user lacks access - normal case, not an error
	// Playbooks API returns 404 for both "not found" and "no permission" scenarios
	if resp.StatusCode == 404 {
		return nil, nil
	}

	// 401 means authentication failed - graceful degradation if user token creation failed
	if resp.StatusCode == 401 {
		return nil, nil
	}

	// Any other non-200 status is an error
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("playbooks API returned status %d", resp.StatusCode)
	}

	var run PlaybookRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &run, nil
}
