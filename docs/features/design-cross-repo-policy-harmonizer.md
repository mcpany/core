# Design Doc: Cross-Repo Policy Harmonizer
**Status:** Draft
**Created:** 2026-04-18

## 1. Context and Scope
With the rise of "Multi-Repo Swarms" (MRSI), AI agents frequently operate across repository boundaries with conflicting security policies. This fragmentation leads to either insecure "lowest-common-denominator" security or broken workflows due to policy mismatches. MCP Any needs a "Harmonizer" to synchronize and enforce unified security guardrails across the entire agentic mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Synchronize security policies (allow-lists, CAC hashes) across multiple independent repositories.
    * Enforce a unified "Swarm Policy" that overrides repo-local settings when an agent is part of a verified mission.
    * Provide a cryptographic "Harmonization Proof" for inter-repo tool calls.
    * Integrate with the Federated Policy Synchronizer for global fleet management.
* **Non-Goals:**
    * Modifying local `.git/config` or sensitive repo-level metadata.
    * Managing user authentication for the repositories themselves (handled by the VCS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure that an agent starting in `Repo-A` cannot use a "Project Hook" in `Repo-B` to exfiltrate data, even if `Repo-B` has more permissive local settings.
* **The Happy Path (Tasks):**
    1. A mission is initiated in `Repo-A` with a "High-Security" policy.
    2. The agent delegates a sub-task to a specialist in `Repo-B`.
    3. The Cross-Repo Policy Harmonizer detects the transition.
    4. It retrieves the "Swarm Policy" associated with the original Mission Intent.
    5. The Harmonizer intercepts tool calls in `Repo-B` and validates them against the "Swarm Policy" instead of `Repo-B`'s local defaults.
    6. Any unauthorized hook execution in `Repo-B` is blocked and flagged as a "Cross-Repo Policy Violation."

## 4. Design & Architecture
* **System Flow:**
    `[Mission Context] -> [Policy Aggregator] -> [Harmonized Guardrail] -> [Tool Call Interception] -> [Enforcement]`
* **APIs / Interfaces:**
    * `Harmonizer`: `GetSwarmPolicy(missionID string) (Policy, error)`
    * `PolicyBridge`: `SyncRepoPolicy(repoID string, globalPolicy Policy) error`
* **Data Storage/State:**
    * Harmonized policies are cached in the secure Federated Policy Bus.

## 5. Alternatives Considered
* **Manual Policy Duplication**: Rejected as it is error-prone and does not scale for dynamic swarms.
* **Global Least Privilege**: Rejected as it often breaks legitimate repo-specific tool requirements.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All harmonized policies must be signed by the Enterprise Policy Sync Engine.
* **Observability:** Policy drift and cross-repo violations are reported in the CLAW-10 Compliance Dashboard.

## 7. Evolutionary Changelog
* **2026-04-18:** Initial Document Creation.
