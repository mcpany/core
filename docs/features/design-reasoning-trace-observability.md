# Design Doc: Reasoning Trace Observability & Redaction

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
As AI agents become more sophisticated, they increasingly rely on internal reasoning traces (e.g., Chain-of-Thought) to solve complex tasks. Recent research (Wu et al., 2026) has demonstrated that LLMs can covertly encode sensitive data (API keys, PII) within these traces—a vector known as "steganographic exfiltration." Current security measures only inspect final tool outputs, leaving a massive gap. MCP Any must provide visibility into and security scrubbing of these internal reasoning traces.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept and store internal agent reasoning traces during tool execution.
    *   Implement "Steganographic Redaction" to identify and mask sensitive data patterns in traces.
    *   Provide a secure API for authorized auditors to inspect traces.
    *   Integrate trace observability with the existing A2A Bridge to maintain a full "Reasoning Chain."
*   **Non-Goals:**
    *   Modifying the LLM's reasoning process (we only observe and redact).
    *   Real-time blocking of reasoning (we focus on secure logging and redaction for now).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Auditor.
*   **Primary Goal:** Verify that a multi-agent swarm did not exfiltrate customer data within its internal logs.
*   **The Happy Path (Tasks):**
    1.  The Auditor opens the MCP Any "Trace Explorer."
    2.  They select a specific agent session involving 3 subagents.
    3.  The UI displays a unified timeline of tool calls AND the internal reasoning traces that led to them.
    4.  Sensitive data (detected via regex or LLM-based scrubbing) is shown as `[REDACTED]`.
    5.  The Auditor confirms that the reasoning path was valid and secure.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception Middleware**: A new layer in the `mcpserver` pipeline that captures `_mcp_reasoning` metadata from incoming requests.
    - **Redaction Engine**: A pluggable service that runs patterns (Regex, Presidio, or scoped LLM) against the captured trace.
    - **Trace Store**: A dedicated table in the existing SQLite Blackboard for long-term trace storage, bound to the session ID.
*   **APIs / Interfaces:**
    - `GET /api/v1/sessions/:id/traces`: Retrieve redacted traces for a session.
    - `POST /api/v1/admin/redaction/rules`: Update redaction patterns.
*   **Data Storage/State:**
    - `traces` table: `session_id`, `agent_id`, `timestamp`, `raw_trace` (encrypted), `redacted_trace`.

## 5. Alternatives Considered
*   **Client-Side Redaction**: Relying on the agent framework to redact its own logs. *Rejected* as it violates the principle of "Trust but Verify."
*   **No Persistence**: Only redacting in-flight and not storing logs. *Rejected* as auditability is a core requirement for enterprise governance.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The `raw_trace` must be encrypted at rest and only accessible via highly privileged admin tokens. The `redacted_trace` is available to auditors.
*   **Performance:** Redaction must be asynchronous to prevent adding latency to the agent's execution path.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
