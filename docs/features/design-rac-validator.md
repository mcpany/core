# Design Doc: Recursive Attestation Chain (RAC) Validator
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become deeper and more autonomous, the risk of "Fragment Splicing"—where a compromised intermediate subagent injects malicious reasoning or instructions into the cognitive chain—has become a critical threat. Current identity-based models only verify the immediate caller, not the complete lineage of the intent.

The RAC Validator is designed to provide a multi-layered cryptographic proof system that ensures every subagent instruction is semantically and cryptographically linked back to the original mission root. This provides a "Zero-Trust Lineage" for all tool calls and state mutations within a swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a recursive token system where each subagent must provide proof of parentage.
    * Neutralize "Fragment Splicing" by validating the entire chain of command for every high-risk tool call.
    * Provide hardware-attested (TPM) anchors for the mission root.
    * Support cross-framework (OpenClaw, Claude Code, AutoGen) lineage verification.
* **Non-Goals:**
    * This system WILL NOT replace transport-layer security (mTLS/Named Pipes).
    * It WILL NOT perform real-time LLM reasoning; it focuses on the cryptographic and structural integrity of the lineage metadata.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Architect
* **Primary Goal:** Ensure that a specialized "Database Optimizer" subagent cannot be coerced by a malicious intermediate agent into exfiltrating schema data.
* **The Happy Path (Tasks):**
    1. The Mission Root (Supervisor) initiates a task and signs it with a TPM-bound RAC token.
    2. The Supervisor delegates a sub-task to the Intermediate Agent, appending its own RAC signature.
    3. The Intermediate Agent delegates to the Specialist Agent, adding its layer of attestation.
    4. The Specialist Agent attempts a tool call (e.g., `read_database_schema`).
    5. The MCP Any RAC Validator intercepts the call and verifies the complete cryptographic chain back to the Mission Root.
    6. Validation succeeds, and the tool call is authorized.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -- RAC Token V1 --> IA[Intermediate Agent]
        IA -- RAC Token V1+V2 --> SA[Specialist Agent]
        SA -- Tool Call + RAC Token Chain --> RV[RAC Validator]
        RV -- Lineage Verification --> AB[Auth Broker]
        AB -- Authorize --> TS[Tool Server]
    ```
* **APIs / Interfaces:**
    * `/v1/rac/verify`: Endpoint for validating a recursive attestation chain.
    * `X-MCP-RAC-Chain`: Header containing the nested cryptographic signatures.
* **Data Storage/State:**
    * Ephemeral mission-bound public keys are stored in the RAC Registry.

## 5. Alternatives Considered
* **Flat Identity Tokens:** Rejected because they do not protect against intermediate hijacking.
* **Centralized Authorization Service:** Rejected to avoid being a single point of failure and to maintain the decentralized nature of swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory hardware-bound (TPM) root of trust for the initial RAC token.
* **Observability:** Detailed lineage logs are exported to the Agent Chain Tracer.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
