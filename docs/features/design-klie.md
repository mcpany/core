# Design Doc: Kernel-Level Intent Enforcement (KLIE) Gateway
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward fully autonomous local execution, the risk of "Runtime Compromise"—where a subagent or tool bypasses the application-level sandbox—has become a critical threat. Current security models rely on the agent runtime (e.g., Node.js, Python) to enforce policies. If the runtime is escaped, host-level access is granted.

The KLIE Gateway moves intent validation from the application layer to the container kernel using eBPF (Extended Berkeley Packet Filter). It ensures that every syscall initiated by an agent is verified against a hardware-attested "Mission Manifest" before it is executed by the CPU.

## 2. Goals & Non-Goals
* **Goals:**
    * Interdict unauthorized syscalls at the kernel level.
    * Bind process IDs to specific hardware-attested mission roots.
    * Provide zero-latency intent validation for high-frequency tool calls.
* **Non-Goals:**
    * Implementing a full container runtime.
    * Validating the semantic "correctness" of tool outputs (handled by AID Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise DevOps
* **Primary Goal:** Prevent an agent from performing a `connect()` syscall to an unauthorized IP, even if the agent's Node.js process is hijacked.
* **The Happy Path (Tasks):**
    1. User defines a mission-root manifest in `mcp.yaml`.
    2. MCP Any generates a BPF bytecode program tailored to that manifest.
    3. The BPF program is loaded into the host kernel and pinned to the agent's cgroup.
    4. The agent attempts an unauthorized `execve` or `connect`.
    5. The kernel interdicts the syscall and sends an async alert to MCP Any.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Process] -> [Syscall] -> [eBPF Hook (KLIE)] -> [Verification against Manifest Map] -> [Allow/Deny]`
* **APIs / Interfaces:**
    * `POST /v1/klie/load`: Load a mission manifest and attach it to a PID/Cgroup.
    * `GET /v1/klie/violations`: Stream kernel-level violation logs.
* **Data Storage/State:**
    * Mission manifests are stored in kernel-resident BPF Maps (Hash Tables) for O(1) lookups.

## 5. Alternatives Considered
* **gVisor (Sentry):** Rejected due to the 5-15% performance overhead of syscall interception in user-space.
* **Seccomp-BPF:** Rejected because it is too static; KLIE requires dynamic, context-aware intent validation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** eBPF programs are verified by the kernel verifier to ensure they cannot crash the host.
* **Observability:** KLIE events are exported via perf-buffers to the MCP Any logging pipeline.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
