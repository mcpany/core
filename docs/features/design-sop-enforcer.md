# Design Doc: Same-Origin Policy (SOP) Enforcer for MCP
**Status:** Draft
**Created:** 2026-03-14

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) has demonstrated that the "Local Trust" model—where localhost connections are implicitly trusted—is a critical vulnerability. Malicious websites can use JavaScript to initiate cross-site requests (CSWSH/CSRF) to local agent gateways, stealing authentication tokens and achieving RCE. MCP Any must implement a mandatory Same-Origin Policy (SOP) enforcement layer for all its local listeners.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory validation of `Origin` and `Sec-Fetch-Site` headers for all local HTTP and WebSocket endpoints.
    * Provide a configurable "Trusted Origins" allow-list for local IDEs, CLI tools, and authorized dashboards.
    * Automatically block any request from a non-local or non-allow-listed origin.
    * Log blocked attempts for the Security Dashboard.
* **Non-Goals:**
    * Implementing full OIDC or OAuth2 (covered by other auth modules).
    * Protecting against physical host access.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer
* **Primary Goal:** Ensure that only authorized local tools can communicate with the MCP Any gateway.
* **The Happy Path (Tasks):**
    1. User starts the MCP Any gateway on their local machine.
    2. A malicious website (e.g., `https://malicious-agent-exploit.com`) tries to connect to `ws://localhost:3000`.
    3. The SOP Enforcer intercepts the request, sees the unauthorized origin, and immediately rejects the handshake.
    4. The user's local VS Code extension (allow-listed) connects to the same port and is granted access.

## 4. Design & Architecture
* **System Flow:**
    `HTTP Request/WS Handshake` -> `SOP Enforcer Middleware` -> `Origin Check` -> `Allow-list Validation` -> `Route Handler`
* **APIs / Interfaces:**
    * `SOPMiddleware`: Intercepts all incoming requests.
    * `OriginRegistry`: Manages the list of trusted origins (e.g., `vscode-webview://`, `http://localhost:5173`).
* **Data Storage/State:**
    * Trusted origins are loaded from `config.yaml` or managed via the Enterprise Policy Sync Engine.

## 5. Alternatives Considered
* **Requiring a custom Header (e.g., `X-MCP-Token`)**: While effective, browsers do not send custom headers during standard WebSocket handshakes without complex workarounds.
* **IP-based Filtering**: Ineffective as most browser-based exploits originate from `127.0.0.1` via the browser.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All local transport is treated as untrusted until the origin is verified.
* **Observability:** Blocked origins are reported to the "Origin Violation Real-time Monitor" in the UI.

## 7. Evolutionary Changelog
* **2026-03-14:** Initial Document Creation.
