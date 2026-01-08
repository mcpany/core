# Context Optimizer Middleware

The Context Optimizer middleware automatically truncates large text outputs in JSON responses to prevent "Context Bloat" and reduce token usage.

## Configuration

To enable the context optimizer:

```yaml
context_optimizer:
  max_chars: 1000 # Maximum characters per text field
```

## How it works

The middleware intercepts JSON responses. If it detects a `result.content` array with `text` fields exceeding the configured `max_chars`, it truncates them and appends a notice (e.g., `...[TRUNCATED X chars]`).
