# Design Doc: Generalist Agent Delegation Middleware

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
With the rise of "Generalist Agents" (e.g., Gemini CLI v0.32.0) and "Agent Swarms" (OpenClaw), the primary bottleneck has shifted from "How does an agent use a tool?" to "How does a generalist agent delegate tasks to the right specialist?". MCP Any is uniquely positioned to handle this delegation by acting as the intelligent router between high-level intent and low-level MCP tool capabilities.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement an orchestration layer that translates high-level task descriptions into specific tool calls for specialized subagents.
    *   Provide a unified session state that persists across multiple agent handoffs.
    *   Support autonomous "Specialist Discovery" where the middleware suggests the best-fit agent for a given sub-task based on MCP tool metadata.
    *   Integrate with the "Adaptive Tool Scoping" to limit subagent visibility to relevant tools only.
*   **Non-Goals:**
    *   Building the LLM models themselves.
    *   Replacing the high-level orchestrator's planning logic (MCP Any facilitates the execution of that plan).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Architect.
*   **Primary Goal:** Deploy a "Manager Agent" that can securely delegate filesystem tasks to a "DevOps Subagent" and data analysis to a "Data Subagent."
*   **The Happy Path (Tasks):**
    1.  The Manager Agent receives a complex request: "Audit the logs and generate a summary PDF."
    2.  The Manager calls the `delegate_task` tool provided by MCP Any.
    3.  MCP Any identifies the "DevOps Subagent" (via its MCP-connected logs tool) and the "Writer Subagent" (via its PDF tool).
    4.  The middleware manages the sequence: DevOps audits logs -> passes data to Writer -> Writer generates PDF.
    5.  The final result is returned to the Manager Agent.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Analysis**: The middleware uses a lightweight "Router LLM" or embedding-based search to map tasks to agents.
    - **Session Handoff**: Each delegation event creates a sub-session that inherits context from the parent (Recursive Context Protocol).
    - **State Blackboard**: All agents in the delegation chain read/write to the "Shared KV Store."
*   **APIs / Interfaces:**
    - `mcp_delegate(task: string, constraints: object)`
    - `mcp_register_specialist(name: string, capabilities: string[])`
*   **Data Storage/State:** Session lineage and blackboard state are stored in the internal SQLite store.

## 5. Alternatives Considered
*   **Hardcoded Routing**: Forcing the user to define every handoff. *Rejected* because it doesn't scale with dynamic agent swarms.
*   **Purely LLM-based Routing**: Letting the parent LLM handle everything. *Rejected* because it's expensive, slow, and prone to "Context Pollution."

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The delegation middleware enforces "Intent-Bound Scoping." A subagent can only access tools that are explicitly relevant to the delegated task.
*   **Observability:** The UI "Agent Chain Tracer" provides a waterfall view of the delegation sequence.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
