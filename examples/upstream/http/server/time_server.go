package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		now := time.Now()
		resp["current_time"] = now.Format("2006-01-02 15:04:05")
		resp["timezone"] = now.Location().String()
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
	})

	port := os.Getenv("TIME_SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting time server on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
