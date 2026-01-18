package playbooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		// Verify authorization header with user token
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer user-test-token", authHeader)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedRun) //nolint:gosec
	}))
	defer server.Close()

	api := &plugintest.API{}
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("user-test-token"), nil)

	client := NewClient(api, server.URL, "bot123")

	run, err := client.GetPlaybookRunByChannel("channel012", "testuser")

	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, expectedRun.ID, run.ID)
	assert.Equal(t, expectedRun.Name, run.Name)
	assert.Equal(t, expectedRun.ChannelID, run.ChannelID)
}

func TestGetPlaybookRunByChannel_NotFound(t *testing.T) {
	// Mock server returning 404 - not a playbook channel or user lacks access
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	api := &plugintest.API{}
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("user-test-token"), nil)

	client := NewClient(api, server.URL, "bot123")

	run, err := client.GetPlaybookRunByChannel("regular-channel", "testuser")

	// 404 should return nil without error (normal case)
	assert.NoError(t, err)
	assert.Nil(t, run)
}

func TestGetPlaybookRunByChannel_Unauthorized(t *testing.T) {
	// Mock server returning 401 Unauthorized (invalid or expired token)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	api := &plugintest.API{}
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("invalid-token"), nil)

	client := NewClient(api, server.URL, "bot123")

	run, err := client.GetPlaybookRunByChannel("channel123", "testuser")

	// 401 should return nil without error (graceful degradation)
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
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("user-test-token"), nil)
	// Story 8.6: Mock LogWarn since errors are now logged (graceful degradation AC6)
	api.On("LogWarn", "Failed to get playbook run",
		"channel_id", "channel123",
		"requester_user_id", "testuser",
		"error", mock.MatchedBy(func(s string) bool {
			return strings.Contains(s, "playbooks API returned status 500")
		}),
		"circuit_state", "closed").Return()

	client := NewClient(api, server.URL, "bot123")

	run, err := client.GetPlaybookRunByChannel("channel123", "testuser")

	// Story 8.6 AC6: No user-visible errors - errors are logged but not returned
	assert.NoError(t, err)
	assert.Nil(t, run)
	api.AssertExpectations(t)
}

func TestGetPlaybookRunByChannel_Timeout(t *testing.T) {
	// Mock server with slow response (> 500ms)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond) // Exceed 500ms timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	api := &plugintest.API{}
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("user-test-token"), nil)
	// Story 8.6: Mock LogWarn since timeouts are now logged (graceful degradation AC6)
	api.On("LogWarn", "Failed to get playbook run",
		"channel_id", "channel123",
		"requester_user_id", "testuser",
		"error", mock.Anything, // Error message contains timeout details
		"circuit_state", "closed").Return()

	client := NewClient(api, server.URL, "bot123")

	run, err := client.GetPlaybookRunByChannel("channel123", "testuser")

	// Story 8.6 AC6: No user-visible errors - timeouts are logged but not returned
	assert.NoError(t, err)
	assert.Nil(t, run)
	api.AssertExpectations(t)
}

func TestGetPlaybookRunByChannel_NetworkError(t *testing.T) {
	// Invalid URL to simulate network error
	api := &plugintest.API{}
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("user-test-token"), nil)
	// Story 8.6: Mock LogWarn since network errors are now logged (graceful degradation AC6)
	api.On("LogWarn", "Failed to get playbook run",
		"channel_id", "channel123",
		"requester_user_id", "testuser",
		"error", mock.Anything, // Error message contains network details
		"circuit_state", "closed").Return()

	client := NewClient(api, "http://invalid-url-that-does-not-exist.local", "bot123")

	run, err := client.GetPlaybookRunByChannel("channel123", "testuser")

	// Story 8.6 AC6: No user-visible errors - network errors are logged but not returned
	assert.NoError(t, err)
	assert.Nil(t, run)
	api.AssertExpectations(t)
}

func TestGetPlaybookRunByChannel_InvalidJSON(t *testing.T) {
	// Mock server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json {{{")) //nolint:gosec
	}))
	defer server.Close()

	api := &plugintest.API{}
	// Mock getUserToken to return a test token
	api.On("KVGet", "user_playbooks_token_testuser").Return([]byte("user-test-token"), nil)
	// Story 8.6: Mock LogWarn since JSON decode errors are now logged (graceful degradation AC6)
	api.On("LogWarn", "Failed to get playbook run",
		"channel_id", "channel123",
		"requester_user_id", "testuser",
		"error", mock.MatchedBy(func(err string) bool {
			return len(err) > 0 && (len(err) < 100 || err[:20] == "failed to decode res")
		}),
		"circuit_state", "closed").Return()

	client := NewClient(api, server.URL, "bot123")

	run, err := client.GetPlaybookRunByChannel("channel123", "testuser")

	// Story 8.6 AC6: No user-visible errors - JSON errors are logged but not returned
	assert.NoError(t, err)
	assert.Nil(t, run)
	api.AssertExpectations(t)
}

