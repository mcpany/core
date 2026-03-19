# Design Doc: Phase-Bound Reasoning Budget (PBRB) Firewall
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Modern agentic swarms often suffer from "Resource Phase-Exhaustion," where a complex discovery phase consumes the majority of the reasoning budget (tokens/compute), leaving the agent with insufficient resources for the actual execution or verification phases. This often leads to "Mission Abandonment" or "Hallucinatory Shortcuts" at the most critical moment.

The PBRB Firewall evolves the Reasoning-Budget Firewall (RBF) to be phase-aware, allowing users to partition budgets for Discovery, Planning, and Execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement tiered budgeting for Discovery, Planning, and Execution phases.
    * Support hardware-attested budget roll-over or lockout.
    * Provide real-time telemetry for budget consumption per phase.
* **Non-Goals:**
    * Predicting the exact cost of a mission branch (budgets are limits, not guarantees).
    * Modifying upstream model pricing or billing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Resource Administrator
* **Primary Goal:** Prevent a fleet of coding agents from "spinning their wheels" in the discovery phase and exhausting the department's token budget.
* **The Happy Path (Tasks):**
    1. Administrator defines a Mission Profile with a $1.00 Discovery cap and a $5.00 Execution cap.
    2. Agent starts the mission; MCP Any monitors ARE headers to identify the "Discovery" phase.
    3. If the agent exceeds $1.00 in Discovery, MCP Any triggers a "Phase-Halt" signal, requiring user re-attestation or mission refinement.
    4. Upon successful transition to "Planning," a new budget bucket is activated.

## 4. Design & Architecture
* **System Flow:**
    [Agent Request: Discovery Phase] -> [PBRB Firewall: Check Discovery Bucket] -> [Upstream LLM]
    [LLM Response: Usage Metrics] -> [PBRB Firewall: Update Phase Totals]
* **APIs / Interfaces:**
    * `InitializePhaseBudget(MissionID, {Discovery: X, Planning: Y, Execution: Z})`
    * `TransitionPhase(MissionID, TargetPhase)`
* **Data Storage/State:**
    * Redis/SQLite store for mission-bound phase accumulators.
    * Hardware-attested budget manifests.

## 5. Alternatives Considered
* **Global Mission Budget:** Rejected as it fails to prevent "front-loading" of costs in non-critical phases.
* **Manual HITL Gating:** Rejected due to the high latency and human overhead in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Phase transitions must be attested by the parent agent or Mission Root to prevent "Phase Spoofing."
* **Observability:** Dashboard visualization of "Reasoning Efficiency" per phase.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
