# Design Doc: A2A Stateful Residency (Persistent Mailbox)

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As agent swarms (e.g., OpenClaw, CrewAI) evolve into asynchronous, long-running systems, the assumption of synchronous, always-on connectivity between agents is failing. Agents frequently go "offline" during intense reasoning cycles or when waiting for external events. MCP Any must provide a "Stateful Residency" layer—a persistent mailbox—that buffers Agent-to-Agent (A2A) messages, ensuring reliable communication and state consistency in decentralized swarms.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a persistent buffer (Mailbox) for A2A messages using the A2A protocol.
    *   Enable "Asynchronous Handoffs" where Agent A can send a task to Agent B even if Agent B is currently busy or disconnected.
    *   Provide message TTL (Time-to-Live) and delivery guarantees (at-least-once).
    *   Expose a "Mailbox API" for agents to poll or subscribe to incoming messages.
*   **Non-Goals:**
    *   Replacing full-featured message brokers like RabbitMQ or Kafka for non-agent workloads.
    *   Implementing complex orchestration logic (MCP Any remains the *bus*, not the *orchestrator*).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw).
*   **Primary Goal:** Delegate a 30-minute research task from a "Planner Agent" to a "Worker Agent" without maintaining an active connection.
*   **The Happy Path (Tasks):**
    1.  Planner Agent calls the `a2a_send` tool via MCP Any, targeting the Worker Agent.
    2.  MCP Any acknowledges receipt and stores the message in the `Stateful Residency` buffer.
    3.  Worker Agent, upon finishing its current task, polls its mailbox via the `a2a_receive` tool.
    4.  Worker Agent processes the message and sends the result back to the Planner's mailbox.

## 4. Design & Architecture
*   **System Flow:**
    - **A2A Gateway**: Intercepts A2A-formatted MCP calls.
    - **Persistence Layer**: An embedded SQLite store (leveraging the `Shared KV Store` infrastructure) to hold queued messages.
    - **Dispatcher**: Handles message routing, TTL enforcement, and delivery status tracking.
*   **APIs / Interfaces:**
    - `a2a_send(recipient_id, message_body, priority, ttl)`
    - `a2a_receive(agent_id, limit)`
    - `a2a_status(message_id)`
*   **Data Storage/State:**
    - Table: `a2a_messages` (id, sender, recipient, body, status, created_at, expires_at).

## 5. Alternatives Considered
*   **Synchronous-Only Proxy**: Force agents to be online. *Rejected* because reasoning-heavy agents (e.g., o1/o3-style) have multi-minute latencies that break standard timeouts.
*   **Stateless Retries**: Have the sender retry. *Rejected* as it increases network overhead and doesn't solve the "offline agent" problem.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Messages are encrypted at rest. Access to a mailbox requires a valid `AgentIdentityToken` corresponding to the `recipient_id`.
*   **Observability:** A dedicated "A2A Mesh Tracer" in the UI to visualize the flow of buffered messages and their delivery status.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
