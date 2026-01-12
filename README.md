# Mattermost Approval Workflow Plugin

Fast, authoritative approvals for time-sensitive decisions - entirely within Mattermost.

## Overview

This plugin enables teams to request and grant official approvals without leaving Mattermost. It creates immutable, auditable approval records that provide confidence to requesters, protection to approvers, and clarity to auditors.

## Features (MVP)

- `/approve new` - Create approval requests via slash command
- `/approve list` - View your approval history
- `/approve get [ID]` - View specific approval records
- `/approve cancel [ID]` - Cancel pending requests
- `/approve help` - Display available commands
- Immutable approval records
- DM-based notifications (coming in Story 2.x)
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
Shows all approval requests you've submitted or received. (Coming in Story 3.1)

### Get Help
```
/approve help
```
Displays available commands and usage information.

## Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/mattermost/mattermost-plugin-approver2.git
cd mattermost-plugin-approver2

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

## Development Status

This is an MVP implementation. Current story: **1.1 - Plugin Foundation & Slash Command Registration**

**Completed:**
- ✅ Plugin initialization and activation
- ✅ Slash command registration (`/approve`)
- ✅ Help text display
- ✅ Error handling for unknown commands

**Upcoming Stories:**
- 1.2 - Approval Request Data Model & KV Storage
- 1.3 - Create Approval Request via Modal
- 1.4 - Generate Human-Friendly Reference Codes
- And more...

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please ensure:
- All tests pass (`make test`)
- Code follows Mattermost Go Style Guide
- Changes are documented in commit messages

## Support

For issues and questions:
- [GitHub Issues](https://github.com/mattermost/mattermost-plugin-approver2/issues)
- [Mattermost Community](https://community.mattermost.com/)
