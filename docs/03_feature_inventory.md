# Feature Inventory: MCP Any

## Current Backlog (P0/P1)
- **Policy Firewall**: Rego/CEL based hooking for tool calls.
- **HITL Middleware**: Suspension protocol for user approval flows.
- **Recursive Context Protocol**: Standardized headers for subagent inheritance.
- **Shared KV Store**: Embedded SQLite "Blackboard" tool for agents.

## Evolution: [2026-05-01] Updates

### Proposed Additions
- **Contextual Quorum (CQ) Hub**: (P0) Coordination service for multi-agent attestation, requiring a consensus of specialized subagents before high-risk tool execution.
- **Adaptive Intent Budgeting (AIB) Middleware**: (P1) Resource management layer that dynamically scales token and compute leases based on agent reasoning confidence.
- **Project-Local Snapshot (PLSS) Sync**: (P0) OS-level bridge for rapid environment snapshotted recovery, enabling speculative agent actions with near-instant rollbacks.

### Priority Shifts
- **S2S Trust Broker**: (Promoted to P0) Critical for neutralizing negotiation overhead in maturing inter-swarm coordination.
- **Consensus Tool Validation Hub**: (Re-affirmed P0) Evolving into the CQ Hub to support OpenClaw v2026.5.0 requirements.

## Evolution: [2026-04-30] Updates

### Proposed Additions
- **Mesh-Aware Blackboard Adaptor**: (P0) Transformation of the Shared KV Store into a graph-based intent mesh, enabling complex intent reconciliation for multi-agent swarms.
- **Kernel-Level Inode Pinning (KLIP) Middleware**: (P0) A kernel-resident security layer for the Shadow-FS that prevents symlink-racing (SIR) by pinning hardware Inodes to session-bound file handles.
- **UACO v3.0 S2S Trust Broker**: (P0) Multi-signature coordination service for Swarm-to-Swarm (S2S) task negotiation and identity management.

### Priority Shifts
- **Mesh-Aware Intelligence**: (Promoted to P0) Critical for reconciling conflicting intents in deep, heterogeneous swarms.
- **KLIP Enforcement**: (Promoted to P0) Designated as the primary defense against the evolved BoryptGrab SIR exploit.

## Evolution: [2026-04-29] Updates

### Proposed Additions
- **PII-Sovereign Context Scrubber**: (P0) Mandatory sanitization middleware for hybrid-cloud deployments, ensuring de-biometricization of context before cloud propagation.
- **ContextEngine Security Bridge**: (P0) A core integration service that maps OpenClaw ContextEngine lifecycle signals to MCP Any security policies for "Session-Bound" capability enforcement.
- **Speculative Integrity Quorum Hub**: (P1) A coordination service for the Shadow-FS that orchestrates multi-agent consensus for high-risk filesystem commits.

### Priority Shifts
- **De-biometricization Sanitizer**: (Promoted to P0) Critical for data sovereignty in hybrid reasoning loops.
- **Ephemeral Privilege Manager (EPM)**: (Re-affirmed P0) Now elevated with the requirement for "Lifecycle-Bound" revocation.

## Evolution: [2026-04-28] Updates

### Proposed Additions
- **Ephemeral Privilege Manager (EPM)**: (P0) Core security service that manages "Just-in-Time" privilege escalation for high-risk tools, neutralizing the "BoryptGrab" persistent access vector.
- **Shadow-FS Virtualization Adapter**: (P0) A virtualized filesystem overlay that allows agents to perform speculative edits in isolation, only committing to the host after successful validation.
- **De-biometricization Sanitizer**: (P1) Local context middleware that scrubs biometric and PII data before it is propagated to external LLM providers, ensuring local data sovereignty.

### Priority Shifts
- **Semantic Risk HITL Arbiter**: (Promoted to P0) Upgrading the HITL Middleware with context-aware risk assessment to reduce user approval fatigue.
- **LFTA ARL Middleware**: (Re-affirmed P0) Critical for immediate revocation of privileges during the ongoing "BoryptGrab" crisis.

## Evolution: [2026-04-27] Updates

### Proposed Additions
- **LFTA ARL Middleware**: (P0) A high-priority security listener that ingests Attestation Revocation Lists from trust-roots to provide sub-millisecond revocation of compromised trust leases.
- **Intent-Gated Shard Manager**: (P0) Advanced extension of the Context Sharding middleware that enforces cryptographic intent-alignment before mounting or unmounting specific context shards.
- **Adaptive Anchor Pruner**: (P1) Optimization service that implements the OpenClaw v2026.3.9 pruning logic, dynamically shedding irrelevant cognitive anchors to prevent context bloat.

### Priority Shifts
- **Cognitive Anchor Manager**: (Re-affirmed P0) Now elevated with the requirement for "Smart Pruning" to support deep, long-running agent swarms.
- **A2A Safety Proof Validator**: (Re-affirmed P0) Expanded to integrate with the LFTA ARL Middleware for real-time reputation and revocation checks.

---

## Evolution: [2026-04-26] Updates

### Proposed Additions
- **Multi-Hop Trust Relay**: (P0) Security middleware implementing LFTA v2.0 multi-hop trust delegation, allowing attestation strength to persist across deep agent swarms.
- **Cognitive Anchor Manager**: (P0) Extension for the ContextEngine Adapter that manages the lifecycle of immutable "Mission Anchors" to prevent semantic drift.
- **A2UI Interactive Delegation Bridge**: (P0) Hardened A2UI component for multi-agent task delegation, supporting rich user approvals for high-risk handoffs.

### Priority Shifts
- **A2A Session Persistence Middleware**: (Re-affirmed P0) Now integrates with the Multi-Hop Trust Relay for long-haul reasoning sessions.
- **ContextEngine Plugin Adapter**: (Re-affirmed P0) Expanded to support Cognitive Anchoring as a core sovereignty utility.

---

## Evolution: [2026-04-25] Updates

### Proposed Additions
- **A2A Session Persistence Middleware**: (P0) A core security service that manages token refresh and trust persistence for long-running A2A reasoning sessions, neutralizing "Session Decay."
- **DAP Enforcement for Pre-Flight Validator**: (P0) Mandatory extension for the Pre-Flight Sandbox Validator that enforces "Deterministic Absence Proofs" as a prerequisite for agent boot.

### Priority Shifts
- **ContextEngine Plugin Adapter**: (Re-affirmed P0) Now elevated to a critical requirement for supporting "Cognitive Anchoring" and "Context-Splicing" defense.
- **A2A Authenticated Handshake Provider**: (Re-affirmed P0) Now designated as the primary backend for the A2A Session Persistence Middleware.

---

## Evolution: [2026-04-24] Updates

