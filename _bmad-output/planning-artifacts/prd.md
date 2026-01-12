---
stepsCompleted: [1, 2, 3, 4, 6, 7, 8, 9, 10, 11]
inputDocuments:
  - '_bmad-output/planning-artifacts/product-brief-Mattermost-Plugin-Approver2-2026-01-09.md'
documentCounts:
  productBrief: 1
  research: 0
  brainstorming: 0
  projectDocs: 0
workflowType: 'prd'
lastStep: 11
completedDate: 2026-01-10
date: 2026-01-09
author: Wayne
---

# Product Requirements Document - Mattermost-Plugin-Approver2

**Author:** Wayne
**Date:** 2026-01-09

## Executive Summary

The Mattermost Approval Workflow Plugin solves a critical but overlooked problem: **teams need to make authorized decisions quickly, but lack a way to formalize those decisions that's both fast enough for urgent situations and authoritative enough to rely on later.**

Organizations currently accept hidden risk by relying on informal chat approvals. While direct messages and channel posts enable quick decisions, they are non-authoritative and difficult to defend during audits, compliance reviews, or after-action analysis. Conversely, formal workflow systems (ServiceNow, ITSM, JIRA) are too heavyweight for ad-hoc, time-sensitive scenarios where authority matters more than process.

**This plugin creates "bridge authorization"** - official enough to act on immediately, but not intended to replace downstream formal systems. It fills the gap between informal decision-making and formal accountability, giving teams a safe way to move fast.

The core question teams ask is: **"Can I safely act on this?"** Informal approvals don't provide confidence; formal systems don't provide speed. This plugin provides both.

**Target Users:**
- **Requesters (Alex Carter)** - Frontline operators who need explicit authorization beyond their normal authority (sales associates, military personnel, incident responders)
- **Approvers (Jordan Lee)** - Decision-makers who must balance speed with judgment and need defensible records of their decisions
- **Auditors (Morgan Patel)** - Compliance officers who verify approvals after the fact and need clear, unambiguous records

**Primary Value:**
- Psychological safety for requesters (confidence to act)
- Procedural safety for organizations (defensible records)
- Clarity for auditors (verification without reconstruction)

### What Makes This Special

**1. True Mattermost-Native Integration**

Being native to Mattermost provides advantages external tools cannot replicate:
- Deep awareness of users, channels, roles, and context
- First-class notifications (DMs, mentions, threads) without integration friction
- Trust inherited from the same system already used for mission-critical communication
- Native slash commands, interactive messages, and permission model
- Native deployment model (on-prem/air-gapped)

For Defense, Intelligence, Security, and Critical Infrastructure (DISC) customers, this means:
- No external SaaS dependencies
- No data egress
- No new trust boundary

**2. Intentional Simplicity at the Approval Moment**

The core approval experience is intentionally lightweight and friction-free. Requesting or granting approval takes seconds and requires no predefined workflows or heavy configuration.

As the product evolves, optional configuration (templates, roles, chains) may be introduced increasing admin overhead slightly — but **simplicity at the moment of request and approval remains a non-negotiable design principle.**

Configuration/setup may grow over time, but the act of requesting and granting approval must never feel heavy.

**3. "Bridge Authorization" Positioning**

This is not a ServiceNow replacement, Playbooks competitor, or compliance platform. It occupies a deliberately narrow but critical space: the moment when authority is needed immediately, and formal documentation can follow later.

The key insight: **formal and informal approvals are not opposites—they're adjacent.** Organizations don't fail because they lack formal workflows. They fail in the gap between decisions made in real time and records created after the fact.

This plugin formalizes just enough of the informal moment to make it safe.

**4. Right Timing with Converging Forces**

Three factors make this the right time:
- **Chat has become the operational backbone** - distributed and hybrid work has made chat the primary coordination layer
- **Compliance pressure is increasing even for fast teams** - more teams must justify decisions without slowing operations
- **Mattermost Playbooks have matured** - with repeatable workflows well-covered, the remaining gap is ad-hoc decision-making

## Project Classification

**Technical Type:** Developer Tool (Mattermost Plugin/Extension)
**Primary Domain:** General (with strong applicability to GovTech/Defense)
**Complexity:** Low to Medium
**Project Context:** Greenfield - new plugin

**Technical Characteristics:**
- Mattermost plugin architecture (Go backend, React frontend)
- Slash command interface
- Interactive modals and message buttons
- Plugin KV store for data persistence
- DM-based notifications
- Integration with Mattermost user directory and permissions

**Domain Considerations:**
While the plugin architecture itself is straightforward, serving DISC (Defense, Intelligence, Security, Critical Infrastructure) customers as the primary market means careful attention to:
- Security and audit requirements
- On-premise and air-gapped deployment support
- No external dependencies or data egress
- Immutable, tamper-evident record keeping

**Market Applicability:**
Primary focus is DISC customers, but the fundamental problem of fast, authoritative approvals applies broadly to:
- Commercial enterprises needing expense or operational exception approvals
- Startups and SMBs wanting lightweight approval processes without heavyweight ITSM
- Any organization where chat is the operational backbone but informal approvals create risk

The DISC focus shapes the design (security, no external dependencies, audit-grade records), but those same qualities make it valuable for a much wider market.

## Success Criteria

### User Success

Success for the Mattermost Approval Workflow Plugin is measured by **behavioral change**: when something matters, people stop using DMs and start using the approval tool.

This behavioral shift validates that all three personas are getting value simultaneously:

**Alex Carter (Requester) - Confidence to Act**

