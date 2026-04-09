# Design Doc: Mission-Root Conflict Resolver (MRCR)
**Status:** Draft
**Created:** 2026-07-23

## 1. Context and Scope
In heterogeneous meshes, parallel teammates (e.g., Claude Code and OpenClaw specialists) often operate on a shared state (the Blackboard). As swarms become more parallel, multiple agents may attempt to mutate the same context shard or filesystem path simultaneously, leading to "Subagent Collision" or "State Deadlocks."

The Mission-Root Conflict Resolver (MRCR) acts as the authoritative kernel-level arbiter for the Blackboard, resolving conflicting state mutations using mission-aligned priority rules.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect simultaneous mutation attempts on shared state shards.
    * Enforce atomicity for Blackboard write operations.
    * Provide a hierarchical resolution engine based on "Teammate Authority Scores."
    * Log and audit all conflict resolution events for mission-root transparency.
* **Non-Goals:**
    * Managing conflicts between agents from different mission roots (handled by Mesh-Resident Identity Attestation).
    * Resolving semantic reasoning conflicts (handled by AIR Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Resolve a conflict where a "Refactoring Agent" and a "Linter Agent" try to edit the same file in the `.mcpany/scratchpad`.
* **The Happy Path (Tasks):**
    1. Both agents issue a `write` request to the same context key.
    2. The MRCR intercepts both requests and identifies the collision.
    3. The MRCR checks the Mission Manifest and determines the "Refactoring Agent" has a higher authority score for "Filesystem Mutation."
    4. The MRCR commits the Refactoring Agent's change and sends a "State Conflict / Retry" signal to the Linter Agent.
    5. The Linter Agent re-ingests the updated state and applies its changes on top of the new baseline.

## 4. Design & Architecture
* **System Flow:**
    * [Subagent A/B] -> (Concurrent Write) -> [MRCR Arbiter]
    * [MRCR Arbiter] -> (Authority Check) -> [Mission Manifest]
    * [MRCR Arbiter] -> (Atomic Commit / Reject) -> [Blackboard]
* **APIs / Interfaces:**
    * `PUT /v1/state/shard/{id}`: Now mediated by MRCR middleware.
    * `x-mcpany-authority-score`: Metadata embedded in agent tokens.
* **Data Storage/State:**
    * Conflict resolution policies are stored in the mission manifest. Mutation logs are appended to the **Immutable State Trail**.

## 5. Alternatives Considered
* **First-Writer Wins**: Rejected as it leads to "Silent Shadowing" where high-priority changes are overwritten by fast, low-trust agents.
* **Global Mailbox Locks**: Rejected due to prohibitive latency in horizontal Agent Teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Authority scores must be hardware-attested to prevent "Privilege Escalation" via bid shadowing.
* **Observability:** Monitor "Contention Density" to identify bottlenecks in mission-root task allocation.

## 7. Evolutionary Changelog
* **2026-07-23:** Initial Document Creation.

### Update: 2026-07-25 - Interrupt-Aware Conflict Resolution (IACR)
**Context:** Market feedback indicates that standard consensus heartbeats cause "Coordination Deadlocks" during rapid mission pivots.
**Architecture Adjustment:**
* Introducing **Interrupt-Aware Conflict Resolution (IACR)** in Section 4.
* Implementing a "Fast-Path Interrupt" signal that allows mission-root instructions to bypass standard sharding locks and authority-score checks.
* Forcing immediate state reconciliation upon receipt of a high-priority mission-root pivot.
**Security Impact:** Ensures mission-root sovereignty is maintained even during high-contention parallel execution, preventing subagents from "locking out" parent corrections.
