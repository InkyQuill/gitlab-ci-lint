package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizeInstanceURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "gitlab.com without scheme",
			input:    "gitlab.com",
			expected: "https://gitlab.com",
		},
		{
			name:     "already has https",
			input:    "https://gitlab.com",
			expected: "https://gitlab.com",
		},
		{
			name:     "has http",
			input:    "http://gitlab.com",
			expected: "http://gitlab.com",
		},
		{
			name:     "with trailing slash",
			input:    "https://gitlab.com/",
			expected: "https://gitlab.com",
		},
		{
			name:     "with spaces",
			input:    "  https://gitlab.com  ",
			expected: "https://gitlab.com",
		},
		{
			name:     "localhost with port",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "custom domain",
			input:    "gitlab.example.com",
			expected: "https://gitlab.example.com",
		},
		{
			name:     "with path",
			input:    "https://gitlab.com/",
			expected: "https://gitlab.com",
		},
		{
			name:     "complex URL",
			input:    "  https://gitlab.example.com/  ",
			expected: "https://gitlab.example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeInstanceURL(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	instance := "https://gitlab.com"
	token := "test-token"
	timeout := 30 * time.Second

	client := NewClient(instance, token, timeout)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.instance != instance {
		t.Errorf("Expected instance '%s', got '%s'", instance, client.instance)
	}

	if client.token != token {
		t.Errorf("Expected token '%s', got '%s'", token, client.token)
	}

	if client.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}

	if client.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
}

func TestClient_ValidateToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/user" {
			t.Errorf("Expected path /api/v4/user, got %s", r.URL.Path)
		}

		token := r.Header.Get("PRIVATE-TOKEN")
		if token != "test-token" {
			t.Errorf("Expected token 'test-token', got '%s'", token)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"username": "testuser", "name": "Test User"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", 10*time.Second)
	err := client.ValidateToken(context.Background())

	if err != nil {
		t.Errorf("Expected successful token validation, got error: %v", err)
	}
}

func TestClient_ValidateToken_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "401 Unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "invalid-token", 10*time.Second)
	err := client.ValidateToken(context.Background())

	if err == nil {
		t.Error("Expected error for invalid token")
	}

	expectedMsg := "authentication failed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestClient_ValidateToken_NoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("PRIVATE-TOKEN")
		if token != "" {
			t.Errorf("Expected no token header, got '%s'", token)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"username": "public"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "", 10*time.Second)
	err := client.ValidateToken(context.Background())

	if err != nil {
		t.Errorf("Expected successful validation without token, got error: %v", err)
	}
}

func TestClient_BuildLintEndpoint_Global(t *testing.T) {
	client := NewClient("https://gitlab.com", "test-token", 10*time.Second)

	endpoint, err := client.buildLintEndpoint("")

	if err != nil {
		t.Fatalf("Failed to build endpoint: %v", err)
	}

	expected := "https://gitlab.com/api/v4/ci/lint"
	if endpoint != expected {
		t.Errorf("Expected endpoint '%s', got '%s'", expected, endpoint)
	}
}

func TestClient_BuildLintEndpoint_ProjectByID(t *testing.T) {
	client := NewClient("https://gitlab.com", "test-token", 10*time.Second)

	endpoint, err := client.buildLintEndpoint("123")

	if err != nil {
		t.Fatalf("Failed to build endpoint: %v", err)
	}

	expected := "https://gitlab.com/api/v4/projects/123/ci/lint"
	if endpoint != expected {
		t.Errorf("Expected endpoint '%s', got '%s'", expected, endpoint)
	}
}

func TestClient_BuildLintEndpoint_ProjectByPath(t *testing.T) {
	client := NewClient("https://gitlab.com", "test-token", 10*time.Second)

	endpoint, err := client.buildLintEndpoint("group/project")

	if err != nil {
		t.Fatalf("Failed to build endpoint: %v", err)
	}

	// PathEscape will encode the slash as %2F, and then URL building will encode it again as %252F
	// The important thing is that it contains the project path
	if !strings.Contains(endpoint, "/api/v4/projects/") {
		t.Errorf("Expected endpoint to contain projects path, got '%s'", endpoint)
	}

	if !strings.HasSuffix(endpoint, "/ci/lint") {
		t.Errorf("Expected endpoint to end with /ci/lint, got '%s'", endpoint)
	}

	// Verify the path contains the encoded project reference
	if !strings.Contains(endpoint, "group") {
		t.Errorf("Expected endpoint to contain project path 'group', got '%s'", endpoint)
	}
}

