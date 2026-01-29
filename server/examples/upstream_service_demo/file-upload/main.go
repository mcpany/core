// Package main implements a file upload demo server.

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "failed to get file from form", http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()

		// Create a temporary file to store the uploaded content
		tmpfile, err := os.CreateTemp("", "upload-*.txt")
		if err != nil {
			http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
			return
		}
		defer func() { _ = os.Remove(tmpfile.Name()) }() // clean up

		// Copy the uploaded file to the temporary file
		if _, err := io.Copy(tmpfile, file); err != nil {
			http.Error(w, "failed to copy file", http.StatusInternalServerError)
			return
		}

		// Respond with the file name and size
		_, _ = fmt.Fprintf(w, "File '%s' uploaded successfully (size: %d bytes)", header.Filename, header.Size)
	})

	fmt.Println("Server started on :8082")
	server := &http.Server{
		Addr:              ":8082",
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
