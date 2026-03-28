# Design Doc: Dynamic Usage Quota Monitor
**Status:** Draft
**Created:** 2026-03-23

## 1. Context and Scope
The rapid adoption of Claude Code and Gemini CLI has exposed a critical bottleneck: "Quota Economics." High-frequency agentic swarms can consume weekly token budgets in minutes, especially during peak pricing or usage promotions. MCP Any needs an authoritative, infrastructure-level monitor to provide agents with "Economic Awareness" and prevent mission-stalling due to resource exhaustion.

## 2. Goals & Non-Goals
* **Goals:**
    * Track real-time token, credit, and usage quotas across all connected LLM frameworks.
    * Provide a unified API for agents to query their remaining mission budget.
    * Enforce automated "Economic Throttling" based on user-defined priority rules.
    * Integrate with the Policy Firewall to block low-priority tool calls when quotas are low.
* **Non-Goals:**
    * Managing billing or payments (we track usage, not dollars).
    * Providing model-specific cost estimation (we rely on framework headers like `x-gemini-usage`).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agentic SRE (Site Reliability Engineer)
* **Primary Goal:** Prevent a low-priority background swarm from exhausting the main project's Claude quota during peak hours.
* **The Happy Path (Tasks):**
    1. The user defines a "Mission Budget" of 1M tokens for the "Project Indexing" swarm.
    2. The Quota Monitor intercepts every tool call from the swarm.
    3. The monitor extracts usage data from response headers (e.g., `x-anthropic-usage`).
    4. When usage reaches 90%, the monitor signals the Policy Firewall to restrict the swarm to "Read-Only" mode.
    5. The user receives a notification in the UI dashboard about the quota status.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [MCP Any Proxy] -> [LLM API] -> [Header Extraction] -> [Quota Store] -> [Policy Check]`
* **APIs / Interfaces:**
    * `/v1/quota/status`: Get current usage and remaining budget for a specific profile or mission.
    * `/v1/quota/reserve`: Optimistically "lock" tokens for high-stakes reasoning blocks.
* **Data Storage/State:**
    * Uses the SQLite Blackboard to store persistent counters and per-mission budget limits.

## 5. Alternatives Considered
* **Framework-Specific Quotas:** Rejected because it doesn't allow for cross-framework budgeting (e.g., spending Claude tokens to save Gemini credits).
* **Manual HITL Approval:** Rejected as it becomes a bottleneck for autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Quota data is sensitive; access to the status API is restricted to authorized supervisor agents.
* **Observability:** Provides a "Quota Burn Rate" chart in the UI dashboard.

## 7. Evolutionary Changelog
* **2026-03-23:** Initial Document Creation.
