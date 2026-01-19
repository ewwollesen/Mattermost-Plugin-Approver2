# Matterpoll Plugin: Interactive Voting Buttons - Complete Analysis

## Executive Summary

The Matterpoll plugin successfully implements interactive voting buttons using **custom post types** with working `Integration` callbacks. This proves that custom post types CAN have functional interactive buttons in Mattermost, contrary to some documentation suggesting otherwise.

**Key Finding:** The pattern uses `model.ParseSlackAttachment(post, actions)` which appears to preserve Integration URLs even with custom post types.

---

## 1. POLL CREATION FLOW

### Entry Points

**A. Slash Command** (`/poll "question" "answer1" "answer2"...`)
- **File**: `Matterpoll/server/plugin/command.go`
- **Function**: `ExecuteCommand()` (line 89)
- **Flow**: Parses command → Creates poll → Posts to channel

**Supported Formats:**
```bash
/poll "What's for lunch?" "Pizza" "Tacos" "Salad"
/poll "Deploy to prod?" "Yes" "No" --progress --votes=1
/poll                      # Opens interactive dialog
```

**Settings Flags:**
- `--anonymous` - Hide voter names
- `--anonymous-creator` - Hide poll creator
- `--progress` - Show vote counts in buttons
- `--public-add-option` - Allow anyone to add options
- `--votes=N` - Limit votes per user

**B. Interactive Dialog** (Fallback)
- **File**: `Matterpoll/server/plugin/command.go` (lines 110-114)
- Opens Mattermost dialog when `/poll` has no arguments
- User fills out form → Submits → Creates poll

---

## 2. POLL POST CREATION WITH VOTING BUTTONS

### Core Function: `ExecuteCommand()` → Poll Creation

**File**: `Matterpoll/server/plugin/command.go` (lines 116-227)

```go
func (p *MatterpollPlugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
    split := strings.Fields(args.Command)
    command := split[0]

    if command != "/poll" {
        return &model.CommandResponse{}, nil
    }

    // Parse command arguments
    pollRequest, err := poll.NewPollRequest(split[1:])

    // Create Poll object
    displayName := p.ConvertUserIDToDisplayName(args.UserId)
    newPoll, errMsg := poll.NewPoll(pollRequest, displayName, args.UserId)

    // Save poll to database
    if err := p.Store.Poll().Insert(newPoll); err != nil {
        return nil, err
    }

    // CRITICAL: Generate post actions with Integration URLs
    actions := newPoll.ToPostActions(p.bundle, root.Manifest.Id, displayName)

    // Create post with CUSTOM TYPE
    post := &model.Post{
        UserId:    p.botUserID,
        ChannelId: args.ChannelId,
        RootId:    args.RootId,
        Type:      MatterpollPostType,  // "custom_matterpoll"
        Props: map[string]interface{}{
            "poll_id": newPoll.ID,
        },
    }

    // CRITICAL: Use ParseSlackAttachment (not direct Props assignment)
    model.ParseSlackAttachment(post, actions)

    // Add interactive card for progress mode
    if newPoll.Settings.Progress {
        post.AddProp("card", newPoll.ToCard(p.bundle, p.ConvertUserIDToDisplayName))
    }

    // Create the post
    rPost, appErr := p.API.CreatePost(post)
    if appErr != nil {
        return nil, appErr
    }

    // Save post ID for later updates
    newPoll.PostID = rPost.Id
    if err := p.Store.Poll().Update(nil, newPoll); err != nil {
        p.API.LogWarn("failed to update poll", "error", err.Error())
    }

    return &model.CommandResponse{}, nil
}
```

**Key Constants** (`Matterpoll/server/plugin/plugin.go`, line 58):
```go
const MatterpollPostType = "custom_matterpoll"
```

---

## 3. GENERATING INTERACTIVE BUTTONS

### Function: `ToPostActions()`

**File**: `Matterpoll/server/poll/transform.go` (lines 67-159)

This is the **critical function** that creates voting buttons with Integration URLs:

