// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LocalCommandTool implements the Tool interface for a tool that is executed as a
// local command-line process.
//
// Summary: Tool implementation for local command-line services.
//
// It maps tool inputs to command-line arguments and environment variables.
type LocalCommandTool struct {
	tool           *v1.Tool
	mcpTool        *mcp.Tool
	mcpToolOnce    sync.Once
	service        *configv1.CommandLineUpstreamService
	callDefinition *configv1.CommandLineCallDefinition
	policies       []*CompiledCallPolicy
	callID         string
	sandboxArgs    []string
	initError      error
	allowedParams  map[string]bool
}

// NewLocalCommandTool creates a new LocalCommandTool instance.
//
// Summary: Initializes a new LocalCommandTool.
//
// Parameters:
//   - tool: *v1.Tool. The protobuf definition of the tool.
//   - service: *configv1.CommandLineUpstreamService. The service configuration.
//   - callDefinition: *configv1.CommandLineCallDefinition. The call configuration.
//   - policies: []*configv1.CallPolicy. The security policies.
//   - callID: string. The unique identifier for the call.
//
// Returns:
//   - Tool: The created LocalCommandTool.
func NewLocalCommandTool(
	tool *v1.Tool,
	service *configv1.CommandLineUpstreamService,
	callDefinition *configv1.CommandLineCallDefinition,
	policies []*configv1.CallPolicy,
	callID string,
) Tool {
	compiled, err := CompileCallPolicies(policies)

	allowedParams := make(map[string]bool)
	if callDefinition != nil {
		for _, param := range callDefinition.GetParameters() {
			if schema := param.GetSchema(); schema != nil {
				allowedParams[schema.GetName()] = true
			}
		}
	}

	t := &LocalCommandTool{
		tool:           tool,
		service:        service,
		callDefinition: callDefinition,
		policies:       compiled,
		callID:         callID,
		allowedParams:  allowedParams,
	}
	if err != nil {
		t.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}

	cmd := service.GetCommand()
	base := filepath.Base(cmd)
	if base == "sed" || base == "gsed" {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		checkCmd := exec.CommandContext(ctx, cmd, "--sandbox", "--version")
		if err := checkCmd.Run(); err == nil {
			t.sandboxArgs = []string{"--sandbox"}
			logging.GetLogger().Info("Enabled sandbox mode for sed tool", "tool", tool.GetName())
		} else {
			t.initError = fmt.Errorf("sed tool %q detected but --sandbox is not supported (error: %v); execution blocked for security", tool.GetName(), err)
			logging.GetLogger().Error("Failed to enable sandbox for sed", "tool", tool.GetName(), "error", err)
		}
	}

	return t
}

// Tool returns the protobuf definition of the command-line tool.
//
// Returns:
//   - *v1.Tool: The underlying protobuf definition.
//
// Summary: Executes Tool operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (t *LocalCommandTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Summary: Executes MCPTool operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (t *LocalCommandTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig returns the cache configuration for the command-line tool.
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration, if any.
//
// Summary: Retrieves GetCacheConfig operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (t *LocalCommandTool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDefinition == nil {
		return nil
	}
	return t.callDefinition.GetCache()
}

// Execute handles the execution of the command-line tool.
//
// Summary: Executes the local command-line tool.
//
// It constructs a command with arguments and environment variables derived from the tool inputs,
// runs the command, and returns its output.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Side Effects:
//   - Executes a subprocess on the local system.
//   - Consumes system resources (CPU, memory).
//   - Logs execution details.
//
// IsStreaming returns true if the tool supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *LocalCommandTool) IsStreaming() bool {
	return false
}

// StreamExecute executes the tool in streaming mode.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
func (t *LocalCommandTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	ch := make(chan any, 1)
	go func() {
		defer close(ch)
		res, err := t.Execute(ctx, req)
		if err != nil {
			ch <- err
		} else {
			ch <- res
		}
	}()
	return ch, nil
}

