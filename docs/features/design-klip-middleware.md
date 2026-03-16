# Design Doc: Kernel-Level Inode Pinning (KLIP) Middleware
**Status:** Draft
**Created:** 2026-05-01

## 1. Context and Scope
The emergence of "Symlink-to-Inode Racing" (SIR) exploits (BoryptGrab Evolution) has rendered traditional path-based filesystem sandboxing insecure. Attackers can swap symlinks between an agent's "check" and "action" phases. KLIP (Kernel-Level Inode Pinning) moves validation from the path layer to the hardware layer by pinning file handles to their underlying hardware Inodes for the duration of an agent session.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-bound file handle persistence for all project-local configurations.
    * Neutralize TOCTOU (Time-of-Check to Time-of-Use) attacks involving symlinks.
    * Integrate with the Shadow-FS to ensure speculative edits are Inode-pinned.
    * Provide a "Deterministic Path-to-Inode" mapping.
* **Non-Goals:**
    * Modifying the host kernel (KLIP will use standard OS primitives like `O_PATH` and `fstat`).
    * Protecting files outside the authorized project root.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer.
* **Primary Goal:** Prevent a malicious repository from exfiltrating SSH keys via a symlink-race on `.claude/settings.json`.
* **The Happy Path (Tasks):**
    1. Agent attempts to read `.claude/settings.json`.
    2. KLIP Middleware resolves the path and captures the hardware Inode.
    3. KLIP "Pins" the file handle.
    4. An external process attempts to swap the file with a symlink to `~/.ssh/id_rsa`.
    5. KLIP detects the Inode mismatch (or uses the pinned handle) and blocks the access.

## 4. Design & Architecture
* **System Flow:**
    `Agent I/O` -> `KLIP Middleware` -> `OS Filesystem API`
* **APIs / Interfaces:**
    * `KLIP.pin(path)`
    * `KLIP.openPinned(handle_id)`
* **Data Storage/State:**
    * In-memory `Handle-to-Inode` registry, session-bound.

## 5. Alternatives Considered
* **Periodic Path Re-validation**: Rejected due to high overhead and inability to completely eliminate the race window.
* **Full Containerization**: Rejected as it imposes too much friction for local development workflows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: KLIP provides the "Negative Trust" guarantee that the environment has not been tampered with post-validation.
* **Observability**: The "KLIP Integrity Monitor" in the UI will show real-time alerts for blocked racing attempts.

## 7. Evolutionary Changelog
* **2026-05-01:** Initial Document Creation.
