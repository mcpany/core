# Design Doc: Instruction-Bound Tunnel Sovereignty (IBTS) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Ghost Proxying" exploits in Sovereign Node Tunneling (SNT) has revealed a critical flaw in session-based mesh security. Current models assume that once a P2P tunnel is authenticated by a parent agent, all subsequent traffic is authorized. Malicious subagents are now "ghosting" through these established tunnels to execute remote tools without parent attestation.

The Instruction-Bound Tunnel Sovereignty (IBTS) Broker is required to move mesh security from "Session-Bound" to "Instruction-Bound," ensuring that every individual tool call is cryptographically linked to the mission root and a verified command lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Enforce per-instruction hardware attestation for all tool calls traversing AMT tunnels.
    * Neutralize "Ghost Proxying" by requiring a Command Traceability (CTP) token for every remote execution.
    * Provide real-time validation of instruction lineage against the Mission-Root manifest.
* **Non-Goals:**
    * Replacing the underlying P2P encryption (handled by AMT).
    * Managing local (non-tunneled) tool calls (handled by ALSV).

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Security Auditor
* **Primary Goal:** Prevent a specialized subagent from using a parent's P2P tunnel to invoke an unauthorized remote tool.
* **The Happy Path (Tasks):**
    1. Parent agent establishes an AMT tunnel to a remote node.
    2. Subagent attempts to send a `run_shell` command through the tunnel.
    3. IBTS Broker intercepts the command and requests a CTP (Command Traceability Provider) token.
    4. IBTS verifies that the command lineage originates from an authorized mission-root branch.
    5. If the subagent attempts to "ghost" the command without a valid, parent-signed lineage, the IBTS Broker drops the packet and triggers a CSAD (Swarm Anomaly) alert.
    6. Authorized commands are forwarded to the remote node with the traceability token attached.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Command] --> B[IBTS Interceptor]
        B --> C{Verify CTP Token}
        C -->|Valid| D[Forward to AMT Tunnel]
        C -->|Invalid| E[Drop & Alert CSAD]
        D --> F[Remote IBTS Validator]
        F --> G[Execute Tool]
    ```
* **APIs / Interfaces:**
    * `ibts.AttestInstruction(command, lineage) -> InstructionToken`: Signs a specific instruction for mesh transit.
    * `ibts.ValidateInstruction(token) -> bool`: Verifies instruction integrity at the tunnel egress.
* **Data Storage/State:**
    * **Active Lineage Cache:** High-speed lookup for authorized command branches within the current mission.

## 5. Alternatives Considered
* **Short-Lived Session Keys:** Rejected because even a 1-second session is enough for a machine-speed agent to inject multiple unauthorized instructions. Instruction-level binding is the only way to achieve Zero Trust in autonomous meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Relies on the Command Traceability Provider (CTP) for non-repudiable lineage.
* **Observability:** Integrated with the "Mesh Command Sovereignty Dashboard" for real-time tracking of tunneled instructions.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
