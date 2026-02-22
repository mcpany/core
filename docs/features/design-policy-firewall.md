# Design Doc: Policy Firewall Engine

**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw that can execute local commands and manage files, there is a critical need for a security layer that sits between the agent and the system. Existing MCP implementations often grant "all-or-nothing" permissions. The Policy Firewall Engine (PFE) will provide fine-grained, rule-based control over every tool call.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce Zero Trust principles for all tool executions.
    *   Support Rego (OPA) or CEL (Common Expression Language) for policy definitions.
    *   Intercept `tools/call` requests and validate them against active policies.
    *   Provide audit logs for all blocked and allowed actions.
*   **Non-Goals:**
    *   Implementing the actual tool logic (handled by upstream adapters).
    *   Managing user authentication (handled by the core Auth layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Prevent an autonomous agent from executing `rm -rf /` or accessing `/etc/shadow` while allowing it to manage files in a specific project directory.
*   **The Happy Path (Tasks):**
    1.  Architect defines a Rego policy restricting `fs` tools to `/home/user/project`.
    2.  Agent attempts to call `fs_read` on `/etc/shadow`.
    3.  PFE intercepts the call, evaluates the policy, and returns a "Policy Violation" error to the agent.
    4.  Agent attempts to call `fs_read` on `/home/user/project/README.md`.
    5.  PFE evaluates the policy, allows the call, and proxies it to the Upstream Adapter.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Agent[AI Agent] -->|tools/call| Core[MCP Any Core]
        Core -->|Intercept| PFE[Policy Firewall Engine]
        PFE -->|Check| Rego[Rego/CEL Engine]
        Rego -->|Result| PFE
        PFE -->|Allow| Adapter[Upstream Adapter]
        PFE -->|Block| Core
        Core -->|Error| Agent
    ```
*   **APIs / Interfaces:**
    *   Internal `Middleware` interface implementation.
    *   `POST /api/v1/policies`: Manage policy definitions.
*   **Data Storage/State:**
    *   Policies stored in the main SQLite database.
    *   In-memory cache for high-performance policy evaluation.

## 5. Alternatives Considered
*   **Static Config Blocks**: Too rigid for complex, dynamic agent behaviors.
*   **App-Level Hardcoding**: Not maintainable or extensible for a "Universal" adapter.
*   **Selected**: Rego/CEL provides the industry-standard flexibility required for "Zero Trust".

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The PFE is the core of the Zero Trust strategy. Fail-closed by default.
*   **Observability:** Every policy evaluation is logged with detailed context (Input, Policy ID, Decision, Latency).

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
