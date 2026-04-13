# Design Doc: CLAW-10 Compliance Mapper
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
With the release of the "CLAW-10" Enterprise Evaluation Matrix, organizations require a way to automatically verify if their AI agent deployments meet standardized security and governance criteria. MCP Any, as the core infrastructure layer, is uniquely positioned to collect the necessary telemetry and map it against these requirements. This feature will provide an automated compliance engine that maps internal state (Zero-Trust policies, hardware attestation status, tool reputation scores) to the CLAW-10 matrix.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically map MCP Any security features to CLAW-10 criteria.
    * Provide a real-time compliance dashboard/report for IT administrators.
    * Support export of compliance proofs to external auditing systems.
* **Non-Goals:**
    * Enforcing compliance (that is the job of the individual policy/security components).
    * Providing legal advice or guarantees of security.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise IT Security Auditor
* **Primary Goal:** Verify that a new agent team deployment meets the organization's "Safe-by-Default" requirements.
* **The Happy Path (Tasks):**
    1. Auditor opens the MCP Any Compliance Dashboard.
    2. The system scans the current active mission-root and associated subagents.
    3. The mapper evaluates the "Hardware-Attested Provenance" and "Safe-by-Default" scores.
    4. Auditor downloads a signed "CLAW-10 Compliance Report" showing 100% adherence.

## 4. Design & Architecture
* **System Flow:**
    * `Security Collector` -> `Mapping Engine` (Rego/CEL) -> `CLAW-10 Report Generator`
* **APIs / Interfaces:**
    * `GET /api/v1/compliance/claw-10`: Returns the current compliance status.
    * `POST /api/v1/compliance/attest`: Generates a signed report.
* **Data Storage/State:**
    * Compliance rules are stored as CEL (Common Expression Language) files.

## 5. Alternatives Considered
* **Manual Auditing:** Rejected due to the machine-speed nature of agent swarms.
* **Third-party Scanners:** Often lack the deep integration with the transport layer that MCP Any provides.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The compliance engine itself must be hardware-attested to prevent spoofed reports.
* **Observability:** Mapping events are logged to the central audit sink.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
