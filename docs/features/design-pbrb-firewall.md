# Design Doc: PBRB (Phase-Bound Reasoning Budget) Firewall
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As models like O1 and O3 introduce internal reasoning chains, the "Reasoning Timing Attack" (RTA) and "Budget Exhaustion" have become major risks for autonomous swarms. A subagent can inadvertently enter a deep, non-terminating reasoning cycle, consuming thousands of compute units and stalling the entire mission.

MCP Any must transition from simple tool gating to **Phase-Bound Reasoning Budgets (PBRB)**, enforcing hard compute limits per agent turn.

## 2. Goals & Non-Goals
* **Goals:**
    * Throttle reasoning effort (e.g., `x-gemini-reasoning-effort`) based on mission priority.
    * Enforce a hard "Compute Budget" per phase (turn) of the swarm.
    * Automatically down-rank reasoning effort if the mission-root budget is depleted.
    * Integrate with cloud-provider "Reasoning Effort" headers.
* **Non-Goals:**
    * Direct credit-card billing (PBRB works on token/compute-hour abstractions).
    * Real-time text sanitization (handled by AID Hub).
    * Managing model training data.

## 3. Critical User Journey (CUJ)
* **User Persona:** Budget-Conscious Developer
* **Primary Goal:** Prevent an autonomous "Test Generator" from consuming $50 in compute effort for a trivial bug fix.
* **The Happy Path (Tasks):**
    1. User starts a mission with a global `Mission-Budget: 100.0` units.
    2. Subagent 1 (Linter) is spawned with a `Phase-Budget: 1.0`.
    3. PBRB Firewall monitors the `ARE` (Advanced Reasoning Effort) headers from the provider.
    4. If the provider signals excessive effort, PBRB injects an interrupt or down-ranks subsequent turns.
    5. The mission completes with 99.0 units remaining.

## 4. Design & Architecture
* **System Flow:**
    `[Model Response (ARE-Header)] -> [PBRB Firewall] -> [Reasoning Telemetry]`
    The Firewall intercepts response metadata and updates the mission-root budget in the Blackboard.
* **APIs / Interfaces:**
    * `X-MCP-Phase-Budget: <float>`
    * `X-MCP-Cumulative-Budget: <float>`
    * `EnforceReasoningPolicy(MissionID, PolicyID)`
* **Data Storage/State:**
    * Mission-root session store in the Blackboard (Shared KV).
    * Budget ledger for mission-wide reconciliation.

## 5. Alternatives Considered
* **Model-Side Quotas:** Often inaccessible to users (cloud providers enforce their own, but don't allow granular per-subagent control).
* **Reactive Alerting:** Rejected; real-time termination/throttling is required to prevent cost spikes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budget tokens are signed by the mission-root identity and cannot be escalated by subagents.
* **Observability:** Real-time budget tracking visible in the UI via the PBRB Budget Tracker dashboard.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
