# Design Doc: Agentic Entropy Monitor (AEM)
**Status:** Draft
**Created:** 2026-04-01

## 1. Context and Scope
As AI agent swarms grow in complexity and autonomy, a critical failure mode known as "Semantic Drift" has emerged. Specialized subagents, while executing valid tool calls, may gradually diverge from the primary mission root's intent. Current security models focus on "what" a tool does, but fail to monitor "why" it is being called in the context of the broader mission.

The Agentic Entropy Monitor (AEM) provides a cross-framework observability and enforcement layer that measures the semantic divergence (entropy) of subagent reasoning traces. By providing a real-time "Coherence Score," MCP Any can proactively interdict rogue subagents before they commit unauthorized state changes or exhaust mission budgets.

## 2. Goals & Non-Goals
* **Goals:**
    * Measure semantic divergence between subagent reasoning and mission-root manifests.
    * Provide a standardized "Coherence Score" (0.0 to 1.0) for all connected agent frameworks.
    * Trigger automated "Cognitive Resets" or mission termination when entropy thresholds are exceeded.
    * Support hardware-attested logging of entropy scores for forensic auditing.
* **Non-Goals:**
    * Modifying the internal model weights of connected agents.
    * Providing natural language explanations for entropy spikes (handled by higher-level auditors).
    * Replacing existing per-tool capability gates (AEM is a behavioral layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Detect and contain a "hallucinatory loop" in a specialized code-refactoring subagent before it begins deleting valid source files.
* **The Happy Path (Tasks):**
    1. The Architect defines an entropy threshold (e.g., 0.7) in the Mission Manifest.
    2. The Refactoring Subagent begins a deep reasoning chain.
    3. The AEM continuously ingests the subagent's reasoning fragments via the `SRM Provider`.
    4. At reasoning depth 15, the AEM detects a divergence where the subagent stops referencing the "safety guardrails" in the mission root.
    5. The Coherence Score drops to 0.65.
    6. The AEM immediately revokes the subagent's `fs:write` capability and signals a `COGNITIVE_RESET` to the supervisor agent.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Subagent] -> (Reasoning Trace) -> [SRM Provider]
                                            |
                                            v
    [Mission Manifest] -> (Reference Intent) -> [AEM Engine]
                                            |
                                            v
    [Policy Engine] <- (Coherence Score) <- [AEM Engine]
    ```
* **APIs / Interfaces:**
    * `GET /v1/entropy/{session_id}`: Returns real-time coherence metrics.
    * `POST /v1/policy/entropy-gate`: Configures threshold-based interdiction rules.
* **Data Storage/State:**
    * Entropy scores are stored in a time-series buffer within the Shared KV Store (Blackboard).
    * High-entropy events are persisted to the Hardware-Attested Audit Log.

## 5. Alternatives Considered
* **Manual HITL Review:** Rejected due to the "Machine-Speed" requirement of autonomous swarms; humans cannot review 100+ reasoning fragments per second.
* **Keyword-Based Filtering:** Rejected as it fails to capture the nuanced semantic meaning of complex reasoning paths.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The AEM Engine itself must be isolated. Subagents cannot "spoof" their own entropy scores as the ingestion happens via the hardware-bound SRM Provider.
* **Observability:** Entropy spikes are visualized in the `Swarm Topology Monitor` and the `Global Agent Activity Map`.

## 7. Evolutionary Changelog
* **2026-04-01:** Initial Document Creation.

### Update: 2026-07-25 - Entropy-Aware Governance Integration
**Context:** Today's market sync revealed that passive reasoning monitoring is insufficient against high-velocity "Reasoning Drift" in autonomous swarms.
**Architecture Adjustment:**
* Integrating hardware-attested Reasoning Entropy Attestation (REA) scores into the AEM Engine.
* Adding a `REA Validator` component to Section 4 to perform real-time verification of model-provided entropy metrics.
**Security Impact:** Prevents hallucination-driven tool execution by revoking capabilities before high-entropy reasoning leads to state corruption.
