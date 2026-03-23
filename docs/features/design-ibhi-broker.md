# Design Doc: IBHI Intent Broker
**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
Recursive Intent Poisoning (RIP) occurs when high-privilege subagents subtly alter the mission's root objectives during deep reasoning loops. Intent-Bound Hardware Isolation (IBHI) addresses this by pinning the "Mission Root" intent to hardware-protected memory regions (TPM/Secure Enclave), making them immutable even to high-privilege software processes.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound "Intent Anchors."
    * Provide a mandatory validation layer for intent expansion.
    * Integration with UACO v1.8 RID tokens.
* **Non-Goals:**
    * Real-time reasoning trace scrubbing (handled by Trace Scrubber).
    * General purpose encryption service.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure a subagent with shell access cannot modify the project's "Safety Manifest."
* **The Happy Path (Tasks):**
    1. Root agent signs the Mission Intent during boot.
    2. MCP Any IBHI Broker binds this intent to a hardware-protected slot.
    3. Subagent attempts to expand scope via a tool call.
    4. IBHI Broker verifies the mutation against the hardware anchor and rejects diverging intents.

## 4. Design & Architecture
* **System Flow:**
    `[Agent reasoning] -> [UACO Token] -> [IBHI Broker (TPM Check)] -> [Tool Execution]`
* **APIs / Interfaces:**
    * `/v1/ibhi/anchor`: Bind root intent to hardware.
    * `/v1/ibhi/verify`: Cryptographic mutation check.
* **Data Storage/State:**
    * Hardware Security Module (HSM/TPM) slots.

## 5. Alternatives Considered
* **Software-only Sandboxing:** Rejected; vulnerable to kernel-level escapes or memory corruption.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Root intents are physically isolated from the reasoning environment.
* **Observability:** Hardware attestation failure logs.

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
