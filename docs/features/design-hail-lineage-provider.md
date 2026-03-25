<!-- markdownlint-disable -->
# Design Doc: HAIL (Hardware-Attested Intent Lineage)
**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope
As agent swarms evolve, specialized subagents are increasingly prone to "Logic Grafting" and identity spoofing. HAIL provides a standardized protocol for cryptographically signing every sub-instruction and linking it back to a hardware-attested mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically sign inter-agent sub-instructions.
    * Link every sub-instruction to a hardware-attested mission-root intent.
    * Detect and neutralize "Logic Grafting" attacks in real-time.
    * Provide non-repudiable lineage for all reasoning fragments.
* **Non-Goals:**
    * Providing long-term archival of all reasoning traces.
    * Blocking instructions without attestation in "Audit-Only" mode.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Auditor / Mission-Root Orchestrator
* **Primary Goal:** Verify the entire lineage of a high-risk tool call performed by a 3rd-degree subagent.
* **The Happy Path (Tasks):**
    1. Parent agent issues a mission-root intent with a TPM-bound signature.
    2. Subagent A receives instruction and signs its sub-delegation to Subagent B using HAIL.
    3. Subagent B attempts a tool call.
    4. MCP Any validates the tool call by tracing the HAIL-chain back to the hardware-attested mission-root.
    5. Tool call is authorized.

## 4. Design & Architecture
* **System Flow:**
    * Intent -> Hardware-Attestation (TPM) -> HAIL Token -> Instruction -> Verification -> Tool Call.
* **APIs / Interfaces:**
    * `POST /v1/hail/sign`: Sign an instruction fragment.
    * `POST /v1/hail/verify`: Verify a lineage chain.
* **Data Storage/State:**
    * Lineage tokens are ephemeral and session-bound.

## 5. Alternatives Considered
* **JWT-based Lineage:** Rejected due to the risk of credential theft on compromised hosts.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory hardware-attestation at every hop.
* **Observability:** Integrated with the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-19:** Initial Document Creation.
