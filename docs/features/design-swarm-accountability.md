# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Swarm Accountability Audit Log
**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
The rapid rise of autonomous agent swarms has created a "visibility vacuum." Users are increasingly concerned about "rogue" sub-agents or untraceable actions taken by deep agent chains. Current logs are flat and lack the hierarchical context of agentic delegation. MCP Any must provide a verifiable, hierarchical record of every action taken within a swarm.

## 2. Goals & Non-Goals
* **Goals:**
    * Create a cryptographically signed, immutable record of all agent-to-agent handoffs and sub-agent spawns.
    * Provide a hierarchical "parent-child" relationship for all tool calls within a swarm session.
    * Support "Proof of Intent" where a parent agent's mission statement is linked to every sub-agent action.
* **Non-Goals:**
    * Storing full LLM chat histories (this remains in the agent framework).
    * Enforcing real-time blocking (this is handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Auditor / Developer
* **Primary Goal:** Verify why a specific "delete_file" tool was called by a sub-sub-agent in a 50-agent swarm.
* **The Happy Path (Tasks):**
    1. The user opens the **Agent Chain Tracer** in the MCP Any UI.
    2. They select the suspicious "delete_file" event.
    3. The UI displays the full delegation path: `Orchestrator -> Researcher -> Janitor Sub-agent -> delete_file`.
    4. Each node shows the "Intent" signature and the "Parent Authorization" token.

## 4. Design & Architecture
* **System Flow:**
    `Agent Action -> Middleware Hook -> Signature Service -> Append-Only Audit DB (SQLite/AAL) -> UI Tracer`
* **APIs / Interfaces:**
    * `GET /audit/swarm/{session_id}`: Returns the full hierarchical trace for a swarm session.
    * `POST /audit/verify`: Validates the cryptographic signatures of a specific trace segment.
* **Data Storage/State:**
    * An internal "Agent Accountability Ledger" (AAL) using SQLite with WAL mode, with every row containing a HMAC of the previous row to ensure integrity.

## 5. Alternatives Considered
* **Standard Structured Logging**: Rejected because it doesn't provide the cryptographic proof needed for "Zero Trust" accountability.
* **Blockchain-based Audit**: Rejected due to excessive latency and complexity for local-first agent execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The AAL itself is protected by system-level permissions. Audit logs cannot be modified by agents themselves.
* **Observability**: Direct integration with the UI "Agent Chain Tracer" for visual debugging.

## 7. Evolutionary Changelog
* **2026-03-08**: Initial Document Creation.
