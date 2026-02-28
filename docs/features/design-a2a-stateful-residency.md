# Design Doc: A2A Stateful Residency (Stateful Buffer)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The shift towards "Headless Agentic Infrastructure" (e.g., OpenClaw Multi-Agent Mode) and the maturity of the A2A protocol have created a need for reliable, asynchronous communication between agents. Agents often have intermittent connectivity or different execution speeds. MCP Any must act as a "Resident" home for A2A state, providing a stateful mailbox/buffer that ensures messages are delivered and sessions are maintained even when agents are offline or transitioning between tasks.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a persistent, stateful buffer for A2A messages (Mailbox).
    *   Support asynchronous delivery with acknowledgement (ACK) mechanisms.
    *   Maintain session-aware state across intermittent agent connections.
    *   Ensure "Full-Chain Traceability" for every message to prevent cascading failures (ASI08).
*   **Non-Goals:**
    *   Replacing message brokers like RabbitMQ or Kafka for non-agentic workloads.
    *   Implementing long-term archival of messages (focus is on active session residency).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator.
*   **Primary Goal:** Ensure a "Researcher Agent" can send a 10MB dataset to a "Summarizer Agent" that is currently busy or offline, without the Researcher Agent timing out.
*   **The Happy Path (Tasks):**
    1.  Researcher Agent posts an A2A message to the Summarizer's address via MCP Any.
    2.  MCP Any validates the sender's identity and stores the message in the `Stateful Buffer`.
    3.  Summarizer Agent comes online or completes its current task and polls MCP Any for new messages.
    4.  MCP Any delivers the message and records the delivery in the `Traceability Engine`.
    5.  Summarizer Agent acknowledges receipt; MCP Any updates the session state.

## 4. Design & Architecture
*   **System Flow:**
    - **Ingestion**: The `A2A Residency Middleware` intercepts A2A tool calls and persists them.
    - **Persistence**: Messages are stored in the `Shared KV Store` (SQLite-backed) with TTL and status (QUEUED, DELIVERED, ACKED).
    - **Traceability**: Every state change is logged to the `Traceability Engine` with a unique `Chain-ID`.
*   **APIs / Interfaces:**
    - `POST /v1/a2a/mailbox/post`: For sending messages.
    - `GET /v1/a2a/mailbox/poll`: For receiving messages.
    - `POST /v1/a2a/mailbox/ack`: For confirming receipt.
*   **Data Storage/State:** SQLite table `a2a_messages` within the existing `Shared KV Store`.

## 5. Alternatives Considered
*   **Direct P2P Communication**: Agents talk directly to each other. *Rejected* due to lack of reliability and centralized observability/security (ASI07).
*   **In-Memory Only Buffer**: Using Go channels or memory maps. *Rejected* because state must survive MCP Any restarts.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Access to mailboxes is restricted via the `Identity-Based Microsegmentation` layer. Agents can only read messages addressed to them.
*   **Observability:** The UI provides a `Stateful A2A Mailbox` view to monitor message queues and delivery latencies.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
