# Design Doc: Cascading Failure Circuit Breaker
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As AI agent ecosystems transition from single-agent tasks to multi-agent swarms (e.g., OpenClaw, CrewAI), the risk of "cascading failures" increases. A single malfunctioning agent or a poisoned tool can trigger a chain reaction of errors, infinite loops, or resource exhaustion across the entire swarm. Current agent frameworks lack a centralized mechanism to detect and halt these patterns. MCP Any, acting as the universal gateway, is uniquely positioned to implement a "Circuit Breaker" that monitors inter-agent communication and tool execution for signs of systemic failure.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect "Infinite Tool Loops" (e.g., Agent A calls Tool 1 -> Tool 1 fails -> Agent A retries indefinitely).
    * Halt "Error Propagation" (e.g., Agent A passes a corrupted state to Agent B, causing Agent B to fail).
    * Implement "Threshold-Based Suspension" (e.g., if 5 tool calls fail within 10 seconds, suspend the agent session).
    * Provide a "Black Box" log for post-mortem analysis of the failure cascade.
* **Non-Goals:**
    * Automatically "fixing" the logic of the agents.
    * Replacing local retry logic within agent frameworks (MCP Any provides a global safety net).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator managing a fleet of 10+ specialized agents.
* **Primary Goal:** Prevent a single misconfigured "Researcher Agent" from exhausting the entire project's token budget through infinite retries.
* **The Happy Path (Tasks):**
    1. The "Researcher Agent" encounters a 404 error from a search tool.
    2. Instead of pivoting, the agent retries the same search call 5 times in rapid succession.
    3. MCP Any's Circuit Breaker detects the "High-Frequency Identical Call" pattern.
    4. MCP Any trips the breaker for that specific agent session and notifies the Orchestrator.
    5. The rest of the swarm remains active while the Researcher Agent is quarantined for review.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any (Circuit Breaker Middleware)` -> `Tool/Other Agent`
    1. **Observation**: The middleware tracks every tool call and A2A message in a circular buffer per session.
    2. **Analysis**: The `Pattern Engine` evaluates the buffer against rules (e.g., "Max Retries", "Error Density", "State Drift").
    3. **Action**:
        - **Closed**: Normal operation.
        - **Open**: Breaker tripped; all subsequent calls for that session return a `CIRCUIT_BREAKER_OPEN` error.
        - **Half-Open**: Periodic trial calls to check if the underlying issue is resolved.
* **APIs / Interfaces:**
    * `GET /v1/health/circuit-breakers`: Returns status of all active breakers.
    * `POST /v1/sessions/{id}/reset`: Manually reset a tripped breaker.
* **Data Storage/State:**
    * In-memory Redis or SQLite store for tracking call frequencies and error states.

## 5. Alternatives Considered
* **Agent-Side Circuit Breakers**: Rejected because they cannot be enforced across different frameworks (Claude Code vs. OpenClaw).
* **Network-Level Rate Limiting**: Too blunt; doesn't understand the "intent" or "identity" of the agent causing the failure.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The circuit breaker can also detect "Prompt Injection Cascades" where a malicious prompt causes an agent to start attacking other agents in the swarm.
* **Observability**: Integration with the "Agentic Black Box Recorder" to provide a visual timeline of why the breaker was tripped.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