```go
func (p *Poll) ToPostActions(bundle *utils.Bundle, pluginID, authorName string) []*model.SlackAttachment {
    actions := []*model.PostAction{}

    // Create voting button for each answer option
    for i, o := range p.AnswerOptions {
        answer := o.Answer

        // Show vote counts if progress mode enabled
        if p.Settings.Progress {
            answer = fmt.Sprintf("%s (%d)", answer, len(o.Voter))
        }

        // CRITICAL: PostAction with Integration URL
        actions = append(actions, &model.PostAction{
            Id:    fmt.Sprintf("vote%v", i),
            Name:  answer,  // Button text
            Type:  model.PostActionTypeButton,
            Style: "default",
            Integration: &model.PostActionIntegration{
                URL: fmt.Sprintf("/plugins/%s/api/v1/polls/%s/vote/%v",
                    pluginID, p.ID, i),
                // Note: No Context map - uses URL path parameters instead
            },
        })
    }

    // Add "Reset Votes" button (only for creator)
    actions = append(actions, &model.PostAction{
        Id:    "resetVotes",
        Name:  bundle.LocalizeDefaultMessage(&i18n.LocalizeConfig{
            DefaultMessage: commandResetPollName,
        }),
        Type:  model.PostActionTypeButton,
        Style: "default",
        Integration: &model.PostActionIntegration{
            URL: fmt.Sprintf("/plugins/%s/api/v1/polls/%s/votes/reset",
                pluginID, p.ID),
        },
    })

    // Add "Add Option" button (if public-add-option enabled)
    if p.Settings.PublicAddOption {
        actions = append(actions, &model.PostAction{
            Id:    "addOption",
            Name:  bundle.LocalizeDefaultMessage(&i18n.LocalizeConfig{
                DefaultMessage: commandAddOptionName,
            }),
            Type:  model.PostActionTypeButton,
            Style: "default",
            Integration: &model.PostActionIntegration{
                URL: fmt.Sprintf("/plugins/%s/api/v1/polls/%s/option/add/request",
                    pluginID, p.ID),
            },
        })
    }

    // Add "End Poll" button (only for creator/admins)
    actions = append(actions, &model.PostAction{
        Id:    "endPoll",
        Name:  bundle.LocalizeDefaultMessage(&i18n.LocalizeConfig{
            DefaultMessage: commandEndPollName,
        }),
        Type:  model.PostActionTypeButton,
        Style: "danger",
        Integration: &model.PostActionIntegration{
            URL: fmt.Sprintf("/plugins/%s/api/v1/polls/%s/end",
                pluginID, p.ID),
        },
    })

    // Add "Delete Poll" button (only for creator/admins)
    actions = append(actions, &model.PostAction{
        Id:    "deletePoll",
        Name:  bundle.LocalizeDefaultMessage(&i18n.LocalizeConfig{
            DefaultMessage: commandDeletePollName,
        }),
        Type:  model.PostActionTypeButton,
        Style: "danger",
        Integration: &model.PostActionIntegration{
            URL: fmt.Sprintf("/plugins/%s/api/v1/polls/%s/delete",
                pluginID, p.ID),
        },
    })

    // Build poll display text
    var textBuffer bytes.Buffer
    for i, option := range p.AnswerOptions {
        textBuffer.WriteString(fmt.Sprintf("%v. %s", i+1, option.Answer))
        if p.Settings.Progress {
            textBuffer.WriteString(fmt.Sprintf(" (%d votes)", len(option.Voter)))
        }
        textBuffer.WriteString("\n")
    }

    // Create SlackAttachment with all buttons
    return []*model.SlackAttachment{{
        AuthorName: authorName,
        Title:      p.Question,
        Text:       textBuffer.String(),
        Actions:    actions,
    }}
}
```

**Pattern Note:** Matterpoll uses **URL path parameters** for state (poll ID, option number), while Zoom uses **Context maps**. Both work!

---

## 4. BUTTON CLICK HANDLING - VOTE PROCESSING

### API Router Setup

**File**: `Matterpoll/server/plugin/api.go` (lines 75-95)

```go
func (p *MatterpollPlugin) initializeAPI() {
    p.router = mux.NewRouter()
    apiRouter := p.router.PathPrefix("/api/v1").Subrouter()

    // Poll creation
    apiRouter.HandleFunc("/polls/create", p.handleCreatePoll).Methods(http.MethodPost)

    // Voting endpoints
    apiRouter.HandleFunc("/polls/{id}/vote/{optionNumber:[0-9]+}",
        p.handlePostActionIntegrationRequest(p.handleVote)).Methods(http.MethodPost)

    apiRouter.HandleFunc("/polls/{id}/votes/reset",
        p.handlePostActionIntegrationRequest(p.handleResetVotes)).Methods(http.MethodPost)

    // Option management
    apiRouter.HandleFunc("/polls/{id}/option/add/request",
        p.handlePostActionIntegrationRequest(p.handleAddOption)).Methods(http.MethodPost)

    apiRouter.HandleFunc("/polls/{id}/option/add",
        p.handleAddOptionConfirm).Methods(http.MethodPost)

    // Poll lifecycle
    apiRouter.HandleFunc("/polls/{id}/end",
        p.handlePostActionIntegrationRequest(p.handleEndPoll)).Methods(http.MethodPost)

    apiRouter.HandleFunc("/polls/{id}/end/confirm",
        p.handleEndPollConfirm).Methods(http.MethodPost)

    apiRouter.HandleFunc("/polls/{id}/delete",
        p.handlePostActionIntegrationRequest(p.handleDeletePoll)).Methods(http.MethodPost)

    apiRouter.HandleFunc("/polls/{id}/delete/confirm",
        p.handleDeletePollConfirm).Methods(http.MethodPost)

    // Metadata (for webapp)
    apiRouter.HandleFunc("/polls/{id}/metadata",
        p.handlePollMetadata).Methods(http.MethodGet)
}
```

