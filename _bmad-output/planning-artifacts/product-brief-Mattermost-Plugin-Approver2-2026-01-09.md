---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments: []
date: 2026-01-09
author: Wayne
---

# Product Brief: Mattermost-Plugin-Approver2

## Executive Summary

The Mattermost Approval Workflow Plugin solves a critical but overlooked problem: **teams need to make authorized decisions quickly, but lack a way to formalize those decisions that's both fast enough for urgent situations and authoritative enough to rely on later.**

Organizations currently accept hidden risk by relying on informal chat approvals. While direct messages and channel posts enable quick decisions, they are non-authoritative and difficult to defend during audits, compliance reviews, or after-action analysis. Conversely, formal workflow systems (ServiceNow, ITSM, JIRA) are too heavyweight for ad-hoc, time-sensitive scenarios where authority matters more than process.

**This plugin creates "bridge authorization"** - official enough to act on immediately, but not intended to replace downstream formal systems. It fills the gap between informal decision-making and formal accountability, giving teams a safe way to move fast.

**Primary value:** Psychological safety for requesters (confidence to act) and procedural safety for organizations (defensible records).

**Target customers:** Defense, Intelligence, Security, and Critical Infrastructure (DISC) organizations where compliance requirements, chain-of-command authority, and on-premise/air-gapped deployment are critical.

---

## Core Vision

### Problem Statement

**There is no lightweight, chat-native approval mechanism that supports urgent, ad-hoc decisions while still producing an authoritative, auditable record.**

Teams face urgent situations daily:
- Sales associates need spending authority that exceeds their limit
- Military personnel need chain-of-command approvals in time-sensitive operational scenarios
- Incident responders need emergency change control approval during active outages

Current approaches force an impossible trade-off:
- **Informal chat approvals** (DMs, channel posts) are fast but non-authoritative - they work socially but fail procedurally
- **Formal workflow systems** (ServiceNow, ITSM, JIRA) are authoritative but too slow and heavyweight for ad-hoc decisions

The core question teams ask is: **"Can I safely act on this?"** Informal approvals don't provide confidence; formal systems don't provide speed.

### Problem Impact

**The cost of failure is asymmetric:** approvals are cheap to formalize, but expensive to question later.

When approvals are informal:
- **Requesters experience uncertainty** - they don't know if they're protected when challenged later
- **Organizations accept hidden risk** - approvals are inconsistent, authority boundaries are unclear, and records are mutable or incomplete
- **Finance, compliance, and audit teams must reconstruct intent after the fact** - often with incomplete or ambiguous evidence

When approvals are delayed by heavyweight systems:
- Sales opportunities are lost
- Operational decisions are postponed with potential safety or mission consequences
- Incident response is slowed during critical outages

### Why Existing Solutions Fall Short

Existing solutions optimize for the wrong axis.

**Enterprise workflow systems** (ServiceNow, BMC Remedy, JIRA) assume:
- Approvals are planned and predefined
- Workflows are known in advance
- The goal is enforcement and compliance

These tools are excellent at enforcing rules, but bad at enabling judgment in real time. They assume you already know the process - but this plugin supports decisions when the process is still forming.

**Mattermost Playbooks** are excellent for repeatable operational workflows (incidents, onboarding, releases), but assume a known sequence of steps and roles. Approvals in ad-hoc scenarios are not "steps" in a playbook - they are exceptions that arise organically during work.

**Slack workflow builders and approval bots** often require pre-built workflows or admin configuration, and many are SaaS-based, creating data residency and compliance concerns unsuitable for air-gapped or highly regulated environments.

**Email approval chains** are slow, asynchronous, and disconnected from where work is happening.

**Informal chat approvals** work socially but fail procedurally - they lack structure, consistency, and defensibility.

There is a persistent gap where teams fall back to informal DMs - not because they're ideal, but because they're the only option that matches the urgency of the moment.

### Proposed Solution

The Mattermost Approval Workflow Plugin creates **formalized informal approvals** - authoritative decisions that happen at the speed of chat.

