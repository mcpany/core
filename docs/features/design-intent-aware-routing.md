# Design Doc: Intent-Aware Routing Middleware

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As tool catalogs expand into the thousands, even "On-Demand Discovery" (Lazy-MCP) can return too many results for an LLM to process effectively. Furthermore, simple keyword search often fails to capture the nuanced *intent* of a complex multi-step agentic workflow. MCP Any needs a routing layer that understands the agent's high-level goal and dynamically shapes the tool-space to match that intent, while also considering historical performance and success rates.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Dynamically filter and prioritize tools based on the active agent session's "Intent Context."
    *   Reduce tool selection errors (hallucinations) by 40% in large catalogs.
    *   Incorporate historical success/failure rates into the routing logic.
    *   Provide "Intent Hints" to the LLM to guide its selection process.
*   **Non-Goals:**
    *   Replacing the LLM's reasoning (the router suggests, the LLM decides).
    *   Defining the intent (the agent/client must provide the high-level goal).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Orchestrator managing a "Full Stack Developer" swarm.
*   **Primary Goal:** The swarm should only see "Frontend" tools when working on UI tasks, and "DB" tools when working on migrations, without manual reconfiguration.
*   **The Happy Path (Tasks):**
    1.  The orchestrator starts a session with intent: "Migrate the user table to include MFA fields."
    2.  LLM requests tools.
    3.  Intent-Aware Router analyzes the intent and boosts "Postgres" and "Migration" tools.
    4.  The LLM receives a focused tool list where irrelevant tools (e.g., "Slack Send") are suppressed or ranked lower.
    5.  The router logs the success of the chosen tool for future boosting.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Extraction**: The router intercepts the `SessionContext` or a new `X-MCP-Intent` header.
    - **Scoring Engine**: A hybrid scoring model combining:
        - Semantic similarity between Intent and Tool Description.
        - Historical success rate for that Intent + Tool pair.
        - Token cost / Latency (from Resource Telemetry).
    - **Dynamic Filtering**: The `tools/list` response is truncated or re-ordered based on the scores.
*   **APIs / Interfaces:**
    - Header: `X-MCP-Intent: "Migrate database tables for MFA support"`
    - Metadata: `_mcp_routing_score: 0.95`
*   **Data Storage/State:** Intent-to-Tool performance metrics stored in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Hard-coded Scopes**: Defining static tool groups (e.g., "Frontend-Group"). *Rejected* as it is too brittle for autonomous agents.
*   **LLM Pre-routing**: Using a small LLM to pick tools. *Rejected* due to latency and cost overhead.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The router cannot override the Policy Firewall; it can only further restrict the set of available tools.
*   **Observability:** Visualizing the "Intent-to-Tool" mapping in the UI to help developers understand why certain tools were promoted or suppressed.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
