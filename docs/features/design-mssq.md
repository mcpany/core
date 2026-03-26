# Design Doc: Machine-Speed Swarm Quarantine (MSSQ)
**Status:** Draft
**Created:** 2026-06-01

## 1. Context and Scope
The 2026 Armis Cyberwarfare Report highlights a critical collapse in Mean Time to Compromise (MTTC), which has moved from hours to seconds due to "Agentic Swarms." Traditional human-in-the-loop security operations cannot respond at this "Machine Speed." MSSQ is designed as an autonomous interdiction layer for the CSAD Hub, capable of isolating compromised agent sub-meshes in sub-milliseconds to prevent lateral movement and zero-day propagation.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement autonomous, sub-millisecond revocation of agent capabilities.
    * Provide "Mission-Wide" quarantine based on CSAD anomaly triggers.
    * Support hardware-attested lockdown signals that cannot be bypassed by compromised agents.
    * Ensure quarantine states are propagated across all framework-neutral UAB adapters.
* **Non-Goals:**
    * Replacing long-term forensic analysis (MSSQ is for immediate containment).
    * Manually reviewing every quarantine event (designed for autonomous action).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Security Mesh (ASM)
* **Primary Goal:** Automatically isolate a "Hivenet" attack within a teammate swarm before it can exfiltrate sensitive mission fragments.
* **The Happy Path (Tasks):**
    1. CSAD Hub detects a "Coordinated Discovery Probe" signature across three parallel teammates.
    2. CSAD issues an MSSQ trigger to the gateway.
    3. MSSQ immediately revokes all A2A and Tool-Call capability tokens for the specific `mission_id`.
    4. The T2T Encryption Bridge terminates all active pipes for the affected sub-mesh.
    5. The UI surfaces a "Machine-Speed Quarantine" alert with a cryptographically signed justification.
    6. The "Mission Root" is preserved in a "Safe-Hold" state for human recovery.

## 4. Design & Architecture
* **System Flow:**
    `[CSAD Pattern Matcher] -> (Trigger) -> [MSSQ Controller] -> (Revoke) -> [Identity & Transport Layers]`
* **APIs / Interfaces:**
    * `mssq.v1.TriggerQuarantine(mission_id, reason_code)`: Internal high-speed trigger.
    * `mssq.v1.GetQuarantineStatus(mission_id)`: Check if a mission is currently isolated.
* **Data Storage/State:**
    * High-priority "Kill Switch" bitset in-memory for sub-millisecond checking during the tool-call middleware path.

## 5. Alternatives Considered
* **Cloud-Based Policy Revocation:** Rejected due to network latency (100ms+), which is too slow to stop machine-speed swarms.
* **Agent Self-Termination:** Rejected as compromised agents cannot be trusted to terminate themselves.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MSSQ triggers must be hardware-attested (TPM/Secure Enclave) to prevent "Defense Hijacking" where an attacker tries to quarantine legitimate teammates.
* **Observability:** Real-time visualization via the "Swarm Anomaly Visualizer" in the UI.

## 7. Evolutionary Changelog
* **2026-06-01:** Initial Document Creation.
<<<<<<< HEAD
* **2026-06-02:** Addressing Spectral Side-Channels in Quarantine Triggers.
    * **Context:** Today's market sync revealed "Spectral Reasoning" timing attacks that can leak mission constraints.
    * **Architecture Adjustment:**
        * Implementing "Reasoning-Aware Timing Jitter" for all quarantine revocation signals.
        * Decoupling the "Kill Switch" bitset update from the network notification loop to prevent timing correlation.
    * **Security Impact:** Neutralizes subagent ability to probe quarantine thresholds via ARE header timing analysis.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