### Proposed Additions
- **A2A Authenticated Handshake Provider**: (P0) Native security middleware implementing Gemini CLI v0.33.0 style HTTP authentication for all agent-to-agent remote communications and card discovery.
- **ContextEngine Plugin Adapter**: (P0) Core adapter for hosting OpenClaw-compatible ContextEngine plugins, enabling sovereignty-aware state management and intent protection.
- **Zero-Trust Discovery Gate**: (P1) Identity-bound access control layer for the A2A Messaging Hub that enforces "Auth-before-Discovery" for agent capabilities.

### Priority Shifts
- **A2A Messaging Hub**: (Re-affirmed P0) Now designated as the primary enforcement point for Authenticated Handshakes.
- **OpenClaw ContextEngine Lifecycle Adapter**: (Re-affirmed P0) Evolving into the ContextEngine Plugin Adapter for broader sovereignty support.

---

## Evolution: [2026-04-23] Updates

### Proposed Additions
- **OpenClaw ContextEngine Lifecycle Adapter**: (P0) A native middleware that implements OpenClaw's matured ContextEngine hooks, allowing MCP Any to act as the authoritative provider for context compression, summarization, and state persistence.
- **Absence Proof (DAP) Generator**: (P0) Extension for the Pre-Flight Sandbox Validator that generates signed manifests proving the non-existence of restricted configuration files, neutralizing CVE-2026-25725.
- **A2UI Secure Component Bridge**: (P0) A hardened rendering interface for declarative A2UI manifests, providing bi-directional, origin-locked state synchronization between agents and the user interface.

### Priority Shifts
- **RL Telemetry Provider**: (Promoted to P0) Now essential for providing high-frequency feedback tokens to OpenClaw-RL asynchronous training loops.
- **Deterministic Attestation Gateway**: (Re-affirmed P0) Expanded to include DAP as a mandatory boot requirement for all compliant agent environments.

---

## Evolution: [2026-04-22] Updates

### Proposed Additions
- **A2A Replay Guard**: (P0) Security middleware for the A2A Messaging Hub that enforces monotonic sequence nonces and session-bound validation to prevent task-proposal replay attacks.
- **Cognitive Fragment Reconciler**: (P1) Optimization service that manages the synchronization and reconciliation of "Encrypted Monologues" between specialized subagents and the A2UI Gateway.
- **Adaptive Context Compaction Engine**: (P0) Upgrade to the WebSocket Context Compactor that supports Gemini-style `x-gemini-reasoning-effort` headers for dynamic compression ratios.

### Priority Shifts
- **Agent-Aware Blackboard Isolation**: (Re-affirmed P0) Expanded to support "Cognitive Sovereignty" via hardware-bound encryption for subagent monologues.
- **A2UI Native Gateway**: (Re-affirmed P0) Now designated as the authoritative decryption point for "Encrypted Monologues" during user reviews.

---

## Evolution: [2026-04-21] Updates

### Proposed Additions
- **A2UI Native Gateway**: (P0) Secure bridge for the Agent-to-User Interface protocol, allowing agents to surface sandboxed, interactive UI fragments.
- **Deterministic Absence Proof (DAP) Provider**: (P0) Security service that generates signed proofs of non-existence for restricted project-local files to prevent malicious hook injection.
- **WebSocket Context Compactor**: (P1) Optimization middleware for WebSocket-first streaming that performs real-time context compaction for adaptive reasoning agents.

### Priority Shifts
- **ASH Consensus Broker**: (Re-affirmed P0) Now integrates with the A2UI Native Gateway for interactive user-in-the-loop consensus voting.
- **Deterministic Attestation Gateway**: (Re-affirmed P0) Expanded to include DAP as a mandatory boot requirement.

## Evolution: [2026-04-20] Updates

### Proposed Additions
- **ASH Consensus Broker**: (P0) Coordination service facilitating swarm-wide voting on reasoning paths and state re-alignment for Autonomous Self-Healing.
- **A2A Safety Proof Validator**: (P0) Mandatory validation layer for the A2A Messaging Hub that evaluates the "Safety Proof" of task proposals before delegation.
- **Origin-Locked Behavioral Attestation**: (P0) Security middleware that binds tool capabilities to a multi-factor token comprising cryptographically verified origin and Ghost Shell behavioral profile.

### Priority Shifts
- **Blackboard Versioning Hub**: (Re-affirmed P0) Now designated as the authoritative state provider for ASH Consensus voting.
- **Distributed Trust Lease Broker**: (Re-affirmed P0) Essential for sub-millisecond validation of A2A Safety Proofs in deep swarms.

---

## Evolution: [2026-04-19] Updates

### Proposed Additions
- **Distributed Trust Lease Broker**: (P0) A high-performance security utility implementing UACO v2.5 LFTA. Manages time-bound, hardware-attested trust leases to reduce per-call attestation latency.
- **Deep Packet Enforcement (DPPE) Middleware**: (P0) L4 network security layer that monitors DNS and ICMP traffic for "Binary Smuggling" exfiltration patterns (CVE-2026-31042).
- **Cognitive Drift Detector**: (P1) A monitoring service that evaluates subagent monologues against the mission-root to trigger ASH (Autonomous Self-Healing) re-alignment cycles.
- **Blackboard Versioning Hub**: (P0) Extends the Shared KV Store to support atomic checkpoints and swarm-wide rollbacks, facilitating autonomous self-healing.

### Priority Shifts
- **Atomic State Rollback Middleware**: Promoted to **P0**. Now a critical dependency for OpenClaw v2.8 ASH compliance.
- **Resident Integrity Monitor (RIM)**: (Re-affirmed P0) Expanded to act as the primary attestation source for the Distributed Trust Lease Broker.

---

## Evolution: [2026-04-18] Updates

### Proposed Additions
- **Foundation Governance Adapter**: (P1) A bridge and translation layer that implements the OpenClaw Foundation's neutral governance protocols for cross-framework agent coordination.
- **Continuous Sandbox Policy Verifier**: (P0) A security service that performs real-time validation of sandbox boundaries against the resident security policy, ensuring zero-drift throughout the agent lifecycle.
- **Unified Persistence Proof Broker**: (P1) A shared attestation utility that allows agents in a swarm to verify the persistence of their execution environment via a centralized hardware-bound proof.

### Priority Shifts
- **Resident Integrity Monitor (RIM)**: (Re-affirmed P0) Now elevated to the primary mechanism for supporting "Continuous Sandbox Persistence Proofs."
- **LFTA Trust Lease Manager**: Promoted to **P0**. Essential for scaling high-frequency attestation in deep swarms.

---

## Evolution: [2026-04-17] Updates

### Proposed Additions
- **LFTA Trust Lease Manager**: (P1) A performance-optimizing security middleware that manages "Trust Leases" for high-frequency agent tool calls, reducing hardware attestation overhead while maintaining mission integrity.
- **Swarm Consensus Alignment Broker**: (P0) A coordination service that periodically reconciles specialized subagent monologues against the parent's verified mission intent to prevent "Consensus Drift" in deep swarms.
- **Reactive Intent Arbitration Hub**: (P0) Advanced extension of the RIG that performs recursive deconstruction and validation of "Boundary Expansion" requests to block "Intent Smuggling" attempts.

