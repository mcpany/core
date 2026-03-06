# Design Doc: Shadow-Guard Configuration Sandbox
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
With the rise of project-local MCP configurations (e.g., `.mcp.json`, `.claude/settings.json`), a new class of supply chain attacks has emerged. Attackers with commit access can inject malicious shell commands as "hooks" or override API base URLs to steal credentials. MCP Any must provide a secure buffer that validates these configurations before they can influence the execution environment or tool registry.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all project-local configuration loading attempts.
    * Validate configuration schema and content against a "Safe List" of allowed commands and URLs.
    * Provide a Human-in-the-Loop (HITL) approval flow for any configuration changes.
    * Isolate the parsing of untrusted JSON/YAML in a restricted environment.
* **Non-Goals:**
    * Replacing the underlying agent frameworks' configuration systems entirely.
    * Providing a full sandbox for the *tools* themselves (this is handled by other layers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / Agent Orchestrator
* **Primary Goal:** Open a new repository and allow an agent to use local tools without risking RCE from malicious project configs.
* **The Happy Path (Tasks):**
    1. User points MCP Any to a new repository.
    2. Shadow-Guard detects `.mcp.json` and `.claude/settings.json`.
    3. Shadow-Guard parses the files in an isolated process.
    4. Shadow-Guard identifies a new "hook" command that is not on the global safe-list.
    5. MCP Any UI/CLI prompts the user: "Untrusted hook detected: `rm -rf /`. Allow?"
    6. User denies the hook; Shadow-Guard loads the config *without* the malicious hook.
    7. Agent proceeds with only verified tools.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Repo Root] --> B[Shadow-Guard Watcher]
        B --> C{Detect Config Files}
        C -->|Yes| D[Isolated Parser]
        D --> E[Policy Evaluator]
        E -->|Violation| F[HITL Approval Flow]
        E -->|Safe| G[Validated Config Cache]
        F -->|Approved| G
        F -->|Denied| H[Sanitized Config]
        H --> G
        G --> I[MCP Any Tool Registry]
    ```
* **APIs / Interfaces:**
    * `POST /v1/sandbox/validate`: Accepts raw config content, returns identified risks.
    * `GET /v1/sandbox/policies`: Returns current safe-list and auto-deny patterns.
* **Data Storage/State:**
    * `config_approvals.db`: SQLite table tracking hashes of approved/denied configurations to prevent re-prompting.

## 5. Alternatives Considered
* **Disabling Local Configs Entirely**: Rejected because it breaks the developer experience and compatibility with tools like Claude Code.
* **OS-Level Sandboxing (e.g., gVisor)**: Considered for the tools themselves, but overkill for just parsing the configuration files.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The Isolated Parser must run with minimal privileges (no network, no write access to host).
* **Observability**: All blocked config attempts must be logged to the security audit trail.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
