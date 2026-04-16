# Design Doc: Reasoning-Responsive Resource Allocation (RRRA) v2
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents move toward "Deep Thought" and reinforcement-learning-driven reasoning, their resource consumption (tokens, compute time) becomes highly non-linear. Traditional static budgets lead to "Cognitive Stall," where an agent is terminated mid-reasoning because it hit a fixed threshold.

MCP Any needs to implement RRRA v2 to act as a "Reasoning Governor." This system will detect real-time reasoning-intensity signals (e.g., `x-gemini-reasoning-effort`) and dynamically adjust the available budget to ensure mission-critical reasoning paths can reach convergence.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically scale token and compute budgets based on LLM-emitted reasoning signals.
    * Prioritize mission-root reasoning over low-stakes specialist sub-tasks.
    * Provide hardware-attested budget "leases" that prevent subagent budget hijacking.
* **Non-Goals:**
    * Explicitly optimizing the model's internal reasoning logic.
    * Providing infinite budgets (limits will still be anchored to the mission root).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (Enterprise)
* **Primary Goal:** Ensure a complex code-refactoring mission completes even if the agent requires 3x the usual reasoning tokens for a specific architectural decision.
* **The Happy Path (Tasks):**
    1. The agent identifies a complex circular dependency and signals "High Reasoning Intensity."
    2. The RRRA Adapter intercepts the signal and cross-references it with the mission-root's "Deep Thought" policy.
    3. The Adapter issues an ephemeral budget expansion lease.
    4. The agent completes the reasoning loop without being throttled.
    5. The expanded budget is automatically reclaimed upon task completion.

## 4. Design & Architecture
* **System Flow:**
    `LLM -> [Intensity Signal] -> RRRA Adapter -> [Budget Check/Update] -> Gateway Resource Controller`
* **APIs / Interfaces:**
    * `ReasoningIntensitySignal`: Incoming header/metadata payload.
    * `ExpandBudgetLease(session_id, delta)`: Internal service call to the RBF.
* **Data Storage/State:**
    * State is managed in-memory within the RRRA session context and periodically persisted to the Blackboard for mission-wide budget reconciliation.

## 5. Alternatives Considered
* **Static Over-provisioning:** Rejected because it leads to "Resource Squatting" where a single agent exhausts the entire pool on a non-critical loop.
* **Manual User Re-approval:** Rejected as it breaks autonomous swarms and introduces human-in-the-loop latency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevent "Budget Hijacking" by mandating that only hardware-attested reasoning traces can trigger a budget expansion.
* **Observability:** Implement a "Budget Timeline" in the UI to visualize when and why resource scaling occurred.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
