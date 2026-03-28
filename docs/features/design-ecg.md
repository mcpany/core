# Design Doc: Ephemeral Context Grafter (ECG)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As specialized agents require increasingly complex execution environments (e.g., specific CUDA versions, proprietary database drivers, or large ML model weights), maintaining a monolithic "Universal Sandbox" has become impractical. Specialist agents often need to "graft" their specific environment dependencies onto the shared mission sandbox.

The Ephemeral Context Grafter (ECG) provides a secure coordination service to manage these temporary grafts, ensuring they are isolated from other mission branches and are atomically reaped upon task completion to prevent resource leaks and cross-mission contamination.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Grafting Manager" that can mount/unmount ephemeral dependencies (FS mounts, env vars, ports) onto an active sandbox.
    * Provide hardware-attested cleanup (Reaping) triggered by task termination signals.
    * Enforce "Graft Isolation" to prevent grafted dependencies from creating side-channels.
    * Integrate with the Discovery Sandbox Middleware for pre-flight graft validation.
* **Non-Goals:**
    * Providing long-term persistent storage for specialists.
    * Replacing general-purpose container orchestration (handles in-sandbox grafts).

## 3. Critical User Journey (CUJ)
* **User Persona:** ML Specialist Agent
* **Primary Goal:** Graft a 50GB model-weight volume onto the mission sandbox for a single inference task without affecting other teammates.
* **The Happy Path (Tasks):**
    1. ML Agent requests a "Context Graft" for a specific volume ID via the ECG.
    2. ECG validates the request against the mission-root manifest and hardware identity.
    3. ECG performs the "Graft" (e.g., a read-only bind mount) into the agent's specific execution sub-path.
    4. ML Agent performs the inference task using the grafted weights.
    5. Upon task completion, the ML Agent (or the ECG monitor) signals task end.
    6. ECG "Grafting Reaper" atomically unmounts the volume and purges any associated ephemeral state.

## 4. Design & Architecture
* **System Flow:**
    `[Graft Request] -> [ECG Validator] -> [Sandbox Mount] -> [Task Execution] -> [Reaping Signal] -> [Atomic Unmount]`
* **APIs / Interfaces:**
    * `ecg.GraftContext(taskID string, manifest GraftManifest) -> error`: Mounts the requested dependencies.
    * `ecg.ReapContext(taskID string) -> error`: Forcefully removes all grafts for a task.
* **Data Storage/State:**
    * **Active Graft Registry:** A kernel-bound tracking table of all active mounts and their associated hardware owners.

## 5. Alternatives Considered
* **Static Sandbox Templates:** Rejected as they lead to "Environment Explosion" (thousands of combinations) and increase startup latency.
* **Agent-Managed Cleanup:** Rejected as compromised agents can skip cleanup to leave persistent backdoors. ECG provides host-mediated "Guaranteed Reaping."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Grafts must be read-only by default and strictly isolated via Inode-pinning to prevent "Mount-Point Hopping."
* **Observability:** Grafting events and reaping latencies are tracked in the "Sandbox Monitor."

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
