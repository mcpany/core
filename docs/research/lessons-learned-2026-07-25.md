# Lessons Learned: Evolution to Universal Agent Infrastructure (2026-07-25)

## Technical Learnings

### 1. Protobuf API_OPAQUE Migration
The recent migration to `API_OPAQUE` in the project's Protobuf definitions has significant impact on Go code:
- Direct field access (e.g., `config.Enabled`) is no longer permitted.
- **Getters** must be used for reading (e.g., `config.GetEnabled()`).
- **Builders** must be used for instantiation (e.g., `configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()`).
- This applies to all generated messages including configuration and router protocol messages.

### 2. Bazel Dependency Management in Multi-Module Repos
When managing dependencies between local packages in a Bazel workspace (e.g., between `server` and `k8s/operator`):
- Prefer relative workspace paths (`//k8s/operator/api/v1alpha1`) over external-style references (`@com_github_mcpany_core//...`) for internal components to avoid resolution issues with Gazelle.
- Ensure `BUILD.bazel` files accurately reflect the local directory structure.

### 3. MCP Middleware Implementation
The `MetadataSanitizationGateway` (MSG) implementation highlighted the structure of MCP results:
- `mcp.CallToolResult.Content` is a slice of `mcp.Content` interfaces.
- Type assertions to `*mcp.TextContent` are required to access and modify the `Text` field.
- The `Text` field in the current SDK version is a raw `string`, not a pointer or a wrapper that requires method calls.

## Strategic Findings

### 1. Mesh-Resident Governance
Research into OpenClaw and Agent Swarms reveals a shift from centralized gateways to **Mesh-Resident Governance**. Components like the **Mesh-Resident Lease Arbiter (MRLA)** are essential to prevent hardware deadlocks in distributed local LLM environments.

### 2. Congestion Sovereignty
Autonomous agents operating at scale require P2P congestion control. Standard HTTP backoff is insufficient; agents need to negotiate priority based on "mission criticality" rather than just network availability.

### 3. Reasoning-Aware Proof Aggregation (RAPA)
With the rise of "reasoning" models (like Gemini 2.0 / Claude 3.5), tool calls often carry implicit Chain-of-Thought metadata. The Universal Agent Infrastructure must provide a way to aggregate these proofs into ZK-verifiable state transitions without leaking the raw reasoning traces.
