# Design Doc: Environment Metadata Sanitizer
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
The OpenClaw security crisis highlighted a new attack vector: indirect prompt injection via environment metadata (CVE-2026-27001). Agents often ingest directory names, git branch names, or file metadata as context. Attackers can craft malicious metadata (e.g., a directory named `"; DROP TABLE sessions; --"`) that, when included in a system prompt, redirects agent behavior. MCP Any needs to solve this by providing a sanitization layer for all environmental context.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and sanitize all "environmental metadata" before it reaches the LLM.
    * Provide a configurable "deny-list" of control characters and command-like patterns.
    * Standardize how file-system tools report metadata to prevent raw injection.
* **Non-Goals:**
    * Sanitizing the actual content of files (this is handled by the Policy Firewall).
    * Replacing OS-level filesystem security.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Prevent an agent from being subverted by a malicious repository it just cloned.
* **The Happy Path (Tasks):**
    1. Developer enables `EnvironmentMetadataSanitizer` in MCP Any config.
    2. Agent calls `list_directory` on a repo containing a folder named `[system_prompt_override]`.
    3. MCP Any middleware intercepts the tool response.
    4. Malicious metadata is replaced with a sanitized placeholder (e.g., `folder_001` or `sanitized_name`).
    5. Agent receives safe context and continues normal operation.

## 4. Design & Architecture
* **System Flow:**
    `Tool Execution -> Metadata Sanitizer Middleware -> LLM Context Assembly`
* **APIs / Interfaces:**
    * `Middleware: OnToolResponse`: Hook that scans for keys tagged as `metadata` or `filepath`.
    * `Config: sanitization_rules`: Regex-based rules for identifying and replacing malicious patterns.
* **Data Storage/State:**
    * Stateless; rules are loaded from configuration.

## 5. Alternatives Considered
* **Agent-Side Sanitization**: Rejected because it's prone to developer error and inconsistent across agent implementations.
* **OS-Level Restrictions**: Ineffective against prompt injection, which is a logic-level attack, not a permission-level attack.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Sanitization is part of the "Zero-Trust Context" pillar.
* **Observability**: Sanitization events will be logged as `SECURITY_SANITIZATION` events for audit.

## 7. Evolutionary Changelog
* **2026-03-01**: Initial Document Creation.
