# Design Doc: Recursive Depth-Limit Enforcer (RDLE)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms grow deeper (e.g., A -> B -> C -> D), managing delegation
depth becomes increasingly complex. Exploits like CVE-2026-71001 (Recursive
Shadow Handoffs) utilize nested UACO bids to bypass parent-imposed limits,
leading to "Delegation Escape" where subagents can recruit unauthorized peers.
The Recursive Depth-Limit Enforcer (RDLE) ensures that all delegations are
cryptographically bound to the mission-root manifest, providing non-bypassable
recursive depth control.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement cryptographically bound depth limits for all inter-agent
      delegations.
    * Mandate hardware-attested manifest validation for all UACO bid
      proposals.
    * Provide real-time "Shadow Handoff" detection and neutralization.
* **Non-Goals:**
    * Limiting the number of parallel teammates (handled by the Swarm
      Governer).
    * Restricting the type of tools a subagent can use (handled by the DPG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from recursively delegating
  its task to an unauthorized subagent via nested bids.
* **The Happy Path (Tasks):**
    1. Parent agent (Depth 0) delegates a task to Subagent A (Depth 1) with a
       manifest-locked limit of 2.
    2. Subagent A attempts to delegate a sub-task to Subagent B (Depth 2).
    3. RDLE verifies the mission-root manifest and allows the Depth 2
       delegation.
    4. Subagent B attempts to delegate to Subagent C (Depth 3).
    5. RDLE detects the Depth-Limit Violation and automatically blocks the
       delegation, revoking Subagent B's bid-capability token.

## 4. Design & Architecture
* **System Flow:** UACO Bid -> RDLE Validator -> Mission-Root Manifest Check ->
  Delegation Grant.
* **APIs / Interfaces:** `POST /api/v1/uaco/bid/validate` requiring a signed
  mission-root manifest and depth-token.
* **Data Storage/State:** Merkle-tree based manifest storage for efficient
  lineage verification.

## 5. Alternatives Considered
* **Implicit Depth Headers:** Rejected as they can be stripped or spoofed by
  misaligned agents.
* **Centralized Orchestration:** Rejected due to performance bottlenecks in
  decentralized meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All RDLE tokens must be hardware-attested and
  bound to the session-identity.
* **Observability:** All blocked delegations are logged for "Depth-Exhaustion"
  analysis.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
