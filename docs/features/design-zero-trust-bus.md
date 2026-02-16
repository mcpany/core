# Design Doc: Zero Trust Inter-Agent Bus
**Status:** Draft
**Created:** 2026-02-16

## 1. Context and Scope
As AI agent swarms (e.g., OpenClaw, CrewAI) become more prevalent, the need for secure, isolated communication between subagents is critical. Currently, agents often use insecure local sockets or shared filesystem directories, leading to "Yolo Mode" execution that bypasses standard security controls. MCP Any needs to provide a secure, authenticated message bus that facilitates inter-agent coordination without compromising host security.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide isolated communication channels for agents.
    *   Authenticate all messages between agents.
    *   Minimize host-level exposure (e.g., using named pipes or Docker-bound sockets).
    *   Integrate with MCP Any's policy engine for fine-grained access control.
*   **Non-Goals:**
    *   Replacing general-purpose message brokers (like RabbitMQ) for non-agentic tasks.
    *   Providing a public-facing communication API.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Coordinate 3 specialized agents to complete a coding task without exposing local environment variables to all agents.
*   **The Happy Path (Tasks):**
    1.  Orchestrator starts MCP Any with a "Swarm Profile".
    2.  MCP Any creates a secure, isolated bus channel (e.g., `/run/mcpany/swarm-1.sock`).
    3.  Agent 1 (Researcher) writes findings to the bus.
    4.  Agent 2 (Coder) receives findings and generates code.
    5.  Agent 3 (Reviewer) validates the code.
    6.  MCP Any logs all inter-agent traffic and enforces that Agent 1 cannot see Agent 2's specific secrets.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        AgentA[Agent A] <--> Bus[MCP Any Secure Bus]
        AgentB[Agent B] <--> Bus
        Bus <--> Policy[Policy Engine]
        Policy <--> Audit[Audit Log]
    ```
*   **APIs / Interfaces:**
    *   `mcp_send(target_agent_id, message_body)`
    *   `mcp_subscribe(topic_id)`
*   **Data Storage/State:**
    *   Transient message state held in memory; persistence via SQLite if configured for long-running swarms.

## 5. Alternatives Considered
*   **Shared Filesystem**: Rejected due to race conditions and lack of fine-grained access control.
*   **Standard HTTP/gRPC**: Rejected due to higher overhead and increased attack surface on the host network.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All agents must present a valid session token to access the bus. Messages are encrypted at rest (if persisted).
*   **Observability:** All inter-agent communication is logged to the central MCP Any audit stream.

## 7. Evolutionary Changelog
*   **2026-02-16:** Initial Document Creation.
