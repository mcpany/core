# Design Doc: Agent Identity Attestation (AIA)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
In a multi-agent ecosystem, tool calls are no longer initiated solely by a human user. Subagents, swarms, and federated nodes frequently act as intermediaries. Without a verified identity for each agent, a compromised subagent could escalate privileges or "spoof" another agent to perform unauthorized actions (OWASP ASI07). AIA provides a cryptographic mechanism for agents to prove their identity to MCP Any.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unique, verifiable identity (Agent ID) for every agent interacting with MCP Any.
    * Support short-lived, rotated attestation tokens.
    * Bind tool permissions to specific Agent IDs via the Policy Firewall.
* **Non-Goals:**
    * Replacing user-level authentication (AIA is for *Agent* identity, not *User* identity).
    * Providing a full PKI (Public Key Infrastructure).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect.
* **Primary Goal:** Ensure that only the "Verified Research Agent" can access the "Internal Knowledge Base" tool.
* **The Happy Path (Tasks):**
    1. The Research Agent starts and generates a key pair.
    2. The Agent requests an Attestation Token from MCP Any by signing a challenge.
    3. MCP Any verifies the signature and issues a JWT-based Agent Identity Token (AIT).
    4. For every subsequent tool call, the Research Agent includes the AIT in the headers.
    5. The Policy Firewall checks the AIT and allows the call if the policy `agent.id == "research-agent"` matches.

## 4. Design & Architecture
* **System Flow:**
    - **Handshake**: Challenge-Response protocol using ECDSA (P-256).
    - **Token Issuance**: MCP Any issues a JWT containing the `agent_id`, `permissions_scope`, and `expiry`.
    - **Enforcement**: The `AIA-Middleware` extracts the token and populates the evaluation context for the Policy Firewall.
* **APIs / Interfaces:**
    - `POST /auth/attest`: Handshake endpoint.
    - `Header: X-Agent-Identity`: Used in all MCP JSON-RPC calls.
* **Data Storage/State:** Agent public keys and active tokens are stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Static API Keys per Agent**: *Rejected* due to risk of leakage and lack of rotation.
* **mTLS (Mutual TLS)**: *Rejected* as it is too heavy for many local/stdio-based agent transports.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are short-lived (e.g., 1 hour). Policy Firewall enforces "least privilege" based on the Agent ID.
* **Observability:** Audit logs will explicitly record which *Agent* performed which tool call.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
