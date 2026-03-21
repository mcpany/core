# Design Doc: State Trust Labeling (STL) Provider
**Status:** Draft
**Created:** 2026-05-19

## 1. Context and Scope
With the rise of multi-framework agent swarms (OpenClaw, Claude Code, AutoGen), the "Shared Blackboard" (KV Store) has become a primary target for Cross-Framework State Injection. An agent from a low-trust framework can currently overwrite state that a high-trust framework relies on for mission-critical reasoning. The STL Provider introduces cryptographic trust-tagging for all blackboard entries to ensure framework-level provenance and integrity.

## 2. Goals & Non-Goals
* **Goals:**
    * Attach immutable "Trust Labels" to every KV pair in the Blackboard.
    * Enable framework-specific "Write Isolation" policies.
    * Provide a standardized API for agents to query the trust level of a state fragment.
* **Non-Goals:**
    * It will not perform deep semantic analysis of the *content* of the state (handled by Semantic Integrity Bridge).
    * It will not manage agent-level identities (handled by A2A Messaging Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Prevent an unverified subagent from clobbering a "Root Mission Anchor" in the shared blackboard.
* **The Happy Path (Tasks):**
    1. A "High-Trust" parent agent writes a mission anchor to the Blackboard with a hardware-attested label.
    2. A "Low-Trust" subagent attempts to overwrite the same key.
    3. The STL Provider intercepts the request, compares the trust labels, and rejects the write.
    4. The parent agent is notified of the attempted integrity violation.

## 4. Design & Architecture
* **System Flow:**
    `Agent Tool Call` -> `Blackboard Wrapper` -> `STL Policy Engine` -> `SQLite Store`
* **APIs / Interfaces:**
    * `SetWithLabel(key, value, trust_token)`
    * `GetWithMetadata(key) -> (value, trust_label, origin_framework)`
* **Data Storage/State:**
    * Schema update for `blackboard` table to include `trust_label_id` and `signature_blob`.

## 5. Alternatives Considered
* **Namespace Isolation:** Rejected because it prevents legitimate cross-framework coordination which is a core value proposition of MCP Any.
* **Read-Only Shards:** Rejected because agents often need to collaboratively refine state, just with different levels of authority.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All trust labels are cryptographically bound to the session's hardware attestation.
* **Observability:** Every label-check failure is logged as a security event in the `audit_log`.

## 7. Evolutionary Changelog
* **2026-05-19:** Initial Document Creation.
