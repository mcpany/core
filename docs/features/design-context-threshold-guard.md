# Design Doc: Universal Context Threshold Guard
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
Large Language Models (LLMs) have finite context windows. As MCP ecosystems scale to hundreds of tools, the tool definitions alone can consume a majority of the available tokens (the "Context Bloat" problem). While Claude Code has introduced a proprietary 10% threshold for tool search, MCP Any requires a **Universal, Model-Agnostic** solution that protects context windows for any LLM (Ollama, Gemini, GPT-4, etc.) by dynamically switching between "Full Manifest" and "Lazy Discovery" modes.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor tool schema volume (token count) in real-time.
    * Automatically transition to "Lazy Discovery" mode (hiding detailed schemas and providing a search tool) when a user-defined threshold is met.
    * Provide a consistent "Discovery Tool" interface across all models.
* **Non-Goals:**
    * Modifying the underlying LLM's token counting logic.
    * Caching tool execution results (this is handled by other middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a local model with a small context window (e.g., Llama 3 8B on a laptop).
* **Primary Goal:** Use a library of 200 tools without the agent forgetting the initial prompt due to context overflow.
* **The Happy Path (Tasks):**
    1. User configures a 5% context threshold in MCP Any.
    2. User adds several large MCP servers (OpenAPI/gRPC).
    3. MCP Any detects that the tool schemas now exceed 5% of the target model's context.
    4. Upon the next `tools/list` call, MCP Any replaces the 200 schemas with a single tool: `mcpany_search_tools`.
    5. The LLM remains "lean" and uses the search tool only when it needs a specific capability.

## 4. Design & Architecture
* **System Flow:**
    * **Telemetry Hook**: A middleware hook that runs before `tools/list` responses are sent to the client.
    * **Token Estimation**: Uses a fast tokenizer (e.g., Tiktoken or a simple heuristic) to estimate the manifest size.
    * **Mode Switcher**: If `size > threshold`, the response is transformed into the "Lazy" format.
* **APIs / Interfaces:**
    * Config options: `context_guard.threshold_percent`, `context_guard.model_max_tokens`.
* **Data Storage/State:**
    * Threshold settings are persisted in the global config.

## 5. Alternatives Considered
* **Static Lazy Loading**: Always use lazy loading. *Rejected* because it adds an extra "hop" (search call) for small toolsets where it's not needed.
* **Client-Side Truncation**: Letting the IDE/Client handle it. *Rejected* because MCP Any aims to be the "Universal Bus" that protects the agent regardless of the client.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The search tool must respect the same Policy Firewall rules as the full manifest.
* **Observability:** Dashboard widgets will show current "Context Pressure" and transition events.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation. Inspired by Claude's 10% threshold but expanded for universal model support.
