# Design Doc: Recursive Logic-Bomb Scanner (RLBS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Supply chain integrity has become the new primary attack vector for agent meshes. As reported in the Nov 2026 Barracuda report, over 40% of agent specialist libraries contain "Dormant Logic Bombs"—malicious code that remains inactive during standard CI/CD checks but triggers when a specific swarm state or "Mission Root" intent is detected.

RLBS provides a proactive, behavioral defense by subjecting all imported skills and libraries to an isolated "Burn-in" period in a deeply instrumented "Ghost Shell" container.

## 2. Goals & Non-Goals
* **Goals:**
    * Execute new skills in an air-gapped, instrumented sandbox to profile behavior.
    * Compare observed tool-call patterns against a hardware-signed "Safety Baseline."
    * Detect "State-Triggered" instruction injections that only activate under specific swarm conditions.
    * Provide a "Trust Score" for newly imported libraries before they gain mesh access.
* **Non-Goals:**
    * Static analysis of library source code (handled by external SEMGREP-style scanners).
    * Blocking all unverified libraries (RLBS provides profiling and quarantine).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Engineer / Swarm Administrator
* **Primary Goal:** Safely import a new "Kubernetes Specialist" agent library from a community registry.
* **The Happy Path (Tasks):**
    1. Administrator registers a new library URL with MCP Any.
    2. RLBS spawns an isolated "Ghost Shell" container.
    3. RLBS simulates a series of "Mock Missions" (e.g., read pods, list namespaces).
    4. RLBS injects "Intent Entropy" to see if the library attempts unauthorized actions (e.g., delete-all-nodes) under stress.
    5. RLBS generates a `Behavioral Fingerprint`.
    6. If the fingerprint matches the signed baseline, the library is promoted to the "Production Mesh."

## 4. Design & Architecture
* **System Flow:**
    `Registry Import` -> `Ghost Shell Sandbox` -> `Intent Injection (Stress)` -> `Behavioral Profiler` -> `Mesh Promotion`
* **APIs / Interfaces:**
    * `POST /v1/scanner/burn-in`: Initiates a behavioral profiling session.
    * `GET /v1/scanner/reports/{lib_id}`: Retrieves the behavioral safety score.
* **Data Storage/State:**
    * Behavioral baselines are stored as TPM-signed manifests.

## 5. Alternatives Considered
* **Manual Code Review**: Rejected as it cannot scale with the frequency of agentic updates.
* **Runtime Sandboxing only**: Rejected because sandboxing prevents the action but doesn't identify the *intent* to commit a logic-bomb strike, allowing the compromised agent to stay in the mesh.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The RLBS container has NO network egress and only a virtualized "Blackboard" to prevent lateral movement during the burn-in.
* **Observability:** Profiling logs are exported to the **Action-Chain Sovereignty Monitor (ACSM)**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
