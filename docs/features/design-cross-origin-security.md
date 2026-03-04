# Design Doc: Cross-Origin Security Middleware

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
The "Agent-Hijack" exploit (OpenClaw) discovered in March 2026 demonstrated that local agentic services are vulnerable to Cross-Origin Request Forgery (CORF). A malicious website running in a user's browser could send requests to a local AI agent (typically listening on `localhost`) to execute tools or steal data. MCP Any needs a robust defense-in-depth middleware to validate the origin and identity of every incoming request.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement mandatory `Origin` and `Host` header validation for all HTTP/WebSocket listeners.
    *   Introduce a "Trusted App Handshake" for non-browser local connections.
    *   Provide a "Secure Bridge" protocol for browser-based agents (e.g., Claude Code, Gemini CLI) to prove their identity via local file-based secrets.
    *   Reject any request from an unverified or blacklisted origin by default.
*   **Non-Goals:**
    *   Implementing a full OAuth2 server.
    *   Blocking legitimate cross-origin requests that have been explicitly authorized by the user.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using Claude Code with MCP Any.
*   **Primary Goal:** Prevent a malicious tab in Chrome from calling MCP Any tools.
*   **The Happy Path (Tasks):**
    1.  MCP Any starts with `Cross-Origin Security Middleware` enabled.
    2.  Claude Code initiates a connection.
    3.  Middleware detects the connection and challenges Claude Code to provide a `session_token` found in `~/.mcpany/sessions/`.
    4.  Claude Code provides the token; the origin is marked as "Trusted".
    5.  A malicious website tries to `fetch('http://localhost:3000/tools/list')`.
    6.  Middleware sees the `Origin: https://evil-site.com` header, finds no session token, and returns `403 Forbidden`.

## 4. Design & Architecture
*   **System Flow:**
    - **Origin Inspector**: First layer of the middleware. Compares the `Origin` header against a `trusted_origins` whitelist.
    - **Handshake Challenger**: If the origin is unknown, it looks for a `X-MCP-Session-Token`.
    - **Session Manager**: Validates tokens against a temporary, local-only store.
*   **APIs / Interfaces:**
    - New Header: `X-MCP-Session-Token`
    - Configuration block:
      ```yaml
      security:
        cross_origin:
          enabled: true
          trusted_origins: ["http://localhost:5173", "vscode-webview://*"]
          enforce_session_token: true
      ```
*   **Data Storage/State:** Persistent storage of trusted session tokens in a secure local directory.

## 5. Alternatives Considered
*   **CORS Only**: Relying solely on standard CORS headers. *Rejected* because CORS is a browser-side enforcement and can be bypassed by non-browser clients or misconfigured browsers.
*   **Mutual TLS (mTLS)**: Requiring certificates for all local connections. *Rejected* as it adds significant friction for local development and setup.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a critical component of the "Safe-by-Default" initiative. It closes the "Local-to-Web" vulnerability gap.
*   **Observability:** Every blocked cross-origin request must be logged with the origin URL and requested endpoint for audit purposes.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
