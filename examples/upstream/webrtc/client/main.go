// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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
	defer conn.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			log.Println("Data channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from data channel: %s", string(msg.Data))
			d.SendText("Hello, back!")
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
		conn.WriteJSON(Signal{Type: "candidate", Payload: string(payload)})
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
				json.Unmarshal([]byte(signal.Payload), &offer)
				peerConnection.SetRemoteDescription(offer)
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					log.Fatal(err)
				}
				peerConnection.SetLocalDescription(answer)
				payload, err := json.Marshal(answer)
				if err != nil {
					log.Fatal(err)
				}
				conn.WriteJSON(Signal{Type: "answer", Payload: string(payload)})
			case "candidate":
				var candidate webrtc.ICECandidateInit
				json.Unmarshal([]byte(signal.Payload), &candidate)
				peerConnection.AddICECandidate(candidate)
			}
		}
	}()

	time.Sleep(5 * time.Second)
}
