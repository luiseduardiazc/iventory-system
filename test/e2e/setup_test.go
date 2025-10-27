package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	baseURL = "http://localhost:8080"
	apiV1   = baseURL + "/api/v1"
	apiKey  = "dev-key-store-001" // API Key para tests E2E
)

// TestClient encapsula el cliente HTTP para tests E2E
type TestClient struct {
	client  *http.Client
	baseURL string
}

// NewTestClient crea un nuevo cliente de test
func NewTestClient() *TestClient {
	return &TestClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: apiV1,
	}
}

// POST realiza una petición POST
func (tc *TestClient) POST(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", tc.baseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := tc.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute POST request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

// GET realiza una petición GET
func (tc *TestClient) GET(t *testing.T, path string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest("GET", tc.baseURL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := tc.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute GET request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

// PUT realiza una petición PUT
func (tc *TestClient) PUT(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("PUT", tc.baseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := tc.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute PUT request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

// DELETE realiza una petición DELETE
func (tc *TestClient) DELETE(t *testing.T, path string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest("DELETE", tc.baseURL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := tc.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute DELETE request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBody
}

// AssertStatusCode verifica el código de estado HTTP
func AssertStatusCode(t *testing.T, expected, actual int, body []byte) {
	t.Helper()
	if actual != expected {
		t.Errorf("Expected status %d, got %d. Response: %s", expected, actual, string(body))
	}
}

// AssertNoError verifica que no haya error en la respuesta JSON
func AssertNoError(t *testing.T, body []byte) {
	t.Helper()
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return // No es JSON o no tiene estructura esperada
	}
	if errMsg, ok := response["error"]; ok {
		t.Errorf("Unexpected error in response: %v", errMsg)
	}
}

// ParseJSON parsea el cuerpo de respuesta como JSON
func ParseJSON(t *testing.T, body []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, string(body))
	}
}

// WaitForServer espera a que el servidor esté disponible
func WaitForServer(t *testing.T, maxAttempts int) {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Log("✅ Server is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	t.Fatalf("Server not available after %d attempts", maxAttempts)
}

// RandomSKU genera un SKU único para tests
func RandomSKU(prefix string) string {
	return fmt.Sprintf("%s-E2E-%d", prefix, time.Now().UnixNano())
}

// RandomID genera un ID único para tests
func RandomID(prefix string) string {
	return fmt.Sprintf("%s-E2E-%d", prefix, time.Now().UnixNano())
}
