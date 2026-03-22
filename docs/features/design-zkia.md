# Design Doc: Zero-Knowledge Intent Attestation (ZKIA)
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As AI agents move toward long-running, autonomous missions, they increasingly rely on a vast "Blackboard" of shared context. However, the rise of "Context-Window Flooding" (CWF) and "Reasoning Noise" attacks means that an agent's reasoning can be subtly hijacked by unauthorized data fragments. MCP Any needs a way to prove that an agent's tool-call reasoning was derived *strictly* from authorized context fragments without requiring the agent to reveal its private internal monologue (which may contain PII or sensitive thought traces).

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested cryptographic proofs of reasoning derivation.
    * Neutralize "Context-Window Flooding" by verifying the "Allow-List" of context inputs.
    * Preserve agent reasoning privacy via Zero-Knowledge proofs.
* **Non-Goals:**
    * Real-time monitoring of the model's weights or internal activations.
    * Preventing all forms of hallucination (ZKIA only proves *derivation*, not *truth*).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Verify that a specialist subagent's shell command was based only on the user's mission-root instructions, not a malicious `GEMINI.md` file.
* **The Happy Path (Tasks):**
    1. The gateway tags authorized context fragments with hardware-bound IDs.
    2. The agent reasoning engine generates a tool-call request.
    3. The ZKIA Validator requests a derivation proof for the tool-call arguments.
    4. The agent provides a ZK-proof linking the output to the allow-listed fragment IDs.
    5. The gateway validates the proof and the hardware signature before executing the tool.

## 4. Design & Architecture
* **System Flow:**
    `[Context Registry] -> (Tagging) -> [Agent Reasoning] -> (ZK-Proof Generation) -> [ZKIA Validator] -> (Attestation) -> [Tool Execution]`
* **APIs / Interfaces:**
    * `POST /v1/attest/derivation`: Submit a ZK-proof for a specific tool call.
    * `GET /v1/context/tags`: Retrieve the hardware-bound IDs for active mission context.
* **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) to track the lineage of context fragment IDs and their associated hardware-attestation status.

## 5. Alternatives Considered
* **Full Monologue Audit:** Rejected because it violates agent privacy and significantly increases token costs/latency.
* **Path-Based Isolation:** Rejected because it cannot protect against "Injected Context" that is already within the agent's authorized filesystem boundary but semantically unauthorized.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ZKIA is a core component of the Zero Trust Reasoning architecture, moving security from "who called it" to "why was it called."
* **Observability:** Detailed logs of derivation proofs and "Fragment Miss" events (where an agent uses un-attested context).

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
