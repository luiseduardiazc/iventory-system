package e2e

import (
	"net/http"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	// Para health check usamos directamente la URL base sin /api/v1
	tc := &TestClient{
		client: &http.Client{
			Timeout: 10 * http.DefaultClient.Timeout,
		},
		baseURL: baseURL,
	}

	resp, body := tc.GET(t, "/health")

	AssertStatusCode(t, http.StatusOK, resp.StatusCode, body)
	AssertNoError(t, body)

	var response map[string]interface{}
	ParseJSON(t, body, &response)

	// Verificar campos esperados
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	if _, ok := response["timestamp"]; !ok {
		t.Error("Missing 'timestamp' field")
	}

	if _, ok := response["version"]; !ok {
		t.Error("Missing 'version' field")
	}

	if dbStatus, ok := response["database"].(string); !ok || dbStatus != "healthy" {
		t.Errorf("Expected database status 'healthy', got %v", response["database"])
	}

	t.Logf("✅ Health check passed: %s", string(body))
}
