# Design Doc: Recursive Token Budgeting (RTB) Controller
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
Current resource management in agent swarms is often flat, leading to "Budget Squatting" where a single subagent can exhaust the entire mission's token or reasoning budget on redundant paths. As swarms scale horizontally and vertically, MCP Any needs a recursive system to assign and enforce cryptographically bound budgets at the intent-branch level.

## 2. Goals & Non-Goals
* **Goals:**
    * Assign cryptographically bound reasoning-effort and token budgets to sub-intent branches.
    * Prevent "Token Exhaustion" attacks by rogue or inefficient subagents.
    * Support real-time budget reconciliation across mission phases.
    * Forcefully terminate branches that exceed their assigned budget.
* **Non-Goals:**
    * Managing financial billing for LLM providers (infrastructure layer only).
    * Optimizing the reasoning paths themselves (policy-neutral).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (e.g., Claude Code Team Lead)
* **Primary Goal:** Delegate a complex research task to 3 specialized agents, ensuring each agent only consumes 15% of the total mission token budget.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent defines a total budget for a research mission.
    2. The RTB Controller issues 3 sub-budget tokens bound to the respective intent branches.
    3. Specialized agents include these tokens in their reasoning and tool requests.
    4. The RTB Controller tracks consumption in real-time.
    5. Agent 2 attempts to exceed its 15% allocation; the RTB Controller blocks the request and signals the parent agent to re-allocate or terminate.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Allocate Budget| RTB[RTB Controller]
        RTB -->|Issue Budget Token| S1[Subagent 1]
        S1 -->|Tool Call + Token| GW[Secure Gateway]
        GW -->|Verify Budget| RTB
        RTB -->|Decrement| Ledger[Budget Ledger]
        Ledger -->|Exhausted| Reject[Reject & Terminate Branch]
        Ledger -->|Available| Accept[Accept Request]
    ```
* **APIs / Interfaces:**
    * `POST /v1/budget/allocate`: Assign a budget to a new intent branch.
    * `GET /v1/budget/status/{branch_id}`: Retrieve real-time budget consumption.
    * `POST /v1/budget/reclaim`: Reclaim unused budget from a terminated branch.
* **Data Storage/State:**
    * Budgets are managed in a hardware-attested ledger (SQLite sidecar) with monotonic counters.

## 5. Alternatives Considered
* **Global Quotas:** Rejected as they don't allow for granular control over individual swarm branches.
* **Reactive Monitoring:** Rejected as it cannot prevent "Token Storms" before they occur.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budget tokens are cryptographically linked to mission-root intents and cannot be shared between branches.
* **Observability:** Real-time visualization of budget consumption via the Resource Attribution Overlay.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
