# Story 1.1: Plugin Foundation & Slash Command Registration

**Status:** done
**Epic:** Epic 1 - Approval Request Creation & Management
**Story ID:** 1-1-plugin-foundation-slash-command-registration
**Estimated Complexity:** Medium

---

## Story

As a Mattermost user,
I want to type `/approve` and see available commands,
So that I can discover how to create approval requests without reading documentation.

---

## Acceptance Criteria

### AC1: Help Text Display

**Given** the plugin is installed and activated on a Mattermost server
**When** a user types `/approve` without arguments
**Then** the system displays help text showing available commands (`new`, `list`, `get`, `cancel`)
**And** the help text includes brief descriptions of each command
**And** the help text is returned as an ephemeral message (visible only to the user)

### AC2: Invalid Command Handling

**Given** the plugin is installed and activated
**When** a user types `/approve invalid-command`
**Then** the system displays an error message stating the command is not recognized
**And** the system suggests valid commands

### AC3: Plugin Initialization

**Given** the Mattermost server is running Plugin API v6.0+
**When** the plugin initializes
**Then** the plugin registers the `/approve` slash command successfully
**And** the plugin uses the mattermost-plugin-starter-template structure
**And** the plugin implements Mattermost Go Style Guide conventions

---

## Technical Context

### Requirements Coverage

**Functional Requirements:**
- **FR25:** Users can invoke approval functionality via slash commands
- **FR26:** The system provides user-friendly command help and guidance
- **FR27:** Error messages are specific and actionable
- **FR28:** The system validates user input before accepting approval requests

**Non-Functional Requirements:**
- **NFR-C1:** Plugin works on Mattermost Server v6.0+ (Plugin API v6.0+)
- **NFR-M1:** Code simplicity over cleverness
- **NFR-M2:** Implementation follows official Mattermost plugin patterns and conventions

### Architectural Decisions

**From Architecture Document:**

1. **Starter Template:** Use mattermost-plugin-starter-template as foundation
   - Location: `https://github.com/mattermost/mattermost-plugin-starter-template`
   - Provides build infrastructure, plugin boilerplate, Makefile targets

2. **Package Structure:**
   ```
   server/
     plugin.go              # Main plugin struct, lifecycle hooks
     configuration.go       # Plugin configuration
     command/
       approve.go          # Slash command handlers
       router.go           # Command routing logic
   ```

3. **Naming Conventions (Mattermost Go Style Guide):**
   - CamelCase for variables: `approvalID`, `userID`, not `approval_id`
   - Proper initialisms: `HTTPClient`, `APIEndpoint`, not `HttpClient`
   - Method receivers: 1-2 letter abbreviations like `(p *Plugin)`
   - Error variables: Use `err` for standard errors

4. **Error Handling Pattern:**
   ```go
   // ✅ Include user input in validation errors
   if cmd == "" {
       return fmt.Errorf("invalid command: must provide a subcommand")
   }

   // ✅ Wrap errors with context (use %w for unwrapping)
   err := p.registerCommand()
   if err != nil {
       return fmt.Errorf("failed to register slash command: %w", err)
   }
   ```

5. **Anti-Patterns to AVOID:**
   - ❌ NO `pkg/`, `util/`, or `misc/` packages
   - ❌ NO boolean validation without errors
   - ❌ NO double-logging (log at highest layer only)
   - ❌ NO custom ToJSON methods
   - ❌ NO `else` after returns

### Plugin API Integration Points

**Hooks Required:**
- `OnActivate()` - Plugin initialization, slash command registration
- `ExecuteCommand()` - Handle `/approve` slash command invocations

**Plugin API Methods:**
- `p.API.RegisterCommand()` - Register `/approve` slash command
- `p.API.LogInfo()` / `p.API.LogError()` - Structured logging
- `&model.CommandResponse{}` - Return command responses

---

## Tasks / Subtasks

### Task 1: Initialize Project with Starter Template

