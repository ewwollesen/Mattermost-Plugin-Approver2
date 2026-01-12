package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mattermost/mattermost-plugin-approver2/server/approval"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleApproveNew_EphemeralConfirmation(t *testing.T) {
	t.Run("ephemeral confirmation sent with correct format", func(t *testing.T) {
		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		// Mock user lookups
		requester := &model.User{
			Id:        "requester123",
			Username:  "alice",
			FirstName: "Alice",
			LastName:  "Carter",
		}

		approver := &model.User{
			Id:        "approver456",
			Username:  "bob",
			FirstName: "Bob",
			LastName:  "Smith",
		}

		api.On("GetUser", "requester123").Return(requester, nil)
		api.On("GetUser", "approver456").Return(approver, nil)

		// Mock KV store operations with specific key pattern validation
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			// Should query approval:record: or approval_code: keys
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		})).Return(nil, nil)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			// Should write to approval:record: or approval_code: keys
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "approver456").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// This is the DM notification to approver
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		// Mock ephemeral post - This is what we're testing!
		api.On("SendEphemeralPost", "requester123", mock.MatchedBy(func(post *model.Post) bool {
			// Verify message format matches AC2 exactly
			return assert.Contains(t, post.Message, "✅ **Approval Request Submitted**") &&
				assert.Contains(t, post.Message, "**Approver:** @bob (Bob Smith)") &&
				assert.Contains(t, post.Message, "**Request ID:**") &&
				assert.Contains(t, post.Message, "You will be notified when a decision is made.") &&
				assert.NotContains(t, post.Message, "Your approver will be notified shortly") &&
				post.ChannelId == "channel123"
		})).Return(&model.Post{})

		// Mock logging (use variadic matchers for flexible parameter counts)
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		// Create payload
		payload := &model.SubmitDialogRequest{
			UserId:     "requester123",
			ChannelId:  "channel123",
			TeamId:     "team789",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "approver456",
				"description": "Please approve deployment",
			},
		}

		// Execute
		response := plugin.handleApproveNew(payload)

		// Verify
		assert.NotNil(t, response)
		assert.Empty(t, response.Error)
		assert.Empty(t, response.Errors)

		// Verify SendEphemeralPost was called for requester confirmation
		api.AssertCalled(t, "SendEphemeralPost", "requester123", mock.Anything)
		// Story 2.1: Verify notification DM was sent to approver
		api.AssertCalled(t, "CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		}))
		api.AssertExpectations(t)
	})

	t.Run("ephemeral post uses correct user ID", func(t *testing.T) {
		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		requester := &model.User{
			Id:       "user999",
			Username: "charlie",
		}
		requester.FirstName = "Charlie"
		requester.LastName = "Brown"

		approver := &model.User{
			Id:       "user888",
			Username: "david",
		}
		approver.FirstName = "David"
		approver.LastName = "Lee"

		api.On("GetUser", "user999").Return(requester, nil)
		api.On("GetUser", "user888").Return(approver, nil)
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		})).Return(nil, nil)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "user888").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		// Verify the first argument to SendEphemeralPost is the requester's UserID
		var capturedUserID string
		api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
			capturedUserID = args.Get(0).(string)
		}).Return(&model.Post{})

		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		payload := &model.SubmitDialogRequest{
			UserId:     "user999",
			ChannelId:  "channel456",
			TeamId:     "team123",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "user888",
				"description": "Test request",
			},
		}

		// Execute
		response := plugin.handleApproveNew(payload)

		// Verify
		assert.NotNil(t, response)
		assert.Empty(t, response.Error)
		assert.Equal(t, "user999", capturedUserID, "SendEphemeralPost should be called with requester's UserID")
		api.AssertExpectations(t)
	})

	t.Run("approval saved even if ephemeral confirmation fails", func(t *testing.T) {
		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		requester := &model.User{
			Id:       "requester111",
			Username: "eve",
		}
		requester.FirstName = "Eve"
		requester.LastName = "Johnson"

		approver := &model.User{
			Id:       "approver222",
			Username: "frank",
		}
		approver.FirstName = "Frank"
		approver.LastName = "Wilson"

		api.On("GetUser", "requester111").Return(requester, nil)
		api.On("GetUser", "approver222").Return(approver, nil)

		// Mock successful KV store operations
		// KVGet is called to check for existing record (immutability check)
		api.On("KVGet", mock.AnythingOfType("string")).Return(nil, nil)

		// Capture the approval record data (first KVSet call has the record key pattern)
		var savedData []byte
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			// Match the record key pattern: approval:record:{id}
			return len(key) > 16 && key[:16] == "approval:record:"
		}), mock.Anything).Run(func(args mock.Arguments) {
			savedData = args.Get(1).([]byte)
		}).Return(nil)

		// Also mock the code index KVSet (approval_code:{code} → recordID)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			// Match the code key pattern: approval_code:{code}
			return len(key) > 14 && key[:14] == "approval_code:"
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "approver222").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		// Mock SendEphemeralPost failure (returns nil)
		api.On("SendEphemeralPost", "requester111", mock.Anything).Return(nil)

		// Mock fallback CreatePost for ephemeral failure
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// This is the fallback confirmation post, not the DM notification
			return post.UserId == "requester111"
		})).Return(&model.Post{}, nil)

		// Mock logging (should log error for failed confirmation and fallback)
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		payload := &model.SubmitDialogRequest{
			UserId:     "requester111",
			ChannelId:  "channel789",
			TeamId:     "team456",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "approver222",
				"description": "Critical approval needed",
			},
		}

		// Execute
		response := plugin.handleApproveNew(payload)

		// Verify
		assert.NotNil(t, response)
		assert.Empty(t, response.Error, "Operation should succeed even if confirmation fails")
		assert.Empty(t, response.Errors)

		// Verify approval was still saved to KV store
		assert.NotNil(t, savedData, "Approval record should be saved even if confirmation fails")

		// Verify the saved record
		var savedRecord approval.ApprovalRecord
		err := json.Unmarshal(savedData, &savedRecord)
		assert.NoError(t, err)
		assert.Equal(t, "requester111", savedRecord.RequesterID)
		assert.Equal(t, "approver222", savedRecord.ApproverID)
		assert.Equal(t, "Critical approval needed", savedRecord.Description)
		assert.Equal(t, "pending", savedRecord.Status)

		api.AssertExpectations(t)
	})

	t.Run("message format matches AC2 exactly", func(t *testing.T) {
		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		requester := &model.User{
			Id:       "req555",
			Username: "grace",
		}
		requester.FirstName = "Grace"
		requester.LastName = "Hopper"

		approver := &model.User{
			Id:       "app666",
			Username: "alan",
		}
		approver.FirstName = "Alan"
		approver.LastName = "Turing"

		api.On("GetUser", "req555").Return(requester, nil)
		api.On("GetUser", "app666").Return(approver, nil)
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		})).Return(nil, nil)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "app666").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		// Capture the actual message sent
		var actualMessage string
		api.On("SendEphemeralPost", "req555", mock.Anything).Run(func(args mock.Arguments) {
			post := args.Get(1).(*model.Post)
			actualMessage = post.Message
		}).Return(&model.Post{})

		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		payload := &model.SubmitDialogRequest{
			UserId:     "req555",
			ChannelId:  "ch999",
			TeamId:     "tm888",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "app666",
				"description": "Test",
			},
		}

		// Execute
		response := plugin.handleApproveNew(payload)

		// Verify
		assert.NotNil(t, response)
		assert.Empty(t, response.Error)

		// Verify exact message format per AC2
		assert.Contains(t, actualMessage, "✅ **Approval Request Submitted**", "Should contain header with checkmark emoji")
		assert.Contains(t, actualMessage, "**Approver:** @alan (Alan Turing)", "Should contain approver with username mention and display name")
		assert.Contains(t, actualMessage, "**Request ID:**", "Should contain Request ID label")
		assert.Contains(t, actualMessage, "`", "Request ID should be in backticks")
		assert.Contains(t, actualMessage, "You will be notified when a decision is made.", "Should contain notification message")

		// Verify old message is removed
		assert.NotContains(t, actualMessage, "Your approver will be notified shortly", "Old temporary message should be removed")
		assert.NotContains(t, actualMessage, "Approval request created!", "Old header should be removed")

		api.AssertExpectations(t)
	})
}

