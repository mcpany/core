# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
As the Agent-to-Agent (A2A) protocol matures, a major bottleneck is the reliability of delivery between agents with intermittent connectivity (e.g., local developer agents delegating to cloud-based swarms). If the receiver is offline, the message is lost. MCP Any must evolve from a simple protocol bridge to a **Resident Mesh Node** that provides stateful "mailboxes" for A2A messages, ensuring reliable, asynchronous delivery and shared state persistence.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide persistent message queuing (Resident Mailboxes) for A2A messages.
    *   Implement a "Store and Forward" mechanism for agents with intermittent connectivity.
    *   Support asynchronous callbacks and task-state tracking.
    *   Expose a "Mailbox Tool" via MCP for agents to poll or receive notifications of pending tasks.
*   **Non-Goals:**
    *   Building a general-purpose message broker (e.g., RabbitMQ). It is strictly for A2A protocol messages.
    *   Managing long-term archival of agent conversations.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Agent Swarm Developer.
*   **Primary Goal:** Ensure a task delegated from a local laptop agent is successfully received by a cloud-based refinement swarm, even if the laptop goes offline temporarily.
*   **The Happy Path (Tasks):**
    1.  Laptop Agent posts a task to the MCP Any A2A Resident Mailbox.
    2.  MCP Any persists the task in its `Shared KV Store`.
    3.  Cloud Swarm connects to MCP Any and retrieves the pending task.
    4.  Laptop Agent goes offline.
    5.  Cloud Swarm completes the task and posts the result back to the Mailbox.
    6.  Laptop Agent reconnects and retrieves the completed task result.

## 4. Design & Architecture
*   **System Flow:**
    - **Ingress**: A2A messages are received via the `A2A Interop Bridge` and passed to the `ResidencyMiddleware`.
    - **Persistence**: Messages are stored in an embedded SQLite database (leveraging the `Shared KV Store` infrastructure).
    - **Egress**: Messages are delivered via long-polling, WebSockets, or explicit tool-based retrieval.
*   **APIs / Interfaces:**
    - `POST /v1/a2a/mailbox/send`: Send a message to a resident mailbox.
    - `GET /v1/a2a/mailbox/poll`: Retrieve pending messages.
    - MCP Tool: `mcp_a2a_check_mailbox()`: Allows an agent to check for work.
*   **Data Storage/State:**
    - Table: `a2a_messages` (id, sender, receiver, payload, status, expires_at).

## 5. Alternatives Considered
*   **Stateless Retries**: Letting agents handle retries. *Rejected* as it drains battery/resources and doesn't handle long-term offline states.
*   **External Broker (Redis)**: Requiring a separate Redis instance. *Rejected* to maintain the "Single Binary" and "Local First" principles of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Mailboxes are scoped to specific Agent Identities. Access requires cryptographic attestation as defined in the `Safe-by-Default` design.
*   **Observability:** The UI provides an "A2A Mailbox Viewer" to monitor message queues, delivery status, and latency.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