**Core workflow:**
1. **Request creation** - User runs `/approve new`, a modal opens with minimal fields (approver, description, optional justification/amount), submission takes seconds
2. **Approval** - Approver receives DM notification with context and action buttons, clicks Approve or Deny with optional comment
3. **Outcome** - Requester is notified immediately, approval is permanently recorded with timestamp and authority
4. **Later use** - Users can run `/approve list` to view all requests, statuses, and timestamps; records can be referenced or exported as proof of authorization

**Key capabilities:**
- **Single-approver workflows** (v1 MVP) - one requester, one approver, one explicit decision, immutable record
- **Multi-step approval chains** (future) - support for chain-of-command scenarios requiring sequential approvals
- **Playbook integration** (future) - callable as part of incident management workflows for emergency change control
- **Fully self-contained** - all data remains within Mattermost, works fully in on-prem and air-gapped environments

**Design philosophy:**
- **Chat-native** - approvals happen where work and context already live
- **Ad-hoc** - no predefined workflows or admin setup required
- **Lightweight** - creating requests takes seconds, not minutes
- **Authoritative** - produces defensible, auditable records
- **Bridge authorization** - official enough to act on, not replacing downstream formal systems (ServiceNow, change management, expense systems)

This plugin turns informal chat approvals into authoritative decisions - without slowing teams down or forcing them into heavyweight workflows.

### Key Differentiators

**1. True Mattermost-native integration**

Being native to Mattermost provides advantages external tools cannot replicate:
- Deep awareness of users, channels, roles, and context
- First-class notifications (DMs, mentions, threads) without integration friction
- Trust inherited from the same system already used for mission-critical communication
- Native slash commands, interactive messages, and permission model
- Native deployment model (on-prem/air-gapped)

For Defense, Intelligence, Security, and Critical Infrastructure customers, this means:
- No external SaaS dependencies
- No data egress
- No new trust boundary

**2. Intentional simplicity**

Most approval tools are built by workflow vendors who think in terms of policies, states, and enforcement, or automation platforms that push users toward complex configuration. Competitors are structurally incentivized to over-engineer this problem.

This plugin is intentionally minimal:
- No predefined workflows required
- No admin-heavy setup
- No attempt to replace governance systems
- Focus on the single moment when a human needs explicit authorization right now

**3. "Bridge authorization" positioning**

This is not a ServiceNow replacement, Playbooks competitor, or compliance platform. It occupies a deliberately narrow but critical space: the moment when authority is needed immediately, and formal documentation can follow later.

The key insight: **formal and informal approvals are not opposites—they're adjacent.** Organizations don't fail because they lack formal workflows. They fail in the gap between decisions made in real time and records created after the fact.

This plugin formalizes just enough of the informal moment to make it safe.

**4. Right timing with converging forces**

Three factors make this the right time:
- **Chat has become the operational backbone** - distributed and hybrid work has made chat the primary coordination layer
- **Compliance pressure is increasing even for fast teams** - more teams must justify decisions without slowing operations
- **Mattermost Playbooks have matured** - with repeatable workflows well-covered, the remaining gap is ad-hoc decision-making

For DISC customers in particular, informal approvals are no longer "good enough," but heavyweight systems are still impractical in urgent moments.

---

## Target Users

### Primary Users

**Alex Carter - The Requester**

Alex is a frontline operator who regularly needs explicit authorization to proceed with work that exceeds their normal authority. Depending on the organization, Alex may be a sales or field employee operating remotely, a military service member working within a chain of command, or an incident responder managing systems during an outage.

**Role & Environment:**
Alex works in time-sensitive situations, often away from formal systems (ticketing tools, approval portals). Mattermost is their primary coordination and communication tool. Alex is trusted to act - but only within limits.

**Current Approval Experience:**
Today, Alex sends a DM or posts in a channel, explains the situation informally, waits for a response, and takes a screenshot "just in case." If the situation is urgent, Alex may act before receiving clear approval - hoping to justify it later. If it's less urgent, Alex may hesitate - slowing down work to avoid personal or organizational risk.

