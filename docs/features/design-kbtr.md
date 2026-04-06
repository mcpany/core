# Design Doc: Kernel-Bound Tunnel Reaper (KBTR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of distributed P2P meshes (e.g., Attested Mesh Tunneling), the management of long-lived network sockets has become a security critical path. Standard application-level "close" signals are often ignored or bypassed by rogue subagents or during unexpected process crashes, leading to the "Phantom Mesh" vulnerability (CVE-2026-55102). In this scenario, an orphaned tunnel remains open at the kernel level, allowing lateral movement between devices without fresh origin-validation.

KBTR provides a fail-safe, kernel-resident enforcement layer that ensures all tunnels are physically severed upon mission completion or supervisor override.

## 2. Goals & Non-Goals
* **Goals:**
    * Use eBPF/Kernel hooks to monitor and close P2P sockets.
    * Bind socket lifecycle to the hardware-attested Mission Root.
    * Neutralize "Phantom Mesh" lateral movement.
    * Provide tamper-proof logging of socket termination events.
* **Non-Goals:**
    * Replacing standard TCP stack management for non-mesh traffic.
    * Managing inter-node routing (handled by AMT Broker).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Auditor
* **Primary Goal:** Ensure no agent-initiated tunnels remain active after a high-privilege mission is terminated.
* **The Happy Path (Tasks):**
    1. A multi-node mission completes or is revoked by the user.
    2. The Subagent Reaper sends a termination signal to the KBTR.
    3. KBTR identifies the PIDs and file descriptors associated with the mission's AMT tunnels.
    4. KBTR executes an eBPF program to forcefully reset the connections at the kernel level.
    5. The OS kernel clears the socket state, making the tunnel inaccessible even if the application process is still partially resident.
    6. KBTR logs the hardware-attested confirmation of the reaper action.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Termination] --> B[Subagent Reaper]
        B --> C[KBTR Middleware]
        C --> D{eBPF Enforcer}
        D -->|SIGKILL/TCP_RESET| E[Kernel Sockets]
        D --> F[Audit Log]
    ```
* **APIs / Interfaces:**
    * `kbtr.SeverMissionTunnels(missionID)`: Triggers the eBPF reset for all sockets bound to a mission.
* **Data Storage/State:**
    * **Mission-FD Map:** eBPF map tracking the relationship between hardware-attested mission IDs and active kernel-level file descriptors.

## 5. Alternatives Considered
* **User-Space Signal Handling:** Rejected because it can be blocked by a compromised process or ignored during a SIGKILL scenario.
* **Firewall Rules (iptables):** Rejected because dynamic rule generation is slower and less granular than per-socket kernel-level resets.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** KBTR itself must be triggered by a hardware-attested signal from the Supervisor node.
* **Observability:** Alerts are surfaced in the "Service Mesh Topology Monitor" when an orphan is detected and reaped.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