**Desired Outcome:** Move forward quickly without uncertainty or personal risk. Success is knowing that the approval is explicit, authoritative, and will stand up later.

**Key "Aha" Moment:** The first time Alex receives a clear approval notification, proceeds with an action, and doesn't feel the need to document or defend themselves afterward. Alex realizes: **"This isn't just permission - this is protection."**

**Observable Success Indicators:**
- Repeat usage of `/approve new` for high-stakes decisions
- Decrease in informal approval DMs for actions that matter
- Short time between approval and action
- Approvals requested for higher-stakes decisions over time (a trust signal)

**Success Signal:** "I don't have to wonder if I'm covered."

---

**Jordan Lee (Approver) - Controlled Decisiveness**

**Desired Outcome:** Approve requests quickly and responsibly, without exposing themselves or the organization to unnecessary risk. Success is being able to say "yes" or "no" confidently - and prove it later.

**Key "Aha" Moment:** The first time Jordan is asked "Did you approve this?" and can point to a single, authoritative record instead of relying on memory or chat history. Jordan realizes: **"This protects me as much as it helps them."**

**Observable Success Indicators:**
- Jordan responds to approval requests faster than informal DMs
- Jordan redirects informal requests into the plugin ("put it through the approval tool")
- Jordan adds brief comments to approvals instead of long chat explanations

**Success Signal:** "I can approve quickly without exposing myself."

---

**Morgan Patel (Auditor/Compliance) - Fast, Defensible Verification**

**Desired Outcome:** Verify that actions were properly authorized without reconstruction, interpretation, or follow-up. Success is clarity, not interaction.

**Key "Aha" Moment:** The first time Morgan reviews an approval record, finds all required information in one place, and closes the review without follow-up questions. Morgan realizes: **"I'm not reconstructing intent - I'm verifying a decision."**

**Observable Success Indicators:**
- Reduced time spent validating approvals
- Fewer requests for supplemental evidence (screenshots, emails)
- Fewer audit exceptions related to missing or unclear authorization

**Success Signal:** "I don't have to guess what happened."

---

**Cross-Persona Success Signal:**

**The strongest signal of success: When something matters, people stop using DMs and start using the approval tool.**

If this behavioral shift happens:
- Alex feels protected
- Jordan feels safe approving
- Morgan trusts the record

No additional features are needed to prove value.

### Business Success

**Distribution Model:**
- **Primary (near-term):** Open-source project built publicly with transparency and community feedback, designed to align with Mattermost's plugin ecosystem and standards
- **Aspirational (longer-term):** Publish to Mattermost Marketplace if proven useful and stable; potential future path as Enterprise-only plugin if adopted officially

This project is not initially optimized for revenue. It is optimized for **usefulness, correctness, and fit within Mattermost's core philosophy.**

**Success Horizons:**

**3-Month Horizon - Early Validation**

Success at this stage:
- A working plugin that solves the core problem end-to-end
- Real usage by at least a handful of teams (internal or community)
- Clear signals that users prefer this over informal DMs for high-stakes approvals
- Positive feedback specifically around confidence, not just speed

**Key question at 3 months:** *"Do people actually use this when it matters?"*

**12-Month Horizon - Established Utility**

Success at this stage:
- Adoption by multiple organizations, especially regulated or on-prem environments
- Evidence of repeat usage across multiple personas (requesters, approvers)
- Recognition as a "missing piece" in the Mattermost ecosystem
- Clear internal or community consensus that this fills a real gap

**Key question at 12 months:** *"Does this feel obvious in hindsight?"*

**Primary Business Metrics:**

Because this is not initially a revenue-driven product, **usage quality matters more than raw numbers.**

Primary metrics:
- **Adoption depth, not just installs** - Are teams using it repeatedly for meaningful approvals?
- **Behavior change** - Are users choosing this over DMs when stakes are high?
- **Cross-persona value** - Does it serve requesters, approvers, and auditors simultaneously?

Secondary metrics:
- Number of organizations installing the plugin
- Repeat usage by the same users
- Requests per organization over time (indicates trust)

Revenue is explicitly out of scope for early success.

**Strategic Contribution:**

This project strengthens Mattermost's broader strategy by:
- Strengthening Mattermost's position with DISC (Defense, Intelligence, Security, Critical Infrastructure) customers
- Demonstrating Mattermost's advantage in secure, chat-native operational workflows
- Filling a real gap between Playbooks and informal chat
- Providing a concrete example of how Mattermost plugins can solve high-trust problems without external SaaS dependencies

Even if it never becomes a commercial product, it adds value to the ecosystem, reinforces Mattermost's differentiation, and showcases best practices for plugin design in regulated environments.

### Technical Success

**MVP Technical Validation:**
- The core workflow works end-to-end
- Records are immutable and retrievable
- The UX feels fast and official (not clunky)
- No major technical blockers discovered

**Product Validation:**
- Internal Mattermost stakeholders see the value proposition
- People who understand DISC customers recognize the gap this fills
- Feedback is "this makes sense" not "why would anyone use this?"

**Decision Point:** If both validations pass, then proceed to polish for marketplace, add v2.0 features, or pursue official adoption paths.

MVP is treated as a **proof of concept with real code** - validating both the technical approach and the product hypothesis before committing to broader distribution.

### Measurable Outcomes

**Success is measured by qualitative behavioral change, not quantitative throughput metrics.**

This product solves a **confidence and authority problem**, not a throughput or performance problem. Additional metrics (latency, volume, analytics) risk over-optimizing the wrong axis and are better deferred until the product's role expands into higher-frequency or automated workflows.

