# Design Doc: Phase-Bound Reasoning Budget (PBRB) Firewall
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Agentic DoS (Denial of Spend) is an emerging threat where non-convergent reasoning loops in autonomous swarms exhaust API budgets.

The PBRB Firewall enforces hard limits on "Reasoning Effort" per agentic phase, preventing recursive loops from scaling indefinitely without human intervention.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce token and ARE (Average Reasoning Effort) budgets per session.
    * Provide a suspension-of-autonomy signal when budgets are exceeded.
* **Non-Goals:**
    * It will not optimize the reasoning itself, only monitor its consumption.

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer managing agent costs.
* **Primary Goal:** Cap reasoning costs at $5.00 per high-level task.
* **The Happy Path (Tasks):**
    1. User sets a Reasoning Budget for the swarm.
    2. PBRB Firewall tracks token usage and reasoning-depth.
    3. Firewall triggers HITL Middleware if budget hits 90%.
    4. User either expands budget or terminates the loop.

## 4. Design & Architecture
* **System Flow:**
    [Agent Call] -> [PBRB Counter] -> [Budget Validation] -> [Execution]
* **APIs / Interfaces:**
    * `POST /v1/governance/budget`
* **Data Storage/State:**
    * Budget counters stored in the Shared KV Store (Redis-backed).

## 5. Alternatives Considered
* **Per-Model Limits**: Insufficient because complex tasks span multiple models and tools.

## 6. Cross-Cleaning Concerns
* **Security (Zero Trust):** Budgets are cryptographically bound to the Mission Root ID.
* **Observability:** Real-time budget tracking in the UI roadmap.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
