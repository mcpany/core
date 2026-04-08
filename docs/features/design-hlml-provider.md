# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward full autonomy, the risk of persistent privilege escalation becomes a critical failure point. Standard session tokens are often too coarse or too long-lived for high-risk operations. The HLML Provider aims to solve this by tying specific agent capabilities to a hardware-attested "Mission Lease" that expires automatically upon the completion of a root task.

This system ensures that even if a subagent is compromised, its elevated privileges are strictly time-bound and cryptographically restricted to a verified mission scope, anchored in the host's Trusted Platform Module (TPM).

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases.
    * Automate the revocation of privileges upon mission-root termination.
    * Provide non-repudiable hardware evidence of privilege grants.
    * Integrate with the Mission-Root Continuity Provider (MRCP) for lease persistence.
* **Non-Goals:**
    * Replacing general-purpose OIDC/OAuth2 authentication.
    * Managing hardware drivers or TPM provisioning.
    * Real-time monitoring of agent reasoning (handled by ARI Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Grant a specialist agent `sudo` access for exactly one server-restarts task without manual intervention or persistent risk.
* **The Happy Path (Tasks):**
    1. Parent agent defines a Mission Manifest with a requested `sudo` capability for a specialist subagent.
    2. HLML Provider verifies the parent's hardware-attested mission root.
    3. HLML Provider requests a TPM signature to mint an HLML lease token.
    4. Specialist subagent receives the lease token and invokes the `restart_server` tool via MCP Any.
    5. MCP Any validates the lease token against the active mission ID.
    6. Upon task completion, the parent signals mission-root termination.
    7. HLML Provider updates the hardware monotonic counter, effectively revoking the lease mesh-wide.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>HLML_Provider: Request Task Lease (MissionID, Capability)
        HLML_Provider->>TPM: Sign Lease Manifest (SHA256)
        TPM-->>HLML_Provider: TPM_Signed_Lease
        HLML_Provider-->>ParentAgent: HLMLToken
        ParentAgent->>SpecialistAgent: Delegate(Task, HLMLToken)
        SpecialistAgent->>MCPAny_Gateway: CallTool(HLMLToken, Args)
        MCPAny_Gateway->>HLML_Provider: Validate(HLMLToken)
        HLML_Provider-->>MCPAny_Gateway: OK (Hardware-Verified)
        MCPAny_Gateway->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * `POST /v1/leases/mint`: Request a new hardware-locked lease.
    * `POST /v1/leases/validate`: Verify the hardware signature and lifecycle status of a lease.
    * `DELETE /v1/leases/:id`: Forcefully revoke a lease.
* **Data Storage/State:**
    * Lease manifests are stored in the Shared KV Store (Blackboard) but only valid if matching the TPM monotonic counter state.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they can be "squatted" and reused within their expiration window if not tied to a specific hardware state change.
* **Manual User Approval (HITL):** Rejected as the primary mechanism due to "Approval Fatigue" in high-density swarms, though remains a fallback.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are "Denied by Default." Only explicit mission-root entries can trigger a minting event.
* **Observability:** Every lease grant and revocation is logged to the Mission-Root Cold-Storage (MRCS) Sink.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
