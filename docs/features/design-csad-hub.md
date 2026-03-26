# Design Doc: Collective Swarm Anomaly Detection (CSAD)
**Status:** Draft
**Created:** 2026-05-29

## 1. Context and Scope
With the rise of "Hivenet" swarm attacks, individual agent behavioral monitoring is no longer sufficient. Attackers now use coordinated networks of autonomous agents to perform low-and-slow probes that evade single-point anomaly detection. MCP Any needs a collective defense layer that can analyze behavioral patterns across the entire agent mesh in real-time.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-millisecond, cross-agent behavioral analysis.
    * Detect coordinated probe patterns (e.g., distributed port scanning, credential stuffing).
    * Provide automated "Swarm Quarantine" capabilities.
* **Non-Goals:**
    * Replacing existing per-agent firewalls.
    * Performing deep packet inspection (L7) beyond semantic tool call metadata.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Detect and neutralize a coordinated "Hivenet" attack attempting lateral movement across the internal agent mesh.
* **The Happy Path (Tasks):**
    1. Multiple agents in the mesh start performing "low-risk" discovery calls that, when aggregated, reveal a network mapping attempt.
    2. CSAD Hub identifies the correlation between these disparate calls in sub-milliseconds.
    3. CSAD triggers a "Mesh Lockdown," revoking discovery capabilities for all agents in the affected mission scope.
    4. Architect receives an alert with a visual graph of the coordinated attack trace.

## 4. Design & Architecture
* **System Flow:**
    * Every tool call and A2A message is intercepted by the CSAD Middleware.
    * Metadata fragments (caller_id, tool_id, timestamp, mission_root) are pushed to a high-speed, in-memory "Swarm State Buffer."
    * A background "Pattern Matcher" (using sliding-window correlation) evaluates the buffer against known Hivenet signatures.
* **APIs / Interfaces:**
    * `csad.v1.QuarantineSwarm(mission_id)`: Revokes all capabilities for a specific mission.
    * `csad.v1.GetSwarmTrace(mission_id)`: Returns a graph of coordinated actions.
* **Data Storage/State:**
    * Ephemeral, lock-free ring buffer in memory for sub-millisecond processing.
    * Persistent audit logs in the core database for post-mortem analysis.

## 5. Alternatives Considered
* **Centralized Cloud Analysis:** Rejected due to latency. "Hivenet" attacks move too fast for cloud-roundtrips to be effective as a first-line defense.
* **Purely Local Rules:** Rejected as they cannot detect coordination between independent agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The CSAD Hub itself must be hardware-attested to prevent attackers from disabling the defense layer.
* **Observability:** Provides real-time "Mesh Health" metrics to the UI.

## 7. Evolutionary Changelog
* **2026-05-29:** Initial Document Creation.

### Update: 2026-05-30 - Teammate Impersonation Signatures
**Context:** Today's market sync revealed the emergence of "Teammate Impersonation" in horizontal meshes (Claude Code Agent Teams). Compromised teammates are injecting malicious instructions into the shared mailbox to bypass supervisor oversight.
**Architecture Adjustment:**
*   Implementing **Teammate Impersonation Signature Detection**. CSAD will now perform semantic analysis of inter-agent mailbox messages to detect "Instruction Splicing" and role-divergent commands.
*   Integrating with the `T2T Identity Rotation Provider` to flag any request using a stale or un-rotated identity fragment.
**Security Impact:** Prevents lateral movement within horizontal teams by ensuring that inter-teammate coordination remains bound to the mission-root intent and its authorized roles.

### Update: 2026-07-08 - AI Agent Insider Threat Mitigation
**Context:** The disclosure of the GTG-1002 swarm campaign highlights the shift toward agents being used as autonomous insider threats. These threats exploit high-trust system access to perform machine-speed lateral movement.
**Architecture Adjustment:**
*   Introducing **AI Agent Insider Threat Behavior Profiler**. This extension will use the Swarm State Buffer to monitor for "Machine-Speed Escalation" signatures (e.g., rapid credential probing or sequential high-value resource access from sibling agents).
*   Implementing **Sub-Millisecond Mission Revocation**. Upon detection of an insider threat pattern, CSAD will now broadcast a hardware-bound "Termination Signal" that clears session tokens for all agents within the affected mission scope across the entire mesh.
**Security Impact:** Neutralizes coordinated autonomous espionage attempts before they can achieve significant data exfiltration or system compromise.
