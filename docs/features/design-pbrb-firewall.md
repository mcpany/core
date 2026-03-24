# Design Doc: PBRB Firewall
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As models introduce internal reasoning chains, the "Reasoning Timing Attack" (RTA) has become a risk. Malicious subagents can enter deep reasoning cycles, exhausting compute budgets. MCP Any must transition to **Phase-Bound Reasoning Budgets (PBRB)**.

## 2. Goals & Non-Goals
* **Goals:**
    * Throttle reasoning effort based on mission priority.
    * Enforce a hard "Compute Budget" per agent turn.
    * Down-rank reasoning effort if the mission-root budget is depleted.
* **Non-Goals:**
    * Direct credit-card billing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Budget-Conscious Developer
* **Primary Goal:** Prevent an autonomous subagent from consuming excessive compute units.
* **The Happy Path (Tasks):**
    1. User starts a mission with a global `Mission-Budget`.
    2. Subagent is spawned with a `Phase-Budget`.
    3. PBRB Firewall monitors the reasoning effort headers from the provider.
    4. If the provider signals excessive effort, PBRB injects an interrupt.

## 4. Design & Architecture
* **System Flow:**
    `[Model Response] -> [PBRB Firewall] -> [Reasoning Telemetry]`
    The Firewall intercepts response metadata and updates the mission-root budget in the Blackboard.
* **APIs / Interfaces:**
    `X-MCP-Phase-Budget: <float>`
* **Data Storage/State:**
    Mission-root session store in the Blackboard.

## 5. Alternatives Considered
* **Model-Side Quotas:** Often inaccessible or too granular for users to manage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budget tokens are signed by the mission-root identity.
* **Observability:** Real-time budget tracking visible in the UI.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
