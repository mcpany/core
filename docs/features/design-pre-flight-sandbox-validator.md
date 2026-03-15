# Design Doc: Pre-Flight Sandbox Validator
**Status:** Draft
**Created:** 2026-04-09

## 1. Context and Scope
The persistent threat of Remote Code Execution (RCE) via project-local configuration files (CVE-2025-59536, CVE-2026-25725) has exposed a critical gap in agent sandboxing. Agents often assume that if a file doesn't exist at startup, it's safe to allow the agent to create it. However, malicious code running *within* a partially restricted sandbox can create these files to inject persistent hooks. The Pre-Flight Sandbox Validator addresses this by generating a "Full-State Manifest" before any agent execution occurs.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate a cryptographic manifest of the project environment (files, directories, environment variables) before agent initialization.
    * Provide "Non-Existence Proofs" for sensitive configuration paths (e.g., `.claude/settings.json`, `.env`).
    * Enforce immutability on the environment state during the agent's execution lifecycle.
    * Block any attempt to create or modify configuration files that were not part of the initial pre-flight attestation.
* **Non-Goals:**
    * Providing a runtime sandbox (it works *with* existing sandboxes like Docker or bubblewrap).
    * Validating the content of non-configuration files (e.g., source code).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer opening an untrusted repository with an AI agent.
* **Primary Goal:** Ensure that the agent cannot create persistent, malicious configuration hooks even if it finds an exploit in its runtime sandbox.
* **The Happy Path (Tasks):**
    1. User runs `mcpany start-agent --project ./untrusted-repo`.
    2. MCP Any performs a "Pre-Flight Scan" of `./untrusted-repo`.
    3. MCP Any detects that `.claude/settings.json` does NOT exist and records a Non-Existence Proof (hash of null).
    4. MCP Any starts the agent (e.g., Claude Code) within its native sandbox.
    5. A malicious subagent tries to create `.claude/settings.json` to inject a `SessionStart` hook.
    6. MCP Any's `File Proxy Middleware` intercepts the write, compares it against the Pre-Flight Manifest, and blocks the creation since it deviates from the "Safe State".
    7. User is notified of the blocked injection attempt.

## 4. Design & Architecture
* **System Flow:**
    `User` -> `Pre-Flight Validator` -> `Full-State Manifest` -> `Agent Runtime` -> `File Proxy` -> `Manifest Check`
* **APIs / Interfaces:**
    * `ManifestGenerator`: `Generate(path string) (Manifest, error)`
    * `ValidatorMiddleware`: `ValidateWrite(path string, data []byte) error`
* **Data Storage/State:**
    * Ephemeral, session-bound `Manifest` stored in memory and cryptographically signed by MCP Any.

## 5. Alternatives Considered
* **Read-only Filesystems**: Too restrictive; agents often need to write to `dist/`, `build/`, or temporary files.
* **Kernel-level Auditing (eBPF)**: High complexity and requires root privileges, which conflicts with our "Run Anywhere" principle.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Implements "Trust but Verify" for the entire project environment.
* **Observability**: All pre-flight manifests and validation failures are logged to the Audit Log.

## 7. Evolutionary Changelog
* **2026-04-23:** Update: Enforcing Strict Non-Existence Proofs (CVE-2026-25725).
    * **Context:** Discovery of persistent configuration injection in Claude Code (CVE-2026-25725) confirms that missing files are a primary attack vector.
    * **Architecture Adjustment:**
        * Elevating "Non-Existence Proofs" to a mandatory, hardware-signed requirement for all sensitive configuration paths (`.claude/settings.json`, `.env`, etc.).
        * The Validator will now explicitly block any `file_create` event for paths identified as "Absent" in the Pre-Flight Manifest, even if the agent sandbox allows it.
    * **Security Impact:** Neutralizes the "Absence-as-Exploit" pattern by ensuring the sandbox state remains identical to the pre-attested manifest.
* **2026-04-10:** Integrated with the **Deterministic Attestation Gateway** to support "Full-State Manifest" requirements for Claude Code deterministic boot compliance.
* **2026-04-09:** Initial Document Creation.
