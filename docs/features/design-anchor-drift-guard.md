# Design Doc: Anchor-Drift Guard (ADG)
**Status:** Draft
**Created:** 2026-07-24

## 1. Context and Scope
With the rise of "Living Documentation" and dynamic workspace scratchpads (e.g., Claude Code), agents rely on "Attention Anchors" to navigate large codebases. **Anchor-Drift** is a new exploit where malicious or hallucinating sub-agents subtly alter the environment, causing these anchors to point to stale or incorrect logic without triggering standard file-hash alerts.

MCP Any needs to implement ADG to perform syntax-aware, point-in-time comparison of agent anchors against a hardware-attested mission baseline. This ensures that when an agent says "I am editing the auth logic," it is actually editing the *current, verified* version of that logic.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement Inode-pinned "Anchor Snapshots" at the start of a mission.
    * Perform syntax-aware diffing (using Tree-sitter) between active anchors and baselines.
    * Automatically quarantine agent sessions when "Anchor Drift" exceeding a semantic threshold is detected.
* **Non-Goals:**
    * Replacing standard `git` version control.
    * Providing automated code fixes for the drift (handled by the agent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a specialist agent from "hallucinating" a deprecated API version as the current production baseline.
* **The Happy Path (Tasks):**
    1. User starts a mission; ADG creates "Anchor Snapshots" for all files in scope.
    2. Specialist Agent B is assigned to update `api_v2.go`.
    3. Hallucination occurs: Agent B starts reasoning against a cached `api_v1.go` structure.
    4. ADG detects that Agent B's reasoning trace (Attention Anchors) diverges from the syntactical baseline of `api_v2.go`.
    5. ADG triggers a "Drift Alert" and pauses Agent B's capability to call `write_file`.
    6. User/Mission Root reviews the drift and re-aligns Agent B.

## 4. Design & Architecture
* **System Flow:**
    `Filesystem` -> `ADG (Snapshot)` -> `Agent Reasoner` -> `Attention Trace` -> `ADG (Syntax Diff)` -> `Alert/Interdiction`
* **APIs / Interfaces:**
    * `GET /v1/anchors/status`: Check drift levels for active agent sessions.
    * `POST /v1/anchors/lock`: Pin a specific Inode and syntax tree as the mission baseline.
* **Data Storage/State:**
    Snapshots are stored as LZ4-compressed syntax trees in the `Shadow-FS` overlay.

## 5. Alternatives Considered
* **Binary File Hashing:** Rejected because it triggers on trivial changes (comments, formatting) that don't constitute "Cognitive Drift." Syntax-aware diffing is required for reasoning-level security.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The syntax-aware comparison must run in a detached WASM sandbox to prevent malicious code in the baseline from exploiting the ADG itself.
* **Observability:** Metrics for "MTTD" (Mean Time to Detect) drift will be surfaced to the UI.

## 7. Evolutionary Changelog
* **2026-07-24:** Initial Document Creation.
