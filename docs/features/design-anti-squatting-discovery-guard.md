# Design Doc: Anti-Squatting Discovery Guard
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
As agents become more autonomous and reliant on local tool discovery, a new attack vector has emerged: "Registry Squatting." Malicious processes on the host can register themselves as MCP servers on local loopback ports (`127.0.0.1`) just before an agent initiates discovery. This allows them to shadow legitimate tools and intercept sensitive data or execute unauthorized commands. The Anti-Squatting Discovery Guard (ASDG) aims to neutralize this by enforcing a cryptographically bound handshake for all tool registrations.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate a session-bound handshake for all local MCP server registrations.
    * Use hardware-attested (TPM) tokens to verify the identity of the registering process.
    * Provide a real-time monitor for blocked "Shadow Registration" attempts.
* **Non-Goals:**
    * Securing remote MCP servers (this is handled by mTLS and A2A auth).
    * Preventing all local process communication (focus is on MCP tool registration).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Start a local development environment where tool discovery is guaranteed to be untampered.
* **The Happy Path (Tasks):**
    1. User starts the MCP Any gateway.
    2. ASDG generates a unique, TPM-signed "Registration Secret" for the current session.
    3. Legitimate local MCP servers (e.g., a local DB tool) are provided with this secret via an environment variable.
    4. When a tool attempts to register, it must provide the secret.
    5. ASDG validates the secret and the process ID (PID) of the tool.
    6. A malicious "Squatter" process attempts to register without the secret and is immediately blocked.

## 4. Design & Architecture
* **System Flow:**
    * `Tool Registration Request` -> `Handshake Validator` -> `TPM Secret Verification` -> `Registry Commit`
* **APIs / Interfaces:**
    * `POST /api/v1/registry/register`: Requires `x-mcp-any-session-token`.
* **Data Storage/State:**
    * Active session tokens are stored in kernel-bound memory.

## 5. Alternatives Considered
* **Port Pinning:** Rejected because attackers can still use different ports if the agent is configured for auto-discovery.
* **PID Whitelisting:** Too fragile for dynamic development environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Registration Secret" is rotated on every gateway restart.
* **Observability:** Failed registration attempts are logged with high-severity "Potential Attack" flags.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
