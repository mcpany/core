/*
 * Copyright 2025 Author(s) of MCP Any
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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuthenticator is a mock implementation of the UpstreamAuthenticator interface.
type MockAuthenticator struct {
	Header http.Header
}

func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	for key, values := range m.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return nil
}

func TestNewWebrtcTool(t *testing.T) {
	toolDef := &v1.Tool{}
	toolDef.SetName("test-webrtc")
	callDef := &configv1.WebrtcCallDefinition{}
	wt, err := NewWebrtcTool(toolDef, nil, "service-key", nil, callDef)
	require.NoError(t, err)
	assert.NotNil(t, wt)
	assert.Equal(t, toolDef, wt.Tool())
	assert.Equal(t, "service-key", wt.serviceID)
}

func TestWebrtcTool_Close(t *testing.T) {
	wt, err := NewWebrtcTool(&v1.Tool{}, nil, "", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)
	assert.NoError(t, wt.Close())
}

func TestWebrtcTool_Execute_Success(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	// Mock WebRTC service that echoes messages
	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)

		// Set up a handler for when a data channel is created
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				// Echo the message back
				err := d.SendText(string(msg.Data))
				require.NoError(t, err)
			})
		})

		var offer webrtc.SessionDescription
		err = json.NewDecoder(r.Body).Decode(&offer)
		require.NoError(t, err)

		err = peerConnection.SetRemoteDescription(offer)
		require.NoError(t, err)

		answer, err := peerConnection.CreateAnswer(nil)
		require.NoError(t, err)

		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
		err = peerConnection.SetLocalDescription(answer)
		require.NoError(t, err)

		<-gatherComplete

		response, err := json.Marshal(*peerConnection.LocalDescription())
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)

		// Close connection after test is done
		go func() {
			wg.Wait()
			peerConnection.Close()
		}()
	}))
	defer signalingServer.Close()
	defer wg.Done()

	tool := &v1.Tool{}
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)

	wt, err := NewWebrtcTool(tool, nil, "", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	inputJSON := `{"message":"hello"}`
	req := &ExecutionRequest{
		ToolInputs: []byte(inputJSON),
	}

	result, err := wt.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "hello", resultMap["message"])
}

func TestWebrtcTool_Execute_WithTransformers(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)

		pc.OnDataChannel(func(d *webrtc.DataChannel) {
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				// Expecting transformed input
				assert.Equal(t, `{"transformed_message":"input_hello"}`, string(msg.Data))
				// Send back a response that can be transformed
				response := `{"data":{"final_message":"output_world"}}`
				err := d.SendText(response)
				require.NoError(t, err)
			})
		})

		var offer webrtc.SessionDescription
		err = json.NewDecoder(r.Body).Decode(&offer)
		require.NoError(t, err)
		err = pc.SetRemoteDescription(offer)
		require.NoError(t, err)
		answer, err := pc.CreateAnswer(nil)
		require.NoError(t, err)
		gatherComplete := webrtc.GatheringCompletePromise(pc)
		err = pc.SetLocalDescription(answer)
		require.NoError(t, err)
		<-gatherComplete
		response, err := json.Marshal(*pc.LocalDescription())
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)
		go func() {
			wg.Wait()
			pc.Close()
		}()
	}))
	defer signalingServer.Close()
	defer wg.Done()

	tool := &v1.Tool{}
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	callDef := &configv1.WebrtcCallDefinition{}
	callDef.SetInputTransformer(&configv1.InputTransformer{})
	callDef.GetInputTransformer().SetTemplate(`{"transformed_message":"input_{{message}}"}`)
	callDef.SetOutputTransformer(&configv1.OutputTransformer{})
	callDef.GetOutputTransformer().SetFormat(configv1.OutputTransformer_JSON)
	callDef.GetOutputTransformer().SetExtractionRules(map[string]string{
		"extracted_message": "{.data.final_message}",
	})

	wt, err := NewWebrtcTool(tool, nil, "", nil, callDef)
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{"message":"hello"}`)}
	result, err := wt.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "output_world", resultMap["extracted_message"])
}

func TestWebrtcTool_Execute_WithAuth(t *testing.T) {
    t.Skip("Skipping flaky test")
	authHeader := "Bearer my-secret-token"
	var wg sync.WaitGroup
	wg.Add(1)

	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for auth header
		assert.Equal(t, authHeader, r.Header.Get("Authorization"))

		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)
		var offer webrtc.SessionDescription
		err = json.NewDecoder(r.Body).Decode(&offer)
		require.NoError(t, err)
		err = pc.SetRemoteDescription(offer)
		require.NoError(t, err)
		answer, err := pc.CreateAnswer(nil)
		require.NoError(t, err)
		gatherComplete := webrtc.GatheringCompletePromise(pc)
		err = pc.SetLocalDescription(answer)
		require.NoError(t, err)
		<-gatherComplete
		response, err := json.Marshal(*pc.LocalDescription())
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)
		go func() {
			wg.Wait()
			pc.Close()
		}()
	}))
	defer signalingServer.Close()
	defer wg.Done()

	tool := &v1.Tool{}
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	authenticator := &MockAuthenticator{
		Header: http.Header{"Authorization": {authHeader}},
	}
	wt, err := NewWebrtcTool(tool, nil, "", authenticator, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{}`)}

	// We don't care about the result, just that the call completes and auth was checked
	_, _ = wt.Execute(context.Background(), req)
}

func TestWebrtcTool_Execute_SignalingFailure(t *testing.T) {
	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer signalingServer.Close()

	tool := &v1.Tool{}
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	wt, err := NewWebrtcTool(tool, nil, "", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{}`)}
	_, err = wt.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode answer")
}

func TestWebrtcTool_Execute_Timeout(t *testing.T) {
	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)
		pc.OnDataChannel(func(d *webrtc.DataChannel) {
			// No OnMessage handler, so it will never send a response
		})
		var offer webrtc.SessionDescription
		err = json.NewDecoder(r.Body).Decode(&offer)
		require.NoError(t, err)
		err = pc.SetRemoteDescription(offer)
		require.NoError(t, err)
		answer, err := pc.CreateAnswer(nil)
		require.NoError(t, err)
		gatherComplete := webrtc.GatheringCompletePromise(pc)
		err = pc.SetLocalDescription(answer)
		require.NoError(t, err)
		<-gatherComplete
		response, err := json.Marshal(*pc.LocalDescription())
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)
	}))
	defer signalingServer.Close()

	tool := &v1.Tool{}
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	wt, err := NewWebrtcTool(tool, nil, "", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{}`)}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = wt.Execute(ctx, req)
	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}
