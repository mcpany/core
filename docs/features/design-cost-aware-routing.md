# Design Doc: Economic Intelligence Middleware (Cost-Aware Routing)

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As agents perform increasingly complex, multi-step autonomous workflows, inference costs have become a primary bottleneck. Agents often use expensive "reasoning" models (e.g., O3, DeepSeek-R1) for simple tasks where a smaller, cheaper model would suffice. MCP Any, as the universal gateway, is uniquely positioned to inject economic metadata into the tool discovery and execution loop, enabling agents to perform "Economic Reasoning."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Inject real-time cost metadata (USD per 1k tokens) into MCP tool schemas.
    *   Provide a "Routing Advisor" tool that agents can call to get model recommendations based on prompt complexity.
    *   Implement "Cost-Cap" policies that block execution if an agent's session exceeds a financial threshold.
    *   Cache common responses to reduce upstream LLM costs (inspired by ClawRouter).
*   **Non-Goals:**
    *   Replacing external LLM providers (MCP Any remains a gateway).
    *   Handling complex billing/payments (MCP Any provides telemetry, not a bank).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Agent Orchestrator managing a fleet of subagents.
*   **Primary Goal:** Reduce monthly LLM spend by 50% without sacrificing agent accuracy.
*   **The Happy Path (Tasks):**
    1.  Orchestrator enables `EconomicIntelligence` middleware in `config.yaml`.
    2.  Agent requests a list of tools.
    3.  MCP Any returns tool schemas augmented with `_mcp_economic_metadata` (latency, cost, recommended model).
    4.  Agent uses the "Routing Advisor" to determine if a task requires "Reasoning" or "Simple" models.
    5.  Agent executes the tool call; MCP Any logs the cost and increments the session's budget tracker.

## 4. Design & Architecture
*   **System Flow:**
    - **Metadata Injector**: Intercepts `tools/list` and `resources/list` to append cost telemetry.
    - **Routing Engine**: A 15-dimension local classifier that scores incoming prompts (SIMPLE/MEDIUM/COMPLEX/REASONING).
    - **Budget Controller**: A stateful middleware that tracks cumulative costs against a `SessionID`.
*   **APIs / Interfaces:**
    - `tools/call` response extension: `_mcp_cost_summary: { input_tokens: 100, output_tokens: 50, cost_usd: 0.0015 }`
    - New Built-in Tool: `mcpany_route_advisor(prompt: string) -> { model_tier: "eco" | "premium", reason: "simple syntax check" }`
*   **Data Storage/State:** Uses the Shared KV Store (Blackboard) to persist session budgets.

## 5. Alternatives Considered
*   **External Router (ClawRouter)**: Using an external routing service. *Rejected* because it adds latency and complicates the security perimeter. Integrating it directly into the gateway provides better performance and "Zero Trust" context control.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Cost telemetry must not leak sensitive prompt information to 3rd party telemetry providers unless explicitly authorized.
*   **Observability:** New dashboard in the UI to visualize "Savings achieved via Economic Routing."

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
