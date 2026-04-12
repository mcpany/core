# Design Doc: Adaptive State Compression (ASC)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of Sovereign Node Tunneling (SNT) in OpenClaw and the move toward multi-node agent meshes, transport latency has become a critical bottleneck for sub-millisecond tool execution. While zero-copy transport optimizes local handoffs, P2P tunnels across disparate execution environments suffer from serialization overhead and fluctuating link quality.

ASC addresses this by dynamically mediating the granularity of Binary State Handoffs (BSH). It ensures that the Universal Agent Bus (UAB) remains performant in high-latency environments by compressing state fragments based on real-time network telemetry and the semantic importance of the reasoning context.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-node state transfer latency by up to 70% in degraded network conditions.
    * Dynamically scale Protobuf/BSH quantization based on link-RTT.
    * Prioritize "Mission-Root" anchors and critical lineage metadata over high-entropy noise during compression.
* **Non-Goals:**
    * Implementing the underlying P2P tunneling protocol (delegated to AMT Broker).
    * Lossy compression of security-critical identity tokens or hardware attestations.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Agent Orchestrator
* **Primary Goal:** Execute a distributed data-processing task across three local devices without cognitive stall due to tunnel latency.
* **The Happy Path (Tasks):**
    1. The orchestrator agent initiates a tool call on a remote node via an AMT tunnel.
    2. ASC middleware detects a link latency spike (>50ms).
    3. ASC triggers "High-Compression" mode, quantizing subagent reasoning traces while pinning mission-root anchors.
    4. The remote node receives the compressed BSH fragment and executes the tool with minimal delay.
    5. Results are streamed back using adaptive delta-encoding.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Context] --> B[ASC Middleware]
        B --> C{Link Telemetry}
        C -- Low Latency --> D[Zero-Copy / Raw BSH]
        C -- High Latency --> E[Delta-Encoded / Quantized BSH]
        E --> F[AMT Tunnel]
        F --> G[Remote Gateway]
        G --> H[Reconstruction / Execution]
    ```
* **APIs / Interfaces:**
    * `SetCompressionPolicy(IntentID, Strategy)`: Allows mission-root to define which fragments are GC-Immune/Compression-Immune.
    * `X-UAB-Link-Quality`: Header for real-time latency signaling between nodes.
* **Data Storage/State:**
    * Local "Resumption Buffers" for maintaining the uncompressed ground-truth state.

## 5. Alternatives Considered
* **Static Compression (Gzip/Zstd):** Rejected because it doesn't account for the semantic importance of reasoning fragments; "blind" compression can lead to intent drift if critical guardrails are truncated.
* **Aggressive Context Pruning:** Rejected as it leads to "Context Amnesia" (GC Fragility).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Compression must be performed *after* cryptographic signing to ensure integrity, but *before* encryption to maximize ratio.
* **Observability:** Metrics for "Compression Utility Score" and "Reconstruction Fidelity" to monitor for semantic loss.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