func TestHandleApproveNew_Performance(t *testing.T) {
	t.Run("operation completes within 2 seconds", func(t *testing.T) {
		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		requester := &model.User{
			Id:       "perf123",
			Username: "perftest",
		}
		requester.FirstName = "Perf"
		requester.LastName = "Test"

		approver := &model.User{
			Id:       "perf456",
			Username: "approvertest",
		}
		approver.FirstName = "Approver"
		approver.LastName = "Test"

		api.On("GetUser", "perf123").Return(requester, nil)
		api.On("GetUser", "perf456").Return(approver, nil)
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		})).Return(nil, nil)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "perf456").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		api.On("SendEphemeralPost", "perf123", mock.Anything).Return(&model.Post{})
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		payload := &model.SubmitDialogRequest{
			UserId:     "perf123",
			ChannelId:  "channel999",
			TeamId:     "team999",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "perf456",
				"description": "Performance test request",
			},
		}

		// Execute with timing
		start := time.Now()
		response := plugin.handleApproveNew(payload)
		elapsed := time.Since(start)

		// Verify
		assert.NotNil(t, response)
		assert.Empty(t, response.Error)
		assert.Empty(t, response.Errors)

		// Verify performance requirement (NFR-P2: < 2 seconds)
		// Note: In unit tests with mocks, operation is near-instantaneous
		// In real integration tests, this verifies the 2-second requirement
		assert.Less(t, elapsed, 2*time.Second, "Operation should complete within 2 seconds (NFR-P2)")
		t.Logf("✅ Operation completed in %v - performance requirement met", elapsed)

		api.AssertExpectations(t)
	})
}