**Objective:** Set up the plugin project using the official Mattermost starter template.

**Steps:**
1. Clone mattermost-plugin-starter-template as project base
2. Update plugin.json manifest:
   - ID: `com.mattermost.plugin-approver2`
   - Name: `Mattermost Approval Workflow`
   - Description: `Fast, authoritative approvals for time-sensitive decisions`
   - Min server version: `6.0.0`
3. Update go.mod with correct module path
4. Remove webapp/ folder (backend-only plugin for MVP)
5. Verify build: `make` and `make test`

**Files Created/Modified:**
- `plugin.json` - Plugin manifest
- `go.mod` - Go module definition
- `Makefile` - Build configuration (from starter)
- `.gitignore` - Ignore build artifacts

**Testing:**
- Run `make` - should compile successfully
- Run `make test` - should pass default tests

---

### Task 2: Implement Plugin Struct & OnActivate Hook

**Objective:** Create the main plugin struct and initialization logic.

**Implementation:**

**File:** `server/plugin.go`

```go
package main

import (
	"fmt"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is called when the plugin is activated.
func (p *Plugin) OnActivate() error {
	p.API.LogInfo("Activating Mattermost Approval Workflow plugin")

	// Register slash command
	if err := p.registerCommand(); err != nil {
		return fmt.Errorf("failed to register slash command: %w", err)
	}

	p.API.LogInfo("Mattermost Approval Workflow plugin activated successfully")
	return nil
}

// registerCommand registers the /approve slash command
func (p *Plugin) registerCommand() error {
	cmd := &model.Command{
		Trigger:          "approve",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage approval requests",
		AutoCompleteHint: "[new|list|get|cancel|help]",
		DisplayName:      "Approval Request",
		Description:      "Create, manage, and view approval requests",
	}

	if err := p.API.RegisterCommand(cmd); err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}
```

**Key Design Decisions:**
- Use `plugin.MattermostPlugin` embedded struct for Plugin API access
- Implement configuration lock pattern (from starter template)
- Log at plugin initialization layer (highest appropriate layer)
- Error wrapping with `%w` for context preservation
- CamelCase naming: `registerCommand`, not `register_command`
- 1-letter receiver: `(p *Plugin)` not `(me *Plugin)`

**Testing:**
- Unit test for `OnActivate()` with mock Plugin API
- Verify command registration succeeds
- Verify error handling if registration fails

---

### Task 3: Implement ExecuteCommand Hook with Help Text

**Objective:** Handle `/approve` command invocations and display help text.

**Implementation:**

**File:** `server/plugin.go` (continued)

```go
// ExecuteCommand executes a command that has been previously registered via the RegisterCommand API.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Split command arguments
	split := strings.Fields(args.Command)
	if len(split) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Invalid command format.",
		}, nil
	}

	// First argument is "/approve", second is the subcommand
	subcommand := ""
	if len(split) > 1 {
		subcommand = split[1]
	}

	// Route to appropriate handler
	switch subcommand {
	case "":
		return p.executeHelpCommand(), nil
	case "help":
		return p.executeHelpCommand(), nil
	default:
		return p.executeUnknownCommand(subcommand), nil
	}
}

// executeHelpCommand returns help text for the /approve command
func (p *Plugin) executeHelpCommand() *model.CommandResponse {
	helpText := `### Mattermost Approval Workflow

**Available Commands:**

* **/approve new** - Create a new approval request
* **/approve list** - View your approval requests and decisions
* **/approve get [ID]** - View a specific approval by ID
* **/approve cancel [ID]** - Cancel a pending approval request
* **/approve help** - Display this help text

**Example:**
` + "`/approve new`" + ` - Opens a modal to create an approval request

For more information, visit the plugin documentation.`

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         helpText,
	}
}

// executeUnknownCommand returns an error message for unrecognized subcommands
func (p *Plugin) executeUnknownCommand(subcommand string) *model.CommandResponse {
	errorText := fmt.Sprintf("Unknown command: **%s**\n\nValid commands: `new`, `list`, `get`, `cancel`, `help`\n\nType `/approve help` for more information.", subcommand)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         errorText,
	}
}
```

