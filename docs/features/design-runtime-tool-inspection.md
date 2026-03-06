# Design Doc: Runtime Tool Inspection (MCP-DPI)

**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
Existing security for AI agents mostly focuses on "Tool-Level" access (Allow/Deny `execute_command`). However, a single tool can have multiple, high-risk arguments (e.g., `execute_command(cmd="rm -rf /")`). Runtime Tool Inspection (MCP-DPI) introduces Deep Packet Inspection for MCP tool calls, allowing the system to inspect, validate, and redact the *arguments* of a tool call before it reaches the upstream service.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement an inline policy engine for Deep Packet Inspection (DPI) of MCP calls.
    *   Enable regex-based and schema-aware payload filtering for tool arguments.
    *   Implement real-time blocking or redaction of sensitive arguments (PII, dangerous flags).
    *   Link DPI violations to specific Agent Identities (from AIM).
*   **Non-Goals:**
    *   Rewriting the model's tool calls (it should either block, redact, or fail).
    *   Inspecting binary tool outputs (initial focus is on tool *inputs*).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Compliance Officer.
*   **Primary Goal:** Prevent an agent from including a user's Social Security Number in a tool call to a public web search tool.
*   **The Happy Path (Tasks):**
    1.  User configures a DPI policy: `block tool:web_search if args.query matches regex(SSN_PATTERN)`.
    2.  An agent calls `web_search(query="Search for credit report of SSN: 123-456...")`.
    3.  MCP-DPI interceptor detects the SSN pattern.
    4.  MCP-DPI blocks the call and returns a "Security Policy Violation" error.
    5.  The incident is logged: `[2026-03-06] Agent: research-01 (SID: ag_...) BLOCKED from web_search due to SSN_PATTERN violation`.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: A new middleware layer between the protocol gateway and the tool executor.
    - **Policy Matching**: The Policy Engine evaluates the tool name and arguments against configured "Guardrail Rules."
    - **Enforcement**: Rules return `ALLOW`, `BLOCK`, or `REDACT(field, replacement)`.
*   **APIs / Interfaces:**
    - `POST /api/v1/policies/guardrails`: Register a new DPI rule.
    - Extension of Rego/CEL input to include `args` (the JSON payload of the tool call).
*   **Data Storage/State:** Rule definitions stored in `config.yaml` or the internal policy database.

## 5. Alternatives Considered
*   **Post-Execution Auditing Only**: Only logging what the agent did. *Rejected* because "Prevention is better than Cure" for destructive actions.
*   **Client-Side Filtering**: Asking the IDE or agent framework to filter. *Rejected* as it relies on the client being secure and compliant.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This feature provides "Least Privilege" at the argument level. It prevents prompt injection and accidental data exfiltration.
*   **Observability:** DPI logs will be a first-class citizen in the Security Dashboard, highlighting blocked attempts and sensitive data leaks.

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
