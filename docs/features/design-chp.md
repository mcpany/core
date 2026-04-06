# Design Doc: Cognitive Honeypot Provider (CHP)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move toward deeper autonomy and horizontal mesh coordination, the primary attack vector has shifted from direct tool-call exploitation to reasoning-phase subversion. Malicious agents or "jailbroken" specialist subagents can use complex chain-of-thought sequences to probe for mission-root constraints or exfiltrate state from the shared Blackboard.

The Cognitive Honeypot Provider (CHP) is designed to move MCP Any's defense posture from passive filtering to proactive interdiction. By injecting decoy state fragments ("traps") that appear highly sensitive or attractive to common prompt-injection/exfiltration patterns, CHP can detect and neutralize malicious reasoning before any real tool execution occurs.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically generate and inject "intent-attractive" decoy KV pairs into the Blackboard.
    * Detect unauthorized read/write access to honeypot fragments by subagents.
    * Trigger immediate hardware-locked capability revocation and mission-root alerts upon honeypot trigger.
    * Provide a feedback loop to the Stylometric Identity Verifier (SIV) for behavioral fingerprinting of attackers.
* **Non-Goals:**
    * Protecting host-level filesystem assets (handled by Shadow-FS and SII).
    * Validating tool output schemas (handled by DCG and WASM-BSH).
    * Providing long-term persistent storage (handled by SQLite/UEG).

## 3. Critical User Journey (CUJ)
* **User Persona:** Corporate Security Architect
* **Primary Goal:** Proactively detect an "Insider Threat" specialist agent attempting to exfiltrate private API keys from the swarm.
* **The Happy Path (Tasks):**
    1. The CHP generates a decoy Blackboard key: `internal.mission_root.auth_bypass_token`.
    2. A compromised specialist agent (e.g., a Database Expert) performs a `blackboard.list_keys` and identifies the decoy.
    3. The agent attempts a `blackboard.read` on the decoy key to "reason" about bypassing its constraints.
    4. The CHP interceptor detects the access on a marked "Trap" fragment.
    5. MCP Any immediately revokes the Database Expert's session token and notifies the Mission-Root supervisor.
    6. The supervisor agent receives a cryptographically signed "Interdiction Receipt" and halts the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root Initiation] --> B[CHP Generator]
        B --> C[Decoy Injection into Blackboard]
        D[Subagent Reasoning Loop] --> E{Blackboard Access}
        E -->|Legitimate| F[Real State]
        E -->|Honeypot Trigger| G[CHP Interceptor]
        G --> H[Capability Revocation]
        G --> I[Mission-Root Alert]
        G --> J[SIV Behavioral Logging]
    ```
* **APIs / Interfaces:**
    * `chp.GenerateTrap(context_hint string) -> TrapFragment`: Generates semantically relevant decoys.
    * `blackboard.read(key string)`: Extended to check the "IsTrap" metadata bit.
* **Data Storage/State:**
    * Honeypot metadata (Key, Entropy Profile, Trigger Policy) is stored in a secure, non-agent-readable kernel-bound buffer.

## 5. Alternatives Considered
* **Static Honeypots**: Rejected because modern LLMs can easily distinguish between static "safety test" strings and dynamic, mission-relevant state.
* **Pure Reasoning Analysis**: Rejected due to high latency and false-positive rates in complex valid reasoning. Honeypots provide a deterministic binary trigger.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: CHP must ensure that the generation of decoys does not lead to "Decoy Poisoning" where valid reasoning is broken.
* **Observability**: Every honeypot trigger must generate a high-fidelity trace including the full reasoning monologue leading to the access attempt.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
