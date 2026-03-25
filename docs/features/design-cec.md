# Copyright 2026 Author(s) of MCP Any

# SPDX-License-Identifier: Apache-2.0

# Design Doc: Cognitive Entropy Controller (CEC)

**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope

Reasoning traces in deep agent swarms are vulnerable to "Entropy Injection"
attacks, where high-entropy noise is injected into subagent responses to
bypass mission-root constraints or trigger context-window eviction.

The Cognitive Entropy Controller (CEC) provides high-performance,
hardware-accelerated monitoring of semantic entropy within the agentic
bus. It ensures that the reasoning path remains stable and aligned with
the primary intent.

## 2. Goals & Non-Goals

* **Goals:**
 * Real-time calculation of semantic entropy for reasoning fragments.
 * Hardware-acceleration using TPM/GPU for sub-10ms latency.
 * Configurable entropy thresholds pinned to the mission root.
 * Automated quarantine of high-entropy fragments.
* **Non-Goals:**
 * Automatically rewriting or "correcting" reasoning traces.
 * Managing token budgets (handled by RBF).

## 3. Critical User Journey (CUJ)

* **User Persona:** Security Auditor for Autonomous Swarms.
* **Primary Goal:** Prevent a specialized subagent from "blinding" the
 parent agent via high-entropy noise injection.
* **The Happy Path (Tasks):**
 1. The Auditor defines an entropy policy for the mission.
 2. A subagent generates a reasoning trace with anomalous noise.
 3. CEC intercepts the fragment in the middleware chain.
 4. CEC utilizes the TPM to calculate fragment entropy.
 5. Entropy exceeds the threshold; CEC flags and quarantines the fragment.
 6. The parent agent receives a "Security Alert" instead of the noise.

## 4. Design & Architecture

* **System Flow:**
 ```mermaid
 graph LR
  A[Agent] -->|Reasoning Trace| M[CEC Middleware]
  M -->|Calculate| H[Hardware Accelerator]
  H -->|Entropy Score| M
  M -->|Verify| P[Mission Policy]
  P -->|Valid| O[Tool/Next]
  P -->|High Entropy| Q[Quarantine Hub]
 ```
* **APIs / Interfaces:**
 * `POST /v1/entropy/calculate`: Calculate entropy for a text fragment.
 * `GET /v1/entropy/thresholds`: Retrieve active thresholds for mission.
* **Data Storage/State:**
 * Entropy metrics are stored in the Mesh-Resident Lineage Tracker.

## 5. Alternatives Considered

* **Software-based Entropy Calculation:** Rejected due to high latency
 blocking real-time agent coordination.
* **Post-hoc Audit Logs:** Rejected because they don't prevent the immediate
 exploitation of the parent's context window.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Entropy thresholds are cryptographically signed
 at mission boot and cannot be modified by subagents.
* **Observability:** Real-time entropy scores are exported to the
 Entropy Monitor Dashboard.

## 7. Evolutionary Changelog

* **2026-06-18:** Initial Document Creation.
