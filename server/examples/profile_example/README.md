# Profile Example
This example demonstrates how to use the *Profile* feature with authenticated users.
Profiles allow you to define distinct environments (like `dev`, `prod`) with their own authentication keys, and assign users access to specific profiles.

## Configuration
The `config.json` defines:
- A user `alice` with a user-specific API key `alice-secret`.
- A service `echo-service` that exposes a profile `dev` with its own API key `dev-secret`.
- `alice` is granted access to the `dev` profile.

## Running the Example
1. Start the server:
   ```bash
   go run cmd/server/main.go --config examples/profile_example/config.json
   ```
2. In another terminal, start a mock upstream service (e.g., using `nc` or a simple Go server) on port 8081 if you want to test actual proxying, or just verify auth.

## Authentication Testing
You can authenticate using either the Profile Key (higher priority) or the User Key.

### 1. Using Profile Key
```bash
curl -H "X-Profile-Key: dev-secret" http://localhost:8080/mcp/u/alice/profile/dev/tools/list
```

### 2. Using User Key
```bash
curl -H "X-User-Key: alice-secret" http://localhost:8080/mcp/u/alice/profile/dev/tools/list
```

### 3. Invalid Key
```bash
curl -H "X-Profile-Key: wrong-secret" -i http://localhost:8080/mcp/u/alice/profile/dev/tools/list
```
Should return `401 Unauthorized (Profile)`.

### 4. No Key
```bash
curl -i http://localhost:8080/mcp/u/alice/profile/dev/tools/list
```
Should return `401 Unauthorized (User)` (since fallback to user auth fails).
