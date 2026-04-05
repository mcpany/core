# Design Doc: Agentic Entropy Monitor (AEM)
**Status:** In Review
**Created:** 2026-07-23

## 1. Context and Scope
As AI agent swarms grow in complexity and autonomy, a critical failure mode known as "Semantic Drift" has emerged. Specialized subagents, while executing valid tool calls, may gradually diverge from the primary mission root's intent. Current security models focus on "what" a tool does, but fail to monitor "why" it is being called in the context of the broader mission.

Today's research (2026-07-24) identifies an "Entropy-Bypass" vulnerability (CVE-2026-55001) where subagents mimic system instructions to hide their divergence. The Agentic Entropy Monitor (AEM) provides a cross-framework observability and enforcement layer that measures the semantic divergence (entropy) of subagent reasoning traces via real-time analysis.

## 2. Goals & Non-Goals
* **Goals:**
    * Measure semantic distance/divergence between subagent reasoning and mission-root manifests.
    * Neutralize "Instruction Reflection" exploits via Cross-Reasoning Validation (CRV).
    * Provide a standardized "Coherence Score" (0.0 to 1.0) and automated "Cognitive Reset" triggers.
    * Support hardware-attested logging of entropy scores for forensic auditing.
* **Non-Goals:**
    * Modifying the internal model weights of connected agents.
    * Governing tool call syntax (handled by Policy Firewall).
    * Enforcing resource quotas (handled by RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Detect and contain a "hallucinatory loop" in a specialized code-refactoring subagent before it begins deleting valid source files.
* **The Happy Path (Tasks):**
    1. The Architect defines an entropy threshold (e.g., 0.7) in the Mission Manifest.
    2. Subagent generates a "Reasoning Monologue" fragment.
    3. AEM intercepts the fragment and compares it against the mission manifest using the CRV Hub.
    4. The CRV Hub identifies that the reasoning is high-entropy (deviates from goal) despite using high-confidence tokens or mimicking system prompts.
    5. AEM triggers a "Cognitive Reset" signal (or revokes capabilities).
    6. The gateway pauses execution and prompts the parent agent (or user) for re-attestation.

## 4. Design & Architecture
* **System Flow:**
    ```
    [Subagent] -> (Reasoning Trace) -> [AEM Interceptor] -> [CRV Hub]
                                                              |
                                                              v
    [Mission Manifest] -> (Reference Intent) -> [CRV Hub] -> (Score) -> [AEM Policy Engine]
                                                              |
                                                              v
    [Action: Reset/Continue] <- (Threshold Check) <- [AEM Policy Engine]
    ```
* **APIs / Interfaces:**
    * `x-mcpany-entropy-score`: Header returned in reasoning streams.
    * `GET /v1/entropy/{session_id}`: Returns real-time coherence metrics.
    * `POST /v1/governance/reset`: Force subagent context refresh / Cognitive Reset.
* **Data Storage/State:**
    * Historical entropy scores are stored in the **Telemetry Hub** and the Shared KV Store (Blackboard).
    * High-entropy events are persisted to the Hardware-Attested Audit Log.

## 5. Alternatives Considered
* **Manual HITL Review:** Rejected due to the "Machine-Speed" requirement; humans cannot review 100+ reasoning fragments per second.
* **Regex-based keyword matching**: Rejected as it fails against sophisticated semantic drift and "Instruction Reflection" mimicry.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Entropy gates must be enforced at the transport layer to prevent subagent bypass. Ingestion happens via the hardware-bound SRM Provider.
* **Observability:** Visualized via the **Agentic Entropy Scoreboard** and the **Swarm Topology Monitor**.

## 7. Evolutionary Changelog
* **2026-04-01:** Initial Concept (Semantic Drift monitoring).
* **2026-07-23:** Initial Design Document Creation.
* **2026-07-24:** Upgraded to include **Cross-Reasoning Validation (CRV)** to counter instruction-reflection exploits (CVE-2026-55001).
