# Design Doc: Temporal Shard Isolation (TSI) Hub
**Status:** Draft
**Created:** 2026-07-16

## 1. Context and Scope
The disclosure of CVE-2026-44012 (Shadow-Context side-channels) in OpenClaw confirmed that logical isolation in shared reasoning buffers is insufficient. Malicious subagents can probe mission-root constraints by monitoring micro-timing variations in memory access. MCP Any must evolve from simple memory segmentation to hardware-enforced **Temporal Isolation**.

The TSI Hub acts as the mesh-wide broker for access windows, ensuring that shard availability is decoupled from reasoning execution patterns.

## 2. Goals & Non-Goals
* **Goals:**
    * Neutralize timing side-channels in shared reasoning buffers.
    * Enforce hardware-locked (TPM) temporal access windows for memory shards.
    * Inject monotonic, hardware-attested jitter into all state synchronization responses.
* **Non-Goals:**
    * Managing the physical allocation of shared memory (handled by ZCMB/DME).
    * Providing semantic validation of shard content (handled by ARI).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Prevent a specialist "Data Analyst" agent from mapping the "CEO's Mission Root" via shared memory timing.
* **The Happy Path (Tasks):**
    1. Orchestrator requests access to a mission-critical memory shard.
    2. TSI Hub assigns a hardware-attested **Temporal Access Window**.
    3. The ZCMB/DME Broker only allows memory-mapping during the assigned window.
    4. TSI Hub injects 2ms-10ms of hardware-bound jitter into every coordination fragment.
    5. Subagent attempt to map parent attention maps via latency monitoring fails due to normalized temporal response.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [ZCMB/DME Broker] -> [TSI Hub (TPM Timers)] -> [Jitter Injection] -> [Shared Memory]`
* **APIs / Interfaces:**
    * `GET /v1/temporal/window/{shard_id}`: Retrieves the authorized access window.
    * `X-TSI-Monotonic-Counter`: Header containing the hardware-attested timer signature.
* **Data Storage/State:**
    * Stateless broker relying on hardware monotonic counters (TPM 2.0).

## 5. Alternatives Considered
* **Software-Only Jitter (Rejected):** Susceptible to kernel-level latency profiling.
* **Full Process Isolation (Rejected):** Prohibitive latency for sub-millisecond state sharing required by Agent Teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Access is denied outside of the temporal window regardless of cryptographic token validity.
* **Observability:** Logs include temporal window utilization and jitter distribution metrics.

## 7. Evolutionary Changelog
* **2026-07-16:** Initial Document Creation.
