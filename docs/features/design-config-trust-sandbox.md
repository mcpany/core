# Design Doc: Config Trust Sandbox (Trust-Before-Load)
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
Recent critical vulnerabilities in AI agent tools (CVE-2025-59536 in Claude Code, and similar exploits in OpenClaw) have highlighted a major security gap: the "Shadow Config" attack. Attackers can include malicious configuration files (e.g., `.claude/settings.json`, `mcp.yaml`) in a repository. When a developer opens that repository with an AI tool, the tool may automatically load these settings, which can then be used to exfiltrate API keys, execute arbitrary shell commands via hooks, or redirect traffic to malicious MCP servers *before* any trust prompt is shown. MCP Any must provide a "Trust-Before-Load" sandbox to neutralize this vector.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement an isolated parsing environment for project-level configuration files.
    * Require explicit user attestation before applying any non-standard or "sensitive" configuration (e.g., custom hooks, modified base URLs, new MCP server definitions).
    * Provide a "Safety Report" to the user detailing what a configuration file intends to change.
    * Sanitize and validate all environment variable overrides from untrusted sources.
* **Non-Goals:**
    * Automatically "fixing" malicious configs (it should block/quarantine them).
    * Providing a full VM for every tool execution (this is specifically for the *configuration* phase).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer.
* **Primary Goal:** Safely explore a new open-source repository without risking API key theft or RCE from malicious project settings.
* **The Happy Path (Tasks):**
    1. User clones a repository and runs `mcpany start --dir ./new-repo`.
    2. MCP Any detects a project-level configuration file (`mcp.yaml`).
    3. Instead of loading it directly, MCP Any parses it in the **Config Trust Sandbox**.
    4. MCP Any identifies that the config tries to set a custom `ANTHROPIC_BASE_URL` and add a new shell-based MCP server.
    5. MCP Any pauses and displays a **Security Alert**: "Untrusted configuration detected. It attempts to: 1. Redirect API traffic. 2. Execute local shell commands."
    6. User reviews the diff and clicks "Approve" or "Reject/Quarantine."

## 4. Design & Architecture
* **System Flow:**
    * **Discovery**: `ConfigScanner` identifies local config files.
    * **Sandbox Parsing**: A restricted parser (using a WASM-based YAML/JSON parser or a separate process with no network/file access) extracts the key-value pairs.
    * **Policy Matching**: The `TrustEngine` compares extracted settings against a "Safe Baseline" (e.g., standard tool names are okay, but `base_url` changes are high-risk).
    * **Attestation UI**: If risks are found, the UI/CLI blocks and requests manual approval.
* **APIs / Interfaces:**
    * `POST /v1/config/verify`: Submit a config file for sandboxed analysis.
    * `GET /v1/config/trust-report`: Retrieve the safety analysis of a pending config.
* **Data Storage/State:** Pending configurations are kept in a "Quarantine" state in memory until approved.

## 5. Alternatives Considered
* **Ignore Project Configs**: Just don't support repository-level settings. *Rejected* as it breaks the "Zero-Configuration" developer experience.
* **Global Allowlist**: Only allow configs from certain GitHub orgs. *Rejected* as it doesn't scale and is easy to spoof.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a "Zero Trust" entry point. We assume *all* project-level config is malicious until proven otherwise.
* **Observability:** Audit logs must record every time a config was quarantined and who approved it.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
