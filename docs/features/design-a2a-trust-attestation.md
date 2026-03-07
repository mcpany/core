# Design Doc: A2A Trust Attestation Middleware
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
In the emerging Agent-to-Agent (A2A) ecosystem, task delegation often crosses framework boundaries (e.g., an OpenClaw agent delegating to a CrewAI specialist). Without a centralized trust mechanism, this delegation is vulnerable to "Agent-in-the-Middle" attacks and credential exfiltration. The "8,000 Exposed Servers" crisis highlights that even simple tool exposure is dangerous. MCP Any must provide a verifiable trust layer that authenticates agent identities and enforces reputation-based access control for delegated tool calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically verify the identity of agents requesting task delegation or tool access via the A2A protocol.
    * Implement a "Trust Score" system that incorporates historical execution success and security compliance.
    * Issue and validate session-bound "Agent-ID" identity tokens.
    * Integrate with the Policy Firewall to enforce "Trust-Aware" access rules.
* **Non-Goals:**
    * Building a global, decentralized PKI for all agents (this is focused on the MCP Any mesh).
    * Enforcing specific agent logic or reasoning.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Swarm Security Architect.
* **Primary Goal:** Ensure that only verified, high-trust agents from the "Compliance" framework can delegate administrative tasks to the "Infrastructure" tools.
* **The Happy Path (Tasks):**
    1. A "Policy Specialist" agent (CrewAI) attempts to delegate a `modify_network_policy` task to an MCP Any tool.
    2. MCP Any's A2A Trust Attestation Middleware intercepts the request and checks for a valid `Agent-ID` token.
    3. The middleware verifies the token's cryptographic signature against the registered public key for the "Compliance" framework.
    4. The middleware retrieves the agent's current "Trust Score" from the internal registry.
    5. The Policy Firewall confirms that agents with a "Trust Score > 80" are allowed to call `modify_network_policy`.
    6. The task is executed and the result is returned to the verified agent.

## 4. Design & Architecture
* **System Flow:**
    * **Handshake**: Agents perform a mTLS or signed-header handshake to establish identity.
    * **Issuance**: MCP Any issues a short-lived `Agent-ID` token (JWT-based) for the current session.
    * **Validation**: The `TrustAttestationMiddleware` validates the token on every A2A message or tool call.
    * **Reputation Loop**: Success/failure and security violations are logged to update the agent's "Trust Score" in real-time.
* **APIs / Interfaces:**
    * `POST /a2a/attest`: Endpoint for agent identity registration and token issuance.
    * `GET /a2a/trust/{agent_id}`: Retrieves the current trust score and metadata for an agent.
* **Data Storage/State:** Agent identities and trust scores are stored in a secure, encrypted table within the `Shared KV Store` (Blackboard).

## 5. Alternatives Considered
* **Implicit Trust (Whitelist-only):** Hardcoding allowed agent IPs. *Rejected* as it is not scalable for dynamic swarms or cloud-based agents.
* **External Identity Providers (OIDC):** Using Auth0 or similar for agents. *Rejected* due to high latency and complexity for local-first agent workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** If an agent's trust score drops below a threshold (e.g., due to a detected injection attempt), all its active sessions are immediately revoked.
* **Observability:** Trust scores and attestation logs are visible in the "Connectivity & Security Dashboard" for real-time monitoring.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
