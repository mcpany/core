# Design Doc: WASM Tool Sandbox
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
With the rise of "Vibe Coding" and the discovery of over 1,000 malicious skills on ClawHub, the security of MCP servers is a critical concern. Standard MCP servers often run with the full privileges of the user (stdio) or have unrestricted network access (HTTP). This document proposes a WASM-based runtime that isolates MCP server execution, providing a secure "chroot-like" environment for tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure, isolated runtime for MCP servers using WebAssembly (WASM).
    * Restrict filesystem access to explicitly permitted directories.
    * Control network access through a capability-based model.
    * Prevent "tool-poisoning" from escalating to host-level compromise.
* **Non-Goals:**
    * Replacing all native stdio tools (some may still require native execution for performance or deep OS integration).
    * Providing a full virtualized OS environment (WASM is for process-level isolation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Safely execute a 3rd-party MCP server discovered on a marketplace without risking local data exfiltration.
* **The Happy Path (Tasks):**
    1. The user identifies a new MCP server (e.g., `sqlite-explorer`).
    2. MCP Any detects the server is unverified or requests high privileges.
    3. MCP Any compiles/runs the server inside the **WASM Tool Sandbox**.
    4. The sandbox restricts the server's access to only the `./data` directory.
    5. The agent uses the tool; any attempt by the tool to access `~/.ssh` or `0.0.0.0` is blocked by the runtime.

## 4. Design & Architecture
* **System Flow:**
    `LLM -> MCP Any Gateway -> Sandbox Manager -> WASM Runtime (Wazero/Wasmtime) -> Tool Logic`
* **APIs / Interfaces:**
    * `sandbox.execute(tool_name, inputs, constraints)`
    * `VirtualFS`: A mapping layer that translates WASM file calls to restricted host paths.
* **Data Storage/State:**
    * Ephemeral state within the WASM instance; persistent state handled via the `VirtualFS`.

## 5. Alternatives Considered
* **Docker Containers:** Rejected due to high overhead and slow startup times for short-lived tool calls.
* **gVisor/Sandboxed Processes:** Rejected due to complexity of cross-platform support and dependency on host kernel features.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Uses WebAssembly's inherent isolation. All host-calls (WASI) are intercepted and checked against the Policy Engine.
* **Observability:** Logs all attempted "out-of-bounds" calls by the sandboxed tool.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
