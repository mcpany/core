# Design Doc: Intent-Aware Routing Channels

**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
As agent swarms grow in complexity, the communication between parent agents and specialized sub-agents becomes a performance and reliability bottleneck. Current transport layers often use ambiguous paths, leading to "Message Drift" where results are lost or misrouted. OpenClaw's latest updates emphasize the need for dedicated, high-reliability channels for sub-agent coordination. MCP Any needs to implement these channels to ensure stable state handoffs in deep agentic chains.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide isolated, dedicated routing channels for inter-agent communication.
    *   Ensure 100% delivery reliability for sub-agent results to parent agents.
    *   Integrate with existing MCP transport types (Stdio, HTTP, FastMCP).
*   **Non-Goals:**
    *   Replacing the underlying transport protocols.
    *   Managing the internal logic of the agents themselves.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator.
*   **Primary Goal:** Successfully coordinate a research task involving 3 specialized sub-agents without losing state.
*   **The Happy Path (Tasks):**
    1.  The parent agent initiates a task card via UACO.
    2.  MCP Any allocates a dedicated "Intent-Aware Channel" for this session.
    3.  Sub-agents communicate their intermediate results through this isolated channel.
    4.  The parent agent receives all results consistently and completes the task.

## 4. Design & Architecture
*   **System Flow:**
    `Parent Agent -> MCP Any (Channel Allocator) -> Dedicated Intent Channel -> Sub-Agent`
*   **APIs / Interfaces:**
    - `POST /v1/channels/allocate`: Creates a session-bound isolated channel.
    - `GET /v1/channels/{id}/stream`: Persistent stream for channel communication.
*   **Data Storage/State:**
    Channel state is managed in-memory with optional persistence to the Shared KV Store (Blackboard) for long-running tasks.

## 5. Alternatives Considered
*   **Global Message Bus:** Rejected due to "Message Drift" and lack of isolation between unrelated swarms.
*   **Direct A2A P2P:** Rejected because it bypasses MCP Any's security and governance layer.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Channels are bound to a specific "Intent-Scope." Only agents with the correct PoI (Proof-of-Intent) tokens can join or read from a channel.
*   **Observability:** Each channel provides a telemetry stream for the Agent Chain Tracer UI.

## 7. Evolutionary Changelog
*   **2026-03-25:** Initial Document Creation.
