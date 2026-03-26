# Design Doc: Recursive Mission Attestation (RMA) Provider
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
As agent swarms grow in depth, the "Delegation Chain" becomes a primary attack vector. Recent exploits (OpenClaw "Intent Splicing") demonstrate that subagents at the 4th or 5th hop can inject unauthorized high-priority instructions that appear to come from the parent.

RMA provides a cryptographic mechanism to verify that every instruction in a deep swarm is a direct, authorized descendant of the hardware-attested "Mission Root."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a cryptographic "Chain of Command" token for all tool calls.
    * Verify instruction lineage back to the TPM-signed mission root.
    * Prune "spliced" instructions that lack a valid parent signature.
* **Non-Goals:**
    * Performance optimization of the LLM reasoning itself.
    * Replacement of transport-layer encryption (mTLS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent subagent "Ghost Instructions" from triggering unauthorized `sudo` or `git push` actions.
* **The Happy Path (Tasks):**
    1. The user initiates a mission with a TPM-signed Root Intent.
    2. Parent Agent A delegates a task to Subagent B, including a signed Child Token.
    3. Subagent B calls a tool.
    4. MCP Any validates the tool call's RMA token against the parent lineage.
    5. The tool call is executed only if the chain is unbroken.

## 4. Design & Architecture
* **System Flow:**
    `Root (TPM Signed) -> Subagent A (Signed Branch) -> Subagent B (Signed Leaf) -> Tool Call (Verified)`
* **APIs / Interfaces:**
    - `POST /v1/rma/sign`: Issue a child token for delegation.
    - `POST /v1/rma/verify`: Validate a tool call's lineage chain.
* **Data Storage/State:**
    Chain signatures are stored in the session-bound RAMS (Reasoning-Aware Memory Segmentation).

## 5. Alternatives Considered
* **Flat Intent Verification**: Rejected because it cannot distinguish between legitimate parent instructions and "spliced" subagent noise.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RMA is a core Zero Trust component, moving security from "Identity" to "Lineage."
* **Observability:** Every RMA verification event is logged in the Command Traceability Provider (CTP).

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.
