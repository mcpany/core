# Design Doc: Strict WebView Policy Enforcer
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the rise of agentic capabilities embedded directly into web browsers (e.g., Google Chrome's Gemini Live/Glic panel), a new class of "Glic Jacking" vulnerabilities (CVE-2026-0628) has emerged. Malicious browser extensions or websites can exploit insufficient policy enforcement in WebView components to escalate privileges or hijack agent-to-tool communications. MCP Any needs a specialized middleware to enforce strict origin and capability binding for any tool call originating from a web-based agent interface.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate the cryptographic origin of all tool requests originating from browser environments.
    * Implement "Domain-Bound Capabilities": Ensure a web-based agent can only access tools explicitly authorized for its specific origin.
    * Provide a secure, isolated bridge for WebView-to-MCP communication.
    * Log and block any unauthorized cross-origin tool execution attempts.
* **Non-Goals:**
    * Securing the browser itself (MCP Any focuses on the gateway/tool layer).
    * Replacing existing CSRF/CORS protections (it complements them).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an AI-integrated browser panel.
* **Primary Goal:** Ensure that a malicious browser extension cannot use the Gemini panel to trigger local MCP tools (like `fs:write`) on the developer's machine.
* **The Happy Path (Tasks):**
    1. Gemini Panel (Origin: `chrome://glic`) sends a tool request to MCP Any.
    2. `Strict WebView Policy Enforcer` intercepts the request.
    3. Middleware checks the `Sec-Fetch-Site` and `Origin` headers against a cryptographically signed allowlist.
    4. Middleware verifies that `chrome://glic` is authorized for the requested tool (e.g., `google_search`, but not `local_exec`).
    5. Request is allowed and passed to the upstream MCP server.

## 4. Design & Architecture
* **System Flow:**
    `WebView Agent` -> `Browser Bridge` -> `MCP Any (WebView Middleware)` -> `Policy Firewall` -> `MCP Tool`
    1. **Origin Verification**: Uses custom headers and signed attestation tokens to verify the calling environment.
    2. **Capability Mapping**: Maps browser origins to specific "Safe Toolsets" (e.g., `web_only`, `read_only`).
    3. **Enforcement**: Integrates with the `Policy Firewall` (Rego) to deny any request that violates origin-to-tool bindings.
* **APIs / Interfaces:**
    * `X-MCP-Origin-Attestation`: Header containing a signed blob from the trusted browser component.
    * `GET /v1/security/webview/policies`: UI/CLI endpoint to manage origin-to-tool mappings.
* **Data Storage/State:**
    * `webview_policies.db`: SQLite table mapping origins to allowed capability scopes.

## 5. Alternatives Considered
* **Standard CORS**: Too permissive; doesn't handle the complexity of browser-internal `chrome://` or `extension://` origins effectively for local tool access.
* **OIDC/OAuth**: High overhead for local, browser-internal agent communication.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Follows the principle of "Verify, then Trust." No web-based origin is trusted by default.
* **Observability**: All "Blocked Origin" events are flagged in the `Audit Log` with high severity.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
