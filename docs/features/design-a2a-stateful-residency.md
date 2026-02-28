# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
As multi-agent swarms become more complex (e.g., OpenClaw's nested orchestration), the communication between agents is often intermittent and asynchronous. Standard MCP is stateless, making it difficult to maintain reliable "Agent-to-Agent" (A2A) handoffs when one agent is processing or offline. MCP Any must provide a "Stateful Residency" layer that acts as a persistent mailbox and message buffer for all inter-agent communications.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a persistent "Mailbox" for A2A messages that survives agent restarts.
    *   Implement "Message Queuing" for agents with intermittent connectivity.
    *   Store the lineage and context of multi-agent tasks in a centralized "State Residency" layer.
    *   Expose this state via standard MCP tools for agents to "Poll" or "Listen" for tasks.
*   **Non-Goals:**
    *   Becoming a general-purpose message broker (it is specific to A2A protocols).
    *   Implementing agent-side message handling logic.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Agent Orchestrator.
*   **Primary Goal:** Ensure a "Researcher" agent can hand off a task to a "Coder" agent even if the Coder agent is currently busy or re-starting.
*   **The Happy Path (Tasks):**
    1.  Researcher calls the `a2a_send_message` tool in MCP Any.
    2.  MCP Any stores the message in the "Stateful Residency" (SQLite-backed Shared KV Store).
    3.  Coder agent connects (or polls) and calls `a2a_get_mailbox`.
    4.  MCP Any delivers the message and updates its status to "Delivered."
    5.  Full trace of the handoff is visible in the UI Timeline.

## 4. Design & Architecture
*   **System Flow:**
    - **Persistence Layer**: Built on top of the `Shared KV Store` (SQLite).
    - **Middleware**: `A2AResidencyMiddleware` handles the translation between A2A protocol calls and the mailbox storage.
    - **Notification Hub**: Optional SSE/WebSocket stream for real-time message delivery to active agents.
*   **APIs / Interfaces:**
    - `POST /v1/a2a/mailbox` - Send message.
    - `GET /v1/a2a/mailbox` - Poll messages.
    - `POST /v1/a2a/ack` - Acknowledge receipt.
*   **Data Storage/State:** Structured JSON records in the `a2a_mailbox` table within the system database.

## 5. Alternatives Considered
*   **In-Memory Buffering**: *Rejected* because it does not survive restarts, which is critical for long-running swarms.
*   **External Redis/NATS**: *Rejected* for the base version to keep MCP Any as a single, low-dependency binary.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Access to mailboxes is restricted by "Intent-Bound" capability tokens. Agent A cannot read Agent B's mailbox without explicit policy authorization.
*   **Observability:** The `A2A Stateful Mailbox` UI component provides real-time visibility into queued and delivered messages.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
