# Design Doc: Policy Firewall Engine
**Status:** In Review
**Created:** 2026-02-17

## 1. Context and Scope
In a "Universal Agent Bus" model, security cannot rely on simple API keys. Agents often perform high-risk actions (file deletion, database writes) that require fine-grained, context-aware control.

The Policy Firewall Engine provides a programmable security layer that evaluates every tool call against a set of Rego (OPA) or CEL (Common Expression Language) rules before execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Fine-grained control over tool parameters (e.g., "Allow `fs_write` only if path starts with `/tmp/agent-out`").
    * Context-aware policies (e.g., "Allow `delete_user` only if the agent is in 'Admin Mode' and it's during business hours").
    * High-performance evaluation (<5ms overhead per tool call).
* **Non-Goals:**
    * Replacing upstream authentication (OIDC/API Keys).
    * General purpose application authorization (focused specifically on MCP tool calls).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect.
* **Primary Goal:** Prevent an AI agent from accidentally deleting production data while allowing it to manage development resources.
* **The Happy Path (Tasks):**
    1. Architect defines a policy in Rego: `allow { input.tool == "db_delete"; startswith(input.args.table, "dev_") }`.
    2. Architect uploads the policy to MCP Any via config or API.
    3. Agent attempts to call `db_delete` on `prod_users`.
    4. Policy Firewall intercepts the call, evaluates the Rego, and returns a `AccessDenied` error to the agent.
    5. The event is logged to the Audit Trail with the specific policy ID that triggered the block.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCPAny: tools/call (db_delete, {table: "prod"})
        MCPAny->>Firewall: Evaluate(call_data)
        Firewall->>RegoEngine: input: {tool, args, context}
        RegoEngine-->>Firewall: Deny
        Firewall-->>MCPAny: Error: Policy Violation
        MCPAny-->>Agent: JSON-RPC Error -32003 (Forbidden)
    ```
* **APIs / Interfaces:**
    * `PolicyStore` interface for loading rules from YAML or remote Git repos.
    * Middleware hook in the `mcpserver` request pipeline.
* **Data Storage/State:**
    * Compiled Rego policies held in memory for fast execution.

## 5. Alternatives Considered
* **Hardcoded Rules**: Too rigid for complex agent behaviors.
* **Upstream IAM**: Often lacks the granularity to inspect JSON-RPC payloads (MCP tool arguments).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The Policy Engine itself must be isolated; a bug in a Rego script shouldn't crash the server.
* **Observability**: A "Policy Dry Run" mode in the UI to test rules against historical tool call logs.

## 7. Evolutionary Changelog
* **2026-02-17:** Initial Document Creation.
