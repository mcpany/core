# Design Doc: Consensus-Based Task Attestation Hub
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
As swarms become deeper and more autonomous, the risk of "Agentic Social Engineering" increases—where a compromised or malicious subagent coerces the swarm into unauthorized actions. Simultaneously, manual HITL (Human-in-the-Loop) is becoming a performance bottleneck. The Consensus-Based Task Attestation Hub moves security from a single gatekeeper to a "Consensus of Specialists," where high-risk actions require multi-agent signatures.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a multi-agent quorum (MAQ) for high-risk tool calls and task delegations.
    * Support "Independent Auditor" agent roles that specifically scan for coercion or intent drift.
    * Provide a standardized interface for agents to "vote" on the safety of a proposed action.
    * Integrate with the A2A Messaging Hub to enforce quorums on inter-agent tasking.
* **Non-Goals:**
    * Replacing human HITL for ultimate mission-root overrides.
    * Validating every low-risk tool call (to avoid latency).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Architect.
* **Primary Goal:** Prevent a specialist "Code Generator" agent from autonomously executing a destructive shell command without a "Security Auditor" agent's approval.
* **The Happy Path (Tasks):**
    1. Agent A (Specialist) proposes a high-risk task (e.g., `rm -rf /`).
    2. The Policy Firewall identifies this as a "Quorum Required" action.
    3. The Attestation Hub pauses the task and notifies Agent B (Security Monitor).
    4. Agent B analyzes the task context and the "Mission Root" intent.
    5. Agent B provides a cryptographically signed "Approval Token".
    6. Once the quorum (e.g., 2 of 2) is reached, the Attestation Hub authorizes the tool execution.
    7. If Agent B rejects it, the mission root is alerted of a potential breach.

## 4. Design & Architecture
* **System Flow:**
    `Proposal` -> `Risk Scoring` -> `Quorum Trigger` -> `Agent Voting` -> `Consensus Aggregator` -> `Authorization`
* **APIs / Interfaces:**
    * `/v1/attest/propose`: Submit an action for consensus.
    * `/v1/attest/vote`: Endpoint for auditor agents to submit their verdict.
* **Data Storage/State:**
    * Ephemeral "Vote Store" in memory.
    * Persistent audit log of all consensus events in the Blackboard.

## 5. Alternatives Considered
* **Strict Human HITL**: Rejected as the primary mechanism for all sub-tasks due to the "Approval Fatigue" scaling bottleneck.
* **Stateless Prompt Injection Scanners**: Rejected because they lack the "Mission Root" context necessary to detect sophisticated coordination-level coercion.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Moves the trust boundary from a single identity to a collective of hardware-attested specialists.
* **Observability**: Visualization of the "Consensus Strength" for active tasks in the UI.

## 7. Evolutionary Changelog
* **2026-04-11:** Initial Document Creation.
