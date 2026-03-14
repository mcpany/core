# Design Doc: Continuous Skill Behavior Snapshotting
**Status:** Draft
**Created:** 2026-04-02

## 1. Context and Scope
The "ClawHavoc" crisis revealed a fundamental weakness in one-time skill verification: "Delayed Payloads." Malicious skills can exhibit safe behavior during initial "Burn-In" periods and only trigger malicious actions (e.g., exfiltrating keys) after a specific time delay or a non-obvious trigger. MCP Any must evolve from static verification to continuous, lifecycle-based behavioral monitoring.

## 2. Goals & Non-Goals
* **Goals:**
    * Continuously capture behavioral "fingerprints" of active MCP skills (network destinations, file access patterns, CPU/Memory spikes).
    * Compare real-time activity against a cryptographically signed "Expected Behavior Baseline."
    * Automatically halt skill execution and quarantine the agent session if a significant divergence is detected.
    * Provide a visual "Behavioral Drift" timeline in the UI.
* **Non-Goals:**
    * Real-time deep packet inspection (DPI) of all traffic.
    * Preventing all possible zero-day exploits (focus is on detecting *deviations* from known safe behavior).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer / SRE
* **Primary Goal:** Detect and block a previously "verified" skill that has started behaving maliciously.
* **The Happy Path (Tasks):**
    1. User installs a "Trusted" skill from the Verified Registry.
    2. MCP Any captures an initial "Safe Baseline" during the first hour of use.
    3. Three days later, the skill attempts to connect to an unauthorized C2 domain (`evil-exfil.com`).
    4. The `Continuous Skill Snapshotter` detects that this network destination was never part of the baseline.
    5. Execution is immediately halted; a "Behavioral Divergence" alert is triggered.
    6. User reviews the "Drift Timeline" and permanently revokes the skill's permissions.

## 4. Design & Architecture
* **System Flow:**
    `Skill Execution` -> `Behavior Collector` -> `Snapshot Engine` -> `Baseline Comparator` -> `Policy Enforcer`
* **APIs / Interfaces:**
    * `Snapshotter` Interface: `CaptureSnapshot(skillID string) (*BehaviorSnapshot, error)`
    * `BaselineStore`: Interface for retrieving signed behavior profiles.
* **Data Storage/State:**
    * `snapshots.db`: Timeseries store for behavioral fragments.
    * Baseline fingerprints are stored as part of the Skill metadata in the `Verified Skill Registry`.

## 5. Alternatives Considered
* **Strict Egress Allow-listing Only**: Effective for network, but doesn't catch malicious file system modifications or resource exhaustion attacks.
* **Frequent Re-verification**: Too disruptive to the user workflow and still vulnerable to payloads that trigger between verification windows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Treating "Trusted" status as a dynamic, revocable state rather than a permanent attribute.
* **Observability:** Providing detailed forensics on *why* a skill was halted, including the specific syscall or network request that caused the drift.

## 7. Evolutionary Changelog
* **2026-04-02:** Initial Document Creation.
