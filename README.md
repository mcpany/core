# MCP Any

## Elevator Pitch

MCP Any is a powerful universal connectivity platform engineered for scale. Built upon a modern Go microservices architecture, it fundamentally bridges communication between distributed swarms, robust AI agents, and versatile toolchains securely. MCP Any breaks down data silos, enabling seamless, verified runtime operations that unify complex infrastructures into a cohesive ecosystem.

## Architecture

MCP Any is designed with a resilient, microservices-oriented backend. Key patterns include:

- **Universal Agent Bus (UAB):** Acts as the central nervous system, dispatching events rapidly using Kafka or Redis streams.
- **Pluggable Upstream Adapters:** Easily extend the platform to interface with GraphQL, REST, gRPC, and specific toolchains (e.g. Docker, Vector DBs).
- **Middleware-Driven Security:** Every request flows through robust pipelines (RBAC, Ratelimiting, Auditing) ensuring security by default.
- **React/Vite Frontend:** A blazing fast administration dashboard and playground allowing direct interaction and visualization of your runtime environment.

## Getting Started

### Prerequisites
- Go 1.22+
- Docker and Docker Compose
- Node.js 18+ (for frontend development)

### Quickstart

1. **Clone the repository:**
   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```

2. **Start the Infrastructure:**
   ```bash
   docker-compose up -d
   ```

3. **Build the Backend:**
   ```bash
   make build
   ```

4. **Run the Server:**
   ```bash
   make run
   ```

5. **Start the UI:**
   ```bash
   cd ui
   npm install
   npm run dev &
   ```

Visit `http://localhost:5173` to see MCP Any in action.

## Development

The project utilizes `make` to automate common development workflows.

- **Compile Backend:** `make build`
- **Run Tests:** `make test` (Executes the Go test suite and UI tests)
- **Lint Code:** `make lint` (Runs `golangci-lint` ensuring strict code quality)
- **Generate Protobufs:** `make proto` (Regenerates all necessary gRPC bindings)

## Configuration

Configuration is managed flexibly via environment variables and configuration files (`config.yaml`).

**Required Environment Variables:**
- `MCP_DATABASE_URL`: Connection string for the primary Postgres database.
- `MCP_REDIS_URL`: Connection string for the caching layer.
- `MCP_JWT_SECRET`: A secure secret key used for issuing and validating auth tokens.
- `MCP_API_PORT`: (Optional) Overrides the default port `8080`.

For advanced configurations, including OAuth integration and TLS certificates, please refer to the internal `docs/` directory.
