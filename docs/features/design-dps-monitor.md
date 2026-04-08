# Design Doc: Decision-Path Sovereignty (DPS) Monitor
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move from executing single tools to orchestrating complex, multi-step automated workflows, a new class of vulnerability has emerged: **Decision-Path Poisoning**. Malicious inputs or poisoned model outputs can cause a "confused deputy" agent to deviate from its intended mission, initiating a chain of system interactions that results in cascading failures or unauthorized data exfiltration.

The DPS Monitor provides active, real-time validation of these automated action-chains. It ensures that every step in a multi-stage workflow remains anchored to the hardware-attested mission root and complies with organization-wide safety policies, even if the agent's internal reasoning has been compromised.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor and record the complete sequence of agent-to-system interactions (action-chains).
    * Perform real-time semantic validation of each step against the mission-root intent.
    * Detect and interdict "Drift" where an action sequence begins to diverge from authorized goals.
    * Provide a cryptographic "Chain of Sovereignty" for every automated workflow.
* **Non-Goals:**
    * Replacing existing per-call tool gating (Policy Firewall). DPS focuses on the *sequence* and *logic* of the path.
    * Controlling the internal weights of the model.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise DevOps Security Engineer
* **Primary Goal:** Prevent an autonomous "CI/CD Remediation Agent" from accidentally deleting production infrastructure due to a poisoned issue description.
* **The Happy Path (Tasks):**
    1. The agent receives a task to "Fix CVE-2026-1234 in Repository X."
    2. DPS Monitor ingests the mission root and initializes a "Sovereignty Session."
    3. The agent starts a workflow: `Read Code` -> `Run Tests` -> `Propose Fix`.
    4. DPS validates that `Read Code` and `Run Tests` align with the "Fix CVE" intent.
    5. A poisoned comment in the repository triggers a sub-intent: `Delete S3 Bucket`.
    6. DPS analyzes the `Delete S3 Bucket` action against the "Fix CVE" intent chain.
    7. DPS detects a "Decision-Path Divergence" and interdicts the call before the bucket is deleted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Agent Intent] --> B(DPS Session Tracker)
        B --> C{Action Validator}
        C -->|Valid| D[Execute Tool]
        C -->|Divergent| E[Interdict & Alert]
        D --> F(State Update)
        F --> B
    ```
* **APIs / Interfaces:**
    * `mcpany.dps.v1.SessionService`
    * `rpc ValidateAction(ActionFragment) returns (ValidationStatus)`
* **Data Storage/State:**
    * Action-chain graphs are stored in the hardware-attested Universal Episodic Graph (UEG).

## 5. Alternatives Considered
* **Static Graph Templates (Rejected):** Too rigid for dynamic agent reasoning.
* **Manual HITL for Every Step (Rejected):** Causes "Approval Fatigue" and destroys agent autonomy.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DPS relies on hardware-attested lineage to prevent subagents from "Self-Attesting" their way out of monitoring.
* **Observability:** Action-chain traces are surfaced in the Mesh-Resident Lineage Tracker for forensic auditing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
