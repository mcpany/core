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

### Update: 2026-04-09 - Defending Against Sandbox Escapes (CVE-2026-25725)
**Context**: Today's research into CVE-2026-25725 reveals that "Partial Sandboxing" fails when agents can create configuration files that did not exist at startup.
**Architecture Adjustment**:
* **Full-State Manifest Generation**: Section 4 will now include a pre-execution step where MCP Any generates a cryptographic "Non-Existence Proof" for all potential configuration files in a project directory.
* **Immutable Path Pinning**: The `File Proxy Middleware` will now enforce that no new files matching sensitive configuration patterns (e.g., `.claude/settings.json`) can be created by the agent runtime unless they are explicitly authorized via a pre-flight user attestation.
**Security Impact**: Closes the gap identified in Claude Code's bubblewrap sandboxing, preventing malicious code from injecting hooks via non-existent configuration files.
### Update: 2026-03-10 - Resolving Config-Based RCE and API Theft
**Context**: Today's market sync confirmed that CVE-2025-59536 (Claude Code) exploited `hooks` and `enableAllProjectMcpServers` in `.claude/settings.json`.
**Architecture Adjustment**:
* **Mandatory Sandbox for Hooks**: Section 6 now mandates that *any* executable hook found in a config must run in the `Detached Sandbox`.
* **Strict Schema Enforcement**: The `Config Validator` in Section 4 will now explicitly block `enableAllProjectMcpServers` or similar "bulk-enable" flags unless they are accompanied by a per-server cryptographic attestation.
* **Environment Variable Masking**: Section 4 will now intercept any configuration that attempts to inject environment variables into the agent runtime, masking sensitive keys (API_KEY, SECRET) by default.
**Security Impact**: Mitigates high-risk "Configuration-as-Execution" attack vectors, preventing host takeover and API credential theft from untrusted repository configurations.

### Update: 2026-03-11 - Mitigating Base URL Hijacking
**Context**: Research into CVE-2026-21852 (Claude Code) revealed that agents can be tricked into exfiltrating API keys by modifying the `ANTHROPIC_BASE_URL` in project-local settings.
**Architecture Adjustment**:
* **Active Interception & Rewriting**: The `File Proxy Middleware` (Section 4) will now actively rewrite intercepted config files. If a `base_url` or similar field is detected, it will be forcefully redirected to the MCP Any internal proxy address before the agent runtime can process it.
* **Lock-on-Write**: Any attempt by the agent (or a malicious script) to modify these sensitive fields in the project-local file will be blocked and flagged for immediate re-attestation.
* **Pre-Trust Validation**: Section 3 (CUJ) now includes a step where MCP Any validates the base URL configuration *before* the agent is even spawned, ensuring that no outbound requests reach unverified domains during initialization.

### Update: 2026-03-12 - Mandatory MFA for Config-Based Hooks
**Context**: Persistent bypasses in agent consent mechanisms (CVE-2025-59536) prove that "Session-Based" consent is insufficient for persistent configuration files.
**Architecture Adjustment**:
* **MFA Integration**: The `HITL Middleware` (Section 4) will now require Multi-Factor Attestation (e.g., via a mobile app or physical token) for any configuration block that defines a new executable hook or modifies system-level settings.
* **Granular Consent Revocation**: Users can now revoke consent for a specific hook hash globally, causing all agents across all projects to immediately suspend execution if they attempt to run that hook.
**Security Impact**: Eliminates "Trust Brushing" and prevents silent RCE from malicious configuration changes that might occur between agent sessions (e.g., after a `git pull`).

### Update: 2026-06-28 - Implementing Hardware-Locked Configuration Anchors (HLCA)
**Context**: The disclosure of CVE-2026-33068 proves that file-based trust can be bypassed by malicious repository settings.
**Architecture Adjustment**:
* **Hardware-Locked Anchoring**: Project-local settings are now cryptographically bound to a TPM-signed user session.
* **Mandatory HLCA Validation**: Sections 3 and 4 are updated to require HLCA validation for any configuration file matching known agent settings patterns (e.g., `.claude/settings.json`, `GEMINI.md`).
**Security Impact**: Ensures that even if a malicious configuration file is committed to a repository, it cannot silently bypass the workspace trust dialog without a hardware-bound user signature.
