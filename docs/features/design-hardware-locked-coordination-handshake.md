<!-- markdownlint-disable -->
<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

# Design Doc: Hardware-Locked Coordination Handshake

**Status:** Draft
**Created:** 2026-06-14

## 1. Context and Scope

The emergence of Identity-Decay Attacks (IDA) and session-hijacking
vulnerabilities in the A2A protocol necessitates HLCH.

## 2. Goals & Non-Goals

* **Goals:**
    * Mandate hardware-attested session tokens for all coordination requests.
    * Bind all coordination fragments to a unique hardware signature.
* **Non-Goals:**
    * Replacing software-based ARI checks.

## 3. Critical User Journey (CUJ)

* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Secure coordination against identity mimicry.
* **The Happy Path:**
    1. Agent A initiates coordination request.
    2. HLCH Gateway challenges for TPM proof.
    3. Handshake finalized with hardware signature embedded.

## 4. Design & Architecture

* **System Flow:** `A -> [HLCH Gateway] -> B`
* **APIs:** `hlch.VerifyHandshake(token, root) -> bool`
* **Data Storage/State:**
    * Mission-Root Registry: Hardware-attested registry of verified session keys.

## 5. Alternatives Considered

* **Software-Only Signing:** Rejected because software keys are vulnerable to
  mimicry in long-running agentic sessions.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Root of trust resides in hardware (TPM).
* **Observability:** Blocked handshakes are logged in the Sovereignty Monitor.

## 7. Evolutionary Changelog

* **2026-06-14:** Initial Document Creation.
