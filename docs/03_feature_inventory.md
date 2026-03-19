# Feature Inventory: MCP Any

## Current Backlog (P0/P1)
- **Policy Firewall**: Rego/CEL based hooking for tool calls.
- **HITL Middleware**: Suspension protocol for user approval flows.
- **Recursive Context Protocol**: Standardized headers for subagent inheritance.
- **Shared KV Store**: Embedded SQLite "Blackboard" tool for agents.

## Evolution: [2026-06-16] Updates

### Proposed Additions
- **Entangled State Broker (ESB)**: (P0) Authoritative coordination service for "Entanglement Shards" that are cryptographically bound to the mission-root intent.
- **Stylometric Mimicry Mitigator (SMM)**: (P0) Security middleware that performs real-time stylometric analysis of inter-agent messages to detect reasoning-path shadowing.
- **Speculative Branching Guard (SBG)**: (P1) Isolation service for un-executed reasoning paths that prevents speculative attention leakage.
- **Mesh-Resident Key Exchange (MRKE) Provider**: (P0) Hardware-bound session key rotation service for sub-100ms inter-teammate coordination.

### Priority Shifts
- **Atomic Reasoning Integrity (ARI) Validator**: (Re-affirmed P0) Now elevated with the requirement for mandatory **ESB-compliant** state entanglement.
- **Stylometric Metadata Sanitizer (SMS)**: (Re-affirmed P0) Evolving to support the new **Stylometric Mimicry** defense requirements.

## Evolution: [2026-06-08] Updates

### Proposed Additions
- **Atomic Reasoning Integrity (ARI) Validator**: (P0) Advanced security middleware for the Mailbox Integrity Middleware that performs fragment-level semantic validation of shared teammate state.
- **HAMM-Locked MLE Gateway**: (P0) Upgrade for the MLE Gateway to support "Hardware-Attested Mission Manifests," providing an immutable, hardware-locked boundary for tool discovery and execution.
- **Temporal Decay Orchestrator**: (P1) Lifecycle management service for the Temporal Sovereignty Controller that handles "Graceful Mission Decay" signals and manages restricted agency transitions.
- **Fragment-Level Sovereignty Attestation Provider**: (P0) Advanced security service mandating ARI-attestation for all A2A-compliant teammates to access shared shards.

### Priority Shifts
- **Mailbox Integrity Middleware**: (Re-affirmed P0) Now elevated with the requirement for mandatory **ARI** integration to counter fragment-level state-splicing.
- **Mission-Locked Execution (MLE) Gateway**: (Re-affirmed P0) Designated as the primary enforcement point for **HAMM-compliant** mission manifests.

## Evolution: [2026-06-07] Updates

### Proposed Additions
- **Semantic Shadowing Mitigator (SSM)**: (P0) A behavioral security middleware for the AID Hub that performs stylometric and contextual consistency checks to detect mimicry-based intent hijacking.
- **Mission-Locked Execution (MLE) Gateway**: (P0) Core security service that enforces cryptographic locking of tool calls and sub-delegations to a hardware-attested mission-root intent.
- **STR-Native Discovery Provider**: (P1) Upgrade for the PNTD Provider to support "Sovereign Tool Registry" manifests and TPM-signed behavioral baselines.
- **Temporal Sovereignty Controller**: (P1) Lifecycle management service implementing "Ephemeral Mission Roots" to prevent long-term session hijacking.

### Priority Shifts
- **Active Intent-Deconstruction (AID) Hub**: (Re-affirmed P0) Now elevated with the requirement for mandatory **SSM** integration to counter mimicry attacks.
- **Capability Garbage Collection (CGC) Provider**: (Re-affirmed P0) Designated as a critical mechanism for supporting the new **Temporal Sovereignty** requirements.

## Evolution: [2026-05-23] Updates

### Proposed Additions
- **Federated Swarm Identity (FSI) Provider**: (P0) A local identity service that issues hardware-attested, cross-framework tokens for secure teammate verification in heterogeneous meshes.
- **Intent-Leakage Shielding (ILS) Middleware**: (P0) Security extension for the MRP middleware that monitors semantic entropy and blocks subagent requests designed to probe mission-root constraints.
- **Hardware-Attested Discovery Handshake (HADH) Gateway**: (P0) Advanced discovery service that mandates hardware-bound handshakes before revealing any agent capabilities to peers.
- **Reasoning-Effort Quota Controller**: (P0) Resource management middleware that dynamically throttles high-intensity reasoning (e.g., `x-gemini-reasoning-effort`) to prevent Agentic DoS.

