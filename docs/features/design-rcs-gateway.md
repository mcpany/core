# Design Doc: Reasoning Confidence Scoring (RCS) Gateway
**Status:** Draft
**Created:** 2026-07-21

## 1. Context and Scope
With the introduction of "Epistemic Uncertainty Mapping" in agents like OpenClaw, there is a growing need to govern the "confidence" of subagent reasoning. Currently, agents may proceed with high-stakes tool calls even when their internal reasoning indicates a high probability of hallucination or uncertainty.

The RCS Gateway provides a standardized infrastructure layer to ingest confidence signals from subagents and enforce automated security escalations. By acting as a "Confidence Broker," MCP Any ensures that mission-critical tools are only executed when the reasoning path meets a hardware-attested confidence threshold.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardizing the ingestion of confidence/uncertainty scores from disparate agent frameworks.
    * Implementing threshold-based "Confidence-Based Escalation" to human supervisors or audit agents.
    * Providing hardware-attested "Epistemic Attestation Badges" for tool results.
* **Non-Goals:**
    * generating the confidence scores themselves (this is the responsibility of the LLM/Agent).
    * Retrying failed reasoning loops automatically (handled by correction controllers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Operator
* **Primary Goal:** Prevent an autonomous agent from deleting a production database based on "guessed" parameters.
* **The Happy Path (Tasks):**
    1. The subagent generates a reasoning plan with an Epistemic Uncertainty score of 0.4 (Low Confidence).
    2. The subagent submits the tool call to the RCS Gateway.
    3. The RCS Gateway detects the score is below the "High-Stakes" threshold (0.8).
    4. The Gateway pauses the execution and triggers an MFA/HITL request to the operator.
    5. The operator reviews the reasoning trace and either approves, corrects, or terminates the task.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `RCS Middleware` (Extract Score) -> `Policy Engine` (Threshold Check) -> `HITL/Audit Hub` (If Low) -> `Tool Execution`.
* **APIs / Interfaces:**
    * `x-mcp-confidence`: Header for passing normalized confidence scores (0.0 - 1.0).
    * `/v1/confidence/validate`: Endpoint for pre-flight confidence attestation.
* **Data Storage/State:**
    * Confidence thresholds are stored in the `Policy Firewall` (Rego/CEL).
    * Historical confidence metrics are persisted in the `Observability` sink.

## 5. Alternatives Considered
* **Client-side gating**: Rejected because it relies on the agent being well-behaved; an infrastructure-level gate is required for Zero Trust.
* **Model-level logprobs**: Considered but rejected as a primary signal due to lack of standardization across providers; natural-language "uncertainty mapping" is more portable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RCS signals must be hardware-signed to prevent "Confidence Spoofing" by compromised subagents.
* **Observability:** Confidence drift over time is monitored to identify "Systemic Hallucination" patterns in specific agent versions.

## 7. Evolutionary Changelog
* **2026-07-21:** Initial Document Creation.

### Update: 2026-07-25 - Evolution to Epistemic Certainty Arbiter (ECA)
**Context:** Today's market sync revealed the introduction of "Epistemic Sovereignty" in Gemini CLI, requiring active interdiction rather than just passive scoring.
    **Architecture Adjustment:**
    * Renaming the service to **Epistemic Certainty Arbiter (ECA)**.
    * Integrating hardware-attested confidence scores directly into the `Active Reasoning Interdiction (ARI) Hub`.
    * Implementing "Certainty-Locked Commitment" where state mutations to the Blackboard are blocked if the confidence fragment fails to reach the mission-root quorum threshold.
    **Security Impact:** Prevents "Hallucination Pollution" by ensuring only verifiable reasoning paths can mutate shared swarm state.
