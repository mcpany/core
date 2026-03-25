# MCP Any

## Project Identity

**MCP Any** is a Universal Agent Infrastructure designed to support the latest Universal Agent Coordination standards. It provides the foundational core for high-density, high-trust agent swarms, bridging the gap between isolated AI agents and cohesive, synchronized swarm execution.

Our mission is to provide secure, attested, and observable state mediation between diverse agent frameworks, enabling true cross-framework collaboration and capabilities like Relational Intent Lineage and Kernel-Mediated Zero-Copy State Mediation.

## Quick Start

### Prerequisites

* [bazelisk](https://github.com/bazelbuild/bazelisk) (or bazel)
* Docker & Docker Compose (optional, for local services)
* Go 1.26+
* Node.js 20+

### Installation & Run

```bash
# Clone the repository
git clone https://github.com/mcpany/core.git
cd core

# Install dependencies and build the project
bazelisk build //...

# Run the MCP Server
bazelisk run //server/cmd/mcp_server

# Run the UI (in a separate terminal)
cd ui
npm install
npm run dev
```

## Developer Workflow

We rely on Bazel for robust and hermetic builds. However, a Makefile is provided for standard developer commands:

* **`make lint`**: Runs local linting for the codebase.
* **`make test`**: Runs unit and integration tests locally.
* **`make docker-lint`**: Runs linting inside the Bazel docker environment.
* **`make docker-test`**: Runs all tests hermetically using Bazel.
* **`make k8s-e2e`**: Executes end-to-end testing against a Kubernetes cluster.

## Architecture

At a high level, MCP Any is composed of several critical subsystems:

1. **Universal Agent Bus**: A unified event and message bus that routes tasks, memory shards, and intents between registered Agent Frameworks (e.g., CrewAI, AutoGen, OpenClaw).
2. **Memfd-Bound BSH Sanitizer**: A secure, memory-bound sandbox for WebAssembly-based execution, defending against malicious agent payloads and intent ghosting.
3. **Hardware-Attested Validation**: Includes depth-counters and intent-pinned memory shards to ensure that relational intent chains are enforced securely across the cluster.
4. **Relational Intent Lineage**: Ensures trace integrity for multimodal states moving across the security boundaries.
5. **Entangled State Broker (ESB)**: Handles the testing and secure routing of multi-party states, serving as a critical boundary protection mechanism.
