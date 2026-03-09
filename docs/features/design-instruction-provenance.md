# Design Doc: Instruction Provenance Middleware

**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
As demonstrated by recent red-teaming of OpenClaw and other autonomous agent frameworks, agents are highly susceptible to "Indirect Prompt Injection." An agent reading an untrusted email or web page can be "hijacked" by malicious instructions embedded in that data. Current MCP implementations do not distinguish between instructions originating from the primary user and those ingested from untrusted external sources. The Instruction Provenance Middleware aims to solve this by tagging every instruction with its origin and enforcing tool-execution policies based on that provenance.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a tagging system for all text input processed by agents via MCP Any.
    *   Maintain a "Provenance Chain" from the initial input source to the final tool call.
    *   Enforce "Provenance-Aware" policies (e.g., "deny shell execution if any part of the instruction came from an untrusted web scrape").
    *   Provide a standard API for agents to report the provenance of their current task.
*   **Non-Goals:**
    *   Automatically detecting prompt injection (this is a reactive measure; we focus on isolation and enforcement).
    *   Modifying LLM core logic. We act as a gateway/proxy.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Agent Developer.
*   **Primary Goal:** Prevent an autonomous "Email Triage Agent" from being tricked into deleting the user's home directory.
*   **The Happy Path (Tasks):**
    1.  The agent uses an MCP tool to fetch an email. The email tool tags the output with `provenance: "untrusted_email"`.
    2.  The agent's next instruction (derived from the email) is processed by MCP Any. MCP Any sees the `untrusted_email` tag in the task context.
    3.  The agent attempts to call a `shell_execute` tool.
    4.  The Instruction Provenance Middleware checks the policy: `shell_execute` requires `provenance: "trusted_user"`.
    5.  The middleware blocks the call and returns a "Security Policy Violation: Untrusted Provenance" error to the agent.

## 4. Design & Architecture
*   **System Flow:**
    - **Provenance Tagging**: Adapters (Email, Web, Slack) are updated to include provenance metadata in their tool responses.
    - **Context Propagation**: MCP Any's `Recursive Context Protocol` is extended to carry `provenance_tags` in the session state.
    - **Policy Enforcement**: The `Policy Firewall` is updated to include `provenance` as a first-class attribute in Rego/CEL rules.
*   **APIs / Interfaces:**
    - Protocol extension: `JSON-RPC` messages will include a `meta.provenance` field.
    - New Header: `X-MCP-Provenance: source=web; trust_level=low`
*   **Data Storage/State:** Provenance tags are stored in the active session context (Shared KV Store or in-memory session manager).

## 5. Alternatives Considered
*   **LLM-based Detection**: Asking the LLM if an instruction is safe. *Rejected* because LLMs are the component being bypassed; they cannot be the ultimate arbiter of their own security boundaries.
*   **Static Sandboxing**: Running all untrusted tools in a container. *Complimentary* but doesn't prevent "logical" attacks (e.g., sending unauthorized emails or API calls).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust feature. It moves the trust boundary from the identity of the agent to the provenance of the instruction.
*   **Observability:** The UI must visualize the "Provenance Path" of a tool call in the audit logs.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
