# Design Doc: Schema Harmonization Middleware
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As AI agents from different ecosystems (Google Gemini, Anthropic Claude, OpenAI, etc.) connect to MCP Any, they encounter "Schema Friction." Each model family has slightly different requirements for tool definitions (e.g., naming conventions, schema structures like `inputSchema` vs `parameters`, or constraints on nesting).

MCP Any needs a **Schema Harmonization Middleware** that automatically detects the requesting model (via headers or user-agent) and transforms the standard MCP tool definitions into an optimized format for that specific LLM.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically sanitize tool names to match LLM-specific regex (e.g., Google requires `^[a-zA-Z0-9_-]+$`).
    * Transform `inputSchema` to the target model's expected format (e.g., `parameters` for OpenAI).
    * Inject model-specific "hints" or "system prompts" into tool descriptions to improve reliability.
    * Provide a configuration-driven way to add new model mappings.
* **Non-Goals:**
    * Changing the underlying logic of the upstream tool.
    * Providing a general-purpose JSON schema validator (outside of LLM constraints).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using Gemini CLI with MCP Any.
* **Primary Goal:** Use a local tool with complex parameters that originally used spaces or special characters in its name.
* **The Happy Path (Tasks):**
    1. Gemini CLI sends a `tools/list` request to MCP Any.
    2. The **Schema Harmonization Middleware** detects the request is from a Google model.
    3. It iterates through all tools, replaces spaces with underscores in names, and ensures all descriptions are within Google's length limits.
    4. Gemini CLI receives a "clean" list of tools and successfully registers them without errors.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[AI Agent / Client] -->|tools/list| Server[MCP Any Core]
        Server --> Registry[Service Registry]
        Registry -->|Raw Tools| Harmonizer[Schema Harmonization Middleware]
        Harmonizer -->|Detect Model| Rules[Model Ruleset]
        Rules -->|Transform| Harmonizer
        Harmonizer -->|Optimized Tools| Agent
    ```
* **APIs / Interfaces:**
    * `Transform(tool *mcp.Tool, model string) (*mcp.Tool, error)`: Internal Go interface for schema transformations.
    * Configurable rules in `server/config.yaml`.
* **Data Storage/State:**
    * Stateless middleware. All transformations are done in-memory per request.

## 5. Alternatives Considered
* **Manual Configuration**: Forcing users to rename tools manually in config. Rejected because it's high friction and breaks the "Universal Adapter" promise.
* **Client-side Transformation**: Let the client (e.g., Gemini CLI) handle it. Rejected because many clients are "dumb" and expect standard-compliant MCP, while models themselves have unique quirks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Ensure that renaming tools doesn't lead to "shadowing" or unauthorized access. The mapping between the "Internal Name" and "Exposed Name" must be strictly maintained.
* **Observability:** Log which transformations were applied to which tools for debugging "Missing Tool" issues.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
