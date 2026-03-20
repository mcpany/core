# Design Doc: Project-Local Snapshot (PLSS) Sync
**Status:** Draft
**Created:** 2026-05-02

## 1. Context and Scope
AI agents operating on local project files (e.g., refactoring code, modifying configurations) face significant risks of environment corruption or security breaches via malicious hooks. "Project-Local Snapshot Sync" (PLSS) provides an OS-level bridge for near-instantaneous environment recovery, allowing agents to perform speculative actions with a safe, hardware-bound rollback mechanism. This implements the "Deterministic Sandbox Recovery" (DSR) patterns standardized by Claude Code.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified interface for OS-level snapshotting (e.g., Btrfs, APFS, or OverlayFS).
    * Implement "Deterministic Recovery" triggers based on subagent exit codes or policy violations.
    * Enable "Snapshot-and-Merge" workflows for parallel agent reasoning branches.
    * Support "State Checkpoints" for the Shared Blackboard that are synchronized with the filesystem snapshots.
* **Non-Goals:**
    * Implementing a full version control system (like Git).
    * Snapshotting the entire host OS (limited to project-local paths).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer using Agent Swarms
* **Primary Goal:** Automatically revert a complex code refactor if the "Auditor Agent" detects a security policy violation in the generated code.
* **The Happy Path (Tasks):**
    1. Parent Agent initiates a "Speculative Branch" for refactoring.
    2. PLSS Sync creates a project-local snapshot (Snapshot A).
    3. Subagent B performs the refactor and exits with a "Verify-Required" code.
    4. Auditor Agent C scans the changes and detects a violation.
    5. Auditor Agent C triggers a "Recovery" signal.
    6. PLSS Sync atomically reverts the project environment to Snapshot A.

## 4. Design & Architecture
* **System Flow:**
    `Snapshot Request` -> `OS-Specific Driver` -> `Manifest Creation` -> `Speculative Execution` -> `Consensus Check` -> `Commit OR Revert (DSR)`
* **APIs / Interfaces:**
    * `SnapshotController`: Manages the lifecycle of project-local snapshots.
    * `RecoveryTriggerListener`: Maps process exit codes and policy signals to DSR events.
    * `FilesystemDriver`: Abstract interface for Btrfs, APFS, and Docker-bound overlays.
* **Data Storage/State:**
    * Snapshot manifests and rollback logs are stored in the local `.mcpany/snapshots` directory, protected by the "Inode-Pinning" middleware.

## 5. Alternatives Considered
* **Git-Based Rollbacks**: Rejected due to high latency and inability to snapshot non-git-tracked artifacts or internal agent state.
* **Full VM Snapshotting**: Rejected as it is too resource-intensive and slow for rapid agent iterations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Snapshots are read-only and hardware-locked once created, preventing agents from tampering with their own recovery points.
* **Observability:** Snapshot events, diffs, and recovery triggers are visualized in the "Snapshot Rollback Manager."

## 7. Evolutionary Changelog
* **2026-05-02:** Initial Document Creation.
