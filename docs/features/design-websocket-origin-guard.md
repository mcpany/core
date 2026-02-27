# Design Doc: WebSocket Origin Guard
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent security research (2026-02-26) has identified a critical vulnerability in local AI agent gateways (e.g., OpenClaw) where malicious websites can hijack the agent via WebSocket connections to `localhost`. Because browsers do not enforce Same-Origin Policy (SOP) on WebSockets in the same way they do for AJAX, any website visited by a developer can silently send commands to an agent running on the same machine. MCP Any must implement a robust "Origin Guard" to prevent this class of attack.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate the `Origin` header of every incoming WebSocket connection.
    * Only allow connections from trusted origins (e.g., the MCP Any UI, approved local domains).
    * Implement a "Local Secret" handshake for non-browser clients.
    * Provide clear audit logs for rejected connection attempts.
* **Non-Goals:**
    * Implementing a full Web Application Firewall (WAF).
    * Restricting standard HTTP/REST calls (these are already handled by CORS/Auth).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer / Security-Conscious Agent User.
* **Primary Goal:** Prevent a malicious site (e.g., `evil-attacker.com`) from controlling their local MCP Any instance.
* **The Happy Path (Tasks):**
    1. User starts MCP Any; it generates a temporary "Session Secret."
    2. User opens the MCP Any UI (`localhost:3000`).
    3. The UI connects via WebSocket, sending the `Origin: http://localhost:3000` header and the session secret.
    4. Origin Guard verifies the origin and the secret, allowing the connection.
    5. User visits `evil-attacker.com`. The site attempts a background WebSocket connection to `localhost:50050`.
    6. Origin Guard detects `Origin: http://evil-attacker.com`, rejects the connection, and logs a security alert.

## 4. Design & Architecture
* **System Flow:**
    `WebSocket Request -> Origin Guard Middleware -> Origin Allowlist Check -> Secret Handshake -> Upgrade to WS`
* **APIs / Interfaces:**
    * **Middleware**: `WSOriginMiddleware` in the server's transport layer.
    * **Configuration**: `security.allowed_origins` (list of strings).
* **Data Storage/State:**
    * Trusted origins are loaded from `config.yaml`.
    * Session secrets are stored in memory and rotated on restart.

## 5. Alternatives Considered
* **Disabling WebSockets**: Rejected because real-time log streaming and tool execution updates are core to the MCP Any experience.
* **Relying on Auth Tokens Only**: Rejected because an attacker might steal a token via XSS on a trusted site; Origin validation provides an essential second layer of defense (Defense in Depth).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero Trust component for local execution.
* **Observability:** Security alerts are pushed to the UI and recorded in the audit log.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation to address OpenClaw-style localhost hijacking.
