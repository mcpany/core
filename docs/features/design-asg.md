# Design Doc: Atomic Scratchpad Guard (ASG)
**Status:** Draft
**Created:** 2026-07-17

## 1. Context and Scope
Claude Code v3.0 has introduced "Agent Teams" that collaborate via a shared project-local `.scratchpad` directory. This pattern allows for non-linear reasoning but creates a critical security and stability frontier. Without an active guard, parallel agents can suffer from race conditions or, worse, a malicious specialist agent could "pollute" the shared scratchpad with deceptive instructions or exfiltrated context. The ASG provides high-entropy semantic security and atomic locking for shared team workspaces.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic analysis of all writes to shared team scratchpads.
    * Neutralize "Scratchpad Pollution" by blocking unauthorized instruction injection.
    * Provide mission-bound atomic locking for scratchpad fragments to prevent race conditions.
    * Ensure "Stitch-Resistant" isolation for context fragments within the shared workspace.
* **Non-Goals:**
    * Restricting legitimate collaborative reasoning.
    * Managing the inter-agent mailbox (handled by T2T/Mailbox Integrity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator
* **Primary Goal:** Enable 3 parallel agents to collaborate on a codebase using a shared scratchpad while ensuring that a compromised "Documentation Agent" cannot inject shell commands into the scratchpad used by the "Architect Agent."
* **The Happy Path (Tasks):**
    1. The orchestrator enables ASG for the mission's `.scratchpad` path.
    2. Teammate A (Architect) locks a specific fragment of the scratchpad.
    3. Teammate B (Doc) attempts to write an imperative shell instruction to the shared workspace.
    4. ASG's semantic engine identifies the instruction as unauthorized by the mission root.
    5. The write is blocked, and a security alert is broadcast to the swarm.

## 4. Design & Architecture
* **System Flow:**
  [Teammate Write] -> [ASG Middleware (Semantic Scanner)] -> [Atomic Lock Manager] -> [Filesystem]
* **APIs / Interfaces:**
    * `mcpany.asg.v1.ScratchpadGuard`
    * Hook: `preProcessScratchpadWrite(agent_id, content, path)`
* **Data Storage/State:**
    * In-memory wait-graphs for atomic locks; hardware-attested semantic hashes for content validation.

## 5. Alternatives Considered
* **OS-Level File Locking**: Rejected as it is semantically unaware and cannot prevent instruction injection.
* **Full Scratchpad Isolation**: Rejected as it breaks the core collaborative value of "Agent Teams."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All writes are treated as untrusted and scanned for out-of-scope instructions.
* **Observability:** Lock contention and sanitization events are visualized in the Swarm Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-17:** Initial Document Creation.

### Update: 2026-07-18 - Incorporating Reasoning-Aware Redaction (RAR)
**Context**: Today's market sync revealed the launch of OpenClaw v3.6 with RAR, which redacts intents at the edge. To remain the universal infrastructure layer, ASG must integrate with the RAR engine to protect shared scratchpads from "Context-Stitching" (CVE-2026-88012).
**Architecture Adjustment**:
*   **RAR Integration**: Appending the RAR engine as a mandatory pre-processor in Section 4.
*   **Atomic Arbiter**: Upgrading the Lock Manager to include "Conflict-Aware Redaction," ensuring that locked fragments are sanitized based on the reader's trust level.
**Security Impact**: Mitigates stylometric mimicry and cross-agent intent leakage in high-contention team workspaces.

### Update: 2026-07-19 - Neutralizing Shard Replay Cycles
**Context**: Today's research revealed that parallel Agent Teams are suffering from "Shard Replay Cycles" where stale fragments in the shared scratchpad are re-ingested by new teammates, leading to reasoning loops.
**Architecture Adjustment**:
*   **Echo-Immune Fragments**: Mandating monotonic, hardware-bound timestamps for all workspace writes in Section 4.
*   **Temporal Sharding**: Introducing mission-phase scoped shards, ensuring that fragments from terminated sub-tasks are cryptographically invalidated for future reasoning steps.
**Security Impact**: Prevents "State Stuttering" and ensures that the swarm maintains a single, forward-moving mission mainline.

### Update: 2026-07-25 - Kernel-Mediated Atomic Synchronization
**Context:** Claude Code v3.2 has encountered performance bottlenecks with user-space scratchpad locking in horizontal teams.
**Architecture Adjustment:**
*   **Atomic Scratchpad Arbiter**: Introducing a kernel-level arbiter in Section 4 that utilizes OS-native file locking (flock/fcntl) for sharded scratchpad fragments.
*   **Mission-Bound Write-Access**: The arbiter now validates the writer's Mission Root token before granting an atomic lock, ensuring that only authorized teammates can mutate the collaborative state.
**Security Impact**: Eliminates teammate race conditions and provides a hardware-attested barrier against "Scratchpad Pollution."
