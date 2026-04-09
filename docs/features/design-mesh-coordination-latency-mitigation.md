# Design Doc: Non-blocking Mesh Coordination Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move from linear workflows to horizontal teammate collaboration (e.g., Claude Code Agent Teams), synchronous coordination has become the primary performance bottleneck. The current "Mailbox Lock" model, where teammates must wait for global consensus before claiming a task, frequently leads to "Cognitive Stalls" exceeding 5 seconds in high-density meshes.

The Non-blocking Mesh Coordination Arbiter introduces an "Optimistic State Synchronization" pattern. It allows specialist agents to speculatively claim tasks and begin reasoning locally while the arbiter performs background conflict resolution and hardware-attested validation. This architecture ensures that swarm reasoning remains continuous while maintaining mission-root consistency.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement an optimistic task-claiming protocol for horizontal Agent Teams.
    * Neutralize "Cognitive Stall" by decoupling task acquisition from global consensus.
    * Maintain mission-root consistency via asynchronous conflict resolution and rollbacks.
    * Support hardware-attested "Consistency Heartbeats" to detect divergent reasoning paths.
* **Non-Goals:**
    * Replacing the underlying A2A transport layer.
    * Managing the semantic generation of tasks (handled by the mission root).
    * Providing real-time prompt injection filtering (handled by the Injection Shield).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5+ specialist agents in a coding task without synchronization deadlocks.
* **The Happy Path (Tasks):**
    1. The mission root pushes a set of tasks to the sharded teammate mailbox.
    2. Agent A (Specialist) speculatively claims Task #1 and begins local reasoning immediately.
    3. The Arbiter receives the claim and performs a background check for hardware-attested authorization and state conflicts.
    4. Simultaneously, Agent B speculatively claims Task #2.
    5. The Arbiter confirms both claims asynchronously.
    6. If a conflict is detected (e.g., two agents claim the same task), the Arbiter issues a "Speculative Rollback" signal to the lower-priority agent.
    7. Agents commit their reasoning results to the global Blackboard only after final Arbiter attestation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Specialist
        participant Arbiter
        participant Blackboard

        Specialist->>Arbiter: Speculative Claim (TaskID, MissionToken)
        Note right of Specialist: Begin Reasoning Locally
        Arbiter->>Arbiter: Background Conflict Check
        Arbiter-->>Specialist: Claim Ack / Rollback Signal
        Specialist->>Arbiter: reasoning_fragment (Proposal)
        Arbiter->>Arbiter: Hardware-Attested Validation
        Arbiter->>Blackboard: Commit State
        Blackboard-->>Specialist: Commit Confirmed
    ```
* **APIs / Interfaces:**
    * `POST /mesh/claim/speculative`: Allows an agent to notify the arbiter of a tentative task claim.
    * `GET /mesh/consistency/heartbeat`: Periodic check for agents to verify their speculative state remains valid.
    * `POST /mesh/commit`: Final state submission after arbiter approval.
* **Data Storage/State:**
    * **Speculative Buffer:** Ephemeral storage for un-attested task assignments and reasoning fragments.
    * **Conflict-Free Replicated Data Types (CRDTs):** Used for the shared task list to ensure non-blocking updates.

## 5. Alternatives Considered
* **Strict Paxos/Raft Consensus:** Rejected because the latency of achieving quorum on every task claim is the root cause of the "Cognitive Stall."
* **Purely Local State:** Rejected because it leads to "Swarm Divergence" where agents waste tokens reasoning against outdated worldviews.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative claims are still bound by hardware-attested mission tokens. Unauthorized agents cannot claim tasks even optimistically.
* **Observability:** Integrated with the "Lock-Free Coordination Monitor" in the UI, visualizing speculative vs. committed task assignments in real-time.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
