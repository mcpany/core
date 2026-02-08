package middleware

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

// HTTPRequestContextKey is the context key for the HTTP request.
const HTTPRequestContextKey contextKey = "http.request"
