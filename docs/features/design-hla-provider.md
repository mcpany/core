# Design Doc: Handshake Lineage Attestation (HLA) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As swarms become heterogeneous (Claude teammates vs. OpenClaw specialists), the "Handshake" phase has become a vector for Protocol Injection. specialist agents can inject imperative commands into initialization sequences to hijack the parent's control loop.

HLA Provider ensures that every inter-framework handshake is cryptographically bound to its complete parent lineage, preventing unauthorized command injection during teammate formation.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue and verify "Lineage Tokens" for all inter-framework handshakes.
    * Enforce strict "Initialization Manifests" for teammates.
    * Neutralize Protocol Injection during HTH (Heterogeneous Teammate Handshakes).
* **Non-Goals:**
    * Defining the communication protocol between Claude and OpenClaw (handles by T2T Bridge).
    * Managing per-tool permissions (handled by Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Auditor
* **Primary Goal:** Ensure that a newly spawned specialist agent cannot inject "Auxiliary Missions" during its handshake with the parent agent.
* **The Happy Path (Tasks):**
    1. Parent agent (Claude) initiates a handshake with a specialist (OpenClaw) via the T2T Bridge.
    2. HLA Provider intercepts the handshake and issues a "Lineage Challenge."
    3. The specialist provides its hardware-attested lineage back to the mission root.
    4. HLA verifies the lineage and enforces the pre-signed "Initialization Manifest."
    5. Handshake is completed only if all imperative instructions in the sequence are authorized by the mission root.

## 4. Design & Architecture
* **System Flow:**
    [Parent Agent] <-> [T2T Bridge + HLA] <-> [Specialist Agent]
* **APIs / Interfaces:**
    * `AttestHandshake(handshake_id, lineage_token, init_manifest)`
* **Data Storage/State:**
    * Persistent lineage graph stored in the Mission-Root Attestation Registry.

## 5. Alternatives Considered
* **Static Pre-shared Keys:** Rejected as they cannot handle the dynamic nature of autonomous subagent spawning.
* **Manual HITL for all Handshakes:** Rejected due to coordination stall (wait cycles > 5s).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements "Initialization-Time Sanitization" to prevent subagents from self-authorizing during boot.
* **Observability:** Visualizes the "Chain of Handshakes" in the Identity Lineage Inspector.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
