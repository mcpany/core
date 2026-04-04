# Design Doc: Remote Dispatch Hub (RDH)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The transition of agent frameworks like Claude Code from local terminal tools to remote, persistent background workers (Dispatch) requires a new orchestration layer. Managing headless agents in CI/CD pipelines or remote servers introduces challenges in visibility, steering, and mission-root sovereignty.

MCP Any needs the Remote Dispatch Hub (RDH) to act as the authoritative "Remote Control" center. It provides the secure infrastructure to initiate, monitor, and hand off agent sessions across network boundaries without losing context or security bounds.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified REST API and WebSocket stream for steering remote, headless agents.
    * Support "Detach/Re-attach" logic for long-running agent missions.
    * Enforce mission-root sovereignty for all commands issued via remote steering.
    * Facilitate secure handoffs between different human or automated controllers.
* **Non-Goals:**
    * Replacing the agent's internal reasoning engine.
    * Providing a full-featured web IDE (RDH is an infrastructure component).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer
* **Primary Goal:** Trigger a headless Claude Code agent to fix a broken build in CI, monitor its progress via a mobile terminal, and step in manually if the agent stalls.
* **The Happy Path (Tasks):**
    1. CI pipeline triggers MCP Any's RDH to spawn a "Fix Build" session.
    2. RDH initiates a headless agent with a TPM-signed mission-root manifest.
    3. Engineer opens a WebSocket stream via the RDH dashboard to view real-time CoT reasoning.
    4. RDH detects a "Human-in-the-Loop" requirement and sends a push notification.
    5. Engineer issues a corrective intent through the RDH steering API.
    6. Agent completes the task and RDH signs the final commit.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[External Controller] -->|REST/WS| B[Remote Dispatch Hub]
        B -->|Command| C[Isolated Agent Worker]
        C -->|Trace/Output| B
        B -->|Sanitized Stream| A
        B -->|Audit Log| D[Regulatory Vault]
    ```
* **APIs / Interfaces:**
    * `/v1/dispatch/spawn`: Initiate a headless session with intent-scoping.
    * `/v1/dispatch/attach/{session_id}`: Establish a WebSocket for real-time steering.
    * `/v1/dispatch/handoff`: Transfer authority between controllers.
* **Data Storage/State:**
    * Session state is persisted in the Shared KV Store (Blackboard) with "RDH-Bound" isolation.

## 5. Alternatives Considered
* **Direct SSH to Workers:** Rejected due to lack of intent-aware auditing and complexity in context preservation during handoffs.
* **Standard CI Logs:** Rejected because they do not allow bi-directional steering and interactive HITL.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RDH endpoints mandate Origin-Locked (LOWA) authentication and hardware-attested session tokens.
* **Observability:** Session "Wait-Time" and "Controller Latency" are tracked in the Swarm Coherence Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
