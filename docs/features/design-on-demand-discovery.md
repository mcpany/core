# Design Doc: On-Demand MCP Discovery Middleware

**Status:** Draft
**Created:** 2026-02-25

## 1. Context and Scope
As the number of available MCP tools grows, agents face "context pollution"—where the token cost of including all tool definitions exceeds the benefit, often consuming over 70% of the context window. Claude Code recently introduced "MCP Tool Search" to solve this. MCP Any needs a universal version of this capability that works across all LLMs and transport layers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Reduce tool-related token overhead by at least 80%.
    *   Support on-demand tool discovery via similarity search (semantic or regex).
    *   Maintain a "Tool Registry" capable of indexing 10,000+ tools.
    *   Ensure compatibility with standard MCP `tools/list` while providing an optimized `tools/search` extension.
*   **Non-Goals:**
    *   Executing tools (this is handled by the existing adapter layer).
    *   Automatic generation of tool schemas (schemas must be provided by upstreams).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Developer with 50+ MCP servers connected.
*   **Primary Goal:** Use a specific specialized tool without overwhelming the LLM's context window.
*   **The Happy Path (Tasks):**
    1.  User connects 50 MCP servers to MCP Any.
    2.  MCP Any indexes all tool descriptions into a local vector/search index.
    3.  User asks: "I need to analyze our AWS S3 costs for last month."
    4.  LLM calls `mcp_any_search_tools(query="AWS S3 cost analysis")`.
    5.  MCP Any returns the top 3 most relevant tool definitions.
    6.  LLM now has the specific schemas needed to call the `get_s3_usage_report` tool.

## 4. Design & Architecture
*   **System Flow:**
    - **Indexing**: On startup/hot-reload, MCP Any crawls all registered upstreams and populates a search index (using Bleve or a simple BM25 implementation).
    - **Search API**: A new MCP tool `mcpany_search_tools` is exposed to the agent.
    - **Lazy Loading**: Tool schemas are omitted from the initial `tools/list` call if "lazy mode" is enabled, replaced by a single discovery tool.
*   **APIs / Interfaces:**
    ```json
    {
      "method": "tools/call",
      "params": {
        "name": "mcpany_search_tools",
        "arguments": {
          "query": "string",
          "limit": 5
        }
      }
    }
    ```
*   **Data Storage/State:** Persistent index stored in the `MCPANY_DB_PATH` (SQLite/FTS5).

## 5. Alternatives Considered
*   **Upfront Compression**: Trying to compress schemas using LLM-specific techniques. *Rejected* due to lack of universality and risk of losing critical schema details.
*   **Manual Whitelisting**: Forcing users to manually enable/disable tools. *Rejected* as it creates high friction and doesn't scale.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Search results only include tools the user/session is authorized to see based on the Policy Firewall.
*   **Observability:** Log search queries and hit rates to optimize index performance and identify "missing tool" patterns.

## 7. Evolutionary Changelog
*   **2026-02-25:** Initial Document Creation.
*   **2026-03-06:** Evolution to **Lazy-MCP 2.0**.
    - **Context**: Market sync reveals that simple search is insufficient; agents need "Intent-Aware" discovery to minimize attack surface and token usage.
    - **Update**: Introducing semantic intent-matching for discovery. Tools are only "hydrated" into the agent's context if they align with the high-level intent verified by the User/Policy Engine.
