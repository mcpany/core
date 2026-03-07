# Design Doc: MCP App Native Hosting (Universal Bridge)

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
With the introduction of the MCP Apps extension (SEP-1865), MCP servers can now return interactive UI components. However, many current agent clients (CLI tools, legacy web interfaces) lack the native capability to render these sandboxed iframes. MCP Any will fill this gap by acting as a "Universal Bridge," natively hosting the UI assets and providing a stable URL that any client can use to render the tool's interactive interface.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a secure, sandboxed hosting environment for MCP App HTML/JS resources.
    *   Implement a bidirectional JSON-RPC bridge between the hosted UI and the parent MCP server.
    *   Expose a "Rendering URL" in tool results for clients that don't support inline iframes.
    *   Ensure strict isolation between different hosted apps.
*   **Non-Goals:**
    *   Creating a new UI framework (we adhere to SEP-1865).
    *   Hosting full-blown standalone web applications (focus is strictly on MCP-bound UI components).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using a CLI-based agent (e.g., Gemini CLI) that doesn't support interactive UI.
*   **Primary Goal:** View and interact with a complex data visualization returned by a tool.
*   **The Happy Path (Tasks):**
    1.  The agent calls a tool `get_infrastructure_map`.
    2.  The tool returns a response including an MCP App resource definition (`ui://...`).
    3.  MCP Any detects the UI resource, boots a local sandboxed instance, and generates a temporary signed URL: `http://localhost:50050/render/app-abc-123`.
    4.  The CLI agent displays the URL to the user.
    5.  The user clicks the URL, opening their browser to an interactive, real-time map that is securely bridged back to the agent's session.

## 4. Design & Architecture
*   **System Flow:**
    - **Resource Interceptor**: A new middleware intercepts `CallToolResult` messages containing `ui://` resources.
    - **App Sandbox Manager**: Manages the lifecycle of ephemeral web environments.
    - **UI-to-Server Bridge**: A WebSocket or PostMessage-based bridge that routes requests from the hosted UI through MCP Any to the upstream MCP server.
*   **APIs / Interfaces:**
    - `/render/{app_id}`: Secure endpoint for serving the UI component.
    - `postMessage` protocol extension for bidirectional tool calls from within the iframe.
*   **Data Storage/State:** Ephemeral session state mapping `app_id` to the underlying MCP session.

## 5. Alternatives Considered
*   **Client-Side Only Rendering**: Expecting all clients to implement SEP-1865. *Rejected* as it leaves CLI and legacy users behind.
*   **Static HTML Snapshots**: Returning a static HTML file. *Rejected* as it loses the bidirectional interactivity required for modern agent tools.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Iframes must be heavily sandboxed (CSP, no-top-navigation, limited permissions). Signed URLs must be short-lived and tied to the active session.
*   **Observability:** Track rendering performance and bridge latency in the management dashboard.

## 7. Evolutionary Changelog
*   **2026-03-07:** Initial Document Creation.