## Evolution: [2026-05-24] Updates

### Proposed Additions
- **Active Negotiation Broker (ANB)**: (P0) Authoritative bidding bus for multi-agent auctions, utilizing hardware-attested Capability Cards to filter and validate bids locally.
- **Differential Context Guarding (DCG) Middleware**: (P0) Security extension for the Mailbox Integrity Middleware that performs semantic analysis of tool outputs to prevent context-dump exfiltration.
- **Zero-Knowledge Capability Proof (ZKCP) Provider**: (P1) Advanced discovery service allowing agents to prove skill possession without revealing sensitive implementation details during the discovery phase.
- **Self-Correction Loop Arbiter**: (P0) Lifecycle security middleware that monitors subagent refinement drift and terminates sessions bypassing parent intent constraints.

### Priority Shifts
- **Mailbox Integrity Middleware**: (Re-affirmed P0) Now elevated with the requirement for mandatory DCG to counter CVE-2026-39102.
- **`TeammateTool` Orchestration Adapter**: (Re-affirmed P0) Evolving to support ANB-native task auctions.

## Evolution: [2026-05-23] Updates

### Proposed Additions
- **Local-Only WebSocket Auth (LOWA) Gateway**: (P0) A mandatory security layer for all local listeners that enforces session-bound authentication to neutralize "ClawJacked" style brute-force attacks.
- **Teammate-to-Teammate (T2T) Encryption Bridge**: (P0) Infrastructure for secure, peer-to-peer mailbox messaging and task list synchronization between teammates from disparate frameworks.
- **Mailbox Integrity Middleware**: (P0) Security extension for the T2T Bridge that validates inter-agent messages against the "Mission Root" intent to prevent malicious mailbox injection.
- **Full-Mesh Discovery Auth Provider**: (P0) Advanced discovery service that mandates hardware-attested handshakes before revealing agent capability cards in a mesh environment.

### Priority Shifts
- **Inter-Agent Mailbox Guard (IAMG)**: (Evolved to Mailbox Integrity Middleware) Now designated as a mandatory requirement for all mesh-based teammate coordination.
- **Origin-Locked Agent Gateway**: (Re-affirmed P0) Now elevated with the requirement for mandatory session-bound LOWA authentication.

## Evolution: [2026-05-21] Updates

### Proposed Additions
- **Cognitive Load Shedding (CLS) Controller**: (P0) A high-speed stability middleware that dynamically throttles or revokes subagent capabilities based on real-time reasoning intensity and mission stability scores.
- **Temporal Reasoning Attestation (TRA) Provider**: (P0) Security extension for the SRM Provider that adds hardware-attested monotonic timestamps to reasoning fragments to neutralize "Reasoning Timing Attacks."
- **CFRR Reconciliation Adapter**: (P1) Orchestration bridge for OpenClaw's Conflict-Free Replicated Reasoning engine, enabling MCP Any to merge non-conflicting reasoning traces in parallel teams.
- **Hardware-Attested Privacy Enclave (HAPE) Adapter**: (P0) Infrastructure for local, hardware-bound processing of sensitive PII context, providing only sanitized intent fragments to cloud providers.

### Priority Shifts
- **SRM Provider**: (Re-affirmed P0) Now elevated with the requirement for mandatory TRA to prevent context-switch hijacking.
- **TeammateTool Orchestration Adapter**: (Re-affirmed P0) Evolving to support CFRR-native state reconciliation.

## Evolution: [2026-05-20] Updates

### Proposed Additions
- **Policy-Bound Reasoning (PBR) Adapter**: (P0) Infrastructure for hosting and enforcing immutable "Policy Anchors" at the pre-reasoning layer, ensuring cross-framework cognitive governance.
- **Multi-modal Integrity Bridge (MIB)**: (P0) Upgrade for the Semantic Integrity Bridge providing real-time sanitization of non-textual traces (SVG, CSS, Audio metadata) to prevent context smuggling.
- **AIR Reconciliation Broker**: (P1) Decentralized intent reconciliation service utilizing hardware-attested multi-signature quorums to resolve conflicting swarm objectives.

