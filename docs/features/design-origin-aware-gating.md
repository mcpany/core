# Design Doc: Origin-Aware Request Gating (OARG)

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
The March 2026 OpenClaw vulnerability revealed that binding to `localhost` is not enough to secure AI agents. Malicious websites running in a user's browser can make cross-origin requests to local services, effectively hijacking the agent's tools. MCP Any needs a mechanism to distinguish between authorized local applications (like a terminal-based Gemini CLI) and unauthorized browser-based origins.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement strict `Origin` and `Referer` header validation for all incoming requests.
    *   Support a "Trusted App" registry (e.g., specific IDEs, CLI tools).
    *   Introduce a CSRF-like challenge for first-time connections from new origins.
    *   Provide a "Local Browser Approval" flow for legitimate web-based tools.
*   **Non-Goals:**
    *   Replacing traditional Auth (this is an additional layer of defense).
    *   Blocking all cross-origin requests (some legitimate web-apps might need access).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using Claude Code and a local browser.
*   **Primary Goal:** Prevent a malicious site in the browser from calling MCP tools via MCP Any.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any.
    2.  Claude Code (CLI) connects; MCP Any verifies it's a local process and grants access.
    3.  User visits a malicious site; the site tries to `fetch('http://localhost:8080/execute')`.
    4.  MCP Any detects the untrusted `Origin` header and rejects the request with 403 Forbidden.
    5.  User visits a trusted local dashboard; MCP Any prompts the user in the terminal to "Authorize this Origin?".

## 4. Design & Architecture
*   **System Flow:**
    - **Middleware Layer**: An `OriginGuard` middleware sits at the front of the HTTP/WebSocket pipeline.
    - **Origin Validation**: Checks `Origin` and `Host` headers. If they don't match or the origin isn't in the `trusted_origins` list, it triggers a challenge.
    - **Challenge-Response**: Uses a short-lived, out-of-band token (displayed in the terminal or a system notification) that the client must provide to be "paired".
*   **APIs / Interfaces:**
    - `POST /auth/pair`: Endpoint for submitting a pairing token.
    - `GET /auth/origins`: Admin endpoint to view/manage trusted origins.
*   **Data Storage/State:** `trusted_origins.yaml` in the configuration directory.

## 5. Alternatives Considered
*   **CORS Only**: Relying on standard CORS headers. *Rejected* because malicious sites can still send "simple" requests or use DNS rebinding to bypass some CORS protections.
*   **Mutual TLS**: Requiring all clients to have a cert. *Rejected* as too high friction for simple CLI tools.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This mitigates "Agent Hijacking" and is a critical component of the Safe-by-Default strategy.
*   **Observability:** Blocked origin attempts should be logged as high-severity security events.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
