# Design Doc: Autonomous Service Mesh Gateway
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As agents move from linear sequences to parallel teammate coordination (Claude Code Agent Teams), the "Universal Agent Bus" must evolve into a secure service mesh. Coordination stalls due to synchronous locking and "Shadow-Context" vulnerabilities prove that a passive bridge is no longer sufficient.

The Autonomous Service Mesh Gateway provides the secure transport and discovery infrastructure for inter-agent communication. It enforces "Auth-before-Discovery," manages IAMS-prioritized coordination, and hosts the "Kill-Switch" (ASKS) for immediate swarm-wide containment.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a secure, authenticated transport layer for inter-agent coordination (Named Pipes/WebSockets).
    * Implement "Auth-before-Discovery" for agent capability cards to prevent shadow mapping.
    * Host the Autonomous Swarm Kill-Switch (ASKS) for sub-millisecond capability revocation.
    * Support Interrupt-Aware Mailbox Sharding (IAMS) to resolve coordination stalls.
* **Non-Goals:**
    * Directly performing tool execution (handled by specialized adapters).
    * Providing long-term state persistence (handled by the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Revoke all capabilities of a specialist agent that has triggered a "Shadow-Context" logical trap.
* **The Happy Path (Tasks):**
    1. Anomaly Monitor detects an unauthorized tool proposal from a specialist agent.
    2. Monitor sends a high-priority interrupt to the Mesh Gateway.
    3. Gateway activates the Kill-Switch (ASKS) for that agent's identity fragment.
    4. Gateway broadcasts a "Revoke" signal over the IAMS priority channel to all teammates.
    5. Teammates immediately drop all coordination sessions with the rogue agent.
    6. Mission-root sovereignty is preserved.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Agent A] -- Encrypted Message --> B[Mesh Gateway]
        B -- Auth Check --> C[Identity Hub]
        C -- Valid --> D[Route to Agent B]
        D -- IAMS Priority --> E[Agent B Inbox]
        F[Anomaly Monitor] -- Kill Signal --> B
        B -- ASKS Trigger --> G[Revoke All Tokens]
    ```
* **APIs / Interfaces:**
    * `mesh.Discover(authToken) -> CapabilityBus`: Returns authorized tools and peers.
    * `mesh.Connect(targetAgent, missionToken)`: Establishes a secure coordination channel.
    * `mesh.TriggerKillSwitch(anomalyReport)`: Activates the ASKS protocol.
* **Data Storage/State:**
    * **Mesh Topology**: Real-time graph of active agent sessions and coordination channels.
    * **Revocation List (ARL)**: High-speed, in-memory cache of revoked identity fragments.

## 5. Alternatives Considered
* **Distributed Peer-to-Peer Coordination**: Rejected due to the difficulty of enforcing a centralized kill-switch and the latency of distributed consensus.
* **Git-based Mailbox (Claude Code style)**: Rejected due to the scaling bottlenecks (Lock Exhaustion) identified in 10+ agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** No agent capability is visible until a hardware-attested handshake is completed.
* **Observability:** Integrated with the "Swarm Topology Widget" for real-time visualization of inter-agent communication.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