### Priority Shifts
- **Resident Integrity Monitor (RIM)**: Promoted to **P0**. Now a critical requirement for "Sandbox Persistence Proofs" and continuous hardware-bound security.
- **Reactive Intent Gateway (RIG)**: Re-affirmed as **P0** and evolved into the Arbitration Hub.

---

## Evolution: [2026-04-14] Updates

### Proposed Additions
- **Context Sidecar Adapter**: (P1) Middleware that synchronizes state with external Context Engines (like OpenClaw v2026.3.7) via their native plugin interfaces, ensuring consistent "Intent-Bound" context across frameworks.
- **Delegation Attestation Layer**: (P0) A core security service that evaluates A2A task proposals against historical reputation and local policy to generate "Safety Proofs," reducing manual oversight requirements.
- **TPM-Bound Configuration Boot**: (P0) Extension of the Deterministic Boot Manifest to require hardware-bound (TPM) signatures for all project-local hooks and settings, neutralizing "Cloned Repository" attack vectors.

### Priority Shifts
- **A2A Messaging Hub**: (Re-affirmed P0) Expanded to include native support for the Delegation Attestation Layer.
- **Settings Injection Guard**: (Re-affirmed P0) Now mandates TPM-bound attestation for all security-critical configuration overrides.

---

## Evolution: [2026-04-12] Updates

### Proposed Additions
- **A2A Messaging Hub**: (P0) Native messaging hub for the A2A protocol, facilitating secure task delegation and coordination between disparate frameworks with integrated Zero-Trust policy enforcement.
- **Settings Injection Guard**: (P0) Active interception and validation layer for project-local agent configurations (e.g., `.claude/settings.json`) to neutralize configuration-based RCE and exfiltration.
- **Non-Existence Proof Generator**: (P0) Extension for the Deterministic Attestation Gateway to provide signed proofs of the absence of sensitive/malicious files in the project environment.

### Priority Shifts
- **Deterministic Attestation Gateway**: (Re-affirmed P0) Expanded to include Non-Existence Proofs as a mandatory "Deterministic Boot" prerequisite.
- **A2A Interoperability Layer**: (Re-affirmed P0) Transitioning from a bridge to a full Messaging Hub.

---

## Evolution: [2026-03-14] Updates

### Proposed Additions
- **Same-Origin Policy (SOP) Enforcer for MCP**: (P0) Middleware that validates `Origin` and `Sec-Fetch-Site` headers for all local requests to prevent cross-site hijacking (CVE-2026-25253).
- **Context Lifecycle Hooks**: (P1) Pluggable lifecycle hooks for context creation, compression, and retrieval, enabling custom "Intent-Preserving" strategies.
- **Semantic Boundary Detector**: (P0) A specialized scanning module for the Prompt Path Protection middleware that detects malicious instructions hidden in multimodal metadata (SVG, CSS).
- **Session-Resumption mTLS for Swarms**: (P1) Optimized mTLS transport that uses session tickets to reduce handshake latency in high-frequency A2A communication.

### Priority Shifts
- **OpenClaw ContextEngine Bridge**: Promoted to **P0**. Urgent need for interoperability to combat "Context Ghosting" in shared swarms.
- **"Safe-by-Default" Network Hardening**: (Re-affirmed P0) Expanded to include mandatory browser-origin validation for all local listeners.

### Deprecations / Monitoring
- **Unvalidated Local WebSockets**: Monitoring for total deprecation. All local WebSocket connections must provide a valid, allow-listed `Origin` header.

## Evolution: [2026-04-10] Updates

### Proposed Additions
- **Inference-Time Data Sanitizer (IDS)**: (P0) Semantic context governance middleware that sanitizes textual and multimodal data fragments using matured OpenClaw `ContextEngine` hooks.
- **Deterministic Attestation Gateway**: (P0) Extension of the Pre-Flight Sandbox Validator to provide signed environment manifests (including non-existence proofs) for "Deterministic Boot" compliance.
- **Origin-Locked Session Bridge**: (P0) Hardened WebSocket/HTTP session manager that binds tokens to cryptographically verified origins, patching CVE-2026-25253.

### Priority Shifts
- **Pre-Flight Sandbox Validator**: (Re-affirmed P0) Promoted to a mandatory "Deterministic Boot" prerequisite.
- **Cross-Framework Skill Reputation Engine**: (P1) Re-affirmed as the primary mechanism for swarm-based consensus on tool safety.

## Evolution: [2026-04-09] Updates

### Proposed Additions
- **Pre-Flight Sandbox Validator**: (P0) Core security service that generates a "Full-State Manifest" before agent execution, addressing environment-escape vulnerabilities like CVE-2026-25725.
- **Origin-Locked Session Bridge**: (P0) Hardened WebSocket/HTTP session manager that binds tokens to cryptographically verified origins.
- **Cross-Framework Skill Reputation Engine**: (P1) UAB-native middleware for sharing and validating tool reliability scores across agent swarms.

### Priority Shifts
- **Verified Skill Auction (VSA)**: (Re-affirmed P0) Expanded to integrate with the new Reputation Engine for real-time capability revoking.
- **Hardware-Linked Inode Pinning**: (Re-affirmed P0) Promoted as a mandatory requirement for the Pre-Flight Sandbox Validator.

## Evolution: [2026-04-08] Updates

### Proposed Additions
- **Pre-Flight Sandbox Validator**: (P0) Core security service that generates a "Full-State Manifest" before agent execution, addressing environment-escape vulnerabilities like CVE-2026-25725.
- **Cross-Framework Skill Reputation Engine**: (P1) UAB-native middleware for sharing and validating tool reliability scores across agent swarms.
- **Origin-Locked Session Bridge**: (P0) Hardened WebSocket/HTTP session manager that binds tokens to cryptographically verified origins.

### Priority Shifts
- **Verified Skill Auction (VSA)**: (Re-affirmed P0) Expanded to integrate with the new Reputation Engine for real-time capability revoking.
- **Hardware-Linked Inode Pinning**: (Re-affirmed P0) Promoted as a mandatory requirement for the Pre-Flight Sandbox Validator.

---

## Evolution: [2026-04-07] Updates

### Proposed Additions
- **Verified Skill Auction (VSA)**: (P0) Integrating the DCA Auction Broker with skill attestation to ensure only verified agents can bid on sensitive tasks.
- **Social-Agent Privacy Sandbox**: (P1) Middleware to prevent parent-context reconstruction during interactions on multi-agent social platforms (e.g., Moltbook).
- **Federated Reputation Quorum Node**: (P1) Peer-to-peer node for collective tool safety attestation, mitigating "ClawHavoc" style registry attacks.

### Priority Shifts
- **DCA Negotiation Guard**: (Re-affirmed P0) Expanded to support the new VSA protocol and mitigate negotiation exhaustion.
- **Attested Discovery Authority**: (Re-affirmed P0) Promoted as the mandatory gate for all marketplace-sourced skills.

