# Design Doc: Predictive Mesh Coordination (PMC) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale to include dozens of specialists (e.g., Claude Code Agent Teams), the overhead of state look-ups and resource allocation becomes the primary performance bottleneck. Current reactive models wait for an agent to request a state shard before fetching or migrating it, leading to "Cognitive Stall" where the reasoning engine idles during transport.

The PMC Hub transitions MCP Any from a reactive proxy to a predictive orchestrator. By analyzing the current mission intent and subagent trajectories, the PMC Hub proactively re-shards state and pre-warms execution environments across the mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce Mean Time to Coordinate (MTTC) by >50%.
    * Implement real-time intent-based trajectory prediction.
    * Automate speculative state migration across distributed nodes.
* **Non-Goals:**
    * Replacing the underlying transport layer (e.g., AMT).
    * Predicting user intent (it only predicts subagent tool/state needs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Infrastructure Architect
* **Primary Goal:** Minimize latency in a 20-agent heterogeneous swarm distributed across local and cloud enclaves.
* **The Happy Path (Tasks):**
    1. The Mission-Root initiates a complex code-refactoring task.
    2. PMC Hub analyzes the task and identifies that "Database specialist" and "Documentation specialist" subagents will likely be spawned.
    3. PMC Hub proactively migrates relevant DB schemas and code docs to the predicted execution enclaves.
    4. Subagents spawn and find their required state fragments already local to their environment.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root] -> [PMC Trajectory Engine] -> [Distributed State Shards]`
    The Trajectory Engine uses a lightweight HMM (Hidden Markov Model) to score the probability of next-hop tool calls.
* **APIs / Interfaces:**
    * `POST /pmc/predict`: Ingests reasoning traces to update trajectories.
    * `GRPC PreWarmShard`: Internal signal to AMT nodes to fetch state.
* **Data Storage/State:**
    Uses the Universal Episodic Graph (UEG) as the primary source for trajectory training.

## 5. Alternatives Considered
* **Reactive-Only (Status Quo):** Rejected due to the 5s+ wait cycles observed in high-density swarms.
* **Global State Mirroring:** Rejected due to prohibitive bandwidth and memory overhead in large meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative migrations are restricted by the Hardware-Attested Mission Manifest (HAMM).
* **Observability:** Track "Prediction Accuracy" and "Pre-warm Hit Rate" metrics in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
