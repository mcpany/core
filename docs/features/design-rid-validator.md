# Design Doc: Recursive Intent Delegation (RID) Validator
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
As agent swarms become deeper and more autonomous, the risk of "Intent Hijacking" increases. A parent agent may delegate a task to a subagent, but without strict boundaries, that subagent could be coerced by malicious tool output or peer agents to escalate its own permissions or deviate from the original mission.

MCP Any needs to implement the UACO v1.8 RID standard to act as the authoritative gatekeeper for recursive agency, ensuring that every sub-delegation remains within cryptographically signed boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement validation for UACO v1.8 RID headers.
    * Enforce parent-defined recursion depth limits.
    * Enforce intent-mutation boundaries (e.g., read-only sub-intents).
    * Provide a cryptographic audit trail of the intent chain.
* **Non-Goals:**
    * Automatically generating intents for agents.
    * Managing the internal reasoning state of subagents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Code Reviewer" subagent from spawning a "Deployment" subagent with unauthorized credentials.
* **The Happy Path (Tasks):**
    1. Parent agent signs a task card for the "Code Reviewer" with `rid_depth: 1` and `allow_mutation: false`.
    2. The subagent attempts to delegate a further task.
    3. MCP Any RID Validator intercepts the request.
    4. Validator identifies that `rid_depth` is exceeded or mutation is unauthorized.
    5. MCP Any blocks the request and alerts the parent agent/user.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Agent A] -> [RID Middleware] -> [Registry] -> [Agent B]
                      |
              [Lineage Store] -> [Validation Logic]
    ```
* **APIs / Interfaces:**
    * `ValidateIntent(token RIDToken) error`: Core validation function.
    * `rid_depth`: Integer in the task metadata.
    * `mutation_mask`: Bitmask defining allowed intent changes.
* **Data Storage/State:**
    * Lineage is stored in the Shared KV Store (Blackboard) under protected keys.

## 5. Alternatives Considered
* **Flat Token Inheritance:** Rejected because it doesn't allow for depth restriction, leading to infinite spawning vulnerabilities.
* **Stateless Validation:** Rejected because verifying the entire chain of custody requires access to the mission-root lineage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All RID tokens must be TPM-signed.
* **Observability:** Every RID violation is logged with the full intent lineage and the identity of the violating agent.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
