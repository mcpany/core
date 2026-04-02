# Design Doc: ClawHub Ingestion Adapter
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
The transition of the OpenClaw ecosystem to the curated "ClawHub" marketplace (v2026.3.22) demands a specialized adapter in MCP Any. Currently, third-party skills are often ingested without rigorous provenance or sandboxing, leading to risks like the "ClawHavoc" crisis. This adapter will serve as the authoritative gateway for ClawHub-compliant skills, ensuring they meet enterprise-grade security standards before being exposed to agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement hardware-attested provenance checks for all skills fetched from ClawHub.
    * Provide real-time behavioral profiling for new marketplace skills in an isolated sandbox.
    * Enforce mandatory signature validation using the MCP Any **Verified Skill Registry**.
    * Integrate with the **OpenShell SSH Adapter** for secure command execution.
* **Non-Goals:**
    * Hosting the marketplace itself (ClawHub remains the upstream source).
    * Curating the skills manually; the system relies on automated attestation and quorums.

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Officer
* **Primary Goal:** Allow developers to use the `web-search` skill from ClawHub without risking host-level exfiltration.
* **The Happy Path (Tasks):**
    1. Developer requests installation of `web-search` from ClawHub.
    2. ClawHub Ingestion Adapter fetches the skill and its **Hardware Manifest**.
    3. Adapter verifies the manifest signature against the verified registry.
    4. Adapter triggers a "Shadow Execution" in the **Ghost Shell** to profile network/file access.
    5. Once profiled as "Safe," the skill is mapped to an isolated **OpenShell SSH Sandbox**.
    6. Skill is activated and made visible to the agent via the PNTD registry.

## 4. Design & Architecture
* **System Flow:**
    `[ClawHub] -> [Ingestion Adapter] -> [Verified Registry Check] -> [Ghost Shell Profiler] -> [PNTD Registry]`
* **APIs / Interfaces:**
    * `ingestion/clawhub/fetch`: Fetches and validates a skill package.
    * `ingestion/clawhub/attest`: Generates a hardware-bound attestation for a skill.
* **Data Storage/State:**
    Stores skill metadata and profiling reports in `registry.db`.

## 5. Alternatives Considered
* **Direct NPM Ingestion:** Rejected due to lack of curated security signals and high vulnerability rates.
* **Manual Skill Vetting:** Rejected as it cannot keep pace with the 13,000+ available skills in Q1 2026.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory multi-signature attestation (MSSA) for all high-risk skills.
* **Observability:** Visualizes skill provenance and behavioral history in the **Marketplace Browser**.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
