# Design Doc: Output-Linked Proof (OLP) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The discovery of "Z-Proof Poisoning" in Gemini CLI demonstrates that privacy-preserving reasoning proofs can be weaponized if they are only linked to the reasoning path. Malicious subagents can "graft" a valid reasoning proof onto an unauthorized tool output, misleading auditors into believing a high-risk action was constrained by the mission root.

MCP Any needs to implement the Output-Linked Proof (OLP) Hub. By cryptographically binding the ZK-proof to a hash of the tool's semantic output, we ensure that the proof is only valid for the specific result produced, neutralizing context-grafting attacks.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate ZK-proofs that are cryptographically bound to both the reasoning path and the tool output.
    * Provide a verification interface for auditors to validate OLP-compliant proofs.
    * Ensure that sensitive mission context remains private within the proof (Zero-Knowledge).
    * Neutralize "Context-Grafting" (Z-Proof Poisoning) exploits.
* **Non-Goals:**
    * Verifying the correctness of the tool output itself (only that the output matches the proof).
    * Replacing the primary reasoning engine (OLP is an attestation layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Auditor
* **Primary Goal:** Verify that an agent's request to "Delete Temporary Files" didn't secretly include "Delete Production Database" by checking the output-linked proof.
* **The Happy Path (Tasks):**
    1. The Agent executes a "Cleanup" task and generates a result.
    2. The OLP Hub intercepts the result and the reasoning trace.
    3. The Hub generates a ZK-proof that includes `H(ReasoningPath || OutputHash)`.
    4. The Auditor receives the proof and the OutputHash.
    5. The Auditor uses the OLP Hub to verify the proof.
    6. If the agent had grafted a "Valid Path" proof onto a "Database Deleted" output, the verification would fail because the OutputHash wouldn't match the proof's binding.

## 4. Design & Architecture
* **System Flow:**
    `[Tool Result] -> [OLP Hub (Hash & Bind)] -> [ZK-Proof Generation] -> [Auditor Verification]`
* **APIs / Interfaces:**
    * `POST /v1/attestation/generate`: Generate an output-linked proof.
    * `POST /v1/attestation/verify`: Validate an OLP token.
* **Data Storage/State:**
    * Verification keys are stored in the Hardware-Attested Mission Root.
    * Short-term caching of OutputHashes to support asynchronous verification.

## 5. Alternatives Considered
* **Full Monologue Logging:** Rejected due to privacy concerns and the prohibitive token cost of storing multi-gigabyte traces.
* **Path-Only Proofs (Legacy):** Rejected because they are vulnerable to the newly discovered context-grafting attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** OLP Hub provides the "Trust-but-Verify" signal needed for autonomous scale.
* **Observability:** OLP verification failures trigger immediate mission-root revocation via the Subagent Reaper.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
