# Changelog

All notable changes to the Mattermost Approval Workflow Plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Nothing yet

## [1.0.0] - 2026-01-15

🎉 **Production-Ready Release** - Feature Complete for 1.0!

This release marks the completion of all planned MVP features with significant UX polish, making the plugin production-ready for team use.

### Added
- **Slash command autocomplete** - Discover commands and arguments as you type (#7.4)
- **Request timeout system** - Automatically timeout stale pending approvals after configurable period (#6.1)
- **Verification workflow** - Mark approved requests as verified for audit compliance (`/approve verify`) (#6.2)
- **Cancellation reasons** - Dropdown with predefined reasons when canceling requests (#4.3)
- **Cancellation notifications** - DMs sent to approvers when requests are canceled (#4.2, #7.1)
- **Cancellation details** - Optional additional context field for complex cancellations (#7.3)
- **Admin statistics command** - `/approve status` shows system-wide metrics (admin-only) (#3.5)
- **List filtering** - Filter by status: pending, approved, denied, canceled, all (#5.1)
- **Professional table output** - `/approve list` displays results in clean markdown tables (#7.5)

### Changed
- **Default list view** - Changed from "all" to "pending" for better focus on actionable items (#5.2)
- **List output format** - Converted to markdown tables for professional appearance (#7.5)
- **Canceled request sorting** - Moved to bottom of "all" view for de-prioritization (#4.6)
- **Ghost button prevention** - Cancel dialogs properly update approver DM posts (#4.7)
- **Empty state messages** - Filter-specific guidance (e.g., "No pending requests. Use /approve list all") (#5.2)
- **Command delivery** - List command uses ephemeral posts for proper markdown rendering (#7.5)

### Fixed
- **CI/CD pipeline** - Removed duplicate builds on tag pushes (#7.2)
- **Cancellation notification bugs** - Requestor now receives proper notification when request is canceled (#7.1)
- **Additional details dialog** - Fixed text box visibility when "Other" reason selected (#7.3)
- **Error handling consistency** - All SendEphemeralPost calls now have fallback logic (#7.5)

### Technical
- Enhanced test coverage: 443 passing tests
- Improved error handling with consistent fallback patterns
- Code review process integrated with automated fixes
- Comprehensive documentation for all features
- Production-ready code quality and maintainability

### Epic Summary
- ✅ Epic 1: Approval Request Creation & Management
- ✅ Epic 2: Approval Decision Processing
- ✅ Epic 3: Approval History & Retrieval
- ✅ Epic 4: Improved Cancellation UX + Audit Trail
- ✅ Epic 5: List Filtering and Default View Optimization
- ✅ Epic 6: Request Timeout & Verification (Feature Complete for 1.0)
- ✅ Epic 7: 1.0 Polish & UX Improvements

### Breaking Changes
None - fully backward compatible with 0.1.0

### Upgrade Notes
- No data migration required
- All existing approval records remain accessible
- New features activate immediately upon plugin upgrade

## [0.1.0] - 2026-01-12

Initial MVP release with complete approval workflow functionality.

### Added
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

[Unreleased]: https://github.com/mattermost/mattermost-plugin-approver2/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/mattermost/mattermost-plugin-approver2/releases/tag/v1.0.0
[0.1.0]: https://github.com/mattermost/mattermost-plugin-approver2/releases/tag/v0.1.0
