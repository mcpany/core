# Design Doc: Intent-Aware Policy Engine
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
As AI agents become more autonomous and specialized, traditional Role-Based Access Control (RBAC) and static capability tokens are failing to prevent sophisticated "Intent-based" attacks. A malicious or hijacked agent might have the permission to `fs:write`, but using that permission to exfiltrate data or overwrite system configs while the user's stated goal is "Write a thank you email" should be blocked.

MCP Any must evolve its Policy Firewall from static rule-checking to dynamic "Intent-Aware" validation. This ensures that every tool call is not only *authorized* by identity but *justified* by the current task context.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept every tool call and evaluate it against the "High-Level Intent" of the session.
    * Use a lightweight LLM (or a specialized classifier) to determine if a tool call aligns with the stated goal.
    * Provide "Intent-Mismatch" feedback to the agent, allowing for self-correction or triggering HITL approval.
    * Maintain low latency by caching common intent/tool-pattern pairs.
* **Non-Goals:**
    * Replacing the underlying RBAC/Capability system (this is an additional layer).
    * Predicting all possible malicious actions (focuses on misalignment with stated goals).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Developer
* **Primary Goal:** Prevent an autonomous "Research Agent" from performing unauthorized system modifications if it gets hijacked by a prompt injection.
* **The Happy Path (Tasks):**
    1. User starts an agent session with the intent: "Research the latest trends in renewable energy and save the summary to a text file."
    2. MCP Any records this "Session Intent."
    3. Agent calls `web_search` -> Policy Engine checks against intent -> Approved.
    4. Agent calls `fs:write` to `research_summary.txt` -> Policy Engine checks against intent -> Approved.
    5. Agent (hijacked) attempts to call `fs:read` on `~/.ssh/id_rsa` -> Policy Engine detects mismatch with "Renewable Energy Research" -> Blocked.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCP Any: Tool Call (Tool: fs:read, Args: /etc/passwd)
        MCP Any->>Policy Engine: Validate(Call, SessionContext)
        Policy Engine->>Intent Classifier: Compare(Call, StatedIntent)
        Intent Classifier-->>Policy Engine: Result (MISMATCH: 98% Confidence)
        Policy Engine-->>MCP Any: Deny (Intent Violation)
        MCP Any-->>Agent: Error (403 Forbidden: Tool call does not align with session intent)
    ```
* **APIs / Interfaces:**
    * `SetSessionIntent(session_id, intent_string)`: Sets the "Root of Trust" intent for the current agentic flow.
    * `ValidateToolCall(context, call)`: Middleware hook in the execution pipeline.
* **Data Storage/State:**
    * Persistent session store (Redis/SQLite) mapping `session_id` to `CurrentIntent`.

## 5. Alternatives Considered
* **Static Manifests:** Requiring agents to declare all tools upfront. *Rejected* because it breaks the flexibility of autonomous agents and "On-Demand Discovery."
* **Human-in-the-Loop for everything:** *Rejected* due to extreme latency and user fatigue.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Stated Intent" must be immutable once set by the user or a high-trust orchestration layer.
* **Observability:** Logs will include "Intent Alignment Scores" for every blocked or suspicious tool call.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