### Priority Shifts
- **Semantic Integrity Bridge**: (Evolved to MIB) Now designated as the primary defense against multi-modal "Context Smuggling" exploits.
- **Cognitive Truth Attestation Hub**: (Promoted to P0) Critical for providing the verifiable proof required for AIR-mediated intent reconciliation.

## Evolution: [2026-05-19] Updates

### Proposed Additions
- **Signed Reasoning Monologue (SRM) Provider**: (P0) A core security middleware that cryptographically binds internal monologues to hardware-attested sessions, neutralizing "Reasoning Hijacking."
- **Namespace-Locked Discovery (NLD) Gateway**: (P0) Advanced extension for the PNTD Provider that ensures deterministic and collision-free capability mapping across registries.
- **HASS-Compliant PLSS Manager**: (P0) Upgrade for the Project-Local Snapshot Sync supporting TPM-signed environment snapshots for "Deterministic Sandbox Recovery."
- **Cognitive Truth Attestation Hub**: (P1) Advanced orchestration service that provides verifiable proof of reasoning integrity across heterogeneous agent swarms.

### Priority Shifts
- **Project-Local Snapshot (PLSS) Sync**: (Promoted to P0) Now critical for implementing HASS-compliant "Point-in-Time Integrity."
- **PNTD Discovery Provider**: (Re-affirmed P0) Now designated as the mandatory registry for all enterprise swarms to support NLD.

## Evolution: [2026-05-18] Updates

### Proposed Additions
- **Mission-Root Pinning (MRP) Middleware**: (P0) A transport-level security component that protects the "Mission Root" from context-window eviction during high-frequency "noise" injections (MRE defense).
- **State-Trust Labeling (STL) Provider**: (P0) Security extension for the Blackboard that tags all KV data with the trust level of its origin framework, neutralizing PASI (Protocol-Agnostic State Injection).
- **Wait-Graph Deadlock Resolver**: (P1) Advanced orchestration service for the `TeammateTool` Adapter that proactively breaks circular task dependencies in parallel swarms.
- **Intent-Weighted Context Summarizer**: (P1) Upgrade for the ContextEngine Adapter supporting RCE v2.0 logic for mission-anchored context compression.

### Priority Shifts
- **TeammateTool Orchestration Adapter**: (Re-affirmed P0) Now elevated with the requirement for "Multi-Agent Quorum" (MAQ) cross-framework coordination.
- **Contextual Quorum (CQ) Hub**: (Promoted to P0) Critical for supporting the new Claude-led MAQ protocol for high-risk actions.

## Evolution: [2026-05-17] Updates

### Proposed Additions
- **`TeammateTool` Orchestration Adapter**: (P0) Infrastructure for cross-framework "Agent Teams," facilitating Claude-style delegation and synchronization for heterogeneous swarms.
- **Transport-Layer Session Binder (TLSB)**: (P0) A security middleware that cryptographically binds inter-agent transport channels (Named Pipes/WebSockets) to hardware-attested reasoning session tokens.
- **Authenticated Agent Card Discovery**: (P0) Identity-bound discovery service for the A2A Messaging Hub that enforces "Auth-Before-Discovery" for agent capabilities.
- **ContextEngine Lifecycle Adapter (v2026.3.7)**: (P0) Upgrade for the ContextEngine Adapter to support the full OpenClaw v2026.3.7 lifecycle hooks for third-party context plugins.

### Priority Shifts
- **A2A Messaging Hub**: (Re-affirmed P0) Designated as the primary gateway for the new "Authenticated Agent Card Discovery."
- **Isolated Named-Pipe Transport Middleware**: (Re-affirmed P0) Elevated with the requirement for mandatory TLSB to prevent "Team Ghosting."

## Evolution: [2026-05-16] Updates

### Proposed Additions
- **Reasoning Quorum Middleware**: (P0) Infrastructure for agents to reach a cryptographically bound quorum on non-deterministic reasoning outputs, neutralizing "Hallucination Variance."
- **Transport-Layer Session Binder**: (P0) Security middleware that cryptographically binds every named-pipe and local transport connection to a unique hardware-attested reasoning session token.
- **RRRA Budget Controller**: (P1) Advanced resource manager implementing Reasoning-Responsive Resource Allocation, scaling compute/token budgets based on real-time reasoning intensity.
- **Intent-Aware Transport Proxy**: (P1) Efficiency middleware that performs semantic deduplication of coordination messages between parallel agents sharing a mission root.

