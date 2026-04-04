# Design Doc: Marketplace Policy Proxy
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The launch of ClawHub and other agent skill marketplaces (hosting over 4,000 community skills) introduces a massive supply-chain risk. Unlike local MCP servers, marketplace skills are dynamic and often un-vetted. Direct integration of these skills into production swarms can lead to prompt injection, data exfiltration, or "Ghost Execution."

The Marketplace Policy Proxy (MPP) acts as a secure, insulating layer between external marketplaces and the internal agent bus. It performs automated behavioral profiling and policy enforcement on marketplace-sourced tools before they are exposed to the reasoning engine.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all marketplace discovery requests and tool schemas.
    * Perform "Structural Sanitization" on incoming tool descriptions (detect imperitve instructions).
    * Execute new skills in a "Burn-In Sandbox" to profile network and filesystem behavior.
    * Mandate hardware-attested auditor signatures for high-risk marketplace skills.
* **Non-Goals:**
    * Hosting the marketplace itself (MPP is a proxy/gateway).
    * Providing model-level filtering (handled by IDS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer
* **Primary Goal:** Allow developers to experiment with new ClawHub skills while ensuring they cannot access the production database or local environment variables.
* **The Happy Path (Tasks):**
    1. Developer requests a new skill from ClawHub.
    2. MPP intercepts the `discovery` signal and redirects the skill to a "Ghost Shell" Burn-In sandbox.
    3. MPP monitors the skill for 5 minutes, logging all attempted network calls and file reads.
    4. MPP identifies an unauthorized outbound ping to an unknown domain and flags the skill.
    5. MPP automatically generates a "Restricted Capability Card" that blocks network egress for this specific skill.
    6. Skill is promoted to the "Internal Discovery Bus" with a P1 safety warning.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[ClawHub] -->|Skill| B[Marketplace Proxy]
        B -->|Behavioral Profiling| C[Ghost Shell Sandbox]
        C -->|Trace| B
        B -->|Attested Capability| D[Agent Discovery Bus]
    ```
* **APIs / Interfaces:**
    * `MarketplaceProxy`: `ProfileSkill(skillID string, schema JSON) (SafetyScore, error)`
    * `PolicyEnforcer`: `ApplySandboxConstraints(skillID string, constraints Policy) error`
* **Data Storage/State:**
    * Profiling reports are stored in the Verified Skill Registry with cryptographic provenance.

## 5. Alternatives Considered
* **Manual Skill Review:** Rejected due to the scale of ClawHub (4,000+ skills).
* **Static Schema Analysis:** Rejected because it cannot detect runtime exfiltration behaviors.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MPP utilizes "Trace-Aware Identity Propagation" to ensure marketplace skills carry their profiling history throughout the mission.
* **Observability:** Profiling status is visualized in the "Metadata Poisoning Alert Hub."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
