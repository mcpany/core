# Design Doc: Ephemeral Workspace Trust Middleware
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
With the release of OpenClaw v1.6 and Claude Code's "Staged Trust" model, the industry is moving away from implicit localhost trust toward session-bound, ephemeral tokens. However, this creates a major usability gap for headless agents, CI/CD runners, and cross-session swarms that do not share a persistent desktop environment. MCP Any needs a "Trust Broker" that can ingest these ephemeral tokens and translate them into stable, capability-bound agent permissions.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate ephemeral session tokens from local transports (e.g., OpenClaw's `X-Session-Token`).
    * Implement a "Trust Translation" layer that maps temporary desktop tokens to persistent, cryptographically signed Agent Identities.
    * Provide a "Headless Bridge" that allows authorized remote agents to leverage local trust if they provide a valid parent-attestation.
    * Prevent "Config Smuggling" by validating project-local settings against the ephemeral trust scope.
* **Non-Goals:**
    * Replacing the underlying OS-level security (e.g., Keychain, DPAPI).
    * Providing long-term storage for session tokens (tokens remain ephemeral).

## 3. Critical User Journey (CUJ)
* **User Persona:** Headless Agent Orchestrator
* **Primary Goal:** Run a multi-agent swarm in a Docker container that needs to access tools on the host machine protected by OpenClaw v1.6.
* **The Happy Path (Tasks):**
    1. The user performs a one-time "Bootstrap Attestation" on the host machine.
    2. MCP Any generates a "Bridge Token" linked to the current desktop session.
    3. The Headless Agent provides the Bridge Token to MCP Any's Ephemeral Trust Middleware.
    4. The Middleware validates the token against the local session and grants the agent "Scoped Trust" to execute host-level tools.
    5. When the desktop session ends, the Scoped Trust is automatically revoked.

## 4. Design & Architecture
* **System Flow:**
    `Local Transport` -> `Token Interceptor` -> `Trust Broker` -> `Capability Mapper` -> `Policy Engine`
* **APIs / Interfaces:**
    * `TrustBroker` Interface: `Exchange(ephemeralToken string) (*AgentIdentity, error)`
    * New middleware hook in the transport layer to inject `X-MCP-Agent-Trust` headers.
* **Data Storage/State:**
    * Ephemeral mappings are stored in-memory with a TTL bound to the session duration.

## 5. Alternatives Considered
* **Persistent API Keys**: Rejected because it violates the "Safe-by-Default" principle and re-introduces the risk of long-term credential exfiltration.
* **Manual MFA for every call**: Rejected due to high friction for autonomous agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Bridge Token" must be bound to the specific hardware ID of the host to prevent relay attacks.
* **Observability:** Trust exchange events and revocations are logged to the "Local Security Audit Log."

## 7. Evolutionary Changelog
* **2026-03-20:** Initial Document Creation.
* **2026-03-21:** Added **Adaptive Trust Continuity** section.
    * **Context:** Developers reported "Headless Handoff" failures in OpenClaw v1.6.
    * **Adjustment:** Introduced hardware-bound attestation (TPM/Secure Enclave) to persist trust across session boundaries for verified headless agents.
