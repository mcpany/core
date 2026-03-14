# Design Doc: Origin-Locked Session Bridge
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) proved that "Local Trust" is insufficient when browsers can bridge the network gap. Malicious websites can use cross-site WebSocket hijacking (CSWSH) to interact with local agent gateways. MCP Any requires an **Origin-Locked Session Bridge** that cryptographically binds every agent session to its initiating origin.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandatory `Origin` and `Sec-Fetch-Site` header validation for all local listeners.
    * Bind session tokens to a specific origin (e.g., a specific browser extension ID or CLI binary hash).
    * Implement "Session Pinning" where a token cannot be reused by a different origin, even on `localhost`.
    * Provide a UI for users to approve and "pin" specific origins.
* **Non-Goals:**
    * Replacing standard API key authentication (this is an additional layer of defense).
    * Managing network-level firewall rules (focused on application-level origin enforcement).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer with an agent-connected browser extension and a malicious tab open in the same browser.
* **Primary Goal:** Prevent the malicious tab from using the agent's WebSocket connection to execute local commands.
* **The Happy Path (Tasks):**
    1. User opens the MCP Any UI and enables "Strict Origin Binding".
    2. The browser extension initiates a connection. MCP Any prompts the user to approve the extension's Origin ID.
    3. The session token is cryptographically bound to that Origin ID.
    4. A malicious website in another tab attempts to connect to `ws://localhost:50050`.
    5. MCP Any detects a mismatch between the token's bound origin and the request's `Origin` header.
    6. The connection is rejected, and an alert is shown in the MCP Any dashboard.

## 4. Design & Architecture
* **System Flow:**
    `Client Request` -> `Header Validator` -> `Session Identity Mapper` -> `Origin Match Check` -> `Protocol Handler`
* **APIs / Interfaces:**
    * `BindTokenToOrigin(token string, origin string)`
    * `VerifyRequestOrigin(req Request) (bool)`
* **Data Storage/State:**
    * Session-to-Origin mapping stored in the secure, internal SQLite store.

## 5. Alternatives Considered
* **Disabling WebSockets**: Rejected; WebSockets are essential for the real-time agent experience.
* **IP-based Filtering**: Rejected; all requests from the same machine share the same `localhost` IP.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements the principle of "Verify Everything," treating the browser as a potentially hostile intermediary.
* **Observability**: Real-time monitoring of origin violations in the "Security Hub" dashboard.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
