# Design Doc: Universal Multimodal Memory Bus (UMMB)
**Status:** Draft
**Created:** 2026-07-01

## 1. Context and Scope
As AI agent ecosystems transition from linear executions to heterogeneous swarms (e.g., Claude Code, OpenClaw, AutoGen), state synchronization has become a critical bottleneck. Frameworks currently rely on isolated, flat context windows or basic shared stores that lack cryptographic lineage. This leads to **Context Fragmentation** and **Intent Drift** during multi-framework handoffs. Furthermore, multimodal inputs (SVG, Audio, Video) often bypass standard reasoning-path validation, allowing **Context Smuggling** via non-textual metadata.

The Universal Multimodal Memory Bus (UMMB) implements a hardware-attested, intent-pinned memory bus for synchronizing state across disparate frameworks while performing real-time multimodal trace sanitization.

## 2. Goals & Non-Goals
* **Goals:**
    * Synchronize state across heterogeneous agent frameworks with cryptographic lineage.
    * Enforce "Mission-Root Intent" pinning for all memory fragments traversing the bus.
    * Provide real-time, hardware-attested sanitization of multimodal traces (SVG, CSS, Audio).
    * Ensure that non-textual inputs cannot be used to evict critical instructions from the attention window.
* **Non-Goals:**
    * Replacing framework-specific internal short-term memory.
    * Generating multimodal content (UMMB only validates and synchronizes).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Orchestrator
* **Primary Goal:** Pass context from a Claude Code agent to an OpenClaw specialist without losing the primary mission intent or introducing multimodal exploits.
* **The Happy Path (Tasks):**
    1. Agent A (Claude) pushes its task state and reasoning monologue to the UMMB.
    2. The UMMB serializes the state, anchoring it to the hardware-attested mission root.
    3. The Multimodal Trace Validator (MTV) scans associated SVG traces for hidden imperative instructions.
    4. The specialist Agent B (OpenClaw) retrieves the sanitized, intent-pinned state.
    5. Agent B resumes the task, confident that the context is authentic and secure.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Agent Framework A] --> B[Intent-Pinned Serializer]
        B --> C[UMMB Memory Bus]
        C --> D[Multimodal Trace Validator]
        D --> E[Hardware-Attested Shards]
        E --> F[Agent Framework B]
    ```
* **APIs / Interfaces:**
    * `POST /v1/memory/bus/push`: Pushes intent-anchored state fragments.
    * `GET /v1/memory/bus/pull`: Retrieves sanitized state shards for a mission.
* **Data Storage/State:**
    * Memory fragments are stored in **Hardware-Attested Memory Shards (HAMS)** within the Blackboard, requiring TPM-bound proofs for access.

## 5. Alternatives Considered
* **Framework-Specific Serializers:** Rejected as it leads to "context amnesia" when passing state between incompatible agent models.
* **Text-Only Memory:** Rejected due to the increasing reliance on visual and audio traces in modern agency.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** UMMB mandates that every fragment carry a signed lineage back to the root intent, neutralizing "Memory Smearing".
* **Observability:** Tracks "Multimodal Sanitization" events and attention-pinning health in the visual dashboard.

## 7. Evolutionary Changelog
* **2026-07-01:** Initial Document Creation.