---

## Evolution: [2026-02-23] Updates

### Proposed Additions
- **Environment Bridging Middleware**: (P1) Bridge between cloud-sandboxed agents (e.g., Claude Code Sandbox) and local MCP Any tools. Enables seamless state transfer.
- **Machine-Checkable Security Contracts**: (P1) Declarative security models for tools that can be verified by automated agents (inspired by OpenClaw).
- **Zero-Trust Subagent Scoping**: (P0) Capability-based tokens that restrict subagents to a specific "intent-scope" of a parent's permissions.

### Priority Shifts
- **Recursive Context Protocol**: Promoted from **P1** to **P0**. Essential for modern agent swarms to prevent state loss.
- **Shared KV Store**: Promoted from **P1** to **P0**. Critical for coordinating multi-agent actions in complex workflows.

### Deprecations / Monitoring
- *None today.*

---

## Evolution: [2026-02-24] Updates

### Proposed Additions
- **Advanced Multi-Agent Session Management**: (P0) A session-aware middleware that tracks tool state and handoffs between multiple specialized agents.
- **Unified MCP Discovery Service**: (P1) Automated discovery and registry for local and remote MCP servers (Stdio, HTTP, FastMCP).
- **Session-Bound State Persistence**: (P1) Ensuring that multi-agent "long-running" tasks maintain state across tool calls and agent switches.

### Priority Shifts
- **Policy Firewall**: Promoted to **P0** to support secure "Zero Trust" subagent isolation as ecosystems become more complex.

---

## Evolution: [2026-02-25] Updates

### Proposed Additions
- **On-Demand Discovery Middleware (Lazy-MCP)**: (P0) Implements similarity-based tool searching to prevent context pollution. Essential for massive (100+) tool libraries.
- **MCP Provenance Attestation**: (P1) Cryptographic verification of MCP server origins to prevent "Clinejection"-style supply chain attacks.
- **Slash-Command Bridge for Gemini**: (P1) Automatic mapping of MCP prompts to native Gemini CLI slash commands.

### Priority Shifts
- **Environment Bridging Middleware**: Promoted from **P1** to **P0**. The need for secure "Local-to-Cloud" tool bridging is increasing with more agents running in remote sandboxes.
- **Supply Chain Integrity Guard**: (New entry but P0 priority) High urgency due to recent ecosystem exploits.

### Deprecations / Monitoring
- **Upfront Tool Schema Pushing**: Monitoring for deprecation in favor of Lazy-Discovery.

---

## Evolution: [2026-02-26] Updates

### Proposed Additions
- **A2A Interop Bridge (Pseudo-MCP)**: (P0) Allows agents to interact with other agent frameworks using the A2A protocol, exposed as standard MCP tools.
- **Federated MCP Node Peering**: (P1) Secure discovery and proxying of tools across distributed MCP Any instances.
- **Cost & Latency Telemetry Middleware**: (P1) Automatically injects performance metadata into tool schemas to enable resource-aware agent reasoning.

### Priority Shifts
- **MCP Provenance Attestation**: Promoted to **P0** as it is a prerequisite for secure Federated MCP peering.
- **Lazy-MCP Middleware**: Promoted to **P0** (Already P0, but re-affirming importance for Federated Tool Mesh).

### Deprecations / Monitoring
- **Static Tool Schemas**: Moving towards dynamic, metadata-rich schemas that include real-time performance metrics.

---

## Evolution: [2026-02-28] Updates

### Proposed Additions
- **"Safe-by-Default" Network Hardening**: (P0) Transition to local-only default bindings for all services. Requires explicit MFA/Attestation for remote exposure.
- **A2A Stateful Residency (Stateful Buffer)**: (P0) MCP Any acts as a persistent mailbox for A2A messages, enabling reliable communication between agents with intermittent connectivity.
- **Provenance-First Discovery (Attested Discovery)**: (P1) Automatic filtering of MCP servers based on cryptographic signatures and community reputation scores.

---
---

## Evolution: [2026-03-09] Updates

### Proposed Additions
- **Project Configuration Security Guard**: (P0) Validating proxy for project-local agent configurations (e.g., `.claude/settings.json`) to prevent RCE via malicious hooks.
- **Agent-Aware Blackboard Isolation**: (P0) Implements row-level security for the Shared KV Store, ensuring agents can only access state within their assigned "Intent Scope."
- **Detached Sandbox for Automated Hooks**: (P1) Isolated execution environment for automated tool sequences, preventing unauthorized host access.

### Priority Shifts
- **Shared KV Store (Blackboard)**: Re-affirmed as **P0** with new mandatory security isolation requirements.
- **Policy Firewall**: Promoted to **P0** (Already P0, but expanded to include "Project-Local Config Validation").

### Deprecations / Monitoring
- **Unvalidated Project-Local Configs**: Monitoring for total deprecation. All local configs must be attested via MCP Any before ingestion by agents.

---

## Evolution: [2026-03-10] Updates

### Proposed Additions
- **Sandbox-as-a-Service for Config Hooks**: (P0) A natively managed, ultra-lightweight execution environment for approved hooks found in project-local settings.
- **Project Configuration Drift Detection**: (P1) Background monitor that alerts the user if a project-local configuration file is modified (e.g., via `git pull`), requiring re-attestation of any hooks.
- **Intent-Bound Context Isolation**: (P0) Cryptographic enforcement that prevents subagents from accessing state or tools outside their explicitly assigned "Intent-Scope."

### Priority Shifts
- **Detached Sandbox for Automated Hooks**: Promoted from **P1** to **P0**. Urgent requirement to mitigate RCE vulnerabilities discovered in the ecosystem.
- **A2A Interop Bridge**: Re-affirmed as **P0** to support secure state handoffs in multi-agent swarms.

### Deprecations / Monitoring
- **Implicit Hook Execution**: All "hooks" or "auto-exec" commands in configurations are now **Deprecated**. They must be explicitly moved to an "Attested Hooks" registry.

---

## Evolution: [2026-03-11] Updates

### Proposed Additions
- **Project-Local Config Attestation Engine**: (P0) A core service that intercepts and verifies cryptographic signatures on project-local configuration files.
- **Base-URL Hijack Protection (Exfiltration Guard)**: (P0) A middleware that enforces a strict "Allow-List" for LLM base URLs, preventing silent redirection of API traffic.
- **Active Config Rewriter**: (P1) A daemon that monitors agent configuration files and automatically reverts unauthorized changes to security-critical fields.

---

## Evolution: [2026-03-12] Updates

### Proposed Additions
- **Verified Skill Registry**: (P0) A security-first marketplace/registry for agent skills, requiring behavioral profiling and cryptographic signing before installation.
- **Offline-First Resilient Proxy**: (P1) A hardened gateway that handles complex proxy configurations and provides a stable LLM interface for air-gapped or restricted environments.
- **MFA for Project-Local Hooks**: (P0) Extends the HITL Middleware to require multi-factor attestation for any executable hook found in project configurations.

