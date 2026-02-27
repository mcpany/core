# Design Doc: Intent-Alignment Middleware

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As autonomous agents (like OpenClaw) gain stars and usage, they are becoming targets for sophisticated prompt injection attacks. Standard capability-based security ("Can this agent run this tool?") is insufficient because a hijacked agent can use its valid capabilities to perform malicious actions (e.g., using `fs_write` to overwrite `.ssh/authorized_keys` instead of a project file). MCP Any needs a layer that verifies if a tool call aligns with the *original user intent* of the session.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a real-time "Guardian" check for every tool call.
    *   Maintain a "Session Intent" state that captures the high-level goal (e.g., "Fix bug in the login component").
    *   Use lightweight, fast-inference models to score tool-call alignment.
    *   Provide a "Safety Score" for tool calls that can trigger HITL (Human-in-the-Loop) flows if low.
*   **Non-Goals:**
    *   Replacing the primary LLM's reasoning.
    *   Adding significant latency (>100ms) to tool calls.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer using OpenClaw via MCP Any.
*   **Primary Goal:** Prevent an agent from performing a "Redirected Action" (e.g., exfiltrating keys) during a legitimate coding task.
*   **The Happy Path (Tasks):**
    1.  User starts a session with the intent: "Refactor the database helper."
    2.  Agent performs several `fs_read` and `fs_write` calls related to `db_helper.py`.
    3.  Intent-Alignment Middleware verifies these against the intent; alignment is high (>0.9).
    4.  An attacker-controlled webpage (ingested by the agent) injects a command to `fs_read ~/.aws/credentials`.
    5.  Middleware detects the misalignment with "Refactor database helper"; alignment is low (<0.2).
    6.  MCP Any blocks the call and alerts the user.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Capture**: Middleware extracts the initial prompt or "System Goal" and stores it in the session context.
    - **Scoring**: For every `tools/call`, the middleware sends a compact representation (Intent + Tool Name + Params) to a "Guardian" model.
    - **Action**:
        - Score > 0.8: Allow.
        - 0.5 < Score < 0.8: Trigger HITL.
        - Score < 0.5: Block and Log.
*   **APIs / Interfaces:**
    - `intent_alignment/verify`: Internal API for scoring.
    - `sessions/intent/update`: API to refine the goal as the conversation evolves.
*   **Data Storage/State:** Intents are stored in the `Shared KV Store`, bound to the `Session ID`.

## 5. Alternatives Considered
*   **Keyword Filtering**: Simple but easily bypassed by clever prompting.
*   **Heavyweight Reasoning**: Using the main LLM to verify itself. *Rejected* due to cost, latency, and the "Self-Deception" problem (if the main model is hijacked, its self-verification might also be compromised).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The "Guardian" model should ideally be a different architecture or provider than the main model to prevent cross-model prompt injection patterns.
*   **Observability:** Log alignment scores and reasoning for every tool call to the audit trail.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
