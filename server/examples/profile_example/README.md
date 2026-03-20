# Profile Example
This example demonstrates how to use the *Profile* feature with authenticated users.
This example demonstrates how to use the server with profiles.

## Configuration
The `config.json` defines:
- A user `alice` with a user-specific API key `alice-secret`.
- A service `echo-service` that exposes a profile `dev` with its own API key `dev-secret`.
- `alice` is granted access to the `dev` profile.

## Running the Example

1. Build the server:
```bash
make build
```

2. Run the server with the configuration:
```bash
./build/bin/server run --config-path server/examples/profile_example/config.json
```

The server will start on port `8082` (as defined in `config.json`).

## Authentication Testing
You can authenticate using either the Profile Key (higher priority) or the User Key.

### 1. Using Profile Key
```bash
curl -X POST -H "X-Profile-Key: dev-secret" -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' http://localhost:8082/mcp/u/alice/profile/dev/
```

### 2. Using User Key
```bash
curl -X POST -H "X-User-Key: alice-secret" -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' http://localhost:8082/mcp/u/alice/profile/dev/
```

### 3. Invalid Key
```bash
curl -X POST -H "X-Profile-Key: wrong-secret" -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' http://localhost:8082/mcp/u/alice/profile/dev/
```
Should return `401 Unauthorized (Profile)`.

### 4. No Key
```bash
curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' http://localhost:8082/mcp/u/alice/profile/dev/
```
Should return `401 Unauthorized (User)` (since fallback to user auth fails).
