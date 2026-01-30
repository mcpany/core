# Context Optimizer Middleware

The Context Optimizer middleware automatically truncates large text outputs in JSON responses to prevent "Context Bloat" and reduce token usage.

## Configuration

To enable the context optimizer:

```yaml
context_optimizer:
  max_chars: 32000 # Maximum characters per text field (default: 32000, approx 8000 tokens)
```

## How it works

The middleware intercepts JSON responses. If it detects a `result.content` array with `text` fields exceeding the configured `max_chars`, it truncates them and appends a notice (e.g., `...[TRUNCATED X chars]`).

If `max_chars` is not set, it defaults to 32000.
