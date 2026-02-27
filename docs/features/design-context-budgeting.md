# Design Doc: Context Budgeting Middleware

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As agents operate in complex environments with hundreds of available tools, even with "Lazy-Discovery," there is a risk of "Context Drift." This happens when an agent manually searches for and adds too many tools to its active context, eventually crowding out the original task instructions or session history. MCP Any needs a proactive mechanism to manage this "Context Budget."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Maintain a strict limit on the number of active tool schemas in the LLM's context window.
    *   Automatically "evict" least-recently-used (LRU) tool schemas to stay within a configurable token budget.
    *   Provide the LLM with a "Budget Status" notification when tools are evicted.
    *   Ensure evicted tools remain searchable via the `mcpany_search_tools` utility.
*   **Non-Goals:**
    *   Modifying the LLM's internal weights or attention mechanisms.
    *   Compressing tool schemas (this is handled by the `Context Optimizer Middleware`).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator.
*   **Primary Goal:** Run a 24/7 autonomous agent that performs research across various domains without crashing due to context window overflow.
*   **The Happy Path (Tasks):**
    1.  Orchestrator configures MCP Any with a `context_budget` of 10 tools.
    2.  Agent performs a series of tasks, searching for and adding 10 different tools to its context.
    3.  Agent searches for an 11th tool (`analyze_logs`).
    4.  MCP Any identifies the least-recently-used tool (`weather_lookup`) and evicts its schema from the prompt.
    5.  MCP Any injects the `analyze_logs` schema.
    6.  MCP Any appends a metadata hint: `Budget Alert: 'weather_lookup' evicted to save context. Re-search if needed.`

## 4. Design & Architecture
*   **System Flow:**
    - **Tracking**: The `ContextBudgetMiddleware` maintains a per-session LRU cache of active tool IDs.
    - **Interception**: Every `tools/list` or `mcpany_search_tools` call is intercepted to check against the budget.
    - **Eviction Logic**: When the budget is exceeded, the middleware marks the oldest tools as "Hidden."
    - **Prompt Injection**: The middleware modifies the outgoing tool schema list sent to the LLM.
*   **APIs / Interfaces:**
    - **Configuration**:
      ```yaml
      middleware:
        context_budget:
          max_active_tools: 10
          eviction_policy: "lru"
          notify_llm: true
      ```
*   **Data Storage/State:** Session-bound LRU state stored in-memory (volatile) or in the `Shared KV Store` (persistent).

## 5. Alternatives Considered
*   **Token-Based Budgeting**: Evicting based on total tokens instead of tool count. *Rejected* for initial version due to complexity in real-time token counting across different model tokenizers, but planned for V2.
*   **Manual Eviction**: Forcing the LLM to call an `evict_tool` function. *Rejected* as it wastes a model turn and increases the risk of the model forgetting to do so.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The budget manager must respect the Policy Firewall. An evicted tool that the user no longer has access to must not be re-searchable.
*   **Observability:** Expose "Context Budget Health" in the UI, showing which tools are currently "Active" vs "Evicted" in the current session.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
