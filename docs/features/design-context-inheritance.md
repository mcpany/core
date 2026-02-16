# Design Doc: Standardized Context Inheritance
**Status:** Draft
**Created:** 2026-02-16

## 1. Context and Scope
In complex agent workflows, maintaining state across multiple specialized agents is a major hurdle. Subagents often lose critical context (e.g., project goals, user preferences, or session-specific constraints) unless it's manually re-injected into every prompt, leading to "context bloat" and high token costs. MCP Any needs a mechanism to manage and propagate context securely and efficiently across a swarm.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable hierarchical context inheritance (Global -> Profile -> Agent).
    *   Reduce redundant data in prompts via "Shared Context Pointers".
    *   Allow agents to "opt-in" to specific context streams.
    *   Enforce security policies on inherited context (e.g., Agent B cannot inherit Agent A's API keys).
*   **Non-Goals:**
    *   Building a full-blown long-term memory database (Vector DB).
    *   Automatically summarizing context (this is left to the LLM/Orchestrator).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Architect
*   **Primary Goal:** Ensure that all agents in a DevOps swarm inherit the same "Production Guardrails" without manual configuration for each agent.
*   **The Happy Path (Tasks):**
    1.  Architect defines a "Production Guardrail" context block in MCP Any.
    2.  Architect configures the "DevOps Profile" to inherit this block.
    3.  When "DeployAgent" and "MonitorAgent" start, they automatically receive the guardrail context via the MCP `resources` or `prompts` capability.
    4.  The agents include these guardrails in their reasoning without the user needing to repeat them.

## 4. Design & Architecture
*   **System Flow:**
    *   Context is stored as "Blocks" in the configuration or database.
    *   Inheritance is resolved at session start or tool execution time.
*   **APIs / Interfaces:**
    *   `getContext(agent_id, session_id)`: Resolves and returns the merged context for an agent.
*   **Data Storage/State:**
    *   Configuration-based (static blocks) or Database-based (dynamic session context).

## 5. Alternatives Considered
*   **Manual Prompt Injection**: Rejected due to token cost and maintenance overhead.
*   **Centralized Database**: Considered, but configuration-driven inheritance is preferred for local/private deployments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Context inheritance follows a "Least Privilege" model. Redaction rules are applied to inherited blocks.
*   **Observability:** Trace IDs include metadata about which context blocks were inherited.

## 7. Evolutionary Changelog
*   **2026-02-16:** Initial Document Creation.
