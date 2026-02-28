# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The current A2A (Agent-to-Agent) communication model relies on synchronous or near-synchronous connectivity between agents. However, in complex swarms, agents may be transiently available, running in different environments (local vs. cloud), or operating on different timescales. The "8,000 Exposed Servers" and "ClawHavoc" incidents also highlighted that agents need a secure, sovereign "home" for their state and messages to prevent interception or loss during handoffs.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a persistent, encrypted "Mailbox" for A2A messages within MCP Any.
    *   Support asynchronous delivery where Agent A can post a task and Agent B can retrieve it when online.
    *   Enable "Handoff Persistence" so that session state (Context ID, Trace ID) survives agent restarts.
    *   Implement "Dead Letter Queues" for failed agent delegations.
*   **Non-Goals:**
    *   Replacing message brokers like RabbitMQ for high-throughput non-agentic data.
    *   Implementing agent-specific logic for message processing.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Developer.
*   **Primary Goal:** Ensure a "Researcher" agent can delegate to a "Writer" agent even if the Writer agent is temporarily offline or being restarted.
*   **The Happy Path (Tasks):**
    1.  Researcher Agent calls `send_a2a_message(to="Writer", payload={...})`.
    2.  Writer Agent is offline; MCP Any accepts the message and stores it in the `Stateful Buffer`.
    3.  Writer Agent comes online and calls `get_pending_a2a_messages()`.
    4.  MCP Any delivers the message along with the inherited `Recursive Context`.
    5.  Writer processes the task and posts the result back to the buffer for the Researcher.

## 4. Design & Architecture
*   **System Flow:**
    - **Ingress**: Messages are received via the `A2ABridgeMiddleware` and validated against the `Policy Firewall`.
    - **Storage**: Messages are stored in an encrypted SQLite table (using SQLCipher or similar) to ensure "Mesh Sovereignty."
    - **Notification**: If the recipient agent is connected via WebSocket, a "New Message" event is pushed. Otherwise, it waits for a poll.
*   **APIs / Interfaces:**
    - `POST /a2a/mailbox/send`: Enqueue a message.
    - `GET /a2a/mailbox/receive`: Poll for new messages.
    - `ACK /a2a/mailbox/:msg_id`: Confirm message processing.
*   **Data Storage/State:**
    - `a2a_messages` table: `id`, `from_did`, `to_did`, `payload` (encrypted), `context_id`, `status` (queued, delivered, acked).

## 5. Alternatives Considered
*   **Purely Synchronous A2A**: Simply failing if the target agent is offline. *Rejected* as it makes agent swarms too fragile.
*   **External Redis/Kafka**: Requiring an external broker. *Rejected* to maintain MCP Any's "Local-First" and "Zero-Dependency" philosophy for the core agentic bus.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All messages in the buffer are encrypted at rest using the instance's Ed25519-derived key. Only the intended recipient (verified by DID or Session Token) can decrypt/retrieve.
*   **Observability:** The UI "Agent Chain Tracer" will show messages as "Pending in Buffer" or "Delivered."

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
