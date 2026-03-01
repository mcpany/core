# Design Doc: Agentic Loop Prevention (ALOD)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Multi-agent swarms often enter "infinite loops" where two or more agents repeatedly exchange the same or similar messages/state without reaching a conclusion. This is not only a waste of tokens and compute but also blocks the user's ultimate goal. MCP Any, as the central gateway for these communications, is uniquely positioned to detect and terminate these non-productive cycles.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Detect repetitive agent state transitions and tool-call patterns.
    *   Provide a configurable "Circuit Breaker" to terminate suspected loops.
    *   Expose loop metrics (cycle frequency, state similarity) to the UI.
    *   Enable "Self-Healing" by suggesting a state rollback or a human-in-the-loop intervention.
*   **Non-Goals:**
    *   Replacing the agent's reasoning logic.
    *   Defining what a "good" loop is (e.g., valid iterative refinement). This must be configurable via policy.

## 3. Critical User Journey (CUJ)
*   **User Persona:** LLM Swarm Orchestrator.
*   **Primary Goal:** Prevent a coding swarm from spending $50 on a loop between a "Linter" and a "Fixer" agent.
*   **The Happy Path (Tasks):**
    1.  The swarm enters a loop where Agent A fixes a bug, and Agent B's linting rules revert it.
    2.  ALOD middleware observes that the session state has returned to a previously seen hash 3 times within 10 turns.
    3.  ALOD triggers the Circuit Breaker, pausing the session.
    4.  The system notifies the user via the HITL (Human-in-the-Loop) interface, showing the detected cycle.
    5.  The user intervenes or the system automatically rolls back to a "Last Known Good" state with a new instruction to break the cycle.

## 4. Design & Architecture
*   **System Flow:**
    - **State Hashing**: Every A2A message and tool-call payload is hashed and stored in a sliding window per-session.
    - **Frequency Analysis**: ALOD calculates the frequency of identical or highly similar state hashes.
    - **Similarity Detection**: Uses MinHash or similar algorithms to detect "fuzzy" loops (where state changes slightly but intent remains identical).
*   **APIs / Interfaces:**
    - **Internal**: `LoopDetectorHook` interface for the A2A Bridge and Policy Engine.
    - **External**: API endpoints for retrieving loop telemetry and overriding pauses.
*   **Data Storage/State:** Uses an in-memory Redis or SQLite sliding window for high-performance hash lookups.

## 5. Alternatives Considered
*   **LLM-based Loop Detection**: Using a separate LLM call to analyze the history for loops. *Rejected* due to high latency and cost. Infrastructure-level hashing is faster and cheaper.
*   **Agent-side Detection**: Relying on agents to detect their own loops. *Rejected* because "stuck" agents are notoriously bad at self-reflection.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Loop detection must not expose sensitive data. Hashing should be one-way and salt-based per session.
*   **Observability:** The UI must visualize the "Loop Waterfall" to help developers debug their swarms.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