**Pain Points:**
Alex isn't frustrated by getting approvals - they're frustrated by uncertainty. Key pain points include:
- Not knowing if an informal approval is "good enough"
- Worrying about being challenged later
- Having to explain context repeatedly to different stakeholders
- Losing time during urgent situations because formal systems are too slow

Alex feels caught between acting quickly to do the right thing and protecting themselves and their organization.

**Success Vision:**
Success is confidence to act. With the right tool, Alex wants to request approval in seconds, get a clear authoritative decision, move forward immediately, and know there's a permanent record if questions come up later. Alex doesn't want to think about compliance - they just want to know: **"I'm covered. I did this the right way."**

**User Journey:**
- **Trigger:** Alex hits an authority boundary (spend limit, operational threshold, emergency change requirement)
- **First Experience:** Alex sees `/approve new`, opens the modal, fills in minimal fields, and submits
- **Aha Moment:** Alex receives the approval notification - not just "Approved" but approved by a named authority, timestamped, clearly scoped. Alex realizes: **"This isn't just permission - this is protection."**
- **Acting with Confidence:** Alex proceeds immediately without slowing down to cover themselves
- **Becoming Routine:** `/approve new` becomes muscle memory. The mental shift: "If it matters, I use the approval tool. If it doesn't, chat is fine."

---

**Jordan Lee - The Approver**

Jordan is a decision-maker with explicit authority over certain actions, resources, or changes. Depending on the organization, Jordan may be a sales manager approving discretionary spend, a commanding officer approving operational actions, or a senior engineer or manager approving emergency changes.

**Role & Responsibilities:**
Jordan is accountable for the outcomes of approvals. Their decisions may be reviewed later. They must balance speed with judgment. Jordan's authority is real - and so is the risk attached to using it.

**Current Approval Experience:**
Today, Jordan receives approval requests as quick DMs, channel pings, or informal "hey, can I do X?" messages. These requests often lack structure, may omit key context, and create pressure to respond quickly. Jordan may reply informally ("Yeah, that's fine") knowing that the wording is ambiguous, the record is weak, and they might be asked about it later.

**Pain Points:**
Jordan's biggest fear is not being interrupted - it's being on the hook for a decision they can't later defend. Pain points include:
- Approving something without full context
- Worrying that a casual reply will be interpreted too broadly
- Being unable to quickly recall what they approved last week or last month
- Being asked later "Why did you approve this?" without a clean record

Jordan often feels tension between enabling their team to move fast and protecting themselves and the organization.

**What Jordan Needs:**
At minimum: who is asking, what they are asking for, why it's needed now, what the impact or risk is, and what authority Jordan is exercising. Critically, Jordan needs confidence they actually have the authority to approve and a clear record of what they approved (and what they did not).

**Success Vision:**
Success is controlled decisiveness. With the right tool, Jordan wants to see a clear structured request, understand what authority they're exercising, approve or deny confidently with one action, and know there's a permanent accurate record of the decision. Jordan doesn't want to manage workflows - they want to own decisions cleanly.

**User Journey:**
- **First Request:** Jordan receives a notification that's clearly different - structured, scoped, explicitly labeled as an approval request. They immediately know: **"This is official. My decision matters."**
- **Making the Decision:** Jordan sees who, what, and why - then clicks Approve or Deny with optional comment. No workflow to manage, no system to configure, just a decision.
- **Aha Moment:** Later, when asked "Did you approve this?", there's a clean record. Jordan realizes: **"This protects me as much as it helps them."**
- **Behavioral Shift:** Jordan becomes less comfortable approving via DM. When asked informally, they reply: **"Put it through the approval tool."** The plugin becomes the safest way to say yes, the cleanest way to say no, and the clearest way to define authority.

---

### Secondary Users

**Morgan Patel - The Auditor/Compliance Officer**

Morgan is responsible for verifying that decisions were properly authorized after the fact. Depending on the organization, Morgan may be a finance or expense auditor, a compliance officer, an internal reviewer for operational or security incidents, or part of an after-action or post-incident review team.