---

## Evolution: [2026-03-13] Updates

### Proposed Additions
- **OpenClaw ContextEngine Bridge**: (P1) A middleware that enables MCP Any to synchronize state with OpenClaw's new pluggable ContextEngine.
- **Prompt Path Protection Middleware**: (P0) Real-time scanning of tool outputs for "Indirect Prompt Injection" patterns to prevent agent hijacking.
- **Critical Skill Simulation (Dry-Run 2.0)**: (P1) Advanced "what-if" analysis for skills that simulates their impact on sensitive data before they are executed.
- **Swarm Behavioral Baseline**: (P1) Monitoring tool to establish a "normal" behavior pattern for agent swarms and alert on anomalies.

### Priority Shifts
- **Verified Skill Registry**: Re-affirmed as **P0** following the "ClawHavoc" malicious skill crisis.
- **A2A Interop Bridge**: Re-affirmed as **P0** to support the industry shift towards "Agentic Swarms."

### Deprecations / Monitoring
- **Direct Agent-to-LLM Communication**: Monitoring for deprecation in favor of **Exfiltration-Resistant Transport** (Proxied via MCP Any).
- **Unsigned/Unverified Skills**: Moving towards a default-block policy for any skill not present in the Verified Skill Registry.

---

## Evolution: [2026-03-14] Updates

### Proposed Additions
- **Same-Origin Policy (SOP) Enforcer for MCP**: (P0) Middleware that validates `Origin` and `Sec-Fetch-Site` headers for all local requests to prevent cross-site hijacking (CVE-2026-25253).
- **Context Lifecycle Hooks**: (P1) Pluggable lifecycle hooks for context creation, compression, and retrieval, enabling custom "Intent-Preserving" strategies.
- **Semantic Boundary Detector**: (P0) A specialized scanning module for the Prompt Path Protection middleware that detects malicious instructions hidden in multimodal metadata (SVG, CSS).
- **Session-Resumption mTLS for Swarms**: (P1) Optimized mTLS transport that uses session tickets to reduce handshake latency in high-frequency A2A communication.

### Priority Shifts
- **OpenClaw ContextEngine Bridge**: Promoted to **P0**. Urgent need for interoperability to combat "Context Ghosting" in shared swarms.
- **"Safe-by-Default" Network Hardening**: (Re-affirmed P0) Expanded to include mandatory browser-origin validation for all local listeners.

### Deprecations / Monitoring
- **Unvalidated Local WebSockets**: Monitoring for total deprecation. All local WebSocket connections must provide a valid, allow-listed `Origin` header.

---

## Evolution: [2026-03-15] Updates

### Proposed Additions
- **Call-Graph Loop Monitor**: (P0) Middleware to detect and prevent recursive "M2M" tool loops that cause resource exhaustion.
- **Signed Context Chain Protocol**: (P0) Cryptographic signing of subagent requests to prevent identity spoofing (CVE-2026-28190).
- **Universal Agent Bus (UAB) Adapter**: (P1) Native support for the UAB protocol, enabling seamless task handoffs between OpenClaw and AutoGen frameworks.

---

## Evolution: [2026-03-16] Updates

### Proposed Additions
- **Browser-Origin Validation Middleware**: (P0) Mandatory validation of `Origin` and `Sec-Fetch-Site` headers for all local listeners to mitigate cross-site hijacking (CVE-2026-25253).
- **UAB Task Delegation Bridge**: (P1) Extension of the A2A bridge to support UAB-native task cards and authenticated discovery.
- **Cross-Agent Loop Circuit Breaker**: (P0) Real-time monitoring of inter-agent call graphs to prevent "Spiral of Death" loops across framework boundaries.
- **Relational Identity Provider**: (P1) A core service that maps and verifies agent identities between disparate frameworks (e.g., OpenClaw, Gemini CLI).

### Priority Shifts
- **Signed Context Chain Protocol**: Re-affirmed as **P0** with expanded requirements for UAB compatibility.
- **"Safe-by-Default" Network Hardening**: (Re-affirmed P0) Now includes mandatory Browser-Origin enforcement for all adapters.

### Deprecations / Monitoring
- **Implicit Local Trust**: All listeners must now explicitly validate request origins. Standard `localhost` binding without header checks is now **Deprecated**.

---

## Evolution: [2026-03-17] Updates

### Proposed Additions
- **Local-Loopback Rate Limiter**: (P0) Mandatory rate limiting for all `127.0.0.1` and `::1` connections to prevent brute-force attacks on gateway credentials.
- **Behavioral Skill Burn-In Sandbox**: (P1) An isolated environment where new skills are profiled for "Delayed Payload" behaviors before being promoted to "Trusted" status.
- **UAB Authenticated Task Delegation Bridge**: (P0) Full implementation of UAB v1.2 "Authenticated Task Cards" for secure cross-framework delegation.
- **Local Security Audit Log**: (P1) Detailed logging of all local connection attempts, including origin headers and authentication success/failure rates.

### Priority Shifts
- **Universal Agent Bus (UAB) Adapter**: Promoted to **P0**. Essential for cross-framework agentic coordination.
- **Verified Skill Registry**: (Re-affirmed P0) Expanded to include Behavioral Profiling requirements.

### Deprecations / Monitoring
- **Unthrottled Local Access**: All local interfaces must now implement rate limiting. Unthrottled loopback access is now **Deprecated**.

---

## Evolution: [2026-03-18] Updates

### Proposed Additions
- **Local Listener Origin Enforcement**: (P0) Mandatory `Origin` and `Sec-Fetch-Site` validation for all local API/WebSocket listeners to prevent cross-site hijacking.
- **Recursive Depth-Limit Middleware**: (P0) Advanced call-graph monitoring to detect and block infinite tool-calling loops across different agents.
- **UAB Authenticated Task Delegation Core**: (P0) Full implementation of UAB task card verification, ensuring all cross-framework delegations are authenticated.
- **Lineage-Aware Context Signing**: (P1) Cryptographic signing of the entire context chain to prevent subagent identity spoofing.

---

## Evolution: [2026-03-19] Updates

### Proposed Additions
- **UACO-Native Coordination Middleware**: (P0) Full implementation of the Universal Agent Coordination Protocol for task negotiation, bidding, and stateful handoffs.
- **Unified RL Feedback Telemetry Bridge**: (P1) Middleware for collecting and normalizing agent performance and conversation feedback for RL training loops (e.g., OpenClaw-RL).
- **Enterprise Policy Sync Engine**: (P1) Core service for synchronizing security policies and allowed-origin lists from a centralized enterprise management server.

---

## Evolution: [2026-03-20] Updates

