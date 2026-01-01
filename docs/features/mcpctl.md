# Config Validator CLI (`mcpctl`)

**Status**: Implemented

`mcpctl` is a command-line interface tool designed to help developers and CI/CD pipelines validate the configuration of MCP Any servers before deployment.

## Installation

You can build `mcpctl` from source:

```bash
go build -o mcpctl ./server/cmd/mcpctl
```

## Usage

### Validate Configuration

The primary command is `validate`, which checks your configuration files for syntax errors and schema validity.

```bash
./mcpctl validate --config-path ./config.yaml
```

If the configuration is valid, it prints "Configuration is valid." and exits with code 0.
If invalid, it prints the errors and exits with a non-zero status code.

### CI/CD Integration

`mcpctl` is designed to be used in your CI pipeline.

**Example: GitHub Actions**

```yaml
jobs:
  validate-config:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build mcpctl
        run: go build -o mcpctl ./server/cmd/mcpctl
      - name: Validate Config
        run: ./mcpctl validate --config-path config/production.yaml
```
