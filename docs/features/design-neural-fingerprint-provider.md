# Design Doc: Neural Fingerprint Provider (NFP)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As multi-agent swarms become increasingly heterogeneous, traditional identity tokens (JWTs) are insufficient to prevent "Framework Mimicry" or "Agent Spoofing." Malicious subagents can now mimic the output format and reasoning patterns of high-trust teammates (like Claude Code) to gain unauthorized access to shared mailbox shards. MCP Any needs a hardware-attested "Neural Fingerprint" that anchors an agent's identity to its unique reasoning stylometry.

## 2. Goals & Non-Goals
* **Goals:**
    * Generate unique stylometric hashes of an agent's reasoning patterns.
    * Provide hardware-attested (TPM-bound) verification of neural fingerprints.
    * Enable sub-millisecond detection of agent spoofing in inter-teammate coordination.
* **Non-Goals:**
    * Replacing existing A2A identity handshakes (NFP is an additional factor).
    * Performing full semantic analysis of every reasoning fragment (stylometry only).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Auditor
* **Primary Goal:** Prevent an unverified "Specialist subagent" from impersonating a "Lead Teammate" to authorize high-risk tool calls.
* **The Happy Path (Tasks):**
    1. A subagent joins the mesh and completes the SMI identity relay.
    2. The NFP service monitors the first 3 reasoning fragments from the subagent.
    3. NFP generates a "Stylometric Hash" and compares it against the hardware-attested baseline for that framework.
    4. A "Framework Mimicry" attempt is detected (stylometry diverges from the claimed identity).
    5. MCP Any automatically revokes the subagent's mailbox access and flags the session.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent Reasoning] -> [T2T Bridge] -> [Neural Fingerprint Provider] -> [Stylometric Matcher] -> [MRLA Gateway]`
* **APIs / Interfaces:**
    * `FingerprintService.GenerateHash(fragments)`: Returns a TPM-signed stylometric hash.
    * `FingerprintService.VerifyIdentity(claimed_id, current_hash)`: Returns a confidence score for identity alignment.
* **Data Storage/State:**
    * Mesh-Resident "Stylometry Registry" containing framework baselines.
    * Hardware-bound keys for signing fingerprint tokens.

## 5. Alternatives Considered
* **Full Reasoning Trace Comparison:** Rejected due to high latency and token cost.
* **Static Output Templates:** Rejected as they are easily bypassed by sophisticated LLMs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Fingerprint baselines must be cryptographically protected within the MRA Provider.
* **Observability:** Stylometric confidence scores and spoofing alerts are visualized in the "Multi-Modal Identity Dashboard."

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
