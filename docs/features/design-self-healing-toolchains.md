# Design Doc: Self-Healing Toolchains (Active Resilience)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents become more autonomous, their dependency on external tools increases. However, tools often fail due to transient network issues, API changes, or incorrect parameter mappings. Currently, MCP Any provides passive logging, which requires human intervention to fix. "Self-Healing Toolchains" introduce an active resilience layer that uses an internal reasoning loop to diagnose tool failures and automatically attempt repairs or suggest configuration patches.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically retry failed tool calls with corrected parameters based on error feedback.
    * Generate "Hot-Patch" suggestions for `mcpany` configuration files when an upstream API schema changes.
    * Reduce "Agent Stalls" caused by brittle tool interfaces.
* **Non-Goals:**
    * Fixing upstream service logic (we only fix the *interface* or *call*).
    * Fully autonomous configuration updates without optional HITL approval for permanent changes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agentic Swarm Developer
* **Primary Goal:** Maintain 99.9% tool success rate without manual config maintenance.
* **The Happy Path (Tasks):**
    1. Agent calls `get_weather` with an outdated parameter `city_name`.
    2. Upstream returns a 400 error: "Param 'city' is required, 'city_name' is deprecated."
    3. Self-Healing Middleware captures the error and analyzes the message.
    4. Middleware re-maps the argument and successfully retries the call.
    5. Middleware alerts the developer with a proposed YAML patch for `config.yaml`.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Core Server -> [Self-Healing Middleware] -> Upstream Adapter -> Upstream Service`
    When an error occurs:
    `Upstream Error -> Analysis Loop (Internal LLM/Heuristics) -> Corrected Request -> Retry`
* **APIs / Interfaces:**
    * `ResilienceHook`: A new middleware interface for error interception.
    * `/api/resilience/patches`: Endpoint to retrieve suggested config updates.
* **Data Storage/State:**
    * Error-Correction memory stored in the Shared KV Store to prevent repeating the same mistake.

## 5. Alternatives Considered
* **Hardcoded Retries:** Rejected because they don't handle schema changes or semantic errors.
* **Agent-side Correction:** Rejected because it puts the burden on every individual agent rather than the central infrastructure.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Self-healing must not bypass Policy Firewall. Re-mapped calls are re-validated.
* **Observability:** Every "Heal" event is logged in the Audit Log with a "Healed" tag.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
