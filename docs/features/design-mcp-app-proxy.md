# Design Doc: MCP App UI Proxy

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The emergence of the "MCP Apps" standard allows MCP servers to return interactive UI components (dashboards, forms, visualizations) instead of just plain text or JSON. However, different agent frontends (Web, CLI, Messaging apps) have vastly different rendering capabilities. MCP Any needs a middleware layer that intercepts these UI components and translates them into the optimal format for the connected client, ensuring a consistent "App" experience across all platforms.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept `mcp_app/ui_component` payloads from upstream MCP servers.
    *   Translate standardized UI components (e.g., `Button`, `TextField`, `Chart`) into target-specific formats (HTML/React for Web, TUI elements for CLI, Markdown/Interactive Buttons for Messaging).
    *   Provide a "fallback" rendering (text-based) for clients that don't support rich UI.
    *   Manage UI state and interactions (callbacks) between the client and the upstream server.
*   **Non-Goals:**
    *   Creating a new UI framework. We follow the emerging MCP Apps schema.
    *   Handling heavy-weight asset hosting (e.g., large video files).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Agent User on WhatsApp.
*   **Primary Goal:** Fill out a "Travel Reimbursement" form provided by an MCP-native HR tool.
*   **The Happy Path (Tasks):**
    1.  User asks their OpenClaw agent to "Submit travel expenses."
    2.  OpenClaw calls the `submit_expense` tool via MCP Any.
    3.  The HR MCP server returns an MCP App UI payload containing a form.
    4.  `MCP App UI Proxy` detects the WhatsApp client and translates the form into a series of interactive WhatsApp buttons/messages.
    5.  User interacts with the WhatsApp messages.
    6.  `MCP App UI Proxy` collects the responses and sends them back to the HR MCP server as a tool interaction.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: Middleware sits in the `tools/call` response path.
    - **Detection**: Identifies client capabilities via session metadata (e.g., `User-Agent` or `Transport-Type`).
    - **Transformation**: Uses a registry of `UIAdapters` (WebAdapter, CLIAdapter, MessagingAdapter) to convert the payload.
    - **Callback Routing**: Maps client-side UI actions (clicks, submits) back to the appropriate MCP `resources/subscribe` or `notifications` channels.
*   **APIs / Interfaces:**
    - **Internal**: `UIComponentTranslator` interface.
    - **External**: Extensions to the MCP session protocol to negotiate UI capabilities.
*   **Data Storage/State:** Temporary UI session state stored in the `Shared KV Store` to handle asynchronous multi-step forms.

## 5. Alternatives Considered
*   **Client-Side Rendering Only**: Forcing every client to implement the full MCP App spec. *Rejected* because messaging apps and CLIs cannot be easily updated to support dynamic React-like components.
*   **Server-Side Image Rendering**: Rendering the UI to an image and sending it to the client. *Rejected* due to lack of interactivity and high bandwidth/latency.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** UI components must be sanitized to prevent "UI Injection" or "Clickjacking" attacks. Only allow a whitelisted set of component types.
*   **Observability:** Log the transformation latency and "fallback" rates (where a rich UI had to be downgraded to text).

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
