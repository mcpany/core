# Design Doc: OpenShell Runtime Adapter
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the launch of NVIDIA OpenShell™, agent security has moved from reasoning-level "best efforts" to runtime-enforced policy sovereignty. MCP Any must bridge the gap between high-level mission intents and the low-level, policy-enforced guardrails of the OpenShell runtime.

The OpenShell Runtime Adapter allows MCP Any to dynamically compile and push security policies to agents running within OpenShell-compliant environments. This ensures that even if an agent's reasoning is bypassed or compromised (e.g., via prompt injection), its actual actions remain strictly bound to the hardware-attested security perimeter.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Map MCP Any "Mission Roots" to OpenShell policy manifests.
    *   Provide real-time updates to runtime guardrails based on task escalation.
    *   Verify that the target agent is running in a valid, hardware-attested OpenShell environment.
    *   Enforce "Local-Only" or "Network-Gated" execution at the kernel level via OpenShell.
*   **Non-Goals:**
    *   Implementing the OpenShell runtime itself (MCP Any acts as the controller/adapter).
    *   Replacing the reasoning-level Policy Firewall (both layers work in tandem).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Enforce strict data-egress and filesystem boundaries for an autonomous data-migration agent.
*   **The Happy Path (Tasks):**
    1.  The Architect defines a "Mission Root" policy in MCP Any restricting file access to `/data/migration/`.
    2.  The OpenShell Adapter compiles this policy into an OpenShell `policy.json` manifest.
    3.  The agent boots within an OpenShell environment; MCP Any performs a hardware-attested handshake to verify the environment.
    4.  The Adapter pushes the compiled policy to the OpenShell kernel.
    5.  When the agent attempts to read `/etc/passwd` (even if coerced by a prompt injection), the OpenShell runtime blocks the call at the OS level and triggers an alert in MCP Any.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Mission[Mission Root Policy] --> Compiler[Policy-to-Guardrail Compiler]
        Compiler --> Adapter[OpenShell Runtime Adapter]
        Adapter -->|Push Policy| OS_Kernel[OpenShell Runtime Engine]
        OS_Kernel -->|Hardware Attestation| Adapter
        Agent[Autonomous Agent] -->|Tool Call| OS_Kernel
        OS_Kernel -->|Action Allowed/Denied| Agent
    ```
*   **APIs / Interfaces:**
    *   `PUT /v1/openshell/policies/:agent_id`: Updates the runtime guardrails for a specific agent.
    *   `GET /v1/openshell/attestation/:agent_id`: Verifies the hardware-bound identity of the OpenShell instance.
*   **Data Storage/State:** Policies are persisted in the `policy_registry` and synced to the OpenShell environment via the Adapter.

## 5. Alternatives Considered
*   **Pure Stdio Sandboxing:** Rejected because it doesn't provide the same level of granular, network-aware, and hardware-attested security as a dedicated runtime like OpenShell.
*   **Container-Only Isolation:** Rejected because standard containers lack "AI-Aware" guardrails (e.g., semantic analysis of output before egress).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The adapter mandates TPM-bound attestation for the OpenShell environment before any mission credentials are released.
*   **Observability:** Runtime violations are streamed back to the "Connectivity & Security Dashboard" for real-time alerting.

## 7. Evolutionary Changelog
*   **2026-06-18:** Initial Document Creation.
