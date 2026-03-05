# Design Doc: Universal Policy Translator
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
The AI agent ecosystem is suffering from "Policy Fragmentation." Gemini CLI uses one format, Claude Code another, and OpenClaw a third. This forces developers to duplicate security logic across multiple tools, leading to inconsistencies and "Security Shadows"—areas where policies are missing or misconfigured.

The Universal Policy Translator (UPT) allows users to define a single, declarative security policy (using industry standards like Rego/Open Policy Agent) which MCP Any then translates and injects into the native policy engines of connected agent frameworks.

## 2. Goals & Non-Goals
* **Goals:**
    * Single Source of Truth (SSoT) for agent security policies.
    * Bi-directional translation between MCP Any Policies and Native (Gemini/Claude) formats.
    * Real-time policy "Hot Reload" across all connected agents.
    * Validation of native policies against the master MCP Any policy.
* **Non-Goals:**
    * Replacing native engines. UPT *adapts* to them to ensure low-latency local enforcement.
    * Defining the "Perfect Policy." It provides the *mechanism*, not the *content*.

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Officer (CSO)
* **Primary Goal:** Enforce a "No local file writes to /etc" policy across 5 different agent frameworks simultaneously.
* **The Happy Path (Tasks):**
    1. CSO writes a Rego policy: `deny { input.tool == "write_file" and startswith(input.path, "/etc") }`.
    2. CSO uploads policy to MCP Any.
    3. UPT parses the Rego and identifies the equivalent "Seatbelt Profile" for Gemini CLI and the `file_system` filter for Claude Code.
    4. MCP Any pushes these configurations to the respective agents via their config files or API hooks.
    5. Agent attempts a prohibited write and is blocked natively.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        SSoT[Master Rego/CEL Policy] --> UPT[Universal Policy Translator]
        UPT --> GeminiAdapter[Gemini Policy Generator]
        UPT --> ClaudeAdapter[Claude Policy Generator]
        UPT --> OpenClawAdapter[OpenClaw Policy Generator]
        GeminiAdapter --> GeminiConfig[.geminirc / Policy DB]
        ClaudeAdapter --> ClaudeConfig[claude_config.json]
    ```
* **APIs / Interfaces:**
    * `POST /v1/policies/master`: Update the master policy.
    * `GET /v1/policies/translation/{target}`: Get the translated policy for a specific framework.
* **Data Storage/State:**
    * Master policies stored in Git (for versioning) and mirrored in MCP Any's internal SQLite.

## 5. Alternatives Considered
* **Runtime Proxying Only**: Intercepting every call at the MCP level. Rejected as too slow; native enforcement is preferred for performance and "belt-and-suspenders" security.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** UPT itself must be secured. Only authenticated admins can update the master policy.
* **Observability:** Logs indicate when a translation occurs and if the target agent successfully acknowledged the new policy.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