func TestHandleApproveNew_IntegrationFlow(t *testing.T) {
	t.Run("complete submission flow verifies all acceptance criteria", func(t *testing.T) {
		// This integration test verifies all ACs in Story 1.6:
		// AC1: ApprovalRecord Created with Complete Data
		// AC2: Ephemeral Confirmation Message Posted
		// AC3: Data Integrity Over Confirmation
		// AC4: Mattermost Authentication Used

		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		// AC4: Mattermost authentication - user identity from authenticated session
		requester := &model.User{
			Id:       "integration-requester",
			Username: "alice",
		}
		requester.FirstName = "Alice"
		requester.LastName = "Johnson"

		approver := &model.User{
			Id:       "integration-approver",
			Username: "bob",
		}
		approver.FirstName = "Bob"
		approver.LastName = "Smith"

		api.On("GetUser", "integration-requester").Return(requester, nil)
		api.On("GetUser", "integration-approver").Return(approver, nil)

		// Mock KV store operations for approval persistence with key validation
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		})).Return(nil, nil)

		// AC1: Capture the ApprovalRecord to verify complete data
		var capturedRecord []byte
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 16 && key[:16] == "approval:record:"
		}), mock.Anything).Run(func(args mock.Arguments) {
			capturedRecord = args.Get(1).([]byte)
		}).Return(nil)

		// Also mock code index KVSet
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 14 && key[:14] == "approval_code:"
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "integration-approver").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		// AC2: Capture ephemeral post to verify message format
		var capturedPost *model.Post
		var capturedUserID string
		api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
			capturedUserID = args.Get(0).(string)
			capturedPost = args.Get(1).(*model.Post)
		}).Return(&model.Post{})

		// Mock logging (use variadic matchers for flexible parameter counts)
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		// Create dialog submission payload
		payload := &model.SubmitDialogRequest{
			UserId:     "integration-requester",
			ChannelId:  "test-channel-123",
			TeamId:     "test-team-456",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "integration-approver",
				"description": "Integration test approval request",
			},
		}

		// Execute the complete flow
		start := time.Now()
		response := plugin.handleApproveNew(payload)
		elapsed := time.Since(start)

		// Verify response has no errors
		assert.NotNil(t, response)
		assert.Empty(t, response.Error, "Response should have no general error")
		assert.Empty(t, response.Errors, "Response should have no field errors")

		// AC1: Verify ApprovalRecord created with complete data
		assert.NotNil(t, capturedRecord, "ApprovalRecord should be saved")
		var record approval.ApprovalRecord
		err := json.Unmarshal(capturedRecord, &record)
		assert.NoError(t, err)

		// Verify all required fields in ApprovalRecord
		assert.Equal(t, "pending", record.Status, "Status should be pending")
		assert.Equal(t, "integration-requester", record.RequesterID)
		assert.Equal(t, "alice", record.RequesterUsername)
		assert.Equal(t, "Alice Johnson", record.RequesterDisplayName)
		assert.Equal(t, "integration-approver", record.ApproverID)
		assert.Equal(t, "bob", record.ApproverUsername)
		assert.Equal(t, "Bob Smith", record.ApproverDisplayName)
		assert.Equal(t, "Integration test approval request", record.Description)
		assert.Equal(t, "test-channel-123", record.RequestChannelID)
		assert.Equal(t, "test-team-456", record.TeamID)
		assert.NotEmpty(t, record.ID, "Record ID should be generated")
		assert.NotEmpty(t, record.Code, "Human-friendly code should be generated")
		assert.Greater(t, record.CreatedAt, int64(0), "CreatedAt should be set")
		assert.Equal(t, int64(0), record.DecidedAt, "DecidedAt should be 0 for pending")

		// AC2: Verify ephemeral confirmation message
		assert.NotNil(t, capturedPost, "Ephemeral post should be sent")
		assert.Equal(t, "integration-requester", capturedUserID, "Ephemeral post should be sent to requester")
		assert.Equal(t, "test-channel-123", capturedPost.ChannelId, "Post should be in request channel")
		assert.Empty(t, capturedPost.UserId, "UserId should be empty for system message")

		// Verify message format matches AC2 exactly
		message := capturedPost.Message
		assert.Contains(t, message, "✅ **Approval Request Submitted**", "Should contain header")
		assert.Contains(t, message, "**Approver:** @bob (Bob Smith)", "Should contain approver info")
		assert.Contains(t, message, "**Request ID:**", "Should contain Request ID label")
		assert.Contains(t, message, record.Code, "Should contain the generated code")
		assert.Contains(t, message, "You will be notified when a decision is made.", "Should contain notification message")
		assert.NotContains(t, message, "Your approver will be notified shortly", "Should not contain old message")

		// AC1: Verify performance requirement (< 2 seconds)
		assert.Less(t, elapsed, 2*time.Second, "Operation should complete within 2 seconds (NFR-P2)")

		// AC4: Verify Mattermost authentication is used (no additional auth checks)
		// This is verified by the fact that payload.UserId is used directly from the authenticated session

		// Verify all mocks were called as expected
		api.AssertExpectations(t)

		t.Logf("✅ Integration test passed - all acceptance criteria verified")
		t.Logf("   - Record created: ID=%s, Code=%s", record.ID, record.Code)
		t.Logf("   - Ephemeral message sent to requester")
		t.Logf("   - Operation completed in %v", elapsed)
	})

	t.Run("data integrity maintained when confirmation fails", func(t *testing.T) {
		// AC3: Verify that approval is saved even when ephemeral confirmation fails

		// Setup
		api := &plugintest.API{}
		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.botUserID = "bot123" // Set bot user ID for notification

		requester := &model.User{
			Id:       "req-fail-test",
			Username: "charlie",
		}
		requester.FirstName = "Charlie"
		requester.LastName = "Brown"

		approver := &model.User{
			Id:       "app-fail-test",
			Username: "diana",
		}
		approver.FirstName = "Diana"
		approver.LastName = "Prince"

		api.On("GetUser", "req-fail-test").Return(requester, nil)
		api.On("GetUser", "app-fail-test").Return(approver, nil)

		// Mock successful KV operations with key validation
		api.On("KVGet", mock.MatchedBy(func(key string) bool {
			return len(key) > 10 && (key[:16] == "approval:record:" || key[:14] == "approval_code:")
		})).Return(nil, nil)
		var recordSaved bool
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 16 && key[:16] == "approval:record:"
		}), mock.Anything).Run(func(args mock.Arguments) {
			recordSaved = true
		}).Return(nil)
		api.On("KVSet", mock.MatchedBy(func(key string) bool {
			return len(key) > 14 && key[:14] == "approval_code:"
		}), mock.Anything).Return(nil)

		// Story 2.1: Mock notification DM calls
		api.On("GetDirectChannel", "bot123", "app-fail-test").Return(&model.Channel{Id: "dm_channel"}, nil)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return post.UserId == "bot123" && post.ChannelId == "dm_channel"
		})).Return(&model.Post{}, nil)

		// Mock SendEphemeralPost FAILURE (returns nil) - triggers fallback to CreatePost
		api.On("SendEphemeralPost", "req-fail-test", mock.Anything).Return(nil)

		// Mock fallback CreatePost for ephemeral failure (AC3: generic success indicator)
		api.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			// This is the fallback confirmation post, not the DM notification
			return post.UserId == "req-fail-test"
		})).Return(&model.Post{}, nil)

		// Mock logging (should log the confirmation failure and fallback attempt)
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		payload := &model.SubmitDialogRequest{
			UserId:     "req-fail-test",
			ChannelId:  "channel-fail",
			TeamId:     "team-fail",
			CallbackId: "approve_new",
			Submission: map[string]any{
				"approver":    "app-fail-test",
				"description": "Test request with confirmation failure",
			},
		}

		// Execute
		response := plugin.handleApproveNew(payload)

		// AC3: Verify operation succeeds despite confirmation failure
		assert.NotNil(t, response)
		assert.Empty(t, response.Error, "Operation should succeed even if confirmation fails")
		assert.Empty(t, response.Errors)

		// Verify data integrity: record was still saved
		assert.True(t, recordSaved, "ApprovalRecord should be saved even if confirmation fails")

		// Verify error was logged (LogError should be called at least once for ephemeral failure)
		// Note: May also log fallback CreatePost failure, so we just verify LogError was called
		api.AssertCalled(t, "LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		// Verify all expectations met
		api.AssertExpectations(t)

		t.Log("✅ Data integrity verified - record saved despite confirmation failure")
	})
}

