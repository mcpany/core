# Design Doc: Continuous Sandbox Policy Verifier
**Status:** Draft
**Created:** 2026-04-19

## 1. Context and Scope
Point-in-time attestation at boot is no longer sufficient for high-security enterprise environments. Agents operating in deep swarms can be subjected to "Delayed Payload" attacks or "Environment Drift" that compromises the sandbox mid-session. The Continuous Sandbox Policy Verifier provides real-time, "Always-On" validation of sandbox boundaries against the resident security policy.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time validation of sandbox resource access (filesystem, network, compute) against active security policies.
    * Detect and block unauthorized "Boundary Expansion" attempts that bypass standard intent-scoping.
    * Provide a continuous "Policy Compliance Signal" for the A2A Messaging Hub.
    * Support hardware-locked policy enforcement via integration with the Hardware-Locked Policy Enforcer.
* **Non-Goals:**
    * Managing the initial sandbox creation (handled by the Pre-Flight Sandbox Validator).
    * Defining the security policies themselves (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Auditor
* **Primary Goal:** Ensure that a compromised subagent cannot escalate its privileges to access host-level environment variables, even if it successfully manipulates its local configuration.
* **The Happy Path (Tasks):**
    1. Agent is initialized with a cryptographically bound resident security policy.
    2. Continuous Sandbox Policy Verifier starts a background monitoring thread.
    3. Every 15 seconds (randomized), the Verifier checks the current sandbox boundaries against the hardware-locked policy.
    4. If a subagent attempts to mount an unauthorized directory, the Verifier detects the drift and triggers an immediate "Mission Halt."
    5. The Governance Transparency Log records the policy violation and the associated hardware-bound attestation state.

## 4. Design & Architecture
* **System Flow:**
    `[Resident Policy] -> [Continuous Verifier] <-> [Sandbox Runtime] -> [Compliance Signal]`
* **APIs / Interfaces:**
    * `PolicyVerifier`: `StartMonitoring(sessionID)`, `CheckCompliance() (bool, error)`
    * `ComplianceSignal`: A signed heart-beat token indicating the current policy alignment status.
* **Data Storage/State:**
    * Leverages the "Hardware-Locked Policy Enforcer" to store and retrieve immutable policy definitions.

## 5. Alternatives Considered
* **Kernel-Level EBPF Monitoring**: Considered for high performance but rejected for initial draft due to complexity and cross-platform compatibility requirements. We will start with user-space randomized sweeps.
* **Reactive-Only Validation**: Rejected as it allows a "window of vulnerability" between policy drift and detection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Verifier itself must run in a "Detached Sandbox" to prevent it from being disabled by a compromised agent.
* **Observability:** Real-time compliance status is exported to the "Continuous Sandbox Policy Monitor" UI.

## 7. Evolutionary Changelog
* **2026-04-19:** Initial Document Creation.
