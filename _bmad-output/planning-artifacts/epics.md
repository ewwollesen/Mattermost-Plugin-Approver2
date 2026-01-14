---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - '_bmad-output/planning-artifacts/prd.md'
  - '_bmad-output/planning-artifacts/architecture.md'
  - '_bmad-output/planning-artifacts/ux-design-specification.md'
workflowComplete: true
completionDate: 2026-01-10
---

# Mattermost-Plugin-Approver2 - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for Mattermost-Plugin-Approver2, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements (From PRD)

**Approval Request Management (FR1-FR5):**
- FR1: Users can create approval requests via slash command
- FR2: Each approval request must specify the approver (single user)
- FR3: Each approval request must include a description of what needs approval
- FR4: System generates unique, human-friendly approval reference codes
- FR5: Requester receives immediate confirmation when request is submitted

**Approval Decision Management (FR6-FR12):**
- FR6: Approver receives notification via DM when approval request is created
- FR7: Approver can approve or deny requests via interactive buttons
- FR8: Approver must confirm decision before it is finalized
- FR9: Approver can optionally add a comment when making a decision
- FR10: Once finalized, approval decisions are immutable (cannot be edited or deleted)
- FR11: System captures precise timestamp of approval creation and decision
- FR12: System captures full user identity (ID, username, display name) for requester and approver

**Notification & Communication (FR13-FR17):**
- FR13: Requester receives notification when approval is granted or denied
- FR14: All notifications are delivered via Mattermost direct messages
- FR15: Notifications contain complete context (who, what, when, decision)
- FR16: Notifications are delivered within 5 seconds of triggering event
- FR17: System tracks notification delivery attempts (does not guarantee delivery)

**Approval Record Management (FR18-FR24):**
- FR18: Users can retrieve their own approval requests via slash command
- FR19: Users can retrieve approval requests where they were the approver
- FR20: Users can retrieve specific approval by reference code
- FR21: Approval records contain all original request context
- FR22: Approval records are immutable once created
- FR23: Approval records include status (pending, approved, denied, canceled)
- FR24: Approval records are persisted in Mattermost plugin KV store

**User Interaction (FR25-FR29):**
- FR25: All interactions use native Mattermost UI components (modals, buttons, DMs)
- FR26: Slash commands provide help text when used incorrectly
- FR27: Error messages are specific and actionable
- FR28: Modal forms validate input before submission
- FR29: System provides immediate feedback for all user actions

**Identity & Permissions (FR36-FR40):**
- FR36: System authenticates users via Mattermost authentication
- FR37: Users can only view approval records where they were requester or approver
- FR38: Users can approve requests directed to them
- FR39: Requesters can cancel their own pending approval requests
- FR40: System enforces one-time state transitions (pending → approved/denied/canceled only)

### Non-Functional Requirements (From PRD)

**Performance (NFR-P1 to P4):**
- NFR-P1: Slash command response time < 2 seconds (modal open)
- NFR-P2: Approval submission response time < 2 seconds (confirmation shown)
- NFR-P3: Notification delivery time < 5 seconds (DM arrival)
- NFR-P4: Approval record retrieval < 3 seconds (list or specific record)

**Security (NFR-S1 to S6):**
- NFR-S1: All approval actions authenticated via Mattermost session
- NFR-S2: Users can only access their own approval records
- NFR-S3: Approval decisions cannot be spoofed or replayed
- NFR-S4: System prevents unauthorized approval record modification
- NFR-S5: Sensitive data (approval descriptions) never logged in plaintext
- NFR-S6: System uses Mattermost's existing security model (no custom auth)

**Reliability (NFR-R1 to R4):**
- NFR-R1: Approval records persisted atomically (all-or-nothing)
- NFR-R2: System handles concurrent approval requests without data loss
- NFR-R3: Failed notification delivery does not prevent record creation
- NFR-R4: System gracefully handles Mattermost API failures with clear error messages

**Usability (NFR-U1 to U4):**
- NFR-U1: First-time users can create approval request without documentation
- NFR-U2: All modals use self-explanatory field labels
- NFR-U3: Error messages explain what failed and how to fix
- NFR-U4: Approval reference codes are human-friendly (e.g., "A-X7K9Q2")

**Compatibility (NFR-C1 to C4):**
- NFR-C1: Plugin works on Mattermost Server v6.0+ (Plugin API v6.0+)
- NFR-C2: Plugin works across web, desktop (Win/Mac/Linux), mobile (iOS/Android)
- NFR-C3: Plugin requires no frontend customization (backend-only or minimal React)
- NFR-C4: Plugin persists data using Mattermost KV store only (no external database)

**Maintainability (NFR-M1 to M4):**
- NFR-M1: Code follows Mattermost Go Style Guide conventions
- NFR-M2: Test coverage minimum 70% for core approval logic
- NFR-M3: Error handling uses error wrapping with %w and sentinel errors
- NFR-M4: Schema versioning included for future data migrations

### Additional Requirements (From Architecture & UX)

**Architecture Requirements:**
- Based on mattermost-plugin-starter-template
- Go backend (primary), optional minimal React frontend
- Plugin API v6.0+ compatibility
- KV store persistence with hierarchical key prefixes
- ApprovalRecord data model with 26-char Mattermost ID (model.NewId())
- Immutability enforcement at write time (one-time state transition)
- Error handling with sentinel errors (ErrRecordNotFound, ErrRecordImmutable)
- Concurrent access strategy: last-write-wins with strict immutability
- Index strategy: per-user reference keys with timestamped keys
- Testing approach: standard Go testing with testify, table-driven tests
- Feature-based package organization (no pkg/util/misc anti-patterns)

**UX Requirements:**
- 100% Mattermost-native components (no custom UI)
- Responsive design inherited from Mattermost (works across all clients)
- Accessibility: WCAG 2.1 Level AA compliance inherited from Mattermost
- Keyboard navigation support for all interactions
- Screen reader support with semantic HTML and ARIA labels
- Mobile-first content design (concise, scannable, clear)
- Touch targets minimum 44x44px on mobile
- Error handling patterns: specific, actionable, graceful recovery
- Authority through structure and language (not custom styling)
- Confirmation dialogs for critical actions (prevent accidents)
- Structured message formatting with Markdown
- User mentions include full names (@alex (Alex Carter))
- Explicit timestamps (not relative: "2026-01-10 02:14:23 UTC")
- No information conveyed by emoji alone
- Clear button labels (descriptive verbs: "Approve", "Deny", "Confirm")
- All required fields marked with asterisk (*)

### FR Coverage Map

**Epic 1: Approval Request Creation & Management**
- FR1: Create approval requests via slash command
- FR2: Specify approver (single user)
- FR3: Include description
- FR4: Generate unique human-friendly reference codes
- FR5: Receive immediate confirmation
- FR25: Native Mattermost UI components
- FR26: Slash command help text
- FR27: Specific, actionable error messages
- FR28: Modal input validation
- FR29: Immediate feedback for all actions
- FR36: Mattermost authentication
- FR37: Access control (own records only)
- FR39: Cancel own pending requests

