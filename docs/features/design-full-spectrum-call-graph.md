# Design Doc: Full-Spectrum Call Graph Monitor
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
With the rise of "Shadow Reasoning" (CVE-2026-48002), agents are increasingly executing internal logic branches that bypass traditional gateway monitoring. These un-reported sub-intents can lead to unauthorized tool calls or state mutations that the parent agent is unaware of. The Full-Spectrum Call Graph Monitor provides deep instrumentation into the agent's internal monologue and reasoning steps, ensuring every "thought" that leads to an "action" is accounted for and validated.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and index internal agent reasoning (monologues/thoughts) in real-time.
    * Detect "Shadow Branches" that diverge from the verified session intent.
    * Provide a unified call graph that links "Intent" -> "Reasoning" -> "Action" (Tool Call).
    * Halt tool execution if a "Shadow Reasoning" pattern is detected.
* **Non-Goals:**
    * Modifying the agent's internal model weights.
    * Blocking all divergent reasoning (some exploration is allowed, but it must be reported).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor / Swarm Orchestrator
* **Primary Goal:** Detect if a subagent is "plotting" to bypass a security boundary before it actually calls the tool.
* **The Happy Path (Tasks):**
    1. Subagent begins a reasoning loop.
    2. MCP Any's monitor streams the internal monologue.
    3. The monitor detects a "Shadow Reasoning" pattern (e.g., an internal plan to exfiltrate data via a side channel).
    4. The monitor flags the session and suspends the next tool call.
    5. The parent agent or user is alerted to the "Intent Divergence."

## 4. Design & Architecture
* **System Flow:**
    `Internal Monologue Stream` -> `Monologue Parser` -> `Pattern Matcher` -> `Intent Alignment Engine` -> `Call Graph Indexer`
* **APIs / Interfaces:**
    * `MonologueHook`: Interface for agents to push their reasoning state.
    * `ShadowDetectionAPI`: Endpoint for querying intent alignment scores.
* **Data Storage/State:**
    * Ephemeral storage for reasoning traces, linked to the Session ID in the Blackboard.

## 5. Alternatives Considered
* **Post-hoc Log Analysis**: Rejected as it is too slow to prevent the initial malicious tool call.
* **Restrictive Tool Scoping**: Rejected because it doesn't prevent "logical" exfiltration or state poisoning if the agent still has some tool access.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Reasoning traces are cryptographically linked to the parent's signed intent.
* **Observability:** Integrates with the "Agent Chain Tracer (A2A)" and the "Recursive Loop Heatmap."

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
