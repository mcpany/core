# Design Doc: Isolated Skill Sandbox (ISE)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
The rapid proliferation of MCP tools from unverified sources (OpenClaw Marketplace) has introduced significant security risks, including SSRF, token leakage (CVE-2026-25253), and unauthorized filesystem access. Currently, MCP tools often run with the same privileges as the host agent. MCP Any must provide an isolation layer that ensures tools operate within a restricted "Secure Enclave."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Execute MCP tool processes in ephemeral, network-isolated containers/sandboxes.
    *   Restrict tool access to the host network (mitigate SSRF).
    *   Provide granular filesystem mounting (read-only by default).
    *   Use high-performance IPC (Named Pipes or gRPC-over-Unix-Sockets) for tool-to-gateway communication.
*   **Non-Goals:**
    *   Building a full container orchestrator (e.g., Kubernetes).
    *   Sandboxing high-performance local tools that require direct hardware access (e.g., local GPU inference) without explicit user opt-in.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Developer.
*   **Primary Goal:** Run a community-contributed "SQL Optimizer" MCP tool without risking the tool accessing internal company databases or environment variables.
*   **The Happy Path (Tasks):**
    1.  User adds the tool to MCP Any with `isolated: true` flag.
    2.  When the agent calls the tool, MCP Any spins up an ISE container.
    3.  The tool executes in an environment with no network access and only the target database file mounted.
    4.  The result is returned to the agent, and the container is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Trigger**: Tool call received.
    - **Provisioning**: ISE Manager selects a sandbox template (Docker, Firecracker, or WebAssembly).
    - **Execution**: The MCP server is started inside the sandbox.
    - **Communication**: MCP Any communicates with the sandboxed server via a mounted Unix socket or named pipe, bypassing the host's TCP/IP stack.
*   **APIs / Interfaces:**
    - `ISEManager.Spawn(toolID, config)`
    - `ISEManager.Terminate(sessionID)`
*   **Data Storage/State:** Sandboxes are stateless. Any persistent state must be explicitly requested and provided via the `Shared KV Store`.

## 5. Alternatives Considered
*   **Process-level Sandboxing (e.g., seccomp, AppArmor)**: *Rejected* as the primary method due to cross-platform inconsistency and complexity in managing network namespaces for individual processes.
*   **WebAssembly (WASM)**: *Accepted* as a lightweight alternative for simple tools, but Docker/Firecracker remains the primary target for full-featured MCP servers.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** ISE is the enforcement point for the "Intent-Aware" policy engine.
*   **Observability:** ISE logs are streamed to the Audit Log with "Sandbox-ID" metadata to trace potential exploit attempts.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
