# Design Doc: Universal Mission Checkpointer
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
As agentic workflows evolve from simple request-response cycles to long-running, multi-step "Missions," the fragility of the execution state becomes a critical bottleneck. Currently, if a local environment fails or if a user needs to migrate a high-compute task to the cloud, the entire agent context and mission progress are often lost.

The Universal Mission Checkpointer provides a standardized mechanism to generate bit-perfect, portable snapshots of an active mission's entire state—including the Shared Blackboard (KV store), Hierarchical Intent Trees (UACO), and active Upstream Tool Sessions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Generate serialized snapshots of the global Blackboard state.
    *   Capture and restore the cryptographic lineage of the Intent Tree.
    *   Enable cross-host mission migration (e.g., Local CLI to Cloud Gateway).
    *   Integrate with Binary State Handoff (BSH) for high-performance state transfer.
*   **Non-Goals:**
    *   Checkpointing the internal weights or hidden states of the LLM itself.
    *   Automatically resolving network-level socket persistence for non-MCP upstream tools.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Agent Swarm Orchestrator
*   **Primary Goal:** Resume a complex multi-hour code refactoring mission on a remote server after starting it on a local laptop.
*   **The Happy Path (Tasks):**
    1.  The user initiates a mission via `mcpany` CLI on their laptop.
    2.  The mission reaches a logical milestone (e.g., "Files Analyzed").
    3.  The Checkpointer generates a signed `.mcp-snapshot` file containing the BSH-encoded state.
    4.  The user uploads the snapshot to a remote MCP Any instance.
    5.  The remote instance validates the snapshot signature and restores the Blackboard and Intent Tree.
    6.  The agent swarm resumes execution from the exact milestone without re-analyzing files.

## 4. Design & Architecture
*   **System Flow:**
    `Orchestrator` -> `Trigger Snapshot` -> `Checkpointer` -> `Fetch Blackboard State` + `Fetch Intent Lineage` -> `BSH Encoder` -> `Signed Artifact`.
*   **APIs / Interfaces:**
    *   `POST /mission/snapshot`: Returns a unique snapshot ID or binary blob.
    *   `POST /mission/restore`: Accepts a snapshot blob and initializes the session.
*   **Data Storage/State:** Snapshots are stored as immutable, content-addressed blobs. State is primarily derived from the `Shared KV Store` and `UACO Manager`.

## 5. Alternatives Considered
*   **Manual Re-play:** Forcing the agent to re-read all previous logs. Rejected due to extreme token cost and non-deterministic divergence.
*   **Filesystem-Only Snapshots:** Just zipping the `.git` or local files. Rejected because it misses the "Internal Monologue" and Intent state stored in MCP Any's memory/DB.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Snapshots must be encrypted using the user's session key. Restoration requires re-attestation of the target environment's integrity (via Pre-Flight Sandbox Validator).
*   **Observability:** Snapshot events are logged in the `Mission Lifecycle Monitor`.

## 7. Evolutionary Changelog
*   **2026-04-11:** Initial Document Creation.