**Key Design Decisions:**
- Return `CommandResponse` with `ResponseTypeEphemeral` (visible only to user)
- Use Markdown formatting for structured help text
- Include specific, actionable error messages (FR27)
- Suggest valid commands when invalid command provided (AC2)
- No logging for help text display (not an error condition)
- Simple string switch for command routing (avoid over-engineering)

**Testing:**
- Test `/approve` with no arguments → returns help text
- Test `/approve help` → returns help text
- Test `/approve invalid-command` → returns error with suggestions
- Verify help text includes all MVP commands (new, list, get, cancel)
- Verify response type is ephemeral

---

### Task 4: Implement Command Routing Structure

**Objective:** Create extensible command routing for future command handlers.

**Implementation:**

**File:** `server/command/router.go`

```go
package command

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Handler defines the interface for command handlers
type Handler interface {
	Handle(args *model.CommandArgs) (*model.CommandResponse, error)
}

// Router routes slash command invocations to appropriate handlers
type Router struct {
	api plugin.API
}

// NewRouter creates a new command router
func NewRouter(api plugin.API) *Router {
	return &Router{
		api: api,
	}
}

// Route determines which handler should process the command
func (r *Router) Route(args *model.CommandArgs) (*model.CommandResponse, error) {
	split := strings.Fields(args.Command)
	if len(split) < 1 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Invalid command format.",
		}, nil
	}

	// Extract subcommand (first arg after /approve)
	subcommand := ""
	if len(split) > 1 {
		subcommand = split[1]
	}

	// Route to appropriate handler
	switch subcommand {
	case "":
		return executeHelp(), nil
	case "help":
		return executeHelp(), nil
	default:
		return executeUnknown(subcommand), nil
	}
}

// executeHelp returns help text
func executeHelp() *model.CommandResponse {
	helpText := `### Mattermost Approval Workflow

**Available Commands:**

* **/approve new** - Create a new approval request
* **/approve list** - View your approval requests and decisions
* **/approve get [ID]** - View a specific approval by ID
* **/approve cancel [ID]** - Cancel a pending approval request
* **/approve help** - Display this help text

**Example:**
` + "`/approve new`" + ` - Opens a modal to create an approval request

For more information, visit the plugin documentation.`

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         helpText,
	}
}

// executeUnknown returns error for unrecognized commands
func executeUnknown(subcommand string) *model.CommandResponse {
	errorText := fmt.Sprintf("Unknown command: **%s**\n\nValid commands: `new`, `list`, `get`, `cancel`, `help`\n\nType `/approve help` for more information.", subcommand)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         errorText,
	}
}
```

**File:** `server/plugin.go` (updated ExecuteCommand)

```go
import (
	"github.com/your-org/mattermost-plugin-approver2/server/command"
)

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	router := command.NewRouter(p.API)

	response, err := router.Route(args)
	if err != nil {
		p.API.LogError("Command execution failed", "error", err.Error(), "command", args.Command)
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "An error occurred while processing your command. Please try again.",
		}, nil
	}

	return response, nil
}
```

**Key Design Decisions:**
- Feature-based package: `server/command/` (not `pkg/` or `util/`)
- Handler interface allows extensibility for future commands
- Router encapsulates command parsing logic
- Logging at highest layer (plugin.go), not in router
- Simple, readable code over clever abstractions

**Testing:**
- Test routing for all known commands
- Test routing for unknown commands
- Verify error handling propagates correctly
- Verify response types are preserved

---

### Task 5: Configuration Management

**Objective:** Implement configuration struct (from starter template pattern).

**Implementation:**

**File:** `server/configuration.go`

