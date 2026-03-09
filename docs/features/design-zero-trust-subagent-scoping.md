# Design Doc: Zero-Trust Subagent Scoping
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As agents increasingly delegate tasks to subagents, the risk of privilege escalation grows. A parent agent with broad permissions (e.g., full filesystem access) might spawn a subagent that only needs to read a single file. Without granular scoping, a compromised or halluncinating subagent could abuse the parent's full permissions. Today's research on "Prompt-Induced Path Traversal" further emphasizes the need for strict, intent-bound boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement capability-based tokens that restrict subagents to a specific "intent-scope".
    * Enforce strict argument validation for subagents (no path traversal).
    * Provide a mechanism for parents to "attenuate" their permissions before delegating.
* **Non-Goals:**
    * Implementing OS-level sandboxing (this is middleware-level enforcement).
    * Managing LLM prompt engineering (focus on protocol and gateway enforcement).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Delegate a "File Analysis" task to a subagent, restricting it only to `analysis/results/` and blocking all write operations.
* **The Happy Path (Tasks):**
    1. Parent agent requests an "Attenuated Token" from MCP Any, specifying `scope: fs:read:analysis/results/`.
    2. MCP Any issues a scoped token.
    3. Parent agent passes the token to the Subagent.
    4. Subagent attempts to read `analysis/results/data.csv` -> **Success**.
    5. Subagent attempts to read `/etc/passwd` -> **Blocked** by MCP Any.
    6. Subagent attempts to write to `analysis/results/exploit.py` -> **Blocked** by MCP Any.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Parent[Parent Agent] -->|Request Scoped Token| Gateway[MCP Any Gateway]
        Gateway -->|Policy Check| Policy[Policy Engine]
        Policy -->|Issue Token| Gateway
        Gateway -->|Token| Parent
        Parent -->|Delegate| Subagent[Subagent]
        Subagent -->|Tool Call + Token| Gateway
        Gateway -->|Validate Scope| Policy
        Policy -->|Allow/Deny| Gateway
        Gateway -->|Execute| Upstream[Upstream Adapter]
    ```
* **APIs / Interfaces:**
    * `POST /auth/token/attenuate`: Takes a parent token and a list of requested scopes. Returns a new, more restricted token.
    * Header: `X-MCP-Scoped-Capability: <token>`
* **Data Storage/State:**
    * Scoped tokens are short-lived and stored in-memory (backed by Redis/SQLite for persistence if needed).

## 5. Alternatives Considered
* **Separate Configs per Subagent:** Rejected as too cumbersome and static.
* **LLM-Based Permission Enforcement:** Rejected because LLMs are susceptible to prompt injection and cannot be trusted for hard security boundaries.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The core value proposition. Uses the Principle of Least Privilege.
* **Observability:** Scoped calls are logged with a reference to the parent token for auditability.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation. Addressing the need for granular subagent permission attenuation and XAS mitigation.
