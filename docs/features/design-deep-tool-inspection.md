# Design Doc: Deep Tool Inspection (DTI) Middleware

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Traditional MCP security focuses on "Who can call which tool" (Authentication/Authorization). However, the "Agent-as-a-Proxy" (AaaP) attack pattern reveals a deeper vulnerability: a legitimate tool (e.g., `fetch_url`) can be used by a compromised subagent to perform internal network scanning or data exfiltration. Deep Tool Inspection (DTI) moves beyond schema validation to inspect the *behavior* and *side-effects* of tool calls in real-time.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept and inspect egress network calls initiated by MCP tools.
    *   Implement "Domain-Specific Egress Policies" (e.g., the `jira_tool` can only talk to `*.atlassian.net`).
    *   Provide "Behavioral Attestation": Verify that a tool's actual execution matches its declared intent.
    *   Support "Payload Inspection" to prevent sensitive data leakage (PII/Secrets) in tool arguments or results.
*   **Non-Goals:**
    *   Replacing host-level firewalls (DTI operates at the application/middleware layer).
    *   Deep packet inspection of encrypted traffic (requires mitmproxy-like setup, which is out of scope for the initial version).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin.
*   **Primary Goal:** Prevent a coding agent from using a `curl` tool to access the internal metadata service (`169.254.169.254`).
*   **The Happy Path (Tasks):**
    1.  Admin defines a DTI policy: `tools.network_access: [trusted_domains]`.
    2.  An agent attempts to call `http_request(url="http://169.254.169.254/latest/meta-data/")`.
    3.  DTI middleware intercepts the call, identifies the target as a "prohibited internal range."
    4.  DTI blocks the tool execution and raises a "Security Policy Violation" alert.
    5.  The event is logged in the Audit Trail with the full context of the agent session.

## 4. Design & Architecture
*   **System Flow:**
    - **Hook Injection**: MCP Any wraps tool execution in a `DTIContext`.
    - **Egress Proxy**: For tools that perform network IO, MCP Any provides a "Secure Fetch" wrapper that enforces the egress policy.
    - **Content Scanning**: Before returning tool results to the LLM, the DTI middleware scans the payload for patterns matching sensitive data (Regex-based DLP).
*   **APIs / Interfaces:**
    - Policy Schema: `dti_rules: { tool_pattern: string, allow_egress: string[], redact: string[] }`.
*   **Data Storage/State:** Policies are stored in the main `config.yaml` or a dedicated `policies.rego` file.

## 5. Alternatives Considered
*   **Container Isolation**: Running every tool in a sandbox. *Rejected* as too heavy for simple tools and doesn't solve the "data exfiltration via trusted domains" problem.
*   **Manual Code Review of Tools**: *Rejected* as it doesn't scale and can't prevent runtime misuse of legitimate tools.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** DTI is the implementation of "Least Privilege" for tool behavior.
*   **Observability:** The UI must show "Blocked Egress" attempts in the Tool Execution Timeline.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
