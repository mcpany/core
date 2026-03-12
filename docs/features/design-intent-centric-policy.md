# Design Doc: Intent-Centric Policy Middleware

**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
As AI agents evolve from single-purpose tools to complex swarms (e.g., OpenClaw, AutoGen), the risk of "Subagent Redirection" increases. A specialized subagent might be manipulated via prompt injection or state injection to perform actions outside its original mandate. Traditional capability-based security (e.g., "can read files") is no longer sufficient. MCP Any must transition to an "Intent-Centric" model where every tool call is validated against a cryptographically-bound "Master Intent."

## 2. Goals & Non-Goals
*   **Goals:**
    *   Bind all tool calls in a session to a user-verified "Master Intent."
    *   Provide a mechanism for agents to prove their current action aligns with the "Master Intent."
    *   Enable the Policy Engine to reject tool calls that deviate from the established intent.
    *   Support cryptographic signing of intents by the user.
*   **Non-Goals:**
    *   Solving all prompt injection (this is a mitigation, not a complete fix).
    *   Automating intent generation without user verification.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Developer using a multi-agent swarm for codebase refactoring.
*   **Primary Goal:** Ensure that no subagent can exfiltrate `.env` files even if it has "file read" capabilities.
*   **The Happy Path (Tasks):**
    1.  User initiates a task: "Refactor the authentication module."
    2.  MCP Any generates a "Master Intent" token and requests user signature.
    3.  User signs the intent via the MCP Any UI/CLI.
    4.  The orchestrator agent passes this signed token to all subagents.
    5.  A subagent attempts to call `read_file(".env")`.
    6.  The Policy Middleware intercepts the call, checks the "Master Intent" ("Refactor auth module"), and determines that reading `.env` is not required for this intent.
    7.  The call is blocked, and the user is notified of a potential policy violation.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Creation**: User-defined string -> Signed JWT/Token containing the intent and scope.
    - **Propagation**: Token is passed via MCP headers (using the Recursive Context Protocol).
    - **Verification**: The Policy Middleware uses an LLM-in-the-loop or a fast Semantic Classifier to compare the `tool_call` + `arguments` against the `intent_string`.
*   **APIs / Interfaces:**
    - `POST /v1/intent/sign`: Endpoint for user to sign a proposed intent.
    - Header: `X-MCP-Intent-Token: <JWT>`
*   **Data Storage/State:** Intent tokens are session-bound and stored in the Shared KV Store (Blackboard) with Agent-Bound isolation.

## 5. Alternatives Considered
*   **Static Whitelisting**: Hardcoding allowed tools per task. *Rejected* as it's too rigid for dynamic agent behaviors.
*   **Pure LLM Monitoring**: Running a second LLM to watch all traffic. *Rejected* due to latency and cost. Intent-tokens provide a structured way to scope the monitoring.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of "Zero Trust Agency." It moves from "What can you do?" to "Why are you doing it?"
*   **Observability:** The UI must show the "Active Intent" and highlight tool calls that were blocked or flagged by the intent-checker.

## 7. Evolutionary Changelog
*   **2026-03-12:** Initial Document Creation.
