# Design Doc: Zero-Trust Agent Identity Hub
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
In autonomous multi-agent environments, the security perimeter has shifted from the network edge to the identity of the agent itself. As swarms grow in complexity, agents must be able to verify each other's mission-bound authority without constant human-in-the-loop attestation.

The Zero-Trust Agent Identity Hub acts as the authoritative "Identity Mint" for the MCP Any mesh. it provides hardware-attested, task-bound identities that allow disparate agents (Claude Code, OpenClaw, AutoGen) to securely coordinate while maintaining absolute mission-root sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP bound) mesh-resident identity tokens for all connected agents.
    * Maintain a verifiable lineage of agent parentage (Chain of Command).
    * Implement task-bound token lifecycles with automatic revocation upon sub-mission completion.
    * Support SHAQ-compliant reputation scoring integration for dynamic trust adjustments.
* **Non-Goals:**
    * Replacing enterprise human IAM systems (e.g., Okta, AD).
    * Managing model-provider API keys (handled by the Secret Manager).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely delegate a high-trust capability (e.g., database write) to a subagent without exposing host credentials.
* **The Happy Path (Tasks):**
    1. Parent Agent requests a "Specialist Token" from the Identity Hub for a new sub-task.
    2. Hub verifies the Parent's hardware attestation and mission-root signature.
    3. Hub issues a task-bound JWT, cryptographically linked to the mission-root and the Parent's identity.
    4. Specialist Agent presents this token to the Database Adapter.
    5. Database Adapter verifies the token's lineage and task-scope with the Hub.
    6. Access is granted only for the duration and scope of the specific task.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B[TPM/Hardware Attestation]
        B --> C[Identity Hub]
        C --> D{Verify Mission Root}
        D -- Verified --> E[Issue Mesh-Resident Token]
        D -- Failed --> F[Deny & Trigger Alert]
        E --> G[Agent Coordination Bus]
        G --> H[Teammate Verification]
        H --> C
    ```
* **APIs / Interfaces:**
    * `identity.IssueToken(parentToken, taskManifest, attestationData) -> MeshToken`: Mints a new task-bound identity.
    * `identity.VerifyLineage(token) -> LineageReport`: Returns the full parentage chain of an agent.
    * `identity.Revoke(token)`: Immediately invalidates an identity fragment.
* **Data Storage/State:**
    * **Identity Ledger**: An encrypted SQLite store mapping tokens to hardware IDs and mission-roots.
    * **Reputation Cache**: Real-time confidence scores synchronized with the SHAQ provider.

## 5. Alternatives Considered
* **Static API Keys per Agent**: Rejected due to the risk of token exfiltration and lack of task-specific scoping.
* **Centralized Cloud IAM**: Rejected due to latency (the "Attestation Tax") and the requirement for local sovereignty.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All tokens are hardware-bound; any attempt to use a token on an un-attested hardware ID triggers the Kill-Switch (ASKS).
* **Observability:** Integrated with the "Mesh-Resident Lineage Tracker" for real-time visualization of the command hierarchy.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
