# Design Doc: Hierarchical Reasoning-Budget Enforcer (HRBE)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become more complex and decentralized, a single "Mission Root" often delegates tasks to multiple specialist subagents. Current economic guardrails in the Reasoning-Budget Firewall (RBF) are flat, meaning a single aggressive subagent can exhaust the entire budget of the parent mission, leading to "Budget Squatting" and cognitive stall for sibling agents.

HRBE introduces a recursive budgeting model where supervisors can issue cryptographically signed, "Fractional Reasoning Tokens" to their children. This ensures that sub-missions are economically bounded and that resources are preserved for the entire swarm's lifecycle.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement recursive delegation of reasoning effort (ARE) and token budgets.
    * Support "Fractional Tokens" that are subsets of the parent's total mission budget.
    * Provide real-time interdiction of subagents that exceed their allocated fractional budget.
    * Integrate with hardware-attested (TPM) counters to prevent token double-spending.
* **Non-Goals:**
    * Directly managing fiat currency or payment processing.
    * Optimizing the internal reasoning efficiency of the LLM itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (DevOps/AI Architect)
* **Primary Goal:** Prevent a "Code specialist" subagent from consuming 90% of the mission budget during a refactoring task, leaving no tokens for the "Testing specialist."
* **The Happy Path (Tasks):**
    1. The Mission-Root initiates a mission with a 1,000,000 token budget and "High" reasoning effort.
    2. The Mission-Root spawns a "Refactor Agent" and issues it a Fractional Token for 200,000 tokens and "Medium" effort.
    3. The "Refactor Agent" attempts to call a tool that would exceed the 200,000 limit.
    4. HRBE intercepts the request, blocks the tool call, and notifies the Mission-Root.
    5. The Mission-Root reclaims the unused portion and allocates it to the "Test Agent."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -- Issues signed Fractional Token --> B[HRBE Middleware]
        B -- Attests & Records --> C[(Budget Ledger)]
        D[Subagent] -- Presents Fractional Token --> B
        B -- Validates against Parent --> E{Budget Available?}
        E -- Yes --> F[Tool Execution]
        E -- No --> G[Interdiction & Event Notification]
    ```
* **APIs / Interfaces:**
    * `POST /v1/budget/delegate`: Issues a sub-budget token.
    * `GET /v1/budget/status/{mission_id}`: Real-time telemetry of hierarchical consumption.
* **Data Storage/State:**
    * Budget state is managed in a hardware-attested SQLite sidecar (Blackboard extension) utilizing monotonic counters.

## 5. Alternatives Considered
* **Flat Quotas:** Rejected because they don't account for the relative priority of sub-tasks.
* **Centralized Billing (Cloud-side):** Rejected due to latency and the requirement for "Local Sovereignty" in the Universal Agent Bus.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Fractional tokens must be cryptographically bound to the subagent's hardware-attested identity to prevent theft.
* **Observability:** Budget consumption is streamed to the `Reasoning Budget Dashboard` in the UI for real-time monitoring.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
