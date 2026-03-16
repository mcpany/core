# Design Doc: PII-Sovereign Context Scrubber
**Status:** Draft
**Created:** 2026-04-30

## 1. Context and Scope
As AI agents move from local-only execution to hybrid-cloud reasoning loops, the risk of propagating Personally Identifiable Information (PII) or biometric data to external LLM providers increases. MCP Any needs a mandatory sanitization layer that acts as the authoritative "Local Scrubber," ensuring that context is de-biometricized and anonymized before it leaves the secure local environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-performance local middleware for scanning agent context (text and metadata).
    * Automatically redact or mask PII (names, emails, phone numbers) and biometric signals.
    * Provide a verifiable "Sovereignty Receipt" for each scrubbed context fragment.
    * Support custom redaction patterns for enterprise-specific data.
* **Non-Goals:**
    * Encrypting data for storage (handled by the Shared KV Store).
    * Sanitizing model *outputs* (handled by Prompt Path Protection).

## 3. Critical User Journey (CUJ)
* **User Persona:** Privacy-Conscious Enterprise Architect
* **Primary Goal:** Ensure that local customer data accessed by an agent is never sent to a cloud LLM in its raw form.
* **The Happy Path (Tasks):**
    1. Architect enables the PII-Sovereign Scrubber in the MCP Any configuration.
    2. Agent retrieves a customer record from a local database tool.
    3. Before the record is sent to the LLM as context, MCP Any intercepts the payload.
    4. The Scrubber identifies the customer's name and email and replaces them with stable, reversible tokens (e.g., `[USER_A]`).
    5. The sanitized context is sent to the cloud LLM.
    6. The LLM performs reasoning and returns a response referencing `[USER_A]`.
    7. MCP Any optionally re-identifies the tokens in the final output before presenting it to the user.

## 4. Design & Architecture
* **System Flow:**
    `[Tool Result] -> [PII-Sovereign Scrubber] -> [Context Compactor] -> [External LLM]`
* **APIs / Interfaces:**
    * `Scrubber.Process(payload)`: Primary interface for sanitizing context blobs.
    * `Scrubber.GetReceipt(context_id)`: Retrieves the attestation of sanitization.
* **Data Storage/State:**
    * Mapping tables for reversible anonymization are stored in a session-bound memory cache, never written to disk.

## 5. Alternatives Considered
* **Client-Side Sanitization:** Rejected because individual agents cannot be trusted to self-sanitize in a Zero-Trust environment.
* **Cloud-Side Redaction:** Rejected because the data has already left the sovereign boundary by the time it reaches the cloud.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Scrubber itself runs in a restricted sandbox with no network access.
* **Observability:** Scrubbing events are logged with "Sensitivity Scores," highlighting tools that frequently produce high-PII output.

## 7. Evolutionary Changelog
* **2026-04-30:** Initial Document Creation.
