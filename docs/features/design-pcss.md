# Design Doc: Pre-Commit Speculative Sanitizer (PCSS)
**Status:** Draft
**Created:** 2026-06-04

## 1. Context and Scope
Speculative reasoning allows agents to prepare for potential future states, significantly reducing latency. However, malicious subagents can inject "Poisoned Fragments" into these speculative buffers. Since these fragments are ingested before final commitment, they bypass standard post-hoc validation. PCSS provides real-time semantic sanitization of these speculative buffers to ensure integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic scanning of speculative context fragments.
    * Neutralize imperative instructions hidden in speculative data.
    * Ensure no "Poisoned Fragments" are ingested into the primary reasoning engine.
* **Non-Goals:**
    * Validating the final committed state (handled by existing post-commit validators).
    * Predicting all possible agent intents.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Frequency Swarm Orchestrator
* **Primary Goal:** Prevent a specialized subagent from hijacking the parent's speculative reasoning branch.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a speculative reasoning branch.
    2. Subagent provides data to be used in the speculative buffer.
    3. PCSS intercepts the data fragment before ingestion.
    4. PCSS performs semantic analysis against the mission-root intent.
    5. PCSS sanitizes any detected "Poisoned Fragments".
    6. Sanitized data is safely ingested into the speculative buffer.

## 4. Design & Architecture
* **System Flow:**
  [Subagent Data] -> [Speculative Buffer] -> (PCSS Interception) -> [Semantic Scanner] -> [Sanitized Buffer] -> [Reasoning Engine]
* **APIs / Interfaces:**
    * `SpeculativeBuffer.Interpose(fragment)`: Hook for sanitization.
    * `PCSS.Analyze(fragment, missionRoot)`: Main analysis engine.
* **Data Storage/State:**
    * Temporary storage of speculative fragments in an isolated "Grey Zone" until sanitized.

## 5. Alternatives Considered
* **Post-Commit Validation:** Rejected because by the time validation happens, the reasoning engine has already been influenced by the poisoned fragment.
* **Complete Isolation:** Rejected because it defeats the purpose of speculative reasoning (sharing context for speed).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PCSS operates as a mandatory gate for all speculative ingestion.
* **Observability:** Logs all detected and neutralized poisoned fragments for audit.

## 7. Evolutionary Changelog
* **2026-06-04:** Initial Document Creation.
