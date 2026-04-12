# Design Doc: Kernel-Level Syscall Gating (KLSG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward high-density horizontal swarms (e.g., Claude Code Agent Teams, OpenClaw Kinetic Swarms), the reliance on shared memory (`mmap`) for performance creates a new class of "Sandbox Escape" vulnerabilities. If a compromised subagent can manipulate the memory segments of its parent or siblings, it can bypass intent-scoping and security filters.

MCP Any needs to move security from the application layer to the kernel. KLSG will intercept and validate high-risk system calls at the OS level, ensuring they are explicitly authorized by the active, hardware-attested mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate `mmap`, `ptrace`, and `shmget` syscalls in real-time.
    * Require hardware-attested (TPM/Secure Enclave) mission justification for every high-risk syscall.
    * Provide sub-millisecond latency for gated syscalls.
* **Non-Goals:**
    * Replacing general-purpose OS sandboxing (e.g., gVisor, Seccomp). KLSG is mission-aware, not just resource-aware.
    * Gating all syscalls (which would cause prohibitive performance degradation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Architect
* **Primary Goal:** Prevent a specialized "Code Reviewer" agent from using `ptrace` to hijack the process of a "Database Specialist" agent.
* **The Happy Path (Tasks):**
    1. The architect defines a "Mission Manifest" that explicitly denies `ptrace` for the "Code Reviewer" role.
    2. The Mission Manifest is TPM-signed and ingested by MCP Any.
    3. The "Code Reviewer" subagent attempts to execute `ptrace` on a sibling process.
    4. The KLSG kernel-module intercepts the call.
    5. KLSG queries the MCP Any security policy via a fast-path shared buffer.
    6. The syscall is denied, and the subagent session is forcefully terminated.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Process] -> [KLSG Kernel Hook (eBPF/LSM)] -> [MCP Any Fast-Path Policy Engine] -> [Hardware Root of Trust]
* **APIs / Interfaces:**
    * `/v1/security/klsg/attest`: Fast-path endpoint for syscall justification.
    * `/v1/security/manifest/sign`: Endpoint for TPM-signing mission manifests.
* **Data Storage/State:**
    * Policies are cached in kernel-accessible eBPF maps for sub-microsecond lookup.
    * Justification logs are stored in a hardware-locked SQLite "Truth Table."

## 5. Alternatives Considered
* **Standard Seccomp Profiles:** Rejected because they are static and mission-agnostic. They cannot distinguish between a "justified" and "unjustified" `mmap` during the same session.
* **gVisor Sentry:** Rejected due to the 20-30% performance tax, which is unacceptable for OpenClaw-style kinetic state handoffs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** KLSG utilizes hardware-bound identity tokens (SMI) to ensure that the process requesting the syscall is indeed the authorized subagent.
* **Observability:** Syscall denials are broadcast as high-priority events to the **CSAD Hub** and the **Local Security Violation Monitor**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
