# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Security-Hardened Tool Proxy (Sovereign Proxy)

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
The "OpenClaw Security Crisis" (CVE-2026-25253) demonstrated that autonomous agents executing shell commands or local tools without strict isolation are a massive security liability. Enterprises are banning these tools because they lack "Safe-by-Default" infrastructure. MCP Any needs to provide a universal "Sovereign Proxy" that intercepts tool execution and runs it in a cryptographically isolated, short-lived sandbox.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all "command-line" or "filesystem-access" tool calls.
    *   Execute these calls in an isolated environment (e.g., gVisor, Firecracker, or Docker with Seccomp).
    *   Provide a "Disposable Environment" for every high-risk session.
    *   Support "Read-Only" vs "Read-Write" mounting of local paths with granular control.
*   **Non-Goals:**
    *   Replacing the user's host OS.
    *   Implementing the LLM reasoning (MCP Any remains the execution layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Enterprise Developer.
*   **Primary Goal:** Allow an autonomous agent to "refactor a local repository" without the risk of the agent exfiltrating `~/.ssh` or executing a malicious payload on the host.
*   **The Happy Path (Tasks):**
    1.  User enables `SovereignProxy` in `config.yaml`.
    2.  Agent (e.g., OpenClaw) requests to call `execute_shell(command="rm -rf /")`.
    3.  MCP Any intercepts the call and spawns a gVisor-hardened container.
    4.  The container only has access to the `/workspace` directory (the target repo).
    5.  The malicious command fails or is contained within the sandbox.
    6.  The sandbox is destroyed immediately after the tool call completes.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: The `ProxyMiddleware` identifies "Unsafe" tools based on a pre-defined or Rego-defined policy.
    - **Provisioning**: MCP Any uses a local `SandboxManager` to spin up the isolation layer.
    - **Execution**: The tool call is forwarded to the adapter *inside* the sandbox.
    - **Cleanup**: Ephemeral state is wiped.
*   **APIs / Interfaces:**
    - **Sandbox Driver Interface**: Pluggable drivers for Docker, Firecracker, gVisor.
    - **Mount Configuration**: Standardized way to define which local folders are mapped into the proxy.
*   **Data Storage/State:** No persistent state inside the proxy; results are streamed back to MCP Any.

## 5. Alternatives Considered
*   **VM-per-Agent**: Too slow (seconds to boot). *Rejected* in favor of container-based or micro-VM isolation (milliseconds).
*   **OS-level User Scoping**: Too easy to escape via kernel exploits. *Rejected* for gVisor/Firecracker which provide a separate kernel or guest kernel.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The sandbox itself is untrusted. Network access from within the sandbox is disabled by default.
*   **Observability:** All commands executed inside the Sovereign Proxy are logged and audited in the `Audit Dashboard`.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
