# Design Doc: Role-Attested Teammate (RAT) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from linear task delegation to complex horizontal meshes (e.g., Claude Code Agent Teams), the risk of horizontal privilege escalation increases. A specialist subagent (e.g., a "Style Refiner") might be coerced or compromised into performing actions that should be restricted to a "Security Auditor" or "Lead Architect."

MCP Any needs to provide a functional identity layer that goes beyond binary "Allow/Deny" permissions. The RAT Provider introduces hardware-attested role tokens that bind an agent's identity to a specific functional role within the mission. This ensures that only agents with the "Auditor" role can access the security-critical shards of the shared teammate mailbox.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM/SEP-signed "Role Tokens" to agents during the sub-mission spawning phase.
    * Enforce role-based access control (RBAC) on teammate mailbox shards.
    * Provide a verifiable audit trail of role-bound task claims.
    * Integrate with existing Mission-Bound Hardware Leases (MBHL).
* **Non-Goals:**
    * Dynamically changing an agent's role mid-session without re-attestation.
    * Defining the specific reasoning logic used by different roles.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure that only the "Security Specialist" subagent can approve code-generating tool calls in a parallel mesh.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent defines a mission manifest with specific roles (e.g., Lead, Auditor, Coder).
    2. The RAT Provider generates hardware-attested role tokens for each spawned subagent.
    3. The "Coder" subagent submits a task proposal to the shared mailbox.
    4. The mailbox shard for "Tool Approval" rejects claims from agents without the "Auditor" role token.
    5. The "Auditor" subagent provides its RAT token, claims the task, and performs the validation.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Parent Agent->>RAT Provider: Request Subagent (Role: Auditor)
        RAT Provider->>Hardware Enclave: Generate Role Signature
        Hardware Enclave-->>RAT Provider: TPM-signed Role Token
        RAT Provider-->>Parent Agent: Role Token + Execution Identity
        Subagent->>Shared Mailbox: Claim Task (Role Token)
        Shared Mailbox->>RAT Provider: Verify Token Integrity
        RAT Provider-->>Shared Mailbox: Authorized
        Shared Mailbox-->>Subagent: Task Data
    ```
* **APIs / Interfaces:**
    * `POST /v1/rat/issue`: Generates a role-attested token for a given Mission ID and Role.
    * `POST /v1/rat/verify`: Validates the hardware signature and mission-root binding of a role token.
* **Data Storage/State:**
    * Role definitions are stored in the hardware-attested Mission Manifest (HAMM).
    * Active role tokens are ephemeral and tied to the session lifecycle.

## 5. Alternatives Considered
* **Identity-Only Tokens:** Rejected because they don't provide functional boundaries within a swarm, leading to "Specialist Overreach."
* **Software-Signed Roles:** Rejected due to the risk of "Identity Spoofing" (CVE-2026-28190) where a compromised subagent could forge its own role metadata.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All role tokens are cryptographically bound to the Mission Root. Tokens are monotonic and non-reusable across missions.
* **Observability:** Every role-bound task claim is logged in the "Mesh-Resident Lineage Tracker," providing a visual audit of which functional roles were involved in a decision.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
