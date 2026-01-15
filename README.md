# Mattermost Approval Workflow Plugin

Fast, authoritative approvals for time-sensitive decisions - entirely within Mattermost.

## Overview

This plugin enables teams to request and grant official approvals without leaving Mattermost. It creates immutable, auditable approval records that provide confidence to requesters, protection to approvers, and clarity to auditors.

## Features

- **`/approve new`** - Create approval requests via interactive modal
- **`/approve list`** - View all approval requests (submitted and received)
- **`/approve get [CODE]`** - View specific approval record details
- **`/approve cancel [CODE]`** - Cancel pending approval requests
- **`/approve verify [CODE]`** - Mark approved requests as verified
- **`/approve status`** - View approval statistics (admin only)
- **`/approve help`** - Display available commands and usage
- **Autocomplete** - Discover commands and arguments as you type
- **Immutable approval records** - Tamper-proof audit trail
- **DM notifications** - Automatic notifications to approvers
- **Interactive buttons** - One-click approve/deny with confirmation
- **Human-friendly codes** - Easy-to-reference IDs (e.g., A-X7K9Q2)
- **No external dependencies** - Pure Mattermost integration

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
/approve list [pending|approved|denied|canceled|all]
```
Shows all approval requests you've submitted or received. Use filters to see specific statuses.

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

## Releases & Versioning

### Versioning Strategy

This plugin follows [Semantic Versioning](https://semver.org/):
- **Major** (X.0.0) - Breaking changes requiring migration
- **Minor** (0.X.0) - New features, backward compatible
- **Patch** (0.0.X) - Bug fixes and minor improvements

Version numbers are managed via git tags. The build system automatically generates version numbers from tags.

### Release Process

**For Maintainers:**

Creating a new release:

```bash
# 1. Ensure you're on master branch and up to date
git checkout master
git pull origin master

# 2. Update CHANGELOG.md
#    - Move changes from [Unreleased] to new version section
#    - Add release date
#    - Update comparison links at bottom

# 3. Commit changelog
git add CHANGELOG.md
git commit -m "Prepare v0.2.0 release"
git push origin master

# 4. Create and push version tag (triggers CI build)
make minor              # For feature releases (0.X.0)
# OR
make patch              # For bug fixes (0.0.X)
# OR
make major              # For breaking changes (X.0.0)

# For release candidates, use:
make minor-rc           # Creates v0.2.0-rc1, v0.2.0-rc2, etc.
```

The Makefile commands will:
- Verify you're on a protected branch (master or release/*)
- Check for pending pulls from remote
- Prompt for confirmation
- Create a signed git tag
- Push the tag to trigger CI build

**CI/CD Pipeline:**

When a tag matching `v*` is pushed:
1. GitHub Actions runs tests
2. Builds plugin for all architectures
3. Creates release artifacts
4. Attaches `com.mattermost.plugin-approver2-X.Y.Z.tar.gz` to GitHub release

### Upgrade Guide

**Data Migration:**

The plugin uses a KV storage structure designed for backward compatibility:
- Key format: `approval_request:v1:{requestID}`
- Index keys: `index:{type}:{userID}:{timestamp}:{requestID}`

**Breaking Changes:**

When upgrading across major versions, check CHANGELOG.md for:
- Required data migrations
- API changes
- Configuration changes
- Minimum Mattermost version bumps

**Safe Upgrade Steps:**

1. Review CHANGELOG.md for your target version
2. Back up Mattermost database (contains KV store data)
3. Upload new plugin version via System Console
4. Plugin will automatically migrate data if needed
5. Verify existing approvals are accessible via `/approve list`

## Development Status

**MVP Implementation - Feature Complete**

**Completed Epics:**
- ✅ Epic 1: Approval Request Creation and Management
  - Plugin foundation and slash command registration
  - Data model and KV storage
  - Modal-based approval creation
  - Human-friendly reference codes (e.g., TUZ-2RK)
  - Approval cancellation

- ✅ Epic 2: Notifications and Approvals
  - DM notifications to approvers
  - Interactive approve/deny buttons
  - Confirmation modals
  - Immutable decision recording

- ✅ Epic 3: List and Query Functionality
  - Efficient indexing strategy
  - List all approvals (submitted and received)
  - Comprehensive test coverage

**Next Steps:**
- Production testing and user feedback
- Performance optimization for high-volume scenarios
- Additional features based on user requests

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
