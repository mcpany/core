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

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WeatherRequest struct {
	City string `json:"city"`
}

type WeatherResponse struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
	Time        string  `json:"time"`
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		if messageType == websocket.TextMessage {
			var req WeatherRequest
			if err := json.Unmarshal(p, &req); err != nil {
				log.Println("unmarshal error:", err)
				continue
			}
			resp := WeatherResponse{
				City:        req.City,
				Temperature: 25.5,
				Description: "Sunny",
				Time:        time.Now().Format(time.RFC3339),
			}
			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Println("marshal error:", err)
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
				log.Println("write error:", err)
				break
			}
		}
	}
}

func main() {
	port := flag.Int("port", 8091, "port to serve on")
	flag.Parse()

	http.HandleFunc("/weather", weatherHandler)
	log.Printf("Server starting on port %d", *port)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
