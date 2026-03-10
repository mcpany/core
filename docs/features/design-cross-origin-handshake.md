# Design Doc: Cross-Origin Handshake Middleware
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the rise of local AI agents like OpenClaw, Claude Code, and Gemini CLI, a new attack vector has emerged: malicious websites or unauthorized local processes attempting to hijack the agent's MCP tools. As seen in the recent OpenClaw vulnerability, failing to distinguish between a trusted developer environment and an external web origin allows for unauthorized tool execution, potentially leading to data exfiltration or RCE.

MCP Any must act as the "Security Gateway" that enforces a strict identity handshake before any tool calls are processed.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory cryptographic handshake for all incoming agent connections.
    * Verify the "Caller Identity" (e.g., process hash, binary signature, or secure token).
    * Block all requests from unauthorized Origins (web browsers) by default.
    * Provide a mechanism for "First-Use Attestation" (FUA) where users manually approve a new agent binary.
* **Non-Goals:**
    * Implementing a full OS-level sandbox (handled by other components).
    * Managing user authentication for remote connections (handled by MFA/Remote Auth module).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a local LLM agent (e.g., OpenClaw).
* **Primary Goal:** Prevent a malicious website from executing `rm -rf /` via the agent's shell tool.
* **The Happy Path (Tasks):**
    1. User starts MCP Any and then starts OpenClaw.
    2. OpenClaw attempts to connect to MCP Any.
    3. MCP Any intercepts the connection and requests a "Challenge-Response" handshake.
    4. OpenClaw provides its attestation token (generated during installation or a signed payload).
    5. MCP Any verifies the token and the process metadata.
    6. Connection is established; subsequent tool calls are authorized.
    7. A malicious website tries to connect via a browser; MCP Any detects the `Origin` header and lack of attestation, then drops the connection.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant Gateway as MCP Any Gateway
        participant Policy as Policy Engine

        Agent->>Gateway: Connection Request (HTTP/WebSocket)
        Gateway->>Gateway: Check Origin & Headers
        alt Unauthorized Origin
            Gateway-->>Agent: 403 Forbidden
        else Potential Trusted Agent
            Gateway->>Agent: Identity Challenge (Nonce)
            Agent->>Agent: Sign Nonce with Local Private Key
            Agent->>Gateway: Challenge Response + Signed Metadata
            Gateway->>Policy: Validate Attestation
            Policy-->>Gateway: Validated (Scope: [Full])
            Gateway-->>Agent: 101 Switching Protocols / 200 OK
        end
    ```
* **APIs / Interfaces:**
    * `/auth/handshake`: Endpoint for initiating the challenge-response flow.
    * Header: `X-MCP-Any-Attestation`: Stores the signed challenge response.
* **Data Storage/State:**
    * Trusted Agent Registry: A local secure store (SQLite) containing hashes and public keys of approved agent binaries.

## 5. Alternatives Considered
* **Simple Token Auth**: Rejected because tokens can be leaked or stolen from config files.
* **OS-Level IP Filtering**: Rejected because multiple processes (including browsers) share `localhost`.
* **mTLS**: Considered, but might be too complex for most local developer setups; kept as an "Advanced" option.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The handshake ensures that "Possession of the Port" does not equal "Authorization to use Tools."
* **Observability:** Log all failed handshake attempts with source process metadata for audit trails.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
