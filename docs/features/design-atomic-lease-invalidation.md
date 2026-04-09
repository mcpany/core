# Design Doc: Atomic Lease Invalidation (ALI) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms move toward distributed, high-speed coordination, the propagation of security revocation signals has become a primary bottleneck. The 50ms+ lag in existing P2P tunnels (OpenClaw) allows malicious subagents to execute unauthorized turns even after a violation is detected.

The ALI Hub aims to provide kernel-mediated, sub-millisecond invalidation of agent capability leases by integrating directly with the local ebpf-based networking stack.

## 2. Goals & Non-Goals
* **Goals:**
    * Achieve <1ms revocation of Sovereign Node Tunnels upon drift detection.
    * Provide a unified interface for hardware-triggered invalidation.
    * Integrate with the Agentic Entropy Monitor (AEM) for automated triggers.
* **Non-Goals:**
    * Replacing the primary Zero-Trust policy engine.
    * Managing the initial minting of hardware leases (handled by HLML).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Security Architect
* **Primary Goal:** Prevent data exfiltration during a detected "Phantom Intent" hijack.
* **The Happy Path (Tasks):**
    1. The AEM detects stylometric drift in a specialist subagent's reasoning trace.
    2. The AEM issues an `INVALIDATE_MISSION` signal to the ALI Hub.
    3. The ALI Hub executes an ebpf bytecode update that drops all packets associated with the subagent's tunnel ID.
    4. The subagent is instantly isolated, and the parent agent receives a `LEASE_EXPIRED_ATOMIC` signal.

## 4. Design & Architecture
* **System Flow:**
    `AEM -> [ALI Hub (Go)] -> ebpf Map Update -> [Kernel Data Plane] -> Blocked Tunnel`
* **APIs / Interfaces:**
    * `POST /v1/ali/invalidate`: Triggers immediate revocation for a specific `mission_id`.
    * `GET /v1/ali/status`: Returns current active vs. invalidated leases.
* **Data Storage/State:**
    Uses an in-memory ebpf map for high-speed lookup and persistent SQLite for audit trails.

## 5. Alternatives Considered
* **User-Space Tunnel Collapse**: Rejected due to the 50ms+ context-switching and signal propagation overhead.
* **TPM Reset**: Rejected as too destructive; impacts other healthy missions on the same node.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ALI Hub requires a hardware-attested supervisor token to issue invalidation signals.
* **Observability:** Revocation events are logged with the nanosecond timestamp of the ebpf map update.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
