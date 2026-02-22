# Design Doc: Isolated Named Pipes for Inter-Agent Comms

**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
Autonomous agent swarms (like OpenClaw or CrewAI) often deploy subagents that communicate via local MCP servers. Currently, these servers frequently use local HTTP ports (e.g., `localhost:50051`). This exposure creates a security risk: any process on the host machine can potentially access these ports, and rogue subagents can exploit SSRF vulnerabilities to access other local services.

MCP Any needs to provide a secure, isolated transport mechanism for inter-agent communication that bypasses the host network stack.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a transport layer using Unix Domain Sockets or Windows Named Pipes.
    *   Ensure isolation between different agent swarms on the same host.
    *   Support seamless fallback to HTTP/Stdio if pipes are unavailable.
*   **Non-Goals:**
    *   Encrypting traffic (assuming local isolation is sufficient for now).
    *   Providing remote access via named pipes (use VPN/SSH for that).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local HTTP ports to the host network.
*   **The Happy Path (Tasks):**
    1.  Orchestrator starts MCP Any with the `named-pipe` transport enabled.
    2.  MCP Any creates a unique socket file in a protected directory (e.g., `/run/mcpany/swarm-alpha.sock`).
    3.  Subagents connect to this socket using the standard MCP JSON-RPC protocol.
    4.  Host processes (non-root) are unable to access the socket due to filesystem permissions.

## 4. Design & Architecture
*   **System Flow:**
    `[Orchestrator] <--> [MCP Any (Pipe Adapter)] <--> [/run/mcpany/agent.sock] <--> [Subagent]`
*   **APIs / Interfaces:**
    *   New Upstream type: `named_pipe`
    *   Config fields: `socket_path`, `permissions`
*   **Data Storage/State:**
    *   State is managed via existing in-memory registries. Socket files are cleaned up on shutdown.

## 5. Alternatives Considered
*   **mTLS over HTTP:** Rejected due to the complexity of certificate management for ephemeral local agents.
*   **Stdio:** While secure, it's limited to 1:1 parent-child relationships and doesn't support the multi-agent "bus" model well.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Filesystem permissions (0600) ensure only the orchestrator and its children can access the pipe.
*   **Observability:** Socket connection/disconnection events logged to the Audit Log.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