### Vote Handler Function

**File**: `Matterpoll/server/plugin/api.go` (lines 360-425)

```go
func (p *MatterpollPlugin) handleVote(vars map[string]string,
    request *model.PostActionIntegrationRequest) (*i18n.LocalizeConfig, *model.Post, error) {

    // Extract poll ID and option from URL path
    pollID := vars["id"]
    optionNumber, err := strconv.Atoi(vars["optionNumber"])
    if err != nil {
        return commandErrorGeneric, nil, err
    }

    // Get user who clicked the button
    userID := request.UserId

    // Load poll from database
    poll, err := p.Store.Poll().Get(pollID)
    if err != nil {
        return commandErrorGeneric, nil, err
    }

    // Validate poll state
    if poll.PostID != request.PostId {
        return commandErrorGeneric, nil, errors.New("poll post ID mismatch")
    }

    // Check if poll has ended
    if poll.State == poll.StateClosed {
        return responseVotePollEnded, nil, nil
    }

    // VOTE LOGIC: Update poll with user's vote
    previouslyVoted := poll.HasVoted(userID)

    msg, err := poll.UpdateVote(userID, optionNumber)
    if err != nil {
        p.API.LogWarn("Failed to update vote", "error", err.Error())
        return msg, nil, err
    }

    // Save updated poll to database
    prev := poll.Copy()
    if err = p.Store.Poll().Update(prev, poll); err != nil {
        p.API.LogWarn("failed to update poll", "error", err.Error())
        return commandErrorGeneric, nil, err
    }

    // Publish WebSocket event for real-time UI updates
    go p.publishPollMetadata(poll, userID)

    // Send DM if first vote and anonymous mode
    if !previouslyVoted && poll.Settings.Anonymous {
        p.SendDirectMessage(userID,
            p.bundle.LocalizeDefaultMessage(&i18n.LocalizeConfig{
                DefaultMessage: responseFirstVote,
            }))
    }

    // CRITICAL: Create updated post with new vote counts
    post := &model.Post{}
    displayName := p.ConvertUserIDToDisplayName(poll.Creator)

    // Regenerate actions with updated vote counts
    model.ParseSlackAttachment(post,
        poll.ToPostActions(p.bundle, p.manifest.Id, displayName))

    // Update progress card if enabled
    if poll.Settings.Progress {
        post.AddProp("card",
            poll.ToCard(p.bundle, p.ConvertUserIDToDisplayName))
    }

    post.AddProp("poll_id", poll.ID)

    // Return success message and updated post
    return &i18n.LocalizeConfig{DefaultMessage: responseVoteCounted}, post, nil
}
```

### PostAction Integration Request Wrapper

**File**: `Matterpoll/server/plugin/api.go` (lines 97-135)

```go
type postActionHandler func(map[string]string, *model.PostActionIntegrationRequest) (*i18n.LocalizeConfig, *model.Post, error)

func (p *MatterpollPlugin) handlePostActionIntegrationRequest(handler postActionHandler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Decode Mattermost's PostActionIntegrationRequest
        var request *model.PostActionIntegrationRequest
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            p.API.LogWarn("failed to decode PostActionIntegrationRequest", "error", err.Error())
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Validate user ID from header matches request
        userID := r.Header.Get("Mattermost-User-Id")
        if userID != request.UserId {
            http.Error(w, "user ID mismatch", http.StatusUnauthorized)
            return
        }

        // Extract URL path variables (poll ID, option number, etc.)
        vars := mux.Vars(r)

        // Call the specific handler (handleVote, handleEndPoll, etc.)
        msg, post, err := handler(vars, request)

        if err != nil {
            p.API.LogWarn("failed to handle PostActionIntegrationRequest",
                "error", err.Error())
        }

        // Build response
        response := &model.PostActionIntegrationResponse{}

        if msg != nil {
            response.EphemeralText = p.bundle.LocalizeDefaultMessage(msg)
        }

        if post != nil {
            // Update the original post with new content
            response.Update = post
        }

        // Send JSON response back to Mattermost
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            p.API.LogWarn("failed to encode PostActionIntegrationResponse",
                "error", err.Error())
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}
```

