
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const maxUploadSize = 10 * 1024 // 10 KB

// fileCreator defines an interface for creating temporary files.
type fileCreator interface {
	CreateTemp(dir, pattern string) (*os.File, error)
}

// osFileCreator is a concrete implementation of fileCreator that uses os.CreateTemp.
type osFileCreator struct{}

func (c osFileCreator) CreateTemp(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

// copier defines an interface for copying files.
type copier interface {
	Copy(dst io.Writer, src io.Reader) (int64, error)
}

// ioCopier is a concrete implementation of copier that uses io.Copy.
type ioCopier struct{}

func (c ioCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

// uploader holds the dependencies for creating and copying files.
type uploader struct {
	creator fileCreator
	copier  copier
}

// uploadHandler is the http.HandlerFunc for uploading files.
func (u *uploader) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enforce a maximum upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary file to store the uploaded content
	tmpfile, err := u.creator.CreateTemp("", "upload-*.txt")
	if err != nil {
		http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpfile.Name()) // clean up immediately

	// Copy the uploaded file to the temporary file
	bytesWritten, err := u.copier.Copy(tmpfile, file)
	if err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}

	// Respond with the file name and actual size
	fmt.Fprintf(w, "File '%s' uploaded successfully (size: %d bytes)", header.Filename, bytesWritten)
}

func main() {
	uploader := &uploader{
		creator: osFileCreator{},
		copier:  ioCopier{},
	}
	http.HandleFunc("/upload", uploader.uploadHandler)

	fmt.Println("Server started on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
