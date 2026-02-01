package validator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
)

func TestAPIValidator_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v4/ci/lint" {
			t.Errorf("Expected path /api/v4/ci/lint, got %s", r.URL.Path)
		}

		// Check for token header
		token := r.Header.Get("JOB-TOKEN")
		if token != "test-token" {
			t.Errorf("Expected token 'test-token', got '%s'", token)
		}

		// Send success response
		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	// Test validation
	yamlContent := []byte(`
image: alpine:latest

stages:
  - build

build:
  stage: build
  script:
    - echo "test"
`)

	result := validator.Validate(yamlContent)

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}

	if result.Stage != "api" {
		t.Errorf("Expected stage 'api', got '%s'", result.Stage)
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestAPIValidator_WithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send error response
		response := gitlab.LintResponse{
			Valid: false,
			Errors: []gitlab.APIError{
				{
					Message: "jobs build config should be a hash or an array of hashes",
				},
			},
			Warnings: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("invalid: yaml"))

	if result.Valid {
		t.Error("Expected invalid result")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Message != "jobs build config should be a hash or an array of hashes" {
		t.Errorf("Unexpected error message: %s", result.Errors[0].Message)
	}
}

func TestAPIValidator_WithWarnings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{"Job 'build' uses deprecated 'only' syntax"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if !result.Valid {
		t.Error("Expected valid result despite warnings")
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if result.Warnings[0].Message != "Job 'build' uses deprecated 'only' syntax" {
		t.Errorf("Unexpected warning message: %s", result.Warnings[0].Message)
	}
}

func TestAPIValidator_ProjectSpecific(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify project-specific endpoint is used
		expectedPath := "/api/v4/projects/123/ci/lint"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "123")

	result := validator.Validate([]byte("test: yaml"))

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestAPIValidator_ProjectByPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify project path is URL-encoded
		expectedPath := "/api/v4/projects/group%2Fproject/ci/lint"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "group/project")

	result := validator.Validate([]byte("test: yaml"))

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestAPIValidator_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		response := map[string]string{
			"message": "401 Unauthorized",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "invalid-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Valid {
		t.Error("Expected invalid result due to HTTP error")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error message")
	}

	// Error should contain "API validation failed"
	if len(result.Errors) > 0 && result.Errors[0].Message == "" {
		t.Error("Expected error message to be populated")
	}
}

func TestAPIValidator_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than client timeout
		time.Sleep(100 * time.Millisecond)

		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := gitlab.NewClient(server.URL, "test-token", 10*time.Millisecond)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Valid {
		t.Error("Expected invalid result due to timeout")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for timeout")
	}
}

func TestAPIValidator_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send invalid JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Valid {
		t.Error("Expected invalid result due to JSON parsing error")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for invalid JSON")
	}
}

func TestAPIValidator_NetworkError(t *testing.T) {
	// Create client with invalid URL
	client := gitlab.NewClient("http://invalid-url-that-does-not-exist.local:1234", "test-token", 1*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Valid {
		t.Error("Expected invalid result due to network error")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for network failure")
	}

	// Error should mention the connection failure
	if len(result.Errors) > 0 {
		errMsg := result.Errors[0].Message
		if errMsg == "" {
			t.Error("Expected error message to be populated")
		}
	}
}

func TestAPIValidator_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send valid response even for empty content
		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte(""))

	if !result.Valid {
		t.Errorf("Expected valid result for empty content, got errors: %v", result.Errors)
	}
}

func TestAPIValidator_MultipleErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := gitlab.LintResponse{
			Valid: false,
			Errors: []gitlab.APIError{
				{Message: "jobs build config should be a hash"},
				{Message: "jobs test config should be a hash"},
				{Message: "stages config should be an array of strings"},
			},
			Warnings: []string{},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("invalid: yaml"))

	if result.Valid {
		t.Error("Expected invalid result")
	}

	if len(result.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(result.Errors))
	}
}

func TestAPIValidator_Result_Stage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Stage != "api" {
		t.Errorf("Expected stage 'api', got '%s'", result.Stage)
	}
}

func TestNewAPIValidator(t *testing.T) {
	client := gitlab.NewClient("https://gitlab.com", "token", 30*time.Second)

	validator := NewAPIValidator(client, "")

	if validator == nil {
		t.Error("Expected validator to be created")
	}

	if validator.client != client {
		t.Error("Expected client to be set")
	}

	if validator.project != "" {
		t.Errorf("Expected empty project, got '%s'", validator.project)
	}

	validatorWithProject := NewAPIValidator(client, "123")

	if validatorWithProject.project != "123" {
		t.Errorf("Expected project '123', got '%s'", validatorWithProject.project)
	}
}

func TestAPIValidator_WithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request context
		// The client should use context.Background() internally
		response := gitlab.LintResponse{
			Valid:    true,
			Errors:   []gitlab.APIError{},
			Warnings: []string{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	// Test with context (currently uses Background context internally)
	result := validator.Validate([]byte("test: yaml"))

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestAPIValidator_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{
			"message": "Internal server error",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Valid {
		t.Error("Expected invalid result due to server error")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for server error")
	}
}

func TestAPIValidator_RateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		response := map[string]string{
			"message": "Rate limit exceeded",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := gitlab.NewClient(server.URL, "test-token", 30*time.Second)
	validator := NewAPIValidator(client, "")

	result := validator.Validate([]byte("test: yaml"))

	if result.Valid {
		t.Error("Expected invalid result due to rate limit")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for rate limit")
	}
}
