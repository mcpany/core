# Design Doc: Inter-Agent Message Bus (Orchestration Bus)

**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
The rapid evolution of agent swarms, specifically Claude Code's "Agent Teams" and OpenClaw's specialized subagents, has created a need for a standardized, high-performance messaging layer. Currently, agents often operate in silos or rely on fragmented, third-party MCP servers for basic communication. MCP Any will provide a native "Inter-Agent Message Bus" that acts as the backbone for agentic coordination, enabling seamless task delegation and state synchronization across different agent frameworks.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement an asynchronous, message-oriented middleware for inter-agent communication.
    *   Support "Pub/Sub" and "Request/Response" patterns for agent coordination.
    *   Provide a "Team Lead" capability where one agent can orchestrate tasks across multiple peers.
    *   Ensure all messages are authenticated and conform to Zero-Trust policies.
    *   Native integration with Claude Code's `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`.
*   **Non-Goals:**
    *   Replacing the LLM's reasoning for task allocation.
    *   Building a general-purpose chat application for humans.
    *   Persistent long-term storage of all messages (focus is on active session coordination).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Team Lead (e.g., Claude Code session).
*   **Primary Goal:** Delegate a security audit task to a specialized "Security Agent" and wait for a structured report.
*   **The Happy Path (Tasks):**
    1.  The Team Lead agent initializes a "Mission Session" on the MCP Any Bus.
    2.  The Team Lead publishes a task to the `tasks.security_audit` topic.
    3.  A specialized Security Subagent (OpenClaw-based) subscribed to that topic receives the task.
    4.  The Security Subagent processes the task and publishes the result back to the Mission's `results` topic.
    5.  The Team Lead receives the notification and synthesizes the final audit report.

## 4. Design & Architecture
*   **System Flow:**
    - **Message Broker**: An internal, high-speed broker managed by the MCP Any server.
    - **Topic Registry**: Dynamic creation of topics based on agent roles and session IDs.
    - **Identity Mapping**: MCP Any maps agent session tokens to Bus identities to ensure secure routing.
*   **APIs / Interfaces:**
    - `bus.publish(topic, payload)`: Send a message to a topic.
    - `bus.subscribe(topic)`: Listen for messages on a topic.
    - `bus.call(agent_id, task)`: Direct request/response delegation.
*   **Data Storage/State:**
    - In-memory message buffer with optional SQLite backing for session recovery.
    - State is isolated per "Mission" using the Blackboard's row-level security.

## 5. Alternatives Considered
*   **Direct P2P WebSockets**: Agents connecting directly to each other. *Rejected* as it bypasses central security policy enforcement and discovery.
*   **Using standard MCP Tool Calls**: Forcing all communication through synchronous tool calls. *Rejected* because coordination often requires asynchronous "long-running" tasks that don't fit the JSON-RPC tool call timeout model.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Mandatory message signing. Agents can only subscribe to topics within their authorized "Intent Scope."
*   **Observability:** A "Message Tracer" in the UI to visualize the flow of tasks and responses across the agent mesh.

## 7. Evolutionary Changelog
*   **2026-03-10:** Initial Document Creation.
