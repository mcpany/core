# Design Doc: Layer-7 Semantic Inspection Hub (L7SIH)
**Status:** Draft
**Created:** 2026-06-11

## 1. Context and Scope
With the rise of Reasoning Entropy Exhaustion (REE) attacks, autonomous agent swarms are increasingly vulnerable to being overwhelmed by mission-irrelevant, high-entropy coordination noise. Standard transport-layer firewalls cannot detect these attacks because the traffic is semantically valid but mission-destructive. L7SIH provides deep semantic inspection to filter and prioritize agent coordination.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect and throttle high-entropy semantic noise in inter-agent messages.
    * Provide hardware-attested priority for mission-root intent fragments.
    * Enforce reasoning-effort quotas per mission branch.
* **Non-Goals:**
    * Replace existing transport-layer encryption (mTLS).
    * Perform generic LLM safety filtering (handled by separate providers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Maintain mission stability despite 3 subagents being compromised and injecting REE noise.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes mission with a hardware-attested "Mission Root Intent."
    2. Subagents submit coordination messages via the L7SIH.
    3. L7SIH evaluates semantic entropy against the mission root.
    4. Compromised subagent noise is identified as "High Entropy / Low Alignment."
    5. L7SIH automatically throttles the noisy subagent and alerts the mission root.

## 4. Design & Architecture
* **System Flow:**
    [Subagent A] -> [L7SIH (Entropy Scan)] -> [Mission Root]
                       ^
                       |
               [Attestation Provider]
* **APIs / Interfaces:**
    * `Inspect(IntentFragment, ContextShard) -> AnalysisResult`
    * `EnforceQuota(AgentID, TokenBudget) -> Status`
* **Data Storage/State:**
    Uses the Shared KV Store (Blackboard) for real-time entropy baselines.

## 5. Alternatives Considered
* **Rate Limiting (L4):** Rejected. Cannot distinguish between high-value dense reasoning and malicious noise.
* **Manual HITL:** Rejected. REE attacks occur at machine speeds, too fast for human intervention.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** L7SIH requires TPM-signed identity fragments for every inspection request.
* **Observability:** Logs "Semantic Drift" scores for every agent in the swarm.

## 7. Evolutionary Changelog
* **2026-06-11:** Initial Document Creation.
