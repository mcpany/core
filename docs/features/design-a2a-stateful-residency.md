# Design Doc: A2A Stateful Residency (Agent Mailbox)
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As AI agent swarms (OpenClaw, CrewAI, AutoGen) move from synchronous "Call-and-Response" to asynchronous, long-running tasks, the communication between them becomes brittle. If a subagent goes offline or a network connection drops, the state is lost.

MCP Any needs to evolve from a simple protocol bridge to a **Resident State Layer**. By implementing A2A Stateful Residency, MCP Any acts as a persistent "Mailbox" and "Stateful Buffer" for agents, ensuring message delivery and session continuity even in intermittent environments.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a persistent, asynchronous message queue for Agent-to-Agent (A2A) communication.
    * Support "Handoff" semantics where an agent can "Park" a task and another can "Pick" it up later.
    * Maintain session state across agent restarts.
    * Implement TTL and "Dead Letter" policies for agent messages.
* **Non-Goals:**
    * Replacing full-blown message brokers like RabbitMQ for high-throughput non-agent data.
    * Managing the internal model weights or reasoning of the agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Decentralized Swarm Orchestrator.
* **Primary Goal:** Ensure a complex multi-step research task completes even if the specialized "Researcher" agent is temporarily throttled or disconnected.
* **The Happy Path (Tasks):**
    1. Parent Agent sends a "Task Request" to MCP Any's A2A Gateway.
    2. MCP Any validates the request and stores it in the **Stateful Residency Buffer**.
    3. The "Researcher" Agent (Subagent) connects and polls for "Available Tasks."
    4. MCP Any delivers the task and marks it as "In Progress."
    5. Researcher Agent completes the task and posts the result back to the "Mailbox."
    6. Parent Agent receives a notification (or polls) and retrieves the finalized state.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent A->>MCP Any: Post Message (Target: Agent B, Session: 123)
        MCP Any->>SQLite: Store Message (Status: Queued)
        MCP Any-->>Agent A: Ack (MsgID: 456)
        ... Agent B connects ...
        Agent B->>MCP Any: Poll Messages
        MCP Any->>SQLite: Fetch Queued for Agent B
        MCP Any-->>Agent B: Deliver Message (MsgID: 456)
        Agent B->>MCP Any: Update Message (Status: Processing)
    ```
* **APIs / Interfaces:**
    * `mcp_a2a_post(target_agent_id, session_id, payload)`: Post a message to the buffer.
    * `mcp_a2a_poll(agent_id)`: Retrieve messages for the calling agent.
    * `mcp_a2a_ack(msg_id)`: Mark a message as successfully processed.
* **Data Storage/State:**
    * Use the existing **Shared KV Store** (SQLite-based) to store the `messages` and `sessions` tables.

## 5. Alternatives Considered
* **Pure Pub/Sub (NATS/Redis)**: Rejected because it requires additional infrastructure dependencies. We want MCP Any to be "Single Binary" and "Safe-by-Default" for local/edge use.
* **Direct Agent-to-Agent HTTP**: Rejected because it requires agents to be network-addressable and online simultaneously, which fails in many mobile/local scenarios.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Messages are encrypted at rest. Only agents with valid **Capability Tokens** for a specific `session_id` can read/write to that session's mailbox.
* **Observability:** Introduce the **Agent Chain Tracer (A2A)** in the UI to visualize the message flow and latency between agents.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
