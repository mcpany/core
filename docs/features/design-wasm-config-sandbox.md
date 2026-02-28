# Design Doc: WASM-Based Config Sandbox
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
Recent security vulnerabilities in agentic CLI tools (e.g., Claude Code CVE-2025-59536) have demonstrated that trusting local project-level configuration files is dangerous. Attackers can inject malicious hooks or unauthorized MCP server configurations that lead to Remote Code Execution (RCE) or credential exfiltration. MCP Any needs a way to ingest these configurations safely without exposing the host system.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate the parsing and validation of .claudecode, .mcpany, and other configuration files within a restricted WASM environment.
    * Prevent unauthorized shell command execution during configuration ingestion.
    * Provide a clear "Safe/Unsafe" report to the user before applying any configuration.
* **Non-Goals:**
    * Replacing the entire configuration system of third-party tools.
    * Sandboxing the actual execution of tools (this is handled by other layers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Open a community-contributed repository and use its MCP tools without risking host compromise.
* **The Happy Path (Tasks):**
    1. User clones a repository containing a .claudecode/config.json file.
    2. User runs MCP Any in the directory.
    3. MCP Any detects the config file and spins up the WASM Sandbox.
    4. The Sandbox parses the config and identifies a "pre-tool" hook attempting to run `curl | bash`.
    5. MCP Any flags the hook as "Unsigned & Dangerous" and prompts the user for explicit approval or signing.
    6. User rejects the dangerous hook but allows the tool definitions.

## 4. Design & Architecture
* **System Flow:**
    [Config File] -> [WASM-Host Bridge] -> [WASM Validator (Rust/C++)] -> [Validated JSON] -> [MCP Any Core]
* **APIs / Interfaces:**
    * `validate_config(raw_bytes: Vec<u8>) -> ValidationResult`
* **Data Storage/State:**
    * Stateless validation. Results are stored in the MCP Any session state after user approval.

## 5. Alternatives Considered
* **Static Analysis (Regex/AST):** Rejected because it's too easy to bypass with obfuscation.
* **Docker Containers:** Rejected due to high overhead and slow startup times for simple config validation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The WASM environment will have zero access to the filesystem, network, or environment variables unless explicitly shared via the host bridge.
* **Observability:** Log all blocked/flagged configuration attempts for audit trails.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
