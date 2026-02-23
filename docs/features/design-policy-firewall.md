# Design Doc: Policy Firewall Engine with Deep Input Inspection

**Status:** Draft
**Created:** 2026-02-26

## 1. Context and Scope
As AI agents gain more autonomy, the risk of "Agent Hijacking" or "Prompt Injection-to-RCE" increases. A recent 0-day in `gemini-mcp-tool` (CVE-2026-0755) demonstrated that even if a tool call is authorized, malicious arguments can lead to command injection. MCP Any requires a robust, Zero Trust "Policy Firewall" that not only authorizes *who* can call a tool but also *what* data can be passed to it.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement Rego/CEL based policy enforcement for all tool calls.
    *   **Deep Input Inspection (DII)**: Sanitize and validate tool arguments against security patterns (shell injection, path traversal).
    *   **Hardware-Attested Identity**: Integrate TPM/Secure Enclave signatures into the authorization flow.
    *   Support "Intent-Aware" scoping where tools are allowed only if they align with verified parent intent.
*   **Non-Goals:**
    *   Rewriting the security logic of upstream MCP servers.
    *   Static analysis of upstream tool binaries.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Enterprise Admin.
*   **Primary Goal:** Prevent an agent from executing unauthorized shell commands or accessing files outside a specific directory, even if the agent is compromised.
*   **The Happy Path (Tasks):**
    1.  Admin configures a policy: "Only allow `shell_execute` if arguments do not contain `;`, `&&`, or `|`."
    2.  Agent receives a malicious prompt and tries to call `shell_execute(command="ls ; rm -rf /")`.
    3.  MCP Any's Policy Firewall intercepts the call.
    4.  DII engine detects the `;` metacharacter.
    5.  Call is blocked, and a security alert is logged.
    6.  The agent receives a "Permission Denied: Malicious Input Detected" error.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: All `tools/call` requests pass through the Policy Middleware.
    - **Identity Verification**: If required, the middleware verifies the `X-MCP-Agent-Attestation` header against a trusted TPM/Secure Enclave root.
    - **Deep Input Inspection**: Arguments are matched against a library of "Dangerous Patterns" (RegEx or CEL expressions).
    - **Policy Evaluation**: The request context (Agent ID, Tool Name, Arguments, Attestation Status) is passed to the OPA (Open Policy Agent) engine for a final Allow/Deny decision.
*   **APIs / Interfaces:**
    - New Header: `X-MCP-Agent-Attestation` (Cryptographic signature of the agent's environment).
    - Configuration: `policies/` directory containing `.rego` files.
*   **Data Storage/State:** Policy definitions stored in YAML/Rego; Audit logs persisted to SQLite.

## 5. Alternatives Considered
*   **Hardcoded Sanitization**: Rejected because it's not flexible enough for diverse tool sets.
*   **LLM-based Validation**: Rejected as "Security-by-LLM" is prone to the same injection attacks it's meant to prevent.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Policy Firewall is the "Last Line of Defense." It assumes both the agent and the upstream tool could be compromised.
*   **Observability:** All blocked calls generate "Security Traces" in the dashboard with detailed reasons (e.g., "DII: Shell Metacharacter Detected").

## 7. Evolutionary Changelog
*   **2026-02-26:** Initial Document Creation. Integrated Deep Input Inspection (DII) and Hardware Attestation in response to CVE-2026-0755 and OpenClaw v2026.2.23.
