package logging

import (
	"io"
	"log/slog"
	"testing"
)

func BenchmarkLoggerAddSource(b *testing.B) {
	b.Run("WithAddSource", func(b *testing.B) {
		h := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: true})
		l := slog.New(h)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Info("benchmark message", "key", "value")
		}
	})

	b.Run("WithoutAddSource", func(b *testing.B) {
		h := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false})
		l := slog.New(h)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Info("benchmark message", "key", "value")
		}
	})
}
