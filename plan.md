Ah! `get_weather` does not take arguments!
```yaml
        - name: "get_weather"
          description: "Get current weather"
          call_id: "get_weather"
```
There's no `input_schema`. So if my Playwright test sends `get_weather {"location": "San Francisco"}`, it probably fails or succeeds with empty inputs depending on the router's validation. In either case, it's safer to just send `get_weather` or use `get_complex_data` which returns a complex table!
Using `get_complex_data` would test the rich UI renderer even better!

Let me update `ui/tests/audit-logs.spec.ts` to use `get_complex_data` without arguments, or maybe with `{}` if required, and then wait for `Success`.