```go
package main

import (
	"github.com/pkg/errors"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
type configuration struct {
	// No configuration fields needed for MVP
	// Future: approval templates, notification settings, etc.
}

// Clone creates a deep copy of the configuration
func (c *configuration) Clone() *configuration {
	var clone = *c
	return &clone
}

// IsValid checks if the configuration is valid
func (c *configuration) IsValid() error {
	// No validation needed for MVP (no config fields)
	return nil
}

// getConfiguration retrieves the active configuration under lock
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty
		return
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(configuration)

	// Load the public configuration fields from the Mattermost server configuration
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	return nil
}
```

**Key Design Decisions:**
- Configuration struct is empty for MVP (no settings needed)
- Follow starter template pattern for future extensibility
- Thread-safe configuration access with RWMutex
- Configuration validation pattern (IsValid) for future use

**Testing:**
- Test configuration loading
- Test thread-safe access patterns
- Verify validation pattern works (returns nil for MVP)

---

### Task 6: Testing Setup

**Objective:** Create comprehensive tests for plugin initialization and command routing.

**Implementation:**

**File:** `server/plugin_test.go`

```go
package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOnActivate(t *testing.T) {
	t.Run("successfully registers command", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)
		api.On("LogInfo", mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("handles registration failure", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(model.NewAppError("test", "test.error", nil, "", 500))
		api.On("LogInfo", mock.Anything, mock.Anything).Return()

		p := &Plugin{}
		p.SetAPI(api)

		err := p.OnActivate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register slash command")
	})
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		expectedContains string
		expectedType     string
	}{
		{
			name:             "no subcommand shows help",
			command:          "/approve",
			expectedContains: "Available Commands",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
		{
			name:             "help subcommand shows help",
			command:          "/approve help",
			expectedContains: "Available Commands",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
		{
			name:             "unknown subcommand shows error",
			command:          "/approve invalid",
			expectedContains: "Unknown command",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
		{
			name:             "unknown subcommand suggests valid commands",
			command:          "/approve foo",
			expectedContains: "Valid commands",
			expectedType:     model.CommandResponseTypeEphemeral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			p := &Plugin{}
			p.SetAPI(api)

			args := &model.CommandArgs{
				Command: tt.command,
			}

			resp, appErr := p.ExecuteCommand(nil, args)
			assert.Nil(t, appErr)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedType, resp.ResponseType)
			assert.Contains(t, resp.Text, tt.expectedContains)
		})
	}
}
```

**File:** `server/command/router_test.go`

```go
package command

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	api := &plugintest.API{}
	router := NewRouter(api)

	tests := []struct {
		name             string
		command          string
		expectedContains string
	}{
		{
			name:             "empty command shows help",
			command:          "/approve",
			expectedContains: "Available Commands",
		},
		{
			name:             "help command shows help",
			command:          "/approve help",
			expectedContains: "Available Commands",
		},
		{
			name:             "unknown command shows error",
			command:          "/approve unknown",
			expectedContains: "Unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := &model.CommandArgs{
				Command: tt.command,
			}

			resp, err := router.Route(args)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Contains(t, resp.Text, tt.expectedContains)
		})
	}
}
```

**Key Design Decisions:**
- Table-driven tests (Mattermost pattern)
- Use `testify/assert` for assertions
- Use `plugintest.API` for mocking Plugin API
- Test both success and error paths
- Verify error messages contain expected content

**Testing:**
- Run `make test` - all tests should pass
- Verify test coverage for plugin.go and command/router.go

---

### Task 7: Documentation & Build Verification

**Objective:** Update README and verify the plugin builds and runs.

**Implementation:**

**File:** `README.md`