---

## 5. VOTE TRACKING AND PERSISTENCE

### Poll Data Structure

**File**: `Matterpoll/server/poll/poll.go` (lines 13-52)

```go
type Poll struct {
    ID            string
    PostID        string     // Link to Mattermost post
    Question      string
    AnswerOptions []*AnswerOption
    Creator       string     // User ID
    Settings      Settings
    State         State      // Open or Closed
}

type AnswerOption struct {
    Answer string
    Voter  []string  // Array of user IDs who voted for this option
}

type Settings struct {
    Anonymous         bool  // Hide voter names
    AnonymousCreator  bool  // Hide poll creator
    Progress          bool  // Show vote counts
    PublicAddOption   bool  // Allow adding options
    MaxVotes          int   // Votes per user (0 = unlimited)
}
```

### Vote Update Logic

**File**: `Matterpoll/server/poll/poll.go` (lines 120-180)

```go
func (p *Poll) UpdateVote(userID string, index int) (*i18n.LocalizeConfig, error) {
    // Validate option index
    if index < 0 || index >= len(p.AnswerOptions) {
        return responseVoteInvalidOption, errors.New("invalid option index")
    }

    // Find all options user has voted for
    votedAnswers := []int{}
    for i, o := range p.AnswerOptions {
        for _, voter := range o.Voter {
            if voter == userID {
                votedAnswers = append(votedAnswers, i)
                break
            }
        }
    }

    // Check if user already voted for this option
    for _, answer := range votedAnswers {
        if answer == index {
            // Remove vote (toggle off)
            p.AnswerOptions[index].Voter = removeFromStringArray(
                p.AnswerOptions[index].Voter, userID)
            return responseVoteRemoved, nil
        }
    }

    // Check max votes limit
    if p.Settings.MaxVotes > 0 && len(votedAnswers) >= p.Settings.MaxVotes {
        return responseVoteMaxReached, errors.New("max votes reached")
    }

    // Add vote
    p.AnswerOptions[index].Voter = append(
        p.AnswerOptions[index].Voter, userID)

    return nil, nil
}

func (p *Poll) HasVoted(userID string) bool {
    for _, o := range p.AnswerOptions {
        for _, voter := range o.Voter {
            if voter == userID {
                return true
            }
        }
    }
    return false
}
```

### Database Storage

**File**: `Matterpoll/server/store/kvstore/poll_store.go`

Polls are stored in Mattermost's KV store with key pattern: `poll_{pollID}`

```go
func (s *PollStore) Insert(poll *poll.Poll) error {
    data, err := json.Marshal(poll)
    if err != nil {
        return err
    }
    return s.api.KVSet(poll.ID, data)
}

func (s *PollStore) Get(id string) (*poll.Poll, error) {
    data, appErr := s.api.KVGet(id)
    if appErr != nil {
        return nil, appErr
    }

    var p poll.Poll
    if err := json.Unmarshal(data, &p); err != nil {
        return nil, err
    }

    return &p, nil
}

func (s *PollStore) Update(oldPoll, newPoll *poll.Poll) error {
    // Optimistic locking: compare-and-swap based on old value
    data, err := json.Marshal(newPoll)
    if err != nil {
        return err
    }

    var oldData []byte
    if oldPoll != nil {
        oldData, err = json.Marshal(oldPoll)
        if err != nil {
            return err
        }
    }

    success, appErr := s.api.KVCompareAndSet(newPoll.ID, oldData, data)
    if appErr != nil {
        return appErr
    }
    if !success {
        return errors.New("concurrent modification detected")
    }

    return nil
}
```

---

## 6. REAL-TIME UPDATES VIA WEBSOCKET

### Publishing Poll Metadata

**File**: `Matterpoll/server/plugin/api.go` (lines 547-567)

