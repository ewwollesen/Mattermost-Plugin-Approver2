# Changelog

All notable changes to the Mattermost Approval Workflow Plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial MVP implementation of approval workflow plugin
- `/approve new` - Create approval requests via interactive modal
- `/approve list` - View all approval requests (submitted and received)
- `/approve get [ID]` - View specific approval record details
- `/approve cancel [ID]` - Cancel pending approval requests
- `/approve help` - Display available commands
- Human-friendly approval reference codes (e.g., TUZ-2RK)
- DM notifications to approvers when requests are created
- Interactive approve/deny buttons in DM notifications
- Confirmation modals for approve/deny actions
- Immutable approval records stored in KV store
- Comprehensive test coverage for all core functionality
- Audit trail with timestamps and user information

### Technical
- Plugin initialization and slash command registration
- KV storage with composite key strategy for efficient querying
- Approval request data model with status management
- Intelligent indexing system for list operations
- Error handling and validation throughout

## [0.1.0] - TBD

Initial MVP release (pending completion of all Epic 1-3 stories).

---

## Version History Legend

### Types of Changes
- **Added** - New features
- **Changed** - Changes in existing functionality
- **Deprecated** - Soon-to-be removed features
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Security improvements or vulnerability patches
- **Technical** - Internal/developer-facing changes

### Compatibility Notes
- **Minimum Mattermost version**: v6.0.0
- **Minimum Go version**: 1.19+
- **Breaking changes** will always be documented in the CHANGELOG with upgrade instructions

[Unreleased]: https://github.com/mattermost/mattermost-plugin-approver2/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mattermost/mattermost-plugin-approver2/releases/tag/v0.1.0
