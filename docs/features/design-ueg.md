# Design Doc: Universal Episodic Graph (UEG)
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
Current state management in AI agent swarms often suffers from "Context Amnesia" or "Memory Smearing." When multiple specialists from different frameworks (e.g., Claude Code, OpenClaw) interact via the Blackboard, the temporal and semantic relationship between their reasoning steps is lost. The Universal Episodic Graph (UEG) evolves the Shared KV Store into a hardware-attested graph database where every state change is an edge linked to a specific reasoning event, mission root, and agent identity.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a framework-neutral graph representation of agent memory.
    * Cryptographically link every memory fragment to its mission-root intent.
    * Support "Temporal Traversals" to allow agents to reason about previous mission branches.
    * Enable sub-millisecond retrieval of task-relevant context shards.
* **Non-Goals:**
    * Implementing a general-purpose graph database for non-agentic data.
    * Replacing long-term vector storage (UEG focuses on the active mission lifecycle).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent System Architect
* **Primary Goal:** Audit a failed swarm execution to identify exactly which subagent's reasoning fragment led to a conflicting state in the Blackboard.
* **The Happy Path (Tasks):**
    1. Architect opens the **Mesh-Resident Lineage Tracker**.
    2. Tracker queries the UEG for the "Mission Root" ID.
    3. UEG returns a directed acyclic graph (DAG) of all reasoning fragments and their resulting state mutations.
    4. Architect identifies a "Stylometric Mimicry" alert on a specific node, confirming an **Attention Splicing** attempt.

## 4. Design & Architecture
* **System Flow:**
    [Agent Reasoning] -> [SRM Provider] -> [UEG Ingestion]
    [UEG Ingestion] -> [Graph Engine (SQLite Graph Ext)] -> [Hardware Attestation (TPM)]
* **APIs / Interfaces:**
    * `POST /v1/memory/episodic/append`: Links a reasoning fragment to a state mutation.
    * `GET /v1/memory/episodic/trace/:mission_id`: Returns the full episodic DAG.
* **Data Storage/State:**
    * Utilizes a sharded SQLite backend with the `cr-sqlite` extension for Conflict-Free Replicated episodic state.

## 5. Alternatives Considered
* **Flat KV Store with Parent-ID Headers**: Rejected because it cannot represent the "many-to-many" relationships found in parallel teammate coordination (e.g., one mutation influenced by multiple teammates).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Graph nodes are cryptographically hash-chained to their predecessors (Multimodal Hash-Chaining).
* **Observability:** Powers the **Global Agent Activity Map** and **Swarm Topology Monitor**.

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
