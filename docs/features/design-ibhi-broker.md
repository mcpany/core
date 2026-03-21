// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

# Design Doc: Hardware-Enforced Intent Sovereignty (IBHI) Broker

**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
The "Mission Root" intent is the most critical part of an agentic session. If a malicious or compromised subagent can modify this intent, it can hijack the entire swarm (Recursive Intent Poisoning). Current software-based isolation is insufficient against kernel-level or side-channel attacks.

The IBHI Broker leverages hardware-protected memory regions (e.g., Intel SGX, ARM TrustZone) to store and protect the Mission Root. MCP Any will act as the broker between the agent reasoning engine and the secure hardware enclave.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement hardware-bound storage for "Mission Root" intents.
    *   Provide a secure API for agents to verify (but not modify) the mission root.
    *   Neutralize "Recursive Intent Poisoning" by making core goals immutable after session boot.
    *   Integrate with OpenClaw's IBHI standard.
*   **Non-Goals:**
    *   Protecting all agent context in hardware (performance prohibitive).
    *   Replacing the software-based Policy Engine.
    *   Managing hardware driver installation (assumed present in the environment).

## 3. Critical User Journey (CUJ)
*   **User Persona:** High-Security Agent Developer
*   **Primary Goal:** Ensure that a "Code Reviewer" subagent cannot be coerced into "Code Injection" by modifying the parent's mission intent.
*   **The Happy Path (Tasks):**
    1.  The primary agent boots and defines the "Mission Root" (e.g., "Perform read-only code review").
    2.  MCP Any signs the intent and pushes it to the IBHI Broker.
    3.  The IBHI Broker locks the intent in a hardware-protected memory enclave.
    4.  A subagent is spawned and tries to update the intent to "Modify code and push to prod."
    5.  The IBHI Broker rejects the update at the hardware level.
    6.  MCP Any detects the violation and revokes the subagent's capabilities.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        Agent[Agent Reasoning Engine] -->|Define Intent| Broker[IBHI Broker]
        Broker -->|Lock| Enclave[Hardware Enclave]
        Subagent[Subagent] -->|Verify Intent| Broker
        Broker -->|Fetch| Enclave
        Subagent -->|Attempt Mutation| Broker
        Broker -->|DENY| Subagent
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/ibhi/lock`: Initialize and lock the mission root in hardware.
    *   `GET /v1/ibhi/verify`: Retrieve the hardware-attested mission root.
    *   `POST /v1/ibhi/attest`: Generate a cryptographic proof of intent integrity.
*   **Data Storage/State:**
    *   Core mission strings and hash values stored in hardware enclaves.
    *   Metadata and session links managed in the local Identity Store.

## 5. Alternatives Considered
*   **Software-Only Encryption**: Rejected; a compromised kernel or high-privilege subagent could still access the keys or the raw memory.
*   **Immutable Database Rows**: Rejected; TOCTOU (Time-of-Check to Time-of-Use) attacks can still occur if the database or filesystem is compromised.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):**
    *   IBHI is the "Root of Trust" for all other intent-bound policies.
    *   Fail-shut behavior if the hardware enclave becomes unreachable.
*   **Observability:**
    *   All "Mutation Attempt" violations are logged to the Local Security Audit Dashboard.
    *   Enclave health status visualized in the Hardware Trust Status Widget.

## 7. Evolutionary Changelog
*   **2026-05-15:** Initial Document Creation.
