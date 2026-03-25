// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
// Close closes the peer connection.
//
// Summary: Closes the peer connection.
// IsHealthy checks if the peer connection is in a usable state.
//
// Summary: Checks connection health.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - _ (context.Context): Unused context parameter.
//
// Returns:
//   - bool: True if the connection state is valid (New, Checking, Connected, Completed).
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (w *peerConnectionWrapper) IsHealthy(_ context.Context) bool {
	if w.PeerConnection == nil {
// WebrtcTool implements the Tool interface for a tool that is exposed via a
// WebRTC data channel.
//
// Summary: WebRTC Tool implementation.
//
// It handles the signaling and establishment of a peer connection to communicate
// NewWebrtcTool creates a new WebrtcTool.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Summary: Initializes a new WebrtcTool.
//
// Parameters:
//   - tool (*v1.Tool): The protobuf definition of the tool.
//   - poolManager (*pool.Manager): Used to get a client from the connection pool.
//   - serviceID (string): Identifies the specific service connection pool.
//   - authenticator (auth.UpstreamAuthenticator): Handles adding authentication credentials to the signaling request.
//   - callDefinition (*configv1.WebrtcCallDefinition): Contains the configuration for the WebRTC call.
//
// Returns:
//   - (*WebrtcTool): The initialized WebrtcTool.
//   - (error): An error if initialization fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func NewWebrtcTool(
	tool *v1.Tool,
	poolManager *pool.Manager,
	serviceID string,
	authenticator auth.UpstreamAuthenticator,
	callDefinition *configv1.WebrtcCallDefinition,
) (*WebrtcTool, error) {
	t := &WebrtcTool{
		tool:              tool,
		serviceID:         serviceID,
		authenticator:     authenticator,
		parameters:        callDefinition.GetParameters(),
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
		cache:             callDefinition.GetCache(),
	}

	if poolManager != nil {
		p, found := pool.Get[*peerConnectionWrapper](poolManager, serviceID)
// MCPTool returns the MCP tool definition.
//
// Summary: Returns the MCP tool definition.
//
// Execute handles the execution of the WebRTC tool.
//
// Summary: Executes the WebRTC tool.
//
// It establishes a new peer connection (or reuses one), negotiates the session
// via an HTTP signaling server, sends the tool inputs over the data channel,
// and waits for a response.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*ExecutionRequest): The execution request.
//
// Returns:
//   - any: The result of the execution.
// StreamExecute executes the tool in streaming mode.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Executes the tool in streaming mode.
//
// Parameters:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (t *WebrtcTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
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

// Execute executes the WebRTC tool.
//
// Summary: Executes the WebRTC request and waits for the response.
//
// Parameters:
//   - ctx (context.Context): The context for execution.
//   - req (*ExecutionRequest): The request parameters.
//
// Returns:
//   - any: The response from the WebRTC endpoint.
//   - error: An error if the WebRTC communication fails.
//
// Errors:
//   - Returns an error if policy evaluation blocks the execution.
//   - Returns an error if marshalling the inputs fails.
//   - Returns an error if the WebRTC request fails.
//
// Side Effects:
//   - Makes a WebRTC network call.
func (t *WebrtcTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if t.webrtcPool == nil {
		// Fallback to creating a new connection if the pool is not initialized
		return t.executeWithoutPool(ctx, req)
	}

	wrapper, err := t.webrtcPool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer connection from pool: %w", err)
	}
	defer t.webrtcPool.Put(wrapper)

	return t.executeWithPeerConnection(ctx, req, wrapper.PeerConnection)
}

func (t *WebrtcTool) executeWithoutPool(ctx context.Context, req *ExecutionRequest) (any, error) {
	pc, err := t.newPeerConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}
	defer func() { _ = pc.Close() }()

	return t.executeWithPeerConnection(ctx, req, pc.PeerConnection)
}

func (t *WebrtcTool) executeWithPeerConnection(ctx context.Context, req *ExecutionRequest, pc *webrtc.PeerConnection) (any, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	for _, param := range t.parameters {
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			inputs[param.GetSchema().GetName()] = secretValue
		}
	}

	var message []byte
	var err error
	if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" { //nolint:staticcheck
		tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}") //nolint:staticcheck
		if err != nil {
			return nil, fmt.Errorf("failed to create input template: %w", err)
		}
		rendered, err := tpl.Render(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to render input template: %w", err)
		}
		message = []byte(rendered)
	} else {
		message, err = json.Marshal(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal inputs to json: %w", err)
		}
	}

	address := strings.TrimPrefix(t.tool.GetUnderlyingMethodFqn(), "WEBRTC ")

	responseChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)

	dc, err := pc.CreateDataChannel("echo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create data channel: %w", err)
	}

	dc.OnOpen(func() {
		if err := dc.SendText(string(message)); err != nil {
			logging.GetLogger().Warn("failed to send message over webrtc data channel", "error", err)
		}
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		responseChan <- string(msg.Data)
		wg.Done()
	})

	gatheringComplete := webrtc.GatheringCompletePromise(pc)

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	if err = pc.SetLocalDescription(offer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	<-gatheringComplete

	offerJSON, err := json.Marshal(pc.LocalDescription())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal offer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", address, bytes.NewReader(offerJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if t.authenticator != nil {
		if err := t.authenticator.Authenticate(httpReq); err != nil {
			return nil, fmt.Errorf("failed to authenticate request: %w", err)
		}
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send offer to signaling server: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var answer webrtc.SessionDescription
	if err := json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return nil, fmt.Errorf("failed to decode answer: %w", err)
	}

	if err := pc.SetRemoteDescription(answer); err != nil {
		return nil, fmt.Errorf("failed to set remote description: %w", err)
	}

	select {
	case response := <-responseChan:
		if t.outputTransformer != nil {
			parser := transformer.NewTextParser()
			outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
			return parser.Parse(outputFormat, []byte(response), t.outputTransformer.GetExtractionRules(), t.outputTransformer.GetJqQuery())
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(response), &result); err != nil {
// Close is a placeholder for any cleanup logic.
//
// Summary: Cleans up the WebrtcTool.
//
// Currently, it is a no-op as the peer connection is created and closed within
// the Execute method, unless a pool is used.
//
// Returns:
//   - error: Always nil.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (t *WebrtcTool) Close() error {
	if t.webrtcPool != nil {
		_ = t.webrtcPool.Close()
	}
	return nil
}