func TestGetUserToken_ExistingToken(t *testing.T) {
	api := &plugintest.API{}
	api.On("KVGet", "user_playbooks_token_user123").Return([]byte("existing-token-abc"), nil)

	client := NewClient(api, "http://localhost", "bot123")

	token, err := client.getUserToken("user123")

	require.NoError(t, err)
	assert.Equal(t, "existing-token-abc", token)
	api.AssertExpectations(t)
}

func TestGetUserToken_CreateNewToken(t *testing.T) {
	api := &plugintest.API{}
	api.On("KVGet", "user_playbooks_token_user123").Return(nil, nil)
	api.On("CreateUserAccessToken", &model.UserAccessToken{
		UserId:      "user123",
		Description: "Approver Plugin - Playbooks API Access",
	}).Return(&model.UserAccessToken{
		Id:    "token-id-456",
		Token: "new-token-xyz",
	}, nil)
	api.On("KVSet", "user_playbooks_token_user123", []byte("new-token-xyz")).Return(nil)

	client := NewClient(api, "http://localhost", "bot123")

	token, err := client.getUserToken("user123")

	require.NoError(t, err)
	assert.Equal(t, "new-token-xyz", token)
	api.AssertExpectations(t)
}

func TestGetUserToken_CreateTokenError(t *testing.T) {
	api := &plugintest.API{}
	api.On("KVGet", "user_playbooks_token_user123").Return(nil, nil)
	api.On("CreateUserAccessToken", &model.UserAccessToken{
		UserId:      "user123",
		Description: "Approver Plugin - Playbooks API Access",
	}).Return(nil, &model.AppError{Message: "permission denied"})

	client := NewClient(api, "http://localhost", "bot123")

	token, err := client.getUserToken("user123")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to create user access token")
	api.AssertExpectations(t)
}

// Story 8.3 / GitHub Issue #2: Tests for PostMessageToPlaybookChannel (using CreatePost with markdown tables)
func TestPostMessageToPlaybookChannel_Success(t *testing.T) {
	api := &plugintest.API{}
	expectedPostID := "post123"

	// Mock CreatePost
	api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
		return post.UserId == "bot123" &&
			post.ChannelId == "channel123" &&
			strings.Contains(post.Message, "Approval Pending")
	})).Return(&model.Post{Id: expectedPostID}, nil)

	client := NewClient(api, "http://localhost", "bot123")

	postID, err := client.PostMessageToPlaybookChannel("channel123", "### ⏳ Approval Pending\n\n| Field | Value |\n|:------|:------|\n| **Request ID** | A-TEST |")

	require.NoError(t, err)
	assert.Equal(t, expectedPostID, postID)
	api.AssertExpectations(t)
}

func TestPostMessageToPlaybookChannel_APIError(t *testing.T) {
	api := &plugintest.API{}

	// Mock CreatePost to fail
	api.On("CreatePost", mock.Anything).Return(nil, &model.AppError{Message: "API error"})
	// Story 8.6: Mock LogWarn since errors are now logged (graceful degradation AC6)
	api.On("LogWarn", "Failed to post message to playbook channel",
		"channel_id", "channel123",
		"error", mock.Anything,
		"circuit_state", "closed").Return()

	client := NewClient(api, "http://localhost", "bot123")

	postID, err := client.PostMessageToPlaybookChannel("channel123", "Test message")

	// Story 8.6 AC6: No user-visible errors - errors are logged but not returned
	assert.NoError(t, err)
	assert.Empty(t, postID)
	api.AssertExpectations(t)
}

// Story 8.6 Code Review: Integration test for circuit breaker + metrics
func TestCircuitBreakerMetricsIntegration(t *testing.T) {
	// Test that circuit breaker state is properly tracked in metrics
	// This ensures the integration between these components works correctly

	// Create mock server that always fails
	failureCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failureCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	api := &plugintest.API{}
	api.On("KVGet", "user_playbooks_token_user123").Return([]byte("test-token"), nil)
	// Mock all log calls with flexible arg counts
	api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	client := NewClient(api, server.URL, "bot123")

	// Make 5 calls to trigger circuit breaker (threshold = 5)
	for i := 0; i < 5; i++ {
		_, _ = client.GetPlaybookRunByChannel("channel123", "user123")
	}

	// Verify metrics recorded all 5 failures
	metrics := client.metrics.GetSnapshot()
	assert.Equal(t, int64(5), metrics.DetectionCalls)
	assert.Equal(t, int64(0), metrics.DetectionSuccess)
	assert.Equal(t, int64(5), metrics.DetectionFailed)

	// Verify circuit breaker state is Open in metrics
	assert.Equal(t, CircuitOpen, metrics.CircuitBreakerState)
	assert.Equal(t, int64(1), metrics.CircuitBreakerOpens)

	// Verify circuit breaker actually opened (next call should not hit server)
	currentFailureCount := failureCount
	_, _ = client.GetPlaybookRunByChannel("channel123", "user123")
	assert.Equal(t, currentFailureCount, failureCount, "Circuit breaker should prevent API call")

	// Verify metrics still show circuit as open
	metricsAfter := client.metrics.GetSnapshot()
	assert.Equal(t, CircuitOpen, metricsAfter.CircuitBreakerState)
}
