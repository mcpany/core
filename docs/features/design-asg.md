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

### Update: 2026-07-18 - Neutralizing Context-Stitching
**Context:** Today's follow-up on CVE-2026-88012 reveals that "Context-Stitching" attacks are most effective in shared teammate scratchpads.
**Architecture Adjustment:**
- Introducing **Cognitive Salt** to the scratchpad write pipeline. ASG will now automatically inject non-functional, semantic noise into scratchpad fragments to prevent bit-level re-assembly of parent context traces.
- Mandating **Reasoning-Aware Redaction** for all scratchpad writes crossing trust tiers (e.g., from an external specialist to a core architect).
**Security Impact:** Prevents low-trust subagents from exfiltrating sensitive mission fragments by "stitching" together shared state.
