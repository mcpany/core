# Design Doc: Intent-Bound Execution Sandbox
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of autonomous agent teams and the recent discovery of SSRF and Path Traversal vulnerabilities in OpenClaw (CVE-2026-26322, CVE-2026-26329), MCP Any needs a more robust security model. Traditional capability-based security (e.g., granting 'read' access to a directory) is too broad for autonomous swarms.

The "Intent-Bound Execution Sandbox" introduces a middleware layer that requires every tool call to be accompanied by a cryptographically signed "Intent Contract." This contract binds the tool call to the original high-level user request, preventing "Intent Drift" where a compromised agent uses legitimate tools for malicious secondary purposes.

## 2. Goals & Non-Goals
* **Goals:**
    * Prevent SSRF and Path Traversal by validating tool parameters against user-approved intents.
    * Provide a cryptographic audit trail of "Intent-to-Tool" mappings.
    * Enable "Team-Scoped" sandboxing where subagents inherit only the specific intent of their parent task.
* **Non-Goals:**
    * Replacing the Policy Firewall (Rego/CEL); this middleware complements it by adding intent-awareness.
    * Performing deep LLM reasoning on every call; it uses high-performance hashing and metadata matching.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise AI Architect.
* **Primary Goal:** Ensure that a "Web Researcher" agent cannot be tricked into using its HTTP tools to probe internal network metadata (SSRF).
* **The Happy Path (Tasks):**
    1. User initiates a task: "Find the latest pricing for Product X on the public web."
    2. MCP Any generates an **Intent Contract** with scope `public-web-research`.
    3. The Lead Agent delegates a task to a Subagent. The Subagent receives a **Scoped Intent Token**.
    4. Subagent calls `http_fetch(url="https://competitor.com")`.
    5. The Sandbox Middleware verifies the URL against the `public-web-research` policy (e.g., DNS resolution check, public IP verification).
    6. Subagent attempts to call `http_fetch(url="http://169.254.169.254/latest/meta-data")` (SSRF attempt).
    7. The Sandbox Middleware detects that this URL violates the `public-web-research` intent and blocks the call.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `Intent Middleware` -> `Policy Engine (Rego)` -> `Tool Execution`
* **APIs / Interfaces:**
    * `SetIntent(intent_string, scope_metadata)`: Generates a signed Intent Contract.
    * `VerifyCall(tool_id, args, intent_token)`: Middleware hook for validation.
* **Data Storage/State:**
    * Intent contracts are stored in the transient "Blackboard" KV store.

## 5. Alternatives Considered
* **Static Sandboxing**: Rejected because agents need dynamic access to tools.
* **Manual HITL for every call**: Rejected due to high latency and poor user experience.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Uses HMAC-SHA256 for intent tokens to prevent spoofing.
* **Observability**: Logs "Intent Violations" with full context for security audits.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
