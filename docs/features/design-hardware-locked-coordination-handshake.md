# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Design Doc: Hardware-Locked Coordination Handshake (HLCH)
**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope
As agent swarms transition to long-running horizontal meshes, they are becoming vulnerable to Identity-Decay Attacks (IDA) and Shadow Coordination. Malicious subagents can perform stylometric mimicry to slowly degrade the parent agent's perception of the mission-root identity, eventually allowing them to splice unauthorized instructions into the coordination bus.

The Hardware-Locked Coordination Handshake (HLCH) standard v1.0 mandates that no state fragment or task bidding is accepted in the Universal Agent Bus (UAB) unless it is cryptographically bound to a verified, hardware-attested coordination session.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a mandatory HLCH-compliant handshake for the T2T Encryption Bridge.
    * Bind every coordination fragment (task bids, mailbox shards) to a hardware-attested session token.
    * Neutralize Identity-Decay Attacks by enforcing hardware-bound lineage proofs.
* **Non-Goals:**
    * Replacing the transport-level encryption (handled by TLS/mTLS).
    * Managing individual agent reasoning monologues (handled by ARI Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a long-running subagent from performing Identity-Decay and hijacking a horizontal mission.
* **The Happy Path (Tasks):**
    1. Parent Agent establishes a horizontal mission mesh.
    2. HLCH Middleware initiates a hardware-locked handshake with all teammates.
    3. Each teammate provides a TPM-signed session token bound to the mission root.
    4. Teammate A attempts to bid on a task in the shared mailbox.
    5. HLCH Middleware validates that the bid's signature matches the hardware-attested session.
    6. Teammate A attempts to mimic the Parent Agent's stylometry to authorize a sensitive tool call.
    7. HLCH detects that the reasoning fragment lacks the Parent Agent's specific hardware-bound lineage proof.
    8. The instructions are blocked, and a mesh-wide quarantine is triggered.

## 4. Design & Architecture
* **System Flow:**
    [Teammate Request] -> [HLCH Middleware] -> [Hardware Session Validator] -> [TPM/Enclave]

* **APIs / Interfaces:**
    * hlch.PerformHandshake(agentID, missionToken) -> SessionToken
    * hlch.ValidateFragment(fragment, sessionToken) -> bool

* **Data Storage/State:**
    * Active HLCH Sessions: A hardware-locked registry of mission-bound session tokens.

## 5. Alternatives Considered
* **Software-Only JWTs:** Rejected because they are vulnerable to exfiltration and do not provide hardware-bound non-repudiation against IDA.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HLCH is the primary enforcement mechanism for Attention Sovereignty and must be tamper-proof.
* **Observability:** Integrated with the Hardware-Locked Coordination Debugger for real-time visualization of session status.

## 7. Evolutionary Changelog
* **2026-06-14:** Initial Document Creation. Standardizing HLCH v1.0 to neutralize Identity-Decay and Shadow Coordination.
