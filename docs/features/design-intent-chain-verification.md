# Design Doc: Intent-Aware Chain Verification

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents gain more autonomy, they are increasingly targeted by "Tool Chain Escalation" attacks. In these scenarios, an attacker uses a series of benign tool calls (e.g., `list_dir`, `read_file`) to gain enough information to execute a high-impact malicious call (e.g., `execute_command`, `delete_db`). Current security models only look at individual tool calls in isolation. Intent-Aware Chain Verification aims to look at the *sequence* and *logic* of tool calls to ensure they align with a verified, user-authorized intent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Track sequences of tool calls within a session.
    *   Verify that each subsequent tool call is a logical progression from previous calls based on a "Task Graph."
    *   Detect and block "reconnaissance-to-escalation" patterns.
    *   Provide an "Intent Score" for tool chains.
*   **Non-Goals:**
    *   Predicting every possible valid tool chain (focus on detecting known malicious patterns and enforcing user-defined intent boundaries).
    *   Adding significant latency to the tool execution loop.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious DevSecOps Engineer.
*   **Primary Goal:** Prevent an agent from escalating from "read-only" file access to "write/execute" access unless explicitly part of the authorized task.
*   **The Happy Path (Tasks):**
    1.  User authorizes a "Bug Fix" task.
    2.  Agent calls `list_files`, then `read_file` on a specific source file. (Verified: Logical progression).
    3.  Agent proposes a fix and calls `write_file` on the *same* file. (Verified: Logical progression).
    4.  *Malicious Path*: If the agent suddenly tries to call `execute_command("curl ...")` or `read_file("/etc/shadow")`, MCP Any detects the "Intent Drift" and blocks the call.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Capture**: When a task starts, the high-level goal is captured and hashed.
    - **Stateful Tracking**: The `ChainVerificationMiddleware` maintains a session-bound graph of all tool calls.
    - **Pattern Matching**: Uses a set of predefined "Safe Transition" rules (e.g., `READ -> READ`, `READ -> WRITE (same target)`, but NOT `READ -> EXECUTE` without intermediary user approval).
    - **Anomaly Detection**: If a tool call has low similarity to the stated intent or the current chain's context, it triggers an HITL (Human-in-the-Loop) challenge.
*   **APIs / Interfaces:**
    - Internal `CheckChain(sessionID, nextToolCall)` function integrated into the Tool Execution Pipeline.
*   **Data Storage/State:** Chain history is stored in the `Shared KV Store` (Blackboard) for the duration of the session.

## 5. Alternatives Considered
*   **Pure RBAC**: Too rigid; doesn't account for the dynamic nature of agentic workflows.
*   **LLM-based Verification**: Too slow and expensive for every tool call. We use a hybrid approach: fast rule-based checks for common patterns, LLM-based verification only for high-risk transitions.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core component of the Zero Trust architecture. It enforces "Temporal Least Privilege."
*   **Observability:** The UI will display a "Chain Health" indicator and visualize the tool call graph.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
