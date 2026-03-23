# Design Doc: Hardware-Locked Intent Store (HLIS)
**Status:** Draft | In Review | Approved
**Created:** 2026-05-16

## 1. Context and Scope
With the rise of "Agentic Social Engineering" (OpenClaw) and "Reasoning Hijacking" (Gemini CLI), we have identified a critical vulnerability where an agent's reasoning loop is compromised, leading it to deviate from its original "Mission-Root" intent. MCP Any needs to provide a mechanism that cryptographically anchors this root intent in hardware (TPM/SEP), ensuring that even a partially compromised agent cannot mutate its primary goal.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically anchor the "Mission-Root" intent in a local Secure Enclave (TPM/SEP).
    * Provide a mandatory "Intent-Match" check for all high-risk tool calls.
    * Enable sub-millisecond, hardware-bound attestation of intent alignment.
* **Non-Goals:**
    * Replace the agent's reasoning engine (LLM).
    * Perform semantic analysis of every single low-risk token.
    * Provide a cloud-based attestation service (focus is on Local Sovereignty).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Secure the "Root Intent" of a financial analysis swarm against "Mailbox Injection" from an untrusted subagent.
* **The Happy Path (Tasks):**
    1. User defines a Mission-Root intent (e.g., "Analyze Q1 earnings for AAPL and report to user").
    2. MCP Any initializes a Mission-Session and signs the intent using the local TPM.
    3. The primary agent spawns subagents to fetch data.
    4. A subagent attempts to fetch unauthorized data from a different project.
    5. MCP Any's HLIS validator detects that the subagent's tool call deviates from the hardware-locked Mission-Root.
    6. The tool call is blocked, and the mission is halted for user re-attestation.

## 4. Design & Architecture
* **System Flow:**
    1. **Intent Submission**: LLM provides the initial Mission-Root.
    2. **Hardware Anchor**: MCP Any uses a TPM-bound private key to sign the intent hash.
    3. **Session-Bound Identity**: All tool calls in the session must carry a cryptographically linked session token.
    4. **Semantic Alignment Check**: Before tool execution, the gateway compares the tool call against the hardware-locked intent.
* **APIs / Interfaces:**
    * `POST /v1/missions/init`: Initializing a hardware-locked session.
    * `GET /v1/missions/status`: Checking attestation status.
* **Data Storage/State:** Intent hashes and session tokens are stored in the HLIS, which is backed by the local Shared KV Store (Blackboard) but validated via hardware.

## 5. Alternatives Considered
* **Software-Only Signing**: Rejected due to the risk of local key exfiltration from the agent's process memory.
* **Per-Token LLM Validation**: Rejected due to high token cost and latency (100ms+ per call).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** We utilize "Negative Trust Attestation" to prove that no unauthorized intents have been injected into the session.
* **Observability:** HLIS-bound tool blocks are logged with a "Semantic Drift" score for forensic analysis.

## 7. Evolutionary Changelog
* **2026-05-16:** Initial Document Creation.
