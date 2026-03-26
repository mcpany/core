# Design Doc: Adaptive Intent Budgeting (AIB)
**Status:** Draft | In Review | Approved
**Created:** 2026-05-02

## 1. Context and Scope
With the introduction of Intent-Scoped Resource Quotas (ISRQ) in Claude Code and Gemini CLI, a central service is needed to manage and scale resource leases (tokens, compute) dynamically. The **Adaptive Intent Budgeting (AIB)** middleware is a resource management layer that scales agent leases based on reasoning confidence and task complexity.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically adjust token and compute leases for agent swarms.
    * Enforce resource quotas based on predicted tool-call impact.
    * Provide real-time telemetry on resource usage vs. intent budget.
* **Non-Goals:**
    * Modifying agent reasoning logic directly.
    * Managing billing (this is the job of the LLM provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (e.g., Gemini CLI)
* **Primary Goal:** Share secure resource context between 3 agents without exceeding the predicted token budget.
* **The Happy Path (Tasks):**
    1. Parent agent submits an "Intent-Bound Task Proposal" with a predicted resource budget.
    2. AIB middleware evaluates the proposal's reasoning confidence and task complexity.
    3. AIB allocates a \"Resource Lease\" (e.g., 50k tokens, 20s compute) to each specialized subagent.
    4. Subagents report usage metrics back to AIB via the Unified Telemetry Bridge.
    5. AIB dynamically scales leases (e.g., prunes tokens if usage exceeds predictions) or suspends tools.
    6. Parent agent reconciles the final budget after task completion.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] -> (Intent-Bound Proposal) -> [AIB Middleware]
    [AIB Middleware] -> (Resource Lease) -> [Subagent 1, 2, 3]
    [Subagent 1, 2, 3] -> (Usage Metrics) -> [Unified Telemetry Bridge]
    [Unified Telemetry Bridge] -> (Resource Feedback) -> [AIB Middleware]
* **APIs / Interfaces:**
    * `RequestResourceLease(IntentID, PredictedBudget)`: Requests a lease for a subagent.
    * `ReportResourceUsage(LeaseID, Metrics)`: Reports actual resource consumption.
    * `SuspendResourceLease(LeaseID)`: Immediately revokes a lease.
* **Data Storage/State:**
    * Uses a high-performance in-memory state store (Redis/XSync) for real-time lease tracking and enforcement.

## 5. Alternatives Considered
* **Static Resource Quotas:** Rejected as it leads to \"Reasoning Starvation\" for complex tasks or \"Resource Bloat\" for simple queries.
* **Framework-Specific Budgeting:** Rejected as it prevents cross-framework (OpenClaw/Gemini) budget reconciliation in a shared swarm.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Resource leases are bound to the agent's cryptographically verified intent to prevent \"Budget Smuggling.\"
* **Observability:** Integrated with the Unified Telemetry Bridge for real-time monitoring of swarm-wide resource efficiency.

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.

### Update: 2026-05-02 - Integration with ISRQ (Intent-Scoped Resource Quotas)
**Context:** Aligning with Anthropic's ISRQ standard and Google's SACH leases.
**Architecture Adjustment:**
- Budgeting will now enforce **L4 Resource Leases** (Network/Disk IOPS) in addition to token/compute limits.
- Implementation of "Burst Leases" that allow temporary resource spikes during high-confidence reasoning steps.
**Security Impact:** Mitigates silent data exfiltration by capping network bandwidth based on the complexity of the declared intent.
