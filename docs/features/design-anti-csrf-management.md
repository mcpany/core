# Design Doc: Anti-CSRF Management API Protection
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) highlighted a critical vulnerability in local AI agents: the ability for malicious websites to modify an agent's local configuration via Cross-Site Request Forgery (CSRF). Since many agents (including MCP Any) expose a local management API for configuration, an attacker can trick a user's browser into sending unauthorized POST/PUT requests to `localhost` to disable security features (e.g., HITL) and achieve RCE.

This design document outlines the implementation of a hardened Management API for MCP Any that prevents CSRF and unauthorized cross-origin tampering.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement strict Origin and Referer validation for all state-changing management requests.
    * Introduce stateful Anti-CSRF tokens for the management dashboard.
    * Support "Configuration Locking" where security-critical fields cannot be modified via the API without out-of-band (OOB) verification.
* **Non-Goals:**
    * Implementing a full User Authentication system (this is a local-first gateway).
    * Protecting against physical access to the host machine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local AI Agent Developer
* **Primary Goal:** Securely configure the MCP Any gateway without exposing it to web-based attacks.
* **The Happy Path (Tasks):**
    1. User opens the MCP Any Management Dashboard (e.g., `http://localhost:3000`).
    2. The server issues a unique, session-bound Anti-CSRF token.
    3. User modifies the Policy Firewall rules.
    4. The Dashboard includes the Anti-CSRF token in the request header.
    5. The MCP Any server verifies the token and the `Origin` header before applying the change.
    6. A malicious site tries to trigger the same change; the request is rejected due to a missing/invalid token and mismatched Origin.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Browser as User Browser
        participant Malicious as Malicious Website
        participant Server as MCP Any Server

        Note over Browser, Server: Legitimate Flow
        Browser->>Server: GET /dashboard
        Server-->>Browser: 200 OK (Set-Cookie: csrf_token=XYZ)
        Browser->>Server: POST /api/config (Header: X-CSRF-Token: XYZ)
        Server->>Server: Verify Token & Origin
        Server-->>Browser: 200 OK

        Note over Malicious, Server: Attack Flow (CSRF)
        Malicious->>Browser: Trigger hidden POST to localhost:3000/api/config
        Browser->>Server: POST /api/config (Missing X-CSRF-Token)
        Server->>Server: Reject (403 Forbidden)
    ```
* **APIs / Interfaces:**
    * `X-CSRF-Token`: Mandatory header for all POST, PUT, DELETE, and PATCH requests.
    * `GET /api/v1/csrf-token`: Endpoint to retrieve a fresh token for SPA clients.
* **Data Storage/State:**
    * Tokens are stored in a short-lived in-memory map on the server, keyed by a session ID.

## 5. Alternatives Considered
* **Requiring a Password:** Rejected for local-first UX, but considered as an optional "Hardened Mode."
* **Unix Domain Pipes only:** Rejected because the UI needs to run in a browser for ease of use.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This feature is a direct implementation of Zero Trust principles for the management plane.
* **Observability:** All rejected CSRF attempts must be logged with the `Referer` and `Origin` headers for audit purposes.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation. Addressing OpenClaw CSRF-to-RCE patterns.
