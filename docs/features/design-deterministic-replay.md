# Design Doc: Deterministic Agent Replay Engine (DARE)
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
Debugging multi-agent swarms is notoriously difficult due to the non-deterministic nature of LLMs and tool execution. When an agent chain fails, it is often impossible to reproduce the exact state to identify if the failure was due to a model hallucination, a tool error, or a state corruption during handoff. DARE (Deterministic Agent Replay Engine) solves this by providing a "Black Box" recorder for agent sessions, allowing exact replay of tool interactions.

## 2. Goals & Non-Goals
* **Goals:**
    * Capture point-in-time snapshots of tool inputs, outputs, and session metadata.
    * Provide a mechanism to "replay" a session, substituting live tool calls with captured results.
    * Support "Time-Travel Debugging" where a developer can modify a state mid-replay to test alternative outcomes.
* **Non-Goals:**
    * Snapshotting the internal weights or hidden states of the LLM itself.
    * Replacing existing observability tools like LangSmith or Phoenix (DARE focuses on the infrastructure/tool layer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Debugging Engineer
* **Primary Goal:** Reproduce a rare "State Corruption" error that occurs 10 steps into a multi-agent research workflow.
* **The Happy Path (Tasks):**
    1. Engineer enables `DARE_RECORDING` for a session ID.
    2. The agent swarm executes; MCP Any captures every tool call and response into a `dare-snapshot.json`.
    3. The workflow fails at step 10.
    4. Engineer runs `mcpany dare replay --snapshot dare-snapshot.json`.
    5. MCP Any reconstructs the environment, providing cached responses for steps 1-9, allowing the engineer to observe the exact failure at step 10 without re-running expensive or side-effect-heavy tools.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> DARE Middleware -> [Recorder | Replayer] -> Tool Registry`
* **APIs / Interfaces:**
    * `POST /dare/session/start`: Begins a recording.
    * `POST /dare/session/snapshot`: Manually triggers a state capture.
    * `GET /dare/session/{id}/export`: Exports the replay bundle.
* **Data Storage/State:**
    * Replay bundles are stored in a compressed Protobuf format to minimize storage overhead while preserving high-fidelity metadata.

## 5. Alternatives Considered
* **Log-Based Reconstruction**: Rejected because logs often lack the granular metadata (e.g., hidden environment variables) required for exact replay.
* **Database-Level Snapshots**: Rejected because it requires the upstream tools to support snapshotting, which most MCP servers do not.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Snapshots may contain sensitive PII or API keys. DARE will implement an "Auto-Redact" policy that scrubs known secret patterns before persisting the snapshot.
* **Observability**: Replay events are linked to the original session ID, providing a "Shadow Trace" in the UI.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
