package playbooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

// Client handles communication with the Mattermost Playbooks plugin API
type Client struct {
	api      plugin.API
	siteURL  string
	botToken string
}

// NewClient creates a new Playbooks API client
func NewClient(api plugin.API, siteURL, botToken string) *Client {
	return &Client{
		api:      api,
		siteURL:  siteURL,
		botToken: botToken,
	}
}

// GetPlaybookRunByChannel retrieves playbook run information for a given channel
// Returns nil if no playbook is associated with the channel (404)
// Returns error only for unexpected failures (logged, doesn't block approval creation)
func (c *Client) GetPlaybookRunByChannel(channelID string) (*PlaybookRun, error) {
	url := fmt.Sprintf("%s/plugins/playbooks/api/v0/runs/channel/%s",
		c.siteURL, channelID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.botToken)

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Playbooks API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 404 means no playbook for this channel - normal case, not an error
	if resp.StatusCode == 404 {
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
