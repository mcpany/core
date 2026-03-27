# Design Doc: Recursive Hallucination Monitor (RHM)

**Status:** Draft
**Created:** 2026-06-20

## 1. Context and Scope
As agent swarms grow in depth, sub-instructions are prone to "Recursive Hallucination," where the original mission-root intent is semantically degraded. RHM provides a lineage-aware monitoring service to detect and mitigate this drift.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect semantic degradation in deep sub-instruction chains.
    * Provide "Mission Alignment Scores" for every delegation hop.
    * Trigger automated reconstruction if alignment falls below a threshold.
* **Non-Goals:**
    * Replacing the reasoning engine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Ensure a 5th-degree subagent is still executing the user's original request.
* **The Happy Path (Tasks):**
    1. A deep delegation chain is formed (Hops > 3).
    2. RHM intercepts each HAIL-signed instruction fragment.
    3. RHM compares the sub-instruction embedding against the Mission-Root TPM intent.
    4. RHM calculates a "Hallucination Delta."
    5. If the delta is within bounds, execution proceeds; otherwise, HAIR is triggered.

## 4. Design & Architecture
* **System Flow:** Mission-Root -> Sub-Instruction -> HAIL Signer -> RHM (Alignment Check) -> Teammate.
* **APIs / Interfaces:** `POST /v1/rhm/evaluate`.

## 5. Alternatives Considered
* **Periodic Re-summarization**: Rejected as it can introduce its own hallucinations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RHM evaluation is required for any delegation beyond Hop 2.
* **Observability:** Alignment scores are visualized in the Hallucination Drift Alert Hub.

## 7. Evolutionary Changelog
* **2026-06-20:** Initial Document Creation.
