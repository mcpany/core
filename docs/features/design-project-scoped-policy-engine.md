# Design Doc: Project-Scoped Policy Engine
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As agents are deployed across varied organizational boundaries, a flat security model is no longer sufficient. Agents operating on "Project A" should not have access to tools or data reserved for "Project B," even if both are hosted on the same MCP Any instance.

The Project-Scoped Policy Engine introduces a hierarchical evaluation model where policies are applied based on the agent's active project context, enabling secure multi-tenancy and better resource isolation.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hierarchical policy evaluation: Global > Project > Agent.
    * Support "Inheritance" and "Override" semantics for Rego/CEL policies.
    * Automatically filter tool availability based on the active `project_id`.
    * Provide a standardized way for agents to declare their current project context via MCP headers.
* **Non-Goals:**
    * Implementing identity providers (LDAP/OIDC) directly (it should integrate with existing ones).
    * Enforcing filesystem-level isolation (this is handled by the OS/Docker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Platform Admin.
* **Primary Goal:** Restrict a "Financial Analyst Agent" to only use the "Excel-Read" tool when working on the "Annual-Audit" project.
* **The Happy Path (Tasks):**
    1. Admin defines a Project Policy for `project: annual-audit` that allows `excel:read` but denies `excel:write`.
    2. The Agent initiates a session with the header `mcp-project-id: annual-audit`.
    3. MCP Any identifies the project and loads the corresponding policy.
    4. When the Agent calls `excel:write`, the Policy Engine denies the request based on the project-specific override, even if the Global policy allows it.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B{Policy Evaluator}
        B --> C[Global Policies]
        B --> D[Project-Specific Policies]
        B --> E[Agent-Specific Policies]
        D --> F{Project ID Match?}
        F -- Yes --> G[Apply Overrides]
        G --> H[Final Decision]
    ```
* **APIs / Interfaces:**
    * New header support: `X-MCP-Project-ID`.
    * Policy definition schema update to include `scope: { type: "project", id: "..." }`.
* **Data Storage/State:** Policies are stored as `.rego` or `.yaml` files in a structured directory: `policies/global/`, `policies/projects/[id]/`.

## 5. Alternatives Considered
* **Namespace-Based Isolation**: Creating separate MCP Any instances per project. *Rejected* due to high resource overhead and difficulty in managing cross-project global rules.
* **Tag-Based Filtering**: Using simple tags on tools. *Rejected* because tags don't support complex logic (e.g., "Allow if user is X and project is Y").

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Project scoping is a core requirement for Zero Trust in multi-agent environments. It prevents "lateral movement" between projects by a compromised agent.
* **Observability:** Logs must include the `project_id` for every tool call and policy decision.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
