# Design Doc: Project Configuration Security Guard
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As agents increasingly rely on project-local configuration files (e.g., `.claude/settings.json`, `.openclaw/config.yaml`) for project-specific instructions and hooks, a new attack vector has emerged. Malicious actors can commit configuration files containing harmful "hooks" or "auto-execute" commands that run automatically when a collaborator opens the project with an AI agent. MCP Any needs to act as a validating proxy that intercepts these configurations and ensures they meet strict security policies before being ingested by any agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept project-local configuration files before they reach agent runtimes.
    * Sanitize and validate "hook" and "auto-execute" definitions against a Zero-Trust policy.
    * Require explicit user attestation for any non-standard or executable configuration change.
    * Provide a secure, isolated runtime for approved hooks.
* **Non-Goals:**
    * Replacing the agent's internal configuration management entirely.
    * Validating the *intent* of natural language prompts within the config (focus is on executable/structural safety).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer collaborating on a public repository.
* **Primary Goal:** Prevent RCE when using an AI agent (e.g., Claude Code) on a repository with malicious project-local settings.
* **The Happy Path (Tasks):**
    1. Agent attempts to read `.claude/settings.json`.
    2. MCP Any intercepts the read request.
    3. MCP Any identifies a new `hook` definition in the file.
    4. MCP Any pauses the request and notifies the user via the UI/CLI.
    5. User reviews the hook, sees it is a `git checkout` command (safe), and approves it.
    6. MCP Any caches the attestation and allows the agent to proceed.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any (File Proxy Middleware)` -> `Filesystem`
    1. **Interception**: MCP Any wraps the filesystem access tools used by agents.
    2. **Analysis**: The `Config Validator` parser identifies known "danger zones" (hooks, commands, environment overrides).
    3. **Policy Check**: Matches against `mcp-policy.rego`.
    4. **Attestation**: If a policy violation occurs, the request is suspended via the `HITL Middleware`.
* **APIs / Interfaces:**
    * `POST /v1/attest/config`: User endpoint to approve/deny a flagged configuration block.
    * `FileSystem.read` hook: Intercepts `read` calls to specific filenames.
* **Data Storage/State:**
    * `attestations.db`: SQLite database storing hashes of approved configuration blocks and their associated user decisions.

## 5. Alternatives Considered
* **Agent-side plugins**: Rejected because it requires modifying every agent framework (Claude, OpenClaw, etc.).
* **OS-level file permissions**: Too blunt; doesn't allow granular inspection of the *content* of the configuration.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All hooks are treated as untrusted until attested. Approved hooks are executed in a `Detached Sandbox` (cgroups/Docker) with restricted network and disk access.
* **Observability**: Every intercepted config and its attestation status is logged to the `Audit Log`.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
* **2026-03-10:** Added Content-Addressable Attestation and Team Registry integration.

### Update: 2026-03-10 - Content-Addressable Attestation Mesh
**Context:** Today's market sync revealed that "one-off" user approvals are insufficient for enterprise teams. Malicious configs can be subtly modified to bypass previous attestations.
**Architecture Adjustment:**
*   **Hashing Engine**: Implementing a SHA-256 content-addressable hashing system for all intercepted configuration blocks.
*   **Team Registry**: Introducing an optional integration with a central `mcp-registry.json` (or remote API) that pre-approves known-good configuration hashes for entire organizations.
*   **Immutable Attestations**: Moving from "File-Path" based trust to "Hash-Based" trust. If a file's content changes by a single byte, it is treated as a new, untrusted entity.
**Security Impact:** Prevents "Attestation Bypass" attacks where a malicious actor overwrites an approved hook with a similar-looking but harmful one.
