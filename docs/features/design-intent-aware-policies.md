# Design Doc: Intent-Aware Tool Policies

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As agent swarms become more complex, static capability-based permissions (e.g., "can read /tmp") are insufficient. If a subagent is hijacked or malfunctions, it could use its broad permissions to perform actions outside its intended task. Inspired by Gemini CLI's `SessionContext` and the need to mitigate "Session Hijacking" risks, MCP Any needs a way to bind tool authorizations to a specific, cryptographically signed "Intent."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement an "Intent Manifest" that defines the scope of work for a specific agent session.
    *   Require tool calls to be accompanied by an `IntentToken` signed by the orchestrator.
    *   Validate that the tool being called and its arguments align with the signed intent.
    *   Integrate with the existing Policy Firewall (Rego/CEL).
*   **Non-Goals:**
    *   Predicting agent intent (intent must be explicitly provided by the orchestrator or user).
    *   Replacing static permissions (Intent-Aware policies sit on top of static ones).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Ensure a "File Researcher" subagent can only read files related to the current project, even if it has broad filesystem read access.
*   **The Happy Path (Tasks):**
    1.  Orchestrator creates a session and generates an `IntentToken` signed with the instance key, whitelisting `fs:read` for files matching `project_alpha/*`.
    2.  Subagent receives the `IntentToken`.
    3.  Subagent calls `read_file(path="project_alpha/README.md")` and includes the token.
    4.  MCP Any Policy Middleware verifies the signature and checks the path against the manifest.
    5.  Call is allowed.
    6.  If the subagent attempts `read_file(path="~/.ssh/id_rsa")`, the call is blocked because it falls outside the "Intent scope."

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Issuance**: The orchestrator uses the `mcpany` API to issue a short-lived, scoped token.
    - **Middleware Interception**: The `PolicyMiddleware` extracts the `X-MCP-Intent-Token` header.
    - **Verification**: The token is verified against the instance's public key. The manifest (JSON) is then evaluated against the incoming RPC request.
*   **APIs / Interfaces:**
    - `POST /v1/intent/issue`: Issues a new signed intent token.
    - RPC Header: `X-MCP-Intent-Token`.
*   **Data Storage/State:** Tokens are stateless (JWT-like), but active session IDs might be tracked in the Shared KV Store for revocation.

## 5. Alternatives Considered
*   **Dynamic Rego Policy Generation**: Re-generating Rego policies for every subagent task. *Rejected* as too slow and complex to manage at scale.
*   **Manual HITL for Every Call**: Asking the user for every sensitive call. *Rejected* as it breaks autonomy for swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This provides "Least Privilege" at the temporal/task level, not just the resource level.
*   **Observability:** Audit logs must include the `IntentID` and whether a call was blocked due to an "Intent Mismatch."

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
