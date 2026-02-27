# Design Doc: Isolated Local Transport (Named Pipes/UDS)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Standard MCP and A2A implementations often rely on local HTTP servers (loopback) for inter-process communication. This exposes agents to "local network" exploits where rogue processes or subagents can scan and interact with these ports. As agent swarms become more complex and handle sensitive data, we need a transport layer that provides stronger isolation. Named Pipes (Windows) and Unix Domain Sockets (UDS) offer a filesystem-based permissions model that eliminates the need for network port exposure.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement support for Unix Domain Sockets (UDS) on Linux/macOS.
    *   Implement support for Named Pipes on Windows.
    *   Provide a transparent fallback mechanism to TCP if isolated transport is unavailable.
    *   Enable capability-based access control using filesystem permissions on the socket/pipe files.
*   **Non-Goals:**
    *   Replacing TCP for remote (cross-host) communication.
    *   Implementing a new protocol (this is a transport layer for MCP/A2A).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Developer.
*   **Primary Goal:** Run a local MCP server that is only accessible to a specific agent process, without opening any network ports.
*   **The Happy Path (Tasks):**
    1.  Developer configures MCP Any to use a UDS path: `unix:///tmp/mcp-any.sock`.
    2.  MCP Any creates the socket file with restricted permissions (e.g., `0600`).
    3.  The Agent (e.g., Claude Code) connects to the socket path instead of `localhost:port`.
    4.  Communication proceeds normally over the isolated channel.
    5.  A rogue process on the same machine attempts to connect but is blocked by OS-level filesystem permissions.

## 4. Design & Architecture
*   **System Flow:**
    - **Initialization**: MCP Any detects the `unix://` or `pipe://` prefix in the configured address.
    - **Socket/Pipe Creation**: Uses Go's `net.Listen("unix", path)` or a Windows-specific library for Named Pipes.
    - **Access Control**: MCP Any explicitly sets `chmod` on the UDS file or applies an ACL to the Named Pipe.
*   **APIs / Interfaces:**
    - New transport configuration options in `config.yaml`.
    - Internal `Transport` interface abstraction in the server to handle different `net.Listener` types.
*   **Data Storage/State:** Socket/Pipe paths are managed as part of the service runtime state.

## 5. Alternatives Considered
*   **Mutual TLS (mTLS) on Loopback**: Provides encryption and identity but still leaves ports open and is more complex to manage (certificate lifecycle).
*   **Kernel-level Firewalls (e.g., eBPF)**: Highly secure but requires root privileges and is complex to implement across different OSs.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Leverages OS-native security. MCP Any must ensure it cleans up socket files on shutdown to prevent "stale socket" errors.
*   **Observability:** Logs should clearly indicate which transport is being used for each connection.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
