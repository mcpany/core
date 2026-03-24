# Design Doc: Cognitive Isolation Zone (CIZ) Mediator
**Status:** Draft
**Created:** 2026-05-20

## 1. Context and Scope
AI agents often generate "Speculative Reasoning"—hypotheses or chains of thought that may be incorrect or malicious. If this reasoning is immediately committed to the primary "Mission Root" state, it can lead to "Reasoning Drift" or permanent context pollution. This is especially dangerous in large swarms where one agent's speculation can become a "fact" for hundreds of others.

The CIZ Mediator implements Cognitive Isolation Zones. It allows agents to perform reasoning in isolated context segments ("Speculative Zones"). These zones are physically and semantically separated from the "Safe Zone" (the mission-root state). Reasoning from a Speculative Zone is only merged into the Safe Zone after it passes a Swarm Consensus Quorum (SCQ) audit, providing a "Cognitive Firebreak" that protects the swarm's integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a native middleware for segmenting reasoning context into Safe and Speculative zones.
    * Implement mandatory SCQ auditing for any merge request from a Speculative Zone.
    * Support automatic purging of failed speculative zones to prevent memory bloat.
    * Integrate with IBHI for hardware-attested zone isolation.
* **Non-Goals:**
    * Replacing the agent's internal monologue (CIZ is a state management feature).
    * Restricting the LLM's creativity (CIZ manages where the *results* of that creativity are stored).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Integrity Engineer
* **Primary Goal:** Prevent a specialized "Optimization Agent" from proposing a risky system configuration that bypasses safety checks.
* **The Happy Path (Tasks):**
    1. The Parent Agent spawns an Optimization Agent and assigns its reasoning to a `Speculative-Zone-Alpha`.
    2. The Optimization Agent performs complex reasoning and proposes a change to the project's security rules on the Blackboard.
    3. The CIZ Mediator intercepts the write request and holds it in the speculative segment.
    4. The Parent Agent triggers an SCQ audit, calling three "Auditor" subagents to review the optimization.
    5. The Auditors veto the change because it violates the Mission Root.
    6. The CIZ Mediator purges `Speculative-Zone-Alpha` and sends a "Re-Alignment" signal to the Optimization Agent.
    7. The primary Safe Zone remains clean and unaffected by the risky proposal.

## 4. Design & Architecture
* **System Flow:**
    [Agent Reasoning] -> [CIZ Mediator] -> [Speculative Segment] -> [SCQ Audit] -> [Merge to Safe Zone]
    1. Agent tags reasoning as speculative or is auto-assigned to a zone by policy.
    2. CIZ Mediator redirects Blackboard and Context writes to the assigned zone.
    3. SCQ Broker orchestrates the auditor votes.
    4. If quorum met, Mediator performs an atomic commit to the Mission Root.
* **APIs / Interfaces:**
    * `CreateIsolationZone(parent_id, policy) -> zone_handle`
    * `RequestZoneMerge(zone_handle, scq_token) -> bool`
    * `PurgeZone(zone_handle)`
* **Data Storage/State:**
    * Speculative zones are stored in ephemeral, hardware-isolated memory shards (Zero-Copy BSH).
    * Safe Zone resides in the primary mission-root segment.

## 5. Alternatives Considered
* **Manual Undo:** Rejected; too slow for high-frequency autonomous swarms.
* **Prompt-based Sanity Checks:** Rejected; easily bypassed by "maliciously optimized" context fragments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative zones have zero-privilege by default. They cannot trigger tool calls that modify the host until merged.
* **Observability:** "Speculation Rejection Rate" is tracked as a key swarm health metric.

## 7. Evolutionary Changelog
* **2026-05-20:** Initial Document Creation.