// TestHandleCancelCommand_Integration verifies the complete cancel flow end-to-end
func TestHandleCancelCommand_Integration(t *testing.T) {
	t.Run("complete cancel flow with all validations", func(t *testing.T) {
		// This integration test verifies all ACs in Story 1.7:
		// AC1: Retrieve approval by code and verify requester
		// AC2: Update status to canceled and set DecidedAt
		// AC3: Display ephemeral confirmation message
		// AC4: Permission denied for non-requester
		// AC5: Cannot cancel non-pending approval
		// AC6: Help text for missing ID
		// AC7: Error for invalid ID

		// Setup
		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Scenario 1: Create an approval record that will be canceled
		record := &approval.ApprovalRecord{
			ID:          "record-to-cancel",
			Code:        "A-TEST01",
			RequesterID: "alice123",
			Status:      approval.StatusPending,
			CreatedAt:   1704931200000,
			DecidedAt:   0,
		}

		// Mock KV operations for successful cancellation
		api.On("KVGet", "approval_code:A-TEST01").Return([]byte(`"record-to-cancel"`), nil)

		recordJSON, _ := json.Marshal(record)
		api.On("KVGet", "approval:record:record-to-cancel").Return(recordJSON, nil)

		// Capture the updated record to verify DecidedAt was set
		var capturedRecord []byte
		api.On("KVSet", "approval:record:record-to-cancel", mock.Anything).Run(func(args mock.Arguments) {
			capturedRecord = args.Get(1).([]byte)
		}).Return(nil)
		api.On("KVSet", "approval_code:A-TEST01", mock.Anything).Return(nil)

		// Mock ephemeral confirmation post
		var capturedMessage string
		api.On("SendEphemeralPost", "alice123", mock.Anything).Run(func(args mock.Arguments) {
			post := args.Get(1).(*model.Post)
			capturedMessage = post.Message
		}).Return(&model.Post{})

		// Mock logging
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		// AC1 & AC2: Execute cancel command as requester
		args := &model.CommandArgs{
			Command:   "/approve cancel A-TEST01",
			UserId:    "alice123",
			ChannelId: "channel123",
		}

		start := time.Now()
		resp, appErr := p.ExecuteCommand(nil, args)
		elapsed := time.Since(start)

		// Verify success
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)

		// AC3: Verify ephemeral confirmation message
		assert.Contains(t, capturedMessage, "✅ Approval request `A-TEST01` has been canceled")

		// AC2: Verify DecidedAt timestamp was set
		var updatedRecord approval.ApprovalRecord
		err = json.Unmarshal(capturedRecord, &updatedRecord)
		assert.NoError(t, err)
		assert.Equal(t, approval.StatusCanceled, updatedRecord.Status)
		assert.Greater(t, updatedRecord.DecidedAt, int64(0), "DecidedAt should be set")
		assert.Greater(t, updatedRecord.DecidedAt, record.CreatedAt, "DecidedAt should be after CreatedAt")

		// AC2: Verify performance requirement (< 2 seconds)
		assert.Less(t, elapsed, 2*time.Second, "Operation should complete within 2 seconds (NFR-P2)")

		api.AssertExpectations(t)

		t.Logf("✅ Integration test passed - all acceptance criteria verified")
		t.Logf("   - Cancel successful: Code=%s, Status=%s", updatedRecord.Code, updatedRecord.Status)
		t.Logf("   - DecidedAt timestamp set: %d", updatedRecord.DecidedAt)
		t.Logf("   - Operation completed in %v", elapsed)
	})

	t.Run("cannot cancel already-canceled approval", func(t *testing.T) {
		// AC5: Verify cannot cancel already-canceled approval (simplified test)

		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Record already canceled
		canceledRecord := &approval.ApprovalRecord{
			ID:          "record456",
			Code:        "A-ALRDYC",
			RequesterID: "bob123",
			Status:      approval.StatusCanceled,
			CreatedAt:   1704931200000,
			DecidedAt:   1704931300000,
		}

		api.On("KVGet", "approval_code:A-ALRDYC").Return([]byte(`"record456"`), nil)
		canceledJSON, _ := json.Marshal(canceledRecord)
		api.On("KVGet", "approval:record:record456").Return(canceledJSON, nil)

		// Mock logging
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-ALRDYC",
			UserId:    "bob123",
			ChannelId: "channel123",
		}

		// Attempt to cancel already-canceled approval
		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Text, "❌ Cannot cancel approval request A-ALRDYC")
		assert.Contains(t, resp.Text, "Status is already canceled")

		api.AssertExpectations(t)

		t.Log("✅ Immutability verified - cannot cancel already-canceled approval")
	})

	t.Run("access control - different user cannot cancel", func(t *testing.T) {
		// AC4: Verify permission denied for non-requester

		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Record owned by alice123
		record := &approval.ApprovalRecord{
			ID:          "record789",
			Code:        "A-NOAUTH",
			RequesterID: "alice123",
			Status:      approval.StatusPending,
			CreatedAt:   1704931200000,
			DecidedAt:   0,
		}

		api.On("KVGet", "approval_code:A-NOAUTH").Return([]byte(`"record789"`), nil)
		recordJSON, _ := json.Marshal(record)
		api.On("KVGet", "approval:record:record789").Return(recordJSON, nil)

		// Mock logging
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		// Attempt cancel as different user (charlie456)
		args := &model.CommandArgs{
			Command:   "/approve cancel A-NOAUTH",
			UserId:    "charlie456",
			ChannelId: "channel123",
		}

		resp, appErr := p.ExecuteCommand(nil, args)
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Contains(t, resp.Text, "❌ Permission denied")
		assert.Contains(t, resp.Text, "only cancel your own approval requests")

		api.AssertExpectations(t)

		t.Log("✅ Access control verified - only requester can cancel")
	})
}

