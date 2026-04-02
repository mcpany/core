# Design Doc: Inode-Pinning Middleware
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
Path-based security validation is vulnerable to TOCTOU (Time-of-Check Time-of-Use) race conditions, such as the "Symlink-to-Inode Racing" (SIR) exploit patterns identified in recent ecosystem audits (CVE-2026-34812). If an agent validates a file path and then a malicious process swaps that path with a symlink to a sensitive host file before the actual execution, the sandbox is effectively bypassed.

The Inode-Pinning Middleware addresses this by moving from string-based path validation to hardware-bound file handle persistence.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind validated project-local configurations to their hardware Inodes.
    * Ensure that once a file is attested, its underlying Inode remains constant for the duration of the session.
    * Detect and interdict "File Swap" attempts where a path is redirected to a different Inode.
    * Provide OS-agnostic Inode mapping (mapping Linux Inodes to Windows file IDs).
* **Non-Goals:**
    * Will not prevent host-level filesystem corruption; focuses on preventing sandbox escapes via symlink racing.
    * Will not perform deep packet inspection of the file content (handled by the Content Validation Middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Developer / Enterprise IT Auditor
* **Primary Goal:** Ensure that an agent's project-local settings (e.g., `.claude/settings.json`) cannot be manipulated via symlink racing to grant unauthorized host access.
* **The Happy Path (Tasks):**
    1. Agent attempts to load a project-local configuration file.
    2. Inode-Pinning Middleware resolves the path and captures the hardware Inode.
    3. Middleware verifies the Inode against the "Initial Attestation Baseline."
    4. Hardware file handle is "pinned" to the session.
    5. All subsequent read/write operations for that path are performed via the pinned handle, ignoring any path-level string redirections.

## 4. Design & Architecture
* **System Flow:**
    [Agent Load] -> [Path Resolver] -> (Capture Inode) -> [Hardware Pinning Engine] -> [Pinned File Handle]
    [Attacker Swap Path] -> (Ignored by Pinned Handle)
* **APIs / Interfaces:**
    * Native integration with `pkg/shadowfs` to intercept all `open()` and `stat()` calls.
    * `PinHandle(path, inodeId)` internal API.
* **Data Storage/State:**
    * Maintains a `Map[Path]InodeID` in kernel-resident memory (where supported) or secure application memory.

## 5. Alternatives Considered
* **Continuous File-Watchers:** Rejected due to race conditions (the swap can happen faster than the event notification).
* **Mandatory Sandbox Mounts:** Rejected for developer-local workflows where agents need to operate on arbitrary project directories.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Neutralizes SIR and TOCTOU exploits.
* **Observability:** Visualized via the "Inode Security Monitor" in the UI, showing pinned Inodes and alerts for blocked racing attempts.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
