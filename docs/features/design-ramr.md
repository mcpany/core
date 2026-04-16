# Design Doc: Reasoning-Aware Mesh Router (RAMR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move toward multi-node, P2P-connected execution environments (OpenClaw Sovereign Node Tunneling), the overhead of mandatory hardware-attested encryption and high-latency tunnels is becoming a performance bottleneck.

MCP Any needs to solve this "Tunneling Tax" by implementing **Reasoning-Aware Mesh Routing (RAMR)**. This system dynamically selects the optimal inter-node transport based on the real-time reasoning complexity and tool risk, ensuring high-stakes actions remain secure while low-risk coordination stays performant.

## 2. Goals & Non-Goals
* **Goals:**
    * Dynamically route tool calls between "Fast-Path" session tunnels and "High-Sovereignty" hardware tunnels.
    * Reduce inter-node coordination latency by up to 60% for low-risk fragments.
    * Maintain TPM-bound security for all high-privilege operations (e.g., shell, secrets).
    * Provide real-time "Attestation Depth" adjustment to counter Attestation Fatigue.

* **Non-Goals:**
    * Replacing existing P2P protocols (e.g., WireGuard, Tailscale). RAMR acts as an orchestration layer *above* them.
    * Managing model-to-model (M2M) reasoning. It strictly handles tool-call and state-synchronization routing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Swarm Architect
* **Primary Goal:** Execute a 50-agent distributed refactoring mission without "Cognitive Stall" from attestation latency.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a "Code Analysis" task to a remote specialist.
    2. RAMR detects "Read-Only" reasoning intent and routes the task through a Fast-Path session tunnel.
    3. Specialist completes analysis and requests a "File Edit" (High-Risk).
    4. RAMR interdicts the request, upgrades the tunnel to Hardware-Locked Sovereignty, and mandates TPM-attestation.
    5. User approves the high-risk commit via the UI; RAMR routes the result back via the Fast-Path for subsequent analysis.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Request] --> B{RAMR Intensity Analyzer}
        B -- Low Risk / Low Entropy --> C[Fast-Path Session Tunnel]
        B -- High Risk / High Entropy --> D[Hardware-Locked Sovereign Tunnel]
        C --> E[Remote Tool / Node]
        D --> F[TPM-Attested Execution]
        F --> E
        E --> G[RAMR Result Arbiter]
        G --> H[Agent Response]
    ```
* **APIs / Interfaces:**
    * `GET /v1/mesh/route-policy`: Retrieve current routing thresholds.
    * `POST /v1/mesh/delegate`: Request-based delegation with suggested route priority.
    * `X-MCP-Reasoning-Risk`: New protocol header to signal intent-bound risk level.

* **Data Storage/State:**
    * **Route Cache**: Ephemeral, session-bound storage of active tunnel mappings.
    * **Risk Matrix**: Hardware-attested configuration defining risk scores for specific tool schemas.

## 5. Alternatives Considered
* **Static Tunneling**: Rejecting due to 100ms+ overhead for every coordination fragment, causing swarm-wide stall.
* **Pure Fast-Path**: Rejecting due to "Implicit Local Trust" vulnerabilities; high-risk tools MUST be hardware-attested.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** RAMR utilizes "Risk-Adaptive Attestation," ensuring that even on Fast-Paths, the initial session handshake is hardware-bound.
* **Observability:** Real-time visualization of tunnel distribution and latency gains in the UI Dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
