# Design Doc: Strict Schema Enforcement Middleware
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
As AI agents become more autonomous, the reliability of their tool calls is paramount. Current systems often suffer from "schema drift" or LLM hallucinations where tool arguments don't match the expected types. MCP Any needs to provide a robust enforcement layer that guarantees tool call integrity before they reach the underlying MCP server.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all tool calls and validate arguments against the registered JSON schema.
    * Provide detailed error feedback to the agent to allow for automatic correction.
    * Support `strict: true` mode similar to Claude Code's implementation.
* **Non-Goals:**
    * Automatically fixing the arguments (the agent must do this).
    * Modifying the tool's behavior itself.

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Agent Developer
* **Primary Goal:** Ensure the agent never crashes the MCP server with malformed inputs.
* **The Happy Path (Tasks):**
    1. Developer enables "Strict Mode" for a specific service in MCP Any.
    2. Agent attempts a tool call with a missing required field.
    3. MCP Any intercepts the call, detects the validation failure.
    4. MCP Any returns a structured 400 error with the specific JSON schema violation.
    5. Agent receives the error, corrects its output, and retries successfully.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [MCP Any Gateway -> Strict Schema Middleware -> Policy Engine] -> MCP Server`
* **APIs / Interfaces:**
    * New configuration flag: `strict_enforcement: boolean` in service metadata.
    * Middleware hook: `ValidateToolCall(ctx, call_args, schema)`
* **Data Storage/State:**
    * Schemas are cached in-memory from the initial MCP `list_tools` discovery.

## 5. Alternatives Considered
* **Client-side validation**: Rejected because we cannot trust all agent clients to implement validation correctly. Gateway-level enforcement ensures a "Zero Trust" perimeter.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevents injection attacks that rely on malformed JSON to bypass backend logic.
* **Observability:** Log all validation failures as "Reliability Traces" to help developers tune their prompts.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation.
