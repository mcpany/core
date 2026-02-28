# Design Doc: Intent Validation Engine (IVE)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents move from simple tool execution to exploratory orchestration, static API schemas are no longer enough to ensure security. A 2026 audit by Equixly found that 43% of MCP servers are vulnerable to injection attacks. The problem is that agents can use a valid schema to perform an invalid *intent* (e.g., using a file-write tool to overwrite a system config). The Intent Validation Engine (IVE) provides a middle layer that verifies whether a tool call aligns with the user's high-level intent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept every tool call and evaluate it against a "User Intent Policy."
    *   Support both deterministic (Rego/CEL) and heuristic (LLM-based) intent verification.
    *   Provide a way for users to define "Guards" for sensitive operations.
    *   Reduce "Confused Deputy" attacks by ensuring the agent isn't abusing its broad tool access.
*   **Non-Goals:**
    *   Replacing traditional RBAC (this is an *additional* layer).
    *   Defining the user's intent automatically (it must be provided as part of the session context).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer.
*   **Primary Goal:** Prevent an agent from performing a "DELETE" operation that wasn't part of the initial task description.
*   **The Happy Path (Tasks):**
    1.  User starts a session with the intent: "Clean up temporary build files in /tmp/build."
    2.  Agent calls `delete_file(path="/tmp/build/artifact.o")`.
    3.  IVE checks the call against the intent and the session-bound policy.
    4.  IVE approves the call as it aligns with the "Clean up temporary build files" intent.
    5.  Agent later attempts `delete_file(path="/etc/passwd")` (via injection).
    6.  IVE flags this as "Intent Mismatch" and blocks the call.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Capture**: User intent is stored in the `Recursive Context Protocol` headers.
    - **Interception**: Every `tools/call` is routed through the IVE Middleware.
    - **Verification**: IVE uses a plugin system (Rego for static rules, a small LLM for semantic checks) to validate the call.
    - **Decision**: Approve, Block, or Flag for HITL (Human-in-the-Loop) approval.
*   **APIs / Interfaces:**
    - `POST /v1/intent/verify`: Endpoint for the middleware to check a proposed call.
    - `_mcp_intent_token`: A header that cryptographically binds the current tool call to a verified intent.
*   **Data Storage/State:** Intent policies are stored in the Service Registry as a new policy type.

## 5. Alternatives Considered
*   **Schema Hardening**: Making the tool schemas extremely restrictive. *Rejected* as it breaks the flexibility required for autonomous agents.
*   **Manual HITL for Every Call**: Asking the user to approve every call. *Rejected* because it eliminates the autonomy that makes agents useful.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** IVE is a key component of "Intent-Aware" security.
*   **Observability:** The UI should show why a call was approved or blocked (e.g., "Matched Policy: Build-Cleanup-Guard").

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
