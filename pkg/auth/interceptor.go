package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// APIKeyHeader is the default header name for the API key.
	APIKeyHeader = "X-API-Key"
)

var (
	// ErrMissingAPIKey is returned when the API key is missing from the request.
	ErrMissingAPIKey = errors.New("missing API key")
	// ErrInvalidAPIKey is returned when the API key is invalid.
	ErrInvalidAPIKey = errors.New("invalid API key")
)

// AuthenticationInterceptor provides a gRPC interceptor for API key authentication.
type AuthenticationInterceptor struct {
	apiKey string
}

// NewAuthenticationInterceptor creates a new authentication interceptor.
func NewAuthenticationInterceptor(apiKey string) *AuthenticationInterceptor {
	return &AuthenticationInterceptor{
		apiKey: apiKey,
	}
}

// Unary returns a gRPC unary interceptor for authentication.
func (i *AuthenticationInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := i.authenticate(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *AuthenticationInterceptor) authenticate(ctx context.Context) error {
	if i.apiKey == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get(strings.ToLower(APIKeyHeader))
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "%s", ErrMissingAPIKey.Error())
	}

	if values[0] != i.apiKey {
		return status.Errorf(codes.Unauthenticated, "%s", ErrInvalidAPIKey.Error())
	}

	return nil
}

// AuthenticateRequest authenticates an HTTP request.
func (i *AuthenticationInterceptor) AuthenticateRequest(r *http.Request) error {
	if i.apiKey == "" {
		// No API key configured, so all requests are allowed.
		return nil
	}

	key := r.Header.Get(APIKeyHeader)
	if key == "" {
		return ErrMissingAPIKey
	}

	if key != i.apiKey {
		return ErrInvalidAPIKey
	}

	return nil
}

// Wrap wraps an HTTP handler with authentication.
func (i *AuthenticationInterceptor) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := i.AuthenticateRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
