# Design Doc: Recursive Intent Delegation (RID) Validator
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
As agent swarms grow deeper and more autonomous, the risk of "Intent Hijacking" and "Subagent Coercion" increases. A primary agent might authorize a subagent for a specific task, but that subagent could potentially spawn further agents or call tools that escalate its permissions beyond the original mission.

MCP Any needs to implement a **Recursive Intent Delegation (RID) Validator** that enforces depth limits and mutation boundaries on agent intents. This ensures that subagents remain strictly bound to the cryptographically signed intent of the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce maximum delegation depth for subagents.
    * Validate that subagent intents are semantically and cryptographically linked to the parent intent.
    * Provide a mechanism for parents to define "Intent Mutation Boundaries" (e.g., "read-only on this branch").
* **Non-Goals:**
    * Replacing the underlying LLM's reasoning process.
    * Managing the lifecycle of the agents themselves (handled by the Reaper).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Code Auditor" subagent from spawning an "Executor" subagent to bypass security gates.
* **The Happy Path (Tasks):**
    1. Parent agent issues a task to a subagent with a RID token (Depth: 1, Mutations: Allowed-Tools-Only).
    2. Subagent attempts to spawn a second subagent with "All Tools" access.
    3. RID Validator intercepts the spawn request, detects depth/boundary violation.
    4. RID Validator blocks the request and alerts the parent agent/user.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] -->|Intent + RID Token| B[RID Validator]
        B -->|Authorized| C[Subagent]
        C -->|Spawn Request| B
        B -->|Check Depth/Boundary| D{Valid?}
        D -->|Yes| E[Authorized Subagent]
        D -->|No| F[Security Violation Alert]
    ```
* **APIs / Interfaces:**
    * `validateRID(token: RIDToken, nextIntent: Intent): Result`
    * `issueRID(parentToken: RIDToken, constraints: Constraints): RIDToken`
* **Data Storage/State:**
    * RID tokens are stateless, carrying encrypted depth and boundary metadata.

## 5. Alternatives Considered
* **Flat Intent Checking:** Rejected because it doesn't account for the lineage of the agent, making it susceptible to "Intent Ghosting."
* **Centralized Session State:** Rejected due to scaling concerns in large, distributed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All RID tokens must be TPM-signed or hardware-attested.
* **Observability:** Every RID violation is logged with a full intent-chain trace.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
