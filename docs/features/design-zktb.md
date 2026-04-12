# Design Doc: Zero-Knowledge Topology Broker (ZKTB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As MCP Any evolves into a multi-node distributed agent mesh, the exposure of the full mesh topology (physical IP addresses, node relationships, and resource distribution) to connected agents has become a critical security liability. "Mesh Topology Reconnaissance" (MTR) allows a compromised subagent to map the mesh and target specific high-trust nodes for lateral movement or state exfiltration.

The Zero-Knowledge Topology Broker (ZKTB) is designed to act as an authoritative "Mesh Mask." It provides agents with the connectivity they need while hiding the underlying physical and logical structure of the mesh using hardware-attested "Virtual Hops."

## 2. Goals & Non-Goals
* **Goals:**
    * Hide physical node addresses and logical topology from all connected agents.
    * Provide stable "Virtual Node IDs" that map dynamically to physical nodes.
    * Utilize hardware-attested (TPM) "Virtual Hops" to normalize handshake timing and neutralize timing side-channels.
    * Enable seamless multi-node tool invocation without revealing the target node's location.
* **Non-Goals:**
    * Managing low-level network routing (handled by AMT Broker).
    * Providing end-to-end encryption (handled by T2T Encryption Bridge).
    * Replacing identity management (handled by HLNI Provider).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Prevent a specialized research agent from discovering that it is running on a specific restricted-access node or mapping the addresses of its peer agents.
* **The Happy Path (Tasks):**
    1. The orchestrator spawns a subagent across the mesh.
    2. The subagent requests a list of available peers for coordination.
    3. The ZKTB intercepts the request and generates a set of ephemeral "Virtual Node IDs" (e.g., `vnode-8821`, `vnode-3304`).
    4. The subagent initiates coordination with `vnode-8821` via the AMT Broker.
    5. The ZKTB transparently translates `vnode-8821` to its physical AMT tunnel ID.
    6. All handshake responses from peers are processed by ZKTB to inject "Timing Jitter" before reaching the subagent, preventing physical location inference.
    7. The subagent completes the task without ever knowing the physical layout of the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent] -->|Virtual Node ID| B[ZKTB]
        B -->|MTR-Proof Masking| C{Topology Map}
        C -->|Physical Tunnel ID| D[AMT Broker]
        D -->|Hardware-Attested Tunnel| E[Remote Node]
        B -->|Jitter Injection| A
    ```
* **APIs / Interfaces:**
    * `zktb.MaskTopology(physicalNodes) -> VirtualNodes`: Generates a virtualized view of the mesh.
    * `zktb.ResolveVirtualID(vNodeID) -> physicalID`: Authoritative translation for internal brokers.
    * `zktb.InjectHandshakeJitter(response) -> jitteredResponse`: Normalizes response timing to mask network distance.
* **Data Storage/State:**
    * **Virtualization Registry:** Ephemeral, mission-bound mapping of Virtual IDs to Physical AMT Tunnels.
    * **Jitter Profiles:** Hardware-locked timing baselines for different mesh tiers.

## 5. Alternatives Considered
* **Static NAT/Proxying:** Rejected because static IP-to-IP mapping is still susceptible to timing analysis and doesn't provide agent-aware mission isolation.
* **Full Mesh Encryption without Masking:** Rejected because encryption protects the *content* but not the *structure*. Knowledge of the structure is enough to launch targeted DoS or "Mesh Shadowing" attacks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The ZKTB is a core component of "Mesh Sovereignty." It requires HLNI-verified node identities to function.
* **Observability:** Topology masking makes debugging harder; ZKTB will provide "Auditor-Only" unmasking logs (TPM-signed) for the "Service Mesh Topology Monitor."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
