# Design Doc: Agent Identity Provider (AIDP)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents move from single-user local scripts to multi-agent swarms operating across federated networks, the lack of a standardized identity model has become a critical security risk. Tools are currently executed based on the permissions of the *user* or the *server process*, not the specific *agent* initiating the call. This leads to "Anonymous Agent" vulnerabilities where a subagent can exceed its intended scope.

MCP Any needs to solve this by providing a verifiable Agent Identity (AID) that can be attached to every tool call, enabling fine-grained governance and auditability.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue verifiable, short-lived identity tokens to agents.
    * Support "Identity Attestation" (verifying an agent's code/policy matches its identity).
    * Integrate with industry standards like SPIFFE/SVID and OIDC.
    * Enable "Identity-Aware" policy hooks in the Policy Firewall.
* **Non-Goals:**
    * Replacing human user authentication (OIDC/SAML).
    * Managing persistent agent long-term storage (covered by Shared KV Store).

## 3. Critical User Journey (CUJ)
* **User Persona:** Federated Swarm Orchestrator
* **Primary Goal:** Ensure that only "Verified Research Agents" can access the Internal Database tool, even if the request comes through a federated peer node.
* **The Happy Path (Tasks):**
    1. The Research Agent requests an identity token from the local MCP Any AIDP.
    2. MCP Any validates the agent's attestation (e.g., hash of its system prompt and tool definitions).
    3. MCP Any issues a signed JWS (JSON Web Signature) containing the `agent_id`, `org_id`, and `attestation_hash`.
    4. The agent includes this token in the MCP request headers.
    5. The tool-providing node verifies the signature and checks the Policy Firewall to see if `agent_id` has `db:read` permissions.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>AIDP: Request Identity (Attestation Data)
        AIDP->>Policy Engine: Verify Attestation
        Policy Engine-->>AIDP: Valid
        AIDP->>Agent: Issue Signed AID Token
        Agent->>MCP Gateway: Tool Call + AID Token
        MCP Gateway->>Policy Firewall: Validate Token & Scope
        Policy Firewall-->>MCP Gateway: Authorized
        MCP Gateway->>MCP Server: Execute Tool
    ```
* **APIs / Interfaces:**
    * `POST /v1/identity/issue`: Issues a token.
    * `POST /v1/identity/verify`: Verifies a token (used by federated nodes).
* **Data Storage/State:**
    * Short-lived tokens are kept in memory/cache.
    * Agent "Blueprints" (for attestation) are stored in the Configuration Store.

## 5. Alternatives Considered
* **Static API Keys per Agent**: Rejected due to high management overhead and lack of attestation (keys can be stolen/leaked).
* **Direct SPIFFE Integration**: While we use the SPIFFE model, a direct dependency might be too heavy for local-first MCP Any users; AIDP acts as a lightweight bridge.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Tokens are task-scoped and short-lived (default 5 minutes). Revocation is handled via a global CRL (Certificate Revocation List) in federated meshes.
* **Observability**: Every tool call is logged with its `agent_id`, enabling "Agent-Level Auditing."

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
