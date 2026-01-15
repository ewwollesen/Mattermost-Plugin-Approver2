# Mattermost Approval Workflow Plugin

Fast, authoritative approvals for time-sensitive decisions - entirely within Mattermost.

## Why Use This Plugin?

Get formal approval without email chains, paper forms, or leaving your chat workspace. This plugin creates **immutable, auditable approval records** that provide confidence to requesters, protection to approvers, and clarity to auditors.

**Perfect for:**

- Deploy approvals and production changes
- Budget sign-offs and purchase requests
- Emergency access to restricted systems
- Policy exception requests
- Any decision that needs a paper trail

## Quick Start

### Installation

1. **Download the plugin**
   - Visit the [GitHub Releases page](https://github.com/ewwollesen/Mattermost-Plugin-Approver2/releases)
   - Download the latest `com.mattermost.plugin-approver2-X.Y.Z.tar.gz`

2. **Install in Mattermost**
   - Go to **System Console → Plugin Management**
   - Click **Upload Plugin**
   - Select the downloaded tarball
   - Click **Enable** after upload completes

3. **Start using it**
   - Type `/approve help` in any channel
   - Create your first approval with `/approve new`

**Requirements:** Mattermost Server v6.0+

### Your First Approval

```
/approve new
```

This opens a modal where you:

1. Select an approver from your team
2. Describe what needs approval
3. Submit the request

The approver receives an instant DM with **Approve** and **Deny** buttons. They click their decision, confirm in a modal, and you get notified immediately. Every step is recorded with timestamps.

## Features

### Core Functionality

- **Interactive approval requests** - Modal-based creation with user picker
- **Instant DM notifications** - Approvers get one-click approve/deny buttons
- **Human-friendly reference codes** - Easy-to-share IDs like `TUZ-2RK` or `A-X7K9Q2`
- **Professional table output** - Results displayed in clean, readable markdown tables
- **Slash command autocomplete** - Discover commands and options as you type
- **Immutable audit trail** - Approval records can't be modified after creation
- **No external dependencies** - Pure Mattermost integration using KV store

### Advanced Features

- **List filtering** - View pending, approved, denied, or canceled requests
- **Request cancellation** - Cancel pending requests with predefined reasons
- **Verification workflow** - Mark approved requests as verified for compliance tracking
- **Automatic timeouts** - Stale pending approvals timeout after configurable period
- **Admin statistics** - System-wide metrics for approval usage (admin-only)
- **Cancellation notifications** - Approvers get notified when requests are canceled

## How It Works

### The Approval Lifecycle

1. **Request Created** - Requester fills out modal, selects approver, submits
2. **Notification Sent** - Approver receives DM with request details and action buttons
3. **Decision Made** - Approver clicks Approve/Deny, confirms in modal
4. **Outcome Recorded** - Decision saved immutably with timestamp
5. **Requester Notified** - Requester receives DM with approval decision
6. **Optional Verification** - Approved requests can be marked as verified

At any point, either party can view the approval record with `/approve get [CODE]`.

## Usage Guide

### Creating Approval Requests

**Command:** `/approve new`

Opens an interactive modal with:

- **Approver** - Select from your Mattermost team members
- **Request Details** - Describe what needs approval

After submission, you receive a unique reference code (e.g., `TUZ-2RK`) that you can share or use to check status.

**What happens next:**

- Approver receives DM notification with your request
- DM includes **Approve** and **Deny** buttons
- You can check status with `/approve list` or `/approve get TUZ-2RK`

### Managing Your Requests

**List all your approvals:**

```
/approve list
```

Shows pending requests by default. Use filters to see other statuses:

```
/approve list pending      # Only pending requests (default)
/approve list approved     # Only approved requests
/approve list denied       # Only denied requests
/approve list canceled     # Only canceled requests
/approve list all          # Everything
```

Results display in a professional markdown table:

| Code | Status | Requestor | Approver | Created |
|------|--------|-----------|----------|---------|
| TUZ-2RK | Pending | @wayne | @jane | 2026-01-15 09:30 |
| A-X7K9Q2 | Approved | @wayne | @john | 2026-01-14 15:45 |

**View specific approval:**

```
/approve get TUZ-2RK
```

Shows complete details including:

- Request and approver
- Status and timestamps
- Decision reason (if denied)
- Cancellation reason (if canceled)
- Verification status

**Cancel a pending request:**

```
/approve cancel TUZ-2RK
```

Opens a dialog where you select a cancellation reason:

- Requirements changed
- No longer needed
- Wrong approver
- Duplicate request
- Other (with additional details field)

Canceling notifies the approver via DM and updates the request record.

### Approving Requests

When you receive an approval request via DM:

1. **Review the request details** in the notification
2. **Click Approve or Deny** button
3. **Confirm your decision** in the modal (optional: add reason for denial)
4. **Requester gets notified** of your decision automatically

The approval record is immediately updated and immutable.

### Verification Workflow

After an approval is granted, mark it as verified when the approved action is completed:

```
/approve verify TUZ-2RK
```

This creates an additional audit trail step useful for:

- Confirming that approved deployments actually happened
- Compliance tracking for policy exceptions
- Post-approval validation requirements

### Admin Features

**System statistics:**

```
/approve status
```

Admins can view system-wide metrics:

- Total approval requests
- Breakdown by status (pending, approved, denied, canceled)
- Verification statistics
- Timeout information

**Configuration** (via System Console):

- Request timeout period (default: configurable)
- Plugin enable/disable

## Common Scenarios

### Scenario 1: Production Deployment Approval

```
You: /approve new
[Modal: Approver: @lead-engineer, Details: "Deploy v2.1.0 to production"]
System: "Approval request TUZ-2RK created"

@lead-engineer receives DM with deployment details
@lead-engineer clicks "Approve" → confirms

You receive: "Your approval request TUZ-2RK was APPROVED"
You: /approve verify TUZ-2RK (after deployment completes)
```

### Scenario 2: Emergency Access Request

```
You: /approve new
[Modal: Approver: @security-manager, Details: "Emergency read access to prod DB to debug P0 incident"]
System: "Approval request A-X7K9Q2 created"

@security-manager receives DM
@security-manager clicks "Approve" → confirms with reason: "Approved for 2 hours only"

You receive notification with approval
[Complete your work]
You: /approve verify A-X7K9Q2 (confirming access was revoked)
```

### Scenario 3: Canceled Request

```
You: /approve new
[Modal: Approver: @finance-director, Details: "Purchase $5,000 software license"]
System: "Approval request B-3M8PN created"

[You find out the software is already purchased]
You: /approve cancel B-3M8PN
[Select reason: "Duplicate request"]

@finance-director receives cancellation notification
Request marked as canceled in system
```

## FAQ & Troubleshooting

### General Questions

**Q: Can I have multiple approvers for one request?**
A: Not currently. Each approval request has one approver. For multi-stage approvals, create sequential approval requests.

**Q: Can I edit an approval request after submitting?**
A: No. Approval records are immutable for audit integrity. If you need changes, cancel the original and create a new request.

**Q: How long do approval records stay in the system?**
A: Indefinitely. All records are stored in Mattermost's KV store and remain accessible via `/approve list` and `/approve get`.

**Q: Can anyone see all approval requests?**
A: No. You can only see requests where you are the requester or the approver. Admins do not have access to view individual approvals (only aggregate statistics via `/approve status`).

**Q: What happens to pending requests when someone leaves the team?**
A: Requests remain in the system. You can cancel them if the approver is no longer available, then create new requests with a different approver.

### Troubleshooting

**Q: The approver didn't receive the DM notification. What should I do?**
A: Possible causes:

- Check if DMs are blocked between your accounts
- Verify the approver has DMs enabled in their settings
- Check the approver's notification settings
- Have the approver run `/approve list` to see if the request appears

**Q: I canceled a request but the buttons still show in the approver's DM. Can they still approve it?**
A: The buttons are updated to show "Canceled" when someone clicks them. Approvers cannot approve or deny a canceled request - the system will reject the action.

**Q: Can I cancel an already-approved request?**
A: No. Once approved or denied, requests cannot be canceled. This preserves audit integrity. If you need to reverse a decision, create a new approval request for the reversal action.

**Q: What happens when a request times out?**
A: After the configured timeout period, pending requests are automatically marked as timed out. Both the requester and approver receive notifications. Timed-out requests appear in the "all" filter view.

**Q: I'm getting an error when trying to create an approval. What should I check?**
A: Common issues:

- Verify the plugin is enabled (System Console → Plugin Management)
- Ensure you're selecting a valid user as the approver
- Check that you have permission to use slash commands
- Try `/approve help` to verify the plugin is responding

**Q: Can I approve my own requests?**
A: Technically yes (the plugin doesn't prevent it), but this defeats the purpose of approvals and creates a poor audit trail. Always use a different person as the approver.

### Performance

**Q: How many approval requests can the plugin handle?**
A: The plugin uses an efficient indexing strategy for fast retrieval. It's designed to handle thousands of requests per user without performance degradation. The KV store scales with your Mattermost installation.

## Installation & Configuration

### Prerequisites

- **Mattermost Server:** v6.0 or later
- **Plugin API:** v6.0+
- **Permissions:** System Admin rights to install plugins

### Detailed Installation Steps

1. **Download the release artifact**
   ```
   wget https://github.com/ewwollesen/Mattermost-Plugin-Approver2/releases/download/v1.0.0/com.mattermost.plugin-approver2-1.0.0.tar.gz
   ```

2. **Upload via System Console**
   - Navigate to **System Console → Plugins → Plugin Management**
   - Click **Upload Plugin**
   - Select the tarball file
   - Click **Upload**

3. **Enable the plugin**
   - Find "Approval Workflow Plugin" in the plugin list
   - Toggle **Enable Plugin** to **true**
   - Plugin activates immediately (no restart required)

4. **Verify installation**
   - Open any Mattermost channel
   - Type `/approve` and you should see autocomplete suggestions
   - Run `/approve help` to confirm the plugin is working

### Configuration Options

Currently, the plugin works out-of-the-box with sensible defaults. Future versions may add:

- Configurable timeout periods
- Custom approval reasons
- Webhook integrations
- Custom reference code formats

### Upgrading

**From v0.x to v1.0.0:**

1. Review the [CHANGELOG](CHANGELOG.md) for new features
2. Download the v1.0.0 release
3. Upload via System Console (overwrites previous version)
4. No data migration required - all existing approvals remain accessible
5. New features activate immediately

**Safe upgrade process:**

- Back up your Mattermost database before major version upgrades
- Test in a staging environment if available
- Verify existing approvals are accessible after upgrade with `/approve list`

## Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/ewwollesen/Mattermost-Plugin-Approver2.git
cd Mattermost-Plugin-Approver2

# Build plugin
make

# Run tests (443 passing tests)
make test

# Deploy to local Mattermost
# Requires: MM_SERVICESETTINGS_SITEURL and MM_ADMIN_TOKEN environment variables
make deploy
```

### Project Structure

```
server/
  plugin.go              # Main plugin hooks and initialization
  configuration.go       # Configuration management
  command/
    router.go           # Command routing and handler logic
  store/
    store.go            # KV storage interface
    approval.go         # Approval data model and operations
webapp/
  src/
    index.js            # Plugin entry point
    components/         # React components for modals
```

### Contributing

Contributions are welcome! Please:

- Ensure all tests pass (`make test`)
- Follow the [Mattermost Go Style Guide](https://developers.mattermost.com/contribute/style-guide/)
- Add tests for new features
- Update CHANGELOG.md with your changes
- Write clear commit messages

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines (if available).

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test
go test ./server/command -run TestExecuteList
```

### Release Process for Maintainers

**Creating a release:**

```bash
# Update CHANGELOG.md with release notes
# Commit the changelog
git add CHANGELOG.md
git commit -m "Prepare vX.Y.Z release"

# Create and push version tag
make minor              # For feature releases (0.X.0)
make patch              # For bug fixes (0.0.X)
make major              # For breaking changes (X.0.0)
```

The Makefile handles tag creation and pushing. GitHub Actions automatically builds release artifacts.

## Release Information

**Current Version:** v1.0.0 (Released 2026-01-15)

**Download:** [GitHub Releases](https://github.com/ewwollesen/Mattermost-Plugin-Approver2/releases)

**What's New in v1.0.0:**

- Professional markdown table output for approval lists
- Slash command autocomplete
- Request timeout system
- Verification workflow for approved requests
- Cancellation reasons with predefined options
- Admin statistics command
- List filtering by status

**All 7 planned epics completed - Production ready!**

See [CHANGELOG.md](CHANGELOG.md) for complete version history.

### Versioning

This plugin follows [Semantic Versioning](https://semver.org/):

- **Major** (X.0.0) - Breaking changes requiring migration
- **Minor** (0.X.0) - New features, backward compatible
- **Patch** (0.0.X) - Bug fixes and minor improvements

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

**Issues and Questions:**

- [GitHub Issues](https://github.com/ewwollesen/Mattermost-Plugin-Approver2/issues) - Bug reports and feature requests
- [Mattermost Community](https://community.mattermost.com/) - General discussion

**Security Issues:**

Please report security vulnerabilities privately via GitHub Security Advisories or email the maintainer directly.

---

**Made with care for teams that value accountability and speed.**
