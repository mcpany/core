# Design Doc: JIT Tool Sandbox
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of "Self-Healing" agents (e.g., OpenClaw) that propose real-time modifications to tool code, and the need to execute tools in isolated workspaces (e.g., Claude Code), MCP Any needs a way to execute potentially untrusted code without risking the host system. The JIT Tool Sandbox provides a transient, isolated environment for tool execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate tool execution from the host filesystem and network.
    * Support transient "healed" tools that exist only for a single session.
    * Provide strict syscall filtering for WASM or Docker-bound tools.
* **Non-Goals:**
    * Providing long-term persistence for sandboxed environments.
    * High-performance GPU passthrough for sandboxed tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Developer
* **Primary Goal:** Execute an agent-modified tool safely in a local environment.
* **The Happy Path (Tasks):**
    1. Agent detects a tool failure and proposes a "fix" (code snippet or WASM module).
    2. MCP Any intercepts the "healed" tool definition and assigns it to a JIT Sandbox.
    3. The sandbox is provisioned (e.g., a WASM runtime or a locked-down container).
    4. The tool executes within the sandbox boundaries.
    5. Results are returned to the agent, and the sandbox is destroyed.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [JIT Middleware] -> [Sandbox Provider (WASM/Docker)] -> [Tool Execution]`
* **APIs / Interfaces:**
    * `RegisterTransientTool(definition, sandbox_type)`
    * `ExecuteSandboxedTool(tool_id, args)`
* **Data Storage/State:**
    * Transient state only; no persistent storage allowed in the sandbox unless explicitly mounted as "Read-Only".

## 5. Alternatives Considered
* **Virtual Machines**: Rejected due to high overhead and slow provisioning for transient tool calls.
* **OS-level User Isolation (sudo/chroot)**: Rejected due to complexity in managing many transient users and potential for escape.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All sandboxed calls are subject to "Intent-Aware" verification before the sandbox is even provisioned.
* **Observability**: Sandbox logs are streamed back to the MCP Any audit log with a "Sandboxed" tag.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
