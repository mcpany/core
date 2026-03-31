# Design Doc: Tool-Level Reasoning Attribution (TLRA) Provider
**Status:** Draft
**Created:** 2026-07-18

## 1. Context and Scope
The introduction of Reasoning-as-a-Service (RaaS) by model providers allows individual tools and MCP servers to request their own sub-reasoning shards. While powerful, this creates a risk of "Reasoning Fork-Bombs" where a tool recursively calls RaaS without being bound by the parent agent's budget. The TLRA Provider ensures that every reasoning fragment is cryptographically attributed to the initiating tool's lineage and governed by mission-bound resource quotas.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically attribute every RaaS request to its initiating tool and mission-root lineage.
    * Enforce granular, hardware-locked token and reasoning budgets for individual tools.
    * Detect and terminate "Recursive Reasoning Loops" at the infrastructure layer.
    * Provide real-time economic transparency for tool-initiated "thinking."
* **Non-Goals:**
    * Governing the quality of reasoning (handled by ARI Hub).
    * Restricting model-internal reasoning depth (governed by model API limits).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Infrastructure Architect
* **Primary Goal:** Prevent a "Search Tool" from consuming $100 in RaaS tokens by recursively searching and analyzing 50+ documents in a single turn.
* **The Happy Path (Tasks):**
    1. The mission-root defines a reasoning budget of 1M tokens for the "Search Tool."
    2. The tool initiates a RaaS shard via the MCP Any gateway.
    3. TLRA Middleware injects an attribution header and tracks consumption.
    4. The tool attempts a second, recursive RaaS call.
    5. TLRA detects the branch and verifies it against the tool-specific quota.
    6. If the quota is exceeded, the TLRA Interceptor halts the reasoning loop and alerts the parent agent.

## 4. Design & Architecture
* **System Flow:**
  [Tool RaaS Request] -> [TLRA Middleware] -> [Model API] -> [Token Usage Sink] -> [Quota Enforcer]
* **APIs / Interfaces:**
    * `mcpany.tlra.v1.AttributionService`
    * Extension of `x-mcp-reasoning-lineage` headers.
* **Data Storage/State:**
    * In-memory counters for tool-bound reasoning quotas; hardware-attested budget fragments.

## 5. Alternatives Considered
* **Global Agent Budgets**: Rejected as they cannot distinguish between parent reasoning and runaway tool behavior.
* **Model-Side Throttling**: Rejected as it lacks the fine-grained mission-root context of MCP Any.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Attribution tokens are hardware-signed to prevent spoofing by rogue subagents.
* **Observability:** Tool-level reasoning costs are visualized in the Mission Cost Attribution Dashboard.

## 7. Evolutionary Changelog
* **2026-07-18:** Initial Document Creation.
