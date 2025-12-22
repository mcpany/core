// Package main implements a WebRTC server demo.

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Signal represents a WebRTC signal.
type Signal struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// handleWebSocket manages the WebRTC signaling over a WebSocket connection.
func handleWebSocket(conn *websocket.Conn) {
	defer func() { _ = conn.Close() }() // Ensure the WebSocket connection is closed when the handler exits

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Print("Failed to create peer connection:", err) // Changed from Fatal
		return
	}
	defer func() { _ = peerConnection.Close() }() // Ensure the peer connection is closed

	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		log.Print("Failed to create data channel:", err) // Changed from Fatal
		return
	}

	var wg sync.WaitGroup
	wg.Add(1) // Expecting one message for now

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
			log.Println("Error marshaling ICE candidate:", err)
			return
		}
		if err := conn.WriteJSON(Signal{Type: "candidate", Payload: string(payload)}); err != nil {
			log.Println("Error writing ICE candidate to WebSocket:", err)
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Printf("Failed to create offer: %v", err)
		return
	}
	// Changed from Fatal
	// return // This return is removed because log.Fatalf exits the program.

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		log.Print("Failed to set local description:", err) // Changed from Fatal
		return
	}

	payload, err := json.Marshal(offer)
	if err != nil {
		log.Print("Failed to marshal offer:", err) // Changed from Fatal
		return
	}
	if err := conn.WriteJSON(Signal{Type: "offer", Payload: string(payload)}); err != nil {
		log.Println("Error writing offer to WebSocket:", err)
		return
	}

	for {
		var signal Signal
		err := conn.ReadJSON(&signal)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("WebSocket closed normally:", err)
			} else {
				log.Println("Error reading JSON from WebSocket:", err)
			}
			return
		}

		switch signal.Type {
		case "answer":
			var answer webrtc.SessionDescription
			if err := json.Unmarshal([]byte(signal.Payload), &answer); err != nil {
				log.Println("Error unmarshaling answer:", err)
				continue
			}
			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println("Error setting remote description:", err)
			}
		case "candidate":
			var candidate webrtc.ICECandidateInit
			if err := json.Unmarshal([]byte(signal.Payload), &candidate); err != nil {
				log.Println("Error unmarshaling candidate:", err)
				continue
			}
			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println("Error adding ICE candidate:", err)
			}
		default:
			log.Printf("Received unknown signal type: %s", signal.Type)
		}
	}
}

func main() {
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
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
