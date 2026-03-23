# Design Doc: S2S Mesh Gateway (UACO v3.5)
**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
As AI agent deployments move from individual assistants to large-scale swarms, the bottleneck has shifted from agent-to-tool communication to collective-to-collective coordination. MCP Any must facilitate "Swarm-to-Swarm" (S2S) mesh networking to allow entire collectives to negotiate resource allocation and task delegation as single, cryptographically bound entities.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement full support for UACO v3.5 S2S Negotiation.
    * Provide a unified gateway for swarm-level handshakes.
    * Integrate "Swarm Wallets" for collective resource bidding.
* **Non-Goals:**
    * Individual agent-to-agent task routing (handled by previous A2A designs).
    * End-user UI for swarm management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Coordinate a specialized "DevOps Swarm" with a "Security Audit Swarm" for automated infra deployment.
* **The Happy Path (Tasks):**
    1. DevOps Swarm broadcasts a Task Card via UACO v3.5.
    2. MCP Any S2S Gateway authenticates the Security Swarm's collective signature.
    3. Both swarms negotiate resource credits via Swarm Wallets.
    4. Mission-root intents are synchronized across both swarm meshes.

## 4. Design & Architecture
* **System Flow:**
    `[Swarm A Mesh] <-> [MCP Any S2S Gateway] <-> [Swarm B Mesh]`
* **APIs / Interfaces:**
    * `/v1/s2s/handshake`: Multi-signature endpoint for swarm peering.
    * `/v1/s2s/negotiate`: UACO v3.5 bidding interface.
* **Data Storage/State:**
    * Swarm Peering Registry (Persistent).
    * Shared Intent Graph (Distributed via Mesh).

## 5. Alternatives Considered
* **Centralized Orchestrator:** Rejected due to single point of failure and sovereignty concerns.
* **P2P Direct Peering:** Rejected due to excessive complexity in managing N*M trust relationships.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All collective actions require multi-signature attestation from a quorum of specialist subagents.
* **Observability:** Mesh-wide tracing of S2S negotiation events.

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
