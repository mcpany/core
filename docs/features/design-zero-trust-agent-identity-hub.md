# Design Doc: Zero-Trust Agent Identity Hub
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As AI agent ecosystems evolve from single LLM calls to complex, autonomous service meshes, the number of Non-Human Identities (NHI) is exploding. Traditional IAM systems are designed for human lifecycles and are too slow for the millisecond-scale spawning and termination of specialized subagents. This leads to "Identity Fragmentation" and "Credential Squatting," where stale agent tokens are hijacked by malicious actors.

The Zero-Trust Agent Identity Hub acts as the authoritative local "Identity Mint" for all connected agents. It issues hardware-attested, session-bound tokens that allow disparate agents (Claude, OpenClaw, AutoGen) to securely verify each other's lineage and mission-bound authority within the mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) mesh-resident identity tokens.
    * Enforce mission-root lineage for all issued identities.
    * Automate task-bound identity revocation to prevent squatting.
    * Provide a standardized identity verification interface for heterogeneous frameworks.
* **Non-Goals:**
    * Replacing enterprise human IAM (e.g., Okta, Azure AD).
    * Storing PII of human users (only agent metadata).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely delegate a high-trust database task to a specialist subagent without exposing persistent host credentials.
* **The Happy Path (Tasks):**
    1. The primary agent requests a "Specialist Identity" from the Identity Hub, providing the mission-root token.
    2. The Identity Hub verifies the mission scope and issues a hardware-attested, task-bound token to the subagent.
    3. The subagent presents this token to the Database Tool Adapter.
    4. The Database Tool Adapter verifies the token's lineage and hardware signature with the Identity Hub.
    5. Upon task completion, the Identity Hub automatically revokes the token, neutralizing any further access attempts.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Parent->>Identity Hub: Request Sub-Identity (Mission: X)
        Identity Hub->>TPM: Sign Identity Fragment
        TPM-->>Identity Hub: Attested Token
        Identity Hub-->>Parent: Task-Bound Identity
        Parent->>Subagent: Spawn with Identity
        Subagent->>Gateway: Tool Call (with Token)
        Gateway->>Identity Hub: Verify Lineage
        Identity Hub-->>Gateway: OK (Validated)
    ```
* **APIs / Interfaces:**
    * `POST /v1/identity/mint`: Issues a new task-bound agent identity.
    * `GET /v1/identity/verify`: Validates an identity token and returns its lineage.
    * `DELETE /v1/identity/revoke`: Forcefully terminates an agent session.
* **Data Storage/State:**
    * Identity metadata is stored in a RAM-backed, encrypted session store.
    * Roots of trust are pinned to the local TPM.

## 5. Alternatives Considered
* **Short-lived JWTs without Attestation:** Rejected because it allows for token cloning and side-channel hijacking.
* **Central Cloud IAM:** Rejected due to coordination latency and local sovereignty requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Enforces "Identity-before-Tooling," ensuring no agent can execute code without a verified mission-bound lineage.
* **Observability:** Logs all identity minting and verification events to the Audit Log for forensic analysis of swarm behavior.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
