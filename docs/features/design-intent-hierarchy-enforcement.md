# Design Doc: Intent Hierarchy Enforcement
**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope
With the rise of "Context Shadowing" (CVE-2026-39102), subagents can manipulate
the shared context to hijack the primary agent's mission. MCP Any needs an
authoritative mechanism to enforce the priority of intents across the universal
bus.

This feature ensures that instructions from the "Mission Root" (the user's
original intent) cannot be overridden or "shadowed" by malicious or
hallucinating subagents, regardless of the framework they originate from.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Tag every context fragment with a "Lineage Depth" and "Authority Score."
    *   Intercept and block tool calls that originate from shadowed or
        conflicting intents.
    *   Provide an immutable "Mission Root" memory segment that cannot be
        evicted.
*   **Non-Goals:**
    *   Implementing a new LLM reasoning engine.
    *   Replacing framework-specific state managers (e.g., OpenClaw's engine).
    *   Automated re-planning of failed sub-missions.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Architect for an Autonomous Swarm.
*   **Primary Goal:** Prevent a specialized "Research Subagent" from overriding
    the "Safety Subagent's" file-read restrictions.
*   **The Happy Path (Tasks):**
    1.  Primary Agent defines the Mission Root and Safety Constraints.
    2.  MCP Any tags these fragments with `Authority: 0 (Root)`.
    3.  Primary Agent delegates a task to a Subagent.
    4.  Subagent attempts to inject a command `IGNORE PREVIOUS INSTRUCTIONS` via
        a tool output.
    5.  MCP Any detects the conflict, tags the injection as `Authority: 2`, and
        filters it before it reaches the Primary Agent's next reasoning cycle.

## 4. Design & Architecture
*   **System Flow:**
    `Agent Request` -> `IHE Middleware` -> `Authority Tagging` -> `Conflict
    Resolver` -> `Tool Execution / Context Update`.
*   **APIs / Interfaces:**
    *   `POST /v1/intent/anchor`: Locks a fragment as the Mission Root.
    *   `GET /v1/context/validate`: Checks if a proposed context update
        violates hierarchy.
*   **Data Storage/State:**
    Authority scores and lineage metadata are stored in a session-bound SQLite
    "Intent Registry" within the Blackboard.

## 5. Alternatives Considered
*   **Plain Text Filtering:** Rejected because it cannot handle semantic
    variations of shadowing attacks and is prone to bypass via encoding.
*   **Framework-Specific Patches:** Rejected because it doesn't solve the
    interoperability problem for "Universal" swarms where different frameworks
    communicate via MCP Any.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The IHE itself must run in a TEE or isolated
    process to prevent tampering by higher-privilege subagents.
*   **Observability:** All "Intent Conflicts" are logged to the Security Audit
    Trail with full lineage traces.

## 7. Evolutionary Changelog
*   **2026-05-30:** Initial Document Creation.
