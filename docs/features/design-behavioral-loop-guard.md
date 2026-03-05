# Design Doc: Agent Behavioral Loop Guard

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As AI agents become more autonomous and their toolsets more complex, they are susceptible to "Agent Loop Injection" (ALI) attacks or accidental infinite loops (e.g., Agent A calls Tool X, which triggers Agent B to call Tool Y, which calls Tool X again). These loops lead to rapid token consumption ("Token Bankruptcy"), high costs, and potential denial of service for the underlying infrastructure. MCP Any needs a proactive middleware to detect and mitigate these behavioral anomalies.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Detect circular tool call patterns within a single session or across a multi-agent swarm.
    *   Implement configurable "Circuit Breakers" that trip when a loop is detected.
    *   Provide telemetry and alerts when behavioral anomalies occur.
    *   Support "Frequency-Based" rate limiting for tool calls (e.g., "no more than 5 calls to the same tool in 30 seconds").
*   **Non-Goals:**
    *   Determining the "intent" behind a loop (the guard is purely behavioral).
    *   Modifying the logic of the underlying tools.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Platform Security Engineer.
*   **Primary Goal:** Prevent a misconfigured OpenClaw swarm from spending $500 in 5 minutes due to an execution loop between two subagents.
*   **The Happy Path (Tasks):**
    1.  Engineer enables the `LoopGuardMiddleware` in the MCP Any configuration.
    2.  An agent enters a loop: `list_files` -> `read_file` -> `list_files` -> `read_file`...
    3.  The `LoopGuardMiddleware` tracks the call stack and detects the repeating pattern.
    4.  On the 5th iteration, the middleware blocks the call and returns a `LoopDetectedError` to the agent.
    5.  The incident is logged, and the user is notified via the UI Dashboard.

## 4. Design & Architecture
*   **System Flow:**
    - **Call Tracking**: Every tool call is intercepted by the middleware.
    - **Pattern Matching**: The middleware maintains a sliding window of recent tool calls (Tool ID + Hash of Arguments).
    - **Circuit Breaker**: If a pattern repeats beyond the `max_repeat_threshold`, the circuit breaks.
*   **APIs / Interfaces:**
    - **Middleware Hook**: Standard MCP Any middleware interface.
    - **Error Protocol**: New JSON-RPC error code for `LOOP_DETECTED`.
*   **Data Storage/State:**
    - In-memory LRU cache for active sessions.
    - Persistent "Blacklist" in the Shared KV Store for repeatedly offending agent IDs.

## 5. Alternatives Considered
*   **Client-Side Loop Detection**: Relying on the agent framework (e.g., LangChain) to detect loops. *Rejected* because MCP Any must be a "Zero Trust" gateway that protects the infrastructure regardless of the client's implementation.
*   **Static Analysis of Tool Dependencies**: Hardcoding which tools can call which other tools. *Rejected* as too rigid for dynamic agent swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Loop Guard is a core component of the "Active Behavioral Monitoring" strategy.
*   **Observability:** Expose loop statistics (Frequency, Repeats, Tripped Circuits) in the UI.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
