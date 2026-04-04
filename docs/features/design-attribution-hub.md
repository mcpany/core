# Design Doc: NIST-Compliant Attribution Hub
**Status:** Draft
**Created:** 2026-07-26

## 1. Context and Scope
With the release of the NIST AI 200.5 draft and the rise of autonomous "Insider Threats," organizations require a non-repudiable method to attribute AI actions to human-verified intents. MCP Any must evolve from a simple proxy to an **Attribution Root**.

The Attribution Hub solves the "Agency Attribution Gap" where subagent actions are logged but cannot be cryptographically proven to originate from a specific user-authorized mission branch. This is critical for regulatory compliance and forensic auditing.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate TPM-signed "Reasoning Provenance Certificates" (RPCs) for every tool call.
    * Maintain an immutable, hash-chained log of intent-to-action transitions.
    * Support real-time verification of attribution lineage by third-party auditors.
* **Non-Goals:**
    * It will NOT perform semantic validation of the reasoning itself (handled by the ARI Hub).
    * It will NOT store the raw context data (handled by the SRM Provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Compliance Officer
* **Primary Goal:** Verify that a specific $50k wire transfer initiated by a finance agent was explicitly authorized by the CFO's mission-root.
* **The Happy Path (Tasks):**
    1. The officer selects the transaction ID in the audit console.
    2. The console queries the MCP Any Attribution Hub for the RPC.
    3. The Hub retrieves the hash-chained certificate linked to the CFO's hardware-attested mission token.
    4. The console verifies the RPC signature against the CFO's public key.
    5. The officer receives a "Verified: Authorized by CFO" signal.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `Attribution Middleware` -> `Sign RPC` -> `Commit to Chain` -> `Execute Tool`.
* **APIs / Interfaces:**
    * `GET /v1/attribution/cert/{action_id}`: Retrieves the RPC for a specific action.
    * `POST /v1/attribution/verify`: Validates a provided RPC against the Hub's chain.
* **Data Storage/State:**
    * Uses a local-first, hash-chained SQLite database (Attribution Ledger) with periodic snapshots to cloud storage.

## 5. Alternatives Considered
* **Plain Text Logging:** Rejected due to lack of non-repudiation.
* **Blockchain-based Attribution:** Rejected due to prohibitive latency and privacy concerns for mission metadata.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RPC signing key is stored in a hardware enclave (TPM) and never exposed to the agent process.
* **Observability:** Integrated with the Trace-Aware Identity (TAI) bridge for end-to-end mission visibility.

## 7. Evolutionary Changelog
* **2026-07-26:** Initial Document Creation.
