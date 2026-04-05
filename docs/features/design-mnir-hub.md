# Design Doc: Multi-Node Intent Reconciliation (MNIR) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale across distributed Sovereign Nodes (e.g., OpenClaw v3.7.0), coordination becomes a bottleneck. Traditional hierarchical reconciliation requires every subagent conflict to be escalated to a central mission-root supervisor, introducing significant latency and "Resolution Stall," especially over high-latency P2P tunnels.

The Multi-Node Intent Reconciliation (MNIR) Hub solves this by providing a decentralized "Peer Quorum" model. It allows sibling agents on disparate physical nodes to reach consensus on state transitions and intent alignment locally, ensuring high-speed coordination without compromising the security or integrity of the primary mission-root.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate decentralized peer-to-peer voting for intent reconciliation across distributed nodes.
    * Reduce Mean Time to Coordinate (MTTC) by resolving conflicts locally within an authorized peer group.
    * Enforce hardware-attested vote integrity using TPM signatures.
    * Maintain mission-root sovereignty by ensuring peer consensus remains within pre-defined mission boundaries.
* **Non-Goals:**
    * Replacing the mission-root as the ultimate authority for high-risk system-wide changes.
    * Managing the underlying P2P network layer (handled by AMT Broker).
    * Synchronizing non-agent application state.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Resolve a state collision between two specialist agents on different devices without supervisor intervention.
* **The Happy Path (Tasks):**
    1. Agent A (Node 1) and Agent B (Node 2) attempt to update the same key on the shared Blackboard.
    2. The MNIR Hub detects the conflict and initiates a "Peer Quorum" session.
    3. The Hub identifies a set of authorized sibling agents across the mesh to act as "Peer Voters."
    4. Sibling agents review the proposed intents and cast hardware-attested votes.
    5. The MNIR Hub reconciles the votes and commits the "Winning Intent" to the distributed Blackboard.
    6. Both Agent A and Agent B receive the synchronized state update, resolving the collision in <500ms.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        subgraph Mesh Nodes
            A[Agent A] --> Hub1[MNIR Hub]
            B[Agent B] --> Hub2[MNIR Hub]
            C[Peer Agent] --> Hub3[MNIR Hub]
        end
        Hub1 <==>|DPV Protocol| Hub2
        Hub2 <==>|DPV Protocol| Hub3
        Hub1 <==>|DPV Protocol| Hub3
        Hub1 --> BB[Distributed Blackboard]
    ```
* **APIs / Interfaces:**
    * `mnir.InitiateReconciliation(conflictID, intentFragments) -> QuorumID`: Starts a voting session.
    * `mnir.CastPeerVote(quorumID, voteToken) -> Receipt`: Authenticated vote submission.
    * `mnir.GetWinningIntent(quorumID) -> ResolvedIntent`: Retrieves the outcome.
* **Data Storage/State:**
    * **Peer Registry:** In-memory map of authenticated sibling agents and their node IDs.
    * **Quorum Buffer:** Temporary storage for pending votes and conflict metadata.

## 5. Alternatives Considered
* **Centralized Supervisor Arbitration:** Rejected due to the "Resolution Stall" in high-latency environments and the single point of failure risk.
* **Optimistic Concurrency Control:** Rejected because "Last-Writer-Wins" frequently leads to "Intent Erasure" in complex reasoning chains.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every vote must be signed by a node-local TPM. Peer groups are strictly limited to agents sharing the same mission-root lineage.
* **Observability:** Integrated with the "Intent Reconciliation Hub" in the UI for real-time visualization of voting progress and conflict density.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
