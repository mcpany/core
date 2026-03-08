# Design Doc: Intent-Aware Tool Scoping (Zero-Trust Agency)

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Current AI agent frameworks often grant full permissions to any tool once the agent is authorized. This "all or nothing" approach led to the February 2026 "Clinejection" exploits, where a prompt-injected subagent used powerful bash tools to perform unauthorized actions. MCP Any needs to restrict tool access not just by *capability* (can I run bash?) but by *intent* (am I running bash to fix a typo or to delete the database?).

## 2. Goals & Non-Goals
*   **Goals:**
    *   Limit tool access based on a cryptographically signed "Intent Manifest."
    *   Provide a "Verification Layer" where parent agents can sign-off on specific subagent task scopes.
    *   Ensure that tool arguments (e.g., file paths) are validated against the current intent.
    *   Support intent-scoped A2A handoffs.
*   **Non-Goals:**
    *   Automating the generation of intent manifests (these should be provided by the agent architect or orchestrator).
    *   Replacing traditional RBAC/ABAC permissions.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Orchestrator (e.g., OpenClaw Specialist).
*   **Primary Goal:** Delegate a "File Read" task to a subagent without allowing it to "Write" or "Delete" anything.
*   **The Happy Path (Tasks):**
    1.  Parent agent generates an intent: `READ_ONLY_DOCS` with a scope of `/docs/*.md`.
    2.  Parent signs this intent using its session key.
    3.  Subagent receives the signed intent and tries to call `write_file(path="/docs/README.md")`.
    4.  MCP Any Policy Firewall checks the intent manifest, sees it is `READ_ONLY`, and blocks the tool call.
    5.  The subagent's `read_file` call is permitted because it matches the intent scope.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Manifest**: A JSON Web Token (JWT) or similar signed structure that defines allowed actions and resource patterns.
    - **Policy Engine Hook**: The existing `Policy Firewall` is extended to intercept the intent token in the request headers.
    - **Dynamic Scoping**: MCP Any temporarily "prunes" the available tool list based on the active intent manifest.
*   **APIs / Interfaces:**
    - Header: `X-MCP-Intent-Manifest: [signed-token]`
    - Manifest Schema:
      ```json
      {
        "intent_id": "UUID",
        "scope": ["fs:read:/docs/*"],
        "max_tool_calls": 5,
        "exp": "TIMESTAMP"
      }
      ```
*   **Data Storage/State:** Intent state is transient and session-bound, verified by the parent agent's public key stored in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Manual Confirmation (HITL)**: Asking the user for every tool call. *Rejected* as it doesn't scale for complex swarms.
*   **Static Resource Whitelisting**: Hardcoding folder access in the config. *Rejected* because intent changes dynamically based on the task.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature implements the "Principle of Least Privilege" at the task level, not just the user level.
*   **Observability:** The UI timeline will show the active intent for every tool call and flag "Out-of-Scope" attempts as security alerts.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
