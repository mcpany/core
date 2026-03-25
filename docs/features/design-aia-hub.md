# Design Doc: Active Intent Alignment (AIA) Hub
**Status:** Draft
**Created:** 2026-06-21

## 1. Context and Scope
Autonomous swarms often suffer from "Intent Drift" during deep reasoning steps. The AIA Hub provides a hardware-attested mechanism to verify that subagents remain anchored to the mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide TPM-signed alignment heartbeats.
    * Enforce semantic entropy thresholds on inter-agent messages.
* **Non-Goals:**
    * Real-time reasoning modification.
    * Replacement of existing framework-specific alignment logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Ensure subagent reasoning remains aligned with the "Mission Root" without exposing internal monologues.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes mission with AIA Hub.
    2. Hub issues a mission-bound attestation token.
    3. Subagents include alignment heartbeats in coordination fragments.
    4. Hub validates heartbeats using local TPM.

## 4. Design & Architecture
* **System Flow:**
    [Mission Root] -> [AIA Hub (TPM)] -> [Attestation Token] -> [Subagent Reasoning] -> [Alignment Heartbeat] -> [AIA Hub (Validation)]
* **APIs / Interfaces:**
    * `/v1/alignment/heartbeat`: Submit reasoning trace for alignment validation.
    * `/v1/mission/anchor`: Bind a mission root to a hardware key.

## 5. Alternatives Considered
Software-only signing was rejected due to the risk of mission-root key exfiltration by rogue subagents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory TPM 2.0 attestation for all heartbeat signatures.
* **Observability:** Integration with the Reasoning Lineage Visualizer.

## 7. Evolutionary Changelog
* **2026-06-21:** Initial Document Creation.
