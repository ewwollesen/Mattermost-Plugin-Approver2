package playbooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlaybookRunByChannel_Success(t *testing.T) {
	// Mock server returning 200 OK with playbook run
	expectedRun := &PlaybookRun{
		ID:          "run123",
		Name:        "Test Playbook",
		Description: "Test Description",
		OwnerUserID: "user456",
		TeamID:      "team789",
		ChannelID:   "channel012",
		CreateAt:    1705507200000,
		EndAt:       0,
		PlaybookID:  "playbook345",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify endpoint
		assert.Contains(t, r.URL.Path, "/plugins/playbooks/api/v0/runs/channel/")

		// Verify authorization header
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-token", authHeader)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedRun) //nolint:gosec
	}))
	defer server.Close()

	api := &plugintest.API{}
	client := NewClient(api, server.URL, "test-token")

	run, err := client.GetPlaybookRunByChannel("channel012")

	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, expectedRun.ID, run.ID)
	assert.Equal(t, expectedRun.Name, run.Name)
	assert.Equal(t, expectedRun.ChannelID, run.ChannelID)
}

func TestGetPlaybookRunByChannel_NotFound(t *testing.T) {
	// Mock server returning 404 - not a playbook channel
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	api := &plugintest.API{}
	client := NewClient(api, server.URL, "test-token")

	run, err := client.GetPlaybookRunByChannel("regular-channel")

	// 404 should return nil without error (normal case)
	assert.NoError(t, err)
	assert.Nil(t, run)
}

func TestGetPlaybookRunByChannel_ServerError(t *testing.T) {
	// Mock server returning 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	api := &plugintest.API{}
	client := NewClient(api, server.URL, "test-token")

	run, err := client.GetPlaybookRunByChannel("channel123")

	// 5xx should return error
	assert.Error(t, err)
	assert.Nil(t, run)
	assert.Contains(t, err.Error(), "returned status 500")
}

func TestGetPlaybookRunByChannel_Timeout(t *testing.T) {
	// Mock server with slow response (> 500ms)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond) // Exceed 500ms timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	api := &plugintest.API{}
	client := NewClient(api, server.URL, "test-token")

	run, err := client.GetPlaybookRunByChannel("channel123")

	// Timeout should return error
	assert.Error(t, err)
	assert.Nil(t, run)
}

func TestGetPlaybookRunByChannel_NetworkError(t *testing.T) {
	// Invalid URL to simulate network error
	api := &plugintest.API{}
	client := NewClient(api, "http://invalid-url-that-does-not-exist.local", "test-token")

	run, err := client.GetPlaybookRunByChannel("channel123")

	// Network error should return error
	assert.Error(t, err)
	assert.Nil(t, run)
}

func TestGetPlaybookRunByChannel_InvalidJSON(t *testing.T) {
	// Mock server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json {{{")) //nolint:gosec
	}))
	defer server.Close()

	api := &plugintest.API{}
	client := NewClient(api, server.URL, "test-token")

	run, err := client.GetPlaybookRunByChannel("channel123")

	// JSON decode error should return error
	assert.Error(t, err)
	assert.Nil(t, run)
	assert.Contains(t, err.Error(), "failed to decode")
}
