# Design Doc: State-Trust Labeling (STL) Provider
**Status:** Draft
**Created:** 2026-05-19

## 1. Context and Scope
As agent swarms become more heterogeneous, state from multiple frameworks (Claude, OpenClaw, Gemini) is being aggregated in the Shared KV Store (Blackboard). This creates a risk of "Protocol-Agnostic State Injection" (PASI), where low-trust state from a legacy or compromised MCP server is ingested by a high-trust reasoning loop. MCP Any needs to solve this by providing State-Trust Labeling (STL) to ensure every data fragment is cryptographically tagged with its origin's trust level.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically tag all Blackboard entries with a `trust_level` derived from the origin framework.
    * Implement "Trust-Aware Retrieval" where agents can filter results by a minimum trust threshold.
    * Provide a mandatory "Trust Upgrade" path for data that has been verified by a human-in-the-loop or security auditor.
* **Non-Goals:**
    * Encrypting individual KV pairs (focus is on provenance/integrity).
    * Providing a reputation system for all public MCP servers (handled by RBC).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Prevent an unauthenticated local tool's state from influencing a Claude Team's security decisions.
* **The Happy Path (Tasks):**
    1. A legacy MCP server writes `is_admin: true` to the Blackboard.
    2. The STL Provider automatically tags this entry with `trust_level: LOW`.
    3. A Claude "Security Lead" agent queries the Blackboard for user permissions with `min_trust: HIGH`.
    4. The Blackboard excludes the `is_admin` entry from the results.
    5. The mission continues without an unauthorized privilege escalation.

## 4. Design & Architecture
* **System Flow:**
    `[Data Source] -> [STL Middleware] -> [Blackboard Storage]`
* **APIs / Interfaces:**
    * `STL.get_label(origin_id)`: Resolves the trust level for a given framework or tool.
    * `Blackboard.set_with_trust(key, value, trust_label)`: Stores data with its cryptographic provenance tag.
    * `Blackboard.query(filter, min_trust)`: Retrieves data meeting the specified trust threshold.
* **Data Storage/State:**
    * The Blackboard SQLite schema is updated to include a `trust_tag` column containing the cryptographically signed label.

## 5. Alternatives Considered
* **Namespace Isolation**: Rejected as it prevents legitimate cross-framework collaboration (e.g., Claude *wants* to see OpenClaw's output, but needs to know it's "Audited" vs "Speculative").
* **Centralized Trust Authority**: Rejected as it introduces a single point of failure and doesn't scale with local-only swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: STL is a foundational component of the Zero Trust architecture, preventing privilege escalation via state injection.
* **Observability**: The "State Trust-Level Inspector" in the UI will color-code Blackboard data by its trust level.

## 7. Evolutionary Changelog
* **2026-05-19:** Initial Document Creation.
