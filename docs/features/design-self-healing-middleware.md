# Design Doc: Self-Healing Tool Middleware

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
AI agents frequently fail when calling tools due to minor schema mismatches, malformed JSON, or missing optional but expected parameters. In a complex multi-agent swarm, these "shallow" failures often cascade, causing the entire task to fail. While some agent frameworks implement their own retry logic, this logic is often redundant and inconsistent. MCP Any, as the universal gateway, is perfectly positioned to provide a standardized, infrastructure-level "Self-Healing" loop that corrects these errors before they reach the parent agent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically diagnose and correct common tool call errors (e.g., type mismatches, missing non-required fields).
    *   Reduce the frequency of `invalid_arguments` errors returned to agents.
    *   Provide a "Diagnostic Trace" so users can see what was corrected.
    *   Support pluggable "Healing Strategies" (e.g., simple regex, LLM-based correction).
*   **Non-Goals:**
    *   Fixing logical errors in the agent's intent (e.g., if the agent calls the wrong tool).
    *   Executing tools that require explicit human approval (HITL).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Developer.
*   **Primary Goal:** Prevent agent failure when the model generates a tool call with a slight schema error.
*   **The Happy Path (Tasks):**
    1.  Agent calls `github_create_issue` but forgets the `body` field (which the MCP server requires but the agent thought was optional).
    2.  The MCP server returns a 400 Bad Request / Schema Validation Error.
    3.  **Self-Healing Middleware** intercepts the error.
    4.  The middleware uses a "Healing Strategy" (e.g., a small local model or rule-base) to identify that `body` is missing and can be defaulted to a placeholder or extracted from the `title`.
    5.  The middleware re-attempts the call with the corrected arguments.
    6.  The tool call succeeds; the agent receives the successful output, unaware of the internal correction (though a warning is logged).

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: Wraps the `ToolExecutionPipe`.
    - **Diagnosis**: On error, the `ErrorAnalyzer` matches the error against known "Healable Patterns."
    - **Correction**: The `Healer` applies a transformation to the original arguments.
    - **Verification**: The corrected call is validated against the tool's JSON schema before re-execution.
*   **APIs / Interfaces:**
    - `HealerStrategy` Interface: `Heal(originalArgs, error) -> (correctedArgs, error)`
*   **Data Storage/State:** Correction logs are stored in the `Audit Log` with a specific `self-healing` tag.

## 5. Alternatives Considered
*   **Agent-Side Retries**: Let every agent framework (OpenClaw, CrewAI) implement its own healing. *Rejected* because it's inefficient and leads to inconsistent behavior across the "Universal Bus."
*   **Hard Schema Enforcement**: Strictly failing every time. *Rejected* because modern LLMs are stochastic and "good enough" corrections significantly improve UX.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Self-Healing Middleware must NOT be allowed to add unauthorized permissions or change the "Intent Scope" of the call. It can only correct format/schema issues.
*   **Observability:** Visualizing "Healed Calls" in the UI (Activity Feed) so developers can improve their prompt engineering or tool definitions.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
