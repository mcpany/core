# Design Doc: Global Intent-Tracing (GIT)

**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
In a complex multi-agent swarm, a single user request can trigger a chain of subagent spawns and tool calls. Current MCP implementations lack a standardized way to correlate these downstream actions back to the original user intent. This creates a security gap where subagents can perform unauthorized actions, and an observability gap where it's impossible to reconstruct the "why" behind a specific tool call. Global Intent-Tracing (GIT) provides a universal ID and metadata container that persists through the entire agentic lifecycle.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a unique `X-Intent-ID` that is automatically injected into every MCP request and response.
    *   Enable the Policy Engine to enforce "Intent-Bound" permissions (e.g., "Only allow file writes if the original intent was 'Fix Bug'").
    *   Standardize the propagation of intent across different agent frameworks via the A2A Bridge.
    *   Provide a secure, tamper-proof audit trail for swarm actions.
*   **Non-Goals:**
    *   Inferring intent automatically (Intent is declared by the entry-point agent or user).
    *   Replacing existing OpenTelemetry or logging standards (GIT complements them).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Auditor.
*   **Primary Goal:** Trace a suspicious `rm -rf` command called by a 3rd-level subagent back to the original user prompt.
*   **The Happy Path (Tasks):**
    1.  User prompts the Main Agent: "Clean up the temporary build artifacts."
    2.  Main Agent generates a `GIT-ID: 8f2a...` with metadata `intent: cleanup`.
    3.  Main Agent spawns a Subagent to find the files. The `GIT-ID` is passed in the MCP headers.
    4.  Subagent finds files and calls a specialized Deletion Agent. The `GIT-ID` persists.
    5.  Deletion Agent calls the `delete_file` tool.
    6.  The Policy Engine checks the `GIT-ID` and metadata. Since `intent: cleanup` is authorized for `delete_file` in `/tmp`, the call is allowed.
    7.  Auditor later searches for `GIT-ID: 8f2a...` in the MCP Any UI and sees the entire chain from prompt to deletion.

## 4. Design & Architecture
*   **System Flow:**
    - **Injection**: Entry-point adapters (e.g., HTTP, Stdio) generate a `GIT-ID` if none exists.
    - **Propagation**: Every middleware in MCP Any is responsible for extracting and re-injecting the `GIT-ID` into upstream/downstream calls.
    - **Validation**: The Policy Engine uses the `GIT-ID` as a key to look up intent metadata in the Shared KV Store (Blackboard).
*   **APIs / Interfaces:**
    - **MCP Headers**: `x-mcp-intent-id`, `x-mcp-intent-metadata`.
    - **Middleware Interface**: New `IntentAwareMiddleware` interface in Go.
*   **Data Storage/State:** Intent metadata is stored in the `Shared KV Store` with a TTL matching the agent session.

## 5. Alternatives Considered
*   **Using TraceID (OTEL)**: Reusing existing distributed tracing IDs. *Rejected* because TraceIDs are for performance/debugging and often sampled; GIT IDs are for security/policy and must be 100% durable and metadata-rich.
*   **Context Window Injection**: Passing the intent in the LLM prompt. *Rejected* because it's prone to "Prompt Injection" and "Context Drift."

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The `GIT-ID` must be signed or stored in a way that downstream subagents cannot forge or escalate their intent metadata.
*   **Observability:** Integrated into the MCP Any Dashboard "Trace View."

## 7. Evolutionary Changelog
*   **2026-03-11:** Initial Document Creation.
