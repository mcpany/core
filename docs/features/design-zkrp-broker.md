# Design Doc: ZK-Reasoning Proof (ZKRP) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Zero-Knowledge Reasoning Proofs (ZKRP) in Gemini CLI v1.0, agents can now provide cryptographic guarantees about their cognitive paths without revealing the raw reasoning traces. This is critical for multi-tenant swarms and vendor-client delegations where the client needs to verify that the vendor's agent followed safety protocols (e.g., "I did not attempt to access PII") but the vendor wants to protect their proprietary reasoning logic.

The ZKRP Broker in MCP Any provides the infrastructure to generate, proxy, and verify these proofs across the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Support the Gemini ZKRP v1.0 standard for reasoning path attestation.
    * Provide a high-performance verification engine to minimize the "Attestation Tax."
    * Integrate with the SRM Provider to anchor reasoning proofs to hardware-attested sessions.
    * Facilitate "Privacy-Preserving Delegation" where the mission root can verify subagent integrity without seeing raw logs.
* **Non-Goals:**
    * Implementing the underlying ZK-SNARK/STARK proving systems (we act as a broker/orchestrator).
    * Providing proofs for non-agentic computations.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Compliance Officer
* **Primary Goal:** Verify that a third-party "Tax Auditor" agent followed strict data-access boundaries without the third-party revealing their proprietary auditing reasoning.
* **The Happy Path (Tasks):**
    1. The Enterprise agent delegates a task to the third-party Auditor agent via an AMT tunnel.
    2. The Auditor agent completes the task and generates a ZKRP stating it followed the "Data Sovereignty Policy."
    3. The ZKRP Broker on the third-party node signs the proof and sends it to the Enterprise MCP Any instance.
    4. The Enterprise ZKRP Broker verifies the proof against the pre-shared policy manifest and hardware root.
    5. The Enterprise gateway accepts the tool results, cryptographically certain that safety rules were followed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent] -->|Generate CoT| Engine[Reasoning Engine]
        Engine -->|Raw Trace| Prover[ZK-Prover]
        Prover -->|ZKRP| Broker[ZKRP Broker]
        Broker -->|Proof + Public Inputs| Peer[Peer ZKRP Broker]
        Peer -->|Verify| Verifier[ZK-Verifier]
        Verifier -->|Success/Fail| Mission[Mission Root]
    ```
* **APIs / Interfaces:**
    * `zkrp.GenerateProof(trace, policy) -> ZKRP`: Generates a proof for a reasoning trace.
    * `zkrp.VerifyProof(proof, publicInputs) -> Bool`: Verifies an external proof.
    * `zkrp.RegisterPolicy(circuitID, schema)`: Maps a safety policy to a specific ZK circuit.
* **Data Storage/State:**
    * **Policy Circuit Registry:** Stores verification keys and circuit definitions for authorized safety protocols.
    * **Proof Cache:** Temporary storage for verified proofs to avoid redundant computation.

## 5. Alternatives Considered
* **Full Trace Auditing:** Rejected because it violates privacy and IP protections of specialized agent providers.
* **Hardware-Only Attestation (SRM):** Rejected because SRM proves *identity* and *integrity of the message*, but not the *content* of the reasoning logic. ZKRP bridges this gap.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Proofs must include hardware-attested nonces to prevent replay attacks across different mission phases.
* **Reasoning Exhaustion Defense:** To neutralize DoS attacks via proof flooding, the ZKRP Broker implements a "Token-Bucket Rate Limiter" for verification requests. Proofs from unauthenticated or low-reputation peers are automatically throttled to preserve the primary mission's reasoning budget.
* **Observability:** Verification latency and "Reasoning Stutters" are tracked via Prometheus metrics.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
