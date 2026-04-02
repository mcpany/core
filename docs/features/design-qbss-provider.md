# Design Doc: Quorum-Based Skill Scoring (QBSS) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As the agent ecosystem transitions from individual node environments to massive multi-agent meshes, individual node-local behavioral profiling is no longer sufficient. A malicious skill may exhibit "Low-and-Slow" behavior that passes a single node's security gate but collectively threatens the swarm's mission root.

The Quorum-Based Skill Scoring (QBSS) Provider is required to implement a decentralized reputation system where hardware-attested nodes in the mesh can share, aggregate, and vote on the safety and reliability scores of MCP tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate decentralized collection of tool safety scores from mesh-resident nodes.
    * Require hardware-attested (TPM) signatures for all reputation votes.
    * Provide real-time reputation queries for the Discovery Bus.
    * Implement "Audit-before-Consensus" for high-risk capabilities.
* **Non-Goals:**
    * Serving as a general-purpose social network for agents.
    * Replacing local Behavioral Skill Profiling; it acts as an aggregate signal.
    * Managing the underlying P2P transport (delegated to AMT Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Administrator
* **Primary Goal:** Prevent a "Rug-Pull" malicious skill from being executed across the enterprise mesh.
* **The Happy Path (Tasks):**
    1. A new MCP tool is discovered by a node in the mesh.
    2. The node queries the QBSS Provider for the tool's aggregate reputation.
    3. The QBSS Provider aggregates votes from 10+ independent, hardware-verified nodes.
    4. If the score is below the "Trust Quorum" threshold (e.g., 0.8), the tool is automatically quarantined.
    5. A node that performs an audit of the tool submits a TPM-signed "Safety Receipt" to the provider.
    6. The aggregate score is updated in real-time across the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Node A - Tool Audit] -->|TPM-Signed Vote| B(QBSS Provider)
        C[Node B - Tool Usage] -->|Success/Failure Signal| B
        D[Discovery Bus] -->|Query Score| B
        B -->|Aggregate Reputation| D
        subgraph Mesh Consensus
            B
        end
    ```
* **APIs / Interfaces:**
    * `qbss.SubmitVote(toolID, score, signature) -> Status`: Submits a hardware-attested vote.
    * `qbss.GetReputation(toolID) -> ReputationSummary`: Retrieves the collective score and confidence level.
    * `qbss.RequestAudit(toolID) -> AuditID`: Triggers a swarm-wide audit request for a tool with low confidence.
* **Data Storage/State:**
    * **Reputation Ledger:** A tamper-resistant local cache of hardware-attested votes.
    * **Consensus Engine:** Logic for weighting votes based on node reputation and hardware strength.

## 5. Alternatives Considered
* **Centralized Registry (e.g., NPM for Agents):** Rejected due to single-point-of-failure and potential for registry-level instruction poisoning. Mesh-wide consensus provides higher resilience.
* **Local-Only Profiling:** Rejected because it cannot detect coordinated attacks across disparate sessions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All votes must be cryptographically bound to a verified hardware ID to prevent "Sybil Attacks" by un-attested subagents.
* **Observability:** Integrated with the "UAB Reputation Explorer" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
