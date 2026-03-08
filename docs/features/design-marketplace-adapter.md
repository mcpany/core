# Design Doc: Universal Marketplace Discovery Adapter

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As ecosystem leaders like Anthropic (Claude Code) and OpenClaw (ClawHub) introduce centralized marketplaces for agents and tools, the burden of tool management shifts from manual configuration to registry-based discovery. MCP Any needs a standardized way to "subscribe" to these heterogeneous marketplaces and expose their offerings as native MCP tools, ensuring users have access to the latest capabilities without manual installation overhead.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a unified interface to subscribe to multiple agent/tool marketplaces.
    *   Implement "Marketplace Adapters" for Claude Code (official/private) and ClawHub.
    *   Support on-demand installation and proxying of marketplace-discovered tools.
    *   Maintain strict security isolation for marketplace-sourced plugins.
*   **Non-Goals:**
    *   Hosting a marketplace (MCP Any is the adapter/gateway).
    *   Bypassing marketplace authentication or licensing.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using a local LLM with MCP Any.
*   **Primary Goal:** Use a specialized "GitHub Analysis" tool discovered via the Claude Marketplace.
*   **The Happy Path (Tasks):**
    1.  User adds the marketplace: `mcpany marketplace add claude-official`.
    2.  User searches for the tool: `mcpany marketplace search github-analysis`.
    3.  User installs/registers the tool: `mcpany marketplace install [slug]`.
    4.  The tool is now available to the local LLM as `marketplace_claude_github_analysis`.

## 4. Design & Architecture
*   **System Flow:**
    - **Marketplace Discovery Service**: Polls registered marketplaces for metadata and schemas.
    - **Registry Middleware**: Maps marketplace-specific schemas (e.g., Claude's plugin manifest) to the MCP protocol.
    - **Sandbox Execution**: Marketplace-sourced logic is executed within a hardened JS-native path or isolated container.
*   **APIs / Interfaces:**
    - `marketplace/list`: Returns a list of available tools across all subscribed marketplaces.
    - `marketplace/subscribe`: Adds a new marketplace URL/catalog.
*   **Data Storage/State:** Marketplace metadata is cached in the `Shared KV Store` to enable fast `tools/list` responses.

## 5. Alternatives Considered
*   **Manual Tool Porting**: Asking users to manually convert marketplace tools to MCP Any configs. *Rejected* as it doesn't scale with the ecosystem growth.
*   **Direct Marketplace Integration in Agents**: Letting the LLM talk directly to marketplaces. *Rejected* because it lacks the centralized policy and security controls of MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All marketplace-sourced tools are treated as "Unverified" until they pass an Attestation check. They are automatically bound to the most restrictive capability scopes by default.
*   **Observability:** Marketplace provenance is tracked in the `Supply Chain Attestation Viewer` in the UI.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
