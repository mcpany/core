# Design Doc: ACP-Native Messaging Bridge
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rapid industry adoption of the **Agent Communication Protocol (ACP)**, AI agents are moving toward a framework-neutral standard for peer-to-peer coordination. Currently, MCP Any supports framework-specific handoffs (e.g., Claude Code Agent Teams), but lacks a universal translation layer. To remain the indispensable core infrastructure for AI swarms, MCP Any must act as the authoritative "ACP Hub," bridging disparate framework protocols into a unified ACP messaging bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a native translation layer between ACP and framework-specific protocols.
    * Facilitate secure, hardware-attested inter-teammate messaging.
    * Support standardized ACP task objects and status updates.
    * Ensure low-latency message routing between local and remote agents.
* **Non-Goals:**
    * Replacing framework-specific internal coordination logic (e.g., Claude's leader-election).
    * Storing message payloads indefinitely (requires persistent mailbox shards).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Coordinate a Claude-led team with specialized OpenClaw subagents via ACP.
* **The Happy Path (Tasks):**
    1. The Claude team lead issues a task proposal to an OpenClaw specialist.
    2. The ACP Bridge intercepts the proprietary Claude message and translates it into an ACP `TaskProposal`.
    3. The OpenClaw agent receives and accepts the ACP task.
    4. The ACP Bridge routes the OpenClaw `StatusUpdate` back to the Claude team lead.
    5. The coordination is completed within a hardware-attested ACP session.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        C[Claude Code] <-->|Proprietary| AB[ACP Bridge]
        G[Gemini CLI] <-->|Proprietary| AB
        AB <-->|ACP| O[OpenClaw Agent]
        AB <-->|ACP| P[Other ACP Peers]
    ```
* **APIs / Interfaces:**
    * `POST /v1/acp/translate/ingress`: Translate a framework-specific message to ACP.
    * `POST /v1/acp/translate/egress`: Translate an ACP message to a target framework protocol.
    * `GET /v1/acp/peers`: Discover ACP-compliant teammates.
* **Data Storage/State:**
    * Mapping tables for framework-specific session IDs to ACP session tokens.

## 5. Alternatives Considered
* **Framework-Specific Bridges:** Rejected as it leads to $O(N^2)$ complexity for N frameworks.
* **Manual Protocol Mapping:** Rejected due to the high risk of intent leakage during manual translation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All ACP translations must preserve the hardware-attested lineage of the original message.
* **Observability:** Message translation latency and protocol errors are logged in the ACP Status Monitor.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
