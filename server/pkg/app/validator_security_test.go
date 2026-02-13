package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleValidate_SecretLeak(t *testing.T) {
	// Setup: Set a secret environment variable
	secretKey := "TEST_SECRET_LEAK"
	secretVal := "SuperSecretValue123"
	t.Setenv(secretKey, secretVal)

	app := &Application{}

	// Case 1: Regex matches the secret (Oracle: True)
	// If the system validates secrets, this should pass.
	configMatch := `
users:
  - id: attacker
    authentication:
      api_key:
        param_name: X-Key
        value:
          environmentVariable: "` + secretKey + `"
          validation_regex: "^Super.*"
`
	reqMatch := ValidateRequest{
		Content: configMatch,
		Format:  "yaml",
	}
	bodyMatch, _ := json.Marshal(reqMatch)
	wMatch := httptest.NewRecorder()
	rMatch := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(bodyMatch))
	app.handleValidate().ServeHTTP(wMatch, rMatch)

	// Case 2: Regex does NOT match the secret (Oracle: False)
	// If the system validates secrets, this should FAIL.
	// If we successfully patch it, this should PASS (because validation is skipped).
	configMismatch := `
users:
  - id: attacker
    authentication:
      api_key:
        param_name: X-Key
        value:
          environmentVariable: "` + secretKey + `"
          validation_regex: "^Wrong.*"
`
	reqMismatch := ValidateRequest{
		Content: configMismatch,
		Format:  "yaml",
	}
	bodyMismatch, _ := json.Marshal(reqMismatch)
	wMismatch := httptest.NewRecorder()
	rMismatch := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(bodyMismatch))
	app.handleValidate().ServeHTTP(wMismatch, rMismatch)

	// Analysis
	var respMatch, respMismatch ValidateResponse
	err := json.Unmarshal(wMatch.Body.Bytes(), &respMatch)
	require.NoError(t, err)
	err = json.Unmarshal(wMismatch.Body.Bytes(), &respMismatch)
	require.NoError(t, err)

	t.Logf("Match Response: valid=%v, error=%s", respMatch.Valid, respMatch.Error)
	t.Logf("Mismatch Response: valid=%v, error=%s", respMismatch.Valid, respMismatch.Error)

	// VULNERABILITY CHECK:
	// If Match is Valid and Mismatch is Invalid, then we have an oracle.
	if respMatch.Valid && !respMismatch.Valid {
		t.Log("🚨 VULNERABILITY CONFIRMED: Secret leak via regex oracle detected!")
		t.Fail() // Fail the test to indicate vulnerability exists
	} else {
		t.Log("✅ Secure: No oracle detected (validation skipped or consistent).")
	}
}

func TestHandleValidate_FileExistenceLeak(t *testing.T) {
	app := &Application{}

	// Case 1: File Exists
	tmpFile, err := os.CreateTemp("", "exist*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	configExist := `
users:
  - id: attacker
    authentication:
      api_key:
        param_name: X-Key
        value:
          filePath: "` + tmpFile.Name() + `"
`
	reqExist := ValidateRequest{
		Content: configExist,
		Format:  "yaml",
	}
	bodyExist, _ := json.Marshal(reqExist)
	wExist := httptest.NewRecorder()
	rExist := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(bodyExist))
	app.handleValidate().ServeHTTP(wExist, rExist)

	// Case 2: File Does Not Exist
	configNotExist := `
users:
  - id: attacker
    authentication:
      api_key:
        param_name: X-Key
        value:
          filePath: "` + tmpFile.Name() + `.missing"
`
	reqNotExist := ValidateRequest{
		Content: configNotExist,
		Format:  "yaml",
	}
	bodyNotExist, _ := json.Marshal(reqNotExist)
	wNotExist := httptest.NewRecorder()
	rNotExist := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(bodyNotExist))
	app.handleValidate().ServeHTTP(wNotExist, rNotExist)

	// Analysis
	var respExist, respNotExist ValidateResponse
	err = json.Unmarshal(wExist.Body.Bytes(), &respExist)
	require.NoError(t, err)
	err = json.Unmarshal(wNotExist.Body.Bytes(), &respNotExist)
	require.NoError(t, err)

	t.Logf("Exist Response: valid=%v, error=%s", respExist.Valid, respExist.Error)
	t.Logf("NotExist Response: valid=%v, error=%s", respNotExist.Valid, respNotExist.Error)

	// VULNERABILITY CHECK:
	// If Exist is Valid and NotExist is Invalid, then we have an oracle.
	if respExist.Valid && !respNotExist.Valid {
		t.Log("🚨 VULNERABILITY CONFIRMED: File existence leak detected!")
		t.Fail()
	} else {
		t.Log("✅ Secure: No file existence oracle detected.")
	}
}
