// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Signal struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var conn *websocket.Conn
	var err error
	for i := 0; i < 5; i++ {
		conn, _, err = websocket.DefaultDialer.Dial("ws://localhost:8081/ws", nil)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to websocket: %v. Retrying in 1s...", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return err
	}
	defer func() { _ = peerConnection.Close() }()

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			log.Println("Data channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from data channel: %s", string(msg.Data))
			_ = d.SendText("Hello, back!")
		})
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		payload, _ := json.Marshal(c.ToJSON())
		_ = conn.WriteJSON(Signal{Type: "candidate", Payload: string(payload)})
	})

	go func() {
		for {
			var signal Signal
			err := conn.ReadJSON(&signal)
			if err != nil {
				return
			}

			switch signal.Type {
			case "offer":
				var offer webrtc.SessionDescription
				_ = json.Unmarshal([]byte(signal.Payload), &offer)
				_ = peerConnection.SetRemoteDescription(offer)
				answer, _ := peerConnection.CreateAnswer(nil)
				_ = peerConnection.SetLocalDescription(answer)
				payload, _ := json.Marshal(answer)
				_ = conn.WriteJSON(Signal{Type: "answer", Payload: string(payload)})
			case "candidate":
				var candidate webrtc.ICECandidateInit
				_ = json.Unmarshal([]byte(signal.Payload), &candidate)
				_ = peerConnection.AddICECandidate(candidate)
			}
		}
	}()

	time.Sleep(5 * time.Second)
	return nil
}
