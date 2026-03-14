# Design Doc: Pre-Flight Sandbox Validator
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The disclosure of CVE-2026-25725 (Claude Code) revealed a critical failure in partial sandboxing: if an agent can influence the environment *before* or *during* its execution in a way that affects subsequent runs (e.g., by creating a configuration file that didn't exist), it can escape the sandbox. MCP Any needs a "Pre-Flight Sandbox Validator" that ensures the environment is in a known, immutable state before any agentic action is taken.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a "Full-State Manifest" of the project environment at the start of every session.
    * Provide "Non-Existence Proofs" for sensitive files (e.g., `.claude/settings.json`, `.mcpany/config.yaml`) to prevent out-of-band injection.
    * Integrate with hardware-bound Inode pinning to prevent TOCTOU races.
    * Force-halt sessions if an unauthorized environmental mutation is detected.
* **Non-Goals:**
    * Managing the actual sandbox (e.g., Docker/Bubblewrap)—this service provides the *validation* and *manifest* logic for the sandbox.
    * General-purpose filesystem backup.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious developer using a subagent to refactor a complex repository.
* **Primary Goal:** Ensure that the subagent cannot silently inject a malicious configuration file that would compromise the host on the next agent start.
* **The Happy Path (Tasks):**
    1. User initiates an agent session via MCP Any.
    2. The **Pre-Flight Sandbox Validator** scans the project root and generates a signed manifest (SHA-256 of all files + list of "should-not-exist" paths).
    3. The agent executes. A compromised subagent attempts to create `.mcpany/hooks.json` to gain persistence.
    4. The filesystem middleware (connected to the Validator) detects a violation of the "Non-Existence Proof" for that path.
    5. MCP Any immediately terminates the subagent and alerts the user.

## 4. Design & Architecture
* **System Flow:**
    `Session Start` -> `Validator (Manifest Gen)` -> `Hardware Inode Pinning` -> `Agent Execution` -> `Continuous Manifest Monitoring`
    * **Manifest Generation**: Recursively hashes project files and records Inodes. Sensitive paths are explicitly marked as "Empty".
    * **Validation Engine**: Intercepts all `WRITE` and `CREATE` calls via the Filesystem Adapter.
* **APIs / Interfaces:**
    * `GenerateManifest(path string) (Manifest, error)`
    * `ValidateMutation(op FSOperation) (bool, error)`
* **Data Storage/State:**
    * Ephemeral, session-bound manifest stored in memory and optionally mirrored to a secure, write-once log.

## 5. Alternatives Considered
* **Read-Only Filesystems**: Too restrictive; agents often need to write code.
* **Post-Execution Cleanup**: Rejected because the damage (injection) is already done; "Pre-Flight" and "Real-Time" validation are required.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The validator assumes the agent *will* try to escape and treats every file creation as a potential threat.
* **Observability**: All manifest violations are logged with a full stack trace of the tool call that triggered them.

## 7. Evolutionary Changelog
* **2026-04-09:** Initial Document Creation.
