# Design Doc: Domain Socket / Named Pipe Gateway
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The "ClawJacked" exploit demonstrated that AI agents bound to `localhost` TCP/WebSocket ports are vulnerable to cross-origin attacks from web browsers. Since browsers do not restrict WebSocket connections to `localhost`, a malicious website can send commands to a local agent. MCP Any must provide a secure alternative transport for local inter-process communication (IPC) that is inherently immune to browser-based attacks.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement Unix Domain Socket (UDS) support for Linux/macOS.
    * Implement Named Pipes support for Windows.
    * Provide a fallback mechanism for legacy tools with ephemeral, tokenized TCP.
    * Ensure the UI can communicate with the server via these secure IPC channels.
* **Non-Goals:**
    * Replacing remote (over-the-internet) HTTP/gRPC transports.
    * Implementing a custom encryption layer over the sockets (OS-level permissions are sufficient).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local AI Developer / Agent Swarm Orchestrator.
* **Primary Goal:** Run a local agent swarm and a management UI without exposing any TCP ports to potential browser-based hijacking.
* **The Happy Path (Tasks):**
    1. User starts MCP Any server.
    2. Server detects OS and creates a socket file at `~/.mcpany/server.sock` (or a Named Pipe `\\.\pipe\mcpany`).
    3. Server sets file permissions to `0600` (user-only).
    4. UI and Subagents connect directly to the socket/pipe.
    5. No TCP ports are opened for local communication.

## 4. Design & Architecture
* **System Flow:**
    `[Local Agent/UI] <--> [Unix Domain Socket / Named Pipe] <--> [MCP Any Gateway]`
* **APIs / Interfaces:**
    * New transport type: `IPC` (in addition to `Stdio`, `HTTP`, `SSE`).
    * Config option: `transport: ipc` with `path` parameter.
* **Data Storage/State:**
    * Socket files are ephemeral and cleaned up on shutdown.

## 5. Alternatives Considered
* **Mutual TLS (mTLS) for localhost:** Rejected due to the complexity of certificate management for local-only development.
* **Custom Auth Headers in WebSockets:** Rejected because some browser exploits can still initiate the handshake or perform timing attacks; OS-level IPC is a cleaner "Safe-by-Default" solution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** OS-level filesystem permissions (e.g., `0600`) ensure that only the current user can access the socket. This completely bypasses the browser's reach.
* **Observability:** Logs will track socket creation, permission settings, and connection counts.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