**The core success signal is intentionally simple and sufficient:**
- When stakes are high, teams choose the approval tool over informal DMs
- This behavioral change validates cross-persona value
- All other metrics are secondary to this fundamental shift

## Product Scope

### MVP - Minimum Viable Product

**MVP Goal:** Enable a single requester to receive a clear, authoritative approval from a single approver, with a permanent record that can be verified later - entirely within Mattermost.

**Essential MVP Features:**

**1. Request Creation**
- `/approve new` slash command
- Opens a modal with: Approver (user selector), Description (required free text)
- **Why essential:** This is the moment Alex crosses an authority boundary. Must be fast, structured, and explicit.
- **Design constraint:** Seconds, not minutes.

**2. Approver Notification & Decision**
- Approver receives a DM containing: Request description, Requester identity, Clear Approve/Deny buttons, Confirmation dialog after click
- **Why essential:** Jordan must know this is an official action. Confirmation reinforces authority and intent.
- **Design constraint:** One click to decide. No workflow management.

**3. Immutable Decision Recording**
- Approval decision stored in plugin KV store
- Stored data: Request ID, Requester, Approver, Description, Decision (approved/denied), Timestamp(s)
- **Why essential:** This is the actual product. Without this, Morgan has nothing to verify.
- **Important note:** "Permanent" means append-only, no silent edits, decision cannot be changed without trace.

**4. Requester Outcome Notification**
- Requester receives a DM when approved or denied
- **Why essential:** Alex's confidence moment. This is when action happens. Without this, the loop isn't closed.

**5. Viewing Past Approvals**
- `/approve list` command
- Lists approvals submitted by the user
- Shows: status, approver, timestamp
- **Why essential:** This is Morgan's entry point. Also reinforces trust for Alex and Jordan.
- **Design constraint:** Simple text output is fine for MVP. No UI dashboard needed.

**MVP "Invisible" Requirements:**

**A. Authority Clarity**
- The approver identity must be explicit and unambiguous

**B. Scope Clarity**
- The approval record must clearly represent exactly what was requested and exactly what was approved or denied
- No edits after submission in MVP

**C. Tamper Awareness (Lightweight)**
- No "edit approval" feature
- No "overwrite decision"
- If something changes later, it's a new approval

**Out of Scope for MVP:**

The following features are **intentionally deferred** to prevent scope creep and maintain focus on core value:
- Multi-step approval chains
- Approval escalation / reassignment
- Groups or role-based approvers
- Policy enforcement (spend limits, rules)
- Playbook integration
- Export formats (PDF, CSV)
- UI dashboards
- Admin configuration UI
- Analytics or reporting
- External system integrations

### Growth Features (Post-MVP)

**Guiding Principle for Growth:** Future versions should extend trust, not complexity.

Each new capability should:
- Preserve the speed of MVP
- Increase confidence in higher-stakes scenarios
- Avoid turning the plugin into a policy enforcement system

**If a feature makes approvals slower by default, it doesn't belong.**

**Phase 2 (v2.0): Extending Authority Safely**

**1. Multi-Step Approval Chains & Escalation** *(Highest Priority)*
- Sequential approval chains (A → B → C)
- Ability for approver to add additional approver or forward request upward when authority is insufficient
- Supports chain-of-command scenarios while keeping authority human-driven, not rule-driven

**2. Approval Templates (Context-Aware, Not Policy-Driven)**
- Predefined approval chains for common scenarios
- Template selection at request creation
- Automatic population of approvers
- **Important constraint:** Templates define who to ask, not whether approval is allowed

**Phase 3 (v3.0): Reducing Friction, Not Adding Rules**

**3. User-Defined Roles & Relationships**
- Users can define: their manager, their team lead
- Admins can define: global roles (e.g., Release Manager, Duty Officer)
- Templates reference roles instead of specific users
- Still no enforcement - just mapping

**4. Playbook Integration (Optional, Strategic)**
- Playbook steps that trigger approval requests
- Emergency change control during incidents
- MVP must prove standalone value first
- Playbooks integration should consume approvals, not define them

### Vision (Future)

**Long-Term Vision (1-2 Years)**

If successful, this plugin becomes:
- The default way to request ad-hoc authorization in Mattermost
- A trusted bridge between chat and formal governance
- A natural companion to Playbooks, not a competitor
- A quiet but critical piece of operational infrastructure

**Importantly, it does not become:**
- A ticketing system
- A policy engine
- A compliance platform

Those already exist - and are intentionally out of scope.

**Explicit Non-Goals (Even Long-Term)**

Even in future versions, this plugin will not:
- Enforce approval policies automatically
- Replace ITSM or change management systems
- Decide who should approve something
- Become a generic workflow builder

**Human judgment remains central.**

**Evolution Summary:** This plugin evolves from single approvals into a trusted, chat-native authorization layer - scaling from individual decisions to organizational trust without sacrificing speed.

## User Journeys

**Journey 1: Alex Carter - From Uncertainty to Confidence**

Alex is a site reliability engineer on-call when the monitoring system alerts at 2 AM - the payment processing service is failing. After 20 minutes of investigation, Alex identifies the issue: a configuration change deployed yesterday is causing cascading timeouts. The fix is simple - roll back the change - but there's a problem. Emergency rollbacks require approval from the on-call engineering manager, and the change was made by the VP of Engineering's team.

In the past, Alex would send a frantic DM to their manager Sarah, wait anxiously for a reply, take a screenshot "just in case," and then execute the rollback hoping the informal "yeah, do it" would be good enough if questioned later. The anxiety of acting without solid authorization during incidents has cost Alex sleep for months.

