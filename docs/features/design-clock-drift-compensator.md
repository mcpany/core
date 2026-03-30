# Design Doc: Monotonic Clock-Drift Compensator (MCDC)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
With the rise of hardware-attested reasoning traces, time-based integrity has become a critical security frontier. The "Shadow-Attestation" vulnerability (2026-07-11) revealed that nanosecond-level timing drift between a Trusted Platform Module (TPM) and the host system clock can be exploited to inject "Ghost Fragments" into reasoning traces.

The Monotonic Clock-Drift Compensator (MCDC) provides a software-defined, hardware-synced monotonic clock for MCP Any. It ensures that every reasoning fragment and state mutation is timestamped with a normalized value that is immune to oscillator drift and manual system-time manipulation.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-precision monotonic clock with nanosecond resolution.
    * Provide real-time drift compensation between the TPM internal clock and the OS system time.
    * Expose a standard API for agents to request "Drift-Compensated Timestamps."
    * Integrate with the SRM Provider to sign reasoning fragments with compensated time.
* **Non-Goals:**
    * Replacing the system clock (MCDC is a middleware service, not a kernel clock).
    * Providing network-wide time synchronization (focus is local node integrity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer
* **Primary Goal:** Verify that a multi-hop reasoning trace was not subject to timing-drift injection.
* **The Happy Path (Tasks):**
    1. The officer selects a "Chain-of-Thought" lineage for audit.
    2. The audit tool requests the "Temporal Integrity Proof" from MCP Any.
    3. MCP Any retrieves the MCDC-normalized timestamps for each reasoning fragment.
    4. The tool verifies that the interval between fragments is consistent with the hardware-attested monotonic counter.
    5. Any drift detected is within the pre-attested MCDC tolerance window, confirming trace integrity.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        TPM[Hardware TPM] -->|Clock Signal| MCDC[MCDC Middleware]
        System[OS System Clock] -->|Oscillator| MCDC
        MCDC -->|Normalization| Clock[Monotonic Master Clock]
        Clock -->|Timestamp| SRM[SRM Provider]
        SRM -->|Signed Trace| Agent[Agent Reasoning]
    ```
* **APIs / Interfaces:**
    * `GET /mcdc/now`: Returns the current drift-compensated monotonic timestamp.
    * `GET /mcdc/drift`: Returns the current drift metrics between TPM and System clock.
* **Data Storage/State:**
    * MCDC maintains a "Calibration Table" in kernel-resident memory to track drift coefficients over time.

## 5. Alternatives Considered
* **Pure TPM Timestamping:** Rejected due to the high latency (>50ms) of per-call TPM clock queries.
* **NTP/PTP Synchronization:** Rejected because external time sources are not hardware-attested and can be spoofed in a "Shadow-Attestation" attack.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The MCDC calibration coefficients must be hardware-attested during node boot.
* **Observability:** Drift alerts are exported to the Hardware Trust Status Widget in the UI.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
