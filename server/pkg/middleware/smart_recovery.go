package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/llm"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// SmartRecoveryMiddleware handles automatic error recovery using LLM.
type SmartRecoveryMiddleware struct {
	config      *configv1.SmartRecoveryConfig
	llmClient   llm.Client
	toolManager tool.ManagerInterface
	mu          sync.RWMutex
}

// NewSmartRecoveryMiddleware creates a new SmartRecoveryMiddleware.
func NewSmartRecoveryMiddleware(config *configv1.SmartRecoveryConfig, toolManager tool.ManagerInterface) *SmartRecoveryMiddleware {
	return &SmartRecoveryMiddleware{
		config:      config,
		toolManager: toolManager,
	}
}

// Execute executes the middleware logic.
func (m *SmartRecoveryMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if m.config == nil || !m.config.GetEnabled() {
		return next(ctx, req)
	}

	m.mu.RLock()
	client := m.llmClient
	m.mu.RUnlock()

	// Initialize LLM client lazily or check if already initialized
	if client == nil {
		m.mu.Lock()
		// Double check
		if m.llmClient == nil {
			apiKey, err := util.ResolveSecret(ctx, m.config.GetApiKey())
			if err != nil {
				m.mu.Unlock()
				logging.GetLogger().Warn("SmartRecovery: Failed to resolve API key", "error", err)
				return next(ctx, req)
			}
			// Assuming OpenAI for now as per config.proto comments
			m.llmClient = llm.NewOpenAIClient(apiKey, m.config.GetBaseUrl())
		}
		// client = m.llmClient // Removed redundant assignment
		m.mu.Unlock()
	}

	maxRetries := int(m.config.GetMaxRetries())
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		res, err := next(ctx, req)
		if err == nil {
			if attempt > 0 {
				logging.GetLogger().Info("SmartRecovery: Successfully recovered from error", "tool", req.ToolName, "attempts", attempt)
			}
			return res, nil
		}
		lastErr = err

		if attempt == maxRetries {
			break
		}

		logging.GetLogger().Info("SmartRecovery: Tool execution failed, attempting recovery", "tool", req.ToolName, "error", err, "attempt", attempt+1)

		newArgs, recoveryErr := m.recover(ctx, req, err)
		if recoveryErr != nil {
			logging.GetLogger().Warn("SmartRecovery: Recovery failed", "error", recoveryErr)
			return nil, lastErr // Return original error
		}

		// Update request with new arguments
		req.Arguments = newArgs
		// Marshal to ToolInputs (which is what usually gets sent to upstreams)
		argsBytes, marshalErr := json.Marshal(newArgs)
		if marshalErr != nil {
			logging.GetLogger().Warn("SmartRecovery: Failed to marshal new arguments", "error", marshalErr)
			return nil, lastErr
		}
		req.ToolInputs = argsBytes
	}

	return nil, lastErr
}

func (m *SmartRecoveryMiddleware) recover(ctx context.Context, req *tool.ExecutionRequest, err error) (map[string]any, error) {
	// Serialize current arguments
	argsJSON, _ := json.Marshal(req.Arguments)
	if len(req.Arguments) == 0 && len(req.ToolInputs) > 0 {
		argsJSON = req.ToolInputs
	}

	prompt := fmt.Sprintf(`You are an expert at fixing tool execution errors.
Tool Name: %s
Arguments: %s
Error: %s

Analyze the error and the arguments. Provide corrected arguments in valid JSON format.
Output ONLY the JSON object of the arguments. Do not include markdown formatting like `+"```json"+` or explanation.`, req.ToolName, string(argsJSON), err.Error())

	// Use local client variable which is thread-safe
	m.mu.RLock()
	client := m.llmClient
	m.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("LLM client not initialized")
	}

	resp, llmErr := client.ChatCompletion(ctx, llm.ChatRequest{
		Model: m.config.GetModel(),
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
	})
	if llmErr != nil {
		return nil, fmt.Errorf("LLM error: %w", llmErr)
	}

	cleanedContent := cleanJSON(resp.Content)
	var newArgs map[string]any
	if parseErr := json.Unmarshal([]byte(cleanedContent), &newArgs); parseErr != nil {
		return nil, fmt.Errorf("failed to parse fixed arguments: %w. Content: %s", parseErr, cleanedContent)
	}

	return newArgs, nil
}

func cleanJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}
