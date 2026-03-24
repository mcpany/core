# Design Doc: Cognitive Loop Circuit Breaker (CLCB)
**Status:** Draft
**Created:** 2026-07-02

## 1. Context and Scope
Autonomous agent swarms are increasingly subject to "Cognitive Denial of Service" (CDoS) attacks, where malicious inputs or non-deterministic tool outputs trigger infinite reasoning refinement loops. These loops exhaust token budgets, cause mission stalls, and lead to "Refinement Exhaustion" where the agent never converges on a final answer. MCP Any needs a framework-agnostic circuit breaker to protect the underlying infrastructure from these high-frequency, low-utility reasoning bursts.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect non-convergent reasoning loops using semantic entropy analysis.
    * Halt refinement cycles that exceed hardware-attested budget or depth thresholds.
    * Provide a standardized "Stall Signal" to the parent agent or user.
    * Maintain zero-latency overhead for healthy (convergent) reasoning chains.
* **Non-Goals:**
    * Replacing the LLM's internal self-correction logic (CLCB is an external governor).
    * Modifying the content of agent reasoning (it only gates execution).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Agent SRE
* **Primary Goal:** Prevent a rogue specialist agent from consuming $5,000 in tokens during a 10-minute "Refinement Loop."
* **The Happy Path (Tasks):**
    1. Specialist Agent receives a complex, ambiguous tool output.
    2. Agent initiates a refinement loop to "correct" its understanding.
    3. After 5 iterations with negligible semantic delta (detected by CLCB), the circuit breaker triggers.
    4. CLCB forcefully terminates the session and notifies the Parent Agent with a `STALL_DETECTED` error.
    5. Parent Agent re-plans or escalates to a human, preserving the budget.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Reasoning] --> B{CLCB Middleware}
        B -->|Entropy Check| C{Convergent?}
        C -->|Yes| D[Allow Next Turn]
        C -->|No| E{Threshold Exceeded?}
        E -->|Yes| F[Trigger Circuit Breaker]
        E -->|No| D
        F --> G[Force Terminate Session]
        G --> H[Notify Parent Agent]
    ```
* **APIs / Interfaces:**
    * `POST /v1/govern/reasoning-step`: Accepts the latest reasoning fragment and session ID. Returns `ALLOW` or `HALT`.
    * `reasoning-circuit-breaker-depth`: Header for setting per-task refinement limits.
* **Data Storage/State:**
    * **Entropy Buffer**: A sliding-window cache of reasoning fragment hashes and their semantic embeddings to calculate "Refinement Drift."

## 5. Alternatives Considered
* **Time-based Throttling**: Rejected because slow, high-value reasoning would be unfairly penalized.
* **Token-only Budgeting**: Rejected because it doesn't distinguish between productive deep reasoning and infinite loops until the budget is already gone.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CLCB signatures must be hardware-bound (TPM) to prevent a compromised subagent from spoofing its own entropy scores.
* **Observability:** Metrics on "Reasoning Convergence Rate" will be surfaced to the Multi-Agent Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-02:** Initial Document Creation.
