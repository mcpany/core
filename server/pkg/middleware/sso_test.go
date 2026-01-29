package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSSOMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := SSOConfig{
		Enabled: true,
		IDPURL:  "https://idp.example.com",
	}

	r := gin.New()
	r.Use(SSOMiddleware(config))

	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test No Auth
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test ID Header
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-MCP-Identity", "alice")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test Bearer Token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-mock-token")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
