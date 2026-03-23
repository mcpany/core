# Design Doc: Phase-Bound Reasoning Budget (PBRB) Firewall
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Agentic DoS (Denial of Spend) threatens autonomous swarms. PBRB enforces hard limits on "Reasoning Effort" per phase.

## 2. Goals & Non-Goals
* **Goals:** Enforce token/ARE budgets, Autonomy suspension.
* **Non-Goals:** Reasoning optimization.

## 3. Critical User Journey (CUJ)
1. User sets budget. 2. PBRB tracks usage. 3. HITL triggered at 90%. 4. User expands or terminates.

## 4. Design & Architecture
[Agent Call] -> [PBRB Counter] -> [Budget Validation] -> [Execution]

## 5. Alternatives Considered
Per-model limits (insufficient for multi-model tasks).

## 6. Cross-Cutting Concerns
Bound to Mission Root ID, Real-time tracking.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
