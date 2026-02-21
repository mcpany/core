# Verification Guide

This document outlines the steps to verify the health, configuration, and functionality of the MCP Any server.

## 1. Automated Tests

The primary way to verify the codebase is through the automated test suite.

### Running All Tests
To run all unit, integration, and end-to-end tests, execute the following command from the project root:

```bash
make test
```

This will trigger the `test` target in the `server/Makefile`, which runs:
- `test-fast`: Unit tests (excluding slow integration tests).
- `e2e-local`: Parallel and sequential end-to-end tests using Docker containers.
- `test-public-api`: Public API verification tests.

### Running Specific E2E Tests
If you want to run only the end-to-end tests:

```bash
make -C server e2e
```

## 2. Server Health Check

Once the server is running (e.g., via `make run` or Docker), you can verify its health using the exposed HTTP endpoints.

### Liveness Probe (`/healthz`)
This endpoint returns a 200 OK status if the server is running and ready to accept connections. It is suitable for Kubernetes liveness probes.

```bash
curl -v http://localhost:50050/healthz
```

### Readiness Probe (`/health`)
This endpoint performs a more comprehensive check, including the status of critical internal components.

```bash
curl -v http://localhost:50050/health
```

## 3. The `doctor` Command

The `doctor` command performs a deep diagnostic check of your configuration and upstream connectivity.

### Usage
If you have built the server binary locally:

```bash
./build/bin/server doctor
```

If you are using the Docker image:

```bash
docker run --rm -v $(pwd)/config.yaml:/etc/mcpany/config.yaml mcpany/server:latest doctor --config-path /etc/mcpany/config.yaml
```

### Checks Performed
1.  **Configuration Syntax**: Validates `config.yaml` schema.
2.  **Environment Variables**: Checks for missing required variables.
3.  **Connectivity**: Pings upstream services defined in your configuration.
4.  **System Resources**: Checks for file permissions and network availability.

## 4. Manual Verification (UI)

To verify the User Interface:
1.  Start the UI server: `cd ui && npm run dev` (or use Docker Compose).
2.  Navigate to `http://localhost:3000`.
3.  Check the **Dashboard** for system metrics.
4.  Use the **Playground** to execute a tool call against a configured service.
5.  Check the **Traces** page to verify the request was logged.
