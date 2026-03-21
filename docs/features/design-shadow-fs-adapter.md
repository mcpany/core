# Design Doc: Shadow-FS Virtualization Adapter
**Status:** Draft
**Created:** 2026-04-28

## 1. Context and Scope
As agents gain the ability to perform complex file manipulations (e.g., refactoring entire modules), the risk of "destructive hallucinations" or malicious "Shadow Edits" increases. A single mistake in a regex replacement or an intentional "BoryptGrab" payload can corrupt a project's source code beyond repair.

The **Shadow-FS Virtualization Adapter** provides a safe, transactional layer for filesystem operations. It allows agents to work on a virtualized "Overlay" of the real filesystem. Edits are only committed to the physical disk after the agent (or the user) verifies the results.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a Copy-on-Write (CoW) virtual filesystem for agent sessions.
    * Support "Atomic Commits" of a batch of file changes.
    * Enable "Speculative Verification" where tests run against the Shadow-FS before committing.
    * Provide a visual "Diff Viewer" for the user to approve changes.
* **Non-Goals:**
    * Replacing the host OS filesystem.
    * High-performance database storage (optimized for source code and config files).

## 3. Critical User Journey (CUJ)
* **User Persona:** Full-Stack Developer using a Swarm of agents.
* **Primary Goal:** Perform a large-scale refactor across 50 files without risking project corruption.
* **The Happy Path (Tasks):**
    1. Agent requests to "Open Shadow-FS Session."
    2. MCP Any creates a virtual overlay of the project root.
    3. Agent performs 50 file edits; edits are written to the Overlay, not the Disk.
    4. Agent runs `npm test` inside the Shadow-FS context.
    5. Tests pass; Agent sends a `commit-shadow` request.
    6. User reviews the "Mega-Diff" in the A2UI dashboard.
    7. User clicks "Commit"; Shadow-FS merges the overlay back to the physical disk.

## 4. Design & Architecture
* **System Flow:**
    * **Read Path:** Check Shadow-FS Overlay -> If missing, Read from Physical Disk.
    * **Write Path:** Always write to Shadow-FS Overlay.
    * **Commit Path:** Atomic `rename` or `patch` operation from Overlay to Physical Disk.
* **APIs / Interfaces:**
    * `shadow_fs_open(session_id)`
    * `shadow_fs_write(path, content)`
    * `shadow_fs_commit(session_id)`
* **Data Storage/State:**
    * Ephemeral storage in `/tmp/mcp-any/shadow-fs/[session_id]`.
    * Metadata manifest (JSON) tracking changed files and hashes.

## 5. Alternatives Considered
* **Git Branches:** Too heavy for rapid agent iteration. Some projects aren't using Git.
* **Full Docker Containers:** High overhead; slow to spin up for small file edits.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shadow-FS acts as a sandbox, preventing agents from "escaping" the project root during the execution phase.
* **Observability:** Users can "peek" into the Shadow-FS at any time to see the agent's progress.

## 7. Evolutionary Changelog
* **2026-04-29:** Addressing "BoryptGrab" payload injection by introducing "Speculative Integrity Quorums" for commits. FS changes now require a background validation pass (e.g., automated test run or monitor agent attestation) before being merged from the shadow overlay.
* **2026-04-28:** Initial Document Creation.