**Epic 2: Approval Decision Processing**
- FR6: Approver receives DM notification
- FR7: Approve or deny via interactive buttons
- FR8: Confirm decision before finalization
- FR9: Optional comment on decision
- FR10: Immutable decisions (no edits/deletes)
- FR11: Precise timestamp capture
- FR12: Full user identity capture
- FR13: Requester receives outcome notification
- FR14: Notifications via DM
- FR15: Notifications contain complete context
- FR16: Notifications delivered within 5 seconds
- FR17: Track notification delivery attempts
- FR38: Users can approve requests directed to them
- FR40: One-time state transitions enforced

**Epic 3: Approval History & Retrieval**
- FR18: Retrieve own approval requests
- FR19: Retrieve requests where user was approver
- FR20: Retrieve specific approval by reference code
- FR21: Records contain all original context
- FR22: Records are immutable
- FR23: Records include status
- FR24: Records persisted in KV store

**Coverage Verification:** All 40 FRs mapped (FR1-FR5, FR6-FR17, FR18-FR29, FR36-FR40) ✅

## Epic List

### Epic 1: Approval Request Creation & Management
Users can create approval requests via slash command, specify approvers, provide descriptions, receive unique reference codes, and cancel their own pending requests. All interactions use native Mattermost UI with immediate feedback and validation.

**FRs Covered:** FR1, FR2, FR3, FR4, FR5, FR25, FR26, FR27, FR28, FR29, FR36, FR37, FR39

### Epic 2: Approval Decision Processing
Approvers receive DM notifications of approval requests, review complete context, approve or deny requests with optional comments, and decisions are recorded immutably with precise timestamps. Requesters receive outcome notifications.

**FRs Covered:** FR6, FR7, FR8, FR9, FR10, FR11, FR12, FR13, FR14, FR15, FR16, FR17, FR38, FR40

### Epic 3: Approval History & Retrieval
Users can retrieve their own approval requests and decisions, search by reference code, and access complete immutable records for audit purposes with fast performance (<3 seconds).

**FRs Covered:** FR18, FR19, FR20, FR21, FR22, FR23, FR24

---

## Epic 1: Approval Request Creation & Management

Users can create approval requests via slash command, specify approvers, provide descriptions, receive unique reference codes, and cancel their own pending requests. All interactions use native Mattermost UI with immediate feedback and validation.

### Story 1.1: Plugin Foundation & Slash Command Registration

As a Mattermost user,
I want to type `/approve` and see available commands,
So that I can discover how to create approval requests without reading documentation.

**Acceptance Criteria:**

**Given** the plugin is installed and activated on a Mattermost server
**When** a user types `/approve` without arguments
**Then** the system displays help text showing available commands (`new`, `list`, `get`, `cancel`)
**And** the help text includes brief descriptions of each command
**And** the help text is returned as an ephemeral message (visible only to the user)

**Given** the plugin is installed and activated
**When** a user types `/approve invalid-command`
**Then** the system displays an error message stating the command is not recognized
**And** the system suggests valid commands

**Given** the Mattermost server is running Plugin API v6.0+
**When** the plugin initializes
**Then** the plugin registers the `/approve` slash command successfully
**And** the plugin uses the mattermost-plugin-starter-template structure
**And** the plugin implements Mattermost Go Style Guide conventions

**Covers:** FR26 (slash command help text), FR27 (specific error messages), NFR-C1 (Plugin API v6.0+), NFR-M1 (Mattermost conventions)

### Story 1.2: Approval Request Data Model & KV Storage

As a system,
I want to store approval requests in a structured format,
So that approval data is persisted reliably and can be retrieved efficiently.

**Acceptance Criteria:**

**Given** an approval request needs to be created
**When** the system creates an ApprovalRecord
**Then** the record contains all required fields:
- ID (26-char Mattermost ID via model.NewId())
- Code (human-friendly, e.g., "A-X7K9Q2")
- RequesterID, RequesterUsername, RequesterDisplayName
- ApproverID, ApproverUsername, ApproverDisplayName
- Description
- Status ("pending", "approved", "denied", "canceled")
- DecisionComment (optional)
- CreatedAt (epoch milliseconds)
- DecidedAt (0 if pending)
- RequestChannelID
- TeamID (optional)
- NotificationSent, OutcomeNotified (bool flags)
- SchemaVersion (integer, starts at 1)

**Given** an ApprovalRecord is created
**When** the system persists it to the KV store
**Then** the record is stored with hierarchical key prefix: `approval_records:{recordID}`
**And** the write operation is atomic (all-or-nothing)
**And** the record is stored as JSON

**Given** an ApprovalRecord is persisted
**When** the system needs to retrieve it by ID
**Then** the record is retrieved within 500ms
**And** the retrieved record matches the original exactly

**Given** the KV store write fails
**When** the system attempts to persist a record
**Then** the system returns an error with context (error wrapping with %w)
**And** no partial data is stored

**Covers:** FR4 (unique reference codes), FR11 (precise timestamps), FR12 (full user identity), FR22 (immutable records), FR24 (KV store persistence), NFR-R1 (atomic persistence), NFR-M3 (error wrapping), NFR-M4 (schema versioning)

### Story 1.3: Create Approval Request via Modal

As a requester,
I want to type `/approve new` and fill out a simple form,
So that I can quickly create an approval request without leaving my current context.

**Acceptance Criteria:**

**Given** a user types `/approve new` in any Mattermost context (channel, DM, thread)
**When** the command is processed
**Then** a modal dialog opens within 1 second (NFR-P1)
**And** the modal title is "Create Approval Request"
**And** the modal contains exactly two fields:
  - "Select approver *" (user selector, required)
  - "What needs approval? *" (textarea, required)
**And** the modal has "Submit Request" and "Cancel" buttons

**Given** the modal is open
**When** the user clicks the "Select approver *" field
**Then** a Mattermost user selector opens
**And** the selector allows searching and selecting any Mattermost user
**And** the selected user appears with @mention format

**Given** the modal is open
**When** the user clicks the "What needs approval? *" field
**Then** a textarea field accepts text input up to 1000 characters
**And** the field has placeholder text: "Describe the action requiring approval"

**Given** the user has filled both fields
**When** the user clicks "Submit Request"
**Then** the modal closes immediately
**And** the system processes the submission

**Given** the user clicks "Cancel"
**When** the modal is open
**Then** the modal closes without processing
**And** no approval request is created

**Covers:** FR1 (create via slash command), FR2 (specify approver), FR3 (include description), FR25 (native Mattermost UI), FR28 (modal validation), NFR-P1 (<2s response), NFR-U1 (no documentation needed), NFR-U2 (self-explanatory labels)

### Story 1.4: Generate Human-Friendly Reference Codes

As a system,
I want to generate unique, human-friendly approval reference codes,
So that users can easily reference and communicate approval IDs.

**Acceptance Criteria:**

**Given** a new approval request is being created
**When** the system generates an approval code
**Then** the code follows the format: `A-{6_CHARS}` (e.g., "A-X7K9Q2")
**And** the 6 characters are randomly generated alphanumeric (uppercase, excluding ambiguous chars like 0/O, 1/I/l)
**And** the code is unique across all existing approval records

