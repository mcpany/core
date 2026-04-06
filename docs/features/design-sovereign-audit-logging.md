# Design Doc: Sovereign Audit Logging (SAL)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
SaaS vendors are increasingly imposing an "Audit Log Tax," charging significant premiums for high-fidelity, granular logs. This financial barrier prevents security teams from accessing the data necessary for effective breach investigation and compliance auditing. Simultaneously, the rise of autonomous agents as "Insider Threats" demands absolute visibility into their actions.

Sovereign Audit Logging (SAL) evolves MCP Any into the authoritative "Local Log Mint." It provides hardware-attested, high-fidelity logs generated at the infrastructure layer, bypassing vendor-imposed visibility gaps.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate high-fidelity, granular audit logs for all agent activities locally.
    * Mandate hardware-attested (TPM) signatures for every log entry to ensure non-repudiability.
    * Provide a standardized export format for SIEM integration.
    * Neutralize the SaaS "Audit Log Tax" by moving truth attestation to the local bus.
* **Non-Goals:**
    * Replacing long-term log storage (SIEMs handle this).
    * Providing natural language log analysis (handled by specialized auditor agents).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Incident Responder
* **Primary Goal:** Investigate a potential data exfiltration event by a subagent without relying on expensive, delayed SaaS vendor logs.
* **The Happy Path (Tasks):**
    1. An anomaly is detected in agent behavior.
    2. The responder accesses the MCP Any Audit Sovereignty Hub.
    3. The responder queries the hardware-attested log for the specific agent session.
    4. SAL provides a complete, timestamped sequence of every tool call, context injection, and reasoning fragment, each signed by the host TPM.
    5. The responder verifies the integrity of the trace using the hardware root of trust.
    6. The responder identifies the exact point of exfiltration and the parentage of the instruction.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Activity] --> B[Middleware Hook]
        B --> C[SAL Provider]
        C --> D[Hardware Attestation Hub]
        D -->|TPM Sign| E[Local Attested Log]
        E --> F[SIEM Export Connector]
    ```
* **APIs / Interfaces:**
    * `sal.LogEvent(eventData, sessionToken) -> LogID`: Records an attested event.
    * `sal.QueryLogs(filter) -> AttestedTrace`: Retrieves verified log sequences.
* **Data Storage/State:**
    * **Attested Buffer:** A high-speed, local append-only log file, periodically rotated and flushed to external storage.

## 5. Alternatives Considered
* **Vendor-native Logging:** Rejected due to cost ("Audit Log Tax") and lack of hardware-bound cross-framework consistency.
* **Ephemeral Logging:** Rejected because it doesn't satisfy compliance (SSDF) or forensic requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The log issuer must be isolated from the agent execution environment to prevent log tampering by compromised agents.
* **Observability:** Integrated with the "Audit Sovereignty Hub" in the UI for real-time visualization of agent traces.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
