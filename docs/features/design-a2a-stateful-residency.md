# Design Doc: A2A Stateful Residency (Agent Mailbox)

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As AI agents evolve into complex swarms (e.g., OpenClaw, CrewAI), the primary bottleneck is reliable communication between specialized subagents. Current Agent-to-Agent (A2A) interactions are often synchronous and ephemeral; if a subagent is offline or a connection drops, the entire workflow fails. MCP Any must provide a "Stateful Residency" layer—a persistent mailbox system—that buffers messages, manages handoffs, and ensures delivery even across intermittent network boundaries.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a persistent message queue (Mailbox) for A2A communication.
    *   Support asynchronous "Fire and Forget" and "Request-Response" patterns.
    *   Provide "Swarm Pulse" telemetry to track message delivery and agent health.
    *   Ensure context inheritance is preserved within the stateful buffer.
*   **Non-Goals:**
    *   Replacing dedicated message brokers like RabbitMQ for massive-scale non-agent workloads.
    *   Building a general-purpose database (state is limited to agent sessions).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Swarm Orchestrator (e.g., a "Manager" agent).
*   **Primary Goal:** Delegate a long-running research task to a "Researcher" agent without staying synchronously connected.
*   **The Happy Path (Tasks):**
    1.  Manager agent sends a task message to the `researcher-mailbox` via MCP Any.
    2.  MCP Any acknowledges receipt and persists the message in SQLite.
    3.  Researcher agent polls its mailbox when ready, retrieves the task, and processes it.
    4.  Researcher posts the result back to the `manager-mailbox`.
    5.  Manager retrieves the result at its next convenience.

## 4. Design & Architecture
*   **System Flow:**
    - **Message Ingest**: A2A Bridge receives an MCP-wrapped message and routes it to the `MailboxManager`.
    - **Persistence Layer**: Messages are stored in an embedded SQLite database with `status` (queued, delivered, acknowledged).
    - **Delivery Logic**: Supports both long-polling and Webhook-based push notifications for agents.
*   **APIs / Interfaces:**
    - `mcp_send_message(to: string, body: object, context_token: string)`
    - `mcp_poll_mailbox(mailbox_id: string) -> Message[]`
    - `mcp_get_swarm_pulse() -> SwarmHealthReport`
*   **Data Storage/State:**
    - SQLite table: `a2a_messages` (id, sender, recipient, payload, context_headers, timestamp, status).

## 5. Alternatives Considered
*   **Purely Synchronous Proxy**: Lower complexity, but fails during agent downtime. *Rejected* due to swarm reliability requirements.
*   **External Redis/NATS**: Highly scalable, but introduces heavy operational dependencies. *Rejected* in favor of "Zero-Config" embedded SQLite for MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Mailboxes are protected by the Policy Firewall. Agents can only read from mailboxes they are explicitly authorized to access via capability tokens.
*   **Observability:** The UI provides an "Agent Chain Tracer" to visualize the flow of messages through the resident state.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
