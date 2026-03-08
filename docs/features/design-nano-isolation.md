# Design Doc: Nano-Isolation Middleware (Ephemeral Tool Runtimes)

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The trust gap in the AI agent ecosystem (e.g., "OpenClaw loses its head") has reached a critical point. Standard tool execution runtimes are too persistent, allowing for potential side-channel attacks or state leakage between subagent calls. "Nano-Isolation" (inspired by the NanoClaw project) aims to provide the highest level of security by ensuring that every single tool call happens in a completely fresh, ephemeral environment that is destroyed immediately after execution.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement per-call ephemeral isolation for all command-based and script-based tools.
    *   Support WASM (WebAssembly) as a primary lightweight isolation tier.
    *   Support micro-containers (e.g., via `runc` or `gVisor`) for tools requiring full OS capabilities.
    *   Zero persistence: ensure no filesystem or memory state survives between tool calls unless explicitly piped through the Shared KV Store.
*   **Non-Goals:**
    *   Replacing traditional MCP server runtimes (this is an optional middleware layer).
    *   Optimizing for long-running processes (by definition, these tools are short-lived).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Orchestrator.
*   **Primary Goal:** Execute a potentially untrusted third-party MCP tool without risking the integrity of the host system or other agent sessions.
*   **The Happy Path (Tasks):**
    1.  User configures a tool with `isolation: nano`.
    2.  Agent calls the tool.
    3.  MCP Any spins up a per-call WASM runtime.
    4.  The tool executes, returns result, and the WASM runtime is immediately purged.
    5.  Any attempt by the tool to write to a "shadow" file or maintain a background process fails as the environment is destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Isolation Provider Interface**: A pluggable interface for different backends (`wasmtime`, `docker-slim`, `gvisor`).
    - **Snapshotting (Optional)**: Support for "Resume-from-Spec" where a clean environment is quickly restored from a verified base image for every call.
    - **Resource Budgeting**: Strict CPU/Memory limits enforced at the isolation layer.
*   **APIs / Interfaces:**
    - Configuration property: `execution_policy: { type: "nano", provider: "wasm" }`
*   **Data Storage/State:** Input/Output is passed via standard MCP JSON-RPC, but the internal execution state is strictly non-persistent.

## 5. Alternatives Considered
*   **Persistent Sandboxes**: Keeping a sandbox alive for the duration of a session. *Rejected* as it allows for cross-call poisoning.
*   **User-Level Isolation (UID/GID)**: Using standard Linux permissions. *Rejected* as it is often bypassable via kernel exploits in shared-kernel environments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the ultimate "Zero Trust" execution model. No trust is granted to the tool's runtime environment.
*   **Observability:** Log the exact isolation provider and resource usage for every nano-isolated call.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
