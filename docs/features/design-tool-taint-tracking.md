# Design Doc: Tool Taint Tracking Engine
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
Recent "Zero-Click" exploits in agentic ecosystems (e.g., Claude Desktop, Claude Code) have demonstrated a critical security flaw: data from untrusted tools (like email or calendar) can be passed directly to high-trust tools (like a shell or filesystem) without a security boundary. MCP Any needs to implement a "Taint Tracking" mechanism to ensure that untrusted tool outputs are sanitized or explicitly approved before being used as inputs for sensitive operations.

## 2. Goals & Non-Goals
* **Goals:**
    * Assign trust levels to all registered tools and MCP servers.
    * Metadata-tag all tool outputs with a "Taint Level" based on the source tool's trust level.
    * Prevent high-trust tools (Sink Tools) from accepting inputs tagged with high taint levels without a "De-Taint" policy check.
    * Provide an audit log of tainted data flows for observability.
* **Non-Goals:**
    * Automatically "cleaning" or "sanitizing" the content of the data (this is the role of the LLM or a specific sanitizer tool).
    * Modifying the underlying MCP protocol.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Developer.
* **Primary Goal:** Prevent an agent from executing a malicious shell command that was extracted from an untrusted email.
* **The Happy Path (Tasks):**
    1. An agent calls `read_email` (Trust Level: LOW, Taint: HIGH).
    2. The email content is returned with a `taint: high` metadata tag.
    3. The agent tries to call `execute_command(command=extracted_text)` (Sink Tool: HIGH TRUST).
    4. The Taint Tracking Engine intercepts the call and sees the `taint: high` input.
    5. The engine blocks the call and triggers a `HITL Approval` request or a "De-Taint" policy check.
    6. The user approves or the policy allows it if certain conditions are met, otherwise the execution is aborted.

## 4. Design & Architecture
* **System Flow:**
    - **Tagging**: The `TaintMiddleware` intercepts every tool output and wraps it in a metadata object: `{ data: "...", metadata: { taint: "high", source: "email_server" } }`.
    - **Tracking**: MCP Any tracks these tags within the `Recursive Context Protocol` headers.
    - **Enforcement**: When a tool call is made, the `Policy Firewall` checks the taint levels of the input arguments against the tool's "Trust Threshold."
* **APIs / Interfaces:**
    - **Trust Configuration**: New configuration block for MCP servers: `trust_level: high | medium | low`.
    - **Sink Policy**: New policy rule: `require_clean_input: true` for sensitive tools.
* **Data Storage/State:** Taint state is stored in the current agent session's memory (Context).

## 5. Alternatives Considered
* **Content Scanning**: Scanning all tool inputs/outputs for malicious patterns (e.g., regex, LLM-based inspection). *Rejected* because it's slow, expensive, and bypassable. Taint tracking is a structural defense.
* **User-Only Execution**: Requiring HITL for all shell calls. *Rejected* as it degrades the "Autonomous" value proposition.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero Trust feature. It prevents "Privilege Escalation via Tainted Input."
* **Observability:** Taint flows will be visualized in the "Agent Chain Tracer" UI.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
