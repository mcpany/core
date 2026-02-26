# Design Doc: Secure Config Sandbox
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent critical vulnerabilities in agentic tools (e.g., CVE-2025-59536 in Claude Code) have demonstrated that repository-bound configuration files can be weaponized to execute arbitrary commands on a developer's machine. As MCP Any acts as a gateway for multiple MCP servers, it must ensure that tool-specific configuration files (like `.mcp.json` or custom tool configs) are parsed and loaded within a secure, sandboxed perimeter that prevents shell injection and unauthorized file access.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Strict Schema Validation" layer for all tool configuration files.
    * Enforce a "No-Exec" policy during configuration parsing (preventing backticks, `$()`, or eval-like behavior).
    * Provide a virtualized view of the filesystem to tool loaders to prevent directory traversal.
    * Create a "Verification Registry" for known-safe configuration templates.
* **Non-Goals:**
    * Replacing the underlying MCP server's own configuration logic (we wrap/gate it).
    * Providing a full OS-level sandbox (e.g., Firecracker) - this is focused on the *config loading* phase.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer.
* **Primary Goal:** Open a community-contributed repository and use its MCP tools without risk of hidden "auto-run" exploits in the config.
* **The Happy Path (Tasks):**
    1. User points MCP Any to a new project directory.
    2. MCP Any detects a `.mcp.json` configuration file.
    3. The Secure Config Sandbox interceptor reads the file as raw text.
    4. The interceptor validates the file against a strict JSON schema and scans for suspicious command patterns.
    5. If valid, it provides a sanitized, read-only object to the Tool Loader.
    6. If suspicious, it alerts the user and blocks the tool from loading.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Project Directory] -->|Read Config| B[Secure Config Sandbox]
        B -->|Scan for Shell Injection| C{Is Safe?}
        C -->|No| D[Alert User & Block]
        C -->|Yes| E[Schema Validation]
        E -->|Success| F[Sanitized Config Object]
        F -->|Inject| G[MCP Tool Loader]
    ```
* **APIs / Interfaces:**
    * `ConfigValidator` interface: `Validate(rawConfig []byte, schema string) (SanitizedConfig, error)`
    * `SandboxedFs`: A restricted `io/fs` implementation that limits tool loaders to specific paths.
* **Data Storage/State:**
    * Policy rules stored in memory, updated via the Global Policy Engine.

## 5. Alternatives Considered
* **User Confirmation for Every Config:** Rejected as it causes "Confirmation Fatigue" and users often click through security warnings.
* **Full Containerization:** Rejected as too heavyweight for local development workflows; configuration-level sandboxing is more targeted.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox assumes all configuration files are potentially malicious until proven otherwise.
* **Observability:** Failed validation attempts are logged to the Security Audit Log with the specific offending patterns highlighted.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation. Addressing CVE-2025-59536 and establishing infrastructure-level config safety.