```markdown
# Mattermost Approval Workflow Plugin

Fast, authoritative approvals for time-sensitive decisions - entirely within Mattermost.

## Overview

This plugin enables teams to request and grant official approvals without leaving Mattermost. It creates immutable, auditable approval records that provide confidence to requesters, protection to approvers, and clarity to auditors.

## Features (MVP)

- `/approve new` - Create approval requests via slash command
- `/approve list` - View your approval history
- `/approve get [ID]` - View specific approval records
- `/approve cancel [ID]` - Cancel pending requests
- Immutable approval records
- DM-based notifications
- No external dependencies

## Installation

### Prerequisites
- Mattermost Server v6.0 or later
- Go 1.19+

### Manual Installation

1. Download the latest release tarball: `com.mattermost.plugin-approver2.tar.gz`
2. In Mattermost System Console, navigate to **Plugin Management**
3. Click **Upload Plugin**
4. Select the tarball and upload
5. Enable the plugin

## Usage

### Create Approval Request
```
/approve new
```
Opens a modal to select an approver and describe what needs approval.

### View Your Approvals
```
/approve list
```
Shows all approval requests you've submitted or received.

### Get Help
```
/approve help
```
Displays available commands and usage information.

## Development

### Setup Development Environment

```bash
# Clone repository
git clone [repository-url]
cd mattermost-plugin-approver2

# Install dependencies
npm install

# Build plugin
make

# Run tests
make test

# Deploy to local Mattermost (requires MM_SERVICESETTINGS_SITEURL and MM_ADMIN_TOKEN)
make deploy
```

### Project Structure

```
server/
  plugin.go              # Main plugin hooks
  configuration.go       # Configuration management
  command/
    router.go           # Command routing
```

## Requirements

- Mattermost Server v6.0+
- Plugin API v6.0+

## License

[License information]

## Contributing

[Contribution guidelines]
```

**Build Verification Steps:**

1. Run `make` - verify compilation succeeds
2. Run `make test` - verify all tests pass
3. Run `make lint` (if configured) - verify no linting errors
4. Inspect generated tarball - verify structure is correct
5. Test installation on local Mattermost instance

**Testing:**
- Build produces valid plugin tarball
- Plugin can be uploaded to Mattermost
- `/approve` command is registered after activation
- `/approve help` displays help text
- `/approve invalid` shows error message

---

## Dev Notes

### Project Structure Notes

**Repository Root:**
- Project initialized using `mattermost-plugin-starter-template`
- Location: `/Users/wayne/Repositories/Mattermost-Plugin-Approver2`

**Key Directories:**
- `server/` - Go backend implementation
- `build/` - Build tools and scripts (from starter template)
- `_bmad-output/` - Planning artifacts (NOT part of plugin)

**Plugin Manifest:**
- File: `plugin.json`
- Plugin ID: `com.mattermost.plugin-approver2`
- Min server version: `6.0.0`

**Build System:**
- Makefile from starter template
- Commands: `make`, `make test`, `make deploy`, `make watch`

### Implementation Order

**This story establishes the foundation. Implement in this order:**

1. **Initialize project** (Task 1)
   - Use starter template
   - Configure plugin.json
   - Remove webapp/ folder (backend-only)

2. **Plugin hooks** (Task 2)
   - Implement OnActivate
   - Register slash command

3. **Command execution** (Task 3)
   - Implement ExecuteCommand hook
   - Add help text and error handling

4. **Command routing** (Task 4)
   - Create command/router.go
   - Extensible routing structure

5. **Configuration** (Task 5)
   - Empty config for MVP
   - Thread-safe access pattern

6. **Testing** (Task 6)
   - Unit tests for plugin hooks
   - Table-driven tests for command routing

7. **Documentation** (Task 7)
   - Update README
   - Verify build and deployment

### Mattermost Conventions Checklist

**Naming:**
- ✅ CamelCase variables: `approvalID`, not `approval_id`
- ✅ Proper initialisms: `APIEndpoint`, not `ApiEndpoint`
- ✅ Method receivers: `(p *Plugin)`, not `(me *Plugin)`

**Error Handling:**
- ✅ Wrap errors with context: `fmt.Errorf("context: %w", err)`
- ✅ Include user input in validation errors
- ✅ Return descriptive errors, not boolean checks

