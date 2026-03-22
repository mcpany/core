# Design Doc: Reasoning-Action Correlation (RAC)
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
The emergence of "Spectral Reasoning" side-channel attacks allows malicious subagents to probe mission-root constraints by monitoring timing variations in reasoning outputs. This reveals sensitive information about the parent agent's attention and security boundaries. RAC provides an active, hardware-locked defense by cryptographically correlating tool actions with their authoritative reasoning lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically correlate tool actions with hardware-attested reasoning fragments.
    * Provide a verifiable Chain-of-Thought Lineage for all tool calls.
    * Neutralize timing-based side-channel probes of mission constraints.
* **Non-Goals:**
    * Does not eliminate all side-channels (e.g., power consumption, which is out of scope for software-only defense).
    * Does not perform reasoning on behalf of agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Sensitive Swarm Auditor
* **Primary Goal:** Verify that a high-risk tool call was directly authorized by a specific, user-vetted reasoning chain.
* **The Happy Path (Tasks):**
    1. Agent generates a reasoning fragment proposing a tool call.
    2. RAC generates a hardware-bound "Correlation Token."
    3. Tool is executed; the Correlation Token is embedded in the tool call metadata.
    4. Auditor verifies the token against the hardware-attested reasoning chain.

## 4. Design & Architecture
* **System Flow:**
    `Reasoning Engine` -> `[Correlation Token (RAC)]` -> `Tool Execution`
* **APIs / Interfaces:**
    * Integration with Signed Reasoning Monologue (SRM) Provider.
    * Middleware for tool call interception and token injection.
* **Data Storage/State:**
    * In-memory buffer for active reasoning fragments and correlation tokens.
    * Integration with TPM/Secure Enclave for hardware-bound signing.

## 5. Alternatives Considered
* **Passive Logging:** Rejected; it allows exfiltration to occur before the audit is conducted.
* **Batching Reasoning:** Rejected; it introduces unacceptable latency for real-time agentic workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RAC is essential for maintaining mission-root sovereignty in horizontal meshes.
* **Observability:** Full lineage for every tool call is recorded in the audit log.

## 7. Evolutionary Changelog
* **2026-03-22:** Initial Document Creation.
