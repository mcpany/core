# Design Doc: Agent Identity Management (AIM)

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As multi-agent swarms (OpenClaw, CrewAI) become more prevalent, the lack of independent agent identities has become a major security bottleneck. Currently, most agents share the same set of API keys or environment variables, making it impossible to enforce granular "least privilege" or perform accurate auditing. Agent Identity Management (AIM) introduces a way to assign and verify unique, cryptographic identities (SIDs) for every agent interacting with MCP Any.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Establish a protocol for unique agent UID assignment using Ed25519 signatures.
    *   Implement per-agent capability tokens that restrict specific agents to specific tools/resources.
    *   Link all audit logs and telemetry to a verified Agent Identity.
    *   Support agent-level "revocation" of access.
*   **Non-Goals:**
    *   Defining the internal logic of the agents.
    *   Providing a full-blown IAM system for humans (focus is strictly on Agent-to-System and Agent-to-Agent identity).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Engineer.
*   **Primary Goal:** Verify exactly which subagent in a swarm attempted to access a sensitive database tool.
*   **The Happy Path (Tasks):**
    1.  The swarm orchestrator registers a new subagent with MCP Any, receiving a unique `agent_sid` and a cryptographic keypair.
    2.  The subagent includes a signed `X-Agent-Identity` header in its MCP tool calls.
    3.  MCP Any verifies the signature against the registered public key.
    4.  The Policy Engine checks if `agent_sid` has the `db:query` capability.
    5.  The audit log records: `[Timestamp] Agent: research-subagent-01 (SID: ag_7f2...) called tool: db_query`.

## 4. Design & Architecture
*   **System Flow:**
    - **Registration**: Agents or orchestrators register via a new `IdentityAPI`.
    - **Authentication**: Tool calls are authenticated via signed JWTs or custom headers containing the agent's identity and a timestamped challenge signature.
    - **Authorization**: The existing Policy Firewall is updated to accept `agent_sid` as a principal in Rego/CEL rules.
*   **APIs / Interfaces:**
    - `POST /api/v1/identity/register`: Returns a SID and secret/key.
    - `GET /api/v1/identity/agents`: List active agent identities.
*   **Data Storage/State:** Agent identities and public keys stored in the secure internal SQLite database.

## 5. Alternatives Considered
*   **Shared API Keys with Metadata**: Using standard API keys and asking agents to "self-report" their name in headers. *Rejected* as it is not cryptographically verifiable and easily spoofed.
*   **OAuth2 for Agents**: Using standard OAuth2 flows. *Rejected* as it adds too much latency and complexity for local, ephemeral subagent swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the "Identity" pillar of Zero Trust. It ensures that "Who" is calling a tool is verified before "What" they can do is evaluated.
*   **Observability:** The Agent Identity Dashboard will provide real-time visibility into the "Active Agent Fleet."

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