**Role & Interaction:**
Morgan is not involved in making decisions - they are involved in validating them. Morgan's involvement is episodic, not continuous, typically during routine audits, incident reviews, exception investigations, or regulatory inquiries.

**Current Experience:**
Today, when asked "Can you confirm this was approved?", Morgan must search through chat logs, request screenshots from employees, interpret informal language ("yeah, that's fine"), and piece together timelines across channels and DMs. This process is time-consuming, error-prone, and dependent on individual memory and cooperation. Even when approvals did happen, Morgan often cannot prove it cleanly.

**Pain Points:**
Morgan's pain is not inefficiency - it's ambiguity. Key concerns include:
- Unclear scope of what was approved
- Inability to confirm the approver's authority
- Missing timestamps or incomplete context
- Records that can be edited or deleted

When Morgan cannot verify an approval, reimbursements may be denied, decisions may be questioned retroactively, organizations may fail audits or reviews, and trust in internal processes erodes.

**What Morgan Needs:**
A "good enough" approval record is explicit (what was approved), authoritative (who approved it), timestamped, immutable or tamper-evident, and easily retrievable. Morgan does not need workflow diagrams, decision rationale debates, or real-time notifications. They need confidence and clarity.

**Success Vision:**
With the plugin in place, Morgan can retrieve a single authoritative approval record where the approver, request, decision, and timestamp are unambiguous. Verification takes minutes instead of hours or days. The best outcome for Morgan is: **"This is clear. I don't need to ask follow-up questions."**

**User Journey:**
- **First Encounter:** Instead of screenshots, vague chat excerpts, or conflicting recollections, Morgan is handed a single approval record
- **Verification:** The record shows what was approved, by whom, when, and with what context - no ambiguity, no interpretation, no follow-up questions
- **Aha Moment:** Morgan realizes: **"I'm not reconstructing intent - I'm verifying a decision."** What used to take hours now takes minutes.
- **Long-term Impact:** Morgan begins to trust approvals from Mattermost, recommend the plugin during reviews, and push back less on fast decisions because they're defensible. The organization becomes more comfortable moving quickly because the record keeps up.

---

### User Journey Convergence

All three personas converge on one artifact: **a clear, authoritative approval record**

- **Alex** sees it as permission to act
- **Jordan** sees it as proof of responsible authority
- **Morgan** sees it as evidence

This convergence creates alignment instead of friction. The plugin doesn't optimize for one role at the expense of others - it creates shared confidence in fast decisions.

**Core "This Is Exactly What I Needed" Moments:**
- **Alex:** "I don't have to wonder if I'm covered."
- **Jordan:** "I can approve quickly without exposing myself."
- **Morgan:** "I don't have to guess what happened."

When all three are true at once, the product has done its job.

---

## Success Metrics

Success for the Mattermost Approval Workflow Plugin is measured by behavioral change: **when something matters, people stop using DMs and start using the approval tool.**

This behavioral shift validates that all three personas are getting value simultaneously:
- Alex (requester) feels protected
- Jordan (approver) feels safe approving
- Morgan (auditor) trusts the record

### User Success Metrics

**Alex Carter (Requester) - Confidence to Act**

**Desired Outcome:** Alex wants to move forward quickly without uncertainty or personal risk. Success is knowing that the approval is explicit, authoritative, and will stand up later.

**How Alex Knows It's Working:**
- Receives approval quickly and clearly
- No longer wonders if a DM or screenshot will be "good enough"
- Acts immediately after approval without hesitation

**Key "Aha" Moment:** The first time Alex receives a clear approval notification, proceeds with an action, and doesn't feel the need to document or defend themselves afterward.

**Observable Success Indicators:**
- Repeat usage of `/approve new`
- Decrease in informal approval DMs for actions that matter
- Short time between approval and action
- Approvals requested for higher-stakes decisions over time (a trust signal)

**Proxy Metrics (if instrumented later):**
- Median time from request → decision
- Percentage of approvals acted on within minutes
- Growth in approval usage among repeat users

---

**Jordan Lee (Approver) - Controlled Decisiveness**

