# Design Doc: Asynchronous Tool Execution (Detached Workers)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents evolve from simple chatbots to complex task orchestrators (e.g., Claude Code, OpenClaw), the need for long-running background tasks has become critical. Currently, the MCP protocol is largely synchronous: the model calls a tool and waits for a response. This blocks the agent's main loop and prevents parallel execution.

MCP Any needs to bridge this gap by providing an asynchronous execution layer that allows tools to be "detached" from the main request-response cycle, managed by a background worker pool, and tracked via standardized status interfaces.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable agents to initiate tool calls that return immediately with a `job_id`.
    * Provide a standardized "Blackboard" tool for agents to poll the status of background jobs.
    * Support callback webhooks for real-time notification of job completion.
    * Ensure state persistence for background jobs across server restarts.
* **Non-Goals:**
    * Implementing a full-blown distributed job queue like Celery or Sidekiq (keep it lightweight and embedded).
    * Modifying the core MCP JSON-RPC spec (layer this on top as a middleware/wrapper).

## 3. Critical User Journey (CUJ)
* **User Persona:** Background Agent Orchestrator (e.g., Claude Code)
* **Primary Goal:** Start an expensive indexing task in the background while continuing to answer user queries.
* **The Happy Path (Tasks):**
    1. Agent sends a tool call to `index_repository` with an `async: true` flag in the metadata.
    2. MCP Any intercepts the call, spawns a background worker, and immediately returns `{"status": "pending", "job_id": "job_123"}`.
    3. Agent continues interacting with the user.
    4. Agent periodically calls `get_job_status(job_id="job_123")` to check progress.
    5. Once complete, the agent retrieves the final result and integrates it into the context.

## 4. Design & Architecture
* **System Flow:**
    1. **Interceptor Middleware**: Detects the `async` intent in the tool call.
    2. **Job Manager**: Generates a UUID, persists the initial state in the embedded SQLite store.
    3. **Worker Pool**: A bounded pool of goroutines/workers that execute the actual upstream MCP call.
    4. **Result Store**: Captures the final output or error and updates the job status.
* **APIs / Interfaces:**
    * `mcp_any_job_status(job_id)`: A built-in MCP tool for status polling.
    * `mcp_any_job_cancel(job_id)`: A built-in MCP tool for job termination.
* **Data Storage/State:**
    * Uses the existing `Shared KV Store` (SQLite) to track `job_id`, `status`, `payload`, and `result`.

## 5. Alternatives Considered
* **Client-side Async**: Forcing the agent framework (e.g., LangGraph) to handle the backgrounding.
    * *Rejected*: This leads to fragmented implementations and loses the "Universal Adapter" benefit of MCP Any.
* **Long-Polling**: Keeping the HTTP/Stdio connection open for the duration.
    * *Rejected*: Prone to timeouts and prevents the agent from doing other work in the meantime.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Background workers must inherit the exact same capability tokens as the parent request.
    * Job results must be restricted so only the initiating agent/session can retrieve them.
* **Observability:**
    * Integrate with the "Tool Activity Feed" in the UI to show active background tasks and their resource consumption.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
