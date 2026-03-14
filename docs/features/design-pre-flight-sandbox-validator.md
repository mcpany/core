# Design Doc: Pre-Flight Sandbox Validator
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The disclosure of CVE-2026-25725 revealed a critical failure in agent environment security: "Partial Sandboxing." Attackers can weaponize the absence of files (e.g., `.claude/settings.json`) or create race conditions (TOCTOU) to inject malicious configurations before an agent session is fully bound. MCP Any needs a "Pre-Flight Sandbox Validator" that generates an immutable manifest of the entire project-local environment before any agent execution begins.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a cryptographically signed "Full-State Manifest" of the project environment.
    * Provide "Non-Existence Proofs" for sensitive configuration paths.
    * Implement hardware-bound Inode pinning to prevent TOCTOU escapes.
    * Block agent execution if the environment diverges from the pre-flight manifest.
* **Non-Goals:**
    * Providing general-purpose file syncing or versioning.
    * Managing remote cloud environments (focused on project-local security).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer working in a shared, potentially compromised repository.
* **Primary Goal:** Ensure the agent only sees and uses verified local configurations.
* **The Happy Path (Tasks):**
    1. User initiates an agent session via MCP Any.
    2. **Pre-Flight**: MCP Any scans the project root, hashing all `.json`, `.yaml`, and `.sh` files.
    3. **Snapshot**: A manifest is generated, including "NULL" entries for sensitive files that *should not* exist.
    4. **Binding**: MCP Any locks the file handles using hardware-bound Inode pinning.
    5. **Execution**: The agent runs. Any attempt to swap a file or create a shadowed config is detected and blocked.

## 4. Design & Architecture
* **System Flow:**
    `Init Session` -> `Manifest Generator` -> `Inode Pinning Engine` -> `Agent Runtime` -> `Continuous Integrity Monitor`
* **APIs / Interfaces:**
    * `GenerateManifest(root string) (Manifest, error)`
    * `VerifyEnvironment(manifest Manifest) error`
* **Data Storage/State:**
    * Ephemeral, session-bound manifests stored in memory.
    * Optional "Trust-Graph" persistence for recurring projects.

## 5. Alternatives Considered
* **Read-only Filesystem Mounts**: Rejected because agents often need to write to specific directories (e.g., `dist`, `logs`), and managing fine-grained RO/RW overlays is complex and OS-dependent.
* **Polling-based Integrity Checks**: Rejected due to high overhead and the race-condition window between polls.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements the principle of "Verify, then Bind." No tool call is allowed until the environment's structural integrity is proven.
* **Observability**: Manifest generation logs and divergence alerts are streamed to the Security Dashboard.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
