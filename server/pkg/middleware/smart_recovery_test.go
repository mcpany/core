package middleware

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/llm"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockLLMClient
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) ChatCompletion(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.ChatResponse), args.Error(1)
}

func TestSmartRecoveryMiddleware_Execute(t *testing.T) {
	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:  "test-tool",
		Arguments: map[string]any{"arg": "bad"},
	}

	t.Run("Disabled", func(t *testing.T) {
		config := configv1.SmartRecoveryConfig_builder{
			Enabled: proto.Bool(false),
		}.Build()
		mw := NewSmartRecoveryMiddleware(config, nil)
		called := false
		next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
			called = true
			return "success", nil
		}
		res, err := mw.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", res)
		assert.True(t, called)
	})

	t.Run("Success first try", func(t *testing.T) {
		config := configv1.SmartRecoveryConfig_builder{
			Enabled: proto.Bool(true),
		}.Build()
		mw := NewSmartRecoveryMiddleware(config, nil)
		// No LLM needed
		next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
			return "success", nil
		}
		res, err := mw.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", res)
	})

	t.Run("Recover success", func(t *testing.T) {
		config := configv1.SmartRecoveryConfig_builder{
			Enabled:    proto.Bool(true),
			MaxRetries: proto.Int32(1),
			Model:      proto.String("gpt-4"),
		}.Build()
		mockLLM := new(MockLLMClient)
		mw := NewSmartRecoveryMiddleware(config, nil)
		mw.llmClient = mockLLM

		// Expectation
		mockLLM.On("ChatCompletion", mock.Anything, mock.MatchedBy(func(r llm.ChatRequest) bool {
			return r.Model == "gpt-4"
		})).Return(&llm.ChatResponse{
			Content: `{"arg": "good"}`,
		}, nil)

		attempts := 0
		next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("invalid argument")
			}
			// Second attempt should have corrected args
			if r.Arguments["arg"] == "good" {
				return "recovered", nil
			}
			return nil, errors.New("still bad")
		}

		res, err := mw.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "recovered", res)
		assert.Equal(t, 2, attempts)
		mockLLM.AssertExpectations(t)
	})

	t.Run("Recover failure (max retries)", func(t *testing.T) {
		config := configv1.SmartRecoveryConfig_builder{
			Enabled:    proto.Bool(true),
			MaxRetries: proto.Int32(1),
		}.Build()
		mockLLM := new(MockLLMClient)
		mw := NewSmartRecoveryMiddleware(config, nil)
		mw.llmClient = mockLLM

		mockLLM.On("ChatCompletion", mock.Anything, mock.Anything).Return(&llm.ChatResponse{
			Content: `{"arg": "still-bad"}`,
		}, nil)

		attempts := 0
		next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
			attempts++
			return nil, errors.New("fail")
		}

		res, err := mw.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, 2, attempts) // 1 initial + 1 retry
	})

	t.Run("LLM Failure", func(t *testing.T) {
		config := configv1.SmartRecoveryConfig_builder{
			Enabled:    proto.Bool(true),
			MaxRetries: proto.Int32(1),
		}.Build()
		mockLLM := new(MockLLMClient)
		mw := NewSmartRecoveryMiddleware(config, nil)
		mw.llmClient = mockLLM

		mockLLM.On("ChatCompletion", mock.Anything, mock.Anything).Return(nil, errors.New("llm down"))

		attempts := 0
		next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
			attempts++
			return nil, errors.New("fail")
		}

		res, err := mw.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Equal(t, "fail", err.Error()) // Should return original error
		assert.Nil(t, res)
		assert.Equal(t, 1, attempts) // Only initial attempt
	})
}
