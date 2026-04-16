# Design Doc: Episodic Mesh Sharding (EMS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents handle long-running complex projects, they often hit context window limits. Gemini's "Chapters" model demonstrates the need to group interactions into semantic episodes.
MCP Any needs to host these "Episodes" as hardware-attested shards, allowing agents to swap state in and out without losing mission-root continuity or bloat.

## 2. Goals & Non-Goals
* **Goals:**
    * Host hardware-attested episodic shards for "Chapters" of work.
    * Allow sub-100ms resumption of previous project chapters.
    * Prevent "Mission-Root Erasure" during episode swaps.
* **Non-Goals:**
    * Unlimited long-term raw log storage (focuses on semantic state).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Chapter Research Swarm
* **Primary Goal:** Switch from "Competitor Analysis" chapter to "Price Strategy" chapter without reloading 500k tokens of previous research data.
* **The Happy Path (Tasks):**
    1. Researcher Agent completes "Competitor Analysis" and requests an "Episode Commit".
    2. EMS generates a hardware-attested shard of the semantic findings.
    3. Strategy Agent initiates the "Price Strategy" chapter.
    4. Strategy Agent requests the "Competitor Analysis" summary shard from EMS.
    5. EMS verifies the mission-root lineage and mounts the shard as a "Pinned Context Anchor".
    6. Agent proceeds with strategy reasoning using the compact, verified anchor.

## 4. Design & Architecture
* **System Flow:**
    `[Agent] -> [Episode Request] -> [EMS Shard Manager] -> [UEG Graph Ingestion] -> [TPM Signing] -> [Sharded Storage]`
* **APIs / Interfaces:**
    * `PUT /ems/v1/episode/commit`: Snapshots current semantic state into a shard.
    * `GET /ems/v1/episode/mount`: Provides an attested pointer to a previous chapter shard.
* **Data Storage/State:**
    * Sharded SQLite/Graph segments managed by the UEG Memory Broker.

## 5. Alternatives Considered
* **Vector RAG**: Rejected because RAG often misses mission-critical instructions that must remain pinned, not just retrieved by similarity.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shards are cryptographically bound to specific mission IDs.
* **Observability:** Track shard hit/miss rates and token savings in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
