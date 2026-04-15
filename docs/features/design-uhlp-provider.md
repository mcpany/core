# Design Doc: Universal Hardware-Lease Provider (UHLP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of framework-specific hardware leases (e.g., Claude Code's MBHL), a new interoperability gap has emerged. Agents from different frameworks cannot easily share or verify hardware-locked capabilities. The Universal Hardware-Lease Provider (UHLP) aims to act as the authoritative broker for "Lease Interoperability," translating proprietary TPM-signed leases into a unified, mission-bound capability manifest that can be consumed by any framework connected to MCP Any.

## 2. Goals & Non-Goals
* **Goals:**
    * Translate proprietary hardware leases (Claude Code, Gemini) into a framework-neutral manifest.
    * Provide a centralized registry for all active hardware-locked capability leases.
    * Enforce automated lease expiration upon mission-root task completion.
    * Facilitate cross-framework lease verification via hardware-attested handshakes.
* **Non-Goals:**
    * Replacing the underlying hardware attestation mechanisms (TPM/SEP).
    * Defining a single hardware security standard for all devices.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Allow an OpenClaw specialist to inherit a "Filesystem Write" lease originally issued to a Claude Code supervisor.
* **The Happy Path (Tasks):**
    1. Claude Code supervisor requests a "Filesystem Write" capability for a specific task.
    2. UHLP issues a TPM-signed MBHL-compliant lease.
    3. The supervisor delegates a sub-task to an OpenClaw specialist.
    4. The OpenClaw specialist requests access to the "Filesystem Write" tool.
    5. UHLP verifies the specialist's lineage and mission-bound authority.
    6. UHLP "translates" the Claude Code lease into an OpenClaw-compliant intent token.
    7. The OpenClaw specialist executes the tool call securely.

## 4. Design & Architecture
* **System Flow:**
    `[Framework A Lease] -> [UHLP Translator] -> [Universal Manifest] -> [UHLP Broker] -> [Framework B Agent]`
* **APIs / Interfaces:**
    * `uhlp.TranslateLease(sourceLease, targetFramework) -> UnifiedToken`: Converts proprietary leases.
    * `uhlp.VerifyLease(unifiedToken, missionID) -> bool`: Validates lease integrity and mission-binding.
    * `uhlp.RevokeLease(leaseID) -> error`: Forcefully terminates a capability lease.
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed storage within the MCP Any server tracking active leases and their hardware fingerprints.
    * **Translation Mapping:** A set of pluggable adapters for different framework lease formats.

## 5. Alternatives Considered
* **Direct Framework Bridging:** Rejected because it leads to N^2 complexity as new frameworks emerge.
* **Legacy API Keys:** Rejected due to lack of hardware attestation and mission-binding.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** UHLP mandates hardware-attested identity handshakes for all translation and verification requests.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active and expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.

    ### Update: 2026-07-26 - Multi-Hop Lease Propagation (MHLP)
    **Context:** Today's market sync revealed Claude Code v3.3.0 supporting multi-hop lease relays, signaling a need for depth-invariant trust.
    **Architecture Adjustment:**
    *   Introducing the **Universal Lease Relay (ULR)** middleware in Section 4.
    *   UHLP will now facilitate the propagation of TPM-signed leases across 5+ levels of sub-delegation.
    *   Each hop includes a "Chain-of-Custody" metadata block that is verified against the mission-root manifest.
    **Security Impact:** Maintains absolute sovereignty in deep swarms without the 100ms+ latency of re-attesting at every delegation hop.
