# Design Doc: Context Quota Enforcement
**Status:** Draft
**Created:** 2026-04-15

## 1. Context and Scope
In deep agent swarms, a single runaway subagent can consume the entire available context window or memory budget of the parent or the gateway, leading to "Context Storms" and resource exhaustion. OpenClaw v2026.4.0-alpha has introduced the concept of "Context Quotas" to mitigate this. MCP Any must implement similar enforcement to ensure fair resource distribution across all agents in a session and protect the stability of the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hard and soft quotas for token usage and state memory per subagent.
    * Provide a mechanism for parent agents to dynamically allocate and reclaim quotas from their children.
    * Halt tool execution or context mounting when a subagent exceeds its assigned quota.
    * Report quota consumption metrics in real-time.
* **Non-Goals:**
    * Automatically optimizing agent prompts to fit within quotas.
    * Managing LLM provider-level rate limits (this is handled by the transport/provider layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Research Subagent" from bloating the shared session state with 50MB of raw HTML.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a session and spawns a "Research Subagent" with a 1MB memory quota.
    2. Research Subagent attempts to store a large dataset in the Blackboard.
    3. MCP Any checks the subagent's current consumption against its quota.
    4. The quota is exceeded; MCP Any rejects the write operation with a `QuotaExceeded` error.
    5. Parent agent receives the error and decides whether to grant more quota or prune the subagent's task.

## 4. Design & Architecture
* **System Flow:**
    * Tool Call/State Write -> Quota Middleware -> Enforcement Engine -> Action.
* **APIs / Interfaces:**
    * `SetQuota(agent_id, limits)`: Parent-only API to define subagent boundaries.
    * `GetConsumption(agent_id)`: Retrieve current resource usage.
* **Data Storage/State:**
    * Quota definitions and current counters stored in the session-scoped Blackboard.

## 5. Alternatives Considered
* **Implicit Limits:** Rejected as they lead to non-deterministic failures that are hard for LLMs to reason about.
* **Centralized Orchestration:** Rejected to maintain framework-neutrality (MCP Any should enforce, not necessarily orchestrate).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Quotas prevent resource-exhaustion DoS attacks within a swarm.
* **Observability:** New dashboard in the UI for "Context Quota Monitoring."

## 7. Evolutionary Changelog
* **2026-04-15:** Initial Document Creation.
