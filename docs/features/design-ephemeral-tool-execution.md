# Design Doc: Ephemeral Tool Execution Environment (JIT Sandbox)
**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
With the rise of "Programmatic Tool Calling" (e.g., Claude's code execution for tool orchestration), agents now require more than just an API to call. They need a runtime to execute orchestration logic (loops, conditionals, data transformations) that bridges multiple MCP tool calls. Currently, this execution happens in unverified or broad-permission environments. MCP Any needs to provide a Just-In-Time (JIT) sandboxed environment to execute these programmatic blocks safely and efficiently.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a restricted execution environment for agent-generated orchestration scripts.
    * Enforce strict resource limits (CPU, Memory, Network) on ephemeral sandboxes.
    * Enable "Tool-to-Tool" communication within the sandbox via internal named pipes or loopback.
* **Non-Goals:**
    * Building a general-purpose cloud IDE.
    * Supporting long-lived persistent background processes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Orchestrator
* **Primary Goal:** Execute a complex workflow (e.g., "Find all users in SQL, then enrichment via Clearbit API, then post to Slack") in a single, secure, atomized execution block.
* **The Happy Path (Tasks):**
    1. Agent submits a "Programmatic Call" block to MCP Any.
    2. MCP Any validates the request against the Policy Engine.
    3. MCP Any spins up an ephemeral, isolated container/process (JIT Sandbox).
    4. The sandbox executes the logic, calling internal MCP tools as needed.
    5. The sandbox returns the final result and self-destructs.

## 4. Design & Architecture
* **System Flow:**
    `LLM -> MCP Any Gateway -> Policy Engine -> Sandbox Manager -> [Ephemeral Sandbox] -> Upstream Tools`
* **APIs / Interfaces:**
    * `POST /v1/sandbox/execute`: Accepts a script (JS/Python) and a set of tool-capability tokens.
* **Data Storage/State:**
    * Stateless by default. Optional access to the "Shared KV Store" (Blackboard) via capability tokens.

## 5. Alternatives Considered
* **Host-level Execution:** Rejected due to high risk of "NeighborJack" and code injection exploits.
* **Pre-provisioned Sandboxes:** Rejected as they consume idle resources and increase the attack surface. JIT (Just-In-Time) is preferred for security.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox has zero network access except to the specific MCP Any internal tool gateway. Filesystem access is limited to a `/tmp` mount.
* **Observability:** Full execution logs (stdout/stderr) are streamed to the Audit Log and visible in the UI Timeline.

## 7. Evolutionary Changelog
* **2026-02-28:** Initial Document Creation.
