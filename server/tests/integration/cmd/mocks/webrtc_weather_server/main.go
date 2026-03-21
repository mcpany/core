// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a mock WebRTC weather server for testing.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

var (
	peerConnection *webrtc.PeerConnection
	mu             sync.Mutex
)

var weatherData = map[string]string{
	"new york": "Sunny, 25°C",
	"london":   "Cloudy, 15°C",
	"tokyo":    "Rainy, 20°C",
}

func signalHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	slog.Info("webrtc_weather_server: Received signal request")
	var offer webrtc.SessionDescription
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("webrtc_weather_server: Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &offer); err != nil {
		slog.Error("webrtc_weather_server: Failed to unmarshal offer", "error", err)
		http.Error(w, "Failed to unmarshal offer", http.StatusBadRequest)
		return
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		slog.Error("webrtc_weather_server: Failed to set remote description", "error", err)
		http.Error(w, "Failed to set remote description", http.StatusInternalServerError)
		return
	}

	gatheringComplete := webrtc.GatheringCompletePromise(peerConnection)

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		slog.Error("webrtc_weather_server: Failed to create answer", "error", err)
		http.Error(w, "Failed to create answer", http.StatusInternalServerError)
		return
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		slog.Error("webrtc_weather_server: Failed to set local description", "error", err)
		http.Error(w, "Failed to set local description", http.StatusInternalServerError)
		return
	}

	<-gatheringComplete

	response, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		slog.Error("webrtc_weather_server: Failed to marshal answer", "error", err)
		http.Error(w, "Failed to marshal answer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		slog.Error("webrtc_weather_server: Failed to write response", "error", err)
	}
	slog.Info("webrtc_weather_server: Sent answer")
}

func setupPeerConnection() error {
	var err error
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		slog.Info("webrtc_weather_server: New DataChannel", "label", d.Label(), "id", d.ID())
		d.OnOpen(func() {
			slog.Info("webrtc_weather_server: Data channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			slog.Info("webrtc_weather_server: Received message", "message", string(msg.Data))
			weather, ok := weatherData[string(msg.Data)]
			if !ok {
				weather = "Location not found"
			}
			if err := d.SendText(weather); err != nil {
				slog.Error("webrtc_weather_server: Failed to send message", "error", err)
			}
			slog.Info("webrtc_weather_server: Sent weather", "weather", weather)
		})
		d.OnClose(func() {
			slog.Info("webrtc_weather_server: Data channel closed")
		})
	})

	return nil
}

// main starts the mock WebRTC weather server.
func main() {
	if err := setupPeerConnection(); err != nil {
		slog.Error("webrtc_weather_server: Failed to setup peer connection", "error", err)
		os.Exit(1)
	}

	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		slog.Error("webrtc_weather_server: Failed to listen on a port", "error", err)
		os.Exit(1)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	slog.Info("webrtc_weather_server: Listening on port", "port", actualPort)

	if *port == 0 {
		fmt.Printf("%d\n", actualPort)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/signal", signalHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		slog.Error("webrtc_weather_server: Server failed", "error", err)
		os.Exit(1)
	}

	slog.Info("webrtc_weather_server: Server shut down.")
}
