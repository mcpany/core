# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agent swarms are increasingly moving toward autonomous delegation, where a primary agent spawns specialists to handle sub-tasks. Current permission models are often too coarse, leading to "Privilege Escalation" if a subagent is compromised. The disclosure of persistent subagent squatting and the transition to Agent Teams (Claude Code) necessitate a JIT, hardware-bound leasing model.

The HLML Provider implements a system where capabilities are granted as TPM-signed, task-specific leases. These leases are automatically revoked upon mission completion and support **Recursive Lease Inheritance (RLI)**, allowing safe, fractionalized delegation through deep swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, time-bound capability leases for all high-risk agent operations.
    * Support Recursive Lease Inheritance (RLI), allowing parent agents to sub-lease their authority to teammates.
    * Enforce absolute mission-root alignment, ensuring sub-leases cannot exceed the parent's authorized scope or duration.
    * Facilitate sub-millisecond, hardware-attested lease revocation.
* **Non-Goals:**
    * Providing long-term persistent credentials; all HLML leases are mission-bound.
    * Managing human user authentication; HLML focuses on agent-to-agent and agent-to-tool authority.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Sub-delegate a secure database query task from a Generalist agent to a Specialist agent without exposing the primary user's full SQL credentials.
* **The Happy Path (Tasks):**
    1. Primary user attests to a 30-minute mission for a Generalist agent with "Read-Only" DB access.
    2. Generalist agent identifies the need for a complex query and spawns a SQL-Specialist teammate.
    3. Generalist agent requests a "Sub-Lease" from the HLML Provider, fractionalizing its mission duration (e.g., 5 mins) and specific capability (e.g., `db:query:users`).
    4. HLML Provider issues a TPM-signed RLI token to the Specialist agent.
    5. Specialist agent invokes the DB tool through MCP Any.
    6. MCP Any verifies the RLI token against the Generalist's parent lease and mission-root intent.
    7. Once the Specialist finishes or the 5-minute sub-lease expires, the capability is forcefully revoked by the hardware root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[User] -->|TPM Sign| B[Root Mission Lease]
        B --> C[Parent Agent]
        C -->|Request RLI| D[HLML Provider]
        D -->|Fractionalize| E[Sub-Lease Token]
        E --> F[Teammate Agent]
        F --> G[MCP Any Gateway]
        G -->|Validate Lineage| H[Hardware Enclave]
    ```
* **APIs / Interfaces:**
    * `hlml.IssueRootLease(capabilities, duration, missionID) -> RootLeaseToken`
    * `hlml.InheritLease(parentToken, subCapabilities, subDuration) -> SubLeaseToken`
    * `hlml.VerifyLease(token, toolCall) -> boolean`
* **Data Storage/State:**
    * **Lease Lineage Registry:** In-memory, hardware-attested graph of parent-child lease relationships.
    * **TPM Nonce Cache:** Prevents replay attacks for lease verification.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they can be exfiltrated and reused if the signing key is compromised. HLML requires per-lease TPM attestation.
* **Coarse OS-level Sandboxing:** Rejected because it lacks the "Intent-Awareness" required for granular sub-delegation within a single execution session.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RLI ensures that subagents always operate with the minimum necessary authority.
* **Observability:** Integrated with the "Hierarchical Trust Monitor" UI for real-time visualization of the lease inheritance tree.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