```go
func (p *MatterpollPlugin) publishPollMetadata(poll *poll.Poll, userID string) {
    // Get metadata for this user
    canManagePoll := p.canManagePoll(userID, poll)
    metadata := poll.GetMetadata(userID, canManagePoll)

    // Broadcast WebSocket event
    p.API.PublishWebSocketEvent(
        "has_voted",
        map[string]interface{}{
            "poll_id": poll.ID,
            "metadata": metadata,
        },
        &model.WebsocketBroadcast{
            ChannelId: "", // Send to specific user only
            UserId:    userID,
        },
    )
}
```

### Metadata Structure

**File**: `Matterpoll/server/poll/poll.go` (lines 200-220)

```go
type Metadata struct {
    VotedAnswers         []int  `json:"voted_answers"`
    CanManagePoll        bool   `json:"can_manage_poll"`
    SettingProgress      bool   `json:"setting_progress"`
    SettingPublicAddOption bool `json:"setting_public_add_option"`
}

func (p *Poll) GetMetadata(userID string, canManagePoll bool) *Metadata {
    votedAnswers := []int{}

    // Find which options user has voted for
    for i, option := range p.AnswerOptions {
        for _, voter := range option.Voter {
            if voter == userID {
                votedAnswers = append(votedAnswers, i)
                break
            }
        }
    }

    return &Metadata{
        VotedAnswers:           votedAnswers,
        CanManagePoll:          canManagePoll,
        SettingProgress:        p.Settings.Progress,
        SettingPublicAddOption: p.Settings.PublicAddOption,
    }
}
```

---

## 7. CUSTOM REACT COMPONENT (WEBAPP)

### Component Registration

**File**: `Matterpoll/webapp/src/index.jsx` (lines 20-35)

```javascript
export default class Plugin {
    initialize(registry, store) {
        // Register custom post type component
        registry.registerPostTypeComponent(
            'custom_matterpoll',
            PostTypeMatterpoll
        );

        // Register reducer for poll metadata
        registry.registerReducer(reducer);

        // Register WebSocket event handler
        registry.registerWebSocketEventHandler(
            'custom_matterpoll_has_voted',
            handleWebSocketHasVoted(store)
        );
    }
}
```

### Post Type Component

**File**: `Matterpoll/webapp/src/components/post_type/post_type.jsx` (lines 26-93)

```jsx
import React from 'react';
import ActionView from './action_view';

export default class PostTypeMatterpoll extends React.PureComponent {
    render() {
        const {post} = this.props;

        // Extract SlackAttachment (added by ParseSlackAttachment on server)
        const attachment = post.props.attachments?.[0];
        if (!attachment) {
            return null;
        }

        const pollId = post.props.poll_id;

        // Fetch poll metadata on mount
        this.props.actions.fetchPollMetadata(pollId);

        return (
            <div className='attachment attachment--matterpoll'>
                {/* Author name */}
                {attachment.author_name && (
                    <div className='attachment__author-name'>
                        {attachment.author_name}
                    </div>
                )}

                {/* Poll title (question) */}
                {attachment.title && (
                    <div className='attachment__title'>
                        <h3>{attachment.title}</h3>
                    </div>
                )}

                {/* Poll text (options list) */}
                {attachment.text && (
                    <div className='attachment__text'>
                        <pre>{attachment.text}</pre>
                    </div>
                )}

                {/* Render action buttons */}
                <ActionView
                    postId={post.id}
                    pollId={pollId}
                    actions={attachment.actions || []}
                />
            </div>
        );
    }
}
```

### Action View (Button Renderer)

**File**: `Matterpoll/webapp/src/components/post_type/action_view.jsx` (lines 65-109)

```jsx
import React from 'react';
import {connect} from 'react-redux';
import ActionButton from './action_button';
import {voteAnswer} from '../../actions/vote';

class ActionView extends React.PureComponent {
    render() {
        const {actions, pollId, metadata, currentUserId} = this.props;

        if (!actions || actions.length === 0) {
            return null;
        }

        return (
            <div className='attachment-actions'>
                {actions.map((action, index) => {
                    // Filter buttons based on permissions
                    const canManage = metadata?.can_manage_poll || false;

                    // Hide management buttons if user can't manage
                    if (['resetVotes', 'endPoll', 'deletePoll'].includes(action.id) && !canManage) {
                        return null;
                    }

                    // Hide "Add Option" if not public-add-option
                    if (action.id === 'addOption' && !metadata?.setting_public_add_option) {
                        return null;
                    }

                    // Check if user voted for this option
                    const voted = metadata?.voted_answers?.includes(index) || false;

                    return (
                        <ActionButton
                            key={action.id}
                            action={action}
                            voted={voted}
                            postId={this.props.postId}
                            voteAnswer={this.props.actions.voteAnswer}
                        />
                    );
                })}
            </div>
        );
    }
}

const mapStateToProps = (state, ownProps) => ({
    metadata: state['plugins-matterpoll'].pollsMetadata[ownProps.pollId],
    currentUserId: state.entities.users.currentUserId,
});

const mapDispatchToProps = (dispatch) => ({
    actions: {
        voteAnswer: (postId, actionId) => dispatch(voteAnswer(postId, actionId)),
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(ActionView);
```

