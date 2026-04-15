# Design Doc: Ephemeral Registry Hook (ERH) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of autonomous agent swarms, the tool discovery phase has become a primary attack vector. The emergence of **Registry Squatting**—where malicious subagents inject dormant or "shadow" tools into the global discovery bus—demands a transition from persistent registries to a "Just-in-Time" discovery model.

The Ephemeral Registry Hook (ERH) Provider secures the discovery phase by issuing session-locked, single-use discovery tokens. This ensures that a tool's capabilities are only visible and accessible for a specific, parent-authorized discovery event, neutralizing the risk of unauthorized capability mapping.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Just-in-Time" discovery token issuance service.
    * Mandate session-locked, single-use tokens for all tool discovery requests.
    * Neutralize Registry Squatting by ensuring tokens expire immediately after capability mapping.
    * Integrate with hardware-attested (TPM) session identities to bind discovery events.
* **Non-Goals:**
    * Managing tool execution permissions (handled by Policy Firewall).
    * Providing long-term tool metadata storage (handled by PNTD).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Swarm Supervisor
* **Primary Goal:** Securely discover tools for a specialized sub-mission without exposing the entire tool bus to the subagent.
* **The Happy Path (Tasks):**
    1. Supervisor Agent initiates a tool discovery event for a specific Mission Root.
    2. ERH Provider generates an Ephemeral Discovery Token, cryptographically bound to the Mission Root and the current session.
    3. The Subagent uses the token to query the tool registry.
    4. The registry returns only the tools authorized for that specific mission branch.
    5. The token is immediately invalidated upon successful capability mapping.
    6. Any subsequent attempt to reuse the token or probe the registry without a new token is blocked.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Supervisor Request] --> B[ERH Provider]
        B --> C[Token Mint (Session-Bound)]
        C --> D[Ephemeral Discovery Token]
        D --> E[Tool Registry]
        E --> F[Authorized Capability Map]
        F --> G[Token Invalidation]
    ```
* **APIs / Interfaces:**
    * `erh.IssueToken(missionToken, scope) -> EphemeralToken`: Issues a single-use discovery token.
    * `erh.ValidateToken(discoveryToken) -> bool`: Verifies token integrity and expiration status.
* **Data Storage/State:**
    * **Ephemeral Nonce Store:** A high-speed, in-memory store for tracking active discovery nonces and their associated mission scopes.

## 5. Alternatives Considered
* **Persistent Allow-Lists:** Rejected as they are difficult to manage in dynamic, multi-agent swarms and do not prevent "Shadow Mapping" of the allowed tools.
* **Static RBAC for Discovery:** Rejected as it lacks the temporal granularity required to prevent "Registry Squatting."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All discovery tokens must be cryptographically bound to a hardware-attested mission identity.
* **Observability:** Discovery latency and token expiration events are logged for auditability.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Introducing Ephemeral Registry Hooks to neutralize Registry Squatting.
