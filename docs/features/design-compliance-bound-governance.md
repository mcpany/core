# Design Doc: Compliance-Bound Governance (CBG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the approach of the August 2026 EU AI Act enforcement deadline, enterprise users of MCP Any require a way to prove that their autonomous agents are operating within legal and safety boundaries. Current logging is sufficient for debugging but lacks the non-repudiable, hardware-attested traceability required for regulatory audits.

Compliance-Bound Governance (CBG) provides the infrastructure to generate and manage non-repudiable audit logs that trace every tool call, system action, and reasoning fragment back to a hardware-attested mission root and user intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested (TPM/SEP), non-repudiable audit logs for all agent actions.
    * Enable "Reasoning Traceability" mapping every tool call to a verified reasoning step.
    * Support "Legal Attestation" exports that satisfy EU AI Act documentation requirements for high-risk systems.
    * Neutralize "Shadow Actions" by requiring cryptographic links between reasoning and execution.
* **Non-Goals:**
    * Enforcing ethical or moral judgments (this is handled by the Policy Firewall).
    * Providing long-term storage for audit logs (will interface with external SIEMs or the Asynchronous Telemetry Sink).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Compliance Officer
* **Primary Goal:** Verify that a specific autonomous filesystem edit was explicitly authorized by the mission-root intent during a regulatory audit.
* **The Happy Path (Tasks):**
    1. Compliance officer selects a specific system action from the Audit Dashboard.
    2. CBG retrieves the hardware-attested "Action Provenance" token for that action.
    3. The token is used to resolve the complete "Chain of Reason" leading back to the Mission Root.
    4. The system validates the TPM signatures at every step of the chain.
    5. A signed "Compliance Report" is generated, proving the action was authorized and traceable.

## 4. Design & Architecture
* **System Flow:**
    `[Action] -> [Action Provenance Provider] -> [CBG Hub] -> [Hardware Attestation Registry] -> [Audit Log]`
* **APIs / Interfaces:**
    * `cbg.AttestAction(actionID, reasoningFragmentID) -> ProvenanceToken`: Links an action to its reasoning source.
    * `cbg.VerifyChain(provenanceToken) -> AuditTrail`: Validates the lineage of an action.
    * `cbg.GenerateComplianceReport(missionID) -> SignedDocument`: Exports a legal audit trail.
* **Data Storage/State:**
    * **Provenance Registry:** SQLite-backed storage for cryptographically linked action-reasoning pairs.
    * **Audit Shards:** Mission-bound, read-only shards for high-fidelity reasoning traces.

## 5. Alternatives Considered
* **Plain JSON Logging:** Rejected because logs can be tampered with or "ghosted." CBG requires hardware-bound signatures.
* **Full-State Mirroring:** Rejected due to prohibitive storage and privacy overhead. CBG focuses on "Traceable Fragments."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The CBG Hub itself operates in a high-trust enclave. Access to the Audit Dashboard requires MFA.
* **Observability:** Integrated with the "Compliance Audit Dashboard" in the UI for real-time traceability monitoring.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
