# Design Doc: Cognitive Debt Arbiter (CDA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The introduction of "Adaptive Reasoning Quotas" (ARQ) in the agent ecosystem allows agents to borrow reasoning effort (`x-gemini-reasoning-effort`) from future mission phases. While this provides flexibility, it introduces the risk of "Cognitive Bankruptcy," where recursive borrowing leads to a total exhaustion of the mission budget early in the lifecycle, causing swarm-wide stalls.

The CDA is an infrastructure-level resource manager that governs reasoning debt. It ensures that borrowing is sustainable and prevents recursive debt loops from destabilizing the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Track "Reasoning Debt" across all subagent branches in real-time.
    * Apply "Cognitive Interest Rates" (resource multipliers) to borrowed effort to discourage over-borrowing.
    * Enforce a "Hard Debt Ceiling" for every mission-root branch.
    * Provide "Debt Forgiveness" triggers based on hardware-attested task completion.
* **Non-Goals:**
    * Modifying the underlying model's reasoning logic.
    * Managing financial budgets (CDA is focused on *reasoning effort* and *tokens*).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Administrator
* **Primary Goal:** Prevent a complex research swarm from stalling due to excessive reasoning borrowing by specialist agents.
* **The Happy Path (Tasks):**
    1. Administrator defines a mission-root budget and a CDA policy (Interest Rate: 5%, Debt Ceiling: 20% of total budget).
    2. A "Code Auditor" subagent encounters a complex bug and requests a reasoning boost (borrowing from the next phase).
    3. The CDA intercepts the request and checks the current branch debt.
    4. CDA approves the borrow but applies the interest rate, deducting 105% of the requested effort from the future phase.
    5. A second subagent attempts to borrow another 20%, exceeding the branch ceiling.
    6. CDA interdicts the request and triggers a "Cognitive Refinement" signal to the parent agent to simplify the task.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent] -->|ARQ Borrow Request| B(CDA Gateway)
        B --> C{Debt Check}
        C -->|Below Ceiling| D[Apply Interest & Approve]
        C -->|Above Ceiling| E[Deny & Trigger Refinement]
        D --> F[Future Phase Budget Ledger]
        E --> G[Parent Notification]
    ```
* **APIs / Interfaces:**
    * `BorrowReasoningEffort(amount int, branchID string) -> success/fail`
    * `GetDebtStatus(missionID string) -> debtSummary`
* **Data Storage/State:**
    * Debt ledgers are stored in the hardware-locked mission state (TPM-bound).

## 5. Alternatives Considered
* **Static Budgets**: Rejected because they lack the flexibility needed for high-stakes, non-deterministic reasoning.
* **Global Debt Capping**: Rejected because it penalizes efficient branches for the over-borrowing of a single runaway subagent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Borrowing must be authorized by a hardware-attested parent token.
* **Observability:** Debt levels and interest accrual are visualized in the Mission Budget Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
