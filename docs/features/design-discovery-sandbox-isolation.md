# Design Doc: Discovery-Phase Sandbox Isolation
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Recent ecosystem audits of OpenClaw and Gemini CLI have identified the tool discovery phase as a critical attack vector. Discovery-time commands (e.g., `discoveryCommand` in MCP manifests) often execute with the same privileges as the primary agent, allowing malicious project-local configuration files to achieve Remote Code Execution (RCE) before any security gates are even initialized.

MCP Any must solve this by isolating all discovery-time execution in a zero-trust environment. This ensures that the "Pre-Flight" phase of agent initialization cannot be weaponized against the host system or used for token exfiltration.

## 2. Goals & Non-Goals
* **Goals:**
    * Isolate all `discoveryCommand` executions in an ephemeral, resource-constrained sandbox.
    * Mandate "Negative Discovery Attestation" to prove no unauthorized hooks were executed.
    * Provide a secure bridge for transferring validated discovery results back to the primary gateway.
* **Non-Goals:**
    * General-purpose tool sandboxing (covered by `design-sandbox-service.md`).
    * Full filesystem virtualization for discovery commands.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Initialize an agent in a potentially untrusted repository without risking host-level RCE.
* **The Happy Path (Tasks):**
    1. The agent triggers a discovery request for a local MCP server.
    2. MCP Any intercepts the `discoveryCommand` defined in the server manifest.
    3. The gateway spawns an ephemeral Docker container or gVisor sandbox with no host network access and read-only filesystem mounts.
    4. The command executes inside the sandbox and outputs the tool schema.
    5. MCP Any validates the output against the PNTD schema before exposing the tool to the agent.
    6. The sandbox is immediately destroyed.

## 4. Design & Architecture
* **System Flow:**
    Discovery Trigger -> Discovery Interceptor -> Ephemeral Sandbox Manager -> Isolated Execution -> Result Validator -> Capability Bus
* **APIs / Interfaces:**
    * Internal `SandboxProvider` interface for spawning ephemeral environments.
    * `DiscoverySession` token for tracking state between the sandbox and the gateway.
* **Data Storage/State:**
    No persistent state is stored within the sandbox. Discovery results are held in-memory by the gateway until validated.

## 5. Alternatives Considered
* **Native Process Isolation (chroot/nsenter):** Rejected due to the complexity of maintaining secure boundaries across different OS kernels. Docker/gVisor provides better uniformity and security guarantees.
* **Static Analysis of Discovery Commands:** Rejected because many discovery commands are dynamic scripts that cannot be accurately analyzed without execution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The sandbox has zero access to environment variables, local network, or host home directory. All communication with the gateway is via a restricted RPC channel.
* **Observability:** All sandbox logs and resource usage metrics are streamed to the `Config Sandbox Monitor` in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
