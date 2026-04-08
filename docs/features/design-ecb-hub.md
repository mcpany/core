# Design Doc: Epistemic Circuit Breaker (ECB) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Decision-Path Poisoning," autonomous agents are increasingly susceptible to silently approving high-risk transactions or system changes based on corrupted model outputs. Traditional tool-gating is insufficient because the malicious intent is often "smuggled" through a series of seemingly benign reasoning steps.

The ECB Hub introduces automated, intent-aware checkpoints into the agent's decision path. It acts as an "Epistemic Safety Valve," requiring higher levels of attestation (e.g., multi-agent consensus or human-in-the-loop) when the agent's reasoning confidence drops or when the cumulative risk of a workflow exceeds a predefined safety threshold.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically inject safety checkpoints into multi-step agent workflows.
    * Evaluate reasoning confidence scores from subagents to trigger escalations.
    * Orchestrate multi-agent quorums (MAQ) for high-stakes decision points.
    * Provide hardware-attested "Confidence Proofs" for approved actions.
* **Non-Goals:**
    * Modifying the underlying model's temperature or sampling parameters.
    * Blocking every tool call (it only triggers on "Decision Points" identified by the policy engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Financial Systems Compliance Officer
* **Primary Goal:** Ensure that a "Treasury Specialist Agent" cannot authorize a wire transfer over $1M without a secondary reasoning audit, even if the primary agent's CoT looks plausible.
* **The Happy Path (Tasks):**
    1. The agent decides to initiate a $1.2M wire transfer.
    2. The ECB Hub intercepts the "Decision Point" based on the transaction value and mission constraints.
    3. ECB Hub evaluates the agent's "Epistemic Confidence" (RCS).
    4. Since the amount is P0-critical, ECB Hub triggers a "Consensus Quorum."
    5. Two independent "Auditor Agents" review the reasoning trace and mission root.
    6. Both auditors provide hardware-attested approval tokens.
    7. ECB Hub issues an "Epistemic Attestation Badge" and allows the tool call to proceed.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Decision] -> [ECB Evaluator] -> {Risk/Confidence Check} -> [Quorum/MFA Trigger] -> [Attestation Issue] -> [Execution]`
* **APIs / Interfaces:**
    * `mcpany.ecb.v1.CheckpointService`
    * `rpc EvaluateDecision(DecisionContext) returns (ActionPolicy)`
* **Data Storage/State:**
    * Risk thresholds and checkpoint policies are stored in the hardware-locked (TPM) secure enclave.

## 5. Alternatives Considered
* **Static Thresholds (Rejected):** Fails to account for the nuances of agent reasoning and intent.
* **Pure HITL (Rejected):** Unscalable for high-velocity autonomous systems.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ECB Hub relies on the hardware-attested mission root to prevent "Consensus Hijacking."
* **Observability:** Checkpoint triggers and quorum results are visualized in the "ASH Consensus Dashboard."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
