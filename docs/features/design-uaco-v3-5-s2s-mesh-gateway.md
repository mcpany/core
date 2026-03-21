// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

# Design Doc: UACO v3.5 S2S Mesh Gateway

**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
As agent swarms become more specialized and autonomous, the bottleneck has shifted from individual agent coordination to inter-swarm collaboration. Different agent frameworks (OpenClaw, CrewAI, AutoGen) often operate as closed ecosystems. The Universal Agent Coordination Protocol (UACO) v3.5 introduces Swarm-to-Swarm (S2S) negotiation to bridge these silos.

MCP Any needs to implement the UACO v3.5 S2S Mesh Gateway to act as the authoritative broker for these high-level negotiations. It will manage "Swarm Wallets" and multi-signature handshakes, allowing collectives to interact as single units while maintaining Zero-Trust security.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement UACO v3.5 S2S handshake protocol.
    *   Manage "Swarm Wallets" for collective identity and resource attribution.
    *   Provide a secure, framework-agnostic gateway for inter-swarm task delegation.
    *   Enforce multi-signature (M-of-N) attestation for cross-swarm commitments.
*   **Non-Goals:**
    *   Replacing individual agent transport (Stdio/HTTP/WebSocket).
    *   Managing the internal reasoning logic of connected swarms.
    *   Providing long-term persistence for swarm-to-swarm chat history (handled by external Telemetry Sinks).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Orchestrator
*   **Primary Goal:** Delegate a security audit task from an "App Dev Swarm" (OpenClaw) to a specialized "SecOps Swarm" (CrewAI) without exposing individual subagent credentials.
*   **The Happy Path (Tasks):**
    1.  The App Dev Swarm (Source) broadcasts a "Task Proposal" to the MCP Any S2S Gateway.
    2.  The S2S Gateway identifies the SecOps Swarm (Target) via the UACO v3.5 discovery bus.
    3.  The SecOps Swarm submits a "Bid" with a "Swarm Wallet" signature.
    4.  The S2S Gateway validates the bid against the source swarm's "Mission Root" and budget.
    5.  A multi-signature handshake is established between the two swarms.
    6.  Task context is handed off via the S2S Gateway, and progress is tracked on the Shared Blackboard.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        SourceSwarm[App Dev Swarm] -->|Proposal| S2SGateway[MCP Any S2S Gateway]
        S2SGateway -->|Discovery| TargetSwarm[SecOps Swarm]
        TargetSwarm -->|Bid + Multi-Sig| S2SGateway
        S2SGateway -->|Attestation| PolicyEngine[Policy Engine]
        PolicyEngine -->|Approval| S2SGateway
        S2SGateway -->|Context Handoff| TargetSwarm
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/s2s/propose`: Submit a task proposal from a swarm.
    *   `POST /v1/s2s/bid`: Submit a bid for a proposed task.
    *   `GET /v1/s2s/discovery`: Discover available swarm capabilities.
    *   `GRPC S2SService`: High-speed bidirectional streaming for swarm-to-swarm state sync.
*   **Data Storage/State:**
    *   Swarm Identities and Wallets stored in the encrypted Identity Store.
    *   Active S2S Negotiations tracked in the Shared KV Store (Blackboard) with "Swarm-Bound" isolation.

## 5. Alternatives Considered
*   **Direct P2P Swarm Comms**: Rejected due to the "Trust Fragmentation" problem; individual swarms would need to manage hundreds of trust relationships. A central gateway provides a single attestation point.
*   **Legacy A2A Bridge**: Rejected because A2A v1 is agent-centric and doesn't support collective identity or multi-signature swarm handshakes required by UACO v3.5.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):**
    *   All S2S handshakes require hardware-attested "Swarm Wallet" tokens.
    *   Inter-swarm context handoffs are semantically sanitized by the Inference-Time Data Sanitizer (IDS).
*   **Observability:**
    *   Real-time tracing of S2S negotiation cycles via the Parallel Team Coordination Dashboard.
    *   Latency and token-cost metrics exported to the RL Telemetry Provider.

## 7. Evolutionary Changelog
*   **2026-05-15:** Initial Document Creation.
