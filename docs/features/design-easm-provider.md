# Design Doc: Enclave-Aware Session Migration (EASM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents move from single-device sessions to multi-device meshes (e.g., Gemini CLI IBEM), the overhead of re-authenticating mission roots on every hardware transition is becoming a performance bottleneck. EASM allows an active reasoning session to be securely "handed over" between disparate physical TPM enclaves without full re-attestation.

MCP Any needs to solve this to enable seamless, low-latency agent mobility across the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate sub-100ms session migration between trusted enclaves.
    * Maintain cryptographic lineage and mission-root sovereignty during migration.
    * Support monotonic counter synchronization between source and target enclaves.
* **Non-Goals:**
    * Migrating sessions to non-TPM/non-attested environments.
    * Handling full filesystem migration (limited to reasoning state and tokens).

## 3. Critical User Journey (CUJ)
* **User Persona:** Mobile-to-Desktop Power User
* **Primary Goal:** Continue a 10-step complex code refactor started on a phone when moving to a workstation.
* **The Happy Path (Tasks):**
    1. Agent on phone initiates migration signal to MCP Any.
    2. MCP Any validates target workstation TPM signature.
    3. Source enclave signs a "Migration Manifest" containing current BSH state and MHL counters.
    4. Manifest is encrypted for target workstation enclave.
    5. Target workstation resumes session with hardware-attested continuity.

## 4. Design & Architecture
* **System Flow:**
    [Source Enclave] --(Encrypted Manifest)--> [EASM Broker] --(Identity Validation)--> [Target Enclave]
* **APIs / Interfaces:**
    * `POST /v1/migration/propose`: Propose session handover.
    * `POST /v1/migration/commit`: Finalize manifest ingestion.
* **Data Storage/State:**
    * Session metadata is held in kernel-locked memory during transition.
    * Handover tokens are one-time use.

## 5. Alternatives Considered
* **Full Re-attestation:** Rejected due to 2s+ user latency and "Attestation Exhaustion" on legacy TPMs.
* **Cloud-only Persistence:** Rejected to maintain "Local-First" Zero-Trust requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Handover requires hardware-attested "Acceptance Token" from target before source releases manifest.
* **Observability:** Migration events are logged in the Mission-Root Attestation Registry.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
