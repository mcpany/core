# Design Doc: Hardware-Attested Attestation Relay (HAAR)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Managing hardware-attested identity fragments across heterogeneous cloud enclaves is leading to "Trust-Decay," where session-bound credentials lose attestation strength during cross-enclave handoffs. HAAR provides a cryptographically signed bridge for hardware-attested identities, ensuring trust continuity and attestation strength across multi-cloud swarm boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Maintain hardware-bound attestation strength across multi-cloud boundaries.
    * Provide a universal relay for session-bound identity fragments.
    * Support sub-100ms identity synchronization between disparate enclaves.
    * Ensure non-repudiable lineage tracking for cross-cloud agent actions.
* **Non-Goals:**
    * Acting as a central identity provider (federated model only).
    * Managing non-attested (low-trust) identities.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Cloud Swarm Architect
* **Primary Goal:** Delegate a sensitive task from an agent in AWS Nitro Enclave to an agent in Azure confidential computing without losing the hardware-bound trust signal.
* **The Happy Path (Tasks):**
    1. The AWS-resident agent generates a hardware-attested task proposal.
    2. HAAR intercepts the proposal and verifies the AWS TPM signature.
    3. HAAR wraps the identity fragment in a "Cross-Enclave Relay Token" signed by its own hardware root.
    4. The Azure-resident agent receives the relay token and verifies the HAAR signature.
    5. The task is executed with the same trust level as if it were local to the original enclave.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        E1[Cloud Enclave A] -->|Attested Identity| HAAR[HAAR]
        HAAR -->|Verify & Sign| HAAR
        HAAR -->|Relay Token| E2[Cloud Enclave B]
        E2 -->|Verify Relay| Action[Execute Task]
    ```
* **APIs / Interfaces:**
    * `POST /v1/relay/sign`: Wrap a local enclave attestation in a relay token.
    * `POST /v1/relay/verify`: Verify a relay token from a remote HAAR instance.
* **Data Storage/State:**
    * Relay keys are rotated every 15 minutes and resident only in volatile hardware memory.

## 5. Alternatives Considered
* **Direct Enclave-to-Enclave Attestation:** Rejected as there is no standardized protocol for cross-provider TPM verification.
* **Centralized JWT Bridge:** Rejected as it moves the root of trust away from hardware.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HAAR instances must peer using mutual hardware attestation. Compromise of one node triggers immediate global revocation.
* **Observability:** Cross-enclave "Handshake Latency" and "Trust Strength" visualized in the Connectivity & Security Dashboard.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
