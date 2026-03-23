<!-- markdownlint-disable -->
<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Hardware-Locked Coordination Handshake

**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

The emergence of Identity-Decay Attacks (IDA) and session-hijacking vulnerabilities in the A2A protocol necessitates a transition from software-only handshakes to hardware-locked sessions.

## 2. Goals & Non-Goals

* **Goals:**
  * Mandate hardware-attested (TPM/Secure Enclave) session tokens for all inter-agent coordination.
  * Neutralize Identity-Decay Attacks (IDA) by binding coordination to unique hardware signatures.
* **Non-Goals:**
  * Replacing software-based ARI checks (this is an additional layer).

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Secure inter-agent coordination against identity mimicry.
* **The Happy Path (Tasks):**
  1. Agent A initiates a coordination request to Agent B.
  2. The HLCH Gateway intercepts the request and challenges for a hardware-attested session token.
  3. Agent A provides a TPM-signed token bound to the current mission root.
  4. The HLCH Gateway verifies the token and allows the handshake to proceed.

## 4. Design & Architecture

* **System Flow:**
  - `A -> [HLCH Gateway] -> B`
  - Gateway validates TPM signature against the mission-root registry.
* **APIs / Interfaces:**
  - `hlch.VerifyHandshake(sessionToken) -> bool`
* **Data Storage/State:**
  - Hardware-Attested Mission-Root Registry (shared state).

## 5. Alternatives Considered

* **Software-Only Signing:** Rejected because it is vulnerable to exfiltration and mimicry in long-running agentic sessions.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Root of trust resides in hardware (TPM).
* **Observability:** Blocked handshakes are logged in the "Sovereignty Monitor."

## 7. Evolutionary Changelog

* **2026-06-14:** Initial Document Creation.