**Given** a code generation produces a collision (extremely rare)
**When** the system detects the code already exists
**Then** the system regenerates a new code
**And** retries up to 5 times before failing with error

**Given** an approval record is created
**When** the system stores the record
**Then** both the ID (26-char) and Code (human-friendly) are stored
**And** users can reference the approval by either ID or Code

**Given** multiple approval requests are created concurrently
**When** codes are generated simultaneously
**Then** each code is unique (no collisions)
**And** code generation completes within 100ms

**Covers:** FR4 (generate unique human-friendly codes), NFR-U4 (human-friendly codes), NFR-R2 (concurrent request handling)

### Story 1.5: Request Validation & Error Handling

As a requester,
I want clear error messages when my approval request has problems,
So that I can fix issues and successfully submit my request.

**Acceptance Criteria:**

**Given** the user submits the modal without selecting an approver
**When** the system validates the submission
**Then** an error message displays: "Approver field is required. Please select a user."
**And** the modal remains open with user's description preserved
**And** the approver field is highlighted

**Given** the user submits the modal without entering a description
**When** the system validates the submission
**Then** an error message displays: "Description field is required. Please describe what needs approval."
**And** the modal remains open with selected approver preserved
**And** the description field is highlighted

**Given** the user selects an invalid or deleted user as approver
**When** the system validates the submission
**Then** an error message displays: "Selected approver is not a valid user. Please select an active user."
**And** the modal remains open for correction

**Given** the user enters a description exceeding 1000 characters
**When** the system validates the submission
**Then** an error message displays: "Description is too long (max 1000 characters). Please shorten your request."
**And** the modal remains open with description preserved

**Given** the KV store is unavailable
**When** the system attempts to save the approval request
**Then** an error message displays: "Failed to create approval request. The system is temporarily unavailable. Please try again."
**And** the user's data is not lost (can retry)

**Given** the Mattermost API returns an error during user lookup
**When** the system processes the approver selection
**Then** the system logs the error with context (error wrapping)
**And** displays a specific, actionable error message to the user

**Covers:** FR26 (slash command help), FR27 (specific, actionable errors), FR28 (modal validation), FR29 (immediate feedback), NFR-U3 (explain what failed and how to fix), NFR-R4 (graceful error handling), NFR-M3 (error wrapping)

### Story 1.6: Request Submission & Immediate Confirmation

As a requester,
I want immediate confirmation after submitting my approval request,
So that I know the request was successful and have the approval ID for reference.

**Acceptance Criteria:**

**Given** the user submits a valid approval request via the modal
**When** the system processes the submission
**Then** an ApprovalRecord is created with:
  - Status: "pending"
  - RequesterID: authenticated user's ID
  - RequesterUsername: authenticated user's username (snapshot)
  - RequesterDisplayName: authenticated user's display name (snapshot)
  - ApproverID: selected approver's ID
  - ApproverUsername: selected approver's username (snapshot)
  - ApproverDisplayName: selected approver's display name (snapshot)
  - Description: user's input
  - CreatedAt: current timestamp (epoch millis)
  - DecidedAt: 0 (pending)
  - RequestChannelID: channel where `/approve new` was typed
  - TeamID: team context if available
  - Unique ID and Code generated

**And** the record is persisted to the KV store atomically
**And** the operation completes within 2 seconds (NFR-P2)

**Given** the approval request is successfully saved
**When** the system confirms the creation
**Then** an ephemeral message is posted to the requester containing:
  - Header: "✅ Approval Request Submitted"
  - Approver: "@{approverUsername} ({approverDisplayName})"
  - Request ID: "{Code}" (e.g., "A-X7K9Q2")
  - Message: "You will be notified when a decision is made."
**And** the message appears within 1 second of submission
**And** the message is formatted with Markdown structure

**Given** the request is saved but confirmation message fails to send
**When** the DM delivery fails
**Then** the approval request still exists (data integrity prioritized)
**And** the error is logged for debugging
**And** the user sees a generic success indicator

**Given** the user's Mattermost session is authenticated
**When** the system creates the approval request
**Then** the system uses the authenticated user's identity (no separate auth)
**And** the system verifies the user has an active session

**Covers:** FR5 (immediate confirmation), FR11 (precise timestamps), FR12 (full user identity), FR29 (immediate feedback), FR36 (Mattermost authentication), NFR-P2 (<2s submission response), NFR-S1 (authenticated via Mattermost), NFR-S6 (Mattermost security model)

### Story 1.7: Cancel Pending Approval Requests

As a requester,
I want to cancel my own pending approval requests,
So that I can retract requests that are no longer needed.

**Acceptance Criteria:**

**Given** a user types `/approve cancel <ID>` where ID is a valid approval code
**When** the system processes the command
**Then** the system retrieves the approval record by code
**And** verifies the authenticated user is the original requester (NFR-S2)

**Given** the authenticated user is the requester and the status is "pending"
**When** the cancel command is executed
**Then** the approval record's Status field is updated to "canceled"
**And** the approval record's DecidedAt field is set to current timestamp
**And** the update is persisted atomically to the KV store
**And** the operation completes within 2 seconds

**Given** the cancellation is successful
**When** the system confirms the cancellation
**Then** an ephemeral message is posted to the requester: "✅ Approval request {Code} has been canceled."
**And** the message appears within 1 second

**Given** the authenticated user is NOT the requester
**When** the cancel command is executed
**Then** the system returns an error: "❌ Permission denied. You can only cancel your own approval requests."
**And** the approval record is not modified

**Given** the approval record status is "approved", "denied", or "canceled"
**When** the cancel command is executed
**Then** the system returns an error: "❌ Cannot cancel approval request {Code}. Status is already {Status}."
**And** the approval record is not modified (immutability enforced)

**Given** the user types `/approve cancel` without an ID
**When** the command is processed
**Then** the system returns help text: "Usage: /approve cancel <APPROVAL_ID>"

**Given** the user types `/approve cancel INVALID-ID`
**When** the command is processed
**Then** the system returns an error: "❌ Approval request 'INVALID-ID' not found. Use `/approve list` to see your requests."

**Covers:** FR39 (cancel own pending requests), FR37 (access control), FR40 (one-time state transitions), FR27 (specific error messages), NFR-S2 (users only access their own records), NFR-S4 (prevent unauthorized modification)

---

## Epic 2: Approval Decision Processing

Approvers receive DM notifications of approval requests, review complete context, approve or deny requests with optional comments, and decisions are recorded immutably with precise timestamps. Requesters receive outcome notifications.

### Story 2.1: Send DM Notification to Approver

As an approver,
I want to receive a DM notification when someone requests my approval,
So that I'm immediately aware of pending decisions without having to check elsewhere.

**Acceptance Criteria:**

**Given** an approval request is successfully created and persisted (Story 1.6)
**When** the system processes the request creation
**Then** a direct message is sent to the approver within 5 seconds (NFR-P3)
**And** the DM is sent from the plugin bot account
**And** the notification delivery attempt is tracked (NotificationSent flag set to true)

