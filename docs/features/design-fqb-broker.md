# Design Doc: Federated Quorum Broker (FQB)
**Status:** Draft
**Created:** 2026-05-06

## 1. Context and Scope
With the release of OpenClaw v2026.5.5, swarms are increasingly operating across organizational and domain boundaries. This creates a "Siloed Authority" problem where a subagent may need permission from multiple independent "Mission Roots" to perform a single high-stakes action. The Federated Quorum Broker (FQB) aims to be the neutral, multi-swarm coordination layer for negotiating and verifying these joint quorums.

## 2. Goals & Non-Goals
* **Goals:**
    * Orchestrate joint quorum negotiations between independent agent swarms.
    * Provide a standardized interface for cross-swarm signature collection (MAQ).
    * Enforce federated safety constraints defined in different mission-roots.
* **Non-Goals:**
    * Modifying individual swarm internal monologues.
    * Storing sensitive context within the broker (broker only handles metadata/signatures).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cross-Organizational Swarm Orchestrator.
* **Primary Goal:** Perform a data migration from Organization A's server to Organization B's database, requiring approvals from both Org A's Auditor and Org B's Security Monitor.
* **The Happy Path (Tasks):**
    1. Migration Agent (Subagent) initiates a task requiring a "Joint Quorum" (FQO).
    2. FQB identifies the required signatories based on the Federated Policy.
    3. FQB broadcasts a "Quorum Proposal" to Org A's Auditor and Org B's Security Monitor.
    4. Both signatories verify the reasoning path (RPW) and return cryptographically bound approval tokens.
    5. FQB verifies both signatures and issues a "Federated Quorum Receipt" to the Migration Agent.
    6. Migration Agent performs the task, presenting the receipt to the respective MCP servers.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
      participant Agent
      participant FQB
      participant OrgA Auditor
      participant OrgB Monitor
      Agent->>FQB: Quorum Proposal (Task, RPW)
      FQB->>OrgA Auditor: Request Signature
      FQB->>OrgB Monitor: Request Signature
      OrgA Auditor-->>FQB: Signature Token (A)
      OrgB Monitor-->>FQB: Signature Token (B)
      FQB->>FQB: Verify Joint Quorum
      FQB->>Agent: Federated Quorum Receipt
    ```
* **APIs / Interfaces:**
    * `InitiateFederatedQuorum(Task, Signatories[], RPWProof) -> ProposalID`
    * `SubmitQuorumSignature(ProposalID, SignatureToken)`
    * `GetFederatedQuorumStatus(ProposalID) -> Status (Pending/Approved/Denied)`
* **Data Storage/State:**
    * Persistent `Joint Quorum Attestation Ledger` for auditability.
    * Ephemeral proposal state in the `UACO Coordination Layer`.

## 5. Alternatives Considered
* **Individual Swarm Bidding (UACO v1.0):** Rejected because it doesn't handle the "Joint" nature of the authority; it treats approvals as individual, additive events rather than a unified consensus.
* **Direct Org-to-Org Negotiation:** Rejected as it requires N^2 integrations between all potential organizations; FQB provides a 1-to-N hub.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** FQB must ensure that all tokens are bound to the specific `ProposalID` and `RPWProof` to prevent "Quorum Hijacking."
* **Observability:** Real-time visualization in the `UACO Negotiation Dashboard`.

## 7. Evolutionary Changelog
* **2026-05-06:** Initial Document Creation.
