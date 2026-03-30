# Design Doc: Continuous Inode-to-Intent Binding (CIIB)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
The "Symlink-Shadowing" exploit has exposed a critical TOCTOU (Time-of-Check to Time-of-Use) vulnerability in AI agent file interactions. Even if a path is validated during the discovery phase, an attacker can replace that path with a symlink to a sensitive host file just before the agent executes a tool.

Continuous Inode-to-Intent Binding (CIIB) eliminates this window by cryptographically binding the target file's hardware Inode to the active mission-root intent at the exact micro-second of execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide real-time verification of file Inodes during tool execution.
    * Bind Inode identity to the cryptographically signed mission-root intent.
    * Support "Last-Mile" attestation for all filesystem-based MCP tools.
    * Integrate with the Resident Integrity Monitor (RIM) for continuous session-wide verification.
* **Non-Goals:**
    * Replacing the Pre-Flight Sandbox Validator (CIIB is the "last-mile" defense).
    * Managing general file permissions (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a subagent from reading `~/.ssh/id_rsa` via a symlink-shadowing attack during a legitimate `read_file` call.
* **The Happy Path (Tasks):**
    1. Agent A requests `read_file("config.json")`.
    2. The CIIB Provider resolves the path to Inode 12345.
    3. The CIIB Provider verifies that Inode 12345 matches the "Authorized Manifest" for the current mission intent.
    4. The file handle is opened, and the Inode is re-verified at the kernel level during the `openat` syscall.
    5. If a symlink was injected (pointing to Inode 99999), the CIIB interdicts the call and triggers a security violation.

## 4. Design & Architecture
* **System Flow:**
    `[Tool Call] -> [CIIB Middleware] -> [Kernel Inode Check] -> [Mission Manifest Validation] -> [Execution]`
* **APIs / Interfaces:**
    * `VerifyFileIdentity(path, mission_id) -> error`
    * `BindInode(fd, mission_id) -> error`
* **Data Storage/State:**
    * Mission-bound Inode allow-list stored in hardware-protected volatile memory.

## 5. Alternatives Considered
* **Path-only Validation:** Rejected due to TOCTOU vulnerability (symlinks can change under the path).
* **Full Sandbox Virtualization (gVisor):** Effective but introduces 20-30% performance overhead. CIIB provides kernel-level security with <1% overhead.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CIIB is the final gate in the Zero-Trust filesystem path.
* **Observability:** Blocked shadowing attempts are visualized in the "Symlink Security Inspector."

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
