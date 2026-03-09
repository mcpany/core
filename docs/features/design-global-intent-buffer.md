# Design Doc: Global Intent Buffer
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
In complex agent swarms (e.g., OpenClaw swarms or multi-agent planning in Gemini CLI), subagents often lose sight of the high-level goal (the "Global Intent") as they descend into deep execution branches. This leads to "Intent Drift," redundant tool calls, and misaligned actions. The Global Intent Buffer is a persistent context layer that is automatically injected into every tool call across an entire agent session, regardless of swarm depth.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Maintain an immutable "Global Intent" string throughout an agent session.
    *   Automatically inject this intent into the headers/metadata of all downstream MCP calls.
    *   Provide a standardized way for subagents to query the current session goal without bloated prompt histories.
    *   Support "Intent-Aware" policy enforcement in the Policy Firewall.
*   **Non-Goals:**
    *   Automatically updating the intent (it must be set by the root agent or user).
    *   Replacing the standard LLM context window or conversation history.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Developer.
*   **Primary Goal:** Ensure a subagent at depth 5 knows it is part of a "Financial Audit" task and doesn't wander into unrelated tasks.
*   **The Happy Path (Tasks):**
    1.  User starts a session with an "Initial Intent" (e.g., "Analyze Q4 revenue for irregularities").
    2.  MCP Any stores this in the session's Global Intent Buffer (backed by SQLite).
    3.  As the root agent spawns subagents, MCP Any automatically propagates the `X-MCP-Global-Intent` header.
    4.  The subagent's tool calls are logged with the intent, and its LLM prompt is augmented with the intent for better alignment.

## 4. Design & Architecture
*   **System Flow:**
    - **Intent Middleware**: A new layer in the request pipeline that intercepts incoming agent requests and injects the buffer content.
    - **Session Store**: Uses the existing `Shared KV Store` (SQLite) to persist the intent for the duration of the session.
    - **Header Propagation**: Standardizes the use of `X-MCP-Intent` across all adapters.
*   **APIs / Interfaces:**
    - New endpoint: `POST /session/{id}/intent`
    - Tool schema extension: `intent_aware: true` (optional metadata).
*   **Data Storage/State:** Stored in the `session_metadata` table in the shared KV store.

## 5. Alternatives Considered
*   **Recursive Prompting**: Including the intent in every sub-prompt. *Rejected* because it consumes too many tokens and is prone to being "lost" in long histories.
*   **Manual Header Passing**: Forcing developers to pass the header manually. *Rejected* as it's brittle and increases integration friction.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The intent can be used by the Policy Firewall to verify if a tool call is "Within Intent."
*   **Observability:** The UI Roadmap's "Global Intent Viewer" will visualize the intent propagation and session alignment.

## 7. Evolutionary Changelog
*   **2026-03-09:** Initial Document Creation.
