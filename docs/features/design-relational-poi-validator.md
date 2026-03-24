# Design Doc: Relational PoI Validator
**Status:** Draft | In Review | Approved
**Created:** 2026-05-15

## 1. Context and Scope
The emergence of "Recursive Context Splicing" (RCS) vulnerabilities in 2026 has exposed a critical flaw in agentic delegation: current validation only checks individual mission intents, not the relationship between parent and child missions. Relational Proof-of-Intent (PoI) is needed to cryptographically validate the linkage between missions, ensuring subagents cannot be coerced into unauthorized sub-goals.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically validate the relational linkage between parent and child missions.
    * Detect and block "Recursive Context Splicing" (RCS) attacks during Binary State Handoff (BSH).
    * Enforce mission-root alignment for all delegated tasks.
* **Non-Goals:**
    * Replacing existing per-mission PoI validation.
    * Managing the underlying transport layer (UAB/BSH).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a malicious subagent from "splicing" an unauthorized secret-exfiltration goal into a legitimate code-review mission.
* **The Happy Path (Tasks):**
    1. Parent agent proposes a task to a child agent via UACO.
    2. The proposal includes a Relational PoI token signed by the parent's mission root.
    3. MCP Any's Relational PoI Validator intercepts the proposal.
    4. The validator verifies the signature and the relational metadata against the root mission intent.
    5. If aligned, the task is allowed; if an RCS attempt is detected (e.g., mismatched parentage), the delegation is blocked.

## 4. Design & Architecture
* **System Flow:**
    `[Parent Agent] -> [UACO Proposal w/ Relational Token] -> [Relational PoI Validator] -> [Child Agent]`
* **APIs / Interfaces:**
    * `Validator.verifyRelationalLinkage()`: Core method for validating mission tokens.
    * `TokenGenerator.createRelationalToken()`: Utility for parents to sign child delegations.
* **Data Storage/State:**
    * Transient storage of mission-tree metadata in the Shared KV Store.

## 5. Alternatives Considered
* **Flattened Mission IDs:** Rejected as it loses the relational hierarchy needed for RCS defense.
* **Manual HITL for all Delegations:** Rejected due to machine-speed coordination requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Token validation must use hardware-bound keys (TPM) to prevent signature forgery.
* **Observability:** Relational validation failures are logged in the "Context Chain Inspector."

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
