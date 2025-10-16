/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/transformer"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/pion/webrtc/v3"
)

// WebrtcTool implements the Tool interface for a tool that is exposed via a
// WebRTC data channel. It handles the signaling and establishment of a peer
// connection to communicate with the remote service. This is useful for
// scenarios requiring low-latency, peer-to-peer communication directly from the
// server.
type WebrtcTool struct {
	tool              *v1.Tool
	poolManager       *pool.Manager
	serviceKey        string
	authenticator     auth.UpstreamAuthenticator
	parameterMappings []*configv1.WebrtcParameterMapping
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
}

// NewWebrtcTool creates a new WebrtcTool.
//
// tool is the protobuf definition of the tool.
// poolManager is used to get a client from the connection pool.
// serviceKey identifies the specific service connection pool.
// authenticator handles adding authentication credentials to the signaling request.
// callDefinition contains the configuration for the WebRTC call, such as
// parameter mappings and transformers.
func NewWebrtcTool(
	tool *v1.Tool,
	poolManager *pool.Manager,
	serviceKey string,
	authenticator auth.UpstreamAuthenticator,
	callDefinition *configv1.WebrtcCallDefinition,
) (*WebrtcTool, error) {
	return &WebrtcTool{
		tool:              tool,
		poolManager:       poolManager,
		serviceKey:        serviceKey,
		authenticator:     authenticator,
		parameterMappings: callDefinition.GetParameterMappings(),
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
	}, nil
}

// Tool returns the protobuf definition of the WebRTC tool.
func (t *WebrtcTool) Tool() *v1.Tool {
	return t.tool
}

// Execute handles the execution of the WebRTC tool. It establishes a new peer
// connection, negotiates the session via an HTTP signaling server, sends the
// tool inputs over the data channel, and waits for a response.
func (t *WebrtcTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	address := strings.TrimPrefix(t.tool.GetUnderlyingMethodFqn(), "WEBRTC ")

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}
	defer pc.Close()

	responseChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)

	dc, err := pc.CreateDataChannel("echo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create data channel: %w", err)
	}

	dc.OnOpen(func() {
		var inputs map[string]any
		if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
			fmt.Printf("failed to unmarshal tool inputs: %v\n", err)
			return
		}

		var message []byte
		var err error
		if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTextTemplate(t.inputTransformer.GetTemplate())
			if err != nil {
				fmt.Printf("failed to create input template: %v\n", err)
				return
			}
			rendered, err := tpl.Render(inputs)
			if err != nil {
				fmt.Printf("failed to render input template: %v\n", err)
				return
			}
			message = []byte(rendered)
		} else {
			message, err = json.Marshal(inputs)
			if err != nil {
				fmt.Printf("failed to marshal inputs to json: %v\n", err)
				return
			}
		}

		if err := dc.SendText(string(message)); err != nil {
			fmt.Printf("failed to send message over webrtc data channel: %v\n", err)
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
	defer resp.Body.Close()

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
			return parser.Parse(outputFormat, []byte(response), t.outputTransformer.GetExtractionRules())
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

// Close is a placeholder for any cleanup logic. Currently, it is a no-op as the
// peer connection is created and closed within the Execute method.
func (t *WebrtcTool) Close() error {
	return nil
}
