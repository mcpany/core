# Design Doc: Intent-Aware GC (IAGC) Controller
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As LLM context windows scale to 1M+ tokens, aggressive "Context-Window Garbage Collection" (CWGC) has become a performance necessity. However, current pruning strategies are "Intent-Blind," often evicting critical mission-root instructions (Silent Anchors) while retaining high-entropy noise from specialist subagents. This leads to "Instruction Eviction" vulnerabilities and swarm divergence.

The IAGC Controller is designed to provide active governance over the context eviction process, ensuring that behavioral guardrails remain permanent while speculative thoughts are pruned based on real-time reasoning confidence.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically scale context eviction thresholds based on real-time "Reasoning Confidence" scores.
    * Provide mandatory "GC-Immune" pinning for mission-critical intent fragments.
    * Neutralize "Instruction Eviction" during aggressive token-saving cycles.
    * Reduce token costs by pruning low-confidence speculative reasoning paths.
* **Non-Goals:**
    * Modifying the underlying model's attention mechanism.
    * Managing long-term memory retrieval (this is handled by the ContextEngine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Agent Swarm (e.g., Gemini CLI Specialist)
* **Primary Goal:** Maintain 100% adherence to mission-root constraints while operating in a 1.5M token window.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent initiates a task with "GC-Immune" behavioral anchors.
    2. Subagents generate high-frequency reasoning traces during complex task execution.
    3. The IAGC Controller monitors the "Reasoning Confidence" headers (`x-gemini-reasoning-confidence`) for every fragment.
    4. As the context window nears its limit, the IAGC triggers a pruning cycle.
    5. IAGC identifies "Speculative Fragments" (confidence < 0.7) and prunes them aggressively.
    6. IAGC verifies that "GC-Immune" anchors remain pinned in the active attention window.
    7. The agent continues reasoning without losing its primary mission constraints.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        AG[Agent] -->|Reasoning Fragment| IAGC[IAGC Controller]
        IAGC -->|Analyze Confidence| CE[Confidence Evaluator]
        CE -->|High Confidence| P[Pin Fragment]
        CE -->|Low Confidence| S[Mark for Pruning]
        IAGC -->|Prune Event| PR[Pruning Engine]
        PR -->|Check Immune Anchors| IA[Immune Anchor Registry]
        IA -->|Retain| AG
    ```
* **APIs / Interfaces:**
    * `SetImmuneAnchor(fragmentID string, priority int)`: Mark a fragment as GC-Immune.
    * `GetEvictionStatus(contextID string) -> eviction metrics`: Retrieve real-time pruning statistics.
* **Data Storage/State:**
    * Eviction priorities and anchor status are stored in the hardware-attested Session Metadata.

## 5. Alternatives Considered
* **Static Token Budgeting:** Rejected because it doesn't account for the varying semantic importance of reasoning fragments.
* **LRU (Least Recently Used) Eviction:** Rejected because "Silent Anchors" (instructions provided at the start) are often the oldest fragments but the most important.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "GC-Immune" status can only be granted by the hardware-attested Mission Root.
* **Observability:** Pruning events and "Instruction-Loss" alerts are logged to the Visual Attention Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
