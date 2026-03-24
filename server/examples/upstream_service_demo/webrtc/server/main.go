// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

type Signal struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func handleWebSocket(conn *websocket.Conn) {
	defer func() { _ = conn.Close() }()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Print("Failed to create peer connection:", err)
		return
	}
	defer func() { _ = peerConnection.Close() }()

	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		log.Print("Failed to create data channel:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	dataChannel.OnOpen(func() {
		log.Println("Data channel opened")
		if err := dataChannel.SendText("Hello, world!"); err != nil {
			log.Println("Error sending text on data channel:", err)
		}
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Printf("Message from data channel: %s", string(msg.Data))
		wg.Done()
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		payload, err := json.Marshal(c.ToJSON())
		if err != nil {
			return
		}
		_ = conn.WriteJSON(Signal{Type: "candidate", Payload: string(payload)})
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		return
	}

	payload, err := json.Marshal(offer)
	if err != nil {
		return
	}
	if err := conn.WriteJSON(Signal{Type: "offer", Payload: string(payload)}); err != nil {
		return
	}

	for {
		var signal Signal
		err := conn.ReadJSON(&signal)
		if err != nil {
			return
		}

		switch signal.Type {
		case "answer":
			var answer webrtc.SessionDescription
			if err := json.Unmarshal([]byte(signal.Payload), &answer); err == nil {
				_ = peerConnection.SetRemoteDescription(answer)
			}
		case "candidate":
			var candidate webrtc.ICECandidateInit
			if err := json.Unmarshal([]byte(signal.Payload), &candidate); err == nil {
				_ = peerConnection.AddICECandidate(candidate)
			}
		}
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func run() error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		handleWebSocket(conn)
	})

	log.Println("Starting server on :8081")
	server := &http.Server{
		Addr:              ":8081",
		ReadHeaderTimeout: 3 * time.Second,
	}
	return server.ListenAndServe()
}