### Action Button Component

**File**: `Matterpoll/webapp/src/components/post_type/action_button.jsx` (lines 33-88)

```jsx
import React from 'react';
import styled from 'styled-components';

class ActionButton extends React.PureComponent {
    handleAction = (e) => {
        e.preventDefault();

        const actionId = e.currentTarget.getAttribute('data-action-id');

        // Call Redux action to vote
        this.props.voteAnswer(this.props.postId, actionId);
    };

    render() {
        const {action, voted} = this.props;

        // Determine button color based on style
        let backgroundColor = '#166DE0'; // primary (blue)
        if (action.style === 'danger') {
            backgroundColor = '#D24B4E'; // danger (red)
        } else if (action.style === 'default') {
            backgroundColor = '#3DB887'; // default (green)
        }

        // Darken background if user voted for this option
        if (voted) {
            backgroundColor = '#1B2126'; // Dark gray for voted
        }

        return (
            <StyledButton
                data-action-id={action.id}
                onClick={this.handleAction}
                backgroundColor={backgroundColor}
            >
                {action.name}
            </StyledButton>
        );
    }
}

const StyledButton = styled.button`
    border-radius: 4px;
    border: none;
    color: white;
    background-color: ${props => props.backgroundColor};
    padding: 10px 16px;
    margin: 5px 5px 0 0;
    cursor: pointer;
    font-weight: 600;

    &:hover {
        opacity: 0.8;
    }
`;

export default ActionButton;
```

### Vote Action (Redux)

**File**: `Matterpoll/webapp/src/actions/vote.js` (lines 1-8)

```javascript
import {doPostAction} from 'mattermost-redux/actions/posts';

export const voteAnswer = (postId, actionId) => async (dispatch) => {
    // Use Mattermost's built-in doPostAction
    // This POSTs to the Integration.URL specified in the PostAction
    return dispatch(doPostAction(postId, actionId));
};
```

**What `doPostAction` does:**
1. Finds the PostAction by `actionId` in the post's attachments
2. Extracts the `Integration.URL`
3. Sends POST request to that URL with `PostActionIntegrationRequest` body
4. Receives `PostActionIntegrationResponse` from server
5. Updates the post in Redux store if `response.Update` is present

---

## 8. COMPLETE FLOW DIAGRAM

```
User types: /poll "Lunch?" "Pizza" "Tacos" "Salad"
        ↓
ExecuteCommand() in command.go:89
        ↓
Parse command arguments
        ↓
NewPoll() creates Poll object
        ↓
Store.Poll().Insert(poll) → Save to KV store
        ↓
poll.ToPostActions() generates buttons with Integration URLs
    ├─ vote0: /api/v1/polls/{id}/vote/0
    ├─ vote1: /api/v1/polls/{id}/vote/1
    ├─ vote2: /api/v1/polls/{id}/vote/2
    ├─ resetVotes: /api/v1/polls/{id}/votes/reset
    ├─ addOption: /api/v1/polls/{id}/option/add/request
    ├─ endPoll: /api/v1/polls/{id}/end
    └─ deletePoll: /api/v1/polls/{id}/delete
        ↓
Create Post object
    Type: "custom_matterpoll"
    Props: {poll_id: "abc123"}
        ↓
CRITICAL: model.ParseSlackAttachment(post, actions)
    → Adds SlackAttachment to post.props.attachments
    → Preserves Integration URLs even with custom type
        ↓
API.CreatePost(post) → Posted to channel
        ↓
[Frontend] PostTypeMatterpoll component renders
    ├─ Displays poll question (attachment.title)
    ├─ Displays options (attachment.text)
    ├─ Renders ActionView with buttons
    └─ Fetches poll metadata via /api/v1/polls/{id}/metadata
        ↓
[User clicks "Pizza" button]
        ↓
ActionButton.handleAction() triggered
        ↓
voteAnswer(postId, "vote0") Redux action
        ↓
doPostAction(postId, "vote0")
    → Finds action with id="vote0" in post
    → Extracts Integration.URL: /api/v1/polls/{id}/vote/0
    → POSTs to plugin endpoint
        ↓
handleVote() in api.go:360
    ├─ Extract pollID and optionNumber from URL
    ├─ Load poll from KV store
    ├─ poll.UpdateVote(userID, 0)
    │   ├─ Check if already voted for option 0 → toggle off
    │   ├─ Check max votes limit
    │   └─ Add userID to AnswerOptions[0].Voter array
    ├─ Store.Poll().Update(poll) → Save to KV store
    └─ publishPollMetadata(poll, userID)
        → WebSocket event: has_voted
        ↓
Generate updated post
    ├─ poll.ToPostActions() with new vote counts
    ├─ model.ParseSlackAttachment(post, updatedActions)
    └─ post.AddProp("poll_id", pollID)
        ↓
Return PostActionIntegrationResponse
    ├─ EphemeralText: "Your vote has been counted"
    └─ Update: <updated post with new vote counts>
        ↓
[Frontend] Mattermost updates post in UI
    ├─ Post re-renders with updated attachments
    ├─ Button now shows "Pizza (1)" if progress mode
    └─ WebSocket handler updates Redux metadata
        → Voted button changes to dark background
```

