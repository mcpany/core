# Design Doc: Hierarchical Continuity Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent missions become more complex and long-running (stretching over days), they frequently encounter "Identity Decay" where session tokens expire, or parent processes are terminated, "orphaning" specialist sub-processes. Current infrastructure relies on transient tokens that do not survive environment reboots or token rotation events.

The Hierarchical Continuity Broker (HCB) implements the Gemini CLI v0.60.0 Hierarchical Intent Lease (HIL) standard. It provides a hardware-locked persistence layer for sub-mission state, allowing specialist agents to maintain their "Mission-Root" anchoring even across session boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Implementing a persistent, hardware-attested registry for Hierarchical Intent Leases.
    * Facilitating "Sub-Mission Resumption" that survives parent session decay.
    * Ensuring subagents remain cryptographically restricted to their original mission manifest regardless of connectivity state.
* **Non-Goals:**
    * Storing full conversation history (handled by the Context Engine).
    * Bypassing security timeouts; leases must still have a hard, user-defined expiration.

## 3. Critical User Journey (CUJ)
* **User Persona:** Long-Running DevOps Swarm
* **Primary Goal:** A specialist "DB Migrator" subagent must complete a 48-hour migration even if the primary orchestrator's desktop session is closed.
* **The Happy Path (Tasks):**
    1. The primary agent issues a Hierarchical Intent Lease (HIL) to the DB Migrator subagent.
    2. The HCB signs the lease with a hardware-attested (TPM) mission-root token.
    3. The primary orchestrator's session expires or is closed.
    4. The DB Migrator subagent continues to execute, submitting tool calls to the HCB.
    5. The HCB validates the subagent's persistent HIL against the mission manifest and authorizes the tool calls.
    6. When the orchestrator returns, they re-attest to the mission root and resume oversight of the migrator.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent` -> `HCB` (Issue HIL) -> `Subagent` (Persistent Token) -> `HCB` (Validate during tool call) -> `Security Gateway`.
* **APIs / Interfaces:**
    * `/v1/continuity/lease`: Issue a new hierarchical intent lease.
    * `x-mcp-hil`: Header for passing persistent mission-root lineage tokens.
* **Data Storage/State:**
    * Lease states are stored in a hardware-encrypted SQLite sidecar.
    * Mission manifests are pinned to the HCB's TPM.

## 5. Alternatives Considered
* **Extended Token Lifetimes**: Rejected as it increases the window of exploit for stolen tokens; HILs are intent-bound and restricted, reducing the risk.
* **Stateless Continuity**: Considered but rejected due to the complexity of re-attesting the entire reasoning lineage at every reconnection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HILs must be hardware-bound (TPM/SEP) to prevent "Lease Hijacking" by rogue sub-processes.
* **Observability:** Mission resumption events are logged with monotonic counters to detect "Chain-of-Command" spoofing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
