# Design Doc: `TeammateTool` Orchestration Adapter
**Status:** Draft
**Created:** 2026-05-17

## 1. Context and Scope
With the official launch of Claude Code "Agent Teams," the `TeammateTool` has emerged as the standard for multi-agent orchestration within the Anthropic ecosystem. However, these swarms are currently restricted to Claude-native agents. MCP Any aims to break this silo by providing a universal `TeammateTool` Orchestration Adapter. This will allow a Claude-led team to discover, spawn, and coordinate with specialized agents from other frameworks (e.g., OpenClaw, AutoGen) via the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a protocol-neutral bridge for the `TeammateTool` 13-operation orchestration layer.
    * Enable cross-framework agent spawning and task delegation (e.g., Claude Lead -> OpenClaw Specialist).
    * Ensure "Mission Root" and context consistency across heterogeneous teammate boundaries.
    * Provide "Snapshot-and-Merge" state reconciliation for parallel teammate branches.
* **Non-Goals:**
    * Modifying the internal reasoning logic of the connected agents.
    * Providing a new agent framework (MCP Any remains an infrastructure layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Full-Stack AI Swarm Architect
* **Primary Goal:** Use a Claude Code "Team Lead" to orchestrate a specialized OpenClaw "Security Auditor" agent.
* **The Happy Path (Tasks):**
    1. The user starts Claude Code in a project monitored by MCP Any.
    2. Claude Code (Team Lead) identifies a need for a security audit and calls `TeammateTool.spawn()`.
    3. MCP Any's adapter intercepts the request and maps it to a verified OpenClaw "Security" agent card.
    4. The OpenClaw agent is spawned in an isolated sandbox, and its communication pipe is bound to the Claude Team session.
    5. The two agents coordinate via the `TeammateTool` operations (message, handoff, wait), with MCP Any handling state synchronization and mission-root anchoring.

## 4. Design & Architecture
* **System Flow:**
    `[Claude Team Lead] <-> [TeammateTool Adapter] <-> [Universal Agent Bus] <-> [OpenClaw Specialist]`
* **APIs / Interfaces:**
    * `TeammateTool.spawn(agent_card_id)`: Maps framework-neutral agent IDs to local/remote capabilities.
    * `TeammateTool.send_message(teammate_id, content)`: Proxies inter-agent communication via the A2A Messaging Hub.
    * `TeammateTool.reconcile_state()`: Implements "Snapshot-and-Merge" logic for the Shared KV Store (Blackboard).
* **Data Storage/State:**
    * Teammate session metadata and mission-root anchors are stored in the Shared KV Store, isolated by `swarm_id`.

## 5. Alternatives Considered
* **Direct A2A Mapping**: Rejected because `TeammateTool` provides a higher-level orchestration semantic (spawning, waiting, merging) that raw A2A lacks.
* **Framework-Specific Proxies**: Rejected as it would lead to an `N x M` integration nightmare. A universal adapter for the `TeammateTool` protocol is more scalable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Every `spawn` and `send_message` call is validated against the cryptographically signed "Mission Root" by the Delegation Attestation Layer (DAL).
* **Observability**: The "Agent Chain Tracer (A2A)" in the UI will visualize `TeammateTool` operations as a hierarchical orchestration tree.

## 7. Evolutionary Changelog
* **2026-05-17:** Initial Document Creation.
* **2026-05-18:** Added support for **Multi-Agent Quorum (MAQ)** coordination. The adapter now facilitates cross-framework approval tokens, allowing an OpenClaw specialist to participate in a Claude-led quorum for high-risk actions. Integrated **Wait-Graph Deadlock Resolution** to identify and break circular task dependencies on the Blackboard during parallel execution.

### Update: 2026-03-24 - Transition to Horizontal Mesh Coordination
**Context:** Today's research confirms Claude Code's shift from hierarchical subagents to "Agent Teams" where teammates share a global task list and communicate directly.
**Architecture Adjustment:**
* Deprecating parent-mediated messaging in favor of direct **LFTC-based** peer communication.
* Integrating **CRDT-native task lists** to allow teammates to asynchronously claim and update work without global locks.
**Security Impact:** Enhances swarm resilience and prevents "Coordinator Bottleneck" attacks, while mandating hardware-attested task-claim integrity.
