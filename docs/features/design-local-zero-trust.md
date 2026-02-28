# Design Doc: Local Zero Trust (LZT) & Config Sandbox

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
Recent high-severity exploits in OpenClaw and Claude Code have demonstrated that the "local environment" is no longer a safe zone. Malicious websites can hijack local agent gateways via WebSockets ("Porous Membrane" attack), and malicious repositories can execute arbitrary code via project-level configuration hooks. MCP Any must evolve to treat local connections and configurations with the same "Zero Trust" rigor as remote ones.

## 2. Goals & Non-Goals
*   **Goals:**
    *   **Cryptographic Connection Binding**: Every local connection (WebSocket/Stdio) must be authenticated with a session-specific token.
    *   **Config Sandbox**: Project-level configuration files must be validated against a strict schema and cannot execute external commands without explicit user attestation.
    *   **Origin Validation**: Enforce strict Origin/CORS checks for all WebSocket gateways to prevent browser-based hijacking.
*   **Non-Goals:**
    *   Providing a full OS-level sandbox (e.g., Docker) for every tool (though we may integrate with them).
    *   Replacing user-level OS permissions.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer.
*   **Primary Goal:** Safely use MCP Any with a multi-agent swarm while browsing the web and cloning new repositories.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any; it generates a master session secret.
    2.  Agent (e.g., OpenClaw) connects to MCP Any and provides the session token.
    3.  User visits a malicious website; the site's JS tries to connect to `localhost:3000` but fails because it lacks the session token and the Origin header is rejected.
    4.  User clones a repository with a `.mcpany/config.yaml` containing a `pre-hook: "rm -rf /"`.
    5.  MCP Any detects the untrusted hook, blocks execution, and prompts the user for "Explicit Attestation" before loading the config.

## 4. Design & Architecture
*   **System Flow:**
    - **Token Handshake**: MCP Any issues short-lived "Agent Access Tokens" (AAT) via a secure out-of-band channel (e.g., a file in a protected directory).
    - **CORS/Origin Guard**: A middleware that rejects any WebSocket/HTTP request with an `Origin` header not matching the explicitly allowed list.
    - **Config Attestation Engine**: Project-level configs are hashed and checked against a local "Known Good" database. Any modification or new hook triggers a HITL (Human-in-the-Loop) approval.
*   **APIs / Interfaces:**
    - `POST /auth/token`: Exchange master secret for session AAT.
    - `WS /gateway?token=<AAT>`: Authenticated gateway access.
*   **Data Storage/State:** Local SQLite database for storing attested config hashes and active session tokens.

## 5. Alternatives Considered
*   **IP Whitelisting**: *Rejected* because malicious JS in the browser shares the user's IP (`127.0.0.1`).
*   **Mutual TLS**: *Rejected* due to high setup friction for local developers, though remains an option for "Enterprise Mode."

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The core of this design. Moves from "Trust by Location" to "Trust by Identity & Attestation."
*   **Observability:** Audit logs will record every failed connection attempt and every blocked config hook.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation focusing on LZT and Config Attestation.
