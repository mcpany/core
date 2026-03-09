# Design Doc: Origin-Bound Security Middleware

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
The March 2026 OpenClaw vulnerability revealed that AI agents running on local machines are susceptible to hijacking via malicious websites. Since these agents often expose an HTTP or WebSocket interface for tool interaction, a rogue web page can send requests to `localhost` and trigger sensitive actions. MCP Any, as a universal gateway, must ensure that only authorized origins (e.g., specific IDEs, verified agent CLI tools) can communicate with it.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement strict `Origin` and `Host` header validation for all incoming requests.
    *   Introduce a "Handshake Token" that must be provided by local clients during session initiation.
    *   Enable cryptographic validation of client identity for high-privilege tool calls.
    *   Provide a default-deny policy for all cross-origin requests unless explicitly whitelisted.
*   **Non-Goals:**
    *   Replacing standard authentication (this is an additional layer specifically for origin verification).
    *   Managing browser-level security policies (we rely on existing CORS but add server-side enforcement).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using a local Agent (e.g., OpenClaw) with MCP Any.
*   **Primary Goal:** Prevent a malicious website from executing local tools via the agent's gateway.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any. It generates a one-time `SessionSecret` and stores it in a local-only file.
    2.  Authorized Agent (e.g., Claude Code) reads the `SessionSecret` from the local filesystem.
    3.  Agent includes `X-MCP-Session-Token: <SessionSecret>` in its requests.
    4.  A malicious website attempts to fetch `localhost:3000/tools`. Since it cannot read the local `SessionSecret` (due to Same-Origin Policy for file access), its request is rejected by MCP Any.

## 4. Design & Architecture
*   **System Flow:**
    - **Request Interceptor**: Every incoming HTTP/WS request passes through the `OriginValidator`.
    - **Header Check**: The validator checks `Origin`, `Host`, and `Referer` headers.
    - **Token Validation**: For `localhost` requests, it requires the `X-MCP-Session-Token`.
*   **APIs / Interfaces:**
    - `middleware.OriginGuard`: Rejects requests with non-whitelisted origins.
    - `auth.SessionProvider`: Manages the generation and rotation of session-bound secrets.
*   **Data Storage/State:**
    - Active session tokens are stored in memory.
    - Whitelisted origins are stored in `config.yaml`.

## 5. Alternatives Considered
*   **Pure CORS**: Relying only on browser CORS headers. *Rejected* because `curl` or non-browser tools can bypass CORS, and some "no-cors" requests still reach the server.
*   **IP Whitelisting**: Restricting access to `127.0.0.1`. *Rejected* because the OpenClaw exploit specifically targeted `localhost` from a browser running on the same IP.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a direct implementation of the "Execution Boundary" protection.
*   **Observability:** Log all rejected origin attempts with the offending header values for security auditing.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
