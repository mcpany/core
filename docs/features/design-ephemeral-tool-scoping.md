# Design Doc: Ephemeral Tool Scoping
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
As agent swarms become more complex, subagents are often granted broad tool permissions by their parents. Currently, these permissions often persist for the duration of a session, creating a large attack surface if a subagent is compromised or suffers from "hallucinated intent." Ephemeral Tool Scoping aims to minimize this risk by providing short-lived, task-specific capabilities.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mechanism to issue tool tokens that expire after a single task or a very short TTL.
    * Cryptographically bind tokens to a specific subagent ID and parent intent.
    * Provide automatic revocation upon task completion signal.
* **Non-Goals:**
    * Replacing the underlying Rego/CEL policy engine (it will build upon it).
    * Managing persistent user-level permissions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Grant a "Researcher" subagent read-only access to a specific `/docs/project-x` directory for exactly 5 minutes or until it submits its report.
* **The Happy Path (Tasks):**
    1. Parent agent requests an ephemeral token for `fs:read` scoped to `/docs/project-x`.
    2. MCP Any issues a signed JWS (JSON Web Signature) token with an `exp` claim and a `task_id`.
    3. The Researcher subagent uses this token to call the `read_file` tool.
    4. MCP Any validates the token and the path constraint.
    5. Once the Researcher subagent returns its output, the parent signals task completion.
    6. MCP Any invalidates the token immediately.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent -> Token Issuer Service (MCP Any) -> Signed Ephemeral Token -> Subagent -> Tool Execution Middleware -> Policy Engine -> Upstream Tool`
* **APIs / Interfaces:**
    * `POST /v1/tokens/ephemeral`: Request a new scoped token.
    * `POST /v1/tokens/revoke`: Manually revoke a token/task.
* **Data Storage/State:**
    * In-memory cache (Redis/LRU) for active ephemeral tokens to allow for O(1) revocation checks.

## 5. Alternatives Considered
* **Short-lived standard JWTs**: Rejected because they lack the ability to be revoked immediately upon a "task complete" signal without a centralized blacklist.
* **Dynamic Rego Policy Updates**: Too slow and complex for high-frequency subagent spawning.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Tokens are signed with an internal private key. Any attempt to use a token for a path outside its scope results in immediate session termination.
* **Observability**: All ephemeral token issuance and usage are logged to the audit trail with `parent_id` and `subagent_id` context.

## 7. Evolutionary Changelog
* **2026-03-08:** Initial Document Creation.
