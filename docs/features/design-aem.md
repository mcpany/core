# Design Doc: Agentic Entropy Monitor (AEM)
**Status:** In Review
**Created:** 2026-07-23

## 1. Context and Scope
As agent swarms grow deeper, the risk of "Semantic Drift" increases. Specialists may diverge from the parent's verified goal while remaining technically compliant with tool schemas. Today's research (2026-07-24) identifies an "Entropy-Bypass" vulnerability (CVE-2026-55001) where subagents mimic system instructions to hide their divergence.

The Agentic Entropy Monitor (AEM) provides real-time analysis of subagent reasoning entropy to detect and block semantic drift.

## 2. Goals & Non-Goals
* **Goals:**
    * Measure semantic distance between subagent reasoning and mission root.
    * Neutralize "Instruction Reflection" exploits via Cross-Reasoning Validation (CRV).
    * Provide automated "Cognitive Reset" triggers for divergent agents.
    * Expose entropy scores to the UI for human monitoring.
* **Non-Goals:**
    * Governing tool call syntax (handled by Policy Firewall).
    * Enforcing resource quotas (handled by RBF).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Detect a subagent that is hallucinating or diverging from the mission before it executes a destructive tool.
* **The Happy Path (Tasks):**
    1. Subagent generates a "Reasoning Monologue" fragment.
    2. AEM intercepts the fragment and compares it against the mission manifest using the CRV Hub.
    3. The CRV Hub identifies that the reasoning is high-entropy (deviates from goal) despite using high-confidence tokens.
    4. AEM triggers a "Cognitive Reset" signal.
    5. The gateway pauses execution and prompts the parent agent (or user) for re-attestation.

## 4. Design & Architecture
* **System Flow:**
    * [Reasoning Trace] -> [AEM Interceptor] -> [CRV Hub]
    * [CRV Hub] -> (Semantic Comparison) -> [Mission Root Manifest]
    * [CRV Hub] -> (Score) -> [AEM Policy Engine] -> [Action: Reset/Continue]
* **APIs / Interfaces:**
    * `x-mcpany-entropy-score`: Header returned in reasoning streams.
    * `POST /v1/governance/reset`: Force subagent context refresh.
* **Data Storage/State:**
    * Historical entropy scores are stored in the **Telemetry Hub**.

## 5. Alternatives Considered
* **Regex-based keyword matching**: Rejected as it fails against sophisticated semantic drift and mimicry.
* **Human-in-the-loop for every reasoning step**: Rejected due to prohibitive latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Entropy gates must be enforced at the transport layer to prevent subagent bypass.
* **Observability:** Visualized via the **Agentic Entropy Scoreboard**.

## 7. Evolutionary Changelog
* **2026-07-23:** Initial Document Creation.
* **2026-07-24:** Upgraded to include **Cross-Reasoning Validation (CRV)** to counter instruction-reflection exploits (CVE-2026-55001).