**Logging:**
- ✅ Log at highest layer (plugin.go)
- ✅ Use snake_case keys: `"approval_id"`, not `"approvalId"`
- ✅ Avoid double-logging (don't log and return error)

**Code Structure:**
- ✅ Feature-based packages (command/), not util/pkg/
- ✅ Synchronous by default (no unmanaged goroutines)
- ✅ Return structs, accept interfaces
- ✅ Avoid `else` after returns

**Testing:**
- ✅ Table-driven tests
- ✅ Co-located with implementation (*_test.go)
- ✅ Use testify/assert and plugintest.API

### Anti-Patterns to AVOID

- ❌ NO `pkg/`, `util/`, or `misc/` packages
- ❌ NO boolean validation without errors
- ❌ NO custom ToJSON methods
- ❌ NO double-logging errors
- ❌ NO unmanaged goroutines
- ❌ NO `else` after early returns
- ❌ NO camelCase in log keys

### References

**Architecture Document:**
- `/Users/wayne/Repositories/Mattermost-Plugin-Approver2/_bmad-output/planning-artifacts/architecture.md`
- Sections: Package Organization, Naming Conventions, Error Handling, Logging Patterns

**PRD:**
- `/Users/wayne/Repositories/Mattermost-Plugin-Approver2/_bmad-output/planning-artifacts/prd.md`
- Relevant sections: Functional Requirements, Developer Tool Requirements

**Epics:**
- `/Users/wayne/Repositories/Mattermost-Plugin-Approver2/_bmad-output/planning-artifacts/epics.md`
- Story 1.1 complete details with acceptance criteria

**Mattermost Resources:**
- Plugin Starter Template: `https://github.com/mattermost/mattermost-plugin-starter-template`
- Mattermost Go Style Guide: `http://developers.mattermost.com/contribute/more-info/server/style-guide/`
- Plugin API Documentation: `https://developers.mattermost.com/integrate/plugins/`

---

## File List

**New Files:**
- `server/command/router.go` - Command routing logic
- `server/command/router_test.go` - Router unit tests

**Modified Files:**
- `plugin.json` - Updated ID, name, description, min server version; removed webapp config
- `go.mod` - Updated module path and Go version
- `go.sum` - Updated dependency checksums
- `server/plugin.go` - Implemented OnActivate, registerCommand, ExecuteCommand
- `server/plugin_test.go` - Added comprehensive unit tests for plugin hooks
- `server/configuration.go` - Added IsValid method
- `README.md` - Complete documentation for Approver2 plugin

**Deleted Files:**
- `webapp/` - Removed (backend-only plugin)
- `server/command/command.go` - Replaced with router.go
- `server/command/command_test.go` - Replaced with router_test.go
- `server/job.go` - Removed unused background job code

---

## Change Log

**2026-01-11 - Code Review Fixes Applied**
- Removed all debug logging from plugin.go and router.go (architecture compliance)
- Removed unused Handler interface from router.go (avoid over-engineering)
- Simplified test mocks (no longer mocking removed logging calls)
- Fixed File List documentation (plugin_test.go listed as Modified, added go.sum)
- All 30 tests still passing after changes
- Status: code review complete, ready for production

**2026-01-10 - Story 1.1 Implementation Complete**
- Initialized plugin foundation using mattermost-plugin-starter-template
- Implemented slash command registration (`/approve`)
- Added help text display (`/approve` or `/approve help`)
- Implemented error handling for unknown commands
- Created extensible command routing structure
- Completed comprehensive test coverage (30 tests passing)
- Updated documentation (README.md)
- All 7 tasks completed successfully
- Status: ready for code review

---

## Dev Agent Record

**Story Created:** 2026-01-10
**Epic:** Epic 1 - Approval Request Creation & Management
**Dependencies:** None (foundation story)
**Blocks:** Stories 1.2, 1.3, 1.4, 1.5, 1.6, 1.7

**Key Implementation Notes:**
- This is the foundation story - all subsequent stories depend on this
- Establishes plugin structure, build system, and command routing
- No database/KV store operations yet (Story 1.2)
- No modal interactions yet (Story 1.3)
- Focus: Plugin activation, slash command registration, help text, error handling

**Testing Requirements:**
- Unit tests for OnActivate hook
- Unit tests for ExecuteCommand hook
- Integration tests for command routing
- Manual testing: install plugin, verify `/approve` works

**Definition of Done:**
- Plugin builds successfully (`make`)
- All tests pass (`make test`)
- Plugin can be installed on Mattermost v6.0+
- `/approve` command is registered
- `/approve` (no args) shows help text
- `/approve help` shows help text
- `/approve invalid` shows error message with suggestions
- README documentation complete

**Estimated Story Points:** 3 (Medium complexity)
- Project setup: 0.5
- Plugin hooks: 1.0
- Command routing: 1.0
- Testing & documentation: 0.5

---

### Completion Notes

**Implementation Summary:**

Story 1.1 has been successfully completed. All 7 tasks were implemented following TDD principles and Mattermost coding conventions.

**What Was Implemented:**

1. **Plugin Foundation** (Task 1)
   - Updated plugin.json with Approver2 details (ID, name, description, min server v6.0.0)
   - Configured go.mod with correct module path
   - Removed webapp folder for backend-only architecture
   - Build system verified and working

2. **Plugin Lifecycle Hooks** (Tasks 2-3)
   - Implemented `OnActivate()` with slash command registration
   - Implemented `registerCommand()` for /approve command
   - Implemented `ExecuteCommand()` with command routing
   - Added help text display for `/approve` and `/approve help`
   - Added error handling for unknown commands with suggestions

3. **Command Routing Architecture** (Task 4)
   - Created extensible `command/router.go` structure
   - Implemented Router pattern for future command handlers
   - Clean separation of concerns between plugin.go and routing logic

4. **Configuration Management** (Task 5)
   - Thread-safe configuration access with RWMutex
   - Empty configuration struct for MVP (as designed)
   - IsValid() validation method for future extensibility

5. **Comprehensive Testing** (Task 6)
   - Unit tests for OnActivate (success and failure cases)
   - Table-driven tests for ExecuteCommand
   - Router tests with multiple command scenarios
   - 30 tests total passing in 1.370s
   - Follows Mattermost testing patterns (testify, plugintest)

6. **Documentation** (Task 7)
   - Complete README.md with installation, usage, and development instructions
   - Clear feature status and roadmap
   - Build verification successful

**Acceptance Criteria Validation:**

✅ **AC1: Help Text Display** - `/approve` without arguments displays complete help text with all commands
✅ **AC2: Invalid Command Handling** - Unknown commands show error with valid command suggestions
✅ **AC3: Plugin Initialization** - Plugin registers `/approve` command successfully on activation

**Technical Achievements:**

- ✅ All Mattermost Go Style Guide conventions followed (CamelCase, error wrapping with %w, proper logging)
- ✅ Anti-patterns avoided (no pkg/util packages, no double-logging, no else after returns)
- ✅ Feature-based package structure (command/ not util/)
- ✅ Clean error messages with user input context
- ✅ Ephemeral responses for all command outputs

**Build & Test Results:**

```
Build: SUCCESS
Tests: 30 passed in 1.370s
Plugin Tarball: dist/com.mattermost.plugin-approver2-0.0.0+ed2dcd2.tar.gz
Cross-compilation: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

**Ready for Code Review:**

All Definition of Done criteria met:
- ✅ Plugin builds successfully
- ✅ All tests pass
- ✅ Plugin installable on Mattermost v6.0+
- ✅ `/approve` command registered
- ✅ Help text displays correctly
- ✅ Error handling works as specified
- ✅ README complete

**Next Story:** 1.2 - Approval Request Data Model & KV Storage
