# Design Doc: Caller Attestation Middleware

**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
The "Localhost Hijacking" vulnerability (CVE-2026-25253) in the OpenClaw ecosystem demonstrated that `localhost` is not a sufficient security boundary. Malicious websites can use JavaScript to open WebSocket or HTTP connections to local services, stealing sensitive tokens or executing arbitrary commands.

MCP Any, as a universal gateway for AI agents, must be able to verify the *identity* of the calling process (e.g., an IDE, a specific CLI, or a verified browser extension). The Caller Attestation Middleware will move the system toward a Zero-Trust model where every request is accompanied by a cryptographically signed attestation of the caller's identity.

## 2. Goals & Non-Goals
* **Goals:**
    *   Verify the identity of local clients using OS-level process information (e.g., PID, binary path, signature).
    *   Support OIDC-based attestation for remote or cloud-based clients.
    *   Block any request that does not have a valid, signed attestation token.
    *   Provide a "Developer Mode" that allows manual override with a one-time password (OTP).
* **Non-Goals:**
    *   Implementing a full-blown IAM system (we rely on external identity providers or OS-level trust).
    *   Verifying the integrity of every individual packet (focus is on session/connection identity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an MCP-enabled IDE (e.g., VS Code or Cursor).
* **Primary Goal:** Securely connect the IDE to the local MCP Any gateway without risk of browser-based hijacking.
* **The Happy Path (Tasks):**
    1.  User starts the MCP Any gateway.
    2.  IDE extension connects to the gateway.
    3.  Gateway requests an attestation token.
    4.  IDE extension uses a local helper to sign a challenge with its private key (or OS-provided identity).
    5.  Gateway verifies the signature against a list of approved "Trusted Callers."
    6.  Gateway establishes a secure, attested session.

## 4. Design & Architecture
* **System Flow:**
    `Client` -> `MCP Any (Attestation Middleware)` -> `Policy Engine` -> `Tool Execution`
    1.  **Handshake**: Client initiates connection.
    2.  **Challenge**: Gateway sends a nonce/challenge.
    3.  **Proof**: Client returns an `AttestationToken` containing the signed challenge and identity metadata.
    4.  **Verification**: Middleware uses `identity-verifier` (OS-specific or OIDC) to validate the proof.
* **APIs / Interfaces:**
    *   `GET /v1/auth/challenge`: Obtain a nonce for attestation.
    *   `POST /v1/auth/attest`: Submit the proof and receive a session token.
    *   Header: `X-MCP-Attestation-Token`: The token passed with subsequent tool calls.
* **Data Storage/State:**
    *   `trusted_callers.yaml`: Configuration file listing approved binary paths, public keys, or OIDC subjects.
    *   `sessions.db`: In-memory store (e.g., Redis or SQLite) for active attested sessions.

## 5. Alternatives Considered
*   **Simple API Keys**: Rejected because they are easily stolen (as seen in OpenClaw).
*   **IP Whitelisting**: Rejected because `127.0.0.1` is shared by the browser.
*   **mTLS for Localhost**: Considered, but deemed too complex for standard developer workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is the core of our Zero Trust strategy. It ensures that even if a port is exposed, an attacker cannot interact with tools without a verified identity.
* **Observability:** All attestation attempts (success and failure) must be logged with detailed metadata (Caller PID, Binary Path, Result).

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
