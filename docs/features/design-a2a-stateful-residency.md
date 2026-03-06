# Design Doc: A2A Stateful Residency (Resident Hub)

**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
Multi-agent swarms (e.g., OpenClaw, CrewAI) often face "Session Drift" where the primary orchestrator disconnects or loses state between intermittent tool calls. The "Stateful Residency" feature evolves MCP Any from a simple message bridge into a "Resident Hub"—a persistent home for agent sessions. This ensures that even if the calling client drops, the sub-agent task and its associated context remain resident and retrievable.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide persistent storage for A2A message queues and agent session states.
    *   Implement "Lease-based" residency where agent state is held for a configurable duration.
    *   Support asynchronous "Push/Pull" for agent task results.
    *   Enable "Handoff Recovery" if a primary agent fails during a delegation.
*   **Non-Goals:**
    *   Replacing external databases for long-term data archival.
    *   Providing a full-blown workflow engine (MCP Any focuses on state residency).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Developer.
*   **Primary Goal:** Ensure a 30-minute research task survives a transient network failure of the orchestrating CLI.
*   **The Happy Path (Tasks):**
    1.  Orchestrator initiates a task via MCP Any's A2A Bridge and requests "Resident Status."
    2.  MCP Any assigns a `session_residency_id` and stores the initial state in the `Shared KV Store`.
    3.  The Orchestrator CLI disconnects.
    4.  The sub-agent completes the task and posts the result to the Resident Hub.
    5.  The Orchestrator reconnects, provides the `session_residency_id`, and retrieves the completed state.

## 4. Design & Architecture
*   **System Flow:**
    - **Session Ingestion**: The `ResidentHubMiddleware` intercepts A2A calls and creates a residency record.
    - **Persistence Layer**: State is serialized and stored in the embedded SQLite `Shared KV Store`.
    - **TTL Manager**: A background routine prunes expired sessions based on the lease policy.
*   **APIs / Interfaces:**
    - `POST /a2a/sessions/residency`: Create or update a resident session.
    - `GET /a2a/sessions/residency/{id}`: Retrieve current state and mailbox.
*   **Data Storage/State:** Extension of the `Shared KV Store` schema to include `residency_metadata` (Owner, Lease, Status).

## 5. Alternatives Considered
*   **Client-Side Persistence**: Forcing the CLI to manage state. *Rejected* because it doesn't solve for "Disconnected Orchestration."
*   **External Redis**: Using Redis for state. *Rejected* to maintain MCP Any's "Zero-Dependency" local-first philosophy, though it may be an option for Enterprise deployments.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Session retrieval requires the same `AttestationToken` used during creation. Access is bound to the original "Identity Chain."
*   **Observability:** The UI "Agent Chain Tracer" will show "Resident" status for active sessions, including TTL and last-sync timestamps.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
