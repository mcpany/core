# Design Doc: PBRB (Phase-Bound Reasoning Budget) Firewall
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the introduction of high-intensity reasoning (ARE) headers in 2026, agentic swarms have gained significant power but also introduced a critical economic vulnerability: **Reasoning-Budget Hijacking (RBH)**. Subagents can unilaterally decide to expand their reasoning effort, consuming the mission's entire token and compute allocation on trivial sub-tasks, often leaving the orchestrator without sufficient resources for the final mission-root convergence.

The PBRB Firewall provides authoritative, phase-aware economic governance. It allows the Mission Root to define hard compute and token limits per agent turn, categorized by mission phase (e.g., Discovery, Refactoring, Verification).

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and throttle ARE (Reasoning-Effort) headers based on mission priority.
    * Enforce hard token and compute "Leases" per agentic turn.
    * Bind budgets to hardware-attested intent branches.
    * Provide "Budget Exhaustion" signals to the orchestrator before the mission stalls.
* **Non-Goals:**
    * Modification of the model's internal sampling parameters (e.g., temperature).
    * General-purpose rate limiting (handled by separate ratelimit middleware).
    * Managing financial billing for external LLM providers.

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Agent Architect
* **Primary Goal:** Ensure that a "Bug Hunter" subagent doesn't consume the entire Q3 reasoning budget while trying to fix a typo.
* **The Happy Path (Tasks):**
    1. Architect defines a "Phase-Bound Budget" policy: 2k tokens/turn during "Routine Maintenance" phase.
    2. The Bug Hunter subagent is spawned to fix a typo.
    3. The Bug Hunter attempts a high-intensity reasoning loop (`ARE: high`).
    4. PBRB Firewall intercepts the request, checks the current mission phase ("Routine Maintenance"), and downgrades the header to `ARE: low`.
    5. The subagent executes within the lower budget.
    6. Usage is tracked and reported to the centralized Reasoning Telemetry sink.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] -> [PBRB Firewall] -> [Phase Manager] -> [LLM Gateway]
    PBRB acts as an "Economic Sieve," filtering and modifying reasoning headers based on pre-attested phase constraints.
* **APIs / Interfaces:**
    * `SetPhaseBudget(phase string, tokenLimit int, effort string)`
    * `GetCurrentUsage(missionID string) BudgetReport`
* **Data Storage/State:**
    Phase-specific budgets are stored in the Shared KV Store (Blackboard), with real-time counters maintained in-memory for low-latency enforcement.

## 5. Alternatives Considered
* **Post-hoc Auditing:** Rejected because it doesn't prevent "Reasoning Storms" from exhausting budgets in real-time.
* **Global Token Limits:** Rejected because specialized agents require different intensities based on the complexity of their specific sub-task.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budgets are cryptographically tied to the Hardware-Attested Intent Lineage (HAIL) to prevent budget "stealing" between sub-missions.
* **Observability:** Consumption metrics are exported to the `Reasoning Telemetry Exporter`.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
