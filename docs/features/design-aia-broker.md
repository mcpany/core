# Copyright 2026 Author(s) of MCP Any

# SPDX-License-Identifier: Apache-2.0

# Design Doc: Active Intent Alignment (AIA) Broker

**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope

Heterogeneous agent swarms often experience "Semantic Drift" where specialist
subagents diverge from the user's primary objectives during long-running
reasoning chains. This drift is often invisible to passive monitors until
an unauthorized or nonsensical tool call occurs.

The Active Intent Alignment (AIA) Broker introduces a proactive verification
layer. It issues hardware-attested "Alignment Heartbeats" that mandate
periodic proof of consistency between the active reasoning path and the
mission-root intent.

## 2. Goals & Non-Goals

* **Goals:**
    * Periodic verification of subagent reasoning traces against mission root.
    * Hardware-attested heartbeats to prevent alignment spoofing.
    * Automated suspension of sessions exhibiting cumulative drift.
    * Framework-neutral alignment signals.
* **Non-Goals:**
    * Real-time semantic correction (handled by parent agents).
    * Blocking tool calls based on static rules (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Ensure a deep chain of 10+ sub-delegations remains
  anchored to the user's original "Mission Root" intent.
* **The Happy Path (Tasks):**
    1. User initializes a complex mission with AIA enforcement enabled.
    2. A subagent begins a multi-hop reasoning cycle.
    3. AIA Broker intercepts the reasoning fragment and issues a heartbeat.
    4. AIA performs a "Semantic Hash-Comparison" with the mission root.
    5. Alignment score is verified and cryptographically signed by the TPM.
    6. Agent continues reasoning only after receiving the signed alignment
       token.

## 4. Design & Architecture

* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>Middleware: Reasoning Fragment
        Middleware->>AIA Broker: Alignment Heartbeat Request
        AIA Broker->>Mission Root: Fetch Intent Anchor
        AIA Broker->>Hardware: Semantic Hash Comparison
        Hardware-->>AIA Broker: Signed Score
        AIA Broker-->>Middleware: Alignment Token
        Middleware->>Agent: Proceed
    ```
* **APIs / Interfaces:**
    * `POST /v1/alignment/heartbeat`: Submit fragment for alignment check.
    * `GET /v1/alignment/status/{mission_id}`: Retrieve drift history.
* **Data Storage/State:**
    * Alignment scores and history are stored in the Shared KV Store
      (Blackboard) with STL protection.

## 5. Alternatives Considered

* **Parent-only Monitoring:** Rejected as parents can also drift or be coerced
  by deceptive subagent responses.
* **Binary Discovery Gates:** Rejected because drift often occurs *between*
  authorized tool calls.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Alignment tokens are tied to specific hardware
  sessions and cannot be reused across mission branches.
* **Observability:** Semantic drift metrics are visualized in the Active
  Intent Alignment Monitor.

## 7. Evolutionary Changelog

* **2026-06-18:** Initial Document Creation.
