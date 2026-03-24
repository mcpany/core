# Design Doc: Intent-Splicing Detector (ISD)
**Status:** Draft
**Created:** 2026-06-06

## 1. Context and Scope
With the rise of heterogeneous agent swarms, subagents from different frameworks (e.g., OpenClaw, Claude Code) frequently coordinate through a shared message bus. A critical vulnerability, "Intent-Splicing," allows a compromised or malicious subagent to "splice" its own instructions into a parent agent's reasoning stream. This mimics the parent's intent and can coerce the entire swarm into unauthorized actions.

MCP Any needs to provide a robust defense mechanism that performs active deconstruction and structural validation of inter-agent messages. This ensures that every instruction within a message is explicitly authorized and anchored to the hardware-attested mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform real-time semantic deconstruction of inter-agent messages.
    * Validate every instruction fragment against the mission-root intent.
    * Detect and block "piggybacked" or "spliced" instructions that diverge from authorized goals.
    * Provide a hardware-attested audit trail for all intent-validation events.
* **Non-Goals:**
    * General-purpose prompt injection defense (handled by the Injection Shield).
    * Modifying the internal reasoning of individual agents.
    * Enforcing framework-specific syntax rules.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Architect
* **Primary Goal:** Prevent a specialized "Code Reviewer" subagent from "splicing" a "Self-Promote to Admin" instruction into its review report.
* **The Happy Path (Tasks):**
    1. Parent agent sends a mission-root token and a task delegation to the subagent via MCP Any.
    2. Subagent processes the task and generates a response containing its findings and a suggested next step.
    3. MCP Any intercepts the message and passes it to the AID Hub.
    4. AID Hub deconstructs the message into semantic fragments.
    5. Each fragment is checked against the parent's verified mission intent and the subagent's task-bound capability lease.
    6. ISD detects a "spliced" instruction that attempts to modify local security settings.
    7. MCP Any blocks the message, terminates the subagent session, and alerts the parent agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Message] --> B[AID Hub Interceptor]
        B --> C[Semantic Deconstructor]
        C --> D[Instruction Fragment Validator]
        D --> E{Authorized?}
        E -- Yes --> F[Message Forwarded to Parent]
        E -- No --> G[Splicing Alert & Session Termination]
        H[Mission Root Intent] --> D
        I[HAIL Lineage] --> D
    ```
* **APIs / Interfaces:**
    * `isd.ValidateMessage(context, message, missionToken) -> Result`: Primary validation endpoint.
    * `isd.RegisterIntent(missionToken, intentManifest)`: Registers authorized intents for a mission.
* **Data Storage/State:**
    * **Intent Registry:** Local in-memory store for active mission intents, backed by SQLite.
    * **Validation Logs:** Hardware-attested logs of all deconstruction events for forensic auditing.

## 5. Alternatives Considered
* **Framework-Level Validation:** Rejected because it relies on the honesty of the subagent framework. MCP Any provides a framework-neutral security boundary.
* **Simple Keyword Filtering:** Rejected because it is easily bypassed by sophisticated semantic mimicking. Structural deconstruction is required for "Splicing" detection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ISD itself must run in a secure enclave (HAPE) to prevent tampering. All intent mappings must be hardware-attested.
* **Observability:** Integrated with the "Intent-Splicing Audit Log" in the UI for real-time monitoring of detection events.

## 7. Evolutionary Changelog
* **2026-06-06:** Initial Document Creation.
