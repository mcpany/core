# Design Doc: A2A Message Residency
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As agent swarms (like OpenClaw and CrewAI) grow in complexity, the communication between specialized agents becomes a bottleneck. Current inter-agent communication is largely ephemeral and synchronous. If an agent crashes or a connection is dropped, the entire swarm's state can be lost.

MCP Any needs to evolve from a simple protocol bridge to a **Stateful Residency** for Agent-to-Agent (A2A) communication. This ensures that messages are persisted, can be retrieved asynchronously, and provide a "source of truth" for the swarm's progress.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a persistent, searchable mailbox for agent-to-agent messages.
    * Support asynchronous delivery (Producer-Consumer pattern for agents).
    * Enable "Snapshotting" of swarm communication for recovery.
    * Integrate with Zero Trust policies to ensure only authorized agents can read specific message streams.
* **Non-Goals:**
    * Implementing the LLM logic for the agents themselves.
    * Providing a general-purpose message queue (like RabbitMQ) for non-agentic workloads.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Ensure a complex task (e.g., "Analyze codebase and suggest security fixes") continues even if individual sub-agents fail or are rate-limited.
* **The Happy Path (Tasks):**
    1. The Orchestrator agent initializes a `SwarmSession` in MCP Any.
    2. Sub-agent A (Security Auditor) posts a "Vulnerability Report" message to the session.
    3. Sub-agent B (Fix Generator) is currently busy/offline.
    4. MCP Any persists the message in the `A2A Message Residency`.
    5. Sub-agent B comes online, queries MCP Any for new messages in the session.
    6. Sub-agent B retrieves the report and begins work.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>MCP Any (Residency): POST /a2a/send {session_id, recipient, payload}
        MCP Any (Residency)->>Storage (SQLite/Postgres): Save Message
        MCP Any (Residency)-->>Agent A: 202 Accepted (Message ID)
        Agent B->>MCP Any (Residency): GET /a2a/poll {session_id, agent_id}
        MCP Any (Residency)->>Storage (SQLite/Postgres): Query Unread Messages
        Storage (SQLite/Postgres)-->>MCP Any (Residency): Messages List
        MCP Any (Residency)-->>Agent B: 200 OK (Messages)
    ```
* **APIs / Interfaces:**
    * `mcp_a2a_send(recipient, message_type, payload)`: Tool exposed to agents to send messages.
    * `mcp_a2a_receive(session_id)`: Tool to retrieve messages.
    * Internal REST/gRPC API for residency management.
* **Data Storage/State:**
    * Messages stored in a dedicated `a2a_messages` table in the MCP Any database.
    * TTL (Time-to-Live) policies for message cleanup.

## 5. Alternatives Considered
* **Direct WebSockets**: Rejected because it requires both agents to be online simultaneously and doesn't provide native persistence.
* **External MQ (Redis/RabbitMQ)**: Rejected to minimize infrastructure footprint for local-first agent deployments. MCP Any should be self-contained.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Messages are scoped to `SessionIDs`. Agents must present a valid `SessionToken` or `AgentCapability` to access messages.
* **Observability:** Logging of message delivery latency and residency depth (number of pending messages).

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
