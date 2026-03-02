# Design Doc: Universal Browser Agent Adapter

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
With the release of Gemini 3.1 Pro and its experimental "Browser Agent," the scope of agent tools has expanded from structured APIs to the entire web DOM. MCP-native agents (like Claude or local OpenClaw instances) currently lack a standardized way to interact with web pages at this level of fidelity. MCP Any needs to provide a Universal Browser Adapter that exposes high-fidelity browser interaction (navigation, clicking, typing, DOM extraction) as a set of standard MCP tools.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a standardized set of MCP tools for browser interaction (`browser_navigate`, `browser_click`, `browser_type`, `browser_extract_dom`).
    *   Support multiple browser backends (e.g., Playwright, Puppeteer, Selenium).
    *   Implement "Visual Grounding" by providing screenshot-to-element mapping.
    *   Ensure compatibility with Gemini-style browser agent actions for cross-model reasoning.
*   **Non-Goals:**
    *   Building a full browser (this is an adapter for existing browser automation engines).
    *   Replacing traditional web fetch tools (this is for high-fidelity interaction, not just content extraction).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Autonomous Agent Developer.
*   **Primary Goal:** Enable a Claude-based agent to navigate to a specific web page, interact with a dynamic form, and extract the resulting data.
*   **The Happy Path (Tasks):**
    1.  User configures the Browser Adapter in MCP Any, specifying a Playwright backend.
    2.  The agent calls `browser_navigate(url="https://example.com/form")`.
    3.  The agent calls `browser_click(selector="#submit-button")`.
    4.  The agent calls `browser_extract_dom(selector=".result-container")` to get the output.
    5.  MCP Any manages the persistent browser session and provides feedback for each action.

## 4. Design & Architecture
*   **System Flow:**
    - **Session Persistence**: MCP Any maintains a pool of active browser contexts tied to agent session IDs.
    - **Tool Mapping**: The `BrowserAdapterMiddleware` translates MCP tool calls into backend-specific commands (e.g., Playwright API).
    - **Visual Feedback**: Every interaction can optionally return a base64 screenshot for the model's visual reasoning.
*   **APIs / Interfaces:**
    - **MCP Tools**: `browser_navigate`, `browser_click`, `browser_type`, `browser_scroll`, `browser_screenshot`, `browser_extract_dom`.
*   **Data Storage/State:** Browser context state (cookies, local storage) is persisted in the session context, allowing for long-running workflows.

## 5. Alternatives Considered
*   **Direct Playwright Tool for Agents**: Letting agents call Playwright directly. *Rejected* because it requires the agent to understand Playwright's complex API and lacks centralized security/policy control.
*   **Generic Web Fetching**: Using simple HTTP requests. *Rejected* as it cannot handle modern SPA/JavaScript-heavy applications.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Browser interactions are governed by the Policy Firewall. Access to sensitive domains (e.g., internal corp tools) can be restricted via "Seatbelt Profiles."
*   **Observability:** The UI provides a "Live Browser Stream" or "Interaction History" showing exactly what the agent did in the browser.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