### Priority Shifts
- **Coordination Token Optimizer**: (Promoted to P0) Critical for neutralizing the overhead and "Team Ghosting" risks in parallel swarm coordination.
- **Consensus Tool Validation Hub**: (Re-affirmed P0) Evolving to support the new Reasoning-Level Consensus (RLC) requirements.

## Evolution: [2026-05-15] Updates

### Proposed Additions
- **Consensus Tool Validation Hub**: (P0) Distributed security middleware requiring multi-agent signatures for high-risk tool calls and task delegations, neutralizing "Agentic Social Engineering."
- **PNTD Discovery Provider**: (P1) Implementation of Protocol-Neutral Task Discovery to unify capability mapping across MCP, gRPC, and UACO transports, providing a universal discovery bus.
- **Intent-Bound Memory Isolation**: (P0) Extension for the ContextEngine Adapter that ensures "Mission-Root" anchors are cryptographically protected and semantically isolated to prevent "Context Ghosting."
- **Negative Discovery Attestation Provider**: (P0) Mandatory extension for the PNTD Provider that provides cryptographic proof of the absolute absence of unauthorized hook execution during the discovery phase.

### Priority Shifts
- **Consensus Tool Validation Gateway**: (Re-affirmed P0) Designated as a mandatory requirement for all enterprise swarm deployments to counter machine-speed coercion.
- **Shared KV Store (Blackboard)**: (Re-affirmed P0) Expanded to support "Intent-Bound Memory Isolation" as the primary state persistence model.

## Evolution: [2026-05-14] Updates

### Proposed Additions
- **ContextEngine Lifecycle Adapter**: (P0) A native implementation of the OpenClaw v2026.3.7 ContextEngine lifecycle hooks, enabling MCP Any to act as a universal host for pluggable context plugins.
- **Swarm-Aware Rate Limiter**: (P0) A high-speed security middleware designed to detect and neutralize coordinated "Hivenet" swarm attacks at sub-millisecond speeds.
- **Hardware-Attested NHI Identity Wallets**: (P1) Integration of TPM/Secure Enclave-bound machine identities for all connected agents, ensuring non-repudiable agency and Zero-Trust identity.
- **Asynchronous Telemetry Sink**: (P1) High-speed, non-blocking telemetry middleware that acts as the authoritative collector for OpenClaw-RL v1.0 reasoning traces and rollout tokens.

### Priority Shifts
- **Injection-Shielding Middleware**: (Re-affirmed P0) Designated as a mandatory prerequisite for all tool-driven code commits to counter high vulnerability rates in agent-generated PRs.
- **A2A Messaging Hub**: (Re-affirmed P0) Expanded to support "Hardware-Attested NHI Wallets" as the primary identity transport.

## Evolution: [2026-05-13] Updates

### Proposed Additions
- **Loopback Authentication Proxy**: (P0) A mandatory security interceptor for all local network ports that enforces origin-locked authentication, neutralizing "ClawdBot" style loopback hijacking.
- **Injection-Shielding Middleware**: (P0) Pre-execution scanning service that performs SEMGREP-style static analysis and semantic validation on all tool inputs to block prompt and command injection.
- **Coordination Token Optimizer**: (P1) Efficiency middleware for parallel swarms that deduplicates and compresses coordination messages within the named-pipe bus to reduce token overhead.

### Priority Shifts
- **Isolated Named-Pipe Transport Middleware**: (Re-affirmed P0) Designated as the mandatory replacement for all local TCP/UDP coordination channels.
- **Pre-Flight Sandbox Validator**: (Promoted to P0) Critical for integrating the new Injection-Shielding logic before agent boot.

## Evolution: [2026-05-12] Updates

### Proposed Additions
- **Isolated Named-Pipe Transport Middleware**: (P0) A high-performance inter-agent transport layer using Docker-bound named pipes (UNIX domain sockets) to eliminate local port exposure.
- **Subagent Routing Firewall**: (P0) Transport-level security gate that enforces "Auth-at-the-Pipe" identity validation before establishing inter-agent connections.
- **Kernel-Resident Trace Scrubber**: (P1) Real-time semantic sanitization engine for binary state handoffs (BSH) within isolated named-pipe transports.

