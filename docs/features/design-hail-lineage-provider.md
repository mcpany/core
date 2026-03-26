<!-- markdownlint-disable -->
# Design Doc: HAIL (Hardware-Attested Intent Lineage)

**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope

As agent swarms evolve, specialized subagents are increasingly prone to "Logic
Grafting" and identity spoofing. HAIL provides a standardized protocol for
cryptographically signing every sub-instruction and linking it back to a
hardware-attested mission-root intent.

## 2. Goals & Non-Goals

* **Goals:**

  * Cryptographically sign every inter-agent instruction using
    TPM/Secure Enclave.

  * Link all sub-instructions to an immutable mission-root intent.

  * Enable real-time lineage verification for all tool calls in deep swarms.

* **Non-Goals:**

  * Archiving full reasoning traces (handled by Telemetry Provider).

  * Providing transport security (handled by Isolated Named Pipes).

## 3. Critical User Journey (CUJ)

* **User Persona:** Security Auditor / Mission-Root Orchestrator

* **Primary Goal:** Verify that a tool call performed by a 3rd-degree subagent
  was legitimately authorized by the original user intent.

* **The Happy Path (Tasks):**

  1. Parent agent issues a mission-root intent with a TPM-bound signature.

  2. Subagent A receives instruction and signs its sub-delegation to
     Subagent B.

  3. Subagent B attempts a tool call.

  4. MCP Any validates the lineage chain from Subagent B back to the
     Mission-Root.

  5. Tool call is authorized based on verified intent lineage.

## 4. Design & Architecture

* **System Flow:**

  * Intent -> Hardware-Attestation (TPM) -> HAIL Token -> Instruction ->
    Verification -> Tool Call.

* **APIs / Interfaces:**

  * `POST /v1/hail/sign`: Sign an instruction fragment with hardware-bound
    identity.

  * `POST /v1/hail/verify`: Verify the cryptographic lineage of an instruction
    chain.

* **Data Storage/State:**

  * Lineage tokens are session-bound and stored in the Mission-Root Enclave.

## 5. Alternatives Considered

* **JWT-based Lineage:** Rejected due to the risk of token theft on compromised
  hosts; hardware-bound signatures provide non-repudiation.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Mandatory hardware-attestation at every hop in the
  delegation chain.

* **Observability:** Full lineage visibility in the Mesh-Resident Lineage
  Tracker.

## 7. Evolutionary Changelog

* **2026-06-19:** Initial Document Creation.
