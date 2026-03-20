# Design Doc: Origin-Locked Session Bridge
**Status:** Draft
**Created:** 2026-04-10

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) revealed a critical vulnerability where local AI agents implicitly trust any connection from `localhost`. This allows malicious websites to bridge into the agent's local control plane via WebSockets or HTTP requests from the user's browser. The Origin-Locked Session Bridge (OLSB) hardens this boundary by mandating cryptographically bound origin validation for all local listeners.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory `Origin` and `Sec-Fetch-Site` header validation for all WebSocket and HTTP listeners.
    * Bind every agent session token to its initiating browser or CLI origin.
    * Provide a user-facing "Trust Prompt" when a new origin attempts to connect.
    * Automatically block cross-site requests (e.g., from `malicious.com` to `localhost:3000`) without an explicit user-attested exception.
* **Non-Goals:**
    * Replacing standard TLS/mTLS for remote connections.
    * Managing user authentication (e.g., passwords), as it focuses on origin-based trust.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a local agent (e.g., OpenClaw) while browsing the web.
* **Primary Goal:** Prevent a malicious website from executing commands via the agent's local WebSocket gateway.
* **The Happy Path (Tasks):**
    1. User starts MCP Any and a local agent.
    2. User opens the authorized Agent UI (e.g., `http://localhost:5173`).
    3. MCP Any detects a connection from `localhost:5173`, validates the origin against the allow-list, and grants a session token.
    4. User accidentally visits `malicious-site.com`, which tries to connect to `ws://localhost:3000`.
    5. MCP Any's Origin-Locked Bridge intercepts the request, identifies the unauthorized `malicious-site.com` origin, and immediately terminates the connection.
    6. A security alert is logged in the MCP Any dashboard.

## 4. Design & Architecture
* **System Flow:**
    `Browser/CLI` -> `HTTP/WS Interceptor` -> `Origin Validator` -> `Allow-list Check` -> `Session Binder` -> `Agent Gateway`
* **APIs / Interfaces:**
    * `OriginInterceptor`: Middleware that extracts and validates headers before protocol upgrade.
    * `SessionRegistry`: Maps active tokens to verified origins and hardware-bound identifiers.
* **Data Storage/State:**
    * `trusted_origins.yaml`: Persists user-approved origins.
    * Ephemeral `SessionMap` in memory.

## 5. Alternatives Considered
* **Disabling WebSockets**: Too restrictive, as many modern agent UIs rely on real-time communication.
* **MFA for Every Request**: High friction; OLSB provides a "Sign-in Once per Origin" experience.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Eliminates the "Implicit Local Trust" loophole.
* **Observability**: Real-time monitoring of origin violations in the UI.

## 7. Evolutionary Changelog
* **2026-04-10:** Initial Document Creation to address CVE-2026-25253.
