# Design Doc: Annotation-Based Policy Routing

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As the number of MCP tools grows, manually configuring policies for each tool becomes unmanageable. Inspired by Gemini CLI's recent addition of "Tool Annotation Matching," MCP Any needs a way to enforce security and routing policies based on metadata attached to tools. This allows for "Class-Based" security where tools tagged with `@security: destructive` are automatically quarantined regardless of which MCP server they originate from.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable policy enforcement based on `annotations` or `metadata` fields in the MCP Tool schema.
    *   Support global policy rules that target specific annotation patterns (e.g., "Block all tools with `@network: external` by default").
    *   Allow tool providers to declaratively signal their tool's security profile.
*   **Non-Goals:**
    *   Implementing a new annotation standard (we will follow existing MCP/OpenAPI/FastMCP conventions).
    *   Replacing the name-based Policy Firewall (annotations are an *additional* signal).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Platform Security Engineer.
*   **Primary Goal:** Ensure no tool capable of making external network requests can be called by an unauthenticated subagent.
*   **The Happy Path (Tasks):**
    1.  User configures a global Rego policy: `deny if input.tool.annotations["network"] == "external" && input.subject.level < 2`.
    2.  An MCP server (e.g., a Google Search tool) exposes a tool with metadata: `{"annotations": {"network": "external"}}`.
    3.  A subagent tries to call the search tool.
    4.  The Policy Engine detects the annotation and blocks the call because the subagent lacks "Level 2" clearance.

## 4. Design & Architecture
*   **System Flow:**
    - **Discovery**: During tool registration, MCP Any extracts annotations from the JSON-RPC schema.
    - **Policy Hook**: The `PolicyFirewall` middleware is updated to include the `tool.annotations` map in the input document passed to the Rego/CEL engine.
    - **Routing**: Metadata can also be used for "Smart Routing" where an agent request for "a secure database tool" can be resolved by searching for tools with the `@security: hardened` annotation.
*   **APIs / Interfaces:**
    - **Internal**: `ToolMetadata` struct updated to include `map[string]string` for annotations.
    - **Policy**: Rego context expanded with `input.tool.annotations`.
*   **Data Storage/State:** Annotations are stored alongside tool definitions in the system registry (Memory/KV Store).

## 5. Alternatives Considered
*   **Name-Based Regex**: Using regex on tool names (e.g., `*destructive*`). *Rejected* because it's fragile and doesn't capture intent.
*   **Separate Policy Config**: Maintaining a separate YAML file mapping tool names to security classes. *Rejected* as it doesn't scale with dynamic discovery.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Annotations provided by the tool itself are "Self-Attested." In high-security environments, these must be cross-referenced with the `MCP Provenance Attestation` to ensure the tool isn't lying about its security profile.
*   **Observability:** The UI should display tool annotations prominently, allowing users to filter and search tools by their metadata tags.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
