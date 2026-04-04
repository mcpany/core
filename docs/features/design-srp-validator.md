# Design Doc: Speculative Reason Proof (SRP) Validator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Gemini CLI v0.59.0-rc, speculative reasoning is being augmented with Speculative Reason Proofs (SRP). Standard speculative execution provides low latency but lacks cryptographic accountability during the "pre-commit" phase. This creates a window where hallucinated or malicious reasoning can influence the agent's path before the discovery quorum completes.

The SRP Validator provides hardware-attested verification of these intermediate reasoning steps, closing the gap between high-speed execution and Zero-Trust auditing.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested (TPM/Secure Enclave) verification of speculative reasoning fragments.
    * Neutralize "Reasoning-Path Mimicry" by anchoring fragments to a verified mission root.
    * Enable sub-millisecond validation of "Progress Proofs" for ALO extensions.
    * Implement cross-framework translation of reason proofs (Gemini to OpenClaw).
* **Non-Goals:**
    * Validating the final model weights.
    * Replacing the Pre-Commit Speculative Sanitizer (PCSS) - SRP focuses on provenance, PCSS focuses on semantic content.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Agent Team Auditor
* **Primary Goal:** Verify that a speculative tool preparation step was driven by a valid parent instruction and not an injected "Shadow Intent."
* **The Happy Path (Tasks):**
    1. Agent initiates a speculative preparation of a `db_query` tool.
    2. Agent generates an SRP fragment for the reasoning step.
    3. SRP Validator intercepts the fragment.
    4. Validator verifies the hardware signature and lineage against the mission root.
    5. Validator signals "Verified" to the Optimistic Quorum Gateway.
    6. Tool result is held in a probabilistic buffer until final quorum, but with higher trust confidence.

## 4. Design & Architecture
* **System Flow:**
    [Speculative Step] --> [SRP Fragment Generation] --> [SRP Validator (TPM Check)] --> [Mission Lineage Matcher] --> [Verified Conf Signal]
* **APIs / Interfaces:**
    * `VerifyProof(fragment, missionRoot)`: Core verification logic.
    * `x-mcpany-srp`: Header for transporting proof fragments.
* **Data Storage/State:**
    * Proof fragments are transient; validation results are appended to the Immutable State Trail.

## 5. Alternatives Considered
* **Full Session Attestation:** Rejected as too slow (100ms+) for high-frequency speculative steps.
* **Semantic Analysis Only:** Rejected because semantic analysis can be bypassed by sophisticated linguistic mimicry; cryptographic provenance is required for Zero Trust.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SRPs are chained to the hardware-bound Mission Manifest (HAMM).
* **Observability:** Proof verification success rates are tracked as a mesh stability metric.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
