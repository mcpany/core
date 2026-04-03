# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward high-autonomy swarms, the risk of "Privilege Escalation" via autonomous tool discovery has become critical. Static permission models are insufficient for specialist subagents that spawn and terminate dynamically.

The HLML Provider introduces a "Just-in-Time" capability model where tool access is cryptographically bound to a specific mission-root fragment and enforced via hardware (TPM). This ensures that a compromised subagent cannot reuse credentials post-task or escalate authority beyond its pre-declared manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases.
    * Enforce automatic hardware-level revocation upon mission completion.
    * Provide non-repudiable audit trails for all mission-bound tool calls.
* **Non-Goals:**
    * Replacing the primary User Authentication layer (e.g., OAuth/SSO).
    * Managing persistent cloud credentials (handled by NHI Wallets).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent a specialized "Python Executor" agent from accessing the production database after its code-review task is finished.
* **The Happy Path (Tasks):**
    1. Parent agent requests a sub-mission for "Code Review."
    2. HLML Provider issues a TPM-signed lease for `read_file` and `execute_pytest`.
    3. Specialist subagent uses the lease to invoke tools via the MCP Any gateway.
    4. Gateway verifies the lease against the hardware root for every call.
    5. Subagent signals "Task Complete."
    6. HLML Provider broadcasts a hardware-level revocation; the lease becomes cryptographically invalid.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>HLML: Request Mission Lease (Manifest)
        HLML->>TPM: Sign Lease (MissionID, ExpireIn, Scopes)
        TPM-->>HLML: Signed HLML Token
        HLML-->>ParentAgent: Mission-Bound Lease
        ParentAgent->>SubAgent: Delegate with Lease
        SubAgent->>Gateway: ToolCall(Args, Lease)
        Gateway->>HLML: Verify(Lease)
        HLML->>TPM: Validate Signature
        TPM-->>HLML: Valid
        HLML-->>Gateway: Authorized
        Gateway->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * `POST /v1/leases/issue`: Takes a Mission Manifest and returns a signed HLML token.
    * `POST /v1/leases/verify`: Validates a token signature and scope alignment.
* **Data Storage/State:**
    * Leases are stored in the `Mission-Root Registry` (hardware-attested).
    * Active lease IDs are cached in kernel-bound memory for sub-millisecond verification.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they cannot be forcefully revoked at the hardware level if a subagent is compromised before the TTL expires.
* **Static RBAC:** Rejected because it doesn't scale with the dynamic, heterogeneous nature of agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are "Deny-by-Default." Scopes must be explicitly mapped to the mission-root fragment.
* **Observability:** Every issuance and revocation is logged to the `Command Traceability Provider`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
