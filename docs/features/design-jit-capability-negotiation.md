# Design Doc: JIT Capability Negotiation Middleware
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As the number of available MCP tools grows (often 100+), providing the full tool schema to an LLM leads to "Context Window Bloat," increasing costs and degrading reasoning performance. MCP Any needs a way to dynamically serve only the tools relevant to the agent's current intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce the token count of tool schemas in the LLM context.
    * Provide a mechanism for agents to "negotiate" tool access on-the-fly.
    * Integrate with similarity-based tool search (Lazy-MCP).
* **Non-Goals:**
    * Automatically identifying agent intent (this is handled by the agent or a separate classifier).
    * Replacing the base MCP protocol (this is an extension/middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Orchestrator / Swarm Developer
* **Primary Goal:** Minimize context usage while maintaining access to a vast tool library.
* **The Happy Path (Tasks):**
    1. Agent sends a partial intent or keyword to MCP Any.
    2. JIT Middleware performs a similarity search across the Service Registry.
    3. Middleware prunes the total tool list to the top 5-10 most relevant tools.
    4. MCP Any returns the pruned schema to the agent.
    5. Agent calls the tool normally.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any (JSON-RPC: tools/list + intent metadata)` -> `JIT Middleware` -> `Similarity Engine` -> `Service Registry` -> `Pruned Schema` -> `Agent`
* **APIs / Interfaces:**
    * Extension to `tools/list`: Adds an optional `intent` or `query` parameter.
    * New `tools/negotiate` endpoint for explicit capability requests.
* **Data Storage/State:**
    * Uses a vector index (e.g., Faiss or simple cosine similarity in memory) to store tool descriptions.

## 5. Alternatives Considered
* **Static Layering:** Grouping tools into static "profiles." *Rejected* as it's too rigid for dynamic agent needs.
* **Agent-Side Filtering:** Having the agent filter its own tools. *Rejected* because the agent still needs the schemas to filter them, which doesn't solve the context bloat.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pruning must still respect the underlying Policy Engine. A tool cannot be negotiated if the agent doesn't have the base permission.
* **Observability:** Log which tools were suggested vs. which were actually called to refine the similarity engine.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