### Priority Shifts
- **Parallel Team Coordination Hub**: (Re-affirmed P0) Evolved to mandate the use of Isolated Named-Pipe Transport for all inter-teammate coordination.
- **A2A Messaging Hub**: (Promoted to P0) Critical requirement for managing "Auth-at-the-Pipe" tokens across heterogeneous agent swarms.

## Evolution: [2026-05-11] Updates

### Proposed Additions
- **Parallel Team Coordination Hub**: (P0) High-speed coordination bus for Claude Code-style "Agent Teams," providing message passing and Snapshot-and-Merge state reconciliation for parallel branches.
- **Negative Discovery Attestation Provider**: (P0) Extension of the Discovery Sandbox that provides cryptographic proof of the absolute absence of unauthorized hook execution during the discovery phase.
- **Async RL Rollout Orchestrator**: (P1) High-speed, non-blocking telemetry bridge for OpenClaw-RL v1.0, facilitating the export of reasoning traces and PRM evaluations.

### Priority Shifts
- **Discovery Sandbox Middleware**: (Re-affirmed P0) Evolved with the requirement for "Mandatory Discovery-Phase Isolation" to counter CVE-2026-0628.
- **Shared KV Store (Blackboard)**: (Promoted to P0) Critical for implementing the "Snapshot-and-Merge" reconciliation needed for parallel agent teams.

## Evolution: [2026-05-10] Updates

