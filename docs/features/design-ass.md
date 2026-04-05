# Design Doc: Ambient State Sanitizer (ASS)
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
As AI agents move from single-threaded sessions to horizontal teammate meshes (e.g., Claude Code Agent Teams), the risk of side-channel state injection increases. Today's research revealed that un-sanitized environment metadata leaked between teammates can be weaponized by rogue subagents to influence the reasoning of siblings without direct mailbox interaction. MCP Any needs to provide a mandatory sanitization layer for all "Ambient State" shared between teammates to maintain mission-root sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time, high-entropy semantic analysis of environment metadata crossing teammate boundaries.
    * Redact or normalize sensitive host-level variables (e.g., PWD, PATH, SHELL) that are not explicitly authorized by the mission root.
    * Provide hardware-attested environment snapshots per teammate execution context.
* **Non-Goals:**
    * Sanitizing the primary inter-agent mailbox messages (handled by Mailbox Integrity Middleware).
    * Restricting legitimate tool-driven environment modifications.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Prevent a specialized "Search Agent" from leaking unauthorized host filesystem paths to a "Code Writing Agent" via ambient environment metadata.
* **The Happy Path (Tasks):**
    1. The architect defines an ASS policy that restricts "Ambient Path Leaks".
    2. Teammate A (Search) executes a tool and generates environment metadata.
    3. The ASS middleware intercepts the metadata before it is synchronized to the shared teammate mesh.
    4. ASS identifies unauthorized path fragments and redacts them.
    5. Teammate B (Code Writer) receives a sanitized view of the environment, anchored to the mission root.

## 4. Design & Architecture
* **System Flow:**
  [Teammate] -> [Metadata Capture] -> [ASS Middleware (Rego/LLM Sanitizer)] -> [Shared Mesh State]
* **APIs / Interfaces:**
    * `mcpany.sanitizer.v1.AmbientStateSanitizer`
    * Hook: `preSyncAmbientState(metadata: EnvironmentObject)`
* **Data Storage/State:**
    * Ephemeral, hardware-attested environment hashes stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Implicit Trust**: Rejected due to the emergence of side-channel injection exploits.
* **Full Environment Isolation**: Rejected as it breaks necessary coordination where siblings need to know legitimate project-local context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All environment metadata is treated as untrusted input.
* **Observability:** Sanitization events are logged with original vs. redacted diffs in the Audit Log.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.

### Update: 2026-07-25 - Implementing Atomic State Segregation
**Context:** Today's market sync revealed a "Memory Bleed" vulnerability in Claude Code horizontal swarms where specialists could probe sibling scratchpads.
**Architecture Adjustment:** * Transitioning from logical environment isolation to hardware-locked Team-Bound Memory Shards (TBMS).
* Mandating hardware-enclave (TPM/SEP) checks for all shared teammate workspace access.
**Security Impact:** Physically neutralizes lateral state movement and context-stitching exfiltration by rogue subagents.
