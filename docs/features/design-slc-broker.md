# Design Doc: Swarm-Local Consensus (SLC) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Traditional AI agent swarms rely on a central supervisor or "Mission Root" to validate all tool outputs and state changes. As swarms become more parallel and high-density, this supervisor becomes a critical performance bottleneck and a single point of failure.

The Swarm-Local Consensus (SLC) Broker is required to facilitate decentralized, peer-voted quorums for tool results, allowing teammates to reach a consensus on truth before committing state, thereby eliminating the supervisor bottleneck.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate decentralized quorums for tool output validation among peer subagents.
    * Support hardware-attested approval tokens for every consensus vote.
    * Minimize coordination latency via "Speculative Consensus" patterns.
    * Provide framework-neutral consensus hooks for OpenClaw and Claude Code teammates.
* **Non-Goals:**
    * Replacing the Mission Root for high-level goal setting.
    * Managing global security policies; SLC operates within the bounds established by the Policy Firewall.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Verify a complex database schema migration across 3 specialist agents without waiting for the human supervisor.
* **The Happy Path (Tasks):**
    1. Agent A (DB Specialist) proposes a schema mutation.
    2. SLC Broker broadcasts the proposal to Agent B (Security Specialist) and Agent C (Performance Specialist).
    3. Peer agents review the proposal against their local context and "Skill Profiles".
    4. Peer agents submit hardware-attested "Approval Tokens" to the SLC Broker.
    5. Once a quorum (e.g., 2/3) is reached, the SLC Broker signs the "Consensus Seal".
    6. The mutation is committed to the mission-root blackboard.
    7. Results are reported back to the parent agent as "Peer-Verified".

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Proposing Agent] --> B[SLC Broker]
        B --> C[Peer Agent 1]
        B --> D[Peer Agent 2]
        C -->|Attested Vote| B
        D -->|Attested Vote| B
        B --> E{Quorum Met?}
        E -->|Yes| F[Consensus Seal]
        F --> G[Blackboard Commit]
    ```
* **APIs / Interfaces:**
    * `slc.ProposeAction(actionID, data) -> ProposalID`: Proposes a state change for consensus.
    * `slc.SubmitVote(proposalID, vote, attestation) -> Success`: Submits a peer vote.
    * `slc.GetConsensus(proposalID) -> ConsensusSeal`: Retrieves the final seal.
* **Data Storage/State:**
    * **Quorum Registry:** Tracking active proposals, votes, and participant reputations.

## 5. Alternatives Considered
* **Centralized Logic (Status Quo):** Rejected due to the "Supervisor Bottleneck" in horizontal swarms.
* **Byzantine Fault Tolerance (BFT):** Considered, but simplified to a hardware-attested quorum model to reduce computational overhead in local execution environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every vote is hardware-attested, preventing sybil attacks by compromised subagents.
* **Observability:** Consensus events are visualized in the "Swarm Consensus Dashboard".

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
