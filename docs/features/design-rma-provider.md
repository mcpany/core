# Design Doc: Recursive Mission Attestation (RMA) Provider
**Status:** Draft
**Created:** 2026-06-07

## 1. Context and Scope
As agent swarms become more complex and multi-layered, the risk of "Intent Hijacking"--where a subagent or tool call diverges from the original user mission--increases significantly. Current session-bound tokens only verify the *identity* of the caller, not the *alignment* of the specific sub-task with the root mission.

MCP Any needs a mechanism to issue and verify hardware-attested, recursive mission tokens. This ensures that every subagent spawn and tool call carries a "Mission Receipt" that can be traced back to the user's cryptographically signed root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) mission tokens for every delegation hop.
    * Provide a standardized "Mission Receipt" format for cross-framework verification.
    * Maintain a non-repudiable audit trail of the intent delegation chain.
* **Non-Goals:**
    * Defining the semantic content of the mission itself (handled by the Reasoning Engine).
    * Providing long-term storage for reasoning traces (handled by the Telemetry Sink).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Architect
* **Primary Goal:** Verify that a multi-hop delegation (User -> Lead Agent -> Specialist Agent -> File Tool) remained within mission boundaries.
* **The Happy Path (Tasks):**
    1. User initiates a mission with a signed root intent.
    2. Lead Agent requests a sub-mission token from MCP Any.
    3. MCP Any issues an RMA token, binding the sub-task to the root mission receipt.
    4. Specialist Agent presents the token to execute a tool call.
    5. MCP Any validates the token's lineage before permitting execution.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        User->>LeadAgent: Start Mission (Root Intent)
        LeadAgent->>RMALayer: Request Sub-Mission Token (Root Receipt + Sub-Intent)
        RMALayer->>TPM: Sign Sub-Mission Receipt
        TPM-->>RMALayer: Attested Token
        RMALayer-->>LeadAgent: RMA Token
        LeadAgent->>SpecialistAgent: Delegate Task (RMA Token)
        SpecialistAgent->>RMALayer: Execute Tool (RMA Token)
        RMALayer->>RMALayer: Verify Lineage (Root -> Sub -> Task)
        RMALayer-->>Tool: Permit Execution
    ```
* **APIs / Interfaces:**
    * `POST /rma/token/issue`: Generates a new sub-mission receipt.
    * `POST /rma/token/verify`: Validates the lineage of a provided token.
* **Data Storage/State:** RMA receipts are stored in a hardware-bound, append-only log.

## 5. Alternatives Considered
* **Flat Session Tokens:** Rejected because they don't provide hierarchical proof of intent, allowing a compromised subagent to spawn unauthorized tasks under the same session.
* **Centralized Policy DB:** Rejected due to the latency of per-call lookups in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RMA tokens are bound to the specific hardware enclave of the issuer, preventing token reuse on unauthorized nodes.
* **Observability:** Every token issuance and verification event is logged to the Mission-Receipt Logging Service.

## 7. Evolutionary Changelog
* **2026-06-07:** Initial Document Creation.
