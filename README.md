# MCP Any

**Elevator Pitch:** MCP Any is an infrastructure project providing robust management and routing for Machine Control Protocol (MCP) services. It exists to simplify and scale the integration of automated tool calling across diverse platforms.

## Architecture
MCP Any is built using a modern Go backend and Bazel build system. Key patterns include:
- **Upstream Routing:** Robust gRPC and HTTP routing to disparate MCP tools.
- **Middleware & Interceptors:** Centralized logging, auditing, and authentication.
- **Operator Pattern:** A Kubernetes Operator for native orchestration of MCP workloads.
- **Reactive Bus:** An event bus (Redis, NATS, Kafka) for distributed signaling and health checking.

## Getting Started
1. **Clone the repository:**
   `git clone https://github.com/mcpany/core.git`
   `cd core`
2. **Install prerequisites:** Ensure Go 1.24+ and Bazel (via bazelisk) are installed.
3. **Run Hello World:**
   `make run-server` (Starts the backend server)

## Development
- **Run Tests:** `make test` (Executes the Bazel test suite)
- **Lint Code:** `make lint` (Runs formatting and golangci-lint checks)
- **Build Code:** `make build` (Compiles all binaries and tools)

## Configuration
MCP Any requires environment variables and secrets to function in production:
- `MCP_DB_URI`: Connection string for the primary PostgreSQL database.
- `JWT_SECRET`: Secret key for authenticating user tokens.
- `LOG_LEVEL`: Set to `debug`, `info`, `warn`, or `error`.
