# Design Doc: Off-Model Policy Gate (OMPG)
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
Current AI agent frameworks (e.g., Claude Code, OpenClaw) often rely on the LLM itself to enforce negative constraints (e.g., "Never use `curl`"). However, research shows that under high context pressure or complex reasoning tasks, these "Deny" rules are often ignored or bypassed due to token prioritization or model hallucinations. This "Token Pressure" bypass creates a critical security hole where enterprise policies are silently discarded.

MCP Any needs to solve this by moving policy enforcement out of the reasoning loop and into the infrastructure layer. The **Off-Model Policy Gate (OMPG)** will intercept all tool requests at the gateway level (written in Go), applying deterministic pattern matching and regex validation before any tool execution occurs, regardless of the LLM's state or context density.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide 100% deterministic enforcement of "Deny" rules.
    * Eliminate token cost associated with security boundary checks for the LLM.
    * Support regex and exact-match blacklisting for tool names and arguments.
    * Interface with existing Rego/CEL policy engines for complex gating logic.
* **Non-Goals:**
    * Replacing the LLM's ability to reason about tool usage for "Happy Path" tasks.
    * Enforcing subjective or semantic-only constraints that require reasoning (e.g., "Don't be rude").

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent agents from using data exfiltration tools (like `curl` or `wget`) even if the agent is compromised or hallucinates under high context load.
* **The Happy Path (Tasks):**
    1. The Architect defines a `deny` policy in the MCP Any configuration: `{ "deny": ["curl", "wget", "ssh"] }`.
    2. An Agent receives a malicious prompt or enters a complex reasoning state where it decides it needs to use `curl`.
    3. The Agent sends a `call_tool` request for `curl` to MCP Any.
    4. **OMPG** intercepts the request at the Go middleware layer.
    5. OMPG matches `curl` against the blacklist and immediately returns a hard error to the Agent: `Security Policy Violation: Tool 'curl' is forbidden.`
    6. The tool never executes; no host-level network call is made.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Framework] -->|Call Tool| B(MCP Any Gateway)
        subgraph OMPG [Off-Model Policy Gate]
            B --> C{Blacklist Engine}
            C -->|Forbidden| D[Return Security Error]
            C -->|Allowed| E{Rego/CEL Policy Engine}
            E -->|Deny| D
            E -->|Allow| F[Execute Tool]
        end
    ```
* **APIs / Interfaces:**
    * Internal `PolicyGate` interface in `server/pkg/policy`.
    * Configuration schema update to support `gate_policies` blocks.
* **Data Storage/State:**
    * Policies are loaded from YAML/JSON configurations at startup and stored in high-performance memory trie/regex buffers.

## 5. Alternatives Considered
* **Model-Side Fine-Tuning**: Rejected due to high cost, lack of flexibility for users, and the "Catastrophic Forgetting" risk where models still bypass rules.
* **OS-Level AppArmor/SELinux**: Good for defense-in-depth but too complex for standard MCP Any users and doesn't provide fine-grained control over specific tool arguments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** OMPG is the "Kernel" of the MCP Any security model. It must be impossible to bypass via any Model Context Protocol transport.
* **Observability:** Every OMPG interdiction must be logged with high priority, including the agent ID, the forbidden tool, and the raw arguments.

## 7. Evolutionary Changelog
* **2026-04-11:** Initial Document Creation.
