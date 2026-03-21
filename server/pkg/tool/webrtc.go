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
	"github.com/pion/webrtc/v3"
)

type peerConnectionWrapper struct {
	*webrtc.PeerConnection
}

// Summary: Close closes the peer connection. Closes the peer connection.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (w *peerConnectionWrapper) Close() error {
	if w.PeerConnection == nil {
		return nil
	}
	return w.PeerConnection.Close()
}

// Summary: IsHealthy checks if the peer connection is in a usable state. Checks connection health.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (w *peerConnectionWrapper) IsHealthy(_ context.Context) bool {
	if w.PeerConnection == nil {
		return false
	}
	state := w.ICEConnectionState()
	return state == webrtc.ICEConnectionStateNew ||
		state == webrtc.ICEConnectionStateChecking ||
		state == webrtc.ICEConnectionStateConnected ||
		state == webrtc.ICEConnectionStateCompleted
}

// Summary: WebrtcTool implements the Tool interface for a tool that is exposed via a WebRTC data channel. WebRTC Tool implementation. It handles the signaling and establishment of a peer connection to communicate with the remote service. This is useful for scenarios requiring low-latency, peer-to-peer communication directly from the server.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type WebrtcTool struct {
	tool              *v1.Tool
	mcpTool           *mcp.Tool
	mcpToolOnce       sync.Once
	webrtcPool        pool.Pool[*peerConnectionWrapper]
	serviceID         string
	authenticator     auth.UpstreamAuthenticator
	parameters        []*configv1.WebrtcParameterMapping
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
	cache             *configv1.CacheConfig
}

// Summary: NewWebrtcTool creates a new WebrtcTool. Initializes a new WebrtcTool.
//
// Parameters:
//   - tool (*v1.Tool): The tool parameter.
//   - poolManager (*pool.Manager): The poolManager parameter.
//   - serviceID (string): The serviceID parameter.
//   - authenticator (auth.UpstreamAuthenticator): The authenticator parameter.
//   - callDefinition (*configv1.WebrtcCallDefinition): The callDefinition parameter.
//
// Returns:
//   - *WebrtcTool: The resulting *WebrtcTool.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
		if !found {
			var err error
			p, err = pool.New(t.newPeerConnection, 5, 5, 20, 1*time.Minute, false)
			if err != nil {
				return nil, fmt.Errorf("failed to create webrtc pool: %w", err)
			}
			poolManager.Register(serviceID, p)
		}
		t.webrtcPool = p
	}

	return t, nil
}

func (t *WebrtcTool) newPeerConnection(_ context.Context) (*peerConnectionWrapper, error) {
	iceServers := []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	}
	if os.Getenv("MCPANY_WEBRTC_DISABLE_STUN") == "true" {
		iceServers = []webrtc.ICEServer{}
	}

	config := webrtc.Configuration{
		ICEServers: iceServers,
	}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	return &peerConnectionWrapper{PeerConnection: pc}, nil
}

// Summary: Tool returns the protobuf definition of the WebRTC tool. Returns the protobuf tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *v1.Tool: The resulting *v1.Tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *WebrtcTool) Tool() *v1.Tool {
	return t.tool
}

// Summary: MCPTool returns the MCP tool definition. Returns the MCP tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Tool: The resulting *mcp.Tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *WebrtcTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// Summary: GetCacheConfig returns the cache configuration for the WebRTC tool. Returns the cache configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The resulting *configv1.CacheConfig.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *WebrtcTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Summary: Execute handles the execution of the WebRTC tool. Executes the WebRTC tool. It establishes a new peer connection (or reuses one), negotiates the session via an HTTP signaling server, sends the tool inputs over the data channel, and waits for a response.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*ExecutionRequest): The req parameter.
//
// Returns:
//   - any: The resulting any.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
			return response, nil
		}
		return result, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timed out waiting for webrtc response")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Summary: Close is a placeholder for any cleanup logic. Cleans up the WebrtcTool. Currently, it is a no-op as the peer connection is created and closed within the Execute method, unless a pool is used.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (t *WebrtcTool) Close() error {
	if t.webrtcPool != nil {
		_ = t.webrtcPool.Close()
	}
	return nil
}