---

## 9. KEY FILES SUMMARY

| File | Lines | Purpose |
|------|-------|---------|
| `server/plugin/command.go` | 89-227 | `/poll` slash command → creates poll post |
| `server/poll/transform.go` | 67-159 | **CRITICAL** - Generates PostActions with Integration URLs |
| `server/plugin/api.go` | 75-95 | API router setup for all endpoints |
| `server/plugin/api.go` | 97-135 | PostAction wrapper - handles all button clicks |
| `server/plugin/api.go` | 360-425 | Vote handler - processes voting |
| `server/poll/poll.go` | 120-180 | Vote update logic and validation |
| `server/store/kvstore/poll_store.go` | - | Poll persistence in Mattermost KV store |
| `webapp/src/index.jsx` | 20-35 | Register custom post type component |
| `webapp/src/components/post_type/post_type.jsx` | 26-93 | Custom post renderer |
| `webapp/src/components/post_type/action_view.jsx` | 65-109 | Button container with filtering |
| `webapp/src/components/post_type/action_button.jsx` | 33-88 | Individual button with styling |
| `webapp/src/actions/vote.js` | 1-8 | Redux action - calls `doPostAction` |

---

## 10. THE CRITICAL PATTERN (PROVEN TO WORK)

### Server Side

```go
// 1. Create PostActions with Integration URLs
actions := []*model.PostAction{
    {
        Id:    "action_id",
        Name:  "Button Text",
        Type:  model.PostActionTypeButton,
        Style: "default", // or "primary", "danger"
        Integration: &model.PostActionIntegration{
            URL: fmt.Sprintf("/plugins/%s/api/v1/endpoint/%s",
                pluginID, resourceID),
            // Optional: Context map (Matterpoll doesn't use this)
        },
    },
}

attachment := &model.SlackAttachment{
    Title:   "Title",
    Text:    "Description",
    Actions: actions,
}

// 2. Create post with CUSTOM TYPE
post := &model.Post{
    UserId:    botUserID,
    ChannelId: channelID,
    Type:      "custom_yourplugin", // ✅ Custom type works!
    Props: map[string]interface{}{
        "resource_id": resourceID,
    },
}

// 3. CRITICAL: Use ParseSlackAttachment (not direct Props)
model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})

// 4. Create the post
createdPost, err := p.API.CreatePost(post)
```

### Client Side

```jsx
// 1. Register custom post type
registry.registerPostTypeComponent('custom_yourplugin', YourPostComponent);

// 2. Render component
class YourPostComponent extends React.PureComponent {
    render() {
        const {post} = this.props;
        const attachment = post.props.attachments?.[0];

        return (
            <div>
                <h3>{attachment.title}</h3>
                <p>{attachment.text}</p>

                {attachment.actions?.map(action => (
                    <button
                        onClick={() => this.handleAction(action.id)}
                        key={action.id}
                    >
                        {action.name}
                    </button>
                ))}
            </div>
        );
    }

    handleAction = (actionId) => {
        // Use mattermost-redux doPostAction
        this.props.actions.doPostAction(this.props.post.id, actionId);
    }
}
```

---

## 11. COMPARISON: MATTERPOLL VS ZOOM PLUGIN