**Desired Outcome:** Jordan wants to approve requests quickly and responsibly, without exposing themselves or the organization to unnecessary risk. Success is being able to say "yes" or "no" confidently - and prove it later.

**How Jordan Knows It's Working:**
- Approval requests arrive with enough context to make a decision
- The act of approving feels official, not casual
- There's a clean record without extra effort

**Key "Aha" Moment:** The first time Jordan is asked "Did you approve this?" and can point to a single, authoritative record instead of relying on memory or chat history.

**Observable Success Indicators:**
- Jordan responds to approval requests faster than informal DMs
- Jordan redirects informal requests into the plugin ("put it through the approval tool")
- Jordan adds brief comments to approvals instead of long chat explanations

**Proxy Metrics (if instrumented later):**
- Approval response time
- Ratio of approvals vs denials (indicates clarity, not rubber-stamping)
- Repeat approval usage by the same approvers

---

**Morgan Patel (Auditor/Compliance) - Fast, Defensible Verification**

**Desired Outcome:** Morgan wants to verify that actions were properly authorized without reconstruction, interpretation, or follow-up. Success is clarity, not interaction.

**How Morgan Knows It's Working:**
- Approval records are easy to retrieve
- Records are explicit and unambiguous
- Verification takes minutes, not hours
- Morgan stops asking "Can you explain what happened here?"

**Key "Aha" Moment:** The first time Morgan reviews an approval record, finds all required information in one place, and closes the review without follow-up questions.

**Observable Success Indicators:**
- Reduced time spent validating approvals
- Fewer requests for supplemental evidence (screenshots, emails)
- Fewer audit exceptions related to missing or unclear authorization

**Proxy Metrics (future):**
- Time-to-verification during audits
- Reduction in approval-related findings
- Increased acceptance of Mattermost-sourced approvals

---

### Cross-Persona Success Signal

**The strongest signal of success: When something matters, people stop using DMs and start using the approval tool.**

If this behavioral shift happens:
- Alex feels protected
- Jordan feels safe approving
- Morgan trusts the record

No additional features are needed to prove value.

### Business Objectives

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

**Optimization Priorities:**

This project is optimizing for:
- Correctness over completeness
- Trust over feature count
- Real-world usefulness over theoretical scope
- Alignment with Mattermost's values over short-term monetization

If it proves valuable, paths to marketplace inclusion, enterprise differentiation, or official productization can be evaluated later - with evidence.

---

## MVP Scope

### Core Features

**MVP Goal:** Enable a single requester to receive a clear, authoritative approval from a single approver, with a permanent record that can be verified later - entirely within Mattermost.

**Essential MVP Features:**

**1. Request Creation**

- `/approve new` slash command
- Opens a modal with:
  - Approver (user selector)
  - Description (required free text)

**Why it's essential:** This is the moment Alex crosses an authority boundary. It must be fast, structured, and explicit.

**Design constraint:** No extra fields unless they directly increase confidence. Seconds, not minutes.

---

**2. Approver Notification & Decision**

- Approver receives a DM containing:
  - Request description
  - Requester identity
  - Clear Approve / Deny buttons
  - Confirmation dialog after click (to prevent fat-finger approvals)

**Why it's essential:** Jordan must know this is an official action. The confirmation reinforces authority and intent.

**Design constraint:** One click to decide. No workflow management.

---

**3. Immutable Decision Recording**

- Approval decision stored in plugin KV store
- Stored data includes:
  - Request ID
  - Requester
  - Approver
  - Description
  - Decision (approved/denied)
  - Timestamp(s)

**Why it's essential:** This is the actual product. Without this, Morgan has nothing to verify.

**Important note:** "Permanent" does not mean undeletable forever. It means:
- Append-only
- No silent edits
- Decision cannot be changed without trace

This can be simple in MVP, but it must be intentional.

---

**4. Requester Outcome Notification**

- Requester receives a DM when approved or denied

**Why it's essential:** Alex's confidence moment. This is when action happens. Without this, the loop isn't closed.

---

**5. Viewing Past Approvals**

