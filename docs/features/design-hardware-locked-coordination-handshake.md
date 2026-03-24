<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

<!-- markdownlint-disable -->

# Design Doc: Hardware-Locked Coordination Handshake

**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

The emergence of Identity-Decay Attacks (IDA) and session-hijacking
vulnerabilities in the A2A protocol necessitates a transition from
software-only handshakes to hardware-locked sessions. Existing mechanisms
like ARI and HAAL provide semantic and attention-level security, but HLCH
establishes a hardware-bound root of trust for the coordination session itself.

## 2. Goals & Non-Goals

* **Goals:**
    * Mandate hardware-attested (TPM/Secure Enclave) session tokens for all
      inter-agent coordination requests.
    * Bind all coordination fragments (metadata, task bidding) to a unique
      hardware signature.
    * Neutralize Identity-Decay Attacks (IDA) by ensuring subagents cannot mimic
      the parent session's hardware lineage.
* **Non-Goals:**
    * Replacing software-based ARI checks (HLCH is an additional layer).
    * Enforcing local port isolation (handled by BSH Gateway).

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Secure inter-agent coordination against identity mimicry.
* **The Happy Path (Tasks):**
    1. Agent A initiates a coordination request to Agent B.
    2. The HLCH Gateway intercepts the request and challenges for a hardware-
       attested session token.
    3. Agent A provides a TPM-signed token bound to the current mission root.
    4. The HLCH Gateway verifies the token's lineage and hardware signature.
    5. The handshake is finalized, and coordination proceeds with the hardware
       signature embedded in all subsequent fragments.

## 4. Design & Architecture

* **System Flow:**
  A -> [HLCH Gateway] -> B
  Gateway validates TPM signature against the mission-root registry.
* **APIs / Interfaces:**
    * hlch.VerifyHandshake(sessionToken, missionRoot) -> bool
    * hlch.SignFragment(fragment, sessionKey) -> signedFragment
* **Data Storage/State:**
    * Mission-Root Registry: A hardware-attested registry of verified session
      keys.

## 5. Alternatives Considered

* **Software-Only Signing:** Rejected because software keys are vulnerable to
  exfiltration and mimicry in long-running agentic sessions.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Root of trust resides in hardware (TPM).
* **Observability:** Blocked handshakes are logged in the Sovereignty Monitor.

## 7. Evolutionary Changelog

* **2026-06-14:** Initial Document Creation.
