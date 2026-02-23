# Design Doc: Secure Command Execution Gateway (SEG)
**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
Recent critical vulnerabilities (CVE-2026-0755, CVE-2026-0757) have exposed a systemic weakness in the MCP ecosystem: the lack of standard sanitization for tool calls that execute system commands. Agents frequently pass unsanitized LLM output into shell-executing functions like `execAsync`, leading to Remote Code Execution (RCE). MCP Any needs a centralized "Secure Execution Gateway" to intercept and harden these calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all tool calls that match known "execution-risk" patterns (e.g., arguments containing `;`, `&&`, `|`, or backticks).
    * Provide a configurable "Regex Allowlist" for permitted command structures.
    * Implement mandatory argument sanitization for all standard MCP `shell` tools.
    * Enable "Safe Mode" by default, which blocks all shell execution unless explicitly allowed by the Policy Firewall.
* **Non-Goals:**
    * Replacing the underlying shell/system (this is a middleware layer).
    * Validating the *intent* of the command (this is handled by the Policy Engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer.
* **Primary Goal:** Prevent an agent from performing a command injection attack if the LLM is compromised by a prompt injection.
* **The Happy Path (Tasks):**
    1. Engineer enables `secure_execution_gateway` in `config.yaml`.
    2. An agent attempts to call a tool `run_script(path="script.sh; rm -rf /")`.
    3. SEG intercepts the call, detects the `;` injection pattern.
    4. SEG blocks the call and returns a "Security Violation" error to the agent.
    5. The event is logged in the Swarm Accountability Ledger.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: SEG sits in the core middleware pipeline, before the tool execution adapter.
    - **Analysis**: It inspects the `arguments` of any tool call. If the tool is tagged as `type: shell` or matches a risk-list of known exec-tools, it triggers the validator.
    - **Validation**: Applies a multi-pass validator:
        - **Pass 1**: Meta-character check (Detects `|`, `;`, `&`, etc.).
        - **Pass 2**: Path validation (Ensures scripts are within allowed directories).
        - **Pass 3**: Regex Allowlist (Ensures the command structure matches expected patterns).
* **APIs / Interfaces:**
    - New configuration block:
    ```yaml
    seg:
      enabled: true
      deny_patterns: [";", "&&", "|", "`", "$("]
      allowed_paths: ["/usr/bin/git", "/app/scripts/"]
    ```
* **Data Storage/State:** Stateless validation, but logs state transitions to the audit log.

## 5. Alternatives Considered
* **Containerization Only**: Relying solely on Docker isolation. *Rejected* because container escapes are still possible, and RCE within a container can still leak secrets or perform lateral movement.
* **LLM-Based Sanitization**: Asking an LLM to sanitize the command. *Rejected* due to latency and the risk of the LLM itself being bypassable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SEG is a core component of the Zero Trust architecture.
* **Observability:** Metrics on "Blocked Commands" will be visible in the Security Dashboard.

## 7. Evolutionary Changelog
* **2026-02-26:** Initial Document Creation.
