package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/echo", echo)

	port := os.Getenv("WEBSOCKET_ECHO_SERVER_PORT")
	if port == "" {
		port = "8082"
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting echo server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
