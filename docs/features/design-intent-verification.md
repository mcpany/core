# Design Doc: Intent-Verification Middleware (IVM)
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of autonomous agent swarms (OpenClaw, Claude Code), traditional capability-based security (e.g., "can read /tmp") is insufficient. A compromised agent can still cause damage by using authorized tools in unauthorized ways (e.g., reading sensitive files in `/tmp` that weren't part of the original task). Intent-Verification Middleware (IVM) ensures that every tool call aligns with a cryptographically signed "Goal Manifest" established at the start of a session.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Validate tool calls against a session-level "Goal Manifest."
    *   Use a secondary, high-integrity "Security LLM" to perform real-time intent alignment checks.
    *   Provide "Justification Tokens" for every tool execution.
    *   Support "Autonomous HITL Triage" to reduce human approval fatigue.
*   **Non-Goals:**
    *   Replacing low-level RBAC/ABAC (IVM sits on top of them).
    *   Perfectly predicting all possible agent actions (focus is on risk mitigation).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin.
*   **Primary Goal:** Prevent an agent authorized to "manage cloud resources" from accidentally (or maliciously) deleting a production database if that wasn't the specific intent of the current session.
*   **The Happy Path (Tasks):**
    1.  The parent agent starts a session with a Goal Manifest: "Optimize dev environment storage."
    2.  MCP Any signs the manifest and issues a Session IVT.
    3.  The agent calls `delete_ebs_volume(id="vol-123")`.
    4.  IVM intercepts the call, sends the tool, arguments, and Goal Manifest to the Security LLM.
    5.  Security LLM verifies "vol-123" is a dev resource and confirms alignment.
    6.  The tool call proceeds. If it was a prod DB, IVM would block it and escalate to HITL.

## 4. Design & Architecture
*   **System Flow:**
    - **Manifest Creation**: At session start, the initial prompt/goal is hashed and signed.
    - **Interception**: Every `tools/call` is paused by the IVM.
    - **Alignment Check**: The IVM performs a "Similarity check" or "LLM-based reasoning" check between the `tool_name(args)` and the `Goal Manifest`.
    - **Token Issuance**: On success, an ephemeral "Intent-Bound Token" is generated to authorize the upstream call.
*   **APIs / Interfaces:**
    - New Internal Service: `IntentValidator`.
    - Extended MCP Header: `X-MCP-Intent-Token`.
*   **Data Storage/State:** Goal manifests are stored in the `Shared KV Store` tied to the session ID.

## 5. Alternatives Considered
*   **Static Policy (Rego)**: Too rigid for the dynamic nature of agentic reasoning.
*   **Manual HITL for everything**: Does not scale with AI discovery speed (the Claude 4.6 problem).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Security LLM itself must be isolated and have no tool access.
*   **Observability:** The "Reasoning Chain" for intent verification is logged in the Audit Trail.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
