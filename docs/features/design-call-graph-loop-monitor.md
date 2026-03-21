# Design Doc: Call-Graph Loop Monitor
**Status:** Draft
**Created:** 2026-03-15

## 1. Context and Scope
Autonomous agents are prone to "Spiral of Death" loops where two or more tools re-trigger each other indefinitely. This not only wastes computational resources and API credits but can also lead to state corruption in shared databases. MCP Any, as the central gateway, must provide a circuit-breaking mechanism that detects these recursive cycles across different MCP servers and halts them before they cause catastrophic failure.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect recursive tool-calling patterns in real-time.
    * Enforce a configurable "Recursion Depth" limit (default: 10).
    * Provide a "Cycle Detection" algorithm (e.g., Tarjan's or similar) for tool-to-tool event flows.
    * Alert the user/orchestrator when a loop is broken.
* **Non-Goals:**
    * Automatically "fixing" the agent's logic (MCP Any only stops the execution).
    * Preventing intentional, shallow recursion that is part of a valid workflow.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Developer using complex event-driven tool chains.
* **Primary Goal:** Prevent a bug in an agent's logic from draining their OpenAI/Anthropic balance.
* **The Happy Path (Tasks):**
    1. Agent executes Tool A, which triggers a notification.
    2. Notification handler (another agent) executes Tool B.
    3. Tool B inadvertently re-triggers Tool A.
    4. MCP Any's Loop Monitor detects the repeated pattern and depth.
    5. On the 11th call (or detected cycle), MCP Any returns a `429 Too Many Requests (Recursive Loop Detected)` error.
    6. The orchestrator receives the error and pauses the swarm for human review.

## 4. Design & Architecture
* **System Flow:**
    - **Call-Graph Tracker**: A lightweight in-memory store that tracks the lineage of tool calls within a session.
    - **Middleware Integration**: Every tool request passes through the `LoopMonitorMiddleware`.
    - **Signature Matching**: Uses a combination of `AgentID`, `ToolName`, and `ArgumentHash` to identify repeated calls.
* **APIs / Interfaces:**
    - Config option: `security.loop_protection.max_depth: 10`
    - Config option: `security.loop_protection.action: "block" | "warn"`
* **Data Storage/State:**
    - Bloom filters or a sliding-window cache for high-frequency call tracking.

## 5. Alternatives Considered
* **Time-based Rate Limiting**: Rejected because loops can happen very quickly or very slowly, and rate limiting might block valid high-frequency non-recursive calls.
* **LLM-based Loop Detection**: Rejected as it is too expensive and slow to run on every tool call.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevents DoS attacks where a compromised subagent tries to exhaust host resources.
* **Observability:** Loops will be highlighted in the "Tool Execution Timeline" with a "Potential Cycle Detected" warning.

## 7. Evolutionary Changelog
* **2026-03-15:** Initial Document Creation.
