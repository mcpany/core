# Profiles

Profiles allow you to categorize and selectively enable services based on the runtime environment (e.g., "dev", "prod", "staging").

## Configuration

Each service can define a list of `profiles` it belongs to.

```yaml
upstream_services:
  - name: "debug-service"
    profiles:
      - name: "dev"
    http_service:
      address: "http://localhost:8080"
```

The server is started with a set of active profiles (e.g., `--profiles=dev`). Services matching the active profiles are loaded; others are ignored.

## Use Case

You want to expose "dangerous" debugging tools only in your development environment, not in production. You tag those services with `dev` profile and only run with `--profiles=dev` locally.

## Public API Example

If you run with `--profiles=prod` and try to call a tool from the "debug-service", it will not be found because the service was never loaded.
