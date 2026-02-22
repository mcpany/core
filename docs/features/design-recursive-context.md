# Design Doc: Recursive Context Propagation
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents move from single-task execution to complex swarm orchestrations, the need for shared context across a hierarchy of agents becomes critical. Currently, when an orchestrator agent calls a sub-agent via MCP, the sub-agent often lacks the broader "intent" or "constraints" of the original user request unless explicitly passed in the tool arguments. This leads to redundant prompts and "context fragmentation." MCP Any needs to provide a standardized middleware layer that automatically propagates context headers down the execution chain.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize a set of headers (e.g., `X-MCP-Context-ID`, `X-MCP-Parent-Intent`) for context inheritance.
    * Provide middleware in MCP Any to automatically inject these headers into outgoing sub-agent calls.
    * Enable agents to query their "inherited context" via a specialized meta-tool.
* **Non-Goals:**
    * Managing the LLM's internal KV cache (handled by the model provider).
    * Synchronizing full conversation histories (focus is on intent and constraints).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Ensure that a "Security Auditor" sub-agent inherits the "High Privacy" constraint set by the human user to the root "DevOps" agent.
* **The Happy Path (Tasks):**
    1. Human tells Root Agent: "Fix the bug in the auth module, but maintain High Privacy."
    2. Root Agent calls "Security Auditor" tool via MCP Any.
    3. MCP Any middleware detects the "High Privacy" constraint in the root session.
    4. MCP Any injects `X-MCP-Context-Inheritance: privacy=high` into the request to the Security Auditor.
    5. Security Auditor agent receives the call and automatically adjusts its system prompt based on the inherited header.

## 4. Design & Architecture
* **System Flow:**
    `Root Agent -> [MCP Any Middleware: Context Capture] -> MCP Any Gateway -> [MCP Any Middleware: Context Injection] -> Sub-Agent`
* **APIs / Interfaces:**
    *   `mcpany_get_inherited_context()`: A new built-in tool that returns the current inheritance stack.
* **Data Storage/State:**
    *   Ephemeral state stored in the MCP Any session registry, indexed by `Context-ID`.

## 5. Alternatives Considered
*   **Explicit Argument Passing:** Forcing agents to pass a `context` string in every tool call. Rejected because it increases token cost and requires every agent to be manually updated to support the new schema.
*   **Global Blackboard:** Storing context in a central DB. Rejected as it adds latency and complexity for simple inheritance flows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevent "Context Injection" attacks. Headers must be signed by the MCP Any core to prevent spoofing.
* **Observability:** Inherited context stacks will be visible in the "Trace Detail" view in the UI.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
