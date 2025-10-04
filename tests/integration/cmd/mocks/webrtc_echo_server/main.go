/*
 * Copyright 2025 Author(s) of MCPXY
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/pion/webrtc/v3"
)

var (
	peerConnection *webrtc.PeerConnection
	mu             sync.Mutex
)

func signalHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	slog.Info("webrtc_echo_server: Received signal request")
	var offer webrtc.SessionDescription
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("webrtc_echo_server: Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &offer); err != nil {
		slog.Error("webrtc_echo_server: Failed to unmarshal offer", "error", err)
		http.Error(w, "Failed to unmarshal offer", http.StatusBadRequest)
		return
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		slog.Error("webrtc_echo_server: Failed to set remote description", "error", err)
		http.Error(w, "Failed to set remote description", http.StatusInternalServerError)
		return
	}

	gatheringComplete := webrtc.GatheringCompletePromise(peerConnection)

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		slog.Error("webrtc_echo_server: Failed to create answer", "error", err)
		http.Error(w, "Failed to create answer", http.StatusInternalServerError)
		return
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		slog.Error("webrtc_echo_server: Failed to set local description", "error", err)
		http.Error(w, "Failed to set local description", http.StatusInternalServerError)
		return
	}

	<-gatheringComplete

	response, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		slog.Error("webrtc_echo_server: Failed to marshal answer", "error", err)
		http.Error(w, "Failed to marshal answer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		slog.Error("webrtc_echo_server: Failed to write response", "error", err)
	}
	slog.Info("webrtc_echo_server: Sent answer")
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
		slog.Info("webrtc_echo_server: New DataChannel", "label", d.Label(), "id", d.ID())
		d.OnOpen(func() {
			slog.Info("webrtc_echo_server: Data channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			slog.Info("webrtc_echo_server: Received message", "message", string(msg.Data))
			if err := d.SendText(string(msg.Data)); err != nil {
				slog.Error("webrtc_echo_server: Failed to send message", "error", err)
			}
			slog.Info("webrtc_echo_server: Echoed message", "message", string(msg.Data))
		})
		d.OnClose(func() {
			slog.Info("webrtc_echo_server: Data channel closed")
		})
	})

	return nil
}

// main starts the mock WebRTC echo server.
func main() {
	if err := setupPeerConnection(); err != nil {
		slog.Error("webrtc_echo_server: Failed to setup peer connection", "error", err)
		os.Exit(1)
	}

	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("webrtc_echo_server: Failed to listen on a port", "error", err)
		os.Exit(1)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	slog.Info("webrtc_echo_server: Listening on port", "port", actualPort)

	if *port == 0 {
		fmt.Printf("%d\n", actualPort)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/signal", signalHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	server := &http.Server{
		Handler: mux,
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		slog.Error("webrtc_echo_server: Server failed", "error", err)
		os.Exit(1)
	}

	slog.Info("webrtc_echo_server: Server shut down.")
}
