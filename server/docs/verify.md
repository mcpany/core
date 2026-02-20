# Verification Guide

This guide explains how to verify that your `mcpany` installation is correct and fully functional.

## 1. Installation Verification

First, ensure that the `mcpany` binary has been built correctly.

```bash
# Verify the binary exists
ls -l build/bin/server

# Verify the version
./build/bin/server version
```

## 2. Configuration Validation (Doctor)

Use the built-in `doctor` command to validate your configuration files and system environment.

```bash
# Run doctor with a minimal configuration
./build/bin/server doctor --config-path server/config.minimal.yaml
```

**Expected Output:**
The output should indicate that all checks passed:
```text
Running doctor checks...
========================
[ ] Checking Configuration... OK
...
All checks passed!
```

## 3. Server Health Check

Start the server and verify it is responsive.

```bash
# Start the server in the background
./build/bin/server run --config-path server/config.minimal.yaml &
SERVER_PID=$!

# Wait for startup (e.g., 5 seconds)
sleep 5

# Check the health endpoint
curl -v http://localhost:50050/health

# Stop the server
kill $SERVER_PID
```

**Expected Output:**
The `curl` command should return `HTTP/1.1 200 OK` and the body `OK`.

## 4. Running Tests

To verify the codebase integrity, run the test suite.

```bash
# Run unit tests
make test
```

> **Note:** Some E2E tests require Docker. If you are running in an environment without Docker, you can skip them using `SKIP_DOCKER_TESTS=true make test`.

## 5. UI Verification (Optional)

If you have built the UI, you can verify it by navigating to `http://localhost:3000` (development) or the configured UI port.

```bash
# Build UI
make -C ui build
```
