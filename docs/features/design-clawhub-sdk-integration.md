# Design Doc: ClawHub SDK Integration
**Status:** Draft
**Created:** 2026-04-11

## 1. Context and Scope
The rapid evolution of the OpenClaw ecosystem has exposed vulnerabilities in unregulated npm-based plugin distributions. To ensure supply chain integrity and secure skill discovery, MCP Any must integrate with the new ClawHub marketplace SDK. This transition moves the platform from a "blind ingestion" model to a "curated, profiled" model for agent capabilities.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a native adapter for the ClawHub SDK to facilitate secure tool discovery.
    * Implement automated behavioral profiling for all imported skills.
    * Enforce cryptographic provenance verification for marketplace-sourced tools.
    * Support "Burn-In" periods for new skills before granting access to sensitive resources.
* **Non-Goals:**
    * Operating a private marketplace (MCP Any acts as a consumer/bridge).
    * Providing manual code review for every ClawHub package.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Safely discover and install a "Database Optimizer" skill from ClawHub without risking host-level JVM injection.
* **The Happy Path (Tasks):**
    1. The developer searches for "Database Optimizer" in the MCP Any UI (connected to ClawHub).
    2. ClawHub returns the package metadata along with its verified behavioral manifest.
    3. MCP Any verifies the package's cryptographic signature against the ClawHub root CA.
    4. The package is executed in a "Burn-In" sandbox to profile its actual behavior against its manifest.
    5. Once verified, the tool is exposed to the agent discovery bus with limited, capability-based permissions.

## 4. Design & Architecture
* **System Flow:**
    `[ClawHub SDK] -> [MCP Any SDK Bridge] -> [Behavioral Profiler] -> [Discovery Bus]`
* **APIs / Interfaces:**
    * `ClawHubBridge`: `Search(query string) ([]PackageMetadata, error)`
    * `SkillValidator`: `Profile(package Package) (BehaviorScore, error)`
* **Data Storage/State:**
    * Local registry of verified ClawHub package hashes and behavioral profiles.

## 5. Alternatives Considered
* **Legacy npm Support:** Rejected due to persistent JVM injection vulnerabilities and lack of native behavioral manifests.
* **Internal Skill Whitelisting:** Rejected as it doesn't scale with the speed of the ecosystem.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All ClawHub-sourced tools are treated as untrusted until they pass behavioral profiling.
* **Observability:** Profiling logs and reputation scores are surfaced in the "Verified Skill Registry" dashboard.

## 7. Evolutionary Changelog
* **2026-04-11:** Initial Document Creation.
