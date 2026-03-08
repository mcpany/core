# Design Doc: Browser-Origin Lockdown (Cross-Origin Protection)

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
The "Localhost Hijack" vulnerability (as seen in OpenClaw) demonstrates that local agent gateways are vulnerable to browser-based attacks. A malicious website can make requests to `localhost:50050` from a user's browser, potentially executing tools or exfiltrating data if no proper origin verification is in place. MCP Any must implement strict protection against unauthorized cross-origin requests to ensure that only trusted local applications or authorized remote clients can interact with the gateway.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement strict CORS (Cross-Origin Resource Sharing) policies for the HTTP/WebSocket gateway.
    *   Enforce origin verification for WebSocket connections.
    *   Introduce a "Connection Pinning" mechanism where the first connection from a trusted app (like Claude Desktop) can be pinned and others rejected.
    *   Provide a way to whitelist specific browser origins (e.g., a trusted local web UI).
*   **Non-Goals:**
    *   Replacing the existing API Key authentication (this is an additional layer).
    *   Providing a general-purpose proxy (the protection is for the MCP Any gateway itself).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using MCP Any with Claude Desktop and a local web dashboard.
*   **Primary Goal:** Prevent a malicious site in Chrome from calling `tools/call` on the local MCP Any instance.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any.
    2.  Claude Desktop (trusted local app) connects via Stdio or local HTTP. It works normally.
    3.  User opens a malicious site in a browser tab.
    4.  The site attempts a `fetch('http://localhost:50050/tools/call', ...)` or a WebSocket connection.
    5.  MCP Any detects an untrusted `Origin` header and rejects the request with `403 Forbidden`.
    6.  The user is notified (via logs or UI) of a blocked cross-origin attempt.

## 4. Design & Architecture
*   **System Flow:**
    - **Origin Validation Middleware**: Every incoming HTTP and WebSocket request passes through an origin check.
    - **CORS Policy**: The `Access-Control-Allow-Origin` header is set to a strict whitelist (defaulting to none for HTTP).
    - **WebSocket Origin Check**: During the WS handshake, the `Origin` header is compared against the allowed list.
*   **APIs / Interfaces:**
    - New Config fields in `security` block:
      ```yaml
      security:
        allowedOrigins: ["http://localhost:3000", "vscode-webview://*"]
        enforceOriginCheck: true
      ```
*   **Data Storage/State:** Whitelist of origins is loaded from configuration.

## 5. Alternatives Considered
*   **Token-only Auth**: Relying solely on API keys. *Rejected* because browsers often allow "no-cors" requests or CSRF-like attacks if the API key is stored in a way the site can access (though harder, origin locking is a safer "belt and suspenders" approach).
*   **Randomized Ports**: Making it harder for sites to find the gateway. *Rejected* as it's "security by obscurity" and breaks standard integrations.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a critical component of the "Safe-by-Default" pillar. It mitigates a whole class of browser-to-localhost attacks.
*   **Observability:** Blocked origins should be logged with a high severity level to alert the user of potential attacks.

## 7. Evolutionary Changelog
*   **2026-03-07:** Initial Document Creation.
