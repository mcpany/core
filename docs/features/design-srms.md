# Design Doc: Stitch-Resistant Memory Segmentation (SRMS)
**Status:** Draft
**Created:** 2026-07-18

## 1. Context and Scope
The disclosure of CVE-2026-88012 (Context-Stitching) reveals a critical vulnerability in how multi-agent swarms handle shared state. Malicious subagents can observe fragmented outputs in shared mailboxes or scratchpads and "stitch" them together to reconstruct sensitive parent context or mission-root intents. SRMS provides advanced memory protection for the Universal Agent Bus, utilizing "Cognitive Salt" and reasoning-aware redaction to ensure that state fragments cannot be re-composed into parent traces.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement "Cognitive Salt" (non-deterministic semantic noise) for all shared state fragments.
    * Provide reasoning-aware redaction of mission-root intents before state commitment.
    * Neutralize "Context-Stitching" exfiltration by ensuring fragments lack semantic continuity.
    * Enforce cryptographic boundaries between teammate-local shards and the shared mission blackboard.
* **Non-Goals:**
    * Blocking legitimate data sharing between authorized teammates.
    * Managing the transport-layer encryption (handled by T2T/LOWA).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Trust Swarm Architect
* **Primary Goal:** Prevent a specialized "UI Component Agent" from reconstructing the system's "Database Schema Intent" by observing fragmented writes in the shared scratchpad.
* **The Happy Path (Tasks):**
    1. The architect enables SRMS for the mission.
    2. The "Schema Agent" writes a fragment to the scratchpad.
    3. SRMS intercepts the write and applies "Cognitive Salt" (injecting semantically valid but non-critical variations).
    4. The "UI Agent" reads the fragment but cannot correlate it with previous fragments to build a coherent schema map.
    5. The mission-root retains the unsalted "Truth" in its private RAMS shard.

## 4. Design & Architecture
* **System Flow:**
  [Fragment Write] -> [RAR Engine (Intent Redaction)] -> [SRMS Salter] -> [Shared Workspace]
* **APIs / Interfaces:**
    * `mcpany.srms.v1.SegmentationProvider`
    * `ApplyCognitiveSalt(fragment_id, content)`
* **Data Storage/State:**
    * Hardware-attested salt-indices; Versioned shards in the RAMS-compliant Blackboard.

## 5. Alternatives Considered
* **Binary Encryption per Fragment**: Rejected because it prevents legitimate collaboration between agents who need to see the *content* but not the *context*.
* **Strict Read-Only Isolation**: Rejected as it breaks the "Agent Team" collaboration model.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All shared state is treated as a potential exfiltration side-channel.
* **Observability:** "Stitch-entropy" scores are tracked to alert on potential mimicry or correlation attempts.

## 7. Evolutionary Changelog
* **2026-07-18:** Initial Document Creation.

### Update: 2026-07-25 - Mirror-Intent Validation for Scratchpads
**Context:** Today's research into "Reflection Quorums" and teammate coordination confirms that shared workspaces are the primary site for "Reflection Drift."
**Architecture Adjustment:**
* Integrating the **Teammate Mirror-Intent Arbiter (TMIA)** into the SRMS write-pipeline in Section 4.
* Mandatory "Mirror-Intent" check for all multi-agent scratchpad writes to ensure semantic alignment before salting.
**Security Impact:** Prevents "Coordinated Drift" where multiple teammates drift away from mission-root constraints in the shared workspace.
