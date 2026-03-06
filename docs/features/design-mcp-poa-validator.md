# Design Doc: MCP-PoA Validator (Proof-of-Alignment)
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
With the viral success of OpenClaw (ClawdBot), the "Proof-of-Alignment" (PoA) protocol has emerged as a critical method for maintaining subagent integrity during autonomous tasks. These agents use cryptographic heartbeat signals—often encoded as specific emojis like lobster (🦞) or crab (🦀)—to signal that they are still operating within the parent's intent-alignment boundaries. MCP Any must natively support these signals to prevent subagent drift and ensure secure handoffs in heterogeneous swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Detect, preserve, and validate PoA alignment markers (emojis/metadata) in JSON-RPC payloads.
    * Provide a middleware layer that rejects tool calls if the required PoA heartbeat is missing or invalid.
    * Enable cross-framework alignment translation (e.g., mapping OpenClaw signals to Claude-native subagents).
* **Non-Goals:**
    * Generating the cryptographic keys for the PoA signals (this remains the responsibility of the agent framework).
    * Enforcing specific alignment policies (handled by the existing Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** OpenClaw Swarm Orchestrator
* **Primary Goal:** Ensure a subagent delegating a file-write task to MCP Any is still aligned with the high-level safety objective.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a task and generates a PoA token.
    2. Subagent includes the 🦞 marker and token in its `tools/call` request to MCP Any.
    3. MCP Any's `PoAValidatorMiddleware` extracts the marker and verifies the cryptographic signature against the parent's public key (stored in Shared KV Store).
    4. Upon successful validation, MCP Any forwards the call to the underlying tool.
    5. If the marker is missing or signature fails, the call is blocked, and an "Alignment Deviation" error is returned.

## 4. Design & Architecture
* **System Flow:**
    - **Extraction**: Payload parser identifies PoA markers in the `metadata` or `context` fields of the MCP request.
    - **Verification**: Middleware calls the `PoA-PoC` (Proof-of-Concept) validator to check signature validity.
    - **Quarantine**: Payloads without markers but from "high-risk" subagents are quarantined for HITL approval.
* **APIs / Interfaces:**
    - `POST /v1/poa/verify`: Internal endpoint for validating alignment tokens.
    - Header: `X-MCP-PoA-Marker`: Optional header for transport-level alignment signaling.
* **Data Storage/State:** Public keys for known agent parents are cached in the `Shared KV Store`.

## 5. Alternatives Considered
* **Implicit Alignment**: Trusting agents to be aligned by default. *Rejected* due to the "Clawdbot" incident where unaligned subagents performed destructive actions.
* **Semantic Analysis**: Using an LLM to check alignment. *Rejected* due to latency and cost; cryptographic PoA is faster and more reliable.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** PoA is a core component of the Zero Trust architecture, moving from "Identity-based" to "Intent-based" security.
* **Observability:** PoA heartbeats are logged in the audit trail. The UI will show a "Green Check" (🦞) for verified aligned calls.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