### Proposed Additions
- **Ephemeral Workspace Trust Middleware**: (P0) A session-bound attestation service that translates desktop-level trust tokens into persistent agent capabilities.
- **Blackboard Integrity Validator**: (P0) Cryptographic validation for all Shared KV Store operations, ensuring state lineage and intent-bound isolation.
- **UACO Bid Profiling Engine**: (P1) Behavioral monitoring service that evaluates agent bids against historical performance and safety baselines to prevent "Task Card Shadowing."
- **Config Smuggling Scanner**: (P1) Specialized scanner for project-local configurations that detects malicious instructions hidden in binary/metadata blobs.

### Priority Shifts
- **A2A Interop Bridge**: Promoted to **P0**. With UACO maturation, the bridge is now critical for multi-agent task negotiation.
- **Project Configuration Security Guard**: (Re-affirmed P0) Expanded to include support for Enterprise-Managed policy overrides.

### Deprecations / Monitoring
- **Framework-Specific Feedback Logs**: Monitoring for deprecation. Feedback should be normalized via the Unified Telemetry Bridge.

---

## Evolution: [2026-03-21] Updates

### Proposed Additions
- **Content-Addressable Config (CAC) Validator**: (P0) A core security service that enforces hash-based validation for all executable hooks and settings, preventing "Binary Smuggling."
- **UACO v1.5 RCC Validator**: (P0) Implementation of Resource Capability Claims to verify agent toolsets and security posture during task bidding.
- **DNS/ICMP Exfiltration Monitor**: (P1) L4 telemetry middleware to detect and block non-HTTP exfiltration attempts by compromised agents.
- **Hardware-Bound Trust Continuity**: (P1) Extension for the Ephemeral Workspace Trust Middleware that uses TPM/Secure Enclave signatures to persist trust for headless agents.

---

## Evolution: [2026-03-22] Updates

### Proposed Additions
- **UACO Agentic SLA Middleware**: (P0) Enforcement layer for resource contracts (token budget, reasoning time) during UACO task delegation.
- **Federated Policy Synchronizer**: (P1) A secure bus for synchronizing CAC hashes and allowed-origin lists across multiple MCP Any instances.
- **Ghost Shell Execution Mode**: (P0) Isolated, instrumented profiling environment for un-attested hooks, providing behavioral insights before attestation.

### Priority Shifts
- **UACO v1.5 RCC Validator**: Re-affirmed as **P0**. Essential foundation for the new SLA middleware.
- **Shared KV Store (Blackboard)**: (Re-affirmed P0) Expanded to support "SLA-Aware State Locking" to prevent resource-heavy contention.

### Deprecations / Monitoring
- **Unbounded Task Delegation**: Moving toward total deprecation. All UACO delegations must eventually include a resource contract (SLA).

---

## Evolution: [2026-03-23] Updates

### Proposed Additions
- **Proof-of-Intent (PoI) Validator**: (P0) Middleware that implements UACO v1.7 headers to verify that tool calls align with cryptographically signed session intents.
- **Binary State Handoff (BSH) Gateway**: (P1) High-performance binary transport for agent state to mitigate "Token Storms" and JSON overhead.
- **Multi-Signature Skill Attestation**: (P0) Verification mechanism for dynamic skill grafting, requiring signatures from both framework and user policy to prevent "Skill-Squatting."

### Priority Shifts
- **UACO-Native Coordination Middleware**: Re-affirmed as **P0**. Urgent update required to support v1.7 PoI and combat Context-Mirroring.
- **Verified Skill Registry**: (Re-affirmed P0) Expanded to include real-time attestation for dynamic grafting.

### Deprecations / Monitoring
- **JSON-only State Handoffs**: Monitoring for deprecation in favor of **BSH** for high-frequency agent swarms.

---

## Evolution: [2026-03-24] Updates

### Proposed Additions
- **Relational PoI Validator**: (P0) Extends PoI validation to verify the entire "Intent Chain," ensuring subagents cannot be coerced into actions outside the parent's verified goal.
- **BSH State Buffer**: (P1) High-speed memory-mapped buffer for binary state handoffs between agents to minimize context transfer latency.
- **Ghost Shell Hook Profiler**: (P0) Instrumented sandbox for behavioral profiling of un-attested configuration hooks, detecting "Binary Smuggling" before host execution.

### Priority Shifts
- **Binary State Handoff (BSH) Gateway**: Promoted from **P1** to **P0**. Urgent requirement to solve the "Token Storm" crisis in deep swarms.
- **Ghost Shell Execution Mode**: Re-affirmed as **P0**. Critical security defense against malicious project-local hooks.

---

## Evolution: [2026-03-25] Updates

### Proposed Additions
- **WASM-BSH State Sanitizer**: (P0) Pluggable WASM sandbox for the BSH Gateway that validates and sanitizes binary context during handoffs.
- **Zero-Copy Shared Memory Transport**: (P0) High-performance transport layer for BSH using memory-mapped regions to eliminate serialization overhead.
- **Recursive Intent Delegation (RID) Validator**: (P0) UACO v1.8 compliant middleware for enforcing depth-limited intent mutations.
- **Predictive Resource Locking**: (P1) Middleware that pre-emptively locks Blackboard keys based on the signed intent of upcoming UACO tasks.

### Priority Shifts
- **Relational PoI Validator**: Re-affirmed as **P0**. Critical foundation for supporting UACO v1.8 RID.
- **Ghost Shell Hook Profiler**: Re-affirmed as **P0**. Expanded to include "WASM-BSH Pattern Matching" to detect malicious state transformation logic.

---

## Evolution: [2026-03-26] Updates

### Proposed Additions
- **Modular Context Hook Adapter**: (P0) A bridge that maps MCP Any's internal state to the pluggable lifecycle hooks of external frameworks (e.g., OpenClaw ContextEngine).
- **RID Mutation Boundary Enforcer**: (P0) Middleware that validates UACO v1.8 tokens, ensuring subagents cannot exceed their assigned delegation depth or mutate intents beyond authorized boundaries.
- **WASM-BSH Active Sanitizer**: (P0) Integrated WASM sandbox for the BSH Gateway that performs schema-based validation on binary context buffers during handoffs.

---

## Evolution: [2026-03-27] Updates

### Proposed Additions
- **Live Context Sharding Middleware**: (P0) Core service for managing the lifecycle of granular, addressable context shards. Enables on-demand mounting/unmounting of sub-state.
- **Consensus Tool Validation Gateway**: (P0) Distributed HITL middleware that requires multi-agent attestation for high-risk tool calls.
- **PNTD Discovery Provider**: (P1) Implementation of Protocol-Neutral Task Discovery to unify capability mapping across MCP, gRPC, and UACO transports.
- **Shard-Aware State Buffer**: (P1) Optimized BSH buffer extension that supports addressable memory regions for individual context shards.

### Priority Shifts
- **UACO-Native Coordination Middleware**: (Re-affirmed P0) Expanded to support RID Parental Overrides and Consensus Tokens.
- **A2A Interop Bridge**: (Re-affirmed P0) Now a critical transport for Consensus-Based Tool Validation.

