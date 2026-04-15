# Design Doc: Recursive Mission Root (RMR) Orchestrator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve from linear delegations to complex, multi-hop nested hierarchies, the challenge of maintaining security policy consistency becomes critical. Current session-bound tokens often fail to propagate the full restrictive context of a parent mission, leading to "Constraint Leakage" where deeply nested specialist agents execute actions that the original mission-root user would have prohibited.

The RMR Orchestrator provides a standardized, cryptographic mechanism for nested swarms to carry sub-mission roots. These roots inherit the primary mission's identity but allow for further (recursive) restriction of capabilities, ensuring that every tool call, no matter how deep in the chain, is validated against a verifiable lineage of intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested, lineage-bound sub-mission tokens.
    * Enforce strict monotonicity: nested missions can only *reduce* privilege, never escalate.
    * Provide a cryptographic "Intent Chain" for every tool call.
    * Integrate with the A2A protocol for cross-framework nested attestation.
* **Non-Goals:**
    * Replacing existing session management (RMR extends it).
    * Providing long-term identity (RMR is mission-bound and ephemeral).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Orchestrator
* **Primary Goal:** Delegate a sensitive data-analysis task to a sub-swarm without allowing the sub-swarm to access the external network.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a sub-mission via the RMR Orchestrator.
    2. RMR Orchestrator generates a sub-mission token signed by the parent's mission-root key.
    3. The token includes a "Deny: Network" constraint injected by the parent.
    4. Sub-agent receives the token and attempts to call a network-enabled tool.
    5. MCP Any gateway validates the RMR token, sees the "Deny: Network" constraint, and interdicts the call.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] --(Request Sub-Mission)--> [RMR Orchestrator] --(Issue Sub-Token)--> [Nested Swarm]
    [Nested Swarm] --(Tool Call + Sub-Token)--> [MCP Any Gateway] --(Validate Lineage)--> [Local Tool]
* **APIs / Interfaces:**
    * `POST /v1/mission/subroot`: Create a nested mission root with inherited constraints.
    * `GET /v1/mission/validate`: Verify a sub-token's lineage and combined constraints.
* **Data Storage/State:**
    * Ephemeral mission-root keys stored in secure kernel memory.
    * Lineage metadata persisted in the Shared KV Store (Blackboard) with RMR-isolation.

## 5. Alternatives Considered
* **Flat Token Handoffs:** Rejected because it lacks lineage and cannot enforce monotonicity across untrusted sub-agents.
* **Centralized Policy Server:** Rejected to avoid single-point-of-failure and maintain mesh scalability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RMR tokens are TPM-signed and non-replayable. Every hop must re-attest to the orchestrator.
* **Observability:** Sub-mission branching is tracked in the Mesh-Resident Lineage Tracker for real-time visualization.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
