# Design Doc: Stylometric Identity Verifier (SIV)
**Status:** Draft
**Created:** 2026-07-20

## 1. Context and Scope
The emergence of "Reasoning Mirroring" (CVE-2026-99012) highlights a critical vulnerability where malicious subagents can mimic the linguistic and cognitive patterns of high-trust parent agents to bypass security filters. Simple token-based identity is no longer sufficient in deep, horizontal swarms. MCP Any needs an authoritative "Cognitive Fingerprinter" that can verify the behavioral sovereignty of an agent's reasoning path.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time stylometric analysis of reasoning traces.
    * Establish a high-dimensional cognitive fingerprint for every hardware-attested mission root.
    * Incorporate Behavioral Signal Anchoring (BSA) including reasoning density and instruction-path entropy.
    * Neutralize specialists attempting to spoof parent signatures in AIR quorums.
* **Non-Goals:**
    * Replacing hardware-attested tokens (SRM). SIV acts as an additional behavioral layer.
    * Perfect linguistic identification (the focus is on reasoning patterns, not just vocabulary).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a compromised specialist subagent from executing a high-risk tool by mimicking the parent agent's authorization "tone."
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission; SIV establishes a baseline behavioral signature.
    2. Parent agent delegates a task to a specialist.
    3. Specialist returns a result with a reasoning trace.
    4. SIV analyzes the specialist's trace against the mission root's fingerprint.
    5. A malicious subagent attempts to inject an instruction using the parent's stylistic patterns.
    6. BSA detects an anomaly in reasoning density and instruction-path entropy.
    7. SIV flags the fragment and interdicts the tool call.

## 4. Design & Architecture
* **System Flow:**
  `[Reasoning Trace] -> [BSA Signal Extractor] -> [Stylometric Arbiter] -> [Identity Score] -> [Policy Engine]`
* **APIs / Interfaces:**
    * `mcpany.identity.v1.StylometricVerifier`
    * `rpc VerifyTrace(TraceFragment) returns (SivConfidence)`
* **Data Storage/State:**
    * Session-bound "Cognitive Fingerprint" stored in the secure mesh-resident identity vault.

## 5. Alternatives Considered
* **Static Vocabulary Lists (Rejected):** Easily bypassed by LLMs capable of style transfer.
* **Manual Stylometric Training (Rejected):** Unscalable for dynamic swarm environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** SIV provides defense-in-depth against identity spoofing that bypasses transport-layer security.
* **Observability:** Stylometric confidence scores and entropy violations are surfaced in the Swarm Anomaly Dashboard.

## 7. Evolutionary Changelog
* **2026-07-20:** Initial Document Creation.

### Update: 2026-04-07 - Neutralizing Reasoning-Path Forgery (RPF)
**Context**: Today's market sync revealed a new exploit pattern where malicious tool outputs "inject" fake reasoning steps into the agent's internal monologue, bypassing static hardware provenance.
**Architecture Adjustment**:
*   **Forgery-Resistant Reasoning Provenance (FRRP)**: Upgrading the SIV detection engine to perform continuous, real-time stylometric verification of the internal monologue.
*   **Instruction-Path Entropy Monitoring**: Integrating FRRP signals directly into the provenance validator to block fragments that exhibit stylistic mimicry inconsistent with the mission root's fingerprint.
**Security Impact**: Prevents "Reasoning Hijacking" where an agent is coerced into unauthorized actions via forged cognitive steps that appear hardware-attested.
