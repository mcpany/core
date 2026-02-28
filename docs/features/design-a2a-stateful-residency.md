# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
As the Agent-to-Agent (A2A) protocol matures, a critical bottleneck has emerged: reliability. Many agents are intermittent, running in local terminal sessions or on-demand cloud functions that may disconnect before a delegated task is complete. Without a persistent intermediary, A2A messages and "waiting" state are lost during network drops or process restarts. MCP Any must provide "Stateful Residency" to act as a reliable, asynchronous buffer for the agentic mesh.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a persistent "Mailbox" for A2A messages that survives agent disconnections.
    *   Maintain the "Agent Handoff" state across session restarts.
    *   Implement an "Acknowledge and Retain" delivery guarantee for A2A messages.
    *   Expose a "Buffer Status" API to monitor queued tasks and pending callbacks.
*   **Non-Goals:**
    *   Implementing agent reasoning logic.
    *   Providing long-term archival of agent conversations (focus is on active session buffers).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Agent Orchestrator.
*   **Primary Goal:** Ensure a task delegation from a Parent Agent to a specialized Subagent is not lost if the Subagent's process restarts.
*   **The Happy Path (Tasks):**
    1.  Parent Agent sends an A2A task delegation message to MCP Any.
    2.  MCP Any persists the message in its `A2A Mailbox` (SQLite-backed) and returns a `queued` status.
    3.  Subagent connects to MCP Any and "pulls" the pending task.
    4.  Subagent crashes midway through execution.
    5.  On restart, Subagent reconnects and resumes the task based on the state still resident in MCP Any's buffer.

## 4. Design & Architecture
*   **System Flow:**
    - **Buffer Ingestion**: The `A2A Gateway` receives messages and immediately commits them to the `Shared KV Store` (Stateful Residency table).
    - **Delivery Logic**: Uses a "Lease" mechanism where a pulling agent must heart-beat or acknowledge the message, or it returns to the "Pending" pool.
    - **State Synchronization**: Periodically flushes buffered state to the SQLite Blackboard to ensure durability.
*   **APIs / Interfaces:**
    - `POST /a2a/v1/messages/enqueue`: Adds a message to the buffer.
    - `GET /a2a/v1/messages/poll`: Retrieves pending messages for a specific agent identity.
    - `POST /a2a/v1/messages/ack`: Marks a message as successfully processed.
*   **Data Storage/State:** Uses the existing `Shared KV Store` (Blackboard) with a specialized `a2a_mailbox` schema.

## 5. Alternatives Considered
*   **Ephemeral Memory Buffer**: Keeping messages only in RAM. *Rejected* as it doesn't solve the "process restart" problem.
*   **External Message Queue (RabbitMQ/Redis)**: Requiring a separate infrastructure component. *Rejected* to maintain MCP Any's "Indispensable Core" identity as a standalone, zero-dependency binary.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Messages are encrypted at rest. Access to the mailbox is strictly governed by the `A2A Interop Bridge` identity tokens.
*   **Observability:** The UI provides a "Stateful A2A Mailbox" view, showing real-time queue depths and delivery latencies.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