Tonight is different. Alex types `/approve new` in Mattermost. A clean modal appears: select approver (Sarah), describe what needs approval ("Emergency rollback of payment-config-v2 deployment - causing 15% payment failures"). Alex hits submit and continues diagnosing. Within 90 seconds, a notification appears: **"Sarah Chen approved your request: Emergency rollback of payment-config-v2 deployment"** with timestamp and approval ID.

The breakthrough comes in that moment. Alex realizes: **"This isn't just permission - this is protection."** No screenshot needed. No wondering if the approval will hold up. No anxiety about being second-guessed in tomorrow's post-mortem. Alex executes the rollback with confidence, payments recover, and the incident is resolved.

Two weeks later, during the incident review, when the VP asks "Who authorized rolling back my team's deployment?", Alex doesn't scramble for screenshots or try to recall exact DM wording. They simply reference approval #A-1847, and Morgan from the compliance team pulls up the complete record in seconds: Sarah Chen approved it at 2:14 AM with full context. The conversation moves on.

Six months later, Alex has used `/approve new` seventeen times for various exceptions and emergency actions. The pattern is clear: when stakes are high, Alex uses the approval tool. When it doesn't matter, DMs are fine. Alex sleeps better knowing that **when it matters, they're covered.**

---

**Journey 2: Jordan Lee (Sarah Chen) - From Exposed to Protected**

Sarah Chen is an engineering manager who loves enabling her team to move fast but hates the nagging worry that comes after giving quick approvals. She's been managing engineers for five years, and every few months someone asks "Did you approve this?" and she has to dig through chat history, trying to remember exact context from weeks ago.

Last quarter, finance questioned a $2,000 cloud infrastructure expense that Sarah had casually approved via DM ("yeah, spin it up"). The engineer had the screenshot, but Sarah's reply had been ambiguous - did she approve $2,000 or just the concept? The expense was eventually approved, but Sarah spent two hours reconstructing the conversation and writing an explanation. She felt exposed.

One night, Alex's DM notification wakes her at 2:14 AM. But this time it's different - it's clearly labeled as an **Approval Request**, with structured information: who (Alex), what (emergency rollback), why (payment failures), and two clear buttons: **Approve** or **Deny**. Sarah reads the context, sees the urgency, and clicks **Approve**. A confirmation dialog appears: "Confirm you are approving: Emergency rollback of payment-config-v2 deployment." She confirms.

The interface makes it clear: **"This is official. My decision matters."** Sarah goes back to sleep knowing there's a clean record.

The breakthrough comes two weeks later during the incident review. When the VP questions the authorization, Sarah doesn't have to dig through DMs or rely on memory. She simply states: "I approved it - approval #A-1847 has the details." The record shows exactly what she approved, when, and with what context. No ambiguity. No exposure.

Three months later, when an engineer DMs Sarah asking to approve an exception, Sarah replies: **"Put it through the approval tool."** That simple redirect has become her pattern. The approval tool isn't more work for her - it's protection. She can say yes quickly without worrying about being misquoted or losing the record. Sarah realizes: **"This protects me as much as it helps them."**

By year-end, Sarah has approved forty-three requests through the tool. Her approval response time is actually faster than informal DMs because she doesn't second-guess herself. When something matters, both she and her team use the tool. It's become the safest way to say yes and the cleanest way to exercise authority.

---

**Journey 3: Morgan Patel - From Reconstruction to Verification**

Morgan is a compliance auditor who dreads the quarterly operational audit. Every quarter, Morgan must verify that critical changes and exceptions were properly authorized. Every quarter, it's the same painful process: requesting evidence from engineers, receiving a mix of screenshots, email forwards, and conflicting recollections, then spending hours piecing together timelines and inferring intent from vague chat messages like "yeah, go ahead."

Last audit, Morgan spent three days reconstructing a single emergency change authorization. The engineer had a screenshot showing approval, but the manager couldn't remember the specific context. The change was legitimate, but without clear evidence, Morgan had to mark it as a "procedural exception" - which triggers executive review and creates extra work for everyone.

This quarter is different. When Morgan requests evidence for the payment rollback incident from April, Alex sends a single line: "Approval #A-1847." Morgan opens the Mattermost approval list and pulls up the record instantly:

- **Request**: Emergency rollback of payment-config-v2 deployment - causing 15% payment failures
- **Requester**: Alex Carter (SRE)
- **Approver**: Sarah Chen (Engineering Manager)
- **Decision**: Approved
- **Timestamp**: April 15, 2026, 02:14:23 UTC
- **Approval ID**: A-1847

Everything Morgan needs is in one place. No ambiguity. No interpretation needed. No follow-up questions. Morgan realizes: **"I'm not reconstructing intent - I'm verifying a decision."** What used to take hours now takes two minutes.

The breakthrough comes when Morgan presents the Q2 audit results to the executive team. For the first time in three years, there are zero "procedural exceptions" flagged for emergency changes. Every approval is clean, complete, and defensible. The exec team asks what changed. Morgan explains: the engineering team started using a lightweight approval tool that creates authoritative records without slowing them down.

By year-end, Morgan has become an advocate for the approval tool. During incident retrospectives, Morgan actively encourages teams to use it: "Next time, run it through the approval tool - makes my life easier and protects you." The organization becomes more comfortable moving fast because the records keep up. Morgan's quarterly audits go from three weeks to one week, and the quality of evidence has never been better.

---

### Journey Requirements Summary

