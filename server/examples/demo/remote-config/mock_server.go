// Package main implements a mock server for remote configuration.

package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/config.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "remote_config.yaml")
	})

	log.Println("Starting mock server on :8080")
	server := &http.Server{
		Addr:         ":8080",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
