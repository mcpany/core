# Design Doc: Attention-Locked Reasoning Anchors (ALRA)
**Status:** Draft
**Created:** 2026-07-01

## 1. Context and Scope
In deep agent swarms, "Context Window Flooding" (CWF) has emerged as a critical stability and security risk. Subagents or malicious tools can inject high-entropy "noise" into the context window, causing the LLM to evict mission-critical instructions or "Mission-Root" intents from its attention span. Once the primary intent is evicted, the agent becomes susceptible to reasoning drift or instructions injected by subagents.

Attention-Locked Reasoning Anchors (ALRA) provide a hardware-bound mechanism to "pin" critical intent fragments at the attention layer. By utilizing specialized KV-cache management headers and TPM-signed anchors, ALRA ensures that mission-root instructions remain present in the model's attention window despite high-entropy noise injections.

## 2. Goals & Non-Goals
* **Goals:**
    * Protect "Mission-Root Intent" fragments from eviction during context window flooding.
    * Enforce attention-layer isolation between high-trust user instructions and low-trust subagent noise.
    * Provide hardware-attested proof that the reasoning path remains anchored to the verified intent.
    * Integrate with the Universal Multimodal Memory Bus (UMMB) for cross-framework attention governance.
* **Non-Goals:**
    * Modifying the underlying transformer architecture (ALRA works via existing attention-masking and KV-cache priority APIs).
    * Restricting subagent reasoning (only preventing it from evicting parent intent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Execute a long-running code refactoring task involving 5 specialized subagents without the primary goal ("Do not delete tests") being evicted from the attention window by subagent logs.
* **The Happy Path (Tasks):**
    1. The "Orchestrator" defines the mission-root and signs it with its TPM.
    2. ALRA creates "Attention Anchors" for the root intent fragments.
    3. As subagents generate high-volume logs and reasoning traces, the LLM context window reaches capacity.
    4. The ALRA middleware ensures that non-anchored fragments (subagent logs) are prioritized for eviction, while anchored intents are retained.
    5. The Orchestrator completes the task, and the model remains anchored to the "Do not delete tests" instruction until the session terminates.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Intent] --> B[ALRA Provider]
        B --> C[TPM Signing]
        C --> D[Attention-Locked Anchors]
        D --> E[LLM Attention Manager]
        F[Subagent Noise] --> G[Context Window]
        G --> E
        E --> H{Eviction Choice}
        H -->|Non-Anchored| I[Evict Noise]
        H -->|Anchored| J[Retain Intent]
    ```
* **APIs / Interfaces:**
    * `POST /v1/attention/anchor`: Creates an attention-locked fragment from a verified mission root.
    * `GET /v1/attention/status`: Returns the current attention-map visualization.
* **Data Storage/State:**
    * Anchors are stored in the hardware-attested segment of the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Constant Intent Injection:** Rejected because it consumes excessive tokens and increases reasoning latency.
* **Context Compression:** Rejected as it can lead to "Semantic Loss" where the compressed intent loses its original nuance.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALRA ensures that "Injected Context" cannot shadow the root authority by occupying the primary attention tiers.
* **Observability:** The "Visual Attention Dashboard" provides real-time heatmaps of driving fragments.

## 7. Evolutionary Changelog
* **2026-07-01:** Initial Document Creation.
