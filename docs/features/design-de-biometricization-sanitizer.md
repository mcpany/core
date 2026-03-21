# Design Doc: De-biometricization Sanitizer
**Status:** Draft
**Created:** 2026-04-29

## 1. Context and Scope
As agents increasingly handle sensitive project data in hybrid cloud/local environments, the risk of unintentional PII (Personally Identifiable Information) and biometric data leakage to external LLM providers has become a primary concern. Research from Purdue University has demonstrated that even "anonymous" datasets can often be de-anonymized if biometric markers are present.

The **De-biometricization Sanitizer** acts as a local-first security boundary, scrubbing sensitive markers from the agent's context and tool outputs before they leave the local environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically detect and redact PII (names, emails, SSNs) from textual context.
    * Identify and scrub biometric markers (voice prints, facial geometry metadata) from multimodal data.
    * Provide a pluggable architecture for domain-specific scrubbing rules (e.g., HIPAA, GDPR).
    * Maintain a "Sovereignty Audit Log" of all redacted fragments for local review.
* **Non-Goals:**
    * Encrypting data for cloud storage (handled by transport layers).
    * Providing perfect, 100% detection (uses a best-effort, heuristic-driven approach).

## 3. Critical User Journey (CUJ)
* **User Persona:** Healthcare Software Developer
* **Primary Goal:** Use a cloud-based agent to refactor a medical record processing tool without leaking patient PII.
* **The Happy Path (Tasks):**
    1. Developer enables the "PII-Sovereign Context Scrubber" in MCP Any.
    2. Agent retrieves a sample medical record from the local database.
    3. The Sanitizer intercepts the tool output.
    4. Names, birthdates, and insurance IDs are replaced with cryptographic tokens (e.g., `[REDACTED_PII_001]`).
    5. The sanitized text is sent to the cloud LLM.
    6. Agent provides a refactored code snippet.
    7. Developer reviews the logs and sees exactly what was scrubbed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        LocalData[Local Data/Tools] --> Interceptor[MCP Any Interceptor]
        Interceptor --> Scrubber[De-biometricization Engine]
        Scrubber --> RedactedData[Sanitized Context]
        RedactedData --> CloudLLM[Cloud LLM Reasoning]
    ```
* **APIs / Interfaces:**
    * `scrub_context(raw_text, ruleset_id)`
    * `register_scrub_rule(pattern, replacement_strategy)`
* **Data Storage/State:**
    * Local "Redaction Map" to allow reversible de-tokenization for local-only tools.
    * Policy definitions stored in YAML.

## 5. Alternatives Considered
* **Cloud-Side Sanitization:** Rejected because the data has already left the sovereign boundary.
* **Manual Redaction:** Impossible for high-frequency, autonomous agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sanitizer is the last line of defense before data exfiltration.
* **Observability:** Integrated with the "Sovereignty Audit Dashboard" in the UI.

## 7. Evolutionary Changelog
* **2026-04-29:** Initial Document Creation.
* **2026-04-30:** Update: Multi-Modal Biometric Redaction.
    * **Context**: Today's market sync revealed a surge in "Multi-Modal Sovereignty" requirements as agents move from text-only to video/audio task execution.
    * **Architecture Adjustment**: Expanded the `scrub_context` interface to support binary streams and multimedia metadata.
    * **Security Impact**: Prevents the exfiltration of facial geometry and voice-print metadata found in high-resolution multi-modal context buffers.
