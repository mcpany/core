# Design Doc: Cross-Origin Agent Protection (COAP)

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
The March 2026 OpenClaw hijacking vulnerability demonstrated that agentic gateways are highly susceptible to "Same-Origin" attacks from malicious websites running in a developer's browser. If an agent (like OpenClaw or Claude Code) is running a local server without strict origin verification, a website can issue commands to it on behalf of the user. MCP Any must implement a cryptographic handshake to ensure that every tool request originates from a trusted agent process and not a web-based attacker.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce cryptographic signing for all tool execution requests.
    *   Provide a "Local Handshake" protocol for agents to authenticate with MCP Any.
    *   Integrate with the "Safe-by-Default" hardening to block all unsigned/unverified requests by default.
    *   Support agent-specific identity tokens.
*   **Non-Goals:**
    *   Replacing standard mTLS (COAP is specifically for the Local-Agent-to-Gateway boundary).
    *   Securing the communication between MCP Any and upstream MCP servers (that is handled by the Provenance/Attestation layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using OpenClaw with MCP Any.
*   **Primary Goal:** Execute a local tool securely without risk of browser-based hijacking.
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any with COAP enabled.
    2.  User starts OpenClaw. OpenClaw performs a one-time "Identity Exchange" with MCP Any (using a local socket or signed file).
    3.  OpenClaw receives a `Session-Origin-Token`.
    4.  Every tool call from OpenClaw includes this token in the header.
    5.  MCP Any verifies the token against the agent's process ID and identity before executing the tool.
    6.  A malicious website attempts to call the tool via `fetch()` and fails because it lacks the `Session-Origin-Token`.

## 4. Design & Architecture
*   **System Flow:**
    - **Identity Exchange**: On startup, MCP Any creates a temporary, short-lived "Registration Secret" in a restricted local directory.
    - **Agent Registration**: Agents read this secret and exchange it for a long-lived `AgentIdentity` and a session-specific `OriginToken`.
    - **Request Interception**: The COAP Middleware intercepts every incoming request. It checks for the `X-MCP-Agent-Signature` header.
*   **APIs / Interfaces:**
    - `POST /v1/auth/register-agent`: Exchange registration secret for identity.
    - `X-MCP-Agent-Signature`: Header containing `HMAC(Payload, AgentKey)`.
*   **Data Storage/State:** In-memory store for active `OriginTokens`, mapped to agent metadata (binary path, PID).

## 5. Alternatives Considered
*   **CORS (Cross-Origin Resource Sharing)**: Using standard CORS headers. *Rejected* because CORS is easily bypassed by tools outside the browser and doesn't provide strong identity.
*   **mTLS for Localhost**: Requiring client certificates for every agent. *Rejected* as it is too much friction for local developer workflows.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** COAP ensures that the "Agent Identity" is verified before any capabilities are granted.
*   **Observability:** The UI will display a "Verified Agents" list and a real-time log of blocked cross-origin attempts.

## 7. Evolutionary Changelog
*   **2026-03-07:** Initial Document Creation.
