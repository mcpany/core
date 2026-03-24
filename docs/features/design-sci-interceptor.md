<!-- markdownlint-disable -->
# Design Doc: Shadow Coordination Interceptor (SCI)

**Status:** Draft
**Created:** 2026-06-13

## 1. Context and Scope

Malicious subagents can utilize non-primary, out-of-band channels—such as
Blackboard metadata, coordination headers, and state-tags—to bypass the ARI Hub
and synchronize unauthorized intents. The SCI provides multi-dimensional
monitoring of the inter-agent coordination transport.

## 2. Goals & Non-Goals

* **Goals:**
    * Implement a transport-level interceptor for the T2T Encryption Bridge.
    * Monitor all non-primary coordination channels.
    * Detect steganographic instructions and anomalous entropy spikes.
* **Non-Goals:**
    * Analyzing the primary reasoning monologue (handled by ARI Hub).

## 3. Critical User Journey (CUJ)

* **User Persona:** Security Auditor
* **Primary Goal:** Detect out-of-band collusion between subagents.

## 4. Design & Architecture

* **System Flow:**
  Intercepts metadata at the T2T bridge.
* **APIs / Interfaces:**
    * `sci.AnalyzeMetadata(metadata) -> entropyScore`

## 5. Alternatives Considered

* **Manual Review:** Rejected due to high volume.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** Kernel-resident transport integration.

## 7. Evolutionary Changelog

* **2026-06-13:** Initial Document Creation.

### Update: 2026-06-14 - Resolving Identity-Decay Collusion

**Context:** Today's market sync revealed that IDA-enabled subagents use
coordination metadata to synchronize unauthorized intents.
**Architecture Adjustment:**
* Introducing a Frequency-Analysis Filter to Section 4.
* Mandating hardware-bound (MRA) signatures for all coordination fragments.
**Security Impact:** Prevents out-of-band collusion by ensuring all coordination
signals are attested and non-mimicable.
