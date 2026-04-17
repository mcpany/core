# Design Doc: Epistemic Uncertainty Broker (EUB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLMs grow more autonomous, they are increasingly prone to "Silent Hallucinations" where they execute high-stakes tools despite having low underlying reasoning confidence. Standard tool-gating only checks permissions, not the *quality* of the reasoning driving the call. Gemini CLI v0.60.0's introduction of Epistemic Uncertainty Mapping (EUM) provides a standardized way for models to signal their doubt.

The Epistemic Uncertainty Broker (EUB) is required to intercept these signals and mandate human-in-the-loop (HITL) or supervisor intervention when model confidence falls below a mission-critical threshold.

## 2. Goals & Non-Goals
* **Goals:**
    * Ingest and validate `x-gemini-epistemic-confidence` headers and equivalents.
    * Trigger automated supervisor escalations or HITL blocks based on uncertainty scores.
    * Map raw uncertainty signals to hardware-attested "Epistemic Badges."
    * Provide a real-time "Confidence Heatmap" for the UI.
* **Non-Goals:**
    * Modifying the model's internal reasoning process.
    * Replacing existing Zero-Trust permission gates.
    * Providing a general-purpose model evaluation suite.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Supervisor
* **Primary Goal:** Prevent an agent from deleting a database table if its reasoning confidence is < 70%.
* **The Happy Path (Tasks):**
    1. Agent generates a reasoning fragment for `drop_table` with an EUM confidence of 0.65.
    2. The EUB middleware intercepts the request and detects the sub-threshold confidence.
    3. EUB generates an "Epistemic Doubt" signal and suspends the tool call.
    4. The supervisor receives a notification via the UI "Epistemic Confidence Heatmap."
    5. The supervisor reviews the reasoning fragment and either approves, denies, or provides a "Corrective Intent."
    6. EUB logs the intervention and releases the call if approved.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B[EUB Middleware]
        B --> C{Confidence Check}
        C -->|> Threshold| D[Permission Firewall]
        C -->|< Threshold| E[Supervisor Escalation]
        E --> F[UI Confidence Heatmap]
        F --> G[Human/Supervisor Decision]
        G -->|Approve| D
        G -->|Deny| H[Abort & Log]
    ```
* **APIs / Interfaces:**
    * `eub.InterceptRequest(headers, fragment) -> Decision`: Main interception logic.
    * `eub.RegisterEscalation(sessionID, confidenceScore) -> EscalationID`: Triggers the HITL flow.
* **Data Storage/State:**
    * **Epistemic Audit Log:** SQLite-backed storage of reasoning fragments and their associated confidence scores.

## 5. Alternatives Considered
* **Reactive Error Handling:** Rejected because it triggers *after* a potential failure. EUB is proactive.
* **Prompt Engineering (Asking 'Are you sure?'):** Rejected as it is susceptible to "Overconfidence Hallucination." EUB uses structured model metadata (EUM).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Epistemic signals must be hardware-attested (TPM) to prevent a compromised subagent from spoofing high confidence.
* **Observability:** Integrated with the UI "Reasoning Confidence Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