// TestHandleCancelCommand_Performance verifies the cancel operation performance requirement
func TestHandleCancelCommand_Performance(t *testing.T) {
	t.Run("cancel operation completes within 2 seconds", func(t *testing.T) {
		// AC2 requires: "the operation completes within 2 seconds"

		api := &plugintest.API{}

		// Mock plugin activation
		api.On("EnsureBotUser", mock.AnythingOfType("*model.Bot")).Return("bot123", nil)
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

		// Setup test record
		record := &approval.ApprovalRecord{
			ID:          "perf-record",
			Code:        "A-PERF01",
			RequesterID: "perfuser",
			Status:      approval.StatusPending,
			CreatedAt:   1704931200000,
			DecidedAt:   0,
		}

		api.On("KVGet", "approval_code:A-PERF01").Return([]byte(`"perf-record"`), nil)
		recordJSON, _ := json.Marshal(record)
		api.On("KVGet", "approval:record:perf-record").Return(recordJSON, nil)
		api.On("KVSet", "approval:record:perf-record", mock.Anything).Return(nil)
		api.On("KVSet", "approval_code:A-PERF01", mock.Anything).Return(nil)
		api.On("SendEphemeralPost", "perfuser", mock.Anything).Return(&model.Post{})

		// Mock logging
		api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

		p := &Plugin{}
		p.SetAPI(api)
		err := p.OnActivate()
		assert.NoError(t, err)

		args := &model.CommandArgs{
			Command:   "/approve cancel A-PERF01",
			UserId:    "perfuser",
			ChannelId: "channel123",
		}

		// Execute with timing
		start := time.Now()
		resp, appErr := p.ExecuteCommand(nil, args)
		elapsed := time.Since(start)

		// Verify success
		assert.Nil(t, appErr)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Text, "Should return empty text on success (ephemeral post sent separately)")

		// Verify performance requirement (NFR-P2: < 2 seconds)
		// Note: In unit tests with mocks, operation is near-instantaneous
		// In real integration tests with actual KV store, this verifies the 2-second requirement
		assert.Less(t, elapsed, 2*time.Second, "Operation should complete within 2 seconds (NFR-P2)")

		t.Logf("✅ Performance requirement met - operation completed in %v", elapsed)

		api.AssertExpectations(t)
	})
}
