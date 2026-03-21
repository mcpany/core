Ah, the `golangci-lint` is OOM killing in my environment. But in CircleCI it might be printing an actual error!
If it's printing an actual error, what could it be?
Let's manually inspect `server/pkg/app/seed.go`.
```go
	if len(req.SettingsRaw) > 0 {
		settings := configv1.GlobalSettings_builder{}.Build()
		if err := protojson.Unmarshal(req.SettingsRaw, settings); err != nil {
			return fmt.Errorf("invalid json for settings")
		}
		err := withRetry(ctx, logging.GetLogger(), func() error {
			return a.Storage.SaveGlobalSettings(ctx, settings)
		})
		if err != nil {
			return fmt.Errorf("failed to save global settings: %w", err)
		}
	}
```
Wait! `a.Storage.SaveGlobalSettings` is called.
BUT `seed.go` imports `configv1 "github.com/mcpany/core/proto/config/v1"`.
And `a.Storage.SaveGlobalSettings(ctx, settings)` returns an error.
Wait, is `a.Storage.SaveGlobalSettings` expecting `*configv1.GlobalSettings`? Yes.
Is there ANY lint issue?
Let me run `go build` for `server/pkg/app`.
