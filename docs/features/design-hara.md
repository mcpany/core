# Design Doc: Hardware-Attested Role Attribution (HARA)
**Status:** Draft
**Created:** 2026-07-18

## 1. Context and Scope
As Agent Teams move from linear sessions to horizontal teammate meshes, identity-based security is failing. A specialist agent (e.g., "Documentation Agent") often inherits the same broad permissions as the "Lead Architect" because they share the same session token. This allows for "Role Hijacking," where a low-trust agent performs high-trust actions. HARA provides kernel-level, TPM-signed role attribution to enforce strict capability boundaries based on the agent's assigned mission role.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested "Role Tokens" during agent initialization.
    * Enforce role-based capability gating at the tool-proxy layer.
    * Provide non-repudiable audit logs of tool calls linked to specific hardware-bound roles.
    * Neutralize "Role Hijacking" by blocking tool calls that diverge from the attested role manifest.
* **Non-Goals:**
    * Replacing existing A2A identity protocols (HARA complements them).
    * Managing the mission-root lifecycle (handled by MRCP).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Mesh Orchestrator
* **Primary Goal:** Ensure that the "Specialist Refactoring Agent" can only modify code files and cannot access environment secrets or call external shell tools.
* **The Happy Path (Tasks):**
    1. Orchestrator defines a "Code Specialist" role with a specific capability manifest.
    2. HARA Provider issues a TPM-signed Role Token to Agent A.
    3. Agent A attempts to call `get_env_secret`.
    4. HARA middleware intercepts the call and verifies the Role Token against the manifest.
    5. The call is interdicted because the role lacks the `secret:read` capability.

## 4. Design & Architecture
* **System Flow:**
  [Agent Init] -> [HARA Provider (TPM Sign)] -> [Role Token]
  [Tool Call] -> [HARA Middleware (Verify Role + Manifest)] -> [Tool Execution]
* **APIs / Interfaces:**
    * `mcpany.hara.v1.RoleMint`
    * `mcpany.hara.v1.RoleGate`
* **Data Storage/State:**
    * Hardware-protected role manifests; session-bound role registry.

## 5. Alternatives Considered
* **Pure JWT-based Roles**: Rejected as they are susceptible to token-stealing and reuse across processes.
* **Static Permission Lists**: Rejected as they cannot adapt to dynamic horizontal mesh formation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Roles are immutable once signed and bound to the hardware enclave.
* **Observability:** Role-based access events are streamed to the Identity Lineage Inspector.

## 7. Evolutionary Changelog
* **2026-07-18:** Initial Document Creation.
