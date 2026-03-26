# Design Doc: Delegation Attestation Layer
**Status:** Draft
**Created:** 2026-04-14

## 1. Context and Scope
As agent swarms grow in complexity, the "Manual Review Bottleneck" (where 44% of users manually verify inter-agent flows) has become a primary scaling limitation. The Delegation Attestation Layer (DAL) provides a secure, automated way to evaluate the safety and legitimacy of A2A task proposals. It acts as the "Decision Support System" for the A2A Messaging Hub, generating verifiable safety proofs that allow for higher degrees of autonomous operation.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically evaluate A2A task proposals against historical reputation and local Zero-Trust policies.
    * Generate "Safety Proofs" (signed attestations) for low-risk inter-agent delegations.
    * Provide a standardized "Trust Signal" for the A2A Messaging Hub to surface for either autonomous or human approval.
    * Integrate with the "Cross-Framework Skill Reputation Engine" for real-time risk scoring.
* **Non-Goals:**
    * Automatically approving high-risk tasks (e.g., shell execution, large financial transfers).
    * Providing a guarantee of task *success* (it only evaluates *safety* and *intent alignment*).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Swarm Operator
* **Primary Goal:** Enable a trusted subagent to perform a routine data-transformation task without manual user intervention for every handoff.
* **The Happy Path (Tasks):**
    1. Parent Agent sends a task proposal to the A2A Messaging Hub.
    2. The Hub routes the proposal to the Delegation Attestation Layer (DAL).
    3. DAL retrieves the subagent's reputation score and verifies the "Intent Chain."
    4. DAL confirms the task aligns with the mission's "Safe Operation Baseline."
    5. DAL generates and signs a "Safety Proof."
    6. The A2A Messaging Hub sees the proof and automatically delegates the task to the subagent.
    7. The user is notified of the delegation in the "Session Timeline" but was not required to manually approve it.

## 4. Design & Architecture
* **System Flow:**
    `[A2A Hub] -> (Proposal) -> [DAL] -> (Reputation + Policy) -> [Safety Proof] -> [A2A Hub]`
* **APIs / Interfaces:**
    * `EvaluationService`: `EvaluateProposal(proposal TaskProposal) (SafetyProof, error)`
    * `SafetyProof`: A signed JSON object containing the risk score, reasoning fragments, and a cryptographic attestation.
* **Data Storage/State:**
    * Caches reputation data from the UAB Reputation Explorer.
    * Stores "Safe Operation Baselines" for different mission intents.

## 5. Alternatives Considered
* **Binary Allow-Lists:** Rejected as too rigid; it cannot handle the dynamic nature of agentic task bidding.
* **Purely Human Approval:** Rejected as the primary scaling bottleneck we are trying to solve.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The DAL itself must run in a restricted sandbox to prevent "Attestation Hijacking." Safety proofs are short-lived and session-bound.
* **Observability:** DAL's reasoning for every evaluation is logged and visible in the "A2A Governance & Security Center."

## 7. Evolutionary Changelog
* **2026-04-14:** Initial Document Creation.

### Update: 2026-04-15 - Integration with Context Sidecars
**Context:** Today's market sync revealed a new "Shadow Context Injection" pattern where malicious state fragments can bypass safety evaluations if the evaluator lacks full context visibility.
**Architecture Adjustment:**
*   **Context-Aware Scoring:** The `EvaluationService` is being updated to optionally mount "Context Sidecars" during evaluation. This allows the DAL to perform safety proofs against the *actual* state the subagent will receive, not just the task description.
*   **Shadow Fragment Detection:** Integrating a WASM-based "State Sanitizer" directly into the evaluation pipeline to detect hidden instructions in Binary State Handoffs.
**Security Impact:** Prevents "Side-Channel Intent Hijacking" where a subagent is coerced into an unsafe action via malicious context fragments that appear benign to traditional evaluators.
