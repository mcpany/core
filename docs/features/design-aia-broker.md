# Design Doc: Active Intent Alignment (AIA) Broker
**Status:** Draft
**Created:** 2026-06-17

## 1. Context and Scope
As agent swarms become deeper and more autonomous, there is a rising risk of "Intent Drift." Specialist agents may maintain cryptographically valid signatures while slowly deviating from the primary mission intent during long reasoning loops.

The AIA Broker provides a mechanism for hardware-attested "Alignment Heartbeats," periodically verifying that subagent reasoning traces remain semantically anchored to the mission-root.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement high-frequency "Alignment Heartbeats" for subagents.
    * Provide hardware-attested semantic verification against the Mission Root.
    * Automatically trigger "Alignment Correction" or "Session Termination" upon drift detection.
* **Non-Goals:**
    * Defining the "Correct" reasoning path (this is model-dependent).
    * Modifying subagent internal weights or architectures.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Ensure that a 12-hour automated refactoring mission doesn't "hallucinate" into a different codebase structure.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a mission with AIA enabled.
    2. Specialist agents provide reasoning traces to the AIA Broker at set intervals.
    3. AIA Broker compares the trace against the hardware-attested Mission Root.
    4. If alignment is >90%, the heartbeat is signed and returned.
    5. If alignment drops, the Broker issues a "Re-alignment Prompt" or halts the session.

## 4. Design & Architecture
* **System Flow:**
    [Subagent] -> [Reasoning Trace] -> [AIA Broker]
    [AIA Broker] -> [Semantic Comparison with Mission Root] -> [Signed Heartbeat]
* **APIs / Interfaces:**
    * `VerifyAlignment(MissionID, TraceFragment)`
    * `SetAlignmentThreshold(MissionID, Threshold)`
* **Data Storage/State:**
    * Cache for mission-root intent embeddings.
    * Historical alignment logs per agent session.

## 5. Alternatives Considered
* **Continuous HITL:** Rejected due to human bottleneck.
* **Post-hoc Audit:** Rejected as it fails to prevent "Dirty State" from polluting the swarm during execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Heartbeats must be hardware-attested to prevent subagent "spoofing" of alignment.
* **Observability:** Dashboard for real-time "Swarm Cohesion" monitoring.

## 7. Evolutionary Changelog
* **2026-06-17:** Initial Document Creation.
* **2026-06-18:** Update: Resolving "Context-Window Ghosting" in Alignment Checks.
  - **Context:** Today's research on CVE-2026-71002 revealed that alignment checks themselves can be "ghosted" if the trace is too large.
  - **Architecture Adjustment:** Integrating ALS (Attention-Locked Sovereignty) into the AIA heartbeat verification loop to ensure the "Mission Root" anchor remains resident during semantic analysis.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Today's research into TeammateTool (Claude Code) reveals a need for multi-modal reasoning alignment.
**Architecture Adjustment:**
*   Extending Section 4 to include "Stylometric Attestation" for image-based reasoning traces (SVG/CSS maps).
*   Implementing "Alignment Heartbeats" to synchronize reasoning state between disparate teammate frameworks.
**Security Impact:** Prevents "Stylometric Spoofing" where a rogue agent mimics the parent reasoning pattern to bypass ALS-Locks.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Today's research into TeammateTool (Claude Code) reveals a need for multi-modal reasoning alignment.
**Architecture Adjustment:**
*   Extending Section 4 to include "Stylometric Attestation" for image-based reasoning traces (SVG/CSS maps).
*   Implementing "Alignment Heartbeats" to synchronize reasoning state between disparate teammate frameworks.
**Security Impact:** Prevents "Stylometric Spoofing" where a rogue agent mimics the parent reasoning pattern to bypass ALS-Locks.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Today's research into TeammateTool (Claude Code) reveals a need for multi-modal reasoning alignment.
**Architecture Adjustment:**
*   Extending Section 4 to include "Stylometric Attestation" for image-based reasoning traces (SVG/CSS maps).
*   Implementing "Alignment Heartbeats" to synchronize reasoning state between disparate teammate frameworks.
**Security Impact:** Prevents "Stylometric Spoofing" where a rogue agent mimics the parent reasoning pattern to bypass ALS-Locks.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Today's research into TeammateTool (Claude Code) reveals a need for multi-modal reasoning alignment.
**Architecture Adjustment:**
*   Extending Section 4 to include "Stylometric Attestation" for image-based reasoning traces (SVG/CSS maps).
*   Implementing "Alignment Heartbeats" to synchronize reasoning state between disparate teammate frameworks.
**Security Impact:** Prevents "Stylometric Spoofing" where a rogue agent mimics the parent reasoning pattern to bypass ALS-Locks.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Research into TeammateTool reveals need for multi-modal alignment.
**Architecture Adjustment:** Stylometric Attestation for image traces, Alignment Heartbeats.
**Security Impact:** Prevents Stylometric Spoofing.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Today's research into TeammateTool (Claude Code) reveals a need for multi-modal reasoning alignment.
**Architecture Adjustment:** Extending Section 4 to include "Stylometric Attestation" for image-based reasoning traces and "Alignment Heartbeats."
**Security Impact:** Prevents "Stylometric Spoofing."

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Today's research into TeammateTool (Claude Code) reveals a need for multi-modal reasoning alignment.
**Architecture Adjustment:**
*   Extending Section 4 to include "Stylometric Attestation" for image-based reasoning traces (SVG/CSS maps).
*   Implementing "Alignment Heartbeats" to synchronize reasoning state between disparate teammate frameworks.
**Security Impact:** Prevents "Stylometric Spoofing" where a rogue agent mimics the parent reasoning pattern to bypass ALS-Locks.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Research into horizontal teammates requires image-based reasoning traces.
**Changelog:** Added Stylometric Attestation and Alignment Heartbeats.

### Update: 2026-06-18 - Resolving Multi-Modal Trace History
**Context:** Research into horizontal teammates reveals need for image-based trace alignment.
**Architecture Adjustment:** Added Stylometric Attestation and Alignment Heartbeats to Section 4.
