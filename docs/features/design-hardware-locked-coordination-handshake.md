<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Hardware-Locked Coordination Handshake

**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

The emergence of Identity-Decay Attacks (IDA) and session-hijacking
vulnerabilities in the A2A protocol (CVE-2026-48201) necessitates a transition
from software-only handshakes to hardware-locked sessions. Existing mechanisms
rely on long-lived tokens that can be mimicked or recycled.

The Hardware-Locked Coordination Handshake (HLCH) mandates that every
inter-agent task bidding and state fragment transfer be bound to a unique,
TPM-signed session key. This ensures that even if an agent's behavioral trust
is established, its coordination remain anchored to a non-repudiable hardware
root.

## 2. Goals & Non-Goals

- **Goals:**
  - Implement HLCH v1.0 standard for all inter-agent coordination.
  - Mandate hardware-bound (TPM/Secure Enclave) signatures for coordination session keys.
  - Neutralize nonce-recycling and stylometric mimicry (Identity-Decay).
  - Provide collision-resistant attestation for reasoning fragments.
- **Non-Goals:**
  - Replacing primary user authentication (this is for inter-agent coordination).
  - Managing persistent storage encryption (handled by HAPE/RAMS).

## 3. Critical User Journey (CUJ)

- **User Persona:** Multi-Agent Swarm Orchestrator
- **Primary Goal:** Prevent a specialized subagent from performing an Identity-Decay attack via stylometric mimicry during a long-running session.
- **The Happy Path (Tasks):**
    1. Parent Agent initiates a task delegation request to Teammate A.
    2. Teammate A generates a session-bound coordination key pair.
    3. Teammate A requests a hardware-attestation signature for the session key from the local MRA Provider.
    4. MRA Provider utilizes the TPM to sign the session key and mission-root fragment.
    5. Teammate A completes the HLCH handshake with the parent via the T2T Bridge.
    6. All subsequent state fragments from Teammate A must be signed by the hardware-locked session key.
    7. Parent Agent (via SCI) verifies the hardware signature for every fragment, neutralizing mimicry.

## 4. Design & Architecture

- **System Flow:**
  - Inter-agent request triggers session key generation.
  - MRA Provider issues TPM-signed attestation for the key.
  - HLCH Gateway validates the handshake before establishing the T2T pipe.
  - SCI Interceptor enforces hardware-bound signatures on all transit metadata.

- **APIs / Interfaces:**
  - `hlch.InitiateHandshake(missionRoot, targetAgent) -> HandshakeToken`: Begins the hardware-locked handshake.
  - `hlch.VerifySessionSignature(fragment, signature) -> bool`: Validates fragment integrity against the session key.

- **Data Storage/State:**
  - **Handshake Registry**: Ephemeral, memory-mapped storage for active hardware-locked session metadata.

## 5. Alternatives Considered

- **Software-Only Nonces**: Rejected due to vulnerability to CVE-2026-48201 (nonce-recycling).
- **Per-Fragment TPM Signatures**: Rejected due to the high latency (100ms+) of repeated hardware calls; HLCH utilizes hardware-locked session keys for performance.

## 6. Cross-Cutting Concerns

- **Security (Zero Trust)**: Mandatory hardware root of trust for all coordination. HLCH ensures that session keys are non-exportable from the TPM.
- **Observability**: Integrated with the "Coordination Handshake Debugger" for real-time visualization of attestation events.

## 7. Evolutionary Changelog

- **2026-06-14**: Initial Document Creation. Supporting the transition to Hardware-Locked Coordination.
