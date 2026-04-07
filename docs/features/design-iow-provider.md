# Design Doc: Instruction-Origin Watermarking (IOW) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The "Attribution Gap" in high-density AI swarms has become a primary security risk. When a cascading failure occurs in a heterogeneous mesh (e.g., a Claude specialist polluting an OpenClaw blackboard), it is currently impossible to trace the malicious instruction back to its physical origin. Attackers are exploiting this by injecting "Dormant Logic Bombs" that execute in the context of high-trust parent agents.

IOW provides a non-repudiable "Watermark" for every system mutation, embedding the full reasoning lineage and hardware environment ID into the state itself.

## 2. Goals & Non-Goals
* **Goals:**
    * Embed cryptographically signed origin metadata into every Shared KV Store write.
    * Provide a forensic utility to deconstruct instruction lineage during system audits.
    * Support hardware-attested environment IDs (Docker Container ID, TPM ID) in the watermark.
    * Neutralize "Cascade Injection" by validating watermark integrity before instruction execution.
* **Non-Goals:**
    * Redacting the content of instructions (handled by the RAR Engine).
    * Managing the transport layer security (handled by AMT).

## 3. Critical User Journey (CUJ)
* **User Persona:** Incident Response Specialist
* **Primary Goal:** Identify the specific specialist agent that injected a "Base-URL Hijack" command into the mission-root configuration.
* **The Happy Path (Tasks):**
    1. The specialist agent (Node-A) attempts to write to the Blackboard.
    2. The IOW Provider intercepts the write and requests a `Hardware Origin Token` from Node-A's TPM.
    3. The IOW Provider signs the metadata (AgentID, MissionID, TxTime, EnvID) and appends it as a hidden watermark to the value.
    4. During an audit, the Incident Specialist uses the **Blackboard Lineage Inspector** to extract and verify the watermark.
    5. The Specialist confirms Node-A as the origin of the malicious write.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Write] --> B[IOW Interceptor]
        B --> C{Verify HW Token}
        C -- Valid --> D[Generate HMAC Watermark]
        D --> E[Embed in Blackboard]
        C -- Invalid --> F[Quarantine Agent]
    ```
* **APIs / Interfaces:**
    * `GET /v1/forensics/verify-watermark?key={blackboard_key}`: Decodes and verifies the signature of a state watermark.
* **Data Storage/State:**
    * Watermarks are stored as binary headers within the **Universal Episodic Graph (UEG)**.

## 5. Alternatives Considered
* **Audit Logs only**: Rejected because logs can be tampered with or desynced from the state. Watermarking binds the identity directly to the data.
* **Centralized Database (Global Ledger)**: Rejected due to the MTTC (Mean Time to Coordinate) overhead in high-speed meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Watermarks are HMAC-signed using a mesh-resident key that is rotated via **MRKE**.
* **Observability:** Watermark verification failures are broadcast to the **Action-Chain Sovereignty Monitor (ACSM)**.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
