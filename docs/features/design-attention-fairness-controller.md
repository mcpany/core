# Design Doc: Attention-Fairness Controller (AFC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Dynamic Attention Shifting (DAS) in major model frameworks, the ability to prioritize "Mission-Root" reasoning has become a core model capability. However, this has opened a new attack vector: **Attention-Starvation DoS**. A specialist subagent or tool can intentionally trigger high-uncertainty reasoning loops or "Reasoning Storms" to monopolize the model's limited attention tokens and compute priority, effectively blinding the parent agent or mission-root from detecting drift or interdicting unauthorized actions.

MCP Any needs to solve this by moving attention governance from the model (which can be coerced) to the infrastructure. The AFC will act as the kernel-level scheduler for the model's attention window, ensuring fair distribution and preventing subagent monopoly.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce hardware-attested "Attention Quotas" (token density and compute priority) for every subagent.
    * Detect and throttle "Reasoning Storms" (high-frequency, high-uncertainty reasoning loops) before they exhaust the mission budget.
    * Provide "Mission-Root Preemption" allowing the parent agent to forcefully reclaim 100% of the attention window at any time.
* **Non-Goals:**
    * Modifying the internal weights or attention mechanisms of the LLM itself.
    * Managing the content of the reasoning traces (handled by ARI/AID).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent a specialized "Code Debugger" subagent from monopolizing the attention window during a complex refactor, which would prevent the "Security Auditor" subagent from running in parallel.
* **The Happy Path (Tasks):**
    1. Orchestrator defines an AFC Policy assigning 40% attention quota to the Debugger and 60% to the Auditor.
    2. The Debugger subagent encounters a recursive bug and attempts to trigger a "Deep Reasoning" loop (DAS).
    3. The AFC monitors the Debugger's token density in real-time.
    4. When the Debugger exceeds its 40% quota, the AFC injects "Wait-States" and throttles its request frequency.
    5. The Auditor subagent retains its 60% attention window, successfully identifying a vulnerability in the Debugger's proposed code.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Request] -> [AFC Quota Check] -> [DAS Header Injection] -> [LLM Gateway] -> [Reasoning Output] -> [AFC Usage Update]
* **APIs / Interfaces:**
    * `POST /v1/afc/policies`: Define attention weights for subagent roles.
    * `GET /v1/afc/status`: Real-time telemetry of attention utilization across the mesh.
* **Data Storage/State:**
    * In-memory "Attention Shards" (Redis/CRDT) to track per-session token density and uncertainty metrics.

## 5. Alternatives Considered
* **Model-Side Throttling:** Rejected because malicious specialists can manipulate the "Uncertainty" signals that drive model-side DAS.
* **Hard Token Limits:** Rejected because they don't account for "Attention Density" (how much the model "focuses" on a fragment), only total count.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** AFC quotas are cryptographically bound to the Mission-Root; subagents cannot "request" more attention without parent attestation.
* **Observability:** Real-time Heatmap in the UI showing "Attention Occupancy" per subagent.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
