package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Signal struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}

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

		dataChannel, err := peerConnection.CreateDataChannel("data", nil)
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		dataChannel.OnOpen(func() {
			log.Println("Data channel opened")
			dataChannel.SendText("Hello, world!")
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
				log.Println(err)
				return
			}
			conn.WriteJSON(Signal{Type: "candidate", Payload: string(payload)})
		})

		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			log.Fatal(err)
		}

		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			log.Fatal(err)
		}

		payload, err := json.Marshal(offer)
		if err != nil {
			log.Fatal(err)
		}
		conn.WriteJSON(Signal{Type: "offer", Payload: string(payload)})

		for {
			var signal Signal
			err := conn.ReadJSON(&signal)
			if err != nil {
				log.Println(err)
				return
			}

			switch signal.Type {
			case "answer":
				var answer webrtc.SessionDescription
				json.Unmarshal([]byte(signal.Payload), &answer)
				peerConnection.SetRemoteDescription(answer)
			case "candidate":
				var candidate webrtc.ICECandidateInit
				json.Unmarshal([]byte(signal.Payload), &candidate)
				peerConnection.AddICECandidate(candidate)
			}
		}
	})

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
