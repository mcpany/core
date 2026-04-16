# Design Doc: Speculative State Arbiter (SSA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward high-density parallel execution, "Cognitive Stall" has emerged as a primary performance bottleneck. Agents frequently enter long wait cycles (5s+) while attempting to resolve state conflicts on shared resources like the Blackboard or project scratchpads. Current synchronous locking mechanisms are insufficient for the sub-millisecond coordination required by modern meshes.

The Speculative State Arbiter (SSA) introduces "Speculation-Aware" coordination, allowing agents to speculatively commit state changes to a hardware-attested buffer and proceed with reasoning while a consistency quorum is reached in the background.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable non-blocking "Speculative Commits" for parallel agent teammates.
    * Maintain atomic integrity of the Shared Blackboard using versioned shadow-state.
    * Implement a 500ms "Consistency Window" for background conflict resolution.
    * Provide hardware-attested rollback triggers for failed speculations.
* **Non-Goals:**
    * Replacing the primary persistence layer (SQLite/Postgres).
    * Managing non-stateful tool side-effects (e.g., sending an email).
    * Resolving semantic intent conflicts; SSA focuses on data-layer consistency.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Resolve concurrent writes to a shared code-plan without stalling the reasoning loop.
* **The Happy Path (Tasks):**
    1. Agent A and Agent B both attempt to update the `mission_plan` key on the Blackboard simultaneously.
    2. SSA intercepts both requests and creates two versioned "Shadow Fragments" in the speculative buffer.
    3. Both agents receive a "Speculative OK" and continue their reasoning immediately.
    4. SSA background-triggers a "Consistency Quorum" between the Mission-Root and a Monitor Agent.
    5. The Quorum selects Agent A's update as the "Winning Fragment" based on priority rules.
    6. SSA merges Agent A's update into the persistent Blackboard and sends a "Rollback" signal to Agent B.
    7. Agent B's local context is reverted to the pre-speculative state using the Atomic State Rollback middleware.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B{SSA Broker}
        B -->|Speculative| C[Shadow Buffer]
        B -->|Synchronous| D[Blackboard Core]
        C --> E[Consistency Quorum]
        E -->|Success| D
        E -->|Conflict| F[Rollback Trigger]
        F --> G[Atomic State Rollback]
    ```
* **APIs / Interfaces:**
    * `ssa.SpeculativeWrite(key, value, missionToken) -> SpeculationID`: Initiates a speculative commit.
    * `ssa.CommitStatus(speculationID) -> Status`: Queries the progress of background attestation.
    * `ssa.ResolveConflict(strategy) -> void`: Configures the arbitration policy (e.g., "Parent Wins", "First In").
* **Data Storage/State:**
    * **Shadow Buffer:** Hardware-locked, in-memory KV store for pending fragments.
    * **Lineage Map:** Tracks the dependency graph of speculative reasoning paths.

## 5. Alternatives Considered
* **Strict Pessimistic Locking:** Rejected due to the 5s+ "Cognitive Stall" experienced in Claude Code teams.
* **Optimistic Concurrency Control (OCC):** Rejected because agents cannot handle "Write Conflicts" at the application layer without significant reasoning overhead; the infrastructure must handle the rollback.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative fragments are cryptographically isolated and cannot be read by siblings until committed.
* **Observability:** Integrated with the "Speculative State Inspector" UI for real-time visualization of buffer status and rollbacks.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
