# Design Doc: Marketplace Skill Profiler
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
With the launch of agent tool marketplaces like OpenClaw's ClawHub, users are increasingly exposed to third-party agent skills. While these skills extend agent capabilities, they introduce significant supply-chain risks, including "Skill-Squatting" and "Delayed Payload" attacks. The Marketplace Skill Profiler provides a secure, air-gapped sandbox for behavioral profiling of these tools before they are promoted to the active discovery bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Execute third-party skills in a "Ghost Shell" (air-gapped, instrumented container).
    * Profile network, filesystem, and subprocess activity against a "Zero-Trust Baseline."
    * Generate a "Safety Score" and behavioral manifest for every skill.
    * Mandate hardware-attested user approval for tools failing the safety threshold.
* **Non-Goals:**
    * Modifying the source code of third-party tools.
    * Replacing the need for runtime monitoring (this is a pre-discovery gate).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Operator
* **Primary Goal:** Safely install a "GitHub Triage" skill from a community marketplace.
* **The Happy Path (Tasks):**
    1. User initiates installation of a new skill via the Marketplace UI.
    2. MCP Any intercepts the request and routes the skill to the Profiler Sandbox.
    3. The skill is executed in an instrumented environment with simulated inputs.
    4. The Profiler detects an unauthorized attempt to access `~/.ssh/` during the "Dry Run."
    5. The installation is quarantined, and a "Safety Violation" report is generated.
    6. The user reviews the report and blocks the malicious skill.

## 4. Design & Architecture
* **System Flow:**
    * **Ingestion**: Skill metadata and binaries are pulled from the marketplace.
    * **Profiling**: Skill runs in a restricted gVisor sandbox with syscall-level auditing.
    * **Analysis**: Activity is compared against the pre-declared schema.
* **APIs / Interfaces:**
    * `profiler.ProfileSkill(skillPackage) -> ProfileReport`
    * `profiler.GetSafetyScore(skillId) -> float`
* **Data Storage/State:** Behavioral reports stored in the local audit DB; safety scores cached in the Registry.

## 5. Alternatives Considered
* **Static Analysis Only**: Rejected as it cannot detect dynamic payloads or obfuscated commands.
* **Manual Code Review**: Rejected as it doesn't scale with the volume of community tools.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Profiler must be air-gapped from the host network and sensitive volumes.
* **Observability:** Profiling logs are streamed to the "Ghost Shell Profiling Viewer" in real-time.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
