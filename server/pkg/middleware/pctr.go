package middleware

import (
    "fmt"
    "net/http"
    "strings"
)

type PCTRMiddleware struct {
    // any needed state
}

func NewPCTRMiddleware() *PCTRMiddleware {
    return &PCTRMiddleware{}
}

func (m *PCTRMiddleware) APIHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handle rotation requests
    }
}