### Proposed Additions
- **Discovery Sandbox Middleware**: (P0) A secure, ephemeral execution environment for MCP discovery commands (e.g., Gemini's `discoveryCommand`), preventing host-level "Ghost-Execution" exploits.
- **Session-Persistent DAP Provider**: (P0) Advanced extension of the DAP generator that maintains a hardware-attested manifest of non-existent files throughout the mission lifecycle, neutralizing "Shadow-Sandbox" escapes.
- **Async RL Telemetry Orchestrator**: (P1) High-speed, non-blocking telemetry bridge for OpenClaw-RL v1.0, facilitating the export of reasoning traces and PRM evaluations for background policy optimization.

### Priority Shifts
- **Deterministic Absence Proof (DAP) Generator**: (Promoted to P0) Critical for neutralizing CVE-2026-25725 style sandbox escapes in multi-agent environments.
- **RL Telemetry Provider**: (Re-affirmed P0) Evolved with the requirement for "Asynchronous Rollout Collection" to support the OpenClaw-RL v1.0 standard.

## Evolution: [2026-05-09] Updates

### Proposed Additions
- **Cryptographic Lineage Validator**: (P0) A core security middleware that enforces mandatory parent-child token binding for all subagent spawns, neutralizing "Shadow Subagent" context contamination.
- **Continuous CPCP Enforcer**: (P0) A high-frequency validation service for project-local configurations that performs hardware-attested checks during every tool call.
- **ARE-Responsive Budget Controller**: (P1) Resource management layer that consumes Gemini CLI `ARE` headers to dynamically prioritize token allocation for high-intensity reasoning.

### Priority Shifts
- **Deterministic Permission Guard (DPG)**: (Re-affirmed P0) Evolved with the requirement for "Per-Call Integrity" mapping to the CPCP standard.
- **Recursive Depth-Limit Middleware**: (Promoted to P0) Critical for preventing infinite "Shadow Spawning" loops in autonomous swarms.

## Evolution: [2026-05-08] Updates

### Proposed Additions
- **Context Sealed-Fragment Hub**: (P0) Implementation of "Active Fragment Sealing" to protect context shards from semantic side-channel exfiltration (defense against "EchoLeak").
- **Deterministic Permission Guard (DPG)**: (P0) A kernel-level security middleware that enforces project-local "Deny" rules independently of the agent's reasoning state (defense against Bug #8961).
- **Asynchronous RL Rollout Collector**: (P1) AUTHORITATIVE telemetry bridge for OpenClaw-RL v1.0, facilitating high-frequency feedback collection for policy optimization.

### Priority Shifts
- **Distributed Supervisor Mesh (DSM) Orchestrator**: (Promoted to P0) Designated as a critical infrastructure requirement for the 2026 enterprise swarm pivot.
- **Programmatic SDK Boundary Enforcer**: (Re-affirmed P0) Evolved with the requirement for "Context-Poisoning" defense in automated scripts.

## Evolution: [2026-05-07] Updates

### Proposed Additions
- **Programmatic SDK Boundary Enforcer**: (P0) Mandatory security gating for SDK-driven agent interactions (e.g., OpenCode SDK), ensuring programmatic tool calls comply with Zero-Trust policies.
- **Distributed Supervisor Mesh (DSM) Orchestrator**: (P1) Infrastructure for decentralized delegation and oversight, allowing any agent to act as a local supervisor while anchored to a mission root.
- **Autonomous Escalation Resolver**: (P1) Mitigation service for "Negotiation Deadlocks" in autonomous swarms, applying mission-aligned fairness policies to break bidding loops.

### Priority Shifts
- **Inter-Swarm Deadlock Detector**: (Promoted to P0) Critical for neutralizing resource exhaustion in autonomous production swarms.
- **Hierarchical Intent Lease (HIL) Broker**: (Re-affirmed P0) Essential for managing the lifecycle of decentralized supervisors in a DSM.

## Evolution: [2026-05-06] Updates

### Proposed Additions
- **Origin-Locked Agent Gateway**: (P0) A mandatory security layer for all local listeners that enforces `Origin`, `Sec-Fetch-Site`, and session-token binding to neutralize "ClawJacked" style hijacking.
- **Intent-Sealed Blackboard Shards**: (P0) Implementation of Reason-Aware Memory Segmentation (RAMS) providing cryptographically isolated memory regions for subagents within the Shared KV Store.
- **Fast-Path Trust Lease Broker**: (P1) A performance-optimizing security middleware that manages time-bound hardware-attested trust leases to reduce per-call attestation latency.

### Priority Shifts
- **Reasoning-Aware Memory Segmentation (RAMS) Hub**: (Re-affirmed P0) Evolved into the "Intent-Sealed Shards" model for default isolation.
- **Same-Origin Policy (SOP) Enforcer**: (Promoted to P0) Designated as a mandatory prerequisite for all local tool connectivity.

## Evolution: [2026-05-05] Updates

### Proposed Additions
- **Reasoning-Aware Memory Segmentation (RAMS) Hub**: (P0) A core extension for the Blackboard that provides cryptographically isolated "Intent-Sealed Shards" for subagents, neutralizing "Memory Smearing."
- **Hardware-Enclave Path Attestation (HEPA) Provider**: (P0) Security service that utilizes Secure Enclaves (TPM/SEP) to provide hardware-bound path validation during the initial O_PATH open phase.
- **Cross-Swarm Intent Attestation Middleware**: (P1) UACO-native service that facilitates multi-signature attestation of mission-root intents across heterogeneous agent swarms.

### Priority Shifts
- **Kernel-Bound FD Persistence**: (Evolved to HEPA) Upgraded with hardware enclave support for stronger path-resolution guarantees.
- **Semantic Integrity Bridge**: (Promoted to P0) Critical requirement for detecting "Recursive Context Splicing" (RCS) in multi-modal reasoning traces.

## Evolution: [2026-05-04] Updates

### Proposed Additions
- **Semantic Integrity Bridge**: (P0) A monitoring extension for the CQ Hub that utilizes "Intent Drift Detection" and SGC-aware analysis to prevent Recursive Intent Poisoning (RIP).
- **Kernel-Bound FD Persistence Middleware**: (P0) Advanced security layer that utilizes FD-passing and hardware-bound Inode pinning to ensure the absolute immutability of project-local configurations.
- **Bi-directional A2UI State Bridge**: (P1) Infrastructure for two-way state synchronization between the agent reasoning loop and the secure user interface, enabling "Corrective Intent" injection.

### Priority Shifts
- **Depth-Aware Inode Pinning (DAIP)**: (Evolved to Kernel-Bound FD Persistence) Upgraded to handle FD-passing for stronger immutability guarantees.
- **A2UI Native Gateway**: (Evolved to Bi-directional Bridge) Now requires support for user-initiated state pushes back to the agent.

## Evolution: [2026-05-03] Updates

### Proposed Additions
- **Deadlock-Resilient CQ Controller**: (P0) Advanced extension of the CQ Hub that performs "Wait-Graph Analysis" to identify and break circular attestation dependencies in multi-agent swarms.
- **Hierarchical Intent Lease (HIL) Broker**: (P0) Core security service implementing UACO v3.2 HIL. Manages hierarchical, task-bound capability leases that automatically expire upon sub-mission completion.
- **Depth-Aware Inode Pinning (DAIP) Middleware**: (P0) Security layer for the Shadow-FS that enforces mandatory depth-limit validation for recursive symlink tunnels, preventing host-region escapes.

### Priority Shifts
- **Inter-Swarm Deadlock Detector**: (Promoted to P0) Critical for preventing resource exhaustion in the face of malicious attestation loops.
- **KLIP Enforcement**: (Evolved to DAIP) Now requires depth-aware validation to counter recursive symlink tunnels.

## Evolution: [2026-05-02] Updates

### Proposed Additions
- **Risk-Adaptive CQ Controller**: (P0) A dynamic policy engine for the CQ Hub that scales the quorum threshold (number of required signatures) based on real-time tool risk scores and reasoning confidence.
- **Reasoning-Responsive Rate Limiter (RRRL)**: (P1) Safety middleware that throttles tool calls when an agent's reasoning confidence falls below a configured threshold, preventing hallucinatory loops.
- **Inter-Swarm Deadlock Detector**: (P1) UACO-native monitoring service that identifies circular dependencies in multi-agent attestation chains and triggers automated resolution/timeouts.

### Priority Shifts
- **Project-Local Snapshot (PLSS) Sync**: (Promoted to P0) Now critical for implementing the Deterministic Sandbox Recovery (DSR) patterns standardized by Claude Code.
- **Contextual Quorum (CQ) Hub**: (Re-affirmed P0) Evolving to support OpenClaw v2026.5.1 AQT (Adaptive Quorum Thresholds).

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

## Evolution: [2026-03-17] Updates

### Proposed Additions
- **Local-Loopback Rate Limiter**: (P0) Mandatory rate limiting and auditing for all 127.0.0.1 / ::1 traffic to mitigate high-frequency brute-force and token hijacking attempts.
- **Origin-Locked Session Bridge**: (P0) A hardened security layer that cryptographically binds session tokens to the initiating browser origin, neutralizing token exfiltration (CVE-2026-25253).
- **Inter-Agent Mailbox Guard (IAMG)**: (P0) A security gateway for inter-agent messaging (Mailboxes) that enforces Zero-Trust identity and intent-validation on all teammate-to-teammate requests.
- **Verifiable RL Reward Provider**: (P1) A "Truth Attestation" middleware that provides verifiable binary rewards (command success, hash match) to optimize agent reasoning loops.
- **Identity-Bound Discovery (IBD) Enforcer**: (P0) A core security layer for the Discovery Bus that mandates cryptographically bound mission-tokens for all capability discovery requests.

### Priority Shifts
- **Same-Origin Policy (SOP) Enforcer**: (Promoted to P0) Now designated as a mandatory prerequisite for all local listeners to counter CVE-2026-25253.
- **A2A Messaging Hub**: (Re-affirmed P0) Evolving to act as the primary transport for the Inter-Agent Mailbox Guard.
- **RL Telemetry Provider**: (Promoted to P0) Essential for feeding verifiable rewards back to OpenClaw-RL policy engines.

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
- **Active Intent Alignment (AIA) Broker**: (P0) Authoritative alignment service that issues hardware-attested heartbeats to ensure specialist agent reasoning traces remain mission-anchored.
- **Multi-Modal Behavioral Attestation (MMBA) Provider**: (P0) Advanced identity service anchoring stylometric profiles to multi-modal trace history (SVG/Audio).
- **UACO-Native Negotiation Hub**: (P0) Full implementation of UACO task bidding and stateful handoffs for cross-framework swarms.
- **Semantic State-Pinning (SSP) Middleware**: (P1) Attention governance service that pins mission-critical fragments to high-priority attention tiers.
- **Temporal Shard Jitter (TSJ) Injector**: (P0) Security extension for the ESB that injects hardware-attested timing jitter to neutralize CVE-2026-62001.

### Priority Shifts
- **Entangled State Broker (ESB)**: (Re-affirmed P0) Now elevated with the requirement for mandatory **TSJ Injection** for all cross-mission synchronization.
- **A2A Bridge**: (Promoted to P0) Evolving into the **UACO-Native Negotiation Hub** to support standardized task negotiation.
