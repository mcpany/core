# Design Doc: A2A Handoff Middleware (MoltHandoff Support)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents move from single-task executors to complex swarms, the ability to "hand off" a task from one agent to another is becoming critical. OpenClaw's "MoltHandoff" 1.0 protocol is the first standardized attempt at this. MCP Any needs to support this protocol to ensure that when a local agent (like OpenClaw) hands off a task to a cloud-based agent (like a specialized Claude instance), the session state, tool permissions, and context are preserved seamlessly.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Standardize the translation of MoltHandoff tokens into MCP-compatible session state.
    *   Implement "Signed Handoffs" to prevent prompt injection during agent-to-agent transfers.
    *   Provide a unified API for agents to initiate and receive handoffs across different frameworks.
    *   Ensure "Context Continuity" where the receiving agent has access to relevant history without context bloating.
*   **Non-Goals:**
    *   Arbitrating which agent is "better" for a task (this is handled by the orchestrator).
    *   Implementing the LLM reasoning for handoffs.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local Agent User (OpenClaw)
*   **Primary Goal:** Delegate a complex data analysis task to a cloud-based specialist without losing local file context.
*   **The Happy Path (Tasks):**
    1.  User asks local OpenClaw agent: "Analyze this 1GB CSV and give me a summary."
    2.  OpenClaw determines the task is too large for local processing.
    3.  OpenClaw calls `mcpany_initiate_handoff` with a MoltHandoff token.
    4.  MCP Any validates the token and signs it.
    5.  MCP Any routes the task to a cloud-based "Data Specialist" agent.
    6.  The Cloud Agent receives the "Handoff Bundle" which includes temporary access tokens for local file tools via MCP Any.
    7.  Cloud Agent completes task and hands back the result.

## 4. Design & Architecture
*   **System Flow:**
    - **Handoff Bundle**: A JSON object containing `session_id`, `state_snapshot`, `capability_tokens`, and a `signature`.
    - **Handoff Registry**: MCP Any maintains a short-lived registry of active handoffs to coordinate between the source and target agents.
    - **Proxy Layer**: MCP Any acts as the security proxy, ensuring the cloud agent only has access to tools explicitly granted in the handoff token.
*   **APIs / Interfaces:**
    ```json
    {
      "method": "a2a/handoff",
      "params": {
        "token": "string (MoltHandoff format)",
        "target_agent_id": "string",
        "context_scope": ["tool_id_1", "fs:read:/data/"]
      }
    }
    ```
*   **Data Storage/State:** Handoff state is stored in an encrypted Redis or SQLite "Blackboard."

## 5. Alternatives Considered
*   **Direct A2A Communication**: Letting agents talk directly to each other. *Rejected* because it bypasses centralized security policies and makes auditing impossible.
*   **Stateless Handoffs**: Passing the entire context in the handoff message. *Rejected* due to token limits and security risks of passing raw data.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All handoffs must be signed using the source agent's identity and verified by MCP Any. Capability tokens are time-bound and single-use.
*   **Observability:** A2A Handoffs are logged in the "Agent Chain Tracer" for debugging and auditing.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