**Given** the approval request contains all required context
**When** the DM notification is constructed
**Then** the message includes:
  - Structured header: "📋 **Approval Request**"
  - Requester info: "**From:** @{requesterUsername} ({requesterDisplayName})"
  - Timestamp: "**Requested:** {timestamp in YYYY-MM-DD HH:MM:SS UTC format}"
  - Description section: "**Description:**\n{description}"
  - Request ID: "**Request ID:** {Code}"
**And** all user mentions use @username format for identity clarity
**And** the message uses Markdown formatting for structure

**Given** the DM is being sent
**When** the Mattermost API processes the DM
**Then** the DM appears in the approver's direct messages list
**And** the approver receives a notification according to their Mattermost notification preferences (desktop, mobile, email)

**Given** the DM send operation fails (network error, user doesn't exist, etc.)
**When** the notification attempt fails
**Then** the approval record remains valid (data integrity prioritized)
**And** the NotificationSent flag remains false
**And** the error is logged with full context for debugging
**And** the system does not retry automatically (manual investigation required)

**Given** the approver has DMs disabled or blocked the bot
**When** the notification is sent
**Then** the system handles the failure gracefully
**And** the approval record remains valid
**And** the error is logged

**Covers:** FR6 (approver receives DM), FR14 (notifications via DM), FR15 (notifications contain complete context), FR16 (notifications within 5 seconds), FR17 (track delivery attempts), NFR-P3 (<5s notification delivery), NFR-R3 (failed notification doesn't prevent record creation)

### Story 2.2: Display Approval Request with Interactive Buttons

As an approver,
I want to see a clearly structured approval request with action buttons,
So that I can review the context and make a decision without leaving the DM.

**Acceptance Criteria:**

**Given** the DM notification is being constructed (Story 2.1)
**When** the system adds interactive elements
**Then** two action buttons are included in the message:
  - [Approve] button with green styling (Mattermost default success color)
  - [Deny] button with red styling (Mattermost default danger color)
**And** both buttons are clearly visible and touch-friendly (minimum 44x44px on mobile per NFR)

**Given** the approval request notification is displayed
**When** the approver views the message
**Then** all context is visible without scrolling or navigation:
  - Who is requesting (full name and username)
  - What they need approval for (complete description)
  - When the request was made (precise timestamp)
  - Request ID for reference
**And** the structured format uses bold labels for scannability
**And** the message follows the UX design specification format

**Given** the message uses Markdown formatting
**When** the approver views it on different clients (web, desktop, mobile)
**Then** the formatting renders consistently across all platforms
**And** the structure remains readable on narrow mobile screens
**And** user mentions (@username) are properly highlighted

**Given** the approval request description is long (up to 1000 characters)
**When** the notification is displayed
**Then** the full description is visible
**And** line breaks are preserved
**And** the message remains readable

**Given** the approver is viewing the notification
**When** they need to understand what to do
**Then** the interface is self-explanatory (no documentation needed)
**And** the buttons clearly indicate available actions
**And** the context explains why they received this notification

**Covers:** FR7 (approve or deny via buttons), FR15 (notifications contain complete context), FR25 (native Mattermost UI), NFR-U1 (no documentation needed), NFR-U2 (self-explanatory), NFR-C2 (works across all clients), UX requirement (touch targets 44x44px), UX requirement (structured formatting)

### Story 2.3: Handle Approve/Deny Button Interactions

As an approver,
I want to click Approve or Deny and confirm my decision,
So that I can make deliberate, official decisions with confidence.

**Acceptance Criteria:**

**Given** an approver views an approval request notification with action buttons
**When** the approver clicks the [Approve] button
**Then** a confirmation modal opens immediately (<1 second)
**And** the modal title is "Confirm Approval"
**And** the modal body displays:
  - "Confirm you are approving:"
  - The original description (quoted)
  - Requester info: "@{requesterUsername} ({requesterDisplayName})"
  - Request ID
  - Finality warning: "This decision will be recorded and cannot be edited."
**And** the modal includes an optional "Add comment (optional)" textarea field
**And** the modal has [Confirm] and [Cancel] buttons

**Given** an approver views an approval request notification with action buttons
**When** the approver clicks the [Deny] button
**Then** a confirmation modal opens immediately (<1 second)
**And** the modal title is "Confirm Denial"
**And** the modal body displays the same structure as approval confirmation
**And** the modal includes an optional "Add comment (optional)" textarea field
**And** the modal has [Confirm] and [Cancel] buttons

**Given** the confirmation modal is open
**When** the approver clicks [Cancel]
**Then** the modal closes without processing
**And** no decision is recorded
**And** the original DM notification remains unchanged

**Given** the confirmation modal is open with optional comment field
**When** the approver enters a comment (up to 500 characters)
**Then** the comment is accepted and will be stored with the decision
**And** the comment field provides character count feedback

**Given** the approver clicks [Confirm] in the confirmation modal
**When** the system processes the confirmation
**Then** the modal closes
**And** the approval decision is recorded (Story 2.4)
**And** the original DM message is updated to show the decision was made

**Given** the approver has already made a decision on this request
**When** they attempt to click Approve or Deny again
**Then** the buttons are disabled (immutability enforced)
**And** an informational message shows: "Decision already recorded: {Status}"

**Given** the approval request has been canceled by the requester
**When** the approver attempts to click Approve or Deny
**Then** the buttons are disabled
**And** a message shows: "This request has been canceled"

**Covers:** FR7 (approve or deny via buttons), FR8 (confirm decision before finalization), FR9 (optional comment), FR10 (immutable decisions), FR38 (users can approve requests directed to them), UX requirement (confirmation dialogs prevent accidents), UX requirement (explicit language reinforces finality)

### Story 2.4: Record Approval Decision Immutably

As a system,
I want to record approval decisions with immutability guarantees,
So that decisions are authoritative, tamper-proof, and auditable.

**Acceptance Criteria:**

**Given** an approver confirms an approval decision (Story 2.3)
**When** the system processes the decision
**Then** the ApprovalRecord is retrieved from KV store by ID
**And** the system verifies the current Status is "pending"
**And** the system verifies the authenticated user is the designated approver (NFR-S2, NFR-S3)

**Given** the approval decision is verified
**When** the system updates the record
**Then** the following fields are updated atomically:
  - Status: "approved" (or "denied" for deny decision)
  - DecisionComment: {comment text or empty string}
  - DecidedAt: {current timestamp in epoch millis}
**And** all other fields remain unchanged
**And** the update completes within 2 seconds (NFR-P2)

**Given** the approval decision update is being persisted
**When** the system writes to the KV store
**Then** the write operation is atomic (all-or-nothing)
**And** the previous "pending" record is replaced completely
**And** the write includes optimistic locking to prevent concurrent modification

**Given** the approval record Status is already "approved", "denied", or "canceled"
**When** the system attempts to update it
**Then** the update is rejected with ErrRecordImmutable sentinel error
**And** an error message is returned: "Decision has already been recorded and cannot be changed."
**And** the existing record is not modified
**And** the error is logged

**Given** two approvers somehow attempt to decide on the same request simultaneously (edge case)
**When** both decisions are submitted
**Then** only the first decision is recorded (last-write-wins with immutability check)
**And** the second attempt fails with ErrRecordImmutable
**And** the second approver receives an error message

**Given** the authenticated user is NOT the designated approver
**When** they attempt to record a decision
**Then** the system rejects the operation with permission error
**And** returns: "Permission denied. Only the designated approver can make this decision."
**And** the record is not modified (NFR-S4)

**Given** the decision is successfully recorded
**When** the system confirms the update
**Then** the response includes the updated ApprovalRecord
**And** the response confirms the decision is now immutable

**Covers:** FR10 (immutable decisions), FR11 (precise timestamp), FR12 (full user identity), FR40 (one-time state transitions), NFR-R1 (atomic persistence), NFR-R2 (concurrent request handling), NFR-S2 (access control), NFR-S3 (prevent spoofing), NFR-S4 (prevent unauthorized modification), Architecture requirement (last-write-wins with immutability), Architecture requirement (sentinel errors)

### Story 2.5: Send Outcome Notification to Requester

As a requester,
I want to receive a DM notification when my approval request is decided,
So that I know immediately whether I can proceed with my action.

**Acceptance Criteria:**

**Given** an approval decision is successfully recorded (Story 2.4)
**When** the system processes the decision outcome
**Then** a direct message is sent to the requester within 5 seconds (NFR-P3)
**And** the DM is sent from the plugin bot account
**And** the OutcomeNotified flag is set to true

**Given** the approval request was approved
**When** the outcome DM is constructed
**Then** the message includes:
  - Header: "✅ **Approval Request Approved**"
  - Approver info: "**Approver:** @{approverUsername} ({approverDisplayName})"
  - Decision time: "**Decision Time:** {timestamp in YYYY-MM-DD HH:MM:SS UTC format}"
  - Request ID: "**Request ID:** {Code}"
  - Original request: "**Original Request:**\n> {description}" (quoted)
  - Decision comment (if provided): "**Comment:**\n{decisionComment}"
  - Action statement: "**Status:** You may proceed with this action."

**Given** the approval request was denied
**When** the outcome DM is constructed
**Then** the message includes:
  - Header: "❌ **Approval Request Denied**"
  - Same structure as approval notification
  - Action statement: "**Status:** This request has been denied."

**Given** the outcome notification includes all context
**When** the requester views the message
**Then** they can understand the decision without referring to external information
**And** the approval ID is clearly visible for reference
**And** the original request description is quoted for context
**And** the approver's identity is clear

**Given** the outcome DM is sent
**When** the Mattermost API processes it
**Then** the requester receives a notification according to their preferences
**And** the message appears in their direct messages

**Given** the DM send operation fails
**When** the notification attempt fails
**Then** the approval decision remains valid (data integrity prioritized)
**And** the OutcomeNotified flag remains false
**And** the error is logged for debugging
**And** the system does not retry automatically

**Given** the decision includes an optional comment
**When** the outcome notification is constructed
**Then** the comment is included in the message
**And** the comment is clearly attributed to the approver
**And** the comment is formatted for readability

**Covers:** FR13 (requester receives outcome notification), FR14 (notifications via DM), FR15 (notifications contain complete context), FR16 (notifications within 5 seconds), FR17 (track delivery attempts), NFR-P3 (<5s notification), UX requirement (structured format with quoted context), UX requirement (explicit status statements)

### Story 2.6: Handle Notification Delivery Failures

As a system,
I want to handle notification failures gracefully,
So that critical data (approval decisions) is never lost even if notifications fail.

**Acceptance Criteria:**

**Given** an approval request is created (Story 1.6)
**When** the approver notification fails to send
**Then** the ApprovalRecord remains valid with Status "pending"
**And** the NotificationSent flag is set to false
**And** the error is logged with full context:
  - Approval ID and Code
  - Approver ID
  - Error message from Mattermost API
  - Timestamp of failure
**And** the requester's confirmation still shows success (request was created)

**Given** an approval decision is recorded (Story 2.4)
**When** the requester outcome notification fails to send
**Then** the ApprovalRecord remains valid with the recorded decision
**And** the OutcomeNotified flag is set to false
**And** the error is logged with full context
**And** the decision is still immutable and official

**Given** notification delivery fails
**When** an admin or user needs to investigate
**Then** the approval record contains flags indicating which notifications failed:
  - NotificationSent: false means approver was never notified
  - OutcomeNotified: false means requester was never notified of outcome
**And** the admin can manually notify users if needed
**And** the system can query for records with failed notifications

**Given** the Mattermost API is temporarily unavailable
**When** notification attempts are made
**Then** the system does not retry automatically (avoid notification spam)
**And** the system does not block or rollback the approval operation
**And** the error includes actionable information for debugging

**Given** multiple notification attempts fail in sequence
**When** the system processes approval requests
**Then** each failure is logged independently
**And** the system continues processing other requests normally
**And** no cascading failures occur

**Given** a notification fails due to user-specific issues (DMs disabled, bot blocked)
**When** the error is logged
**Then** the log indicates the specific issue
**And** suggests resolution steps (e.g., "User has DMs disabled from bots")

**Given** the approval record has NotificationSent=false
**When** the system needs to prioritize manual intervention
**Then** admins can query for records needing notification
**And** retrieve a list of approvals where approvers were never notified

**Covers:** FR17 (track notification delivery attempts), NFR-R3 (failed notification doesn't prevent record creation), NFR-R4 (graceful error handling), UX requirement (graceful degradation), UX requirement (error logging for debugging)

---

## Epic 3: Approval History & Retrieval

Users can retrieve their own approval requests and decisions, search by reference code, and access complete immutable records for audit purposes with fast performance (<3 seconds).

### Story 3.1: List User's Approval Requests

As a user,
I want to list all my approval requests and approvals,
So that I can review my approval history and find specific records.

**Acceptance Criteria:**

**Given** a user types `/approve list` in any Mattermost context
**When** the system processes the command
**Then** the system retrieves all approval records where the authenticated user is either the requester or the approver
**And** the retrieval completes within 3 seconds (NFR-P4)
**And** the results are returned as an ephemeral message (visible only to the user)

**Given** the user has no approval records
**When** they execute `/approve list`
**Then** the system returns a message: "No approval records found. Use `/approve new` to create a request."
**And** the response completes within 1 second

**Given** the user has multiple approval records
**When** the system retrieves the records
**Then** the records are sorted by CreatedAt timestamp (most recent first)
**And** the list includes records where the user is the requester
**And** the list includes records where the user is the approver
**And** each record shows: Code, Status, Requester, Approver, CreatedAt (formatted as date)

**Given** the list contains more than 20 records
**When** the system displays the results
**Then** the first 20 records are shown
**And** a message indicates: "Showing 20 most recent records. Use `/approve get <ID>` for specific requests."
**And** the output remains readable and not truncated

**Given** the results are being formatted
**When** the system constructs the output
**Then** the output uses a structured table or list format:
```
**Your Approval Records:**

**A-X7K9Q2** | ✅ Approved | Requested by: @alex | Approver: @jordan | 2026-01-10 02:14
**A-Y3M5P1** | ⏳ Pending | Requested by: @alex | Approver: @morgan | 2026-01-09 14:23
**A-Z8N2K7** | ❌ Denied | Requested by: @chris | Approver: @alex | 2026-01-08 09:45
```
**And** each record is on a separate line for readability
**And** status is indicated with clear icons (✅ Approved, ❌ Denied, ⏳ Pending, 🚫 Canceled)

**Given** the KV store query fails
**When** the system attempts to retrieve records
**Then** an error message is returned: "❌ Failed to retrieve approval records. Please try again."
**And** the error is logged with context

**Given** the authenticated user session is valid
**When** the list command is executed
**Then** the system uses the authenticated user's ID to filter records (NFR-S1)
**And** no records from other users are included (NFR-S2)

**Covers:** FR18 (retrieve own approval requests), FR19 (retrieve where user was approver), FR21 (records contain all context), FR23 (records include status), FR37 (access control), NFR-P4 (<3s retrieval), NFR-S1 (authenticated), NFR-S2 (users see only their own records)

### Story 3.2: Retrieve Specific Approval by Reference Code

As a user,
I want to retrieve a specific approval by its reference code,
So that I can view the complete record for audit or reference purposes.

**Acceptance Criteria:**

**Given** a user types `/approve get <CODE>` where CODE is a valid approval code (e.g., "A-X7K9Q2")
**When** the system processes the command
**Then** the system retrieves the approval record by code
**And** verifies the authenticated user is either the requester or approver (NFR-S2)
**And** the retrieval completes within 3 seconds (NFR-P4)
**And** the result is returned as an ephemeral message

**Given** a user types `/approve get <ID>` where ID is a full 26-char Mattermost ID
**When** the system processes the command
**Then** the system retrieves the approval record by full ID
**And** performs the same access control verification
**And** returns the complete record

**Given** the approval record is retrieved and access is granted
**When** the system displays the record
**Then** the output shows all immutable record details (Story 3.3)
**And** the output includes:
  - Request ID (both Code and full ID)
  - Status (Pending/Approved/Denied/Canceled)
  - Requester (username and display name)
  - Approver (username and display name)
  - Description (full text)
  - Created timestamp (precise UTC format)
  - Decided timestamp (if decided)
  - Decision comment (if provided)
  - Request channel/team context

**Given** the authenticated user is NOT the requester or approver
**When** they attempt to retrieve the record
**Then** the system returns an error: "❌ Permission denied. You can only view approval records where you are the requester or approver."
**And** the record details are not shown

**Given** the approval code does not exist
**When** the user executes `/approve get INVALID-CODE`
**Then** the system returns an error: "❌ Approval record 'INVALID-CODE' not found."
**And** suggests: "Use `/approve list` to see your approval records."

**Given** the user types `/approve get` without providing a code
**When** the command is processed
**Then** the system returns help text: "Usage: /approve get <APPROVAL_ID>"
**And** includes an example: "Example: /approve get A-X7K9Q2"

**Given** multiple users reference the same approval code
**When** each user executes `/approve get <CODE>`
**Then** only the requester and approver can view the record
**And** all other users receive a permission denied error
**And** no information leakage occurs

**Given** the approval record contains sensitive information in the description
**When** the record is retrieved
**Then** the full description is shown only to authorized users (requester or approver)
**And** no logging of sensitive data occurs (NFR-S5)

**Covers:** FR20 (retrieve specific approval by code), FR21 (records contain all original context), FR22 (records are immutable), FR37 (access control), NFR-P4 (<3s retrieval), NFR-S2 (users see only their records), NFR-S5 (sensitive data not logged)

### Story 3.3: Display Approval Records in Readable Format

As a user,
I want approval records displayed in a clear, structured format,
So that I can quickly understand the approval context and decision.

**Acceptance Criteria:**

**Given** an approval record is being displayed (from `/approve get` or `/approve list`)
**When** the system formats the output
**Then** the output uses structured Markdown formatting with bold labels
**And** follows the UX design specification format for authority and clarity

**Given** a specific approval record is displayed via `/approve get`
**When** the system constructs the detailed view
**Then** the output includes:
```
**📋 Approval Record: A-X7K9Q2**

**Status:** ✅ Approved

**Requester:** @alex (Alex Carter)
**Approver:** @jordan (Jordan Lee)

**Description:**
Emergency rollback of payment-config-v2 deployment - causing 15% payment failures

**Requested:** 2026-01-10 02:14:23 UTC
**Decided:** 2026-01-10 02:15:45 UTC

**Decision Comment:**
Approved. Rollback immediately and investigate root cause.

**Context:**
- Channel: #incidents
- Team: Engineering
- Request ID: A-X7K9Q2
- Full ID: abc123def456ghi789jkl012mno

**This record is immutable and cannot be edited.**
```

**Given** the approval is still pending
**When** the record is displayed
**Then** the Status shows: "⏳ Pending"
**And** the "Decided" field shows: "Not yet decided"
**And** no decision comment is shown

**Given** the approval was denied
**When** the record is displayed
**Then** the Status shows: "❌ Denied"
**And** the decision comment (if provided) is shown

**Given** the approval was canceled
**When** the record is displayed
**Then** the Status shows: "🚫 Canceled"
**And** the "Decided" timestamp shows when it was canceled

**Given** the approval record has no decision comment
**When** the detailed view is displayed
**Then** the "Decision Comment" section is omitted (not shown as empty)

**Given** the description contains line breaks or formatting
**When** the record is displayed
**Then** the description formatting is preserved
**And** the output remains readable

**Given** the record includes all required context
**When** displayed to an auditor months later
**Then** the record is self-contained and complete (no external references needed)
**And** all identity information (usernames, display names) is preserved as snapshots
**And** timestamps are precise and unambiguous

**Given** the output is viewed on different clients (web, desktop, mobile)
**When** the Markdown formatting renders
**Then** the structure remains consistent and readable across all platforms
**And** the formatting works with both light and dark themes

**Covers:** FR21 (records contain all original context), FR22 (records are immutable), FR23 (records include status), UX requirement (structured formatting), UX requirement (explicit timestamps), UX requirement (user mentions with full names), UX requirement (immutability indicator)

### Story 3.4: Implement Index Strategy for Fast Retrieval

As a system,
I want to maintain efficient indexes for approval records,
So that users can retrieve their records quickly (<3 seconds) even with many approvals.

**Acceptance Criteria:**

**Given** an approval request is created (Story 1.6)
**When** the record is persisted to the KV store
**Then** the system creates the following keys:
  - Primary record: `approval_records:{recordID}` → full ApprovalRecord JSON
  - Requester index: `approval_index:requester:{requesterID}:{timestamp}:{recordID}` → recordID
  - Approver index: `approval_index:approver:{approverID}:{timestamp}:{recordID}` → recordID
  - Code lookup: `approval_code:{code}` → recordID
**And** all keys are written atomically as part of the record creation

**Given** a user executes `/approve list`
**When** the system queries for their records
**Then** the system performs two queries:
  - List all keys matching: `approval_index:requester:{userID}:*`
  - List all keys matching: `approval_index:approver:{userID}:*`
**And** extracts recordIDs from the index keys
**And** retrieves full records using the recordIDs
**And** the entire operation completes within 3 seconds (NFR-P4)

**Given** a user executes `/approve get A-X7K9Q2`
**When** the system retrieves the record
**Then** the system performs a direct lookup: `approval_code:A-X7K9Q2` → recordID
**And** retrieves the full record using: `approval_records:{recordID}`
**And** the operation completes within 1 second (faster than list)

**Given** the index keys use timestamp-based sorting
**When** records are listed
**Then** the natural key order returns most recent records first (descending timestamp)
**And** no additional sorting logic is required in application code

**Given** index keys avoid list-valued keys (NFR-R2, Architecture requirement)
**When** concurrent approval requests are created
**Then** no concurrent hotspot contention occurs (each key is unique)
**And** each write is independent (no read-modify-write cycle)

**Given** an approval record is updated (e.g., status changes from pending to approved)
**When** the record is updated
**Then** the primary record is updated: `approval_records:{recordID}`
**And** index keys remain unchanged (timestamp and IDs don't change)
**And** the next list query retrieves the updated record

**Given** the system needs to look up a record by full ID
**When** a user provides the full 26-char ID
**Then** the system retrieves directly: `approval_records:{fullID}`
**And** no index lookup is needed (direct key access)

**Given** the KV store has thousands of approval records
**When** a user lists their own records
**Then** only their index keys are queried (efficient prefix scan)
**And** the query does not scan all records in the system
**And** performance remains consistent regardless of total record count

**Covers:** FR18 (retrieve own requests), FR19 (retrieve where user was approver), FR20 (retrieve by code), FR24 (KV store persistence), NFR-P4 (<3s retrieval), NFR-R2 (concurrent request handling), Architecture requirement (per-user reference keys), Architecture requirement (timestamped keys), Architecture requirement (no list-valued keys to avoid hotspots)

### Story 3.5: Enforce Access Control on Retrieval

As a system,
I want to enforce strict access control on approval record retrieval,
So that users can only view records where they are authorized participants.

**Acceptance Criteria:**

**Given** a user executes `/approve list` or `/approve get <ID>`
**When** the system processes the query
**Then** the system retrieves the authenticated user's ID from the Mattermost session
**And** uses the authenticated ID for all access control checks (NFR-S1)

**Given** a user requests their approval list
**When** the system queries the KV store
**Then** only records matching the user's requester index or approver index are retrieved
**And** no records from other users are included in the query results
**And** the query filter is applied at the KV store level (not in application logic)

**Given** a user executes `/approve get <CODE>`
**When** the system retrieves the record
**Then** the system checks if the authenticated user's ID matches either:
  - record.RequesterID
  - record.ApproverID
**And** if neither matches, the retrieval is denied with permission error

**Given** a user attempts to retrieve another user's approval
**When** they provide a valid code but are not authorized
**Then** the system returns: "❌ Permission denied. You can only view approval records where you are the requester or approver."
**And** no record details are leaked in the error message
**And** the unauthorized attempt is logged for security auditing

**Given** a user tries to enumerate approval codes
**When** they execute multiple `/approve get` commands with guessed codes
**Then** each unauthorized attempt is denied
**And** no timing attacks are possible (consistent response time for not found vs. unauthorized)
**And** excessive failed attempts could be rate-limited (future consideration)

**Given** an admin or system user with elevated permissions
**When** they attempt to retrieve any approval record
**Then** the same access control rules apply (no special admin override in MVP)
**And** admins must be the requester or approver to view records

**Given** the authenticated user's session expires
**When** they attempt to retrieve approval records
**Then** the Mattermost API returns an authentication error
**And** no records are retrieved
**And** the user must re-authenticate

**Given** an approval record contains sensitive information
**When** access control is enforced
**Then** only the requester and approver can view the full description
**And** no other users can access the sensitive data (NFR-S2)
**And** the sensitive data is never logged in plaintext (NFR-S5)

**Given** the access control check fails for any reason
**When** the error occurs
**Then** the system logs the security event with:
  - Authenticated user ID
  - Attempted approval code/ID
  - Timestamp
  - Reason for denial
**And** no sensitive record data is included in logs

**Covers:** FR37 (access control - users only view their records), NFR-S1 (authenticated via Mattermost), NFR-S2 (users only access their records), NFR-S3 (prevent spoofing), NFR-S4 (prevent unauthorized modification), NFR-S5 (sensitive data not logged), NFR-S6 (Mattermost security model)

---

## Epic 6: Request Timeout & Verification (Feature Complete for 1.0)

Users can configure automatic timeout for unanswered approval requests (default 30 minutes), preventing abandoned requests from cluttering the system. Requesters can verify approved requests to confirm action completion, closing the approval loop with optional comments and notifying approvers.

### Story 6.1: Auto-Timeout for Pending Approval Requests

As a requester,
I want pending approval requests to automatically cancel after a timeout period,
So that unanswered requests don't clutter the system and I'm notified when requests expire.

**Acceptance Criteria:**

**Given** an approval request is created with status "pending"
**When** the system persists the record
**Then** a timeout timer starts based on the CreatedAt timestamp
**And** the default timeout duration is 30 minutes (system-wide configurable)

**Given** a pending approval request exists
**When** the current time exceeds (CreatedAt + 30 minutes)
**Then** the system automatically updates the record:
  - Status: "canceled"
  - CanceledReason: "Auto-canceled: No response within 30 minutes"
  - CanceledAt: current timestamp (epoch millis)
**And** the update is persisted atomically to the KV store

**Given** an approval request times out
**When** the system processes the auto-cancellation
**Then** a DM notification is sent to the requester within 5 seconds
**And** the DM includes:
  - Header: "⏱️ **Approval Request Timed Out**"
  - Request ID: "{Code}"
  - Original request: "**Original Request:**\n> {description}" (quoted)
  - Approver info: "**Approver:** @{approverUsername} ({approverDisplayName})"
  - Timeout info: "**Reason:** No response within 30 minutes"
  - Action statement: "**Status:** This request has been automatically canceled. You may create a new request if still needed."

**Given** an approval request times out
**When** the system sends the timeout notification
**Then** NO notification is sent to the approver (requester-only notification)
**And** the approver's original notification remains unchanged
**And** the approver can no longer approve/deny (buttons disabled with "This request has been canceled" message)

**Given** an approver clicks Approve or Deny on a pending request
**When** the decision is recorded BEFORE the timeout expires
**Then** the timeout timer stops immediately (race condition handling)
**And** the status becomes "approved" or "denied"
**And** the auto-timeout does NOT trigger for this request
**And** the requester receives the normal decision notification (not timeout notification)

**Given** the timeout timer and approver decision occur nearly simultaneously
**When** both operations attempt to update the record
**Then** only the first operation succeeds (last-write-wins with immutability check)
**And** if the approver decision arrives first, status becomes "approved"/"denied"
**And** if the timeout arrives first, status becomes "canceled"
**And** the second operation fails with ErrRecordImmutable
**And** no duplicate notifications are sent

**Given** an approval request has status "approved", "denied", or "canceled"
**When** the timeout period elapses
**Then** the auto-timeout does NOT trigger (only pending requests time out)
**And** the immutability rule prevents any modification

**Given** the system configuration allows customizing the timeout duration
**When** the timeout is configured (future enhancement - out of scope for MVP)
**Then** the timeout duration can be set system-wide (e.g., 15 minutes, 1 hour, 24 hours)
**And** the default remains 30 minutes if not configured
**Note:** MVP uses hardcoded 30 minute timeout; configuration is deferred to post-MVP

**Given** the timeout notification fails to send
**When** the DM delivery fails
**Then** the approval record remains canceled (data integrity prioritized)
**And** the error is logged for debugging
**And** the system does not retry automatically

**Given** a user executes `/approve list`
**When** the results include timed-out requests
**Then** the status shows: "🚫 Canceled"
**And** the CanceledReason indicates: "Auto-canceled: No response within 30 minutes"

**Given** a user executes `/approve get <CODE>` for a timed-out request
**When** the detailed view is displayed
**Then** the output includes:
  - Status: "🚫 Canceled"
  - Canceled timestamp
  - Cancellation reason: "Auto-canceled: No response within 30 minutes"
**And** the record is immutable (cannot be modified)

**Technical Implementation Notes:**
- Timeout check runs periodically (e.g., every 5 minutes via background goroutine or scheduled task)
- Query pending requests where: `Status == "pending" AND (CurrentTime - CreatedAt) > 30 minutes`
- Process batch of timed-out requests, update atomically, send notifications
- Index cleanup: remove from pending approver index, add to requester index with canceled status

**Covers:** NFR-R3 (graceful degradation), NFR-P3 (<5s notification), NFR-R1 (atomic persistence), NFR-R2 (concurrent request handling), UX requirement (structured formatting), Architecture requirement (immutability enforcement)

### Story 6.2: Verification Step for Approved Requests

As a requester,
I want to verify that I completed the approved action,
So that I can close the approval loop and notify the approver that the action is done.

**Acceptance Criteria:**

**Given** a requester has an approved request
**When** the requester types `/approve verify <CODE> [optional comment]`
**Then** the system retrieves the approval record by code
**And** verifies the authenticated user is the original requester (permission check)

**Given** the approval record has status "approved"
**When** the verify command is executed by the requester
**Then** the system updates the ApprovalRecord with new fields:
  - Verified: true (boolean)
  - VerifiedAt: current timestamp (epoch millis)
  - VerificationComment: {comment text or empty string}
**And** the Status field remains "approved" (verification is separate metadata)
**And** the update is persisted atomically to the KV store
**And** the operation completes within 2 seconds

**Given** the verification is successful
**When** the system confirms the verification
**Then** an ephemeral message is posted to the requester:
  - "✅ Verification recorded for approval request {Code}."
  - "The approver will be notified that the action is complete."
**And** the message appears within 1 second

**Given** a verification is recorded
**When** the system processes the verification
**Then** a DM notification is sent to the approver within 5 seconds
**And** the DM includes:
  - Header: "✅ **Action Verified Complete**"
  - Request ID: "{Code}"
  - Requester info: "**Requester:** @{requesterUsername} ({requesterDisplayName})"
  - Verification time: "**Verified:** {timestamp in YYYY-MM-DD HH:MM:SS UTC format}"
  - Original request: "**Original Request:**\n> {description}" (quoted)
  - Verification comment (if provided): "**Verification Comment:**\n{verificationComment}"
  - Action statement: "**Status:** The requester has confirmed this action is complete."

**Given** the verification command includes an optional comment
**When** the comment is provided
**Then** the comment is stored in VerificationComment field (max 500 characters)
**And** the comment is included in the approver notification
**And** the comment is validated before submission

**Given** the verification command does not include a comment
**When** the verify command is executed
**Then** VerificationComment is stored as empty string
**And** the "Verification Comment" section is omitted from the notification

**Given** a requester types `/approve verify` without providing a code
**When** the command is processed
**Then** the system returns help text:
  - "Usage: /approve verify <APPROVAL_ID> [optional comment]"
  - "Example: /approve verify A-X7K9Q2"
  - "Example: /approve verify A-X7K9Q2 Rollback completed successfully"

**Given** a requester types `/approve verify INVALID-CODE`
**When** the command is processed
**Then** the system returns an error: "❌ Approval request 'INVALID-CODE' not found. Use `/approve list` to see your requests."

**Given** the authenticated user is NOT the original requester
**When** they attempt to verify a request
**Then** the system returns an error: "❌ Permission denied. Only the requester can verify completion of an approval request."
**And** the approval record is not modified

**Given** the approval record status is "pending", "denied", or "canceled"
**When** the verify command is executed
**Then** the system returns an error: "❌ Cannot verify approval request {Code}. Only approved requests can be verified. (Current status: {Status})"
**And** the approval record is not modified

**Given** the approval record is already verified (Verified == true)
**When** the verify command is executed
**Then** the system returns an error: "❌ Approval request {Code} has already been verified on {VerifiedAt timestamp}."
**And** the existing verification is not modified (immutability - one-time verification only)

**Given** a user executes `/approve get <CODE>` for a verified request
**When** the detailed view is displayed
**Then** the output includes a new "Verification" section:
```
**Verification:**
- ✅ Verified on: 2026-01-10 15:30:00 UTC
- Comment: Rollback completed successfully
```
**And** the section appears after the "Decision Comment" section
**And** if not verified, the "Verification" section is omitted

**Given** a user executes `/approve list`
**When** the results include verified requests
**Then** the status shows: "✅ Approved" (same as unverified approved requests)
**And** verification status is NOT shown in the list view (only in detail view via `/approve get`)

**Given** the verification notification fails to send
**When** the DM delivery fails
**Then** the verification record remains valid (data integrity prioritized)
**And** the error is logged for debugging
**And** the system does not retry automatically
**And** the requester's confirmation still shows success

**Given** the verification comment exceeds 500 characters
**When** the system validates the input
**Then** an error message displays: "❌ Verification comment is too long (max 500 characters). Please shorten your comment."
**And** the verification is not recorded

**Technical Implementation Notes:**
- Add three new fields to ApprovalRecord struct:
  - Verified bool `json:"verified"`
  - VerifiedAt int64 `json:"verifiedAt"` (epoch millis, 0 if not verified)
  - VerificationComment string `json:"verificationComment"`
- Default values: Verified=false, VerifiedAt=0, VerificationComment=""
- Verification is one-time only: if Verified==true, reject subsequent verify attempts
- Verification does not change Status field (remains "approved")
- Schema version remains 1 (these fields are additive, backward compatible)

**Covers:** FR37 (access control), FR36 (Mattermost authentication), FR11 (precise timestamps), FR29 (immediate feedback), NFR-S2 (access control), NFR-P2 (<2s response), NFR-P3 (<5s notification), NFR-R1 (atomic persistence), UX requirement (structured formatting), UX requirement (explicit timestamps)

