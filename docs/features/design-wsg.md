# Design Doc: Worktree Sovereignty Guard (WSG)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of native worktree support in agent frameworks (e.g., Gemini CLI v0.36.0), AI agents can now manage parallel git-worktrees for independent task execution. However, this creates a significant security risk: "Worktree-to-Host" escapes. Malicious or compromised subagents can leverage shared `.git` metadata, specifically hooks and config files, to execute arbitrary code on the host machine or bypass standard filesystem sandboxing.

The Worktree Sovereignty Guard (WSG) acts as a specialized isolation broker that enforces hardware-attested, enclave-bound boundaries for all agent-spawned git-worktrees.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce strict enclave-bound isolation for all agent-spawned git-worktrees.
    * Cryptographically lock `.git/hooks` and other worktree metadata to prevent unauthorized modification.
    * Provide hardware-attested provenance for any changes to worktree configuration.
    * Neutralize worktree-to-host escape vectors via git metadata injection.
* **Non-Goals:**
    * Replacing standard git functionality (WSG is a security wrapper).
    * Managing remote git repository permissions (handled by the provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Safely delegate parallel feature development to 3 subagents using git-worktrees without exposing host `.git/hooks`.
* **The Happy Path (Tasks):**
    1. The Orchestrator initiates a "Worktree Branch" mission.
    2. MCP Any's WSG intercepts the `git worktree add` command.
    3. WSG spawns a hardware-attested Worktree Enclave for the new path.
    4. The `.git` file in the worktree is redirected to a cryptographically locked metadata vault managed by WSG.
    5. The subagent performs work within the enclave.
    6. Any attempt to write to the `.git/hooks` directory is interdicted and flagged by WSG.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Subagent[Subagent] -- git worktree add --> WSG[Worktree Sovereignty Guard]
        WSG -- Request Enclave --> TPM[TPM / Secure Enclave]
        WSG -- Lock Metadata --> Vault[Locked Metadata Vault]
        WSG -- Mount Worktree --> FS[Isolated Filesystem Path]
        Subagent -- write .git/hooks --> WSG -- BLOCKED --> Subagent
    ```
* **APIs / Interfaces:**
    * `WSG.ProvisionEnclave(path string, mission_root string)`
    * `WSG.ValidateMetadata(path string, signature []byte)`
* **Data Storage/State:**
    * Worktree metadata is stored in a TPM-sealed SQLite sidecar.

## 5. Alternatives Considered
* **Standard Chroot/Docker Isolation:** Rejected because standard containers often share the host's `.git` directory when worktrees are used, failing to isolate the metadata layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** WSG relies on the Hardware-Attested Mission Manifest (HAMM) to define authorized worktree paths.
* **Observability:** Worktree lifecycle events are logged to the `Mesh-Resident Lineage Tracker`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
