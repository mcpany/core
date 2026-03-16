# Design Doc: Contextual Quorum (CQ) Hub
**Status:** Draft
**Created:** 2026-05-02

## 1. Context and Scope
With the release of OpenClaw v2026.5.0, the concept of "Contextual Quorum" (CQ) has emerged as a critical requirement for high-stakes agentic actions. Unlike traditional HITL, CQ requires a consensus from a "Quorum" of independent, specialized subagents (e.g., a "Security Auditor" and a "Privacy Monitor") before a task is authorized. MCP Any needs a centralized "CQ Hub" to orchestrate these multi-agent votes and manage the cryptographic attestation tokens.

## 2. Goals & Non-Goals
* **Goals:**
    * Orchestrate multi-agent consensus flows using the UACO v3.1 CQ protocol.
    * Provide a "Quorum Policy Engine" to define required signatures for specific tool categories.
    * Aggregate independent attestation tokens into a single, verifiable "Quorum Proof."
    * Support "Monitor" and "Auditor" subagent roles with specific capability-based weighting.
* **Non-Goals:**
    * Implementing the internal reasoning of monitor agents.
    * Direct execution of tools (delegated to the Tool Adapter layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure that any tool call involving "Financial Transaction" requires approval from both an `Accountant` agent and a `Compliance` agent.
* **The Happy Path (Tasks):**
    1. A primary agent proposes a tool call to `stripe_payout`.
    2. The Policy Firewall identifies a CQ requirement for `Accountant` and `Compliance` roles.
    3. The CQ Hub initiates a "Quorum Session" and broadcasts the request to the A2A Messaging Hub.
    4. The `Accountant` agent and `Compliance` agent both submit signed approval tokens.
    5. The CQ Hub validates the signatures, aggregates them into a "Quorum Proof," and authorizes the transaction.

## 4. Design & Architecture
* **System Flow:**
    `Primary Intent` -> `CQ Hub Policy Check` -> `Quorum Request Broadcast` -> `Asynchronous Approval Collection` -> `Proof Aggregation` -> `Tool Authorization`
* **APIs / Interfaces:**
    * `CQManager`: `CreateSession(intentID string, policy QuorumPolicy) (sessionID string, err error)`
    * `QuorumListener`: `SubmitApproval(sessionID string, agentToken Token) error`
* **Data Storage/State:**
    * Quorum sessions and tokens are stored in a versioned branch of the Blackboard with "Consensus-Locked" visibility.

## 5. Alternatives Considered
* **Centralized Auditor**: Rejected as it creates a single point of failure and doesn't leverage the specialized intelligence of distributed subagents.
* **Synchronous HITL**: Rejected due to high latency and inability to scale with autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Quorum Proofs must be hardware-bound (TPM/SEP) to the participating agents' identities to prevent spoofing.
* **Observability:** All quorum activities (votes, dissent, timeouts) are logged in the "Consensus Attestation Workspace" for post-mortem auditing.

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.
