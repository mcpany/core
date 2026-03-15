# Design Doc: A2A Session Persistence Middleware
**Status:** Draft
**Created:** 2026-04-26

## 1. Context and Scope
As agent swarms evolve from short-lived command-line tools to long-running, autonomous reasoning agents, the fragility of authenticated A2A sessions has become a critical failure point. "Session Decay" occurs when ephemeral auth tokens expire during deep reasoning loops, causing subagents to lose access to parent context or delegated tools. The A2A Session Persistence Middleware provides a core security service that manages token refresh and trust persistence across deep agent hierarchies.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a centralized token refresh and session persistence service for A2A communications.
    * Support "Attested Lineage" where trust is maintained through multi-hop delegations (integration with Multi-Hop Trust Relay).
    * Provide sub-millisecond trust validation for cached session state.
    * Ensure session revocation is synchronized with the Resident Integrity Monitor (RIM).
* **Non-Goals:**
    * Replacing the primary A2A Handshake Provider (it acts as a middleware for it).
    * Storing long-term PII or user credentials (only ephemeral session tokens).

## 3. Critical User Journey (CUJ)
* **User Persona:** Long-Running Agent Swarm
* **Primary Goal:** Maintain a secure, authenticated link between a parent orchestrator and a specialized subagent over a 12-hour reasoning session.
* **The Happy Path (Tasks):**
    1. Parent agent performs initial A2A Handshake and receives a session token.
    2. Parent registers the session with the Persistence Middleware.
    3. During a deep reasoning loop, the subagent attempts to call a tool using a delegated (expired) token.
    4. The Persistence Middleware intercepts the failure, uses the parent's "Attested Lineage" to refresh the token via the Handshake Provider.
    5. The subagent's tool call is re-authorized and completed without "Cognitive Stall."

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [Persistence Middleware] -> [A2A Handshake Provider]`
    `[Persistence Middleware] <-> [Multi-Hop Trust Relay] <-> [RIM]`
* **APIs / Interfaces:**
    * `/v1/session/register`: Enrolls a new A2A session for persistence.
    * `/v1/session/refresh`: Manually triggers a token refresh based on lineage.
    * `/v1/session/status`: Returns the real-time health and attestation strength of a session.
* **Data Storage/State:**
    * Sessions are tracked in a high-speed, encrypted in-memory buffer, backed by the Shared KV Store (Blackboard) for crash recovery.

## 5. Alternatives Considered
* **Short-Lived Tokens Only**: Rejected as it leads to frequent failures in deep reasoning chains (Session Decay).
* **Infinite-Life Tokens**: Rejected as a severe security risk (Zero Trust violation).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All session refreshes require a valid "Attested Lineage" signature. If the underlying hardware state (RIM) drifts, all persisted sessions for that lineage are instantly purged.
* **Observability:** Session health and refresh events are visualized in the "A2A Session Persistence Dashboard."

## 7. Evolutionary Changelog
* **2026-04-26:** Initial Document Creation.
