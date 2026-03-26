# Design Doc: A2A Authenticated Handshake Provider
**Status:** Draft
**Created:** 2026-04-24

## 1. Context and Scope
As the Agent-to-Agent (A2A) ecosystem matures, the risk of "A2A Coercion" and unauthorized capability discovery has increased. Compromised specialist agents can attempt to trick parent agents into exfiltrating secrets or executing high-risk tasks by spoofing task cards. Gemini CLI v0.33.0 has introduced HTTP authentication for A2A communication to address this. MCP Any needs a native Handshake Provider to ensure all inter-agent communication is authenticated and session-bound.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory HTTP-based handshake for all A2A remote agent communication.
    * Bind agent session tokens to cryptographically verified identities.
    * Support "Auth-before-Discovery" for agent capability cards.
* **Non-Goals:**
    * Implementing a new encryption protocol (will leverage TLS/mTLS).
    * Providing a full identity provider (will bridge to existing providers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Securely delegate a file-analysis task from a Parent Agent to a specialized File-System Subagent without risk of session hijacking or coercion.
* **The Happy Path (Tasks):**
    1. Parent Agent initiates a connection to the Subagent via MCP Any's A2A Bridge.
    2. MCP Any intercepts the request and challenges the Parent Agent for an authentication token (Handshake).
    3. Parent Agent provides a cryptographically signed token.
    4. MCP Any validates the token against the Mission Intent and establishes a session-bound secure channel.
    5. Subagent's capability card is revealed only after the handshake is verified.
    6. Task delegation proceeds over the authenticated channel.

## 4. Design & Architecture
* **System Flow:**
    `[Parent Agent] -> [A2A Bridge (Auth Interceptor)] -> [Handshake Provider] -> [Subagent]`
* **APIs / Interfaces:**
    * `POST /a2a/handshake`: Endpoint for initiating the authenticated handshake.
    * `X-A2A-Auth-Token`: Header for transmitting the session-bound authentication token.
* **Data Storage/State:**
    * Session state is stored in the authenticated session registry, mapping tokens to mission IDs and agent identities.

## 5. Alternatives Considered
* **Implicit Trust (Deprecated):** Rejected due to the "A2A Coercion" risk.
* **Static API Keys:** Rejected as they are prone to leakage and do not provide mission-bound isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enforces identity-bound access control for all inter-agent messages. Neutralizes "Shadow Agent" discovery by masking capabilities until authenticated.
* **Observability:** Logs all handshake attempts (success/failure) and maps them to the Mission Audit Trail.

## 7. Evolutionary Changelog
* **2026-04-25:** Added "Trust Persistence" section to address "Session Decay" in long-running reasoning chains. Introduced token refresh mechanisms linked to the Mission Intent.
* **2026-04-24:** Initial Document Creation based on Gemini CLI v0.33.0 A2A security updates.
