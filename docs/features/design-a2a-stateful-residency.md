# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
As multi-agent swarms become more complex and asynchronous, the need for a reliable "mailbox" or "buffer" for agent-to-agent (A2A) communications has become critical. Agents often operate across different network boundaries, may have intermittent connectivity, or require long-running task processing. MCP Any, as the universal agent bus, is uniquely positioned to provide a stateful residency layer that ensures no A2A messages are lost and that state is preserved across agent handoffs.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a persistent, encrypted "mailbox" for A2A messages.
    *   Enable asynchronous "fire and forget" task delegation between agents.
    *   Support "At-Least-Once" delivery guarantees for A2A messages.
    *   Maintain a historical audit log of all inter-agent communications for observability.
*   **Non-Goals:**
    *   Implementing agent-specific logic (e.g., deciding *what* to send).
    *   Replacing direct, low-latency A2A communication for real-time tasks.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Agent Swarm Orchestrator.
*   **Primary Goal:** Ensure a "Design Agent" can delegate a long-running rendering task to an "Image Agent" without staying online to wait for the result.
*   **The Happy Path (Tasks):**
    1.  Design Agent sends an A2A message to the residency buffer via MCP Any.
    2.  MCP Any persists the message in the `Stateful Buffer` and returns a `receipt_id`.
    3.  Design Agent goes offline or moves to another task.
    4.  Image Agent polls or receives a notification from MCP Any, retrieves the task, and processes it.
    5.  Image Agent posts the result back to the `Stateful Buffer`.
    6.  Design Agent (or a monitoring agent) retrieves the result using the `receipt_id` later.

## 4. Design & Architecture
*   **System Flow:**
    - **Ingestion**: The `A2A Residency Middleware` receives messages and assigns a unique `MessageID` and `CorrelationID`.
    - **Persistence**: Messages are stored in a dedicated table in the `MCPANY_DB_PATH` (SQLite), encrypted at rest.
    - **Notification**: MCP Any can optionally trigger a webhook or push notification to the recipient agent when a new message is residency-bound.
    - **Retrieval**: Agents use a "Pull" or "Long-Polling" API to retrieve pending messages.
*   **APIs / Interfaces:**
    - `POST /a2a/buffer`: Ingest a message.
    - `GET /a2a/buffer/pending`: List pending messages for the caller.
    - `GET /a2a/buffer/message/{id}`: Retrieve a specific message.
*   **Data Storage/State:**
    - Table: `a2a_messages` (id, correlation_id, sender_id, recipient_id, payload, status, created_at, expires_at).

## 5. Alternatives Considered
*   **Direct-Only Communication**: Forcing agents to be online simultaneously. *Rejected* due to lack of resilience and inability to handle long-running tasks.
*   **External Message Queue (e.g., RabbitMQ)**: Requiring users to manage separate infrastructure. *Rejected* to maintain the "Single Binary / Zero Config" principle of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All messages in the residency buffer are tied to the session's Zero Trust identity. Access to the buffer is governed by the Policy Firewall.
*   **Observability:** The UI provides a "Stateful A2A Mailbox" view, allowing users to track the status of queued and delivered messages across the swarm.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