### Deprecations / Monitoring
- **Single-Agent HITL for High-Risk Actions**: Monitoring for deprecation in enterprise profiles in favor of **Consensus-Based Validation**.
- **Monolithic Context Handoffs**: Moving toward deprecation for deep swarms in favor of **Context Sharding**.

---

## Evolution: [2026-03-28] Updates

### Proposed Additions
- **Atomic State Rollback Middleware**: (P0) Enables swarm-wide state checkpoints and rollbacks for the Blackboard and Context Shards.
- **UACO-MAQ Consensus Gateway**: (P0) Support for UACO v1.9 Multi-Agent Quorum, allowing cross-framework approval tokens for high-risk actions.
- **Session-Bound Fast-Path Attestation**: (P1) Hardware-accelerated attestation for sub-calls within a verified mission session.
- **Context Smearing Scanner**: (P1) Binary-level inspection for BSH fragments to detect malicious "Ghost Fragments."

### Priority Shifts
- **Consensus Tool Validation Gateway**: Re-affirmed as **P0**. Urgent need to align with UACO v1.9 MAQ.
- **WASM-BSH State Sanitizer**: (Re-affirmed P0) Expanded to include detection of "Context Smearing" patterns.

### Deprecations / Monitoring
- **Legacy HITL Approval Tokens**: Monitoring for deprecation in favor of UACO-MAQ compliant multi-signature tokens.

---

## Evolution: [2026-03-29] Updates

### Proposed Additions
- **Proactive State Alignment (PSA) Middleware**: (P1) Background service for continuous synchronization of agent-local state with the global Blackboard.
- **UACO v2.0 RIS Validator**: (P0) Implementation of Relational Intent Scoping to prevent Identity Shadowing via hierarchical intent trees.
- **Hardware-Bound Attestation Provider (HAFP)**: (P0) Native integration with TPM/Secure Enclave for zero-latency mission validation.
- **Context Pinning Middleware**: (P1) Implements immutable prompt segments to neutralize Context Smearing attacks.

---

## Evolution: [2026-03-31] Updates

### Proposed Additions
- **UACO v2.2 Intent Barrier Middleware**: (P0) Synchronization engine for parallel sub-intents to prevent race conditions in the Blackboard.
- **Inode-Aware Symlink Validator**: (P0) Security middleware that performs recursive symlink resolution and inode validation for all project-local configurations.
- **Federated Discovery Quorum (FDQ) Node**: (P1) Peer-to-peer discovery service that requires multi-node attestation for new tool beacons.
- **Parallel Intent Branch Manager**: (P0) Implements "Snapshot-and-Merge" logic for parallel agent branches, ensuring deterministic state reconciliation.

### Priority Shifts
- **Shared KV Store (Blackboard)**: Re-affirmed as **P0**. Expanded to include support for "Branch-Aware State Isolation" and "Merge Conflict Resolution."
- **UDP Beacon Discovery Listener**: Promoted from **P1** to **P0**. Essential prerequisite for the new Federated Discovery Quorum.
- **Inode-Aware Symlink Validator**: Re-prioritized to **P0**. Critical for mitigating project-local exfiltration vectors.

---

## Evolution: 2026-04-01 Updates

### Proposed Additions
- **Reasoning-Bound Context Shifter**: (P0) Context management middleware that synchronizes dynamic shifting logic across frameworks.
- **Path Normalization Engine (NaaS)**: (P0) Centralized service for OS-agnostic path normalization to prevent symlink and traversal escapes.
- **Optimistic Capability Loading Middleware**: (P1) Predictive tool registry that handles Gemini-style optimistic loading with built-in TOCTOU protection.

### Priority Shifts
- **Inode-Aware Symlink Validator**: (Re-affirmed P0) Urgent requirement to address "Normalization Fatigue" in project-local config parsing.

### Deprecations / Monitoring
- **OS-Specific Path Joins**: Monitoring for deprecation in favor of the **Path Normalization Engine**.
- **Static Discovery Quorums**: Moving toward **Optimistic Loading** with background attestation.

---

## Evolution: [2026-03-30] Updates

### Proposed Additions
- **UACO v2.1 IPSC Middleware**: (P0) Implementation of Intent-Preserving Self-Correction to prevent "Cognitive Lock" refinement loops.
- **Continuous BSH Integrity Monitor**: (P0) Real-time WASM-based monitor for Binary State Handoffs to detect "Ghost Fragment Mutation" during self-correction.
- **UDP Beacon Discovery Listener**: (P1) High-speed reactive listener for Gemini-style Capability Beacons to reduce discovery noise.
- **Correction Budget Controller**: (P1) Resource management middleware that enforces token and cycle limits on agent self-correction loops.

### Priority Shifts
- **WASM-BSH State Sanitizer**: Re-affirmed as **P0**. Expanded to include "Dormant Fragment" detection as part of GFM defense.
- **PNTD Discovery Provider**: Promoted from **P1** to **P0**. Essential foundation for the new Beacon-First Discovery Hub.

### Deprecations / Monitoring
- **Unbounded Self-Correction**: Moving toward total deprecation. All self-correction loops must eventually be bound by an IPSC token and Correction Budget.

---

## Evolution: [2026-04-02] Updates

### Proposed Additions
- **Speculative Execution Guard**: (P0) Middleware that manages "Shadow State" for speculative tool calls, ensuring rollbacks on attestation failure.
- **Inode-Pinning Middleware**: (P0) Hardware-bound file handle protection that prevents symlink-racing and TOCTOU escapes in project configs.
- **Consensus Delegation Gateway**: (P1) Implementation of "Delegated Authority" models where trusted monitor agents can authorize time-critical tasks.
- **Branch-Purity Blackboard Validator**: (P0) Integrity layer for the Shared KV Store to prevent "Branch Contamination" between divergent reasoning paths.

---

## Evolution: [2026-04-03] Updates

### Proposed Additions
- **Active Subagent Reaper**: (P0) Lifecycle monitor that forcefully terminates orphaned or "Ghost" subagent sessions when their parent intent branch is pruned.
- **Tool Metadata Sanitizer**: (P0) Security middleware that scans JSON schemas and tool descriptions for imperative instructions (Context Poisoning) before LLM ingestion.
- **DCA Auction Broker**: (P1) High-speed negotiation bus for the "Distributed Capability Auction" protocol, managing agent tool bidding.
- **Subagent Heartbeat Provider**: (P1) Standardized heartbeat protocol for subagents to report liveness and intent alignment to the Reaper.

### Priority Shifts
- **Speculative Execution Guard**: Re-affirmed as **P0**. Now requires integration with the Subagent Reaper to ensure speculative "Zombies" are purged.
- **Branch-Purity Blackboard Validator**: (Re-affirmed P0) Expanded to detect "Ghost State" injected by non-terminated subagents.

### Deprecations / Monitoring
- **Unmanaged Subagent Lifecycle**: Moving toward total deprecation. All subagent sessions must be bound to a supervised intent lifecycle.
- **Unsanitized Structural Metadata**: Monitoring for deprecation. Tool schemas will require "Safe Metadata" attestation.

