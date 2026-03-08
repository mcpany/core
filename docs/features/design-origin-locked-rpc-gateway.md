# Design Doc: Origin-Locked RPC Gateway
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
The recent OpenClaw vulnerability (March 2026) demonstrated that local agent gateways are vulnerable to "Origin Hijacking." A malicious website running in a user's browser can send JSON-RPC requests to a local service (like `localhost:50050`) if that service does not strictly validate the `Origin` or `Referer` headers. This allows an attacker to execute arbitrary tools on the user's machine via their own agent. MCP Any must implement strict origin-based locking to prevent this class of cross-origin attacks.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory `AllowedOrigins` whitelist for all HTTP-based MCP endpoints.
    * Automatically reject requests from browser-based origins that are not explicitly trusted.
    * Provide a seamless CLI-based "Trust this Origin" workflow for developers.
    * Enforce `SameSite` and CSRF protection for any browser-integrated dashboard components.
* **Non-Goals:**
    * Replacing standard network-level firewalls.
    * Implementing a full-blown OAuth2 server (keep it focused on simple origin/token validation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a web-based LLM interface (e.g., a local instance of OpenClaw or a specific IDE extension).
* **Primary Goal:** Securely connect the web interface to the local MCP Any server without exposing it to malicious sites.
* **The Happy Path (Tasks):**
    1. User starts MCP Any.
    2. User opens their trusted web-based agent interface (e.g., `http://localhost:3000`).
    3. The interface attempts to connect to MCP Any.
    4. MCP Any rejects the request initially and logs a "Blocked Origin: http://localhost:3000" message.
    5. User runs `mcpany trust add http://localhost:3000`.
    6. Subsequent requests from that origin are allowed.
    7. A malicious site (`http://evil-site.com`) attempts the same call and is permanently blocked.

## 4. Design & Architecture
* **System Flow:**
    `Browser Request -> MCP Any HTTP Listener -> Origin Middleware -> [Whitelist Check] -> [Token Check] -> JSON-RPC Handler`
* **APIs / Interfaces:**
    * `OriginMiddleware`: Intercepts all incoming HTTP requests to check `Origin` and `Referer` headers against the `allowed_origins` configuration.
    * `mcpany trust`: New CLI command group for managing the whitelist.
* **Data Storage/State:**
    * Whitelist is stored in `config.yaml` or a dedicated `security.json` state file.

## 5. Alternatives Considered
* **Relying on CORS alone**: Browser-side CORS is a client-side enforcement mechanism. Malicious actors can bypass it using non-browser clients, but more importantly, "Simple Requests" might still reach the server before being blocked. Server-side validation of the `Origin` header is required for true security.
* **API Keys Only**: While API keys help, they can be stolen if the origin isn't also verified. Combining both provides defense-in-depth.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Fundamental protection against unauthorized tool execution from the web.
* **Observability**: Blocked origin attempts must be logged with high priority and surfaced in the security dashboard.

## 7. Evolutionary Changelog
* **2026-03-07**: Initial Document Creation.
