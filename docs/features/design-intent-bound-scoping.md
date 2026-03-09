# Design Doc: Intent-Bound Dynamic Tool Scoping
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
The "Universal Agent Bus" is increasingly used in multi-agent swarms. A parent agent might have broad permissions (e.g., access to GitHub, AWS, and local FS), but it might only want its subagent to have access to a specific GitHub repo for a limited time. Currently, MCP Any provides static capability tokens. Dynamic scoping will allow a parent to "mint" a restricted token for a specific sub-task.

## 2. Goals & Non-Goals
* **Goals:**
    * Create temporary, short-lived "Intent Tokens" that scope tool access.
    * Enable parent agents to define a subset of tools for a child session.
    * Limit tools to specific arguments or resources (e.g., `fs:read` restricted to `/tmp/task-1`).
* **Non-Goals:**
    * Implementing a full-blown IAM system (too complex for the "local adapter" use case).
    * Managing parent agent credentials (parent must already be authenticated).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw)
* **Primary Goal:** Share a secure, limited context with a subagent.
* **The Happy Path (Tasks):**
    1. Parent agent calls `mint_scoped_token` with a list of tools (e.g., `list_files`, `read_file`) and restrictions (e.g., `path: /tmp/task-1`).
    2. MCP Any generates a JWT-based "Intent Token" with a short TTL (e.g., 5 mins).
    3. Parent agent passes the token to the subagent.
    4. Subagent uses the token for tool calls.
    5. MCP Any's Policy Engine verifies the token and enforces the "Intent Scope."

## 4. Design & Architecture
* **System Flow:**
    - The `PolicyEngine` will be updated to handle `IntentToken` validation.
    - A new `SessionManager` middleware will track active "Intent Scopes."
    - The `Policy Firewall` will use the token's claims to filter the `ServiceRegistry` results.
* **APIs / Interfaces:**
    - `POST /v1/tokens/mint` (or an MCP tool `mint_intent_token`).
    - The token will contain a `scope` claim: `{ "tools": ["fs:read"], "constraints": {"path": "/tmp/*"} }`.
* **Data Storage/State:**
    - Short-lived tokens are stateless (standard JWT).

## 5. Alternatives Considered
* **Sticky Sessions:** Rejected because it makes scaling across multiple MCP Any nodes difficult.
* **Static Config:** Too rigid for dynamic agent behaviors.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If a subagent is compromised, it only has access to its restricted scope, not the parent's full credentials.
* **Observability:** Audit logs will record tool calls under the specific `Intent ID`.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
