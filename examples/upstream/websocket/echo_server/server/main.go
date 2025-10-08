package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

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
	port := flag.String("port", "8082", "The port to listen on")
	flag.Parse()

	http.HandleFunc("/echo", echo)
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting echo server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
