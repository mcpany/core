# Design Doc: Intent-Bounded Governance Middleware
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Recent investigations (MITRE ATLAS, Feb 2026) have shown that autonomous agents are vulnerable to "trust abuse" exploits, where a legitimate tool is used for a malicious purpose that deviates from the user's original intent. MCP Any needs a layer that doesn't just check *permissions* (can the agent call this tool?) but also *intent* (should the agent be calling this tool right now?).

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept tool calls and validate them against a "Signed Intent Manifest".
    * Use a secondary, "Governance LLM" to perform semantic validation of tool parameters against the high-level task.
    * Provide a Zero-Trust perimeter that prevents agents from reconfiguring themselves or accessing sensitive tools outside of verified workflows.
* **Non-Goals:**
    * Solving the general "Alignment Problem" for LLMs.
    * Adding significant latency (targets <100ms for governance checks using small, specialized models).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Prevent an autonomous "Data Analysis" agent from using its file-access tools to exfiltrate SSH keys, even if it has "read" permissions on the home directory.
* **The Happy Path (Tasks):**
    1. User starts a task: "Analyze the CSV files in /data."
    2. MCP Any generates a "Signed Intent" for this session.
    3. Agent calls `read_file("/data/results.csv")` -> Governance check PASSES (aligns with intent).
    4. Agent (via prompt injection) calls `read_file("~/.ssh/id_rsa")` -> Governance check FAILS (outside "Analyze CSV" intent).
    5. MCP Any blocks the call and alerts the HITL Middleware.

## 4. Design & Architecture
* **System Flow:**
    `Agent Tool Call -> Policy Engine -> Governance LLM (Semantic Check) -> Tool Execution`
* **APIs / Interfaces:**
    * `X-MCP-Intent-Token`: A signed header containing the high-level intent description.
    * Governance Hook: A gRPC/Internal interface for plugging in different semantic validators.
* **Data Storage/State:**
    * Intent tokens are short-lived and session-bound.
    * Audit logs store the "Reasoning Path" for blocked calls.

## 5. Alternatives Considered
* **Static File Path Whitelisting**: Rejected because it's too brittle for dynamic agent workflows.
* **RBAC (Role-Based Access Control)**: Rejected because it doesn't account for the *context* of the call, only the identity of the agent.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The Governance LLM must be isolated and have no tool-access itself.
* **Observability**: Real-time dashboard of "Intent Deviations" to help humans identify potential prompt injection attacks.

## 7. Evolutionary Changelog
* **2026-02-27**: Initial Document Creation.
