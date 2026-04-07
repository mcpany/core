# Design Doc: Discovery-Phase Authentication (DPA) Provider
**Status:** Draft
**Created:** 2026-04-07

## 1. Context and Scope
As agent ecosystems move toward autonomous peer-to-peer discovery (A2A), the "Discovery Phase" has become a high-value target for shadow-mapping. Malicious subagents or compromised peers can probe a gateway's capabilities without ever initiating a formal task, allowing them to identify vulnerable tools or sensitive "Agent Cards" for later exploitation.

The Discovery-Phase Authentication (DPA) Provider mandates that any discovery-related request (searching tools, listing capabilities, or fetching agent cards) must be backed by a session-bound, hardware-attested identity token.

## 2. Goals & Non-Goals
* **Goals:**
    * Neutralize "Pre-Flight Shadow Mapping" by unauthenticated peers.
    * Mandate MFA/Hardware attestation for discovery in high-trust missions.
    * Provide a cryptographically bound audit trail of who discovered what and when.
* **Non-Goals:**
    * Authenticating tool *execution* (handled by the Policy Firewall and EPM).
    * Providing discovery for public, zero-trust tools that explicitly opt-out of DPA.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Prevent a guest specialist agent from seeing the full tool manifest of the primary supervisor agent.
* **The Happy Path (Tasks):**
    1. The Guest Agent initiates a `list_tools` request to MCP Any.
    2. The DPA Provider intercepts the request and challenges for a Mission-Bound Identity Token.
    3. The Guest Agent provides a token signed by the supervisor.
    4. DPA validates the signature against the hardware root of trust.
    5. MCP Any returns only the subset of tools authorized for that specific token scope.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Peer Agent->>Gateway: Discovery Request (Search Tools)
        Gateway->>DPA Provider: Authenticate Request
        DPA Provider-->>Peer Agent: Challenge (Nonce + Auth Requirement)
        Peer Agent->>DPA Provider: Auth Response (Signed Token)
        DPA Provider->>TPM/Secure Enclave: Verify Signature
        DPA Provider->>Gateway: Auth Success (Scopes: [DB_ADMIN])
        Gateway->>Registry: Filtered Discovery
        Registry-->>Peer Agent: Capability Manifest
    ```
* **APIs / Interfaces:**
    * `DPAInterceptor`: Middleware that wraps all discovery-related JSON-RPC methods.
    * `IdentityResolver`: Interface for validating tokens against various framework roots (Claude Code, Gemini, OpenClaw).
* **Data Storage/State:**
    * Transient session store for active discovery handshakes.
    * Persistent audit log for capability discovery events.

## 5. Alternatives Considered
* **Implicit Trust on Loopback:** Rejected due to CVE-2026-25253; local ports are no longer considered a secure boundary.
* **Static API Keys:** Rejected in favor of dynamic, hardware-bound tokens that can be tied to specific mission lifetimes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Discovery is "invisible-by-default" until authentication is completed.
* **Observability:** Failed discovery attempts are flagged as "Capability Probing" events in the Security Hub.

## 7. Evolutionary Changelog
* **2026-04-07:** Initial Document Creation.