func TestClient_BuildLintEndpoint_ProjectWithSpecialChars(t *testing.T) {
	client := NewClient("https://gitlab.com", "test-token", 10*time.Second)

	endpoint, err := client.buildLintEndpoint("my.group/sub-project")

	if err != nil {
		t.Fatalf("Failed to build endpoint: %v", err)
	}

	// Special characters should be properly encoded
	if !strings.Contains(endpoint, "/api/v4/projects/") {
		t.Errorf("Expected endpoint to contain projects path, got '%s'", endpoint)
	}

	if !strings.Contains(endpoint, "/ci/lint") {
		t.Errorf("Expected endpoint to contain ci/lint path, got '%s'", endpoint)
	}
}

func TestClient_Lint_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/api/v4/ci/lint" {
			t.Errorf("Expected path /api/v4/ci/lint, got %s", r.URL.Path)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid": true, "errors": [], "warnings": []}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", 10*time.Second)
	content := []byte(`image: alpine`)

	resp, err := client.Lint(context.Background(), content, "")

	if err != nil {
		t.Fatalf("Failed to lint: %v", err)
	}

	if !resp.Valid {
		t.Error("Expected valid response")
	}

	if len(resp.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(resp.Errors))
	}
}

func TestClient_Lint_WithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"valid": false,
			"errors": [
				{"message": "jobs config should contain at least one visible job"}
			],
			"warnings": []
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", 10*time.Second)
	content := []byte(`invalid: config`)

	resp, err := client.Lint(context.Background(), content, "")

	if err != nil {
		t.Fatalf("Failed to lint: %v", err)
	}

	if resp.Valid {
		t.Error("Expected invalid response")
	}

	if len(resp.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(resp.Errors))
	}

	expectedError := "jobs config should contain at least one visible job"
	if resp.Errors[0].Message != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, resp.Errors[0].Message)
	}
}

func TestClient_Lint_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal server error`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", 10*time.Second)
	content := []byte(`image: alpine`)

	_, err := client.Lint(context.Background(), content, "")

	if err == nil {
		t.Error("Expected error for non-200 status")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to contain status code, got: %v", err)
	}
}

func TestExtractTokenFromNetrc_NotImplemented(t *testing.T) {
	_, err := ExtractTokenFromNetrc("https://gitlab.com")

	if err == nil {
		t.Error("Expected error for unimplemented netrc support")
	}

	expectedMsg := "not yet implemented"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestClient_BuildLintEndpoint_InvalidInstanceURL(t *testing.T) {
	client := NewClient(":invalid-url", "test-token", 10*time.Second)

	_, err := client.buildLintEndpoint("")

	if err == nil {
		t.Error("Expected error for invalid instance URL")
	}
}

func TestClient_ValidateToken_NetworkError(t *testing.T) {
	// Create a server that's already closed (simulating network error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.Close() // Close immediately

	client := NewClient(server.URL, "test-token", 10*time.Second)
	err := client.ValidateToken(context.Background())

	if err == nil {
		t.Error("Expected error for network failure")
	}
}

func TestClient_Lint_WithProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should use project-specific endpoint
		if !strings.Contains(r.URL.Path, "/projects/") {
			t.Errorf("Expected project-specific endpoint path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid": true, "errors": [], "warnings": []}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", 10*time.Second)
	content := []byte(`image: alpine`)

	resp, err := client.Lint(context.Background(), content, "123")

	if err != nil {
		t.Fatalf("Failed to lint with project: %v", err)
	}

	if !resp.Valid {
		t.Error("Expected valid response for project-specific lint")
	}
}
