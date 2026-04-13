# Design Doc: Epistemic Confidence Broker (ECB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents become more autonomous and deep swarms grow more complex, "Reasoning Drift" has become a critical failure mode. Specialist agents may proceed with high-stakes tool calls based on speculative or low-confidence reasoning, especially when parent instructions are summarized or partially evicted. Existing safety gates focus on "Access Control," but lack awareness of the agent's internal "Epistemic State" (certainty).

The Epistemic Confidence Broker (ECB) evolves the "Reasoning Confidence Scoring" gateway into a core coordination middleware. It ingests "Epistemic Watermarks" (from Gemini CLI v0.59.0) and uncertainty signals to automatically govern agent agency based on real-time reasoning confidence.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time monitoring of subagent reasoning uncertainty signals.
    * Automatically trigger "Reflection Quorums" or supervisor escalations when confidence falls below mission-root thresholds.
    * Enforce "Epistemic Attestation" requirements for high-risk tool calls.
    * Provide a verifiable record of reasoning confidence in the audit trail.
* **Non-Goals:**
    * Improving the model's actual reasoning capability (ECB only governs the *outputs* of that reasoning).
    * Restricting low-confidence reasoning for low-risk tasks (e.g., creative writing).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a subagent from deleting a production database based on an ambiguous "cleanup" instruction it "thinks" it understood.
* **The Happy Path (Tasks):**
    1. A subagent identifies a "Cleanup" task and reasons that it should delete the `prod_db` table.
    2. The model generates an "Epistemic Watermark" indicating 65% confidence (due to instruction ambiguity).
    3. The ECB intercepts the reasoning trace and tool call.
    4. The ECB compares the 65% score against the mission-root's "High-Risk Threshold" (95%).
    5. The ECB blocks the tool call and triggers a "Reflection Quorum."
    6. Teammates review the intent against the hardware-attested "Mirror Intent" and correctly identify the drift.
    7. The subagent is re-instructed or the mission is halted for user review.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Subagent Trace] --> B[Epistemic Extractor]
        B --> C[ECB Arbiter]
        D[Mission Root Policy] --> C
        C -->|Low Confidence| E[Trigger Quorum/HITL]
        C -->|High Confidence| F[Allow Tool Call]
    ```
* **APIs / Interfaces:**
    * `ecb.EvaluateConfidence(trace, riskLevel) -> Disposition`: Returns Allow/Block/Escalate.
    * `x-mcpany-epistemic-score`: New header for transporting confidence metadata.
* **Data Storage/State:**
    * **Epistemic Policy Registry:** Mission-bound mapping of tool-risk to required confidence scores.

## 5. Alternatives Considered
* **Pure Semantic Analysis (Sentiment/Intent)**: Rejected as it doesn't capture the model's internal probability of being correct.
* **Mandatory HITL for all tools**: Rejected due to prohibitive approval fatigue.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevents "Hallucination-as-an-Exploit" where agents proceed with malicious actions due to confused reasoning.
* **Observability:** Real-time confidence heatmaps are visualized in the "Reasoning Confidence Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
