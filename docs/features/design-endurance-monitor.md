# Design Doc: Autonomous Endurance Monitor

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As agents evolve toward "Agentic Endurance" (Prosus 2026), they are performing autonomous tasks that span several hours and hundreds of individual steps. In these long-horizon scenarios, small compounding errors can lead to catastrophic failure or "infinite loops." MCP Any, as the orchestration layer, needs a way to monitor the health of these extended sessions, provide "heartbeat" verification, and enable state checkpoints to prevent total task loss.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Track session duration and tool-call frequency for long-running autonomous tasks.
    *   Detect "compounding error" patterns (e.g., repeated tool failures with slight parameter variations).
    *   Implement "Checkpoint Tools" that allow agents to save intermediate state to the `Shared KV Store`.
    *   Provide a "Session Health" score to the UI and parent agents.
*   **Non-Goals:**
    *   Replacing the LLM's reasoning loop.
    *   Automatically fixing complex logic errors (the goal is detection and suspension).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Ops Engineer.
*   **Primary Goal:** Ensure a long-running Data Migration Agent (expected duration: 4 hours) doesn't enter a "silent failure" state or exceed its resource quota.
*   **The Happy Path (Tasks):**
    1.  Engineer starts the agent with an `Endurance-Scope` token.
    2.  MCP Any begins tracking the session's "Endurance Metrics" (latency, error rate, entropy).
    3.  Every 30 minutes, the agent calls `checkpoint_save(state_blob)` via MCP Any.
    4.  At hour 3, MCP Any detects a rising error rate in database writes and "Suspends" the session.
    5.  Engineer reviews the "Session Health" log, adjusts the DB permissions, and "Resumes" from the last checkpoint.

## 4. Design & Architecture
*   **System Flow:**
    - **Monitoring Middleware**: Hooks into every `tools/call` to record telemetry and update the `SessionContext`.
    - **Heuristic Engine**: Analyzes the recent history of calls for "Looping" or "Drift" (e.g., repeating the same failed command 5 times).
    - **Checkpoint API**: A set of built-in MCP tools for state persistence.
*   **APIs / Interfaces:**
    - `endurance/status`: Returns current health metrics for a session.
    - `endurance/suspend`: Pause a session for manual intervention.
*   **Data Storage/State:** Checkpoints and telemetry are stored in the local SQLite `Blackboard`.

## 5. Alternatives Considered
*   **Stateless Retries**: Simply retrying failed calls. *Rejected* because it doesn't solve for long-term logic drift or resource exhaustion.
*   **Client-Side Monitoring**: Forcing the agent framework to monitor itself. *Rejected* because MCP Any provides a framework-agnostic, unified observability point.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Endurance tokens must be session-bound and cannot be used to bypass the `Policy Firewall`.
*   **Observability:** The UI must provide a "Real-Time Endurance Dashboard" showing the health of all active long-horizon tasks.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