// Execute handles the execution of the command-line tool.
//
// Summary: Executes the local command-line tool.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - Returns an error if initialization failed.
//   - Returns an error if policy evaluation blocks the execution.
//   - Returns an error if argument substitution or validation fails.
//   - Returns an error if shell injection is detected.
//   - Returns an error if the command execution fails.
//
// Side Effects:
//   - Executes a local command line process.
func (t *LocalCommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}
	var inputs map[string]any

	if len(bytes.TrimSpace(req.ToolInputs)) == 0 {
		req.ToolInputs = []byte("{}")
	}

	if err := fastJSONNumber.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	for k := range inputs {
		if k == "args" {
			continue
		}
		if !t.allowedParams[k] {
			delete(inputs, k)
		}
	}

	args := []string{}
	if len(t.sandboxArgs) > 0 {
		args = append(args, t.sandboxArgs...)
	}

	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	isDocker := t.service.GetContainerEnvironment() != nil && t.service.GetContainerEnvironment().GetImage() != ""

	commandName := t.service.GetCommand()

	if inputs != nil {
		for i := range args {
			for k, v := range inputs {
				placeholder := "{{" + k + "}}"
				if strings.Contains(args[i], placeholder) {
					val := util.ToString(v)
					if err := validateSafePathAndInjection(val, isDocker, commandName); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}

					cmd := t.service.GetCommand()
					if isShellCommand(cmd) {
						if err := checkForShellInjection(val, args[i], placeholder, cmd, isShell(cmd)); err != nil {
							return nil, fmt.Errorf("parameter %q: %w", k, err)
						}
					}
					args[i] = strings.ReplaceAll(args[i], placeholder, val)
				}
			}
		}
	}

	if inputs != nil {
		if argsVal, ok := inputs["args"]; ok {

			argsAllowed := false
			if inputSchema := t.tool.GetInputSchema(); inputSchema != nil {
				if props := inputSchema.Fields["properties"].GetStructValue(); props != nil {
					if _, ok := props.Fields["args"]; ok {
						argsAllowed = true
					}
				}
			}

			if !argsAllowed {
				return nil, fmt.Errorf("'args' parameter is not allowed for this tool")
			}

			if argsList, ok := argsVal.([]any); ok {
				for _, arg := range argsList {
					if argStr, ok := arg.(string); ok {
						if err := validateSafePathAndInjection(argStr, isDocker, commandName); err != nil {
							return nil, fmt.Errorf("args parameter: %w", err)
						}

						cmd := t.service.GetCommand()
						if isShellCommand(cmd) {
							if err := checkForShellInjection(argStr, "", "", cmd, isShell(cmd)); err != nil {
								return nil, fmt.Errorf("args parameter: %w", err)
							}
						}
						args = append(args, argStr)
					} else {
						return nil, fmt.Errorf("non-string value in 'args' array")
					}
				}
			} else {
				return nil, fmt.Errorf("'args' parameter must be an array of strings")
			}
			delete(inputs, "args")
		}
	}

	if filepath.Base(t.service.GetCommand()) == gitCommand {
		for _, arg := range args {

			if strings.Contains(arg, "ext::") {
				return nil, fmt.Errorf("git ext:: protocol is not allowed")
			}
		}
	}

	timeout := t.service.GetTimeout()
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
		defer cancel()
	}

	executor := command.NewLocalExecutor()

	env := []string{}

	if !isDocker {

		allowedEnvVars := []string{"PATH", "HOME", "USER", "SHELL", "TMPDIR", "SYSTEMROOT", "WINDIR"}
		for _, key := range allowedEnvVars {
			if val, ok := os.LookupEnv(key); ok {
				env = append(env, fmt.Sprintf("%s=%s", key, val))
			}
		}
	}

	resolvedServiceEnv, err := util.ResolveSecretMap(ctx, t.service.GetEnv(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service env: %w", err)
	}

	secrets := make([]string, 0, len(resolvedServiceEnv))
	for _, v := range resolvedServiceEnv {
		secrets = append(secrets, v)
	}

	for k, v := range resolvedServiceEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	for _, param := range t.callDefinition.GetParameters() {
		name := param.GetSchema().GetName()
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", name, err)
			}
			secrets = append(secrets, secretValue)
			env = append(env, fmt.Sprintf("%s=%s", name, secretValue))
		} else if val, ok := inputs[name]; ok {
			valStr := util.ToString(val)
			if err := validateSafePathAndInjection(valStr, isDocker, commandName); err != nil {
				return nil, fmt.Errorf("parameter %q: %w", name, err)
			}

			safeForEnv := true
			if isShellCommand(t.service.GetCommand()) {
				if err := checkEnvInjection(valStr); err != nil {
					logging.GetLogger().Warn("Skipping environment variable due to potential shell injection risk", "parameter", name, "error", err)
					safeForEnv = false
				}
			}

			if isDangerousEnvVar(name) {
				logging.GetLogger().Warn("Skipping dangerous environment variable", "parameter", name)
				safeForEnv = false
			}

			if safeForEnv {
				env = append(env, fmt.Sprintf("%s=%s", name, valStr))
			}

			if util.IsSensitiveKey(name) {
				secrets = append(secrets, valStr)
			}
		}
	}

	if req.DryRun {
		logging.GetLogger().Info("Dry run execution", "tool", req.ToolName)
		return map[string]any{
			"dry_run": true,
			"request": map[string]any{
				"command": t.service.GetCommand(),
				"args":    args,
				"env":     env,
			},
		}, nil
	}

	startTime := time.Now()
	limit := getMaxCommandOutputSize()

	redactor := util.NewSecretRedactor(secrets)

	redactedArgs := make([]string, len(args))
	for i, arg := range args {
		redactedArgs[i] = redactor.Redact(arg)
	}

	if t.service.GetCommunicationProtocol() == configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON {
		stdin, stdout, stderr, _, err := executor.ExecuteWithStdIO(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
		if err != nil {

			logging.GetLogger().ErrorContext(ctx, "Failed to execute JSON CLI command with stdio", "tool", t.tool.GetName(), "error", err)
			return nil, fmt.Errorf("failed to execute JSON CLI command")
		}

		var stderrBuf bytes.Buffer
		stderrDone := make(chan struct{})
		go func() {
			defer close(stderrDone)
			defer func() { _ = stderr.Close() }()
			_, _ = io.Copy(&stderrBuf, io.LimitReader(stderr, limit))
		}()

		var unmarshaledInputs map[string]interface{}
		decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&unmarshaledInputs); err != nil {
			_ = stdin.Close()
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}

		go func() {
			defer func() { _ = stdin.Close() }()
			if err := fastJSON.NewEncoder(stdin).Encode(unmarshaledInputs); err != nil {
				logging.GetLogger().Warn("Failed to encode inputs to stdin", "error", err)
			}
		}()

		var result map[string]interface{}
		if err := fastJSON.NewDecoder(io.LimitReader(stdout, limit)).Decode(&result); err != nil {
			<-stderrDone

			redactedStderr := redactor.Redact(stderrBuf.String())
			if redactedStderr != "" {
				logging.GetLogger().WarnContext(ctx, "JSON CLI command failed with stderr", "tool", t.tool.GetName(), "stderr", redactedStderr)
			}
			return nil, fmt.Errorf("failed to execute JSON CLI command: %w", err)
		}
		return result, nil
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
	if err != nil {

		logging.GetLogger().ErrorContext(ctx, "Failed to execute CLI command", "tool", t.tool.GetName(), "error", err)
		return nil, fmt.Errorf("failed to execute command")
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var combinedBuf threadSafeBuffer
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() { _ = stdout.Close() }()
		_, _ = io.Copy(io.MultiWriter(&stdoutBuf, &combinedBuf), io.LimitReader(stdout, limit))
	}()
	go func() {
		defer wg.Done()
		defer func() { _ = stderr.Close() }()
		_, _ = io.Copy(io.MultiWriter(&stderrBuf, &combinedBuf), io.LimitReader(stderr, limit))
	}()

	wg.Wait()
	exitCode := <-exitCodeChan
	endTime := time.Now()

	status := consts.CommandStatusSuccess
	if ctx.Err() == context.DeadlineExceeded {
		status = consts.CommandStatusTimeout
		exitCode = -1
	} else if exitCode != 0 {
		status = consts.CommandStatusError
	}

	result := map[string]interface{}{
		"command":         t.service.GetCommand(),
		"args":            redactedArgs,
		"stdout":          redactor.Redact(stdoutBuf.String()),
		"stderr":          redactor.Redact(stderrBuf.String()),
		"combined_output": redactor.Redact(combinedBuf.String()),
		"start_time":      startTime,
		"end_time":        endTime,
		"return_code":     exitCode,
		"status":          status,
	}

	return result, nil
}

type threadSafeBuffer struct {
	b  bytes.Buffer
	mu sync.Mutex
}

func (tsb *threadSafeBuffer) Write(p []byte) (n int, err error) {
	tsb.mu.Lock()
	defer tsb.mu.Unlock()
	return tsb.b.Write(p)
}

func (tsb *threadSafeBuffer) String() string {
	tsb.mu.Lock()
	defer tsb.mu.Unlock()
	return tsb.b.String()
}
