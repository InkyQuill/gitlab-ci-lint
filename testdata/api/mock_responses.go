package api

import "github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"

// MockResponse represents a mock API response for testing
type MockResponse struct {
	StatusCode int
	Body       string
	Error      error
}

// MockServer represents a mock GitLab API server for testing
type MockServer struct {
	responses map[string]MockResponse
}

// NewMockServer creates a new mock API server
func NewMockServer() *MockServer {
	return &MockServer{
		responses: make(map[string]MockResponse),
	}
}

// SetResponse sets a mock response for a specific endpoint
func (m *MockServer) SetResponse(endpoint string, response MockResponse) {
	m.responses[endpoint] = response
}

// GetLintValidResponse returns a valid lint response
func GetLintValidResponse() gitlab.LintResponse {
	return gitlab.LintResponse{
		Valid:      true,
		Errors:     []gitlab.APIError{},
		Warnings:   []string{},
		MergedYaml: "stages:\n  - build\n  - test\n",
	}
}

// GetLintInvalidResponse returns an invalid lint response
func GetLintInvalidResponse() gitlab.LintResponse {
	return gitlab.LintResponse{
		Valid:    false,
		Errors: []gitlab.APIError{
			{Message: "jobs:test:script config should be an array of strings"},
			{Message: "jobs:build stage is not defined in stages"},
		},
		Warnings: []string{
			"Variable FOO is not used",
		},
	}
}

// GetLintErrorResponse returns a lint response with errors
func GetLintErrorResponse() gitlab.LintResponse {
	return gitlab.LintResponse{
		Valid: false,
		Errors: []gitlab.APIError{
			{Message: "Invalid YAML syntax"},
			{Message: "jobs:deploy:script config should be an array of strings"},
		},
	}
}

// GetUserValidResponse returns a valid user response
func GetUserValidResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":       1,
		"username": "testuser",
		"name":     "Test User",
		"email":    "test@example.com",
	}
}
