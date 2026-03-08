# Design Doc: Origin-Bound Request Validation
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Recent security vulnerabilities in the AI agent ecosystem (specifically the OpenClaw "localhost hijacking" incident) have demonstrated that local services are not inherently secure. Malicious websites running in a user's browser can attempt to connect to local WebSocket or HTTP ports, effectively bypassing traditional network-level firewalls. MCP Any, as a gateway for sensitive tool access and data, must protect its interfaces against these cross-site/browser-initiated attacks.

This design doc outlines a middleware layer for MCP Any that enforces strict, cryptographic origin validation for all incoming requests, even those originating from `localhost`.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent unauthorized browser-based connections to the MCP Any gateway.
    * Enforce strict `Origin` and `Host` header validation.
    * Implement a cryptographic "Local App Token" system for trusted local applications.
    * Support dynamic allow-listing of trusted origins (e.g., local IDEs, specific web-based agent consoles).
* **Non-Goals:**
    * Replacing network-level firewalls for remote access (covered by the "Safe-by-Default Hardening" design).
    * Providing full identity management (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a local LLM swarm with browser-based monitoring tools.
* **Primary Goal:** Ensure that only trusted local applications and authorized browser tabs can interact with the MCP Any gateway, while blocking malicious websites.
* **The Happy Path (Tasks):**
    1. User starts MCP Any server.
    2. MCP Any generates a transient "Local App Token" and stores it in a secure local file (e.g., `~/.mcpany/session_token`).
    3. A trusted local IDE (like Cursor or VS Code) reads the token and includes it in the `Authorization` header of its connection request.
    4. The Origin-Bound Middleware validates the token and the `Host` header.
    5. Connection is established. Malicious website at `evil-attacker.com` attempts a WebSocket connection, fails the origin check, and is rejected.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Client as Local App / Browser
        participant MW as Origin-Bound Middleware
        participant Gateway as MCP Any Gateway

        Client->>MW: Request (Header: Origin, Host, Auth Token)
        Note over MW: 1. Check Host == localhost/127.0.0.1
        Note over MW: 2. If browser, check Origin against Allow-list
        Note over MW: 3. Verify Auth Token against session_token
        alt Valid
            MW->>Gateway: Forward Request
            Gateway-->>Client: Success
        else Invalid
            MW-->>Client: 403 Forbidden / Connection Closed
        end
    ```
* **APIs / Interfaces:**
    * `X-MCP-App-Token`: Custom header for non-browser local applications.
    * `config.yaml` addition:
      ```yaml
      security:
        origin_validation:
          enabled: true
          allowed_origins:
            - "http://localhost:3000"
            - "vscode-webview://*"
          require_app_token: true
      ```
* **Data Storage/State:**
    * Session tokens stored in memory and synchronized to a restricted-permission file (`0600`) for local discovery.

## 5. Alternatives Considered
* **CSRF Tokens only**: Rejected because they don't prevent all types of WebSocket hijacking in misconfigured browsers.
* **Mutual TLS (mTLS)**: Considered for inter-agent comms, but too complex for simple local browser/app integrations. Token-based auth is more ergonomic for this specific use case.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Even local traffic is untrusted until the app token or origin is verified.
* **Observability:** Log all blocked origin attempts with source IP and headers to help users debug misconfigurations.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