---

## Evolution: [2026-04-04] Updates

### Proposed Additions
- **DCA Negotiation Guard**: (P0) Hardware-accelerated (HAN) broker for subagent bidding, mitigating "Negotiation Exhaustion."
- **Metadata Provenance Engine**: (P0) Verification service for structural metadata lineage, ensuring tool schemas are cryptographically signed.
- **Unified Lifecycle Bridge**: (P1) Standardized commit/rollback middleware for cross-framework (OpenClaw/AutoGen) lifecycle synchronization.

### Priority Shifts
- **Tool Metadata Sanitizer**: Promoted from **P1** to **P0**. Critical for mitigating CVE-2026-42001.
- **DCA Auction Broker**: Re-affirmed as **P0** (Already P0, but expanded to include HAN requirements).

---

## Evolution: [2026-04-05] Updates

### Proposed Additions
- **RL Telemetry Provider**: (P1) Standardized middleware for exporting tool performance and feedback metrics to agent training frameworks (e.g., OpenClaw-RL).
- **Attested Discovery Authority**: (P0) Cryptographic identity broker for local MCP servers, providing the "Trust Verification" required by Claude Code.
- **Optimistic Execution Gate**: (P0) Implementation of speculative context loading for tools, synchronized with background discovery quorums.

### Priority Shifts
- **Unified RL Feedback Telemetry Bridge**: (Re-affirmed P1) Now a core strategic requirement to support OpenClaw-RL v1.
- **Provenance-First Discovery**: (Promoted to P0) Critical for satisfying the new Claude Code trust verification requirements.

### Deprecations / Monitoring
- **Implicitly Trusted Local Discovery**: Moving toward total deprecation. All local tool discovery must eventually be backed by an Attested Discovery signal.

---

## Evolution: [2026-04-06] Updates

### Proposed Additions
- **Structural Metadata Sanitizer Middleware**: (P0) A security service that treats tool descriptions and schemas as untrusted input, scanning them for imperative instructions or "Context Poisoning" patterns.
- **Hardware-Linked Inode Pinning**: (P0) Extends path validation to include hardware-bound Inode checks, preventing TOCTOU races in project-local configurations.
- **Speculative Auction Broker (SAB)**: (P1) High-speed negotiation bus for Gemini-style "Intent Probability" bidding in agent swarms.

---

## Evolution: [2026-04-11] Updates

### Proposed Additions
- **A2A Interoperability Layer**: (P0) Native messaging hub implementation for the Agent2Agent (A2A) protocol, facilitating secure task delegation and coordination between disparate frameworks.
- **Deterministic Environment Attestation Gateway**: (P0) Advanced pre-execution security service that generates signed environment manifests, including non-existence proofs for restricted configuration hooks.
- **Structured Context Propagation Middleware**: (P1) Implementation of emerging context propagation standards to ensure rich, structured contextual data (trace IDs, session IDs) flows securely across the agentic lifecycle.

### Priority Shifts
- **Tool Metadata Sanitizer**: Promoted to **P0**. Urgent requirement to address CVE-2026-45201.
- **DCA Negotiation Guard**: (Re-affirmed P0) Expanded to support the new Speculative Auction Broker (SAB) protocol.

### Deprecations / Monitoring
- **Implicitly Trusted Tool Schemas**: Monitoring for total deprecation. All structural metadata must eventually pass through the Sanitizer.

---

## Evolution: [2026-04-13] Updates

### Proposed Additions
- **CLAW-10 Compliance Mapper**: (P1) Middleware that maps MCP Any's internal security state to the CLAW-10 Enterprise Evaluation Matrix for automated compliance reporting.
- **Deterministic Boot Manifest Provider**: (P0) Core service that generates and signs "Environment Integrity Manifests" to fulfill deterministic boot requirements for high-security agent environments.

### Priority Shifts
- **A2A Messaging Hub**: (Re-affirmed P0) Evolving to support the finalized Linux Foundation open governance model for inter-agent task brokering.
- **Settings Injection Guard**: (Re-affirmed P0) Promoted as the primary defense against "Shadow Agent" configuration overrides identified in recent enterprise audits.

---

## Evolution: [2026-04-16] Updates

### Proposed Additions
- **Reactive Intent Gateway (RIG)**: (P0) Security middleware that mediates agent "Boundary Expansion" requests, validating them against the Root Mission Intent to prevent Intent Smuggling.
- **Resident Integrity Monitor (RIM)**: (P1) Background service that performs continuous, hardware-bound sandbox attestation to detect post-boot environment drift or tampering.
- **Self-Healing Consensus Hub**: (P0) A coordination service that provides a standardized interface for swarm state reconciliation, leveraging MAQ for authoritative "Truth Brokering."

### Priority Shifts
- **Deterministic Attestation Gateway**: (Re-affirmed P0) Expanded to support the new Resident Integrity Monitor for continuous lifecycle protection.
- **TPM-Bound Configuration Boot**: (Re-affirmed P0) Now considered the prerequisite foundation for RIG-mediated boundary expansions.

## Evolution: [2026-04-15] Updates

### Proposed Additions
- **Standardized Context Sidecar Interface**: (P1) A core API and "Context Bus" that allows MCP Any to host and bridge framework-specific context strategies (OpenClaw, etc.) across different agent frameworks.
- **Hardware-Attested Boot Manifest Provider**: (P0) Advanced attestation service that binds project-local environment manifests to a TPM/Secure Enclave, ensuring configuration integrity.
- **VTD Autonomous Delegation Engine**: (P0) Automation layer for the Delegation Attestation Layer that executes low-risk A2A handoffs without manual approval, based on safety proofs.

### Priority Shifts
- **Verifiable Task Delegation (VTD)**: (Re-affirmed P0) Now elevated as the primary solution for the "Approval Fatigue" scaling bottleneck.
- **Pluggable Context Bridge**: (Re-affirmed P0) Expanded to support the new Standardized Context Sidecar Interface.

## Evolution: [2026-05-02] Updates

### Proposed Additions
- **Kernel-Bound Intent (KBI) Broker**: (P0) A kernel-resident security utility for translating OpenClaw KBIA tokens into MCP Any security policies, binding tool quotas to the agent's semantic intent.
- **Swarm-Aware Capability Handoff (SACH) Gateway**: (P0) High-performance middleware for the A2A Messaging Hub that manages "Capability Leases," allowing sub-second tool permission delegation within a swarm.
- **Intent-Scoped Resource Quota (ISRQ) Middleware**: (P1) Extension for the Adaptive Intent Budgeting layer that enforces budget suspensions based on predicted tool-call impact.

### Priority Shifts
- **Adaptive Intent Budgeting (AIB)**: (Promoted to P0) Critical for aligning with the latest Claude Code and Gemini CLI ISRQ enforcement.
- **SACH Gateway**: (Promoted to P0) Designated as the primary defense against the newly discovered "Context-Mirroring" identity spoofing exploit.
