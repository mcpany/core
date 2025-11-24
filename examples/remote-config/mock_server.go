package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/config.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "remote_config.yaml")
	})

	log.Println("Starting mock server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
