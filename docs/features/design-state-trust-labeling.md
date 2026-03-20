# Design Doc: State-Trust Labeling (STL) Provider
**Status:** Draft
**Created:** 2026-05-19

## 1. Context and Scope
As AI agent swarms evolve from single-framework to heterogeneous "Agent Teams" (e.g., combining Claude Code, OpenClaw, and Gemini CLI), they are increasingly vulnerable to **Protocol-Agnostic State Injection (PASI)**. This occurs when an agent ingests state from a lower-trust origin and propagates it into a high-trust reasoning loop. The STL Provider solves this by cryptographically tagging every data fragment in the Shared KV Store (Blackboard) with its framework origin and trust level.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Trust Labeling" mechanism for all Blackboard data.
    * Provide a "Trust Translation Layer" to map metadata across UAB, A2A, and MCP.
    * Enforce "Trust-Bound Reading," where high-trust agents are alerted or blocked when accessing low-trust data.
* **Non-Goals:**
    * Providing a global identity for all agents (focus is on data trust).
    * Modifying the internal reasoning of the agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Orchestrator
* **Primary Goal:** Ensure a Claude-led team does not base critical decisions on unauthenticated data injected by a legacy subagent.
* **The Happy Path (Tasks):**
    1. An OpenClaw subagent writes a task result to the Blackboard via the UAB adapter.
    2. The STL Provider intercepts the write and cryptographically tags it with `trust_level: hardware_attested`.
    3. A legacy subagent writes a result to the same Blackboard via an unauthenticated MCP server.
    4. The STL Provider tags it with `trust_level: low_unverified`.
    5. A supervisor agent attempts to read both fragments; the STL Provider flags the `low_unverified` data, preventing trust pollution in the reasoning loop.

## 4. Design & Architecture
* **System Flow:**
    `[Agent Write/Read] -> [STL Provider] -> [Blackboard (SQLite)]`
* **APIs / Interfaces:**
    * `STL.label_data(key, value, origin_token)`: Computes and attaches trust metadata to a write operation.
    * `STL.verify_trust(key)`: Returns the cryptographically verified trust level of a data fragment.
    * `STL.translate_metadata(raw_headers)`: Normalizes disparate framework metadata into the STL standard.
* **Data Storage/State:**
    * Trust labels are stored as metadata columns in the Blackboard SQLite table, signed by the MCP Any master key.

## 5. Alternatives Considered
* **Framework-Specific Tagging**: Rejected as it fails to provide a unified trust worldview across different frameworks.
* **Global Access Control (ACL)**: Rejected as ACLs manage *who* can access, while STL manages the *provenance* and *reliability* of the data itself.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The STL labels must be immutable and signed to prevent "Label Spoofing" by compromised agents.
* **Observability**: Trust distribution and "Trust Violation" events will be visualized in the Blackboard Isolation Inspector.

## 7. Evolutionary Changelog
* **2026-05-19:** Initial Document Creation.
