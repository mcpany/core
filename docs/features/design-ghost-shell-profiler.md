# Design Doc: Ghost Shell Hook Profiler
**Status:** Draft
**Created:** 2026-03-24

## 1. Context and Scope
With the discovery of "Binary Smuggling" in project-local configuration hooks (CVE-2025-59536), simple static analysis is no longer sufficient. Agents are often tricked into executing malicious WASM or shell commands hidden in `.claude/settings.json`. Ghost Shell provides a natively managed, instrumented sandbox that profiles the behavior of un-attested hooks before they are granted access to the host environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Execute un-attested configuration hooks in an air-gapped, instrumented container.
    * Profile hooks for suspicious activities (network requests, unauthorized file access, credential exfiltration).
    * Generate a "Safety Report" and suggested Content-Addressable Config (CAC) attestation policy.
* **Non-Goals:**
    * Provide a permanent execution environment for all tools (this is for pre-flight profiling).
    * Automatically patch malicious hooks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer cloning an untrusted repository.
* **Primary Goal:** Safely evaluate and approve automated hooks defined in the repository's configuration without risking machine takeover.
* **The Happy Path (Tasks):**
    1. User opens a new project repository containing a `.claude/settings.json` with automated hooks.
    2. MCP Any detects un-attested hooks and blocks their execution.
    3. User triggers "Ghost Shell Profiling" via the UI or CLI.
    4. Ghost Shell executes the hooks in a detached, monitored sandbox.
    5. Ghost Shell generates a report showing the hook attempted to access `/etc/passwd` and connect to a remote IP.
    6. User sees the report and permanently blocks the malicious hook.

## 4. Design & Architecture
* **System Flow:**
    `[Config Loader] -> [Un-attested Hook] -> [Ghost Shell Sandbox] -> [Behavioral Telemetry] -> [Safety Report]`
* **APIs / Interfaces:**
    * `profileHook(hookDefinition)` internal API.
    * `GET /v1/reports/ghost-shell/{id}` for retrieving profiling results.
* **Data Storage/State:**
    * Temporary container state (discarded after profiling).
    * Persistent behavioral logs for attestation history.

## 5. Alternatives Considered
* **Static Analysis Only:** Rejected because it cannot detect runtime "Binary Smuggling" or obfuscated shell commands.
* **Manual Code Review:** Rejected as it's error-prone and doesn't scale with complex, multi-layered configurations.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Ghost Shell sandbox must have zero network access by default and strictly limited syscalls.
* **Observability:** Detailed logging of all sandboxed actions is critical for the "Safety Report."

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.

### Update: 2026-03-24 - Ghost Shell as Mandatory Profiling
**Context:** Today's market sync revealed the rise of "Binary Smuggling" in WASM hooks.
**Architecture Adjustment:**
* Ghost Shell is now a mandatory pre-flight step for any un-attested configuration hooks.
* Introducing behavioral "Burn-In" periods for newly discovered skills.
**Security Impact:** Provides a behavioral safety net before any malicious WASM or shell code can reach the host.
