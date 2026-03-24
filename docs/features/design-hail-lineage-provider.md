<!-- markdownlint-disable -->
# Design Doc: HAIL (Hardware-Attested Intent Lineage)
**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope
As agent swarms evolve, specialized subagents are increasingly prone to "Logic Grafting" and identity spoofing. Current transport-layer security is insufficient to verify the *behavioral identity* of a reasoning fragment. HAIL provides a standardized protocol for cryptographically signing every sub-instruction and linking it back to a hardware-attested mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-bound (TPM) signatures for all reasoning fragments.
    * Establish a non-repudiable "Chain of Command" for sub-delegations.
    * Enable real-time verification of "Reasoning Fragment Authorship."
* **Non-Goals:**
    * Real-time sanitization of the *content* of the reasoning (handled by ISD).
    * General-purpose identity for human users.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Verify that a subagent tool call was authorized by the parent mission-root without exposing parent environment variables.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a session with a TPM-signed mission-root token.
    2. Parent spawns a subagent and issues a "Lineage Fragment" token.
    3. Subagent makes a tool call to the HAIL-compliant gateway.
    4. Gateway verifies the fragment's signature against the mission-root lineage.
    5. Tool is executed only if the lineage is valid.

## 4. Design & Architecture
* **System Flow:**
    `[Parent (TPM Root)] -> [Lineage Token Generator] -> [Subagent] -> [HAIL Gateway (Verification)] -> [Tool]`
* **APIs / Interfaces:**
    * `x-mcpany-hail-lineage`: Header containing the signed lineage fragment.
    * `/v1/hail/verify`: Internal endpoint for verifying fragment tokens.
* **Data Storage/State:** Lineage state is ephemeral and bound to the hardware-attested session.

## 5. Alternatives Considered
* **JWT-only signing**: Rejected because it is vulnerable to token theft/leakage. Hardware-bound signatures provide non-repudiation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HAIL is the foundation of Behavioral Zero Trust, ensuring that only "known good" reasoning paths can execute high-risk tools.
* **Observability:** Lineage traces are logged as non-repudiable audit trails.

## 7. Evolutionary Changelog
* **2026-06-19:** Initial Document Creation.
