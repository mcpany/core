# Design Doc: Model-Aware Schema Translator
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
AI models from different providers (Google Gemini, Anthropic Claude, OpenAI) have divergent requirements for tool/function-calling schemas. For instance, Gemini requires strict alphanumeric/underscore names, while Claude is more flexible. Today, developers must manually map MCP tool definitions to each model's specific flavor. MCP Any will solve this by implementing an automated, middleware-based translation layer.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically sanitize tool names and descriptions to meet specific LLM provider constraints.
    * Dynamically reformat JSON schema definitions (e.g., handling `anyOf`, `oneOf`, or specific type requirements) for each provider.
    * Inject model-specific metadata (e.g., Gemini's `FunctionDeclaration` fields).
* **Non-Goals:**
    * Translating the *logic* of the tool (only the interface/schema).
    * Predicting which model the user *should* use.

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Developer using multiple LLMs.
* **Primary Goal:** Use the same MCP server with both Gemini and Claude without manual config changes.
* **The Happy Path (Tasks):**
    1. User connects a standard MCP server to MCP Any.
    2. User's Gemini-based agent requests tools.
    3. MCP Any detects the `User-Agent` or a specific header indicating Gemini.
    4. The `SchemaTranslator` middleware automatically sanitizes a tool named `fetch-data-from-api` to `fetch_data_from_api` and reformats the schema.
    5. The tool call succeeds in the Gemini environment.

## 4. Design & Architecture
* **System Flow:**
    `MCP Server -> Registry -> [SchemaTranslator Middleware] -> LLM Adapter (Gemini/Claude/OpenAI)`
* **APIs / Interfaces:**
    * Internal: `TranslateSchema(mcp_schema, provider_type) -> provider_schema`
* **Data Storage/State:**
    * Stateless translation; mapping rules are stored in a provider-specific configuration file.

## 5. Alternatives Considered
* **Manual Mapping in Config**: Rejected as it scales poorly and increases developer burden.
* **Provider-Specific Adapters**: Rejected as it duplicates logic; a unified middleware is more maintainable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Sanity checks ensure that translation doesn't accidentally expose internal metadata or bypass security hooks.
* **Observability**: Translation logs will show how a schema was modified, helping developers debug schema-related tool call failures.

## 7. Evolutionary Changelog
* **2026-03-02**: Initial Document Creation.
