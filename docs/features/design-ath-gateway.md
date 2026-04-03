# Design Doc: Atomic Teammate Handshake (ATH) Gateway
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of horizontal "Agent Teams" (e.g., Claude Code), coordination has moved from single-agent sessions to multi-agent meshes. This shift has introduced "Teammate Impersonation" vulnerabilities, where a compromised specialist subagent can claim tasks or inject state into a teammate's mailbox by spoofing its identity. The ATH Gateway provides a mandatory, hardware-attested handshake that teammates must complete before they can coordinate, ensuring non-repudiable identity within the mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Mandate hardware-attested identity exchange for all inter-teammate task claiming and delegation.
    * Neutralize teammate impersonation in horizontal meshes.
    * Provide a cryptographically bound "Handshake Token" for every coordination event.
    * Support cross-framework teammate verification (Claude, OpenClaw, AutoGen).
* **Non-Goals:**
    * Implementation of full-mesh encryption (this is handled by the AMT Broker).
    * Providing long-term identity persistence (handled by the MRIA Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialist Subagent in a Horizontal Swarm
* **Primary Goal:** Securely claim a "Database Optimization" task from the shared team mailbox without being shadowed by a rogue specialist.
* **The Happy Path (Tasks):**
    1. The specialist agent identifies a relevant task in the shared mailbox.
    2. The agent attempts to "claim" the task via the MCP Any coordination endpoint.
    3. The ATH Gateway intercepts the claim request and challenges the agent for hardware attestation.
    4. The agent provides a TPM-signed "Identity Fragment" and its mission-root lineage.
    5. The ATH Gateway verifies the signature and checks the agent's authorized role in the mission manifest.
    6. Upon success, the ATH Gateway issues a unique "Atomic Handshake Token" bound to that specific task claim.
    7. The mailbox marks the task as claimed by the verified agent, and coordination proceeds.

## 4. Design & Architecture
* **System Flow:**
    `[Teammate Agent] --(Task Claim + TPM Sig)--> [ATH Gateway] --(Verified Handshake Token)--> [Shared Mailbox]`
* **APIs / Interfaces:**
    * `POST /mesh/handshake/atomic`: Performs a hardware-attested handshake for a specific coordination action.
    * `GET /mesh/handshake/verify`: Validates an Atomic Handshake Token for mailbox or state access.
* **Data Storage/State:**
    * Handshake tokens are ephemeral and task-bound.
    * Hardware-bound public keys for teammates are cached in the `Mesh Peer Registry`.

## 5. Alternatives Considered
* **Bearer-Token Auth:** Rejected as it is vulnerable to token theft/leakage via local environment probing.
* **IP-based Trust:** Rejected due to lack of granularity and vulnerability to subagent process hopping.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** No coordination action is allowed without a valid, hardware-attested handshake.
* **Observability:** Handshake successes, failures, and "Spoofing Detected" events are logged to the `Swarm Integrity Monitor`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
