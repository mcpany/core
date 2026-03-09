# Design Doc: AOTUI Semantic Projection Engine
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
With the rise of Agent-Oriented TUIs (AOTUI) in frameworks like OpenClaw, tool outputs designed for humans (raw JSON or verbose text) are no longer optimal. Agents need tool results projected into "Semantic Markdown"—a structured, minimal, and highly searchable format that emphasizes data relationships over presentation.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically transform upstream tool results (JSON/Text) into AOTUI-compliant Markdown.
    * Provide a configuration-driven way to define "Projection Templates" for different tools.
    * Reduce token usage by stripping non-essential metadata from tool responses.
* **Non-Goals:**
    * Building a full GUI renderer.
    * Altering the fundamental JSON-RPC transport (this is a payload transformation).

## 3. Critical User Journey (CUJ)
* **User Persona:** OpenClaw Agent (using AOTUI runtime)
* **Primary Goal:** Ingest tool results efficiently without bloating the context window with raw JSON noise.
* **The Happy Path (Tasks):**
    1. Agent calls a tool (e.g., `list_files`).
    2. MCP Any executes the tool and receives a complex JSON object from the upstream.
    3. The `SemanticProjectionEngine` applies a Markdown template.
    4. The agent receives a clean, structured Markdown table or list in the `content` field.
    5. Agent reasons over the Markdown to determine the next action.

## 4. Design & Architecture
* **System Flow:**
    - New Middleware: `ProjectionMiddleware` inserted at the end of the execution chain.
    - Template Engine: Uses Go `text/template` or a similar lightweight engine to map JSON fields to Markdown.
    - Default Projections: Fallback to a "Generic Semantic Table" for unconfigured tools.
* **APIs / Interfaces:**
    - `Project(data interface{}, template string) (string, error)`
* **Data Storage/State:**
    - Projection templates are stored in `config.yaml` under each service/tool definition.

## 5. Alternatives Considered
* **Client-side transformation**: Rejected because it requires every agent framework to implement the same logic. Server-side ensures consistency.
* **Hardcoded Projections**: Rejected as it violates MCP Any's "Configuration over Code" principle.

## 6. Cross-Cutting Concerns
* **Performance**: Template execution must be extremely fast to avoid adding latency to the agent loop.
* **Observability**: Traces will show both the raw upstream response and the projected Markdown for debugging.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
