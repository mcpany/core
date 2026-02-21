package public_api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// StartMockServer starts a local test server with the given handler
func StartMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// AgifyHandler mocks the agify.io API
func AgifyHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	resp := map[string]interface{}{
		"name":  name,
		"age":   50,
		"count": 1000,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// BoredHandler mocks the boredapi.com API
func BoredHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"activity":      "Learn a new programming language",
		"type":          "education",
		"participants":  1,
		"price":         0.1,
		"link":          "",
		"key":           "5881028",
		"accessibility": 0.25,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GenericJSONHandler returns a generic success response
func GenericJSONHandler(response interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
