# Design Doc: Isolated Subagent Communication (Named Pipes)
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
Autonomous agent frameworks like OpenClaw currently rely on local HTTP tunneling for inter-agent communication. Recent vulnerability reports indicate that this pattern exposes local ports to CSRF attacks and unauthorized access by other processes on the host. MCP Any needs to provide a secure, isolated alternative for agents to communicate when running on the same host or within the same container cluster.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Eliminate the need for local TCP/HTTP ports for inter-agent communication.
    *   Provide a secure, authenticated channel for subagent routing.
    *   Support Docker-bound named pipes and Unix domain sockets.
    *   Ensure compatibility with existing MCP JSON-RPC protocol.
*   **Non-Goals:**
    *   Replacing remote (cross-network) MCP communication.
    *   Implementing a full message queue (this is a transport layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator (e.g., OpenClaw user).
*   **Primary Goal:** Run a parent agent and three subagents on a single Mac Mini or Docker host without exposing any network ports.
*   **The Happy Path (Tasks):**
    1.  User configures MCP Any to use the `pipe` transport for the subagent registry.
    2.  MCP Any creates a named pipe (or Unix socket) at a specific filesystem path (e.g., `/tmp/mcp-any/subagent-1.pipe`).
    3.  The subagent connects to this pipe instead of an HTTP URL.
    4.  Parent agent sends tool calls to the subagent through MCP Any, which routes them via the isolated pipe.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Parent[Parent Agent] -->|MCP/JSON-RPC| Core[MCP Any Core]
        Core -->|Local Router| Registry[Subagent Registry]
        Registry -->|Named Pipe| Sub1[Subagent 1]
        Registry -->|Unix Socket| Sub2[Subagent 2]
    ```
*   **APIs / Interfaces:**
    *   New Transport type: `pipe`.
    *   Configuration parameter: `pipe_path` (filesystem path).
*   **Data Storage/State:**
    *   Pipes are transient filesystem artifacts managed by the MCP Any lifecycle.
    *   Access is controlled via standard filesystem permissions (POSIX).

## 5. Alternatives Considered
*   **mTLS over Localhost:** Rejected due to complexity of certificate management for local ephemeral agents.
*   **Shared Memory:** Rejected as it requires complex synchronization and is less portable than pipes/sockets.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** By using filesystem-based communication, we leverage OS-level access control. Only processes with specific UID/GID can access the communication channel.
*   **Observability:** MCP Any will log all traffic passing through the pipes (when debug is enabled) and provide latency metrics for local calls.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation. Resolving Local Port Exposure issue identified in OpenClaw market sync.
