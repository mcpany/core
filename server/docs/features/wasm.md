# WASM Plugin System

MCP Any provides a WebAssembly (WASM) plugin system for safe, sandboxed execution of custom logic, such as transformations or custom tool implementations.

## Features

- **Sandboxed Execution**: Plugins run in a secure WASM environment (using [wazero](https://wazero.io/)).
- **Dynamic Loading**: Load plugins from bytecode at runtime.

*(Note: Currently in experimental stage)*