- `/approve list` command
- Lists approvals submitted by the user
- Shows: status, approver, timestamp

**Why it's essential:** This is Morgan's entry point. Also reinforces trust for Alex and Jordan.

**Design constraint:** Simple text output is fine for MVP. No UI dashboard needed.

---

**MVP "Invisible" Requirements**

These aren't flashy, but they are non-negotiable:

**A. Authority Clarity**
- The approver identity must be explicit and unambiguous

**B. Scope Clarity**
- The approval record must clearly represent:
  - Exactly what was requested
  - Exactly what was approved or denied
- No edits after submission in MVP

**C. Tamper Awareness (Lightweight)**
- No "edit approval" feature
- No "overwrite decision"
- If something changes later, it's a new approval

This keeps Morgan happy without building crypto.

---

### Out of Scope for MVP

The following features are **intentionally deferred** to prevent scope creep and maintain focus on core value:

**Deferred Features:**
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

These are explicitly **non-goals for MVP** but may be considered for future versions based on validation and feedback.

---

### MVP Success Criteria

**Technical Validation:**
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

---

### Future Vision

**Guiding Principle for Growth:** Future versions should extend trust, not complexity.

Each new capability should:
- Preserve the speed of MVP
- Increase confidence in higher-stakes scenarios
- Avoid turning the plugin into a policy enforcement system

**If a feature makes approvals slower by default, it doesn't belong.**

---

**Phase 2 (v2.0): Extending Authority Safely**

**1. Multi-Step Approval Chains & Escalation** *(Highest Priority)*

Once single-approver workflows are trusted, the most natural extension is shared or hierarchical authority.

**Capabilities:**
- Sequential approval chains (A → B → C)
- Ability for an approver to:
  - Add an additional approver
  - Forward the request upward when authority is insufficient

**Example:** A middle manager receives a request, recognizes it exceeds their authority, and forwards it to a director without restarting the process.

**Why this comes first:**
- Directly supports chain-of-command scenarios
- Preserves accountability
- Builds naturally on the MVP data model

This keeps authority human-driven, not rule-driven.

---

**2. Approval Templates (Context-Aware, Not Policy-Driven)**

Templates introduce consistency without rigidity.

**Capabilities:**
- Predefined approval chains for common scenarios
- Template selection at request creation
- Automatic population of approvers

**Examples:**
- "Emergency Release"
- "Over-Limit Spend"
- "Operational Exception"

**Why templates come after chains:**
- Templates are just shortcuts over multi-step approvals
- Without chains, templates are shallow
- With chains, templates become powerful and intuitive

**Important constraint:** Templates define who to ask, not whether approval is allowed.

---

**Phase 3 (v3.0): Reducing Friction, Not Adding Rules**

**3. User-Defined Roles & Relationships**

To make templates usable at scale, the system must understand relationships, not org charts.

**Capabilities:**
- Users can define:
  - Their manager
  - Their team lead
- Admins can define:
  - Global roles (e.g., Release Manager, Duty Officer)
- Templates reference roles instead of specific users

**Why this comes later:**
- Requires careful UX and admin design
- Introduces governance concerns
- Only valuable once templates are proven

Still no enforcement - just mapping.

---

**4. Playbook Integration (Optional, Strategic)**

Approval workflows can be invoked from Playbooks via slash commands.

**Capabilities:**
- Playbook steps that trigger approval requests
- Emergency change control during incidents
- Consistent approval handling inside operational workflows

**Why this is later:**
- MVP must prove standalone value first
- Playbooks integration should consume approvals, not define them

This keeps Playbooks as orchestration tools, not approval engines.

---

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

---

**Explicit Non-Goals (Even Long-Term)**

Even in future versions, this plugin will not:
- Enforce approval policies automatically
- Replace ITSM or change management systems
- Decide who should approve something
- Become a generic workflow builder

**Human judgment remains central.**

---

**Evolution Summary:** This plugin evolves from single approvals into a trusted, chat-native authorization layer - scaling from individual decisions to organizational trust without sacrificing speed.

---
