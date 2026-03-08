# Design Doc: Air-Gapped Tool Runner (Sandboxed Execution)

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
The OpenClaw security crisis (CVE-2026-25253) demonstrated that local AI agents are highly vulnerable to Remote Code Execution (RCE) via malicious websites or poisoned tool marketplaces. The "Local is Safe" assumption is broken. MCP Any must provide a secure, isolated perimeter for tool execution. This feature introduces a sandboxed runtime that executes tools in a network-restricted environment (WASM or Docker), preventing unauthorized host access and data exfiltration.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Execute local tools in a cryptographically isolated environment.
    *   Enforce "Air-Gapped" network policies (disable network access for tool processes by default).
    *   Support ephemeral, one-time-use runtimes to prevent state persistence between calls.
    *   Provide a standard interface for multiple sandbox providers (e.g., WASM, Docker, Firecracker).
*   **Non-Goals:**
    *   Sandboxing remote HTTP-based MCP servers (they are already isolated by network boundaries).
    *   Providing a full OS virtualization layer.
    *   Rewriting existing tools to be WASM-compatible (we will focus on providing the runtime).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer using "Autonomous" Agents.
*   **Primary Goal:** Run a community-contributed "S3 Analysis" tool without risking local SSH keys.
*   **The Happy Path (Tasks):**
    1.  User configures the "S3 Analysis" tool in MCP Any with `runtime: sandboxed`.
    2.  The LLM triggers the tool call.
    3.  MCP Any spins up an ephemeral WASM/Docker container.
    4.  The container has access to specific AWS environment variables but *no* access to the host filesystem or network (except the AWS API).
    5.  The tool executes, returns the result, and the container is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Tool Request**: The Adapter receives a tool call and identifies the `sandboxed` requirement.
    - **Sandbox Provisioning**: The `SandboxManager` selects a provider (e.g., `WasmRunner` or `DockerRunner`).
    - **Isolation**: The provider creates an isolated environment, mounting only the necessary secrets and volumes.
    - **Execution**: The tool process runs within the sandbox, communicating with MCP Any via a secure pipe.
    - **Cleanup**: On completion, the sandbox is wiped.
*   **APIs / Interfaces:**
    - `Runner` Interface: `Execute(ctx, tool, args, sandbox_policy) (result, error)`
    - Config Schema: `sandboxed: bool`, `sandbox_type: "wasm" | "docker"`, `network: "none" | "restricted"`.
*   **Data Storage/State:** No persistent state is allowed in the sandbox. All necessary state must be passed via arguments or ephemeral mounts.

## 5. Alternatives Considered
*   **User-Level Virtualization (chroot/jails)**: Lacks the strong isolation and cross-platform consistency of WASM or Docker.
*   **Manual Sandbox Configuration**: Asking users to set up their own Docker containers. *Rejected* as it adds too much friction. MCP Any should manage the lifecycle.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the primary security control for autonomous execution. It mitigates the impact of "Clawjection" and RCE vulnerabilities.
*   **Observability:** Sandbox logs must be captured and streamed to the MCP Any audit log without allowing the tool to bypass logging.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
