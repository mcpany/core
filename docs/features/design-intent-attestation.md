# Design Doc: Intent-Scoped Attestation (ISA)
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As agents evolve to use "Generalist" orchestrators and "Programmatic Tool Calling," the risk of the "Confused Deputy" problem increases. A trusted orchestrator might be tricked into delegating a high-privilege task to a subagent or tool without verifying the underlying intent. MCP Any needs a mechanism to ensure that every tool call is backed by a cryptographically signed intent that aligns with the user's original goal.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a signed "Intent Token" that must accompany high-privilege tool calls.
    * Provide a middleware to verify Intent Tokens against a Rego-based Policy Engine.
    * Enable orchestrators to "delegate" specific intents to subagents with a limited TTL.
* **Non-Goals:**
    * Replacing existing OIDC/OAuth authentication.
    * Implementing a full-blown identity provider.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Architect
* **Primary Goal:** Prevent a subagent from deleting a production database even if it has the `db:delete` capability, unless that specific action was part of the authorized intent.
* **The Happy Path (Tasks):**
    1. The Orchestrator receives a user request: "Clean up temporary logs from last week."
    2. The Orchestrator generates a signed Intent Token for "Log Cleanup" with scope `fs:delete:/tmp/*.log`.
    3. The Orchestrator passes this token to the Cleanup Subagent.
    4. The Cleanup Subagent calls the `delete_file` tool via MCP Any, providing the Intent Token.
    5. MCP Any's ISA Middleware verifies the signature and checks if `delete_file` on the target path matches the "Log Cleanup" intent.
    6. The tool call is allowed.

## 4. Design & Architecture
* **System Flow:**
    `Orchestrator -> [Signs Intent] -> Subagent -> [Tool Call + Intent Token] -> MCP Any (ISA Middleware) -> Policy Engine -> Tool Execution`
* **APIs / Interfaces:**
    * `POST /v1/intent/issue`: (Internal) Issue a new signed intent token.
    * `X-MCP-Intent-Token`: Header for passing the JWT-based intent token in tool calls.
* **Data Storage/State:**
    * Public keys for trusted orchestrators are stored in the MCP Any Key Store.
    * Intent Tokens are stateless (JWT) but may be checked against a short-lived revocation list.

## 5. Alternatives Considered
* **Static Capability Tokens**: Rejected because they don't solve the "Confused Deputy" problem; a token with `delete` permissions can be misused if the *intent* isn't verified.
* **HITL for every call**: Rejected due to high friction and latency in automated swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Uses asymmetric cryptography for token signing. The Policy Engine (Rego) ensures that even with a valid token, global safety constraints are never violated.
* **Observability**: All intent verifications (Success/Failure) are logged in the Audit Trail with the full intent context.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
