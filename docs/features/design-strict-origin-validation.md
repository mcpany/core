# Design Doc: Strict Origin Validation Layer
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
The "OpenClaw Hijacking" incident (March 2026) demonstrated that standard loopback (localhost) HTTP/WebSocket listeners are vulnerable to Cross-Origin Resource Sharing (CORS) bypasses or DNS rebinding when an agent-connected browser visits a malicious site. MCP Any must ensure that only authorized, local processes can communicate with its gateway and adapters.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate browser-based hijacking of local MCP tool endpoints.
    * Implement cryptographic origin validation for all incoming local requests.
    * Transition from standard TCP loopback to Named Pipes (Windows) and Unix Domain Sockets (Linux/macOS) by default.
* **Non-Goals:**
    * Providing a full identity management system (IAM).
    * Securing remote (WAN) connections (covered by the "Safe-by-Default" hardening feature).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using OpenClaw with MCP Any.
* **Primary Goal:** Ensure a malicious website cannot trigger local tools via the OpenClaw agent.
* **The Happy Path (Tasks):**
    1. MCP Any starts and binds to `/tmp/mcp-any.sock` (or named pipe).
    2. MCP Any generates a unique `Gateway-Token` and writes it to a secure, local-only config file (`~/.mcp/session.json`).
    3. The agent (OpenClaw) reads the token and uses it to authenticate every request over the socket.
    4. A malicious website tries to fetch `localhost:PORT` but fails because no port is listening, or tries to hit the socket and fails because it lacks the token and OS-level socket permissions.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Named Pipe / UDS` -> `MCP Any Auth Middleware` -> `Tool Gateway`
    1. **Transport Migration**: Move listeners to `\\.\pipe\mcp-any` or `/var/run/mcp-any.sock`.
    2. **Token Handshake**: Implement a "Local Secret" handshake.
* **APIs / Interfaces:**
    * `X-MCP-Origin-Token`: Mandatory header for all requests.
* **Data Storage/State:**
    * Ephemeral session tokens stored in memory and synchronized via protected local files.

## 5. Alternatives Considered
* **CORS Hardening**: Rejected as insufficient against DNS rebinding or direct socket manipulation from malicious browser extensions.
* **Mutual TLS (mTLS)**: Considered, but the overhead of certificate management for local-only dev environments was deemed too high compared to Unix Sockets + Tokens.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Follows the principle of "Local-Only Attestation."
* **Observability**: Unauthorized connection attempts are logged with source metadata (PID/UID where possible).

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
