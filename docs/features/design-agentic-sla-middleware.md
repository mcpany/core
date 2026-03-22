# Design Doc: Agentic SLA Middleware
**Status:** Draft
**Created:** 2026-03-22

## 1. Context and Scope
As AI agent swarms move from experimental setups to production environments, the lack of deterministic resource control has become a critical bottleneck. Multi-agent systems often encounter "Spiral of Death" loops—where agents recursively call each other or refine their own outputs indefinitely—leading to unpredictable costs and latencies.

The `Agentic SLA Middleware` introduces a governance layer within MCP Any that enforces Service Level Agreements (SLAs) on tool calls and agent task delegations. By binding every autonomous action to a resource contract, we ensure that swarms operate within the economic and temporal boundaries defined by the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce per-call and per-session token budgets (input/output).
    * Implement reasoning-time limits to prevent infinite execution loops.
    * Provide "Intent-Bound Budgets" where a parent agent can sub-delegate portions of its own resource quota.
    * Support hardware-attested budget enforcement to prevent subagent tampering.
* **Non-Goals:**
    * Direct model-provider billing integration (handles quotas at the gateway level).
    * General-purpose cloud cost management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Agent Architect
* **Primary Goal:** Delegate a complex research task to a subagent swarm without risking a "Token Storm" that exceeds $50 in API costs.
* **The Happy Path (Tasks):**
    1. The Parent Agent issues a `delegate_task` call via UACO, including an `SLA_Contract` object (Max Tokens: 500k, Max Time: 120s).
    2. MCP Any validates the contract and issues a signed `BudgetToken`.
    3. The Subagent performs tool calls, each carrying the `BudgetToken`.
    4. The Middleware intercepts each call, updates the consumption counters, and verifies they are within limits.
    5. When the Subagent attempts a call that would exceed the 500k token limit, MCP Any rejects it with `429 SLA Limit Exceeded`.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `SLA Middleware` -> `Quota Checker (In-Memory)` -> `Upstream Adapter` -> `Response Monitor` -> `Quota Updater`
* **APIs / Interfaces:**
    * UACO Header Extension: `x-mcp-sla-contract: { "max_tokens": 1000, "max_seconds": 30, "intent_id": "uuid" }`
    * Internal Management API: `/sla/quotas` (GET/POST) for real-time monitoring.
* **Data Storage/State:**
    * **Active Sessions**: High-speed Redis or thread-safe in-memory map for real-time tracking of token/time consumption.
    * **Audit Log**: SQLite for persistent records of SLA violations and resource usage.

## 5. Alternatives Considered
* **Client-Side Throttling**: Rejected because subagents cannot be trusted to self-throttle in a Zero-Trust environment.
* **Model-Provider Quotas**: Insufficient for granular, task-specific isolation within a shared API key environment.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: SLA contracts are cryptographically bound to the session, preventing subagents from "stealing" or modifying their own budgets.
* **Observability**: Real-time integration with the `Agentic SLA Monitor` dashboard in the UI.

## 7. Evolutionary Changelog
* **2026-03-22:** Initial Document Creation.
