# Design Doc: Nested HITL Approval Routing
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents evolve from simple task-executors to complex swarms with subagents (e.g., OpenClaw, Claude Code), the need for interactive user input becomes nested. A subagent might encounter an ambiguity (like a security confirmation or a clarification on a file path) and need to ask the user. Currently, most MCP implementations only support top-level tool-call approvals. This feature allows subagents to trigger interactive prompts that are routed through the parent agent and MCP Any to the primary user interface.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized MCP tool/protocol for subagents to request user input.
    * Maintain session and context integrity while waiting for user response.
    * Support "Intent-Aware" routing where the prompt includes the high-level goal context.
* **Non-Goals:**
    * Replacing the primary UI of the agent framework.
    * Handling real-time voice/video streams (focus on text/JSON interaction).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using a multi-agent swarm for complex refactoring.
* **Primary Goal:** Approve a risky subagent action (e.g., deleting a file) without losing the swarm's progress.
* **The Happy Path (Tasks):**
    1. Parent Agent spawns "Refactor Subagent".
    2. Subagent determines it needs to delete `old_config.yaml`.
    3. Subagent calls the `nested_ask_user` tool provided by MCP Any.
    4. MCP Any pauses the subagent's execution thread and pushes a notification to the Parent Agent's UI.
    5. User sees the prompt: "Refactor Subagent wants to delete `old_config.yaml`. Allow?"
    6. User clicks "Allow".
    7. MCP Any resumes the subagent with the "Allowed" response.
    8. Subagent completes the task.

## 4. Design & Architecture
* **System Flow:**
    `Subagent` -> `MCP Any Gateway (HITL Middleware)` -> `Primary Agent UI` -> `User` -> `MCP Any Gateway` -> `Subagent`
* **APIs / Interfaces:**
    * `tool: nested_ask_user`:
        * `prompt`: string (The question for the user)
        * `context_id`: string (Mapping to the parent session)
        * `options`: list of strings (e.g., ["Yes", "No", "Explain"])
* **Data Storage/State:**
    * Use the `Shared KV Store` to track "Pending Approvals" associated with `session_id`.

## 5. Alternatives Considered
* **Direct UI Connection for Subagents**: Rejected because it breaks the "Parent-controlled" security model and leads to UI clutter.
* **Polling for Response**: Rejected in favor of an event-driven "Suspension Protocol" to save resources.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All nested prompts must be signed by the subagent and verified against the parent's `intent-scope`.
* **Observability:** All user responses are logged in the `Audit Log` with full context of the subagent's state at the time of the request.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
