# Design Doc: Contextual Quorum (CQ) Hub
**Status:** Draft
**Created:** 2026-05-02

## 1. Context and Scope
As agent swarms become more autonomous and specialized, relying on a single agent's reasoning or a static policy for high-risk actions (e.g., executing code, modifying security settings, making financial transfers) is no longer sufficient. The release of OpenClaw v2026.5.0's "Contextual Quorum" (CQ) signals a move toward collective, agent-in-the-loop validation. MCP Any needs to evolve from a simple tool gateway into a "Collective Attestation Hub" that can orchestrate these multi-agent votes and enforce consensus before action execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Orchestrate multi-agent quorums where high-risk actions require cryptographically bound approval tokens from independent subagents.
    * Support dynamic quorum definitions based on the "Reasoning Effort" and "Semantic Risk" of the intent.
    * Provide a standardized "Quorum Proposal" interface for agents to request peer review.
    * Maintain a "Consensus Trail" for audit and forensic analysis.
* **Non-Goals:**
    * Defining the internal reasoning logic of the monitor/auditor agents.
    * Replacing existing HITL (Human-in-the-Loop) flows; CQ serves as an additional layer of automated peer review.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure that no subagent can modify the project's production environment without approval from both a "Security Monitor" agent and a "Compliance Auditor" agent.
* **The Happy Path (Tasks):**
    1. Subagent A proposes a high-risk tool call (e.g., `deploy_to_prod`).
    2. CQ Hub intercepts the call and identifies the requirement for a "Security & Compliance" quorum.
    3. CQ Hub broadcasts a "Quorum Proposal" to the designated Monitor and Auditor subagents.
    4. The Monitor and Auditor agents review the proposal (and the associated Verifiable Reasoning Trace) and submit their signed "Approval Tokens."
    5. CQ Hub validates the consensus and releases the tool execution lease.

## 4. Design & Architecture
* **System Flow:**
    `Agent Tool Call` -> `Policy Engine (CQ Check)` -> `Quorum Hub (Initiate)` -> `A2A Messaging (Broadcast)` -> `Collect Signatures` -> `Consensus Validation` -> `Execution`
* **APIs / Interfaces:**
    * `CQManager`: `ProposeQuorum(intentID string, req QuorumRequirement) (proposalID string, err error)`
    * `CQListener`: `OnQuorumProposal(proposal QuorumProposal) (ApprovalToken, error)`
    * `QuorumHub`: `SubmitToken(proposalID string, token ApprovalToken) error`
* **Data Storage/State:**
    * Quorum states (proposals, signatures, status) are stored in the Shared KV Store with `Quorum` scope isolation.
    * Consensus trails are appended to the mission-root audit log.

## 5. Alternatives Considered
* **Static HITL Only**: Rejected because human approval becomes a bottleneck in high-frequency, complex swarms.
* **Single-Agent Security Nodes**: Rejected because it creates a single point of failure; collective quorums provide "Defense in Depth."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All Approval Tokens must be hardware-bound (TPM/SEP) and cryptographically linked to the specific Intent ID and Reasoning Trace.
* **Observability:** Quorum status and participant monologues are visualized in the "Contextual Quorum Dashboard."

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.
