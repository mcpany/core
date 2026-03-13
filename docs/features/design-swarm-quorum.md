# Design Doc: Swarm Quorum Attestation Hub
**Status:** Draft
**Created:** 2026-03-31

## 1. Context and Scope
With the rise of multi-agent refinement swarms (e.g., OpenClaw v2.7) and the discovery of context-mirroring exploits like "Mirror-Leak" (CVE-2026-48201), relying on a single agent's intent or a single human-in-the-loop (HITL) approval is increasingly risky. The Swarm Quorum Attestation Hub (SQAH) introduces a decentralized governance model where high-risk tool calls must be attested by a quorum of independent "Monitor" agents before execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Orchestrate multi-agent approval workflows for high-risk tool categories.
    * Support UACO v1.9 MAQ (Multi-Agent Quorum) compliant tokens.
    * Provide a threshold-based enforcement mechanism (e.g., 2 out of 3 monitors must approve).
    * Maintain an immutable audit trail of quorum decisions in the Federated Blackboard.
* **Non-Goals:**
    * Automatically generating the monitor agents (this is handled by the agent framework).
    * Replacing the primary PoI (Proof-of-Intent) validator.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Ensure that no subagent can modify production database schemas without consensus from at least two independent "Security Auditor" agents.
* **The Happy Path (Tasks):**
    1. A "DB Specialist" subagent requests a tool call to `DROP TABLE`.
    2. SQAH intercepts the request and identifies it as "High-Risk."
    3. SQAH broadcasts an "Attestation Request" to the configured Security Auditor swarm.
    4. Two Auditor agents analyze the request against the signed "Mission Anchor" and return signed approval tokens.
    5. SQAH verifies the tokens, achieves the threshold, and appends the "Quorum Proof" to the request headers.
    6. The tool call is dispatched to the target MCP server with full cryptographic backing.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [SQAH] -> [Threshold Check] -> [Broadcast to Monitors] -> [Collect Tokens] -> [Verify Quorum] -> [Upstream]`
* **APIs / Interfaces:**
    * `X-UACO-MAQ` header for transmitting quorum proofs.
    * `QuorumManager` internal service: `InitiateQuorum(requestID, policy)`, `SubmitAttestation(requestID, token)`, `GetQuorumStatus(requestID)`.
* **Data Storage/State:**
    * Pending quorum requests are stored in the Shared KV Store (Blackboard) with a session-bound TTL.
    * Final Quorum Proofs are persisted in the Immutable State Trail.

## 5. Alternatives Considered
* **Centralized HITL Only:** Rejected due to human bottleneck and latency in autonomous swarms.
* **Single-Agent "Supervisor" Model:** Rejected because the supervisor agent itself becomes a single point of failure/compromise.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Monitor agents must have distinct identities and cryptographic keys. SQAH must verify the `Origin` of all attestation tokens.
* **Observability:** Integrate with the "Consensus Attestation Workspace" in the UI to show real-time quorum progress.

## 7. Evolutionary Changelog
* **2026-03-31:** Initial Document Creation.