| Aspect | Matterpoll | Zoom Plugin |
|--------|-----------|-------------|
| **Post Type** | `custom_matterpoll` ✅ | `custom_zoom` ✅ |
| **Interactive Buttons** | YES - voting with callbacks | NO - links only |
| **ParseSlackAttachment** | ✅ Used | ✅ Used |
| **Integration URLs** | ✅ Work perfectly | ⚠️ Only on ephemeral posts |
| **URL Pattern** | Path params: `/vote/{num}` | Context map data |
| **State Storage** | KV store (poll object) | Props + KV store |
| **Custom React** | Full component tree | Basic rendering |
| **Real-time Updates** | WebSocket metadata | WebSocket events |
| **Post Updates** | Button clicks update post | Webhook updates post |

**Key Insight:** Matterpoll proves custom post types with Integration URLs DO work in Mattermost!

---

## 12. SECURITY & PERMISSIONS

### User Validation

```go
// Mattermost-User-Id header must match request.UserId
userID := r.Header.Get("Mattermost-User-Id")
if userID != request.UserId {
    http.Error(w, "user ID mismatch", http.StatusUnauthorized)
    return
}
```

### Permission Checks

```go
func (p *MatterpollPlugin) canManagePoll(userID string, poll *poll.Poll) bool {
    // Creator can manage
    if poll.Creator == userID {
        return true
    }

    // System admins can manage
    user, err := p.API.GetUser(userID)
    if err == nil && user.IsSystemAdmin() {
        return true
    }

    return false
}
```

### Concurrent Modification Protection

Uses optimistic locking via `KVCompareAndSet`:

```go
success, appErr := s.api.KVCompareAndSet(newPoll.ID, oldData, data)
if !success {
    return errors.New("concurrent modification detected")
}
```

---

## 13. TROUBLESHOOTING GUIDE

### If Buttons Don't Appear

1. Check `model.ParseSlackAttachment(post, actions)` is called
2. Verify `Actions` array is not empty
3. Check browser console for errors
4. Inspect post object in Redux DevTools

### If Buttons Appear But Don't Work

1. Verify Integration.URL format: `/plugins/{pluginId}/api/v1/...`
2. Check HTTP route is registered in `ServeHTTP`
3. Look for CORS errors in browser console
4. Verify plugin is activated (`plugin list` in CLI)
5. Check server logs for errors

### If Post Doesn't Update After Click

1. Verify handler returns `PostActionIntegrationResponse`
2. Check `response.Update` contains updated post
3. Ensure `ParseSlackAttachment` called on updated post
4. Verify JSON encoding doesn't fail

### If Permissions Are Wrong

1. Check `Mattermost-User-Id` header validation
2. Verify user has channel permissions
3. Check WebSocket event user ID matches
4. Validate poll creator ID matches

---

## 14. IMPLEMENTATION CHECKLIST

**Server Side:**
- [ ] Create PostAction structs with Integration URLs
- [ ] Use `model.ParseSlackAttachment(post, actions)`
- [ ] Set custom post type: `Type: "custom_yourplugin"`
- [ ] Register HTTP routes for Integration URLs
- [ ] Implement PostActionIntegrationRequest handler
- [ ] Return PostActionIntegrationResponse with Update
- [ ] Store resource ID in post Props for tracking
- [ ] Validate user ID from header

**Client Side:**
- [ ] Register custom post type component
- [ ] Extract attachment from `post.props.attachments[0]`
- [ ] Render buttons from `attachment.actions`
- [ ] Implement button click handler
- [ ] Call `doPostAction(postId, actionId)` from mattermost-redux
- [ ] Handle WebSocket events for real-time updates (optional)
- [ ] Style buttons based on `action.style`

**Testing:**
- [ ] Create post - buttons appear
- [ ] Click button - endpoint receives request
- [ ] Verify PostActionIntegrationRequest contains correct data
- [ ] Post updates after button click
- [ ] Multiple users can interact simultaneously
- [ ] Works on mobile app (not just webapp)
- [ ] Permissions enforced correctly

---

## 15. CONCLUSIONS

1. **Custom post types CAN have working Integration URLs** - Matterpoll proves this conclusively

2. **`model.ParseSlackAttachment()` is critical** - Don't manually set `props.attachments`

3. **URL path params vs Context map** - Both patterns work (Matterpoll uses paths, Zoom uses Context)

4. **State must be stored separately** - Use KV store, not just post Props

5. **Real-time updates** - WebSocket events enhance UX but aren't required

6. **Custom React components** - Give full control over rendering and UX

7. **The pattern is production-ready** - Matterpoll is widely used and stable

**Bottom Line:** Follow Matterpoll's pattern for reliable, working interactive buttons with custom post types in Mattermost plugins.
