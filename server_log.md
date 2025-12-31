make: Entering directory '/usr/local/google/home/kxw/core/server'
Cleaning generated protobuf files...
Preparing development environment...
protoc version v33.1 is already installed.
Installing Go protobuf plugins...
Checking for Go protobuf plugins...
Go protobuf plugins installation check complete.
Checking for Python to install dependencies and pre-commit hooks...
Python found. Installing/updating pre-commit and fastmcp locally...
pre-commit installed at .git/hooks/pre-commit
Installing Node.js dependencies for integration tests...
Found package.json, running npm install...

up to date, audited 620 packages in 1s

153 packages are looking for funding
  run `npm fund` for details

2 high severity vulnerabilities

To address all issues, run:
  npm audit fix

Run `npm audit` for details.
Installing Playwright browsers (skipping system deps)...
Playwright Host validation warning:
╔══════════════════════════════════════════════════════╗
║ Host system is missing dependencies to run browsers. ║
║ Please install them with the following command:      ║
║                                                      ║
║     sudo npx playwright install-deps                 ║
║                                                      ║
║ Alternatively, use apt:                              ║
║     sudo apt-get install libevent-2.1-7t64           ║
║                                                      ║
║ <3 Playwright Team                                   ║
╚══════════════════════════════════════════════════════╝
    at validateDependenciesLinux (/usr/local/google/home/kxw/core/server/tests/integration/upstream/node_modules/playwright-core/lib/server/registry/dependencies.js:269:9)
    at process.processTicksAndRejections (node:internal/process/task_queues:95:5)
    at async Registry._validateHostRequirements (/usr/local/google/home/kxw/core/server/tests/integration/upstream/node_modules/playwright-core/lib/server/registry/index.js:961:14)
    at async Registry._validateHostRequirementsForExecutableIfNeeded (/usr/local/google/home/kxw/core/server/tests/integration/upstream/node_modules/playwright-core/lib/server/registry/index.js:1083:7)
    at async Registry.validateHostRequirementsForExecutablesIfNeeded (/usr/local/google/home/kxw/core/server/tests/integration/upstream/node_modules/playwright-core/lib/server/registry/index.js:1072:7)
    at async i.<anonymous> (/usr/local/google/home/kxw/core/server/tests/integration/upstream/node_modules/playwright-core/lib/cli/program.js:217:7)
Checking for Helm...
Helm is already installed.
Checking for ShellCheck...
ShellCheck is already installed.
Checking for nats-server...
nats-server is already installed.
Checking for golangci-lint...
Installing golangci-lint v2.7.2 to /usr/local/google/home/kxw/core/build/env/bin...
golangci/golangci-lint info checking GitHub for tag 'v2.7.2'
golangci/golangci-lint info found version: 2.7.2 for v2.7.2/linux/amd64
golangci/golangci-lint info installed /usr/local/google/home/kxw/core/build/env/bin/golangci-lint
golangci-lint installed successfully.
Downloading go modules...
Preparation complete.
Removing old protobuf files...
Generating protobuf files...
Using protoc: libprotoc 33.1
Protobuf generation complete.
Building Go project locally...
Starting MCP Any server locally...
Error: failed to load services from config: failed to load config from store: failed to collect config file paths: failed to stat path /usr/local/google/home/kxw/core/examples/config.yaml: stat /usr/local/google/home/kxw/core/examples/config.yaml: no such file or directory
Usage:
  mcpany run [flags]

Flags:
      --api-key string              API key for securing the MCP server. If set, all requests must include this key in the 'X-API-Key' header. Env: MCPANY_API_KEY
      --db-path string              Path to the SQLite database file. Env: MCPANY_DB_PATH (default "data/mcpany.db")
      --grpc-port string            Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPANY_GRPC_PORT
  -h, --help                        help for run
      --profiles strings            Comma-separated list of active profiles. Env: MCPANY_PROFILES (default [default])
      --shutdown-timeout duration   Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT (default 5s)
      --stdio                       Enable stdio mode for JSON-RPC communication. Env: MCPANY_STDIO

Global Flags:
      --config-path strings             Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH
      --debug                           Enable debug logging. Env: MCPANY_DEBUG
      --log-format string               Set the log format (text, json). Env: MCPANY_LOG_FORMAT (default "text")
      --log-level string                Set the log level (debug, info, warn, error). Env: MCPANY_LOG_LEVEL (default "info")
      --logfile string                  Path to a file to write logs to. If not set, logs are written to stdout.
      --mcp-listen-address string       MCP server's bind address. Env: MCPANY_MCP_LISTEN_ADDRESS (default "50050")
      --metrics-listen-address string   Address to expose Prometheus metrics on. If not specified, metrics are disabled. Env: MCPANY_METRICS_LISTEN_ADDRESS

make: *** [Makefile:518: run] Error 1
make: Leaving directory '/usr/local/google/home/kxw/core/server'
