# Design Doc: Universal Browser Bridge (UBB)
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
With the rapid evolution of browser-use tools across major agent frameworks (OpenClaw, Claude Code, Gemini CLI), there is a growing fragmentation in how agents interact with browsers. Each framework uses its own set of MCP tools or native extensions, leading to redundant implementations and limited interoperability. MCP Any needs to provide a unified infrastructure layer that standardizes browser interaction as a shared MCP service.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized MCP toolset for browser interaction (navigation, form filling, element selection).
    * Support secure sharing of browser sessions across different agent frameworks via MCP Any.
    * Implement capability-based access control for browser tools (e.g., restrict to specific domains).
    * Ensure compatibility with existing browser extensions (Chrome, Firefox).
* **Non-Goals:**
    * Building a new browser engine (will use existing Playwright/Puppeteer/Selenium/Chrome extensions).
    * Automating websites that explicitly block automated access (unless bypasses are provided by the user).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Use a specialized "Research Agent" to find data on the web and then hand off the session to a "Booking Agent" to complete a transaction, without re-authenticating.
* **The Happy Path (Tasks):**
    1. Parent orchestrator initializes a "Browser Session" through MCP Any.
    2. Research Agent uses `ubb:navigate` and `ubb:extract_data` tools.
    3. MCP Any maintains the session state (cookies, local storage) in its internal store.
    4. Orchestrator hands the "Session Token" to the Booking Agent.
    5. Booking Agent uses `ubb:fill_form` and `ubb:click` within the same session to complete the task.

## 4. Design & Architecture
* **System Flow:**
    [Agent Framework] <--(MCP RPC)--> [MCP Any Gateway] <--(UBB Middleware)--> [Browser Controller (e.g., Playwright Service)] <--> [Real Browser/Headless]
* **APIs / Interfaces:**
    * `ubb:start_session(profile_id)`: Initializes a browser session.
    * `ubb:navigate(url)`: Navigates to a specific URL.
    * `ubb:click(selector)`: Clicks an element.
    * `ubb:type(selector, text)`: Fills an input.
    * `ubb:get_screenshot()`: Returns a base64 encoded screenshot.
* **Data Storage/State:**
    Session state (cookies, local storage) is stored in the MCP Any "Shared KV Store" and managed by the `UBB Middleware`.

## 5. Alternatives Considered
* **Native Framework Tools:** Rejected because it limits interoperability (e.g., Claude can't easily use OpenClaw's browser session).
* **Direct Browser-Use Protocol:** Rejected in favor of MCP to leverage existing agent integration patterns and security policies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Domain Whitelisting: Restrict browser access to a pre-approved list of domains.
    * Content Redaction: Sensitive information (e.g., passwords in form fields) can be redacted in screenshots/logs.
    * Session Isolation: Each agent swarm gets an isolated browser profile.
* **Observability:**
    * Visual Timeline: Store screenshots of each action in the MCP Any "Tool Activity Feed."
    * Latency Monitoring: Track DOM interaction times.

## 7. Evolutionary Changelog
* **2026-03-02:** Initial Document Creation.
