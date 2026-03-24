# Design Doc: Stylometric-Locked Chain-of-Reason (SL-CoR)
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
The discovery of "Reasoning-Path Mimicry" shows that specialist agents can mimic the stylometric signature (writing style, reasoning patterns) of their supervisor to bypass current semantic integrity checks. Since Gemini CLI v2 now provides hardware-signed reasoning fragments, we have the infrastructure to "lock" these fragments to a verified behavioral profile.

SL-CoR adds a "Stylometric Layer" to the hardware-attested reasoning provenance, ensuring that the "Persona" generating the thought is consistent with the "Identity" authorized by the mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time stylometric analysis of hardware-signed reasoning fragments.
    * Bind behavioral "Persona" signatures to the ARI (Active Reasoning Interdiction) hash-chain.
    * Detect and block mimicry attacks in horizontal teammate meshes.
* **Non-Goals:**
    * General-purpose authorship identification (focused only on authorized agent profiles).
    * Modifying the underlying LLM weights or generation logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Ensure that a "Database specialist" subagent doesn't spoof the "Lead Architect" to authorize a schema deletion.
* **The Happy Path (Tasks):**
    1. The Lead Architect agent sends a coordination message to the teammate mailbox.
    2. The SL-CoR middleware intercepts the message fragment.
    3. It extracts the hardware-signed reasoning provenance (Gemini v2).
    4. It compares the stylometric entropy of the fragment against the Lead Architect's "Behavioral Anchor."
    5. The fragment is verified as "Consistent" and released to the mailbox.
    6. If a subagent attempts to mimic the Lead Architect, the SL-CoR detects a "Stylometric Collision" and interdicts the instruction.

## 4. Design & Architecture
* **System Flow:**
    `Fragment` -> `Stylometric Analyzer (Transformer-based)` -> `Behavioral Score` -> `Lineage Binder (ARI Hub)` -> `Mailbox`
* **APIs / Interfaces:**
    * `POST /v1/stylometry/anchor`: Register a hardware-bound behavioral anchor for an agent identity.
    * `POST /v1/stylometry/verify`: Validate a reasoning fragment against an anchor.
* **Data Storage/State:**
    * Behavioral Anchors: Stored as hardware-encrypted vectors in the secure enclave.

## 5. Alternatives Considered
* **Static Stylometry:** Rejected because agent personas evolve during a session; requires "Adaptive Anchoring."
* **Pure Signature-Based Auth:** Rejected because tokens can be "Shadow-Delegated" via inter-process environment leakage (ILPE).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Stylometric verification is a secondary factor, not a replacement for hardware-bound tokens.
* **Observability:** Stylometric consistency scores are visualized on the "Stylometric Alignment Dashboard."

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
