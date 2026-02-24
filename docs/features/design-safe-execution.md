# Design Doc: Safe-Execution Middleware (Sandbox)
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
Recent critical vulnerabilities (notably CVE-2026-0755 in `gemini-mcp-tool`) have demonstrated that MCP servers can be vectors for Remote Code Execution (RCE) via command injection. MCP Any, as a universal gateway, must protect the host environment by ensuring all tool calls are executed within a restricted security perimeter.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all MCP tool execution requests.
    * Sanitize all input arguments against common injection patterns.
    * Execute tools in an isolated process/container (sandbox).
    * Restrict filesystem and network access for tool execution.
* **Non-Goals:**
    * Rewriting vulnerable MCP servers.
    * Providing a full virtualized OS for every tool call (must be high performance).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Run third-party MCP servers without risking host compromise.
* **The Happy Path (Tasks):**
    1. User configures a third-party MCP server in MCP Any.
    2. User enables "Safe-Execution" for this service.
    3. An agent triggers a tool call with potentially malicious input.
    4. Safe-Execution middleware intercepts the call, validates the input.
    5. Tool is executed in a restricted sub-process with limited syscall access.
    6. Result is returned safely to the agent.

## 4. Design & Architecture
* **System Flow:**
    `LLM Agent -> MCP Any Gateway -> Safe-Execution Middleware (Validation) -> Sandbox Runner -> MCP Server Tool Execution -> Result -> Agent`
* **APIs / Interfaces:**
    * `ISandboxRunner`: Interface for different sandboxing backends (Docker, gVisor, WebAssembly).
    * `ValidationRule`: Pluggable rules for input sanitization.
* **Data Storage/State:**
    * ephemeral state within the sandbox.

## 5. Alternatives Considered
* **Static Analysis of MCP Servers**: Rejected due to complexity and the "black box" nature of many servers.
* **Manual Input Validation**: Rejected as it is error-prone and doesn't protect against zero-day exploits in the server logic.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Principles of Least Privilege applied to the sandbox environment.
* **Observability:** Logs all blocked execution attempts and sanitization triggers.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation.
