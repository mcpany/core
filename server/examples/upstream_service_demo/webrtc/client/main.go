// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a demo WebRTC client.
package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// Signal represents a WebRTC signal.
type Signal struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func main() {
	var conn *websocket.Conn
	var err error
	// Retry connecting to the websocket server to avoid race conditions in tests.
	for i := 0; i < 5; i++ {
		conn, _, err = websocket.DefaultDialer.Dial("ws://localhost:8081/ws", nil)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to websocket: %v. Retrying in 1s...", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to websocket after multiple retries: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Printf("Failed to create peer connection: %v", err)
		return
	}

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			log.Println("Data channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from data channel: %s", string(msg.Data))
			if err := d.SendText("Hello, back!"); err != nil {
				log.Printf("Failed to send text: %v", err)
			}
		})
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		payload, err := json.Marshal(c.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}
		if err := conn.WriteJSON(Signal{Type: "candidate", Payload: string(payload)}); err != nil {
			log.Printf("Failed to write candidate: %v", err)
		}
	})

	go func() {
		for {
			var signal Signal
			err := conn.ReadJSON(&signal)
			if err != nil {
				log.Println(err)
				return
			}

			switch signal.Type {
			case "offer":
				var offer webrtc.SessionDescription
				if err := json.Unmarshal([]byte(signal.Payload), &offer); err != nil {
					log.Printf("Failed to unmarshal offer: %v", err)
					return
				}
				if err := peerConnection.SetRemoteDescription(offer); err != nil {
					log.Printf("Failed to set remote description: %v", err)
					return
				}
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					log.Printf("Failed to create answer: %v", err)
					return
				}
				if err := peerConnection.SetLocalDescription(answer); err != nil {
					log.Printf("Failed to set local description: %v", err)
					return
				}
				payload, err := json.Marshal(answer)
				if err != nil {
					log.Printf("Failed to marshal answer: %v", err)
					return
				}
				if err := conn.WriteJSON(Signal{Type: "answer", Payload: string(payload)}); err != nil {
					log.Printf("Failed to write answer: %v", err)
					return
				}
			case "candidate":
				var candidate webrtc.ICECandidateInit
				if err := json.Unmarshal([]byte(signal.Payload), &candidate); err != nil {
					log.Printf("Failed to unmarshal candidate: %v", err)
					return
				}
				if err := peerConnection.AddICECandidate(candidate); err != nil {
					log.Printf("Failed to add ice candidate: %v", err)
				}
			}
		}
	}()

	time.Sleep(5 * time.Second)
}
