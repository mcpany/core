# Design Doc: Reasoning Path Integrity (RPI) Validator
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
As swarms become deeper and more horizontal, verifying the "Chain-of-Thought" (CoT) becomes as important as verifying the final tool call. Currently, subagents can provide plausible reasoning traces that actually diverge from the mission root, leading to "Logic Grafting" or "Semantic Hallucinations."

The RPI Validator leverages Gemini's ARE v1.8 standard and hardware-attestation to validate internal reasoning steps. It ensures that the reasoning path remains untampered and semantically anchored to the mission root throughout the session.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate hardware-signed internal reasoning fragments.
    * Maintain a "Semantic Hash-Chain" of the entire reasoning process.
    * Detect "Logic Grafting" where a subagent appends unauthorized reasoning paths.
    * Provide a cryptographic proof of reasoning provenance.
* **Non-Goals:**
    * Evaluating the *quality* of the reasoning (it's an integrity check, not a correctness check).
    * Restricting reasoning depth (managed by budget controllers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Forensic AI Auditor
* **Primary Goal:** Verify that a tool call was the direct result of an authorized reasoning path, not a "Spliced" instruction.
* **The Happy Path (Tasks):**
    1. A subagent submits a tool call along with its reasoning trace and RPI headers.
    2. The RPI Validator extracts the semantic hash-chain from the trace.
    3. The Validator verifies the hardware signatures on each reasoning step.
    4. The Validator confirms that the chain-head is a valid descendant of the parent's mission intent.
    5. The tool call is authorized.

## 4. Design & Architecture
* **System Flow:**
    [Reasoning Fragment] -> [TPM Signer] -> [RPI Header]
    [RPI Header] -> [RPI Validator] -> [Semantic Hash Comparator] -> [Attestation Result]
* **APIs / Interfaces:**
    * `POST /v1/rpi/validate`: Validates a reasoning trace against its parent lineage.
    * `GET /v1/rpi/lineage?session_id=[id]`: Retrieves the cryptographically verified reasoning tree.
* **Data Storage/State:**
    Reasoning hash-chains are stored in the SRM (Signed Reasoning Monologue) provider sidecar.

## 5. Alternatives Considered
* **Plaintext Trace Matching:** Rejected because it's vulnerable to adversarial editing within the context window.
* **Full-Trace Re-Inference:** Rejected due to the massive compute and latency overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RPI ensures that subagents cannot bypass PBR (Policy-Bound Reasoning) by "Hallucinating" a valid path.
* **Observability:** UI integration with the "Reasoning Path Integrity Viewer" for real-time visualization of the hash-chain.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