These three journeys reveal the core requirements for the approval plugin:

**From Alex's Journey (Requester):**
- `/approve new` slash command that opens immediately
- Simple modal with minimal required fields (approver selection, description)
- Fast submission (seconds, not minutes)
- Clear approval notification with timestamp and approval ID
- Ability to reference approval later by ID
- No friction during time-sensitive situations

**From Jordan's Journey (Approver):**
- Structured approval request notification (not just a DM)
- Clear presentation of: who is asking, what they want, why they need it
- One-click approval/denial with confirmation dialog
- Permanent record of exactly what was approved
- No workflow management or configuration required
- Ability to redirect informal requests to the tool

**From Morgan's Journey (Auditor):**
- Ability to retrieve approvals by ID
- Complete, unambiguous approval records
- Immutable records (can't be edited after the fact)
- All context preserved: requester, approver, description, decision, timestamp
- Fast retrieval (minutes, not hours)
- Records that require no interpretation or follow-up questions

**Cross-Journey Requirements:**
- Single source of truth for approval records
- No external dependencies or SaaS
- Works entirely within Mattermost
- Approval records that survive as defensible evidence
- Behavioral shift from informal DMs to formal approvals when stakes are high

## Innovation & Novel Patterns

### Conceptual Innovation: "Bridge Authorization"

Mattermost-Plugin-Approver2 introduces a **conceptual shift** in how organizations handle approvals within chat environments. Rather than technical innovation, the novelty lies in **reframing authorization as a structured, auditable moment inside existing conversation flows**.

**The Innovation:**

Traditional approaches force a binary choice:
- **Informal**: DMs and chat messages (fast, but unaccountable)
- **Formal**: Dedicated approval systems (accountable, but heavy)

This plugin creates a **third category** — "bridge authorization" — that combines the speed and context of chat with the structure and auditability of formal systems. By treating authorization as an embedded, structured moment rather than extracting it to external workflows, it closes a long-ignored gap between informal communication and formal accountability.

**Why This Matters:**

This conceptual positioning is particularly valuable for:
- **Regulated environments** where informal approvals create compliance risk
- **On-prem deployments** where external approval tools add complexity
- **High-trust teams** who need "official enough to act on" without bureaucratic overhead

The innovation isn't in the technology stack — it's in **recognizing and naming** a category of authorization that has always existed informally but never been properly served.

### Validation Approach

For conceptual innovation, validation focuses on **resonance and behavioral adoption** rather than technical proof:

**Primary Validation Signal:**
- **Behavioral change**: When stakes are high, do people choose the tool over DMs?

**Secondary Validation Signals:**
- **Problem recognition**: When described, do users immediately recognize the gap?
- **Framing clarity**: Does "bridge authorization" effectively communicate the concept?
- **Adoption patterns**: Do teams adopt it for specific high-stakes scenarios (incident approvals, resource requests)?

**Validation Methodology:**
- Early beta with DISC customers in real incidents
- Interview 5-10 approvers about decision-making confidence
- Track DM volume for approval-related keywords before/after deployment
- Qualitative feedback: "Does this feel official enough to act on?"

### Risk Mitigation

**Primary Risk**: Conceptual framing doesn't resonate with target market.

**Mitigation Strategy:**
- **Fallback positioning**: If "bridge authorization" terminology doesn't land, reposition simply as "lightweight approval tracking for teams" without the conceptual framing
- **Functional value remains**: Even without conceptual buy-in, the core functionality (immutable records, notifications, audit trail) delivers tangible value
- **Adaptive messaging**: Test framing with early customers and adjust positioning based on what resonates

**Secondary Risk**: Product seen as "too simple" compared to enterprise approval systems.

**Mitigation Strategy:**
- **Feature parity not required**: This is intentionally NOT competing with formal systems — that's the point of "bridge authorization"
- **Clear positioning**: Explicitly communicate this is for "official enough" scenarios, not replacing compliance-mandated approval workflows
- **Value proposition**: Speed and context retention are features, not limitations

## Developer Tool (Mattermost Plugin) Specific Requirements

### Project-Type Overview

Mattermost-Plugin-Approver2 is a **standard Mattermost plugin** following the official plugin architecture without deviations. The technical approach prioritizes **consistency and correctness** over innovation at the framework level. This ensures long-term compatibility, easier review/contribution, and predictable behavior for administrators.

**Technical Foundation:**
- **Backend**: Go (following Mattermost plugin patterns)
- **Frontend**: React (if needed, following Mattermost plugin patterns)
- **Architecture**: Standard Mattermost plugin starter template and APIs

This is not a novel framework or experimental plugin architecture — it's a straightforward implementation using proven patterns.

### Language & Framework Matrix

**Backend Implementation:**
- **Language**: Go
- **Framework**: Mattermost Plugin API (v6.0+)
- **Dependencies**: Standard Mattermost plugin dependencies only
- **Rationale**: Full alignment with official Mattermost plugin patterns ensures compatibility and maintainability

**Frontend Implementation:**
- **Language**: JavaScript/TypeScript (React)
- **Framework**: Mattermost Webapp plugin integration
- **UI Components**: Mattermost UI component library where applicable
- **Rationale**: Consistency with Mattermost look-and-feel, minimal custom styling

**No Deviations:**
- No custom frameworks or experimental patterns
- No language mixing beyond standard plugin structure
- No special build tooling beyond Mattermost plugin requirements

### Installation Methods & Distribution

**Primary Distribution (v1.0):**
- **Method**: Manual upload of plugin tarball (.tar.gz)
- **Target Environment**: On-prem and air-gapped Mattermost instances
- **Installation Flow**:
  1. Download release tarball from GitHub
  2. Upload via Mattermost System Console → Plugin Management
  3. Enable plugin

**Rationale:**
- Aligns with DISC customer deployment expectations
- Supports air-gapped and high-security environments
- No external dependencies or internet connectivity required
- Standard Mattermost admin workflow

**Future Distribution Options:**
- **Mattermost Marketplace**: Considered for future versions once proven stable
- **Note**: Marketplace distribution uses same artifact format, so no architectural changes needed

**Packaging Requirements:**
- Standard Mattermost plugin tarball structure
- Includes plugin manifest (plugin.json)
- Compiled binaries for target platforms (Linux, Darwin, Windows)
- Frontend assets bundled
- No special packaging mechanisms beyond Mattermost conventions

### Plugin API Surface & Integration Points

**MVP API Usage:**

The plugin leverages standard Mattermost plugin APIs without custom REST endpoints or external integrations:

**Slash Commands:**
- `/approve new` - Initiate approval request
- `/approve list` - View approval history
- Future: `/approve help`, `/approve get [ID]`

**Interactive Components:**
- **Modals/Dialogs**: Request creation form (approver selection, description input)
- **Message Actions**: Approve/Deny buttons in DM notifications
- **Confirmation Dialogs**: Explicit confirmation before finalizing approval decision

**Data & Messaging APIs:**
- **Plugin KV Store**: Immutable approval record storage
- **Direct Messages**: Notification delivery to approvers and requesters
- **User Lookup**: Identity resolution and user directory queries
- **Permissions**: Leveraging Mattermost user permissions (no custom RBAC in v1)

**Intentionally Excluded from MVP:**
- Custom REST API endpoints
- External system integrations (webhooks, ITSM connectors)
- Automation-first APIs (no programmatic approval triggering)
- Bulk operations or batch processing

**API Design Philosophy:**
- Human-in-the-loop by design
- Slash-command-driven interaction model
- Interactive messages for decisions (not passive notifications)
- No "headless" or automation-first usage in v1

### Code Examples & Documentation Strategy

**Initial Documentation (v1.0):**

**README.md (Primary Documentation):**
- Clear problem statement and value proposition
- Installation instructions for manual tarball upload
- Usage examples:
  - How to request approval
  - How to approve or deny a request
  - How to view approval history
- Explicit scope clarification: what this plugin does and does not do
- Contribution guidelines (if open-source)

**User-Facing Usage Documentation:**
- Quick start guide for end users (requesters and approvers)
- Screenshots of slash command flows
- Common scenarios and patterns
  - Incident approvals
  - Resource request approvals
  - Exception approvals

**Developer Setup Documentation:**
- Local development environment setup
- Plugin build and deployment instructions
- Testing approach and test execution

**Documentation Principles:**
- Focus on **how to use** and **what problem it solves**
- Explicitly state what it does NOT do (avoid feature creep expectations)
- No prescriptive "best practices" — teams define their own approval patterns
- Clear, actionable guidance over comprehensive reference docs

**Future Documentation Evolution:**
- If adoption grows or plugin becomes officially supported:
  - Migration to Mattermost primary docs (docs.mattermost.com)
  - Optional examples of common approval workflow patterns
  - Integration guides (if future versions support external systems)

### Migration Guide & Upgrade Path

**v1.0 Installation:**
- **No migration required** — greenfield plugin with clean installation
- Focus: Correct initial implementation, not backward compatibility

**Design Considerations for Future Versions:**

While v1.0 has no migration concerns, the data model and APIs should be designed with extensibility in mind:

**Data Structure Extensibility:**
- Approval records should support additional fields without breaking existing records
- Schema versioning considered for KV store data
- Avoid assumptions that would block:
  - Multi-step approval chains
  - Approval templates
  - Additional metadata (tags, categories, custom fields)

**API Stability:**
- Slash command interface should remain stable across versions
- New commands can be added without breaking existing workflows
- Interactive message formats should be forward-compatible

**Backward Compatibility Strategy:**
- Future versions should read v1.0 approval records without migration
- If schema changes required, implement transparent upgrade on read
- Avoid forced migrations that disrupt air-gapped deployments

**Upgrade Principles:**
- Simple plugin replacement via System Console
- No database migrations or manual data conversions
- Graceful handling of legacy data formats
- Clear upgrade notes in release documentation

**Non-Goal:**
- Complex migration tooling or scripts
- Backward compatibility at the expense of v1.0 simplicity
- Supporting every possible future feature path

**Philosophy:** Prioritize correctness and clarity in v1.0. Design for extensibility without over-engineering. Future compatibility is considered but not at the expense of MVP simplicity.

### Technical Architecture Considerations

**Mattermost Plugin Architecture:**
- Standard three-tier plugin model: Server (Go), Webapp (React), Plugin API layer
- Leverages Mattermost's built-in user authentication and permissions
- No external database — plugin KV store provides persistence
- Stateless backend logic (all state in KV store)

**Key Technical Constraints:**
- No external dependencies or SaaS services
- No data egress beyond Mattermost instance boundaries
- On-prem and air-gapped deployment support
- Security inherits from Mattermost platform (no custom auth layer)

**Implementation Priorities:**
1. **Correctness**: Immutable records, clear audit trail, no silent failures
2. **Simplicity**: Minimal configuration, straightforward code, predictable behavior
3. **Compatibility**: Full alignment with Mattermost plugin standards

### Implementation Considerations

**Development Approach:**
- Use official Mattermost Plugin Starter Template as foundation
- Follow Mattermost plugin development best practices
- Regular testing against supported Mattermost versions
- Code review and security review before release

**Testing Strategy:**
- Unit tests for core approval logic
- Integration tests for plugin API interactions
- Manual testing of user workflows (slash commands, interactive messages)
- Validation in air-gapped test environment

**Deployment Validation:**
- Test tarball installation process
- Verify compatibility with target Mattermost versions
- Validate behavior in on-prem and air-gapped scenarios
- Confirm no external dependencies or network calls

**Maintenance Philosophy:**
- Keep codebase simple and maintainable
- Avoid premature abstractions or over-engineering
- Clear code comments for non-obvious design decisions
- Responsive to security issues and critical bugs

## Project Scoping & Development Strategy

### Scoping Approach

The scoping strategy for Mattermost-Plugin-Approver2 has been defined comprehensively in the **Product Scope** section above, with clear MVP boundaries, growth features, and explicit non-goals.

**Key Scoping Decisions Affirmed:**

**MVP Philosophy: Problem-Solving MVP**
- Solve the core "bridge authorization" problem with minimal features
- Five essential capabilities that deliver complete user value
- Intentional simplicity at the approval moment (non-negotiable)
- Human-in-the-loop by design

**Development Phasing:**
- **Phase 1 (MVP)**: Single approver, single request, immutable records
- **Phase 2 (v2.0)**: Multi-step chains, approval templates
- **Phase 3 (v3.0)**: User-defined roles, Playbook integration

**Scope Discipline:**
- Clear boundaries prevent feature creep
- Explicit non-goals protect against mission drift
- Future features extend trust, not complexity

### MVP Validation Strategy

**Success Validation:**
- Behavioral change: When stakes are high, people use the tool over DMs
- Cross-persona value: Requesters feel confident, approvers feel protected, auditors trust records
- Qualitative feedback over quantitative metrics

**Technical Validation:**
- Core workflow works end-to-end
- Records are immutable and retrievable
- UX feels fast and official
- No major technical blockers discovered

**Decision Gate:**
If both validations pass, proceed to polish for marketplace, add v2.0 features, or pursue official adoption paths.

### Resource Requirements

**MVP Team (Estimated):**
- 1 Go developer (Mattermost plugin backend)
- 1 React developer (Mattermost plugin frontend, if UI needed beyond modals)
- Part-time: Testing/QA, documentation

**Development Timeline Considerations:**
- MVP is intentionally lean to enable fast iteration
- Phased approach allows for validated learning between releases
- Resource allocation can scale based on MVP success

**Risk Mitigation:**
- Standard Mattermost plugin architecture reduces technical risk
- Small MVP scope reduces resource risk
- Clear success criteria enable early pivot decisions if needed

## Functional Requirements

### Approval Request Management

- FR1: Requesters can initiate a new approval request by specifying an approver and description
- FR2: Requesters can select any Mattermost user as an approver from the user directory
- FR3: Requesters must provide a description of what requires approval before submitting
- FR4: Requesters can view the status of their submitted approval requests
- FR5: Requesters can view a list of all approval requests they have submitted

### Approval Decision Management

- FR6: Approvers can review approval requests directed to them
- FR7: Approvers can approve an approval request
- FR8: Approvers can deny an approval request
- FR9: Approvers must explicitly confirm their approval or denial decision before it is finalized
- FR10: Approvers can view the requester's identity and full request description before making a decision
- FR11: Approvers can view a list of approval requests awaiting their decision
- FR12: Approvers can view a list of approval requests they have previously approved or denied

### Notification & Communication

- FR13: Approvers receive a direct message notification when an approval request is directed to them
- FR14: Requesters receive a direct message notification when their approval request is approved
- FR15: Requesters receive a direct message notification when their approval request is denied
- FR16: Approval request notifications include all relevant context (requester identity, description, timestamp)
- FR17: Approval outcome notifications include decision details (approver identity, decision, timestamp, approval ID)

### Approval Record Management

- FR18: The system creates an immutable approval record when a decision is finalized
- FR19: Approval records include: request ID, requester identity, approver identity, description, decision (approved/denied), timestamp
- FR20: Approval records can be retrieved by approval ID
- FR21: Approval records cannot be edited after creation
- FR22: Approval records cannot be deleted
- FR23: Users can view approval records they participated in (as requester or approver)
- FR24: Each approval record has a unique identifier

### User Interaction

- FR25: Users can invoke approval functionality via slash commands
- FR26: Users interact with approval request creation through a modal interface
- FR27: Approvers interact with approval decisions through interactive message actions (buttons)
- FR28: The system provides user-friendly command help and guidance
- FR29: The system validates user input before accepting approval requests

### Data Integrity & Auditability

- FR30: The system ensures approval records are stored persistently
- FR31: The system ensures approval records maintain integrity (no silent modifications)
- FR32: The system ensures approval decisions are attributed to authenticated Mattermost users
- FR33: The system ensures timestamps are accurate and immutable
- FR34: The system preserves the complete context of approval requests and decisions
- FR35: Approval records are retrievable for audit and verification purposes

### Identity & Permissions

- FR36: The system leverages Mattermost user authentication for all approval operations
- FR37: The system verifies user identity before recording approval requests
- FR38: The system verifies approver identity before recording approval decisions
- FR39: Users can only view approval records they participated in (as requester or approver)
- FR40: The system uses Mattermost's user directory for approver selection

## Non-Functional Requirements

### Performance

**NFR-P1: Approval Request Submission Responsiveness**
- User must receive confirmation that approval request was submitted within 2 seconds of initiating `/approve new` command
- Rationale: "Seconds, not minutes" is core value proposition; slow submission breaks trust

**NFR-P2: Approval Decision Responsiveness**
- Approver must receive approval request notification within 5 seconds of requester submitting
- Rationale: Time-sensitive scenarios (incidents) require immediate notification

**NFR-P3: Approval Outcome Notification Responsiveness**
- Requester must receive approval outcome notification within 5 seconds of approver making decision
- Rationale: Alex's confidence moment depends on fast feedback loop

**NFR-P4: Approval Record Retrieval Responsiveness**
- Viewing approval history (`/approve list`) must return results within 3 seconds
- Rationale: Morgan needs fast audit access; slow retrieval undermines audit value

### Security

**NFR-S1: Data Residency**
- All approval data must remain within the Mattermost instance boundaries with no external data egress
- Rationale: DISC customer requirement for on-prem and air-gapped deployments

**NFR-S2: Authentication & Authorization**
- All approval operations must leverage Mattermost's native authentication system
- No custom authentication layer or credential storage
- Rationale: Security inherits from Mattermost platform; no additional trust boundary

**NFR-S3: Data Immutability**
- Approval records must be append-only with no edit or delete operations after finalization
- Any attempt to modify finalized approval records must be prevented at the system level
- Rationale: Audit integrity depends on tamper-evident records

**NFR-S4: Identity Verification**
- Requester and approver identities must be verified against authenticated Mattermost sessions before recording decisions
- Rationale: Jordan's protection and Morgan's trust depend on authoritative identity

**NFR-S5: Audit Trail Integrity**
- Timestamps must be system-generated and immutable (not user-provided)
- Approval records must include complete context without relying on external references
- Rationale: Morgan's verification depends on self-contained, trustworthy records

**NFR-S6: No External Dependencies**
- Plugin must operate without external SaaS services, APIs, or network dependencies
- All functionality must work in air-gapped environments
- Rationale: DISC customer deployment requirement

### Reliability

**NFR-R1: Data Persistence**
- Approval records must persist reliably in Mattermost plugin KV store with no data loss
- System must handle plugin restart/upgrade without losing approval history
- Rationale: Morgan's audits depend on records being available months later

**NFR-R2: Decision Finality**
- Once an approval decision is recorded, it must remain accessible and unchanged
- No silent failures in recording decisions
- Rationale: Alex's confidence depends on knowing the approval "will stand up later"

**NFR-R3: Notification Delivery**
- Critical notifications (approval requests, outcomes) must be delivered reliably via Mattermost DM system
- If notification delivery fails, system must log error visibly
- Rationale: Jordan's response and Alex's confidence depend on reliable notification

**NFR-R4: Graceful Degradation**
- If underlying Mattermost services are degraded, plugin must fail visibly (not silently)
- Users must know if approval operations failed rather than assuming success
- Rationale: Incorrect assumptions about approval status create risk

### Usability

**NFR-U1: Approval Request Simplicity**
- Creating an approval request must require no more than 2 required inputs (approver, description)
- No mandatory configuration or workflow setup before first use
- Rationale: "Intentional simplicity at the approval moment" is non-negotiable

**NFR-U2: Approval Decision Simplicity**
- Approving or denying must require no more than 2 interactions (button click, confirmation)
- No navigation away from notification required
- Rationale: Jordan's fast response time depends on minimal friction

**NFR-U3: Command Intuitiveness**
- Slash commands must be self-explanatory (e.g., `/approve new`, `/approve list`)
- Help text must be available via `/approve help`
- Rationale: Users should not need extensive documentation to understand core commands

**NFR-U4: Error Messaging Clarity**
- Error messages must clearly state what went wrong and what action is needed
- No technical jargon or stack traces in user-facing errors
- Rationale: Users need actionable feedback, not debugging information

### Compatibility

**NFR-C1: Mattermost Version Support**
- Plugin must support Mattermost Server versions currently under official Mattermost support
- Clear documentation of minimum supported version
- Rationale: Administrators need predictable compatibility with their deployments

**NFR-C2: Platform Support**
- Plugin must compile and run on Linux, macOS, and Windows servers
- Tarball packaging must include binaries for all supported platforms
- Rationale: Standard Mattermost plugin requirement for broad deployment support

**NFR-C3: Air-Gapped Deployment**
- Plugin must install and operate without internet connectivity
- No runtime dependencies on external package repositories or services
- Rationale: DISC customer requirement for secure, isolated deployments

**NFR-C4: Plugin API Compatibility**
- Plugin must use stable Mattermost Plugin API endpoints only
- Avoid experimental or deprecated APIs that may change
- Rationale: Long-term compatibility and maintainability

### Maintainability

**NFR-M1: Code Simplicity**
- Codebase must prioritize readability and simplicity over clever abstractions
- No premature optimization or over-engineering
- Rationale: Small team, open-source project needs maintainable code

**NFR-M2: Standard Patterns**
- Implementation must follow official Mattermost plugin patterns and conventions
- No custom frameworks or non-standard architectures
- Rationale: Easier review, contribution, and long-term maintenance

**NFR-M3: Test Coverage**
- Core approval logic must have unit test coverage
- Critical workflows (request, approve, deny, retrieve) must have integration tests
- Rationale: Simple code with good tests enables confident changes

**NFR-M4: Documentation Currency**
- README and usage documentation must be updated with each release
- Breaking changes must be clearly documented in release notes
- Rationale: Users need accurate guidance for installation and usage
