# Design Doc: A2A Stateful Residency (Agentic Mailbox)

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
In a multi-agent swarm (e.g., OpenClaw, CrewAI), specialized subagents often operate asynchronously or in transient environments. When a parent agent delegates a task via the A2A protocol, the recipient subagent might be temporarily offline or occupied, leading to message loss or timeout. MCP Any needs a "Stateful Residency" layer to act as a reliable, persistent mailbox for agentic communications.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a persistent, disk-backed "mailbox" for A2A messages.
    *   Support "Store-and-Forward" delivery where messages are held until the recipient agent heartbeats or connects.
    *   Implement "Dead Letter Queues" for undeliverable agent tasks.
    *   Expose message status (Pending, Delivered, Acknowledged) to the parent agent via MCP tool outputs.
*   **Non-Goals:**
    *   Implementing a general-purpose message broker (e.g., RabbitMQ). This is specific to the A2A protocol.
    *   Executing agent logic during the "waiting" phase.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent System Architect.
*   **Primary Goal:** Ensure a task delegated to a specialized subagent is eventually delivered even if the subagent is restarting.
*   **The Happy Path (Tasks):**
    1.  Agent A calls `mcp_a2a_send(recipient="AgentB", task="Generate Report")`.
    2.  MCP Any determines Agent B is currently offline.
    3.  MCP Any stores the message in the `A2A_Mailbox` and returns a `tracking_id`.
    4.  Agent B comes online and requests pending messages from MCP Any.
    5.  Agent B processes the task and posts a result back to the mailbox.
    6.  Agent A (or a watcher) retrieves the result using the `tracking_id`.

## 4. Design & Architecture
*   **System Flow:**
    - **Ingress**: A2A Bridge receives a `message/post` and validates the recipient's identity.
    - **Persistence**: Messages are serialized and stored in a local SQLite table (`a2a_messages`).
    - **Egress**: Recipients poll or subscribe (via SSE/WS) to their identity-bound mailbox.
*   **APIs / Interfaces:**
    - `POST /a2a/mailbox/send`: Ingest a message.
    - `GET /a2a/mailbox/receive`: Retrieve pending messages for the caller's identity.
    - `POST /a2a/mailbox/ack`: Acknowledge receipt and/or post a response.
*   **Data Storage/State:**
    - Table: `a2a_messages` (id, sender_id, recipient_id, payload, status, created_at, expires_at).

## 5. Alternatives Considered
*   **Purely Synchronous Handoff**: Fail the tool call if the recipient is offline. *Rejected* as it breaks the reliability of complex swarms.
*   **Using an External Broker**: Requiring Redis/RabbitMQ. *Rejected* to maintain the "Self-Contained Infrastructure" goal of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Messages are encrypted at rest. Only the verified owner of the `recipient_id` (verified via Ed25519 signature) can retrieve messages.
*   **Observability:** The UI provides an "A2A Mailbox Viewer" to monitor queue depths and message latencies.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
