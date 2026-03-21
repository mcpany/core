# Design Doc: Multi-Agent Quorum (MAQ) Gateway
**Status:** Draft
**Created:** 2026-03-28

## 1. Context and Scope
As agents perform higher-risk actions (e.g., executing shell scripts, making financial transactions), relying on a single agent's reasoning (or even single-agent HITL) is insufficient. The UACO v1.9 Multi-Agent Quorum (MAQ) draft standardizes how multiple independent agents can provide cryptographically bound approval tokens for a single tool call. MCP Any must implement a MAQ Gateway to orchestrate this multi-framework consensus.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement the `MAQGateway` to orchestrate consensus flows between disparate agents (OpenClaw, AutoGen, etc.).
    * Support the UACO v1.9 MAQ token schema for multi-signature approvals.
    * Provide a "Quorum Requirement" policy that defines the threshold and types of agents required for specific tools.
    * Integrate with the A2A Bridge for broadcasting "Quorum Requests" to available monitors.
* **Non-Goals:**
    * Defining the reasoning logic for the monitor agents (they decide whether to approve).
    * Providing the final execution (handled by the target Upstream Adapter).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Swarm Architect
* **Primary Goal:** Require an "Audit Quorum" of two independent monitor agents before a subagent is allowed to delete files from a restricted directory.
* **The Happy Path (Tasks):**
    1. Subagent A attempts to call `fs_delete` on a restricted path.
    2. MAQ Gateway intercepts the call and identifies a "Quorum Requirement" of 2 monitors.
    3. MAQ Gateway broadcasts a "Quorum Request" via UACO-MAQ to all active Monitor agents.
    4. Monitor B (OpenClaw) and Monitor C (AutoGen) both submit "Approval Tokens."
    5. MAQ Gateway validates the tokens, aggregates the multi-signature, and authorizes the `fs_delete` call.

## 4. Design & Architecture
* **System Flow:**
    `Tool Call` -> `Policy Check` -> `Quorum Request (UACO-MAQ)` -> `A2A Broadcast` -> `Approval Collection` -> `Multi-Sig Validation` -> `Execution Authorization`
* **APIs / Interfaces:**
    * `QuorumManager` Interface: `InitiateQuorum(req *MAQRequest) (*MAQToken, error)`, `CollectApproval(token *MAQApproval) error`
    * `QuorumPolicy`: Definitions of `required_threshold` and `required_agent_roles` per tool/service.
* **Data Storage/State:**
    * Pending quorum requests and collected tokens are stored in the Shared KV Store (Blackboard) with "Quorum-Scoped" isolation.

## 5. Alternatives Considered
* **Sequential HITL**: Rejected as it is too slow and doesn't provide the cryptographic integrity of a parallel multi-agent quorum.
* **Single-Framework Governance**: Rejected because modern swarms are increasingly heterogeneous (multi-framework).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All approval tokens must be cryptographically signed and tied to the specific "Intent ID" of the original tool call.
* **Observability:** Quorum flows (requests, participants, approvals, timeouts) are visualized in the "Consensus Attestation Workspace."

## 7. Evolutionary Changelog
* **2026-03-28:** Initial Document Creation.

### Update: 2026-03-29 - Identity Shadowing Defense
**Context:** Today's research revealed CVE-2026-45001 (Identity Shadowing) in the MAQ implementation and the release of the UACO v2.0 RIS candidate.
**Architecture Adjustment:**
* Deprecating flat session nonces for token generation.
* Introducing **Relational Intent Scoping (RIS)**: Approval tokens must now include the hierarchical `Intent-Path` from the UACO v2.0 tree structure.
* The MAQ Gateway will now validate that the requester's intent branch is a legitimate child of the branch that authorized the original quorum.
**Security Impact:** Prevents malicious subagents from reusing parent signatures across unauthorized sub-delegations.
