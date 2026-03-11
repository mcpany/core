# Design Doc: Zero-Trust Hook Runtime (Detached Sandbox)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
The recent OpenClaw security crisis (CVE-2026-25253) has demonstrated that simply validating configuration files is insufficient when agents can be tricked into executing malicious hooks. MCP Any needs a secondary layer of defense: a "Detached Sandbox" that executes all automated hooks and tool sequences in an isolated environment, ensuring that even if a malicious hook is approved or bypassed, it cannot damage the host system or access sensitive data.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a strictly isolated runtime for any executable hook or automated tool sequence.
    * Prevent unauthorized access to the host filesystem, network, and environment variables.
    * Use lightweight containerization or OS-level isolation (e.g., Docker-bound named pipes, cgroups, or Firecracker).
    * Implement "Zero-Trust" by default: no access is granted unless explicitly defined in a capability token.
* **Non-Goals:**
    * Sandboxing the entire AI agent (e.g., Claude Code itself). Focus is on the *hooks* it triggers.
    * Providing a full persistent VM for every hook (must be low-latency).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a multi-agent swarm for automated refactoring.
* **Primary Goal:** Execute a "post-refactor" linting hook safely without risking host file corruption.
* **The Happy Path (Tasks):**
    1. Agent triggers a `post-refactor` hook defined in the project config.
    2. MCP Any identifies the hook as an executable command.
    3. MCP Any spawns a "Detached Sandbox" instance (e.g., a minimal Docker container with a mounted read-only copy of the project root).
    4. The hook executes within the sandbox.
    5. MCP Any streams logs back to the agent and user.
    6. Sandbox is destroyed immediately after execution.

## 4. Design & Architecture
* **System Flow:**
    `MCP Any` -> `Sandbox Manager` -> `Isolated Runtime (Docker/Podman/nsjail)`
    1. **Trigger**: Middleware detects an executable action.
    2. **Provisioning**: `Sandbox Manager` requests a pre-warmed isolated environment.
    3. **Mounting**: Filesystem is mounted via `virtio-fs` or read-only binds, restricted to the `Project Root`.
    4. **Communication**: Inter-process communication via isolated named pipes or a secure gRPC bridge.
* **APIs / Interfaces:**
    * `Runtime.Execute(command, context)`: Internal interface for spawning sandboxed tasks.
    * `Config.SandboxPolicy`: Configuration schema for defining sandbox constraints (memory, CPU, disk).
* **Data Storage/State:**
    * Ephemeral storage within the sandbox, wiped on exit.

## 5. Alternatives Considered
* **Local Process with Restricted User**: Rejected as it is too easy to bypass on many OS configurations.
* **WebAssembly (Wasm)**: Considered for high-performance isolation, but rejected for now due to limited support for standard CLI tools used in hooks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The sandbox is the final line of defense. It follows the principle of "Fail-Closed"—if the sandbox cannot be provisioned securely, the hook does not run.
* **Observability**: Sandbox lifecycle events (start, stop, resource usage) are tracked in the `Audit Log`.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
