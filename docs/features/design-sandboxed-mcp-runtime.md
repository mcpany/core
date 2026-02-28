# Design Doc: Sandboxed MCP Runtime

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
The recent OpenClaw security crisis (unauthorized autonomous command execution and file access via malicious skills) has highlighted the critical risk of running agent-controlled tools directly on a host machine. MCP Any needs a way to execute these tools in a restricted, isolated environment. This feature introduces a "Sandboxed Runtime" that sits between the MCP Any gateway and the tool execution layer.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide ephemeral, isolated execution environments for tool calls.
    *   Prevent unauthorized host-level filesystem and shell access by agents/tools.
    *   Enable granular resource limits (CPU, RAM, Network) for tool execution.
    *   Support both containerized (Docker) and lightweight (WASM) backends.
*   **Non-Goals:**
    *   Protecting against all possible side-channel attacks.
    *   Providing a full virtual desktop for agents.
    *   Replacing existing MCP servers (it wraps/hosts their execution).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Allow developers to use third-party "skills" or MCP servers without risking the integrity of their development machines.
*   **The Happy Path (Tasks):**
    1.  User configures a new MCP server in MCP Any with `sandbox: true`.
    2.  An agent requests to call a tool on that server.
    3.  MCP Any spins up an ephemeral, isolated container/WASM instance.
    4.  The tool call is executed inside the sandbox.
    5.  Results are piped back to the gateway; the sandbox is destroyed or returned to a clean state.
    6.  The host filesystem remains untouched except for explicitly mounted volumes.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: The Policy Firewall identifies a tool call requiring a sandbox.
    - **Orchestration**: The Sandbox Manager selects a provider (Docker or WASM) and prepares the environment.
    - **Execution**: The JSON-RPC request is forwarded into the sandbox via a secure bridge (e.g., named pipes or Unix domain sockets).
    - **Clean-up**: The environment is wiped post-execution to ensure no persistent state remains.
*   **APIs / Interfaces:**
    - `SandboxProvider` interface: `Create()`, `Execute()`, `Destroy()`.
    - Configuration schema update:
      ```yaml
      services:
        - name: "untrusted-tool"
          sandbox:
            enabled: true
            provider: "docker"
            image: "mcp-any-runtime:latest"
            mounts: ["/tmp/app:/app:ro"]
      ```
*   **Data Storage/State:** No persistent state within the sandbox unless explicitly configured via mounts.

## 5. Alternatives Considered
*   **Virtual Machines (VMs)**: *Rejected* due to high overhead and slow spin-up times for per-tool execution.
*   **Process Namespacing (chroot/unshare)**: *Rejected* as it is harder to maintain across different OSes and provides less isolation than modern containers/WASM.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox itself is a "Zero Trust" zone. No network access is granted by default.
*   **Observability:** All IO within the sandbox is logged to the Audit Stream, providing a "flight recorder" for agent actions.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation in response to the OpenClaw security crisis.
