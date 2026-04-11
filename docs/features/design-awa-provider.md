# Design Doc: Attention-Weight Attestation (AWA) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents become more autonomous and are integrated into high-stakes environments, verifying the integrity of their reasoning process is paramount. While existing methods like hardware-attested reasoning traces provide a log of what was said, they do not provide visibility into the model's "intent" or what it was "paying attention" to during tool execution. Malicious actors or prompt injections can divert an agent's attention away from the "Mission-Root" guardrails toward unauthorized actions.

The Attention-Weight Attestation (AWA) Provider solves this by providing a hardware-attested summary of the model's attention-head distribution during a tool call. It allows a supervisor or an audit log to cryptographically verify that the agent was primarily driven by the verified mission-root fragments and not by high-entropy noise or injected instructions.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a cryptographically signed summary of attention-head weights per tool call.
    * Enable automated verification of "Attention Anchoring" to the mission-root.
    * Reduce the audit overhead by summarizing multi-gigabyte KV-cache data into verifiable attestation fragments.
    * Integrate with the SRM Provider to provide a complete "Cognitive Provenance" chain.
* **Non-Goals:**
    * Exposing raw attention weights (due to privacy and model IP concerns).
    * Modifying model weights or architecture.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Auditor
* **Primary Goal:** Verify that a code-generation tool call was not influenced by a malicious `GEMINI.md` file found in a cloned repository.
* **The Happy Path (Tasks):**
    1. The agent initiates a tool call to `run_shell_command`.
    2. The AWA Provider captures the attention-weight summary across the last 10 reasoning steps.
    3. The TPM signs the summary, linking it to the tool call's session token.
    4. The Auditor retrieves the AWA proof.
    5. The verification engine confirms that 85% of the attention weight was anchored to the "User Policy" fragment and <2% to the "Project Context" fragment.
    6. The tool call is marked as "Anchored" and approved for the permanent audit trail.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Tool Call] --> AWA[AWA Provider]
        AWA --> Capture[Capture Attention Meta-Weights]
        Capture --> Summarize[Summarize into Fragment-Bound Scores]
        Summarize --> Sign[TPM Attestation Signature]
        Sign --> Proof[AWA Proof Fragment]
        Proof --> Audit[Audit/Verification Hub]
    ```
* **APIs / Interfaces:**
    * `mcpany.awa.v1.AwaProvider`
    * `GetAwaProof(tool_call_id string) (Proof, error)`
* **Data Storage/State:**
    * Temporary attention-meta buffers; finalized proofs are stored in the hardware-attested Blackboard segment.

## 5. Alternatives Considered
* **Full Reasoning Trace Audit**: Rejected due to high latency and inability to detect "Instruction Shadowing" where the model produces "safe" text while being driven by "unsafe" intent.
* **Instruction Tuning for Safety**: Rejected as it cannot be verified in real-time by a third-party gateway like MCP Any.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AWA Proofs are hardware-bound and session-locked, preventing "Proof Replay" attacks.
* **Observability:** Integrated into the "Visual Attention Dashboard" as high-fidelity heatmaps.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
