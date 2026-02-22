# Design Doc: MCP App UI Support
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
Anthropic's "MCP Apps" allows agents to surface UI elements. MCP Any, as a gateway, must support passing this UI metadata from upstream services to the AI host. This enables richer interactions beyond plain text.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Proxy UI metadata (iframe URLs, templates) from upstream adapters to the MCP response.
    *   Provide a standardized way for HTTP and gRPC adapters to declare UI capabilities.
*   **Non-Goals:**
    *   Hosting the UI assets ourselves (we proxy the URLs/templates).
    *   Validating the HTML/JS content (left to the host's sandbox).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Data Scientist using Claude with MCP Any
*   **Primary Goal:** View a live-generated chart from an internal REST API.
*   **The Happy Path (Tasks):**
    1.  User asks: "Show me the revenue trend."
    2.  Agent calls `get_revenue_chart` via MCP Any.
    3.  The Upstream HTTP adapter calls the internal API, which returns a JSON response containing a data URL or a link to a dashboard.
    4.  MCP Any transforms this into the `ui` field of the `CallToolResult`.
    5.  Claude renders the chart in an iframe.

## 4. Design & Architecture
*   **System Flow:**
    [Upstream API] -> [HTTP Adapter] -> [Response Transformer] -> [MCP Any Core] -> [AI Host]
*   **APIs / Interfaces:**
    *   Update `CallToolResult` schema to include the `ui` object.
    *   Add `ui_mapping` configuration to HTTP adapter.
*   **Data Storage/State:** Stateless proxying of UI metadata.

## 5. Alternatives Considered
*   **Returning raw HTML in text:** Rejected as it's insecure and doesn't leverage the host's native UI capabilities.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Enforce strict Content Security Policy (CSP) headers if we were hosting, but primarily rely on host-side sandboxing. Implement "Audit Logging" for all UI-bound URLs.
*   **Observability:** Log the occurrence of UI responses for usage analytics.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
