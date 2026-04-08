# Design Doc: Headless Orchestration Proxy (HOP)
**Status:** Draft
**Created:** 2026-04-08

## 1. Context and Scope
With the release of Claude Code's "Remote Control" and "Dispatch" capabilities, the primary usage pattern for AI agents is shifting from interactive terminal sessions to background, long-running processes. Current agent gateways are often tethered to a single active session or require a local terminal. The Headless Orchestration Proxy (HOP) evolves MCP Any into a persistent coordination layer that manages these background swarms, allowing users to connect, monitor, and "steer" autonomous agents from any authenticated client.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a persistent gateway for "Headless" agent sessions that survive client disconnections.
    * Implement a "Steering API" that allows remote users to inject high-priority intents into running swarms.
    * Support multi-tenant session isolation for background "Dispatch" workers.
    * Facilitate real-time telemetry streaming from headless agents to remote UI dashboards.
* **Non-Goals:**
    * Replacing the agent framework's internal logic (e.g., Claude's reasoning loop).
    * Providing long-term archival of agent monologues (focus is on active session coordination).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agentic DevSecOps Engineer
* **Primary Goal:** Connect to a background "Security Audit" swarm running in CI and steer it to focus on a specific directory.
* **The Happy Path (Tasks):**
    1. The engineer initiates a background swarm via `mcpany dispatch --task "Audit vulnerabilities"`.
    2. The HOP creates a persistent session ID and attaches the swarm to a headless proxy.
    3. The engineer later connects to the session via the MCP Any Web UI from a different machine.
    4. The UI displays the real-time reasoning trace of the CI-bound agents.
    5. The engineer uses the Steering Interface to send an interrupt: "Focus all teammates on `/pkg/auth`."
    6. The HOP propagates the intent to the swarm; the agents pivot their reasoning immediately.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Client[Remote UI/CLI] -- Authenticated WebSocket --> HOP[Headless Orchestration Proxy]
        HOP -- Persistent Pipe --> AgentMesh[Background Agent Swarm]
        AgentMesh -- State Updates --> Blackboard[(Shared Blackboard)]
        Blackboard -- Telemetry --> HOP
    ```
* **APIs / Interfaces:**
    * `HOP.create_session(mission_root)`: Initializes a persistent background context.
    * `HOP.attach(session_id)`: Bridges a remote client to an active session.
    * `HOP.inject_intent(session_id, priority_intent)`: Forces a mission-root update across all teammates.
* **Data Storage/State:**
    * Session metadata is stored in a lightweight SQLite index; active context is held in RAMS-compliant memory shards.

## 5. Alternatives Considered
* **SSH-based Tunneling**: Rejected as it is too low-level and doesn't provide semantic visibility into agent reasoning.
* **Framework-Specific Remote Control**: Rejected to maintain MCP Any's "Universal" value proposition (supporting mixed swarms of Claude, Gemini, and OpenClaw).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All remote connections to the HOP require hardware-attested tokens. Session hijacking is neutralized via mandatory origin-locked handshakes.
* **Observability:** Headless sessions are visualized in the "Global Agent Activity Map" with "Remote Steerable" badges.

## 7. Evolutionary Changelog
* **2026-04-08:** Initial Document Creation.
