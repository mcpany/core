
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockFileCreator struct {
	shouldError bool
}

func (m mockFileCreator) CreateTemp(dir, pattern string) (*os.File, error) {
	if m.shouldError {
		return nil, errors.New("failed to create temp file")
	}
	return os.CreateTemp(dir, pattern)
}

type mockCopier struct {
	shouldError bool
}

func (m mockCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	if m.shouldError {
		return 0, errors.New("failed to copy file")
	}
	return io.Copy(dst, src)
}

// TestUploadHandler_Success tests the happy path of the upload handler.
func TestUploadHandler_Success(t *testing.T) {
	fileContent := "this is a test file"
	fileName := "test.txt"

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(fileContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	uploader := &uploader{creator: mockFileCreator{}, copier: mockCopier{}}
	handler := http.HandlerFunc(uploader.uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf("File '%s' uploaded successfully (size: %d bytes)", fileName, len(fileContent))
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

// TestUploadHandler_RejectsLargeFile tests that the upload handler rejects files
// that are larger than a certain limit.
func TestUploadHandler_RejectsLargeFile(t *testing.T) {
	// Create a payload that is larger than our limit.
	largePayload := make([]byte, maxUploadSize+1)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "largefile.txt")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(largePayload)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	uploader := &uploader{creator: mockFileCreator{}, copier: mockCopier{}}
	handler := http.HandlerFunc(uploader.uploadHandler)
	handler.ServeHTTP(rr, req)

	// The test expects a 400 Bad Request because the file is too large.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

// TestUploadHandler_WrongMethod tests that the handler rejects non-POST requests.
func TestUploadHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/upload", nil)
	rr := httptest.NewRecorder()
	uploader := &uploader{creator: mockFileCreator{}, copier: mockCopier{}}
	handler := http.HandlerFunc(uploader.uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

// TestUploadHandler_NoFile tests that the handler returns an error if no file is provided.
func TestUploadHandler_NoFile(t *testing.T) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	uploader := &uploader{creator: mockFileCreator{}, copier: mockCopier{}}
	handler := http.HandlerFunc(uploader.uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

// TestUploadHandler_CreateTempError tests that the handler returns an error if it fails to create a temporary file.
func TestUploadHandler_CreateTempError(t *testing.T) {
	fileContent := "this is a test file"
	fileName := "test.txt"

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(fileContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	uploader := &uploader{creator: mockFileCreator{shouldError: true}, copier: mockCopier{}}
	handler := http.HandlerFunc(uploader.uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

// TestUploadHandler_CopyError tests that the handler returns an error if it fails to copy the file.
func TestUploadHandler_CopyError(t *testing.T) {
	fileContent := "this is a test file"
	fileName := "test.txt"

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(fileContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	uploader := &uploader{creator: mockFileCreator{}, copier: mockCopier{shouldError: true}}
	handler := http.HandlerFunc(uploader.uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}
