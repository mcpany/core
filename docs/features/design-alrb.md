# Design Doc: Attention-Locked Reasoning Budgets (ALRB)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
"Attention-Density" attacks represent a new class of DoS vulnerabilities in AI agents. A malicious subagent can inject high-entropy, plausible-sounding "noise" into the context window. This noise can evict critical instructions or cause the parent agent to consume its entire "Reasoning Effort" (e.g., `x-gemini-reasoning-effort`) processing irrelevant fragments.

ALRB mitigates this by cryptographically binding reasoning budgets to verified mission anchors. It ensures that a subagent can only consume its assigned budget for reasoning traces that remain "pinned" within the hardware-locked attention tier of the LLM. If reasoning drifts into "Noise Tiers," the budget is automatically halted.

## 2. Goals & Non-Goals
* **Goals:**
    * Bind `x-reasoning-effort` headers to verified attention anchors (HAAL).
    * Implement a "Budget Halt" trigger for high-entropy/drifted reasoning.
    * Provide a granular allocation model for reasoning effort per intent branch.
    * Neutralize "Attention-Density" exhaustion attacks.
* **Non-Goals:**
    * Restricting total token usage (handled by standard token budgets).
    * Evaluating the correctness of reasoning.

## 3. Critical User Journey (CUJ)
* **User Persona:** Resource-Conscious Enterprise Administrator
* **Primary Goal:** Prevent a specialized "Code Auditor" agent from exhausting the team's reasoning budget by "hallucinating" infinitely deep trace logs.
* **The Happy Path (Tasks):**
    1. Administrator sets a "Reasoning Anchor" for the Code Audit mission.
    2. The Auditor agent starts reasoning; ALRB tracks its attention mapping in real-time.
    3. The Auditor attempts a "Noise Injection" to prolong its session.
    4. ALRB detects that the reasoning is no longer anchored to the mission intent.
    5. ALRB forcefully terminates the `x-reasoning-effort` lease.
    6. The Auditor fails gracefully without exhausting the global pool.

## 4. Design & Architecture
* **System Flow:**
    [Agent Reasoning] -> [Attention Map (HAAL)] -> [ALRB Monitor]
    [ALRB Monitor] --(Drift Detected)--> [Budget Firewall] --(Halt)--> [Reasoning Engine]
* **APIs / Interfaces:**
    * `POST /v1/alrb/allocate`: Allocates a reasoning lease to an intent branch.
    * `GET /v1/alrb/status`: Returns real-time anchoring and budget consumption scores.
* **Data Storage/State:**
    Budget leases and anchoring maps are stored in the secure SRM (Signed Reasoning Monologue) provider.

## 5. Alternatives Considered
* **Static Timeouts:** Rejected because they don't account for varying reasoning intensity.
* **Semantic Analysis of Outputs:** Rejected as it's too slow for real-time budget gating.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** ALRB ensures that subagents cannot "squat" on reasoning resources.
* **Observability:** Real-time "Reasoning-Density Heatmap" in the UI.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
