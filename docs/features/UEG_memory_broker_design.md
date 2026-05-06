# Design Doc: Universal Episodic Graph (UEG) Memory Broker

**Author:** Lead Systems Architect (L7)
**Date:** 2026-05-06
**Status:** Draft

## 1. Context and Problem Statement

As the AI agent ecosystem continues to scale from linear executions to complex, horizontal swarms (e.g., Claude Code teammates, OpenClaw specialists), memory management has become a critical bottleneck.

Currently, frameworks rely on isolated, flat context windows or basic shared KV stores without cryptographic lineage. This architecture is leading to two severe industry friction points:
1.  **Context Amnesia**: Agents lose track of the primary mission intent during handoffs because memory serialization is incompatible or lacks attention-pinning mechanisms.
2.  **Memory Smearing**: When shared state is not cryptographically isolated, subagents inject high-entropy noise into the shared context, diluting the "Mission-Root Intent" and exposing the swarm to intent drift.

We need a centralized memory broker that allows horizontal swarms to store, query, and enforce state securely, without relying on flat context windows.

## 2. Goals and Non-Goals

**Goals:**
- Evolve the existing Shared KV Store (Blackboard) into a hardware-attested, graph-based memory structure.
- Store memory as "Episodes" that are cryptographically linked to the mission-root intent.
- Enable subagents to query structural context efficiently without inheriting massive flat text blobs.
- Resolve "Context Amnesia" and "Memory Smearing" in deep swarms.

**Non-Goals:**
- Replacing local, short-term agent scratchpads completely. The UEG acts as the authoritative, shared mesh memory, not a replacement for immediate token window buffers unless they need to be persisted across handoffs.

## 3. Proposed Solution

The **Universal Episodic Graph (UEG) Memory Broker** will replace the traditional Shared KV Store with a graph database optimized for episodic memory storage.

### 3.1. Episodic Graph Structure
Instead of key-value pairs, memory will be stored as a directed graph where:
- **Nodes** represent individual facts, tool outputs, or "episodes".
- **Edges** represent the semantic and chronological relationship between these nodes.

### 3.2. Cryptographic Intent-Linking
Every node (episode) written to the UEG must be cryptographically signed and bound to the "Mission-Root Intent". This ensures that:
- Subagents can only write to branches of the memory graph they are authorized for.
- Reading operations (queries) can filter out high-entropy noise by only returning episodes linked to the verified mission lineage.

### 3.3. Querying & Sanitization
Agents will interact with the UEG via the Universal Multimodal Memory Bus (UMMB). When an agent requests context for a specific task:
- The UEG broker performs a semantic graph traversal to find relevant episodes.
- The returned data is structured context (JSON/schema) rather than raw text blobs, preventing context window flooding.

## 4. Alternative Solutions Considered

- **Hierarchical KV Store**: While simpler to implement on top of the current Blackboard, it fails to capture the complex relational nature of multi-agent handoffs, making context retrieval inefficient.
- **Increasing Context Window Sizes**: Relying purely on 1M+ token windows does not solve "Memory Smearing" or the security implications of un-isolated state sharing across heterogeneous frameworks.

## 5. Rollout Strategy
1.  Implement the core graph database engine alongside the existing Blackboard (dark launch).
2.  Update the UMMB to query both the Blackboard and UEG, validating data consistency.
3.  Migrate high-priority agent teams (e.g., OpenClaw specialists) to write exclusively to the UEG.
4.  Fully deprecate the flat Shared KV Store.

## 6. Security & Compliance
- **Hardware Attestation**: All memory write operations require a TPM-signed mission token to prevent rogue subagents from injecting false episodes.
- **Data Sovereignty**: The UEG will support environment-aware provenance (EAP) to bind memory fragments to specific hardware-attested container IDs, neutralizing trace replay attacks.
