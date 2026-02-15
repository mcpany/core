package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/pool"
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

func TestWebrtcTool_PoolInteraction(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
	var wg sync.WaitGroup
	wg.Add(2)

	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)

		pc.OnDataChannel(func(d *webrtc.DataChannel) {
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				err := d.SendText(string(msg.Data))
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
			_ = pc.Close()
		}()
	}))
	defer signalingServer.Close()

	tool := v1.Tool_builder{}.Build()
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	poolManager := pool.NewManager()
	wt, err := NewWebrtcTool(tool, poolManager, "webrtc-service", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	p, ok := pool.Get[*peerConnectionWrapper](poolManager, "webrtc-service")
	require.True(t, ok)

	// Execute twice to test pooling
	_, err = wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: []byte(`{"message":"test1"}`)})
	require.NoError(t, err)
	assert.Equal(t, 5, p.Len())

	_, err = wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: []byte(`{"message":"test2"}`)})
	require.NoError(t, err)
	assert.Equal(t, 5, p.Len())

	wg.Done()
	wg.Done()
}

func TestWebrtcTool_Execute_Success(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
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
			_ = peerConnection.Close()
		}()
	}))
	defer signalingServer.Close()
	defer wg.Done()

	tool := v1.Tool_builder{}.Build()
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)

	poolManager := pool.NewManager()
	wt, err := NewWebrtcTool(tool, poolManager, "webrtc-service", nil, &configv1.WebrtcCallDefinition{})
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
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
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
			_ = pc.Close()
		}()
	}))
	defer signalingServer.Close()
	defer wg.Done()

	tool := v1.Tool_builder{}.Build()
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	poolManager := pool.NewManager()
	callDef := &configv1.WebrtcCallDefinition{}
	callDef.SetInputTransformer(&configv1.InputTransformer{})
	callDef.GetInputTransformer().SetTemplate(`{"transformed_message":"input_{{message}}"}`)
	callDef.SetOutputTransformer(&configv1.OutputTransformer{})
	callDef.GetOutputTransformer().SetFormat(configv1.OutputTransformer_JSON)
	callDef.GetOutputTransformer().SetExtractionRules(map[string]string{
		"extracted_message": "{.data.final_message}",
	})

	wt, err := NewWebrtcTool(tool, poolManager, "webrtc-service", nil, callDef)
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{"message":"hello"}`)}
	result, err := wt.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "output_world", resultMap["extracted_message"])
}

func TestWebrtcTool_Execute_WithAuth(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
	authHeader := "Bearer my-secret-token"
	var wg sync.WaitGroup
	wg.Add(1)

	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for auth header
		assert.Equal(t, authHeader, r.Header.Get("Authorization"))

		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)

		pc.OnDataChannel(func(d *webrtc.DataChannel) {
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				// Echo message to unblock execution
				err := d.SendText(string(msg.Data))
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
			_ = pc.Close()
		}()
	}))
	defer signalingServer.Close()
	defer wg.Done()

	tool := v1.Tool_builder{}.Build()
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	poolManager := pool.NewManager()
	authenticator := &MockAuthenticator{
		Header: http.Header{"Authorization": {authHeader}},
	}
	wt, err := NewWebrtcTool(tool, poolManager, "webrtc-service", authenticator, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{}`)}

	// We don't care about the result, just that the call completes and auth was checked
	_, _ = wt.Execute(context.Background(), req)
}

func TestWebrtcTool_Execute_SignalingFailure(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer signalingServer.Close()

	tool := v1.Tool_builder{}.Build()
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	poolManager := pool.NewManager()
	wt, err := NewWebrtcTool(tool, poolManager, "webrtc-service", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{}`)}
	_, err = wt.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode answer")
}

func TestWebrtcTool_Execute_Timeout(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
	signalingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		require.NoError(t, err)
		pc.OnDataChannel(func(_ *webrtc.DataChannel) {
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

	tool := v1.Tool_builder{}.Build()
	tool.SetUnderlyingMethodFqn("WEBRTC " + signalingServer.URL)
	poolManager := pool.NewManager()
	wt, err := NewWebrtcTool(tool, poolManager, "webrtc-service", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)

	req := &ExecutionRequest{ToolInputs: []byte(`{}`)}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = wt.Execute(ctx, req)
	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestWebrtcTool_GetCacheConfig(t *testing.T) {
	toolDef := &v1.Tool{}
	cacheConfig := &configv1.CacheConfig{}
	callDef := &configv1.WebrtcCallDefinition{}
	callDef.SetCache(cacheConfig)
	wt, err := NewWebrtcTool(toolDef, nil, "service-key", nil, callDef)
	require.NoError(t, err)
	assert.Equal(t, cacheConfig, wt.GetCacheConfig())
}

func TestWebrtcTool_Execute_InvalidInputTemplate(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
	toolDef := &v1.Tool{}
	poolManager := pool.NewManager()
	callDef := &configv1.WebrtcCallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{ .invalid }}")
	callDef.SetInputTransformer(inputTransformer)
	wt, err := NewWebrtcTool(toolDef, poolManager, "webrtc-service", nil, callDef)
	require.NoError(t, err)

	req := &ExecutionRequest{
		ToolInputs: []byte(`{"message":"hello"}`),
	}

	_, err = wt.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render input template")
}

func TestWebrtcTool_CloseMethod(t *testing.T) {
	wt, err := NewWebrtcTool(&v1.Tool{}, nil, "", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)
	assert.NoError(t, wt.Close())
}
