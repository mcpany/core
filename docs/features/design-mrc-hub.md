# Design Doc: Mesh-Resident Consensus (MRC) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms transition from centralized gateways to distributed P2P meshes (e.g., OpenClaw SNT), the "Single Point of Failure" and latency of centralized coordination have become primary bottlenecks. Agents in a mesh need a way to reach agreement on high-stakes operations, such as filesystem commits or tool execution, without hair-pinning through a central authority.

The Mesh-Resident Consensus (MRC) Hub provides a decentralized protocol for agents to negotiate and reach quorums directly within the transport layer. By leveraging hardware-attested "Voter Tokens" and CRDT-based state reconciliation, the MRC Hub ensures that the swarm maintains mission integrity even in fragmented network environments.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable decentralized quorum collection for high-risk agent actions.
    * Provide hardware-attested proof of consensus for all mesh-wide state commits.
    * Support dynamic "Voter" roles for specialized subagents (e.g., Security Auditor, Performance Monitor).
    * Minimize coordination latency by utilizing SNT-native transport.
* **Non-Goals:**
    * Implementing a general-purpose blockchain (MRC is task-bound and ephemeral).
    * Resolving non-semantic network partitions (handled by the underlying mesh transport).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Authorize a mission-critical database schema migration across three nodes without a central supervisor.
* **The Happy Path (Tasks):**
    1. The "Lead Agent" proposes a migration task card to the mesh.
    2. The MRC Hub identifies the required quorum (e.g., Lead + 2 Security Specialists).
    3. The Specialists review the task and issue cryptographically signed "Approve" tokens.
    4. The MRC Hub aggregates the tokens and verifies the hardware attestation.
    5. Once the quorum is reached, the MRC Hub issues a "Consensus Receipt" to all nodes.
    6. The migration executes simultaneously across the mesh, anchored by the receipt.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Proposing Agent] -->|Task Proposal| B[MRC Hub]
        B -->|Quorum Request| C[Voter Agent 1]
        B -->|Quorum Request| D[Voter Agent 2]
        C -->|Signed Vote| B
        D -->|Signed Vote| B
        B -->|Consensus Receipt| E[Mesh State Commit]
    ```
* **APIs / Interfaces:**
    * `mcp.mesh.propose(task_card)`: Initiates a consensus round.
    * `mcp.mesh.vote(proposal_id, decision)`: Casts a signed vote.
    * `mcp.mesh.get_receipt(proposal_id)`: Retrieves hardware-attested proof of consensus.
* **Data Storage/State:**
    * Ephemeral vote tallies are stored in task-bound CRDT shards within the mesh transport.

## 5. Alternatives Considered
* **Centralized Gatekeeper:** Rejected due to single-point-of-failure risks and high latency in geolocated meshes.
* **Gossip Protocols:** Rejected for high-stakes actions due to "Eventual Consistency" gaps; MRC requires strong consistency for commits.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Votes must be hardware-attested. A compromised agent can only cast one vote and cannot spoof its lineage.
* **Observability:** Consensus rounds are visualized in the `Service Mesh Topology Monitor`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
