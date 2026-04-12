# Design Doc: Task-Bound Inode Locking (TBIL) Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of distributed agents sharing project workspaces across multiple devices, "Racing Symlinks" and configuration-injection attacks have become a primary vector for sandbox escapes. Existing path-based validation is insufficient in environments with high-frequency file synchronization.

The Task-Bound Inode Locking (TBIL) Middleware, inspired by Claude Code v3.2.1, provides kernel-level cryptographic binding between filesystem Inodes and specific mission-root tasks. This ensures that once a file is validated for a task, its underlying physical resource cannot be swapped or tampered with.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically bind filesystem Inodes to mission-root task IDs.
    * Prevent TOCTOU (Time-of-Check to Time-of-Use) attacks during tool execution.
    * Ensure resource sovereignty across distributed filesystem sync events.
* **Non-Goals:**
    * Implementing a full distributed filesystem.
    * Managing low-level disk I/O outside the validation path.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent a subagent from using a symlink race to overwrite a critical project configuration while another agent is reading it.
* **The Happy Path (Tasks):**
    1. Agent A requests access to `.claude/settings.json` for Task X.
    2. TBIL Middleware resolves the path to Inode 12345 and locks it to Task X.
    3. A malicious Subagent B attempts to replace the file with a symlink to `/etc/shadow`.
    4. When Agent A performs a tool call using the file, TBIL verifies that the current Inode still matches 12345 and is authorized for Task X.
    5. If the Inode changed (due to the symlink swap), the tool call is interdicted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Tool Call] --> B[TBIL Middleware]
        B --> C{Verify Inode Lock}
        C -- Valid --> D[Kernel FD Access]
        C -- Invalid --> E[Security Interdiction]
        F[File Watcher] --> B
    ```
* **APIs / Interfaces:**
    * `tbil.LockInode(path, taskID) -> LockID`: Binds an Inode to a task.
    * `tbil.ValidateAccess(fd, taskID) -> Bool`: Confirms the file descriptor matches the locked Inode.
* **Data Storage/State:**
    * **Kernel-Bound Lock Table:** Local map of Inodes to hardware-attested task IDs.

## 5. Alternatives Considered
* **Pure Path Validation:** Rejected because paths can be easily manipulated via symlinks between check and use.
* **Full Filesystem Snapshots:** Rejected as too heavy for high-frequency tool calls; TBIL provides granular, low-latency protection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** TBIL is the final gate before kernel-level file access. It relies on the "Hardware-Linked Inode Pinning" strategic pillar.
* **Observability:** Integrated with the "Inode Security Monitor" for real-time visualization of locked resources.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
