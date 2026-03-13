# Design Doc: Inode-Aware Symlink Validator
**Status:** Draft
**Created:** 2026-04-01

## 1. Context and Scope
Recent vulnerabilities in agentic IDEs (e.g., Claude Code CVE-2026-34812) have demonstrated that simple path validation is insufficient for securing project-local configurations. Attackers can use complex symlink chains and "Normalization Fatigue" to trick agents into reading or writing files outside the authorized project root. MCP Any needs a robust, Inode-aware validation layer to enforce strict filesystem boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform recursive symlink resolution before any file access.
    * Validate that the final resolved path's Inode belongs to the authorized project subtree.
    * Provide a centralized "Normalization-as-a-Service" (NaaS) for all project-local tool calls.
* **Non-Goals:**
    * Replacing OS-level filesystem permissions (ACLs/chmod).
    * Managing remote filesystem mounts (S3/NFS) unless they expose standard Inodes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Open an untrusted repository containing a malicious `.claude/settings.json` that symlinks to `~/.ssh/id_rsa`.
* **The Happy Path (Tasks):**
    1. Agent attempts to load `.claude/settings.json`.
    2. MCP Any intercepts the request via the **Project Configuration Guard**.
    3. The request is passed to the **Inode-Aware Symlink Validator**.
    4. Validator resolves the symlink and detects that the target Inode is outside the project root.
    5. MCP Any blocks the request and alerts the user of a "Symlink Escape Attempt."

## 4. Design & Architecture
* **System Flow:**
    `[Agent Request] -> [Config Guard] -> [NaaS Normalizer] -> [Inode Validator] -> [Filesystem]`
* **APIs / Interfaces:**
    * `fs/normalize`: Returns a canonical, safe path.
    * `fs/validate-boundary`: Boolean check if a path (resolved) is within the project root.
* **Data Storage/State:**
    Caches Inode-to-Path mappings for the current project session to minimize `lstat` overhead.

## 5. Alternatives Considered
* **Chroot/Containerization:** Effective but often too heavy for lightweight local agents or specific tool calls.
* **Simple Realpath Checks:** Rejected due to TOCTOU risks and differences in `realpath` implementation across OS layers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The validator is the primary gatekeeper for all "Settings-as-Code" interactions.
* **Observability:** Logs all blocked traversal attempts with full resolution traces.

## 7. Evolutionary Changelog
* **2026-04-01:** Initial Document Creation.
