# Design Doc: Automated Reasoning Anchor (ARA) Protocol
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Long-running agent missions often suffer from "Anchor Drift" and "Instruction Eviction." As context windows shift or are pruned by aggressive Garbage Collection (GC) logic (like Gemini's CWGC), critical mission-root guardrails can be lost, leading to Hallucination or Intent Divergence. The ARA Protocol provides an automated cognitive governance layer that monitors reasoning coherence and dynamically nominates, refreshes, and pins mission-critical context fragments as GC-Immune anchors.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically identify mission-critical reasoning fragments in real-time.
    * Refresh and pin nominated anchors to prevent eviction during context-window shifts.
    * Neutralize "Anchor Drift" via periodic semantic alignment checks.
    * Provide a standardized interface for agents to "nominate" high-confidence fragments.
* **Non-Goals:**
    * Replacing manual mission-root pinning (ARA is an augmentation layer).
    * Modifying model-side attention mechanisms directly (ARA operates via context-injection/headers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Long-Haul Autonomous Agent
* **Primary Goal:** Maintain strict security guardrails during a 24-hour continuous code migration task.
* **The Happy Path (Tasks):**
    1. Agent initializes mission with primary guardrails (pinned by default).
    2. As task evolves, agent generates a high-confidence summary of new sub-tasks.
    3. ARA Monitor evaluates the summary's coherence score and mission-alignment.
    4. ARA Protocol nominates this fragment as a "Dynamic Anchor."
    5. MCP Any injects the dynamic anchor into the next attention window, marked as `x-mcpany-anchor-priority: high`.

## 4. Design & Architecture
* **System Flow:**
    `Reasoning Stream` -> `ARA Coherence Monitor` -> `Anchor Nomination` -> `Context Injector` -> `Attention Window`
* **APIs / Interfaces:**
    * `AnchorService`: `NominateFragment(fragmentID string, coherence float64) error`, `GetActiveAnchors(sessionID string) ([]Anchor, error)`
* **Data Storage/State:**
    * Anchors are stored in the `Universal Episodic Graph` (UEG) to ensure cross-session persistence and lineage tracking.

## 5. Alternatives Considered
* **Static Context Pinning**: Rejected because it leads to "Attention Bloat" where irrelevant instructions consume the window.
* **Frequent Summarization**: Rejected because summarization often "ghosts" subtle security constraints.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Dynamic anchors are validated against the hardware-attested mission manifest before pinning.
* **Observability:** Active anchors and their coherence scores are visualized in the "GC-Immune Anchor Visualizer" UI component.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
