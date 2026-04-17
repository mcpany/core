# Design Doc: Remote Control Bridge (RCB)
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
With the release of "Remote Control" and "Dispatch" in Claude Code, AI agents are transitioning from terminal-bound solo sessions to distributed infrastructure components. Agents now run as background workers on remote servers, in CI pipelines, or on headless edge devices.

The Remote Control Bridge (RCB) is required to provide a secure, hardware-attested gateway for monitoring and steering these headless sessions. Without the RCB, users are tethered to the machine where the session started, limiting the flexibility and scalability of autonomous agent swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate secure, authenticated connections to headless agent sessions.
    * Provide a real-time "Steering" interface for human intervention in background processes.
    * Implement hardware-attested identity verification for both the remote agent and the human controller.
    * Support "Session Handover" between different control nodes.
* **Non-Goals:**
    * Replacing the primary agent reasoning engine.
    * Providing general-purpose remote shell access (RCB is specific to agent command-and-control).
    * Modifying agent internal state directly without reasoning-engine consent.

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer / Swarm Supervisor
* **Primary Goal:** Monitor a long-running refactoring task on a production build server and intervene when the agent requests a high-stakes credential.
* **The Happy Path (Tasks):**
    1. The Engineer dispatches a refactoring agent to the headless build server via MCP Any.
    2. The agent begins work and encounters a security boundary requiring a production API key.
    3. The agent pauses its reasoning loop and registers a "Waiting for Input" signal with the RCB.
    4. The Engineer receives a notification on their local workstation.
    5. The Engineer uses the RCB to connect to the active remote session, authenticated via a TPM-bound handshake.
    6. The Engineer reviews the agent's reasoning plan, provides the necessary credential/guidance, and resumes the task.
    7. The RCB logs the intervention and closes the remote steering channel.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Headless Agent] <-->|Local Bus| B(RCB Agent Node)
        B <-->|Hardware-Attested Tunnel| C(RCB Control Gateway)
        C <-->|Secure WebSocket| D[User Dashboard/CLI]
        E[TPM/Secure Enclave] -->|Attestation| B
        E -->|Attestation| D
    ```
* **APIs / Interfaces:**
    * `rcb.RegisterSession(sessionID, missionRoot) -> RCB_ID`: Registers a new headless session.
    * `rcb.ConnectControl(rcbID, authChallenge) -> ControlChannel`: Establishes a steering connection.
    * `rcb.PushIntervention(rcbID, inputData) -> status`: Injects human guidance into the agent loop.
* **Data Storage/State:**
    * Session metadata and control-node fingerprints are stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Standard SSH/TMUX**: Rejected because it lacks agentic context and cannot enforce mission-bound safety policies or hardware-attested steering.
* **Direct WebSocket Proxying**: Rejected due to the lack of "Auth-before-Discovery" and the high risk of session hijacking in remote environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Mandatory `LOWA` and `Origin` validation are enforced for the control channel.
* **Observability**: Remote steering events are captured by the `ARC` for policy evaluation and auditing.

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation (Iteration 2).
