// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetLoggerConcurrency(t *testing.T) {
	ForTestsOnlyResetLogger()

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			GetLogger()
		}()
	}
	wg.Wait()
	assert.NotNil(t, GetLogger())
}

func TestInitJson(t *testing.T) {
	ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	Init(slog.LevelInfo, &buf, "json")
	log := GetLogger()
	log.Info("test")
	assert.Contains(t, buf.String(), "{")
	assert.Contains(t, buf.String(), `"msg":"test"`)
}

func TestBroadcastHandler_SourceLogic(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b)
	ctx := context.Background()

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// 1. Source from "source" attr
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.String("source", "my-source"))
	_ = h.Handle(ctx, r)

	select {
	case msg := <-ch:
		assert.Contains(t, string(msg), `"source":"my-source"`)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for log")
	}

	// 2. Source from "toolName" (higher priority)
	r = slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.String("source", "my-source"), slog.String("toolName", "my-tool"))
	_ = h.Handle(ctx, r)

	select {
	case msg := <-ch:
		assert.Contains(t, string(msg), `"source":"my-tool"`)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for log")
	}
}

func TestBroadcastHandler_Error(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b)
	ctx := context.Background()

	// Pass a function, which json.Marshal fails on
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.Any("bad", func() {}))
	err := h.Handle(ctx, r)
	assert.Error(t, err)
}

type failHandler struct{}

func (f failHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (f failHandler) Handle(_ context.Context, _ slog.Record) error { return errors.New("fail") }
func (f failHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return f }
func (f failHandler) WithGroup(_ string) slog.Handler               { return f }

func TestTeeHandlerError(t *testing.T) {
	h := NewTeeHandler(failHandler{}, failHandler{})
	err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0))
	assert.Error(t, err)
}

func TestTeeHandler_Methods(t *testing.T) {
	h := NewTeeHandler(failHandler{})
	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))

	h2 := h.WithAttrs(nil)
	assert.NotNil(t, h2)

	h3 := h.WithGroup("g")
	assert.NotNil(t, h3)
}
