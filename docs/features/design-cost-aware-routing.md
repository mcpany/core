# Design Doc: Cost-Aware Model Router
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As the number of available LLMs grows, users and agent swarms are increasingly faced with the "Economical Reasoning" problem: which model is the most cost-effective for a given task? Frontier models like Claude 3.7 Opus are powerful but expensive, while models like GPT-4o-mini are significantly cheaper but less capable at complex reasoning. MCP Any needs a middleware layer that intelligently routes agent requests to the optimal model based on task complexity, cost, and latency requirements, similar to the `ClawRouter` pattern.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically route tool calls or subagent tasks to the most cost-effective model.
    * Provide a configurable policy for "Model Selection" based on task tags or historical complexity.
    * Support seamless fallback to more capable models if the cheaper model fails.
    * Track and display cost/token usage metrics per model and agent session.
* **Non-Goals:**
    * Implementing a new LLM provider (we bridge to existing ones).
    * Training specialized "router" models (we use rule-based or simple heuristic-based routing).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator (e.g., OpenClaw user).
* **Primary Goal:** Reduce API costs by 40% without sacrificing task success rates.
* **The Happy Path (Tasks):**
    1. Orchestrator defines a multi-agent workflow in MCP Any.
    2. A simple "File List" task is received.
    3. The `Cost-Aware Router` identifies this as a "Low Complexity" task.
    4. The task is routed to `gpt-4o-mini`.
    5. A subsequent "Code Refactoring" task is received.
    6. The router identifies this as "High Complexity" and routes it to `claude-3-7-opus`.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Gateway` -> `Cost-Aware Router Middleware` -> `LLM Provider (OpenAI/Anthropic/Ollama)`
    1. **Complexity Analysis**: Inspects the tool name, description, and input schema.
    2. **Policy Matching**: Matches against `routing-policy.yaml`.
    3. **Model Selection**: Selects the best available model from the configured pool.
    4. **Telemetry Injection**: Injects cost/latency metadata back into the agent session.
* **APIs / Interfaces:**
    * `MCP_ROUTING_POLICY`: Configuration schema for model tiers.
    * `GET /v1/metrics/costs`: Dashboard endpoint for cost visualization.
* **Data Storage/State:**
    * `model_metrics.db`: SQLite database storing historical performance (latency, success rate, cost) for each model-task pair.

## 5. Alternatives Considered
* **Hard-coded models in agents**: Rejected because it's inflexible and requires code changes to the agent itself.
* **External Proxy (e.g., LiteLLM)**: Good, but doesn't have the "MCP-Aware" context (tool schemas) that MCP Any provides for better complexity analysis.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Routing policies must be attested to prevent a "Cost Injection" attack where a malicious agent forces the use of the most expensive model.
* **Observability**: Real-time "Model Usage" waterfall charts in the UI.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation (inspired by the ClawRouter ecosystem shift).
