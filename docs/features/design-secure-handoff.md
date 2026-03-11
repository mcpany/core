# Design Doc: Secure Handoff Protocol (Orchestrator-Executor)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As agents move from single-process scripts to distributed swarms, the boundary between the "Orchestrator" (the agent making decisions, often in the cloud) and the "Executor" (the tool running on a local machine) has become a primary attack surface. Current handoffs often rely on insecure environment variables or long-lived API keys. The Secure Handoff Protocol (SHP) provides a standardized, cryptographic way to transfer ephemeral context and secrets between a remote Orchestrator and a local MCP Any instance.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize the transfer of "Short-Lived Ephemeral Secrets" between agents.
    * Enable "One-Time Use" capability tokens for specific tool calls.
    * Provide a verifiable audit trail of context inheritance across transport boundaries (Cloud-to-Local).
    * Support "Zero-Knowledge" secrets where the Orchestrator never sees the raw local credentials.
* **Non-Goals:**
    * Replacing long-term identity providers (OIDC/SAML).
    * Managing persistent database credentials (focus is on session-specific ephemeral data).

## 3. Critical User Journey (CUJ)
* **User Persona:** Cloud-based Agent Swarm (e.g., Fetch.ai).
* **Primary Goal:** Safely trigger a local `git push` on a developer's machine without the cloud agent ever seeing the developer's SSH keys or GitHub tokens.
* **The Happy Path (Tasks):**
    1. Remote Orchestrator requests a "Local Task Session" from MCP Any.
    2. MCP Any generates an asymmetric key pair and sends the public key to the Orchestrator.
    3. Orchestrator encrypts the "Task Context" (but not the secrets) and signs it.
    4. MCP Any receives the signed context, validates the Orchestrator's identity, and maps it to a local "Ephemeral Capability Token."
    5. The local tool is executed using the ephemeral token, which MCP Any transparently swaps for the real local secret at the last millisecond.

## 4. Design & Architecture
* **System Flow:**
    `Remote Orchestrator` -> `SHP Handshake` -> `MCP Any Gateway` -> `Local Tool`
    1. **Handshake**: JWE (JSON Web Encryption) based exchange to establish a session key.
    2. **Tokenization**: Context and intent are bound to a UUIDv7 token with a 60-second TTL.
    3. **Secret Injection**: The `Secret Manager` middleware injects credentials into the `Detached Sandbox` only for the duration of the tool call.
* **APIs / Interfaces:**
    * `POST /v1/handoff/init`: Initialize a secure session.
    * `POST /v1/handoff/exec`: Execute a tool call within the session context.
* **Data Storage/State:**
    * Ephemeral session keys are stored in memory only (no persistence).

## 5. Alternatives Considered
* **Environment Variable Passing**: Rejected as insecure and prone to logging leaks.
* **Mutual TLS (mTLS)**: Too complex for ad-hoc agent connections; SHP provides a lighter, JWE-based alternative.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements "Perfect Forward Secrecy" for agent sessions. Even if one session is compromised, others remain secure.
* **Observability**: The `Supply Chain Attestation Viewer` will show the "Lineage of Trust" for every remote-triggered tool call.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
