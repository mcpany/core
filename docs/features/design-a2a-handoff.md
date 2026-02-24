# Design Doc: MoltHandoff A2A Adapter

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the release of OpenClaw's MoltHandoff protocol, there is now a standardized way for agents to perform "task handoffs" across framework boundaries. MCP Any needs to support this to maintain its position as the universal agent bus. This feature allows an agent (e.g., a local Claude Code instance) to securely hand off a sub-task (e.g., "Analyze this log file") to a specialized subagent (e.g., an OpenClaw instance with expert log-parsing skills) while preserving context and intent.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Support the MoltHandoff 1.0 JSON-RPC specification.
    *   Provide a "Handoff Proxy" that manages session state between the initiator and receiver.
    *   Enable "Handoff Sovereignty," allowing the initiator to set constraints and veto actions.
    *   Integrate with the MCP Any Policy Firewall for security.
*   **Non-Goals:**
    *   Replacing OpenClaw or other agent frameworks.
    *   Implementing the underlying LLM logic for the handoff reasoning.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator.
*   **Primary Goal:** Delegate a complex sub-task to a remote specialized agent without losing state.
*   **The Happy Path (Tasks):**
    1.  Primary Agent identifies a sub-task suited for a "Log Expert" subagent.
    2.  Primary Agent calls `mcpany_handoff_task(target="log-expert", task="Analyze server.log", context_id="session-123")`.
    3.  MCP Any validates the request and initiates a MoltHandoff handshake with the "log-expert" node.
    4.  The "log-expert" subagent receives the task and the minimal necessary context (via Recursive Context Protocol).
    5.  Once finished, the subagent returns the result to MCP Any, which passes it back to the Primary Agent.

## 4. Design & Architecture
*   **System Flow:**
    - **Handoff Manager**: A new middleware component that tracks active handoffs and their lifecycle (Initiated -> Accepted -> Running -> Completed/Failed).
    - **State Bridging**: Synchronizes the "Blackboard" (Shared KV Store) state relevant to the handoff.
*   **APIs / Interfaces:**
    - `POST /handoff/initiate`: Internal API for the local agent to start a handoff.
    - `POST /handoff/receive`: External-facing endpoint for remote agents to accept handoffs.
*   **Data Storage/State:** Session data stored in the shared SQLite Blackboard, keyed by `handoff_id`.

## 5. Alternatives Considered
*   **Direct A2A Communication**: Letting agents talk directly to each other. *Rejected* because it bypasses the Zero-Trust security and observability layer of MCP Any.
*   **Task-only Handoffs**: Not sharing state/context. *Rejected* as it leads to "Handoff Hallucinations" where the subagent lacks necessary background.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Handoffs require mutually authenticated TLS (mTLS) or signed tokens. The Policy Firewall must approve the handoff target and the data being shared.
*   **Observability:** The "Agent Chain Tracer" UI will visualize the handoff as a node transition in the execution timeline.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
