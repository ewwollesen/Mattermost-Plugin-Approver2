# Mattermost Interactive Buttons Solution

## Problem
Custom post types (`post.Type = "custom_*"`) with PostAction buttons that have `Integration` callbacks don't work in Mattermost. The security layer strips the integration field, so buttons render but clicking them does nothing.

## Solution: Use Regular Posts with SlackAttachments

Instead of custom post types, use **regular posts** with `SlackAttachment` and `PostAction` buttons. This gives you:
- ✅ Working interactive buttons with callbacks
- ✅ Automatic timezone-aware timestamp rendering
- ✅ Works on webapp, desktop, and mobile
- ❌ Less control over custom rendering (limited to SlackAttachment styling)

## Implementation Example

### Server-Side (Go)

```go
func (p *Plugin) postApprovalRequest(channelID, userID, requestID, requestDetails string, timestamp int64) error {
    // Create interactive buttons
    slackAttachment := model.SlackAttachment{
        Title:    "Approval Request",
        Text:     requestDetails,
        Timestamp: timestamp, // Mattermost auto-converts to user's timezone!
        Actions: []*model.PostAction{
            {
                Id:    "approve_" + requestID,
                Name:  "Approve",
                Type:  model.PostActionTypeButton,
                Style: "primary", // Green button
                Integration: &model.PostActionIntegration{
                    URL: fmt.Sprintf("/plugins/yourplugin/api/v1/approve"),
                    Context: map[string]interface{}{
                        "request_id": requestID,
                        "action":     "approve",
                    },
                },
            },
            {
                Id:    "deny_" + requestID,
                Name:  "Deny",
                Type:  model.PostActionTypeButton,
                Style: "danger", // Red button
                Integration: &model.PostActionIntegration{
                    URL: fmt.Sprintf("/plugins/yourplugin/api/v1/deny"),
                    Context: map[string]interface{}{
                        "request_id": requestID,
                        "action":     "deny",
                    },
                },
            },
        },
    }

    // Create post WITHOUT custom type
    post := &model.Post{
        UserId:    p.botUserID,
        ChannelId: channelID,
        RootId:    "", // or parent post ID if threading
        Message:   "New approval request", // Optional top-level message
        // IMPORTANT: NO Type field - this makes it a regular post
        Props: map[string]interface{}{
            "attachments": []*model.SlackAttachment{&slackAttachment},
            "request_id":  requestID, // Store for later reference
        },
    }

    // Create the post
    createdPost, err := p.API.CreatePost(post)
    if err != nil {
        return err
    }

    // Store the post ID for later updates
    p.storeRequestPostID(requestID, createdPost.Id)

    return nil
}
```

### Button Click Handler

```go
func (p *Plugin) handleApprovalAction(w http.ResponseWriter, r *http.Request) {
    // Parse the button click request
    var request *model.PostActionIntegrationRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Extract context data
    requestID := request.Context["request_id"].(string)
    action := request.Context["action"].(string)
    userID := r.Header.Get("Mattermost-User-Id")

    // Validate user permissions
    if !p.userCanApprove(userID, requestID) {
        p.respondWithError(w, "You don't have permission to approve this request")
        return
    }

    // Process the action
    if action == "approve" {
        p.approveRequest(requestID, userID)
    } else {
        p.denyRequest(requestID, userID)
    }

    // Update the post to show the result
    updatedAttachment := model.SlackAttachment{
        Title: "Approval Request",
        Text:  fmt.Sprintf("✅ %s by @%s", strings.Title(action)+"d", p.getUsername(userID)),
        Color: "#00FF00", // Green for approved
    }

    post := &model.Post{
        Id:        request.PostId,
        ChannelId: request.ChannelId,
        UserId:    p.botUserID,
        Props: map[string]interface{}{
            "attachments": []*model.SlackAttachment{&updatedAttachment},
        },
    }

    p.API.UpdatePost(post)

    // Respond to Mattermost
    response := &model.PostActionIntegrationResponse{
        Update: post,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### HTTP Route Registration

```go
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
    router := mux.NewRouter()

    router.HandleFunc("/api/v1/approve", p.handleApprovalAction).Methods("POST")
    router.HandleFunc("/api/v1/deny", p.handleApprovalAction).Methods("POST")

    router.ServeHTTP(w, r)
}
```

## Key Points

1. **Don't set `post.Type`** - Leave it empty/unset to make it a regular post
2. **Use `SlackAttachment.Actions`** - This is where your buttons go
3. **Set `Integration.URL`** - This endpoint gets called when button is clicked
4. **Use `Integration.Context`** - Pass custom data with the button click
5. **Timestamp auto-converts** - Set `SlackAttachment.Timestamp` as Unix timestamp (seconds), Mattermost handles timezone conversion
6. **Update post after action** - Return `PostActionIntegrationResponse` with updated post

## Styling Options

```go
// Button styles
Style: "default"  // Gray button
Style: "primary"  // Blue/green button (brand color)
Style: "danger"   // Red button

// Attachment colors
Color: "#00FF00"  // Green
Color: "#FF0000"  // Red
Color: "#0000FF"  // Blue
```

## Testing Checklist

- [ ] Buttons appear on the post
- [ ] Clicking button triggers your endpoint
- [ ] Context data arrives correctly
- [ ] Post updates after button click
- [ ] Timestamp shows in user's local timezone
- [ ] Works on mobile app (not just webapp)

## Reference

This pattern is used successfully in the Mattermost Zoom plugin:
- File: `server/http.go`
- Function: `askPreferenceForMeeting()` (lines 488-546)
- Shows working interactive buttons with regular posts

## Migration from Custom Post Type

If you're currently using a custom post type:

**Before:**
```go
post := &model.Post{
    Type: "custom_myapp",  // ❌ This breaks button integration
    Props: map[string]interface{}{
        "attachments": attachments,
    },
}
```

**After:**
```go
post := &model.Post{
    // ✅ No Type field = regular post with working buttons
    Props: map[string]interface{}{
        "attachments": attachments,
    },
}
```

You lose custom React rendering, but you gain working buttons and native timezone support.
