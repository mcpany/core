# Design Doc: Adaptive Intent Budgeting (AIB)
**Status:** Draft
**Created:** 2026-05-02

## 1. Context and Scope
Modern agent swarms, particularly those using Gemini CLI v0.36.0+, require dynamic resource management that scales with reasoning complexity. "Adaptive Intent Budgeting" (AIB) allows MCP Any to manage token, compute, and time budgets that are not static, but "Adaptive"—scaling up or down based on the agent's real-time confidence and the verified criticality of the sub-intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Lease-Based" resource allocation model for sub-intents.
    * Integrate with UACO v3.1 `x-gemini-reasoning-effort` headers to adjust budgets dynamically.
    * Provide a "Hard Stop" mechanism for recursive loops that exceed the maximum adaptive threshold.
    * Support "Intent-Scoped Resource Isolation" (ISRI) to prevent resource leakage between branches.
* **Non-Goals:**
    * Managing the underlying OS resource scheduling (handled by the Kernel-Bound Intent Broker).
    * Predicting the exact token count needed (uses a "Budget-and-Replenish" model).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps AI Engineer
* **Primary Goal:** Prevent a specialized "Code Refactoring" agent from consuming the entire monthly token budget on a single complex file.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a task to a refactoring subagent with an initial "Intent Budget."
    2. The AIB Middleware monitors token usage in real-time.
    3. The subagent requests a "Budget Replenishment" as it reaches 80% of its lease.
    4. AIB evaluates the subagent's "Reasoning Confidence" score.
    5. AIB grants a limited replenishment based on the parent's "Overall Mission Budget."

## 4. Design & Architecture
* **System Flow:**
    `Subagent Spawn` -> `Initial Budget Lease` -> `Token/Compute Monitoring` -> `Replenishment Negotiation` -> `Threshold Enforcement`
* **APIs / Interfaces:**
    * `BudgetBroker`: `RequestLease(intentID string, initialReq ResourceMap) (Lease, error)`
    * `Replenish(leaseID string, confidenceScore float64) (Lease, error)`
* **Data Storage/State:**
    * Real-time budget consumption is tracked in the memory-mapped Shared KV Store for sub-millisecond latency.

## 5. Alternatives Considered
* **Static Per-Agent Limits**: Rejected because it leads to "Reasoning Stall" for complex but legitimate tasks.
* **Session-Only Budgeting**: Rejected because a single malicious subagent can exhaust the entire session budget.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budget replenishment requests must be cryptographically signed by the requesting subagent and validated against the parent's intent tree.
* **Observability:** Real-time budget usage and "Replenishment Events" are visualized in the "Adaptive Budgeting Monitor."

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.
