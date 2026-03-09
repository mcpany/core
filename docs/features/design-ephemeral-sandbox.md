# Design Doc: Ephemeral Tool Sandbox (MCP-Sandbox)

**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
As agents increasingly execute arbitrary code or CLI commands, the risk of host-level compromise (RCE, file theft) has become a critical blocker for adoption. Current "allow-list" approaches are insufficient against sophisticated prompt injection. MCP Any must provide a mechanism to execute high-risk tool calls in ephemeral, strictly isolated environments that are destroyed immediately after use.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide automated, containerized isolation for CLI and script-based tools.
    *   Support `containerd` and WASM as primary isolation runtimes.
    *   Enable "Zero-Persist" execution where no state survives the tool call unless explicitly saved to a shared mount.
    *   Minimize latency overhead for spawning sandbox environments (<100ms).
*   **Non-Goals:**
    *   Providing long-lived persistent containers (use standard service definitions for that).
    *   Replacing full-blown orchestration like Kubernetes (MCP Any handles local/node-level isolation).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Developer using OpenClaw subagents.
*   **Primary Goal:** Run a "Python Executor" tool that can process untrusted code without risking the host machine.
*   **The Happy Path (Tasks):**
    1.  User marks a tool in `config.yaml` with `isolation: sandbox`.
    2.  An agent calls the tool with a snippet of Python code.
    3.  MCP Any intercepts the call, pulls/starts a pre-warmed ephemeral container.
    4.  The code executes inside the container; output is streamed back to MCP Any.
    5.  The container is immediately destroyed.

## 4. Design & Architecture
*   **System Flow:**
    - **Sandbox Manager**: A core component that manages a pool of "pre-warmed" containers to reduce startup latency.
    - **Execution Driver**: Handles the bridging of stdio/network between the MCP Any process and the sandbox.
    - **Volume Mapping**: Strictly controlled mounts for specific input/output files.
*   **APIs / Interfaces:**
    - Service Config: `isolation: "host" | "sandbox" | "wasm"`
    - Sandbox Metrics: Integration with `Resource Telemetry Middleware`.
*   **Data Storage/State:** Sandboxes are stateless by default. Temporary state is handled via `/tmp` mounts that are wiped on container exit.

## 5. Alternatives Considered
*   **User-Level Virtualization (Firecracker)**: *Rejected* for local dev use cases due to higher complexity and resource overhead compared to `containerd`.
*   **Static Chroot/Jail**: *Rejected* because it doesn't provide enough isolation for modern multi-tenant or multi-agent environments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core "Defense in Depth" feature. Even if the Policy Firewall is bypassed, the sandbox prevents host access.
*   **Observability:** The UI should show "Sandbox Active" status and resource consumption per-sandbox.

## 7. Evolutionary Changelog
*   **2026-03-01:** Initial Document Creation.
