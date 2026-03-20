# Design Doc: Config Smuggling Scanner
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
With the maturation of Claude Code's "Settings Injection Guard" and OpenClaw's "Project Configuration Guard," the industry has identified a new class of vulnerability: "Config Smuggling." This involves hiding malicious instructions or "Binary Smuggling" (CVE-2026-31042) inside project-local configuration files (e.g., `.claude/settings.json`, `.gemini/skills/config.yaml`) within complex binary or metadata blobs that bypass traditional regex-based scanners. MCP Any needs a specialized "Semantic Scanner" that performs deep inspection of project-local configurations to detect and block these "Dormant Logic Bombs" before they are ingested by the agent reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Deep Semantic Inspection" for all project-local agent configurations.
    * Detect "Instruction Smuggling" hidden in non-primary configuration fields (metadata, tags, base URLs).
    * Block "Binary Smuggling" exfiltration patterns in WASM-based configuration hooks.
    * Provide a "Verified View" of configurations to the agent, with any "smuggled" fragments redacted.
* **Non-Goals:**
    * Replacing the "Structural Metadata Sanitizer" (remains responsible for schema validation).
    * Preventing *all* project-local file creation (only those with high-risk mismatches).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Agent in a cloned repository.
* **Primary Goal:** Safely load the project-local settings without executing a hidden malicious hook.
* **The Happy Path (Tasks):**
    1. The agent attempts to load `.claude/settings.json`.
    2. The Config Smuggling Scanner intercepts the request.
    3. It performs a semantic scan of the "custom_rules" and "metadata" fields.
    4. It identifies a hidden `exec` instruction in a base64-encoded string.
    5. The scanner redacts the malicious fragment and returns a "Sanitized Config" to the agent.
    6. A "Security Alert" is triggered for the user.

## 4. Design & Architecture
* **System Flow:**
    `Local File` -> `File Watcher` -> `Semantic Scanner` -> `Redactor` -> `Agent`
* **APIs / Interfaces:**
    * `ConfigScanner.scan(filePath string) (*SanitizedConfig, error)`
    * `ConfigScanner.redact(filePath string)`
* **Data Storage/State:**
    * Signature patterns are stored in the "Federated Policy Synchronizer."

## 5. Alternatives Considered
* **Regex-based Scanning**: Rejected as insufficient against the "Normalization Fatigue" and "Logic Grafting" observed in recent Claude Code CVEs.
* **Full Sandbox Execution for all configs**: Rejected due to high resource overhead and difficulty in profiling non-executable metadata.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The scanner must be hardware-bound to prevent an agent from "shadowing" its own scanning logic.
* **Observability:** Sanitized fragments and "Smuggling" detections are logged as high-priority security alerts.

## 7. Evolutionary Changelog
* **2026-03-20:** Initial Document Creation.
