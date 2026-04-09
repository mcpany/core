# Design Doc: Asynchronous Execution Queuing (AEQ)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents move from linear, single-threaded interactions to complex, high-frequency missions, the latency of synchronous tool confirmations (waiting for user or system approval) becomes a primary bottleneck. Agents often stall while waiting for a single tool call to resolve, even when they could be performing other non-dependent tasks. The Asynchronous Execution Queuing (AEQ) service provides an event-driven scheduler that allows agents to queue tool calls and confirmations, enabling speculative reasoning and parallel task processing.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement an event-driven scheduler for tool execution and user confirmations.
    * Allow agents to speculatively proceed with reasoning while tool results are pending.
    * Provide a unified state manager for tracking the lifecycle of queued tool calls (Queued, Pending, Confirmed, Executed, Failed).
    * Support "Priority-Aware" execution, allowing safety-critical tools to bypass the queue.
* **Non-Goals:**
    * Replacing the main reasoning engine (AEQ is a support service for the engine).
    * Automating user approvals (AEQ only manages the *state* of the approval request).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an AI agent for large-scale codebase refactoring.
* **Primary Goal:** Execute 10+ filesystem edits across different modules without waiting for each one to be confirmed individually.
* **The Happy Path (Tasks):**
    1. Agent identifies 10 files that need modification.
    2. Agent queues 10 `file_write` calls via AEQ.
    3. MCP Any surfaces all 10 calls in the `Queued Confirmations` UI.
    4. Agent continues analyzing the codebase for potential architectural improvements instead of stalling.
    5. User reviews the batch of 10 edits and clicks "Approve All."
    6. AEQ triggers the execution of all 10 calls in the background and notifies the agent of the results.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `AEQ` -> `Queue` -> `UI/HITL` -> `Execution Engine` -> `AEQ Result Handler` -> `Agent`
* **APIs / Interfaces:**
    * `ExecutionQueue`: `Enqueue(call *ToolCall) (callID string)`, `GetStatus(callID) Status`
    * `ResultStream`: `Subscribe(agentID) chan Result`
* **Data Storage/State:**
    * Persistent task queue stored in the Durable Mission Continuity Provider to ensure recovery after reboots.

## 5. Alternatives Considered
* **Parallel WebSocket Streams**: Rejected because managing 10+ concurrent streams for individual calls is resource-intensive and complex for agents to coordinate.
* **Synchronous Batching**: Rejected because it requires the agent to wait until it has found *all* tasks before sending them, losing the benefits of incremental progress.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Every queued call is cryptographically bound to the hardware-attested agent session. Speculative reasoning is contained within isolated context shards.
* **Observability**: The `Execution Queue Monitor` provides a visual timeline of queued, pending, and completed tool calls.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
