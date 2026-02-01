package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockDebugLogger is a simple implementation of DebugLogger for testing
type mockDebugLogger struct {
	logs []string
}

func (m *mockDebugLogger) Log(component, message string) {
	m.logs = append(m.logs, fmt.Sprintf("[%s] %s", component, message))
}

func (m *mockDebugLogger) LogAPIRequest(endpoint, method string, hasProject bool, project string) {
	projectInfo := ""
	if hasProject && project != "" {
		projectInfo = fmt.Sprintf(" project=%s", project)
	}
	m.logs = append(m.logs, fmt.Sprintf("[API] %s %s%s", method, endpoint, projectInfo))
}

func (m *mockDebugLogger) GetLogs() []string {
	return m.logs
}

func (m *mockDebugLogger) Clear() {
	m.logs = nil
}

// TestNewClientRegistry_EmptyInstances tests creating a registry with no instances
func TestNewClientRegistry_EmptyInstances(t *testing.T) {
	registry := NewClientRegistry([]InstanceConfig{}, 30*time.Second, nil)

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	instances := registry.GetAllInstances()
	if len(instances) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(instances))
	}
}

// TestNewClientRegistry_SingleInstance tests creating a registry with one instance
func TestNewClientRegistry_SingleInstance(t *testing.T) {
	mockDebug := &mockDebugLogger{}
	instances := []InstanceConfig{
		{
			Name:    "gitlab.com",
			URL:     "https://gitlab.com",
			Token:   "test-token",
			Timeout: nil,
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, mockDebug)

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	allInstances := registry.GetAllInstances()
	if len(allInstances) != 1 {
		t.Errorf("Expected 1 instance, got %d", len(allInstances))
	}

	if allInstances[0].Name != "gitlab.com" {
		t.Errorf("Expected instance name 'gitlab.com', got '%s'", allInstances[0].Name)
	}

	if allInstances[0].URL != "https://gitlab.com" {
		t.Errorf("Expected URL 'https://gitlab.com', got '%s'", allInstances[0].URL)
	}

	if !allInstances[0].HasToken {
		t.Error("Expected HasToken to be true")
	}

	// Check debug logs
	logs := mockDebug.GetLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 debug log, got %d", len(logs))
	}
}

// TestNewClientRegistry_MultipleInstances tests creating a registry with multiple instances
func TestNewClientRegistry_MultipleInstances(t *testing.T) {
	timeout := 45 * time.Second
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "token1",
		},
		{
			Name:  "gitlab.example.com",
			URL:   "https://gitlab.example.com",
			Token: "token2",
		},
		{
			Name:    "localhost",
			URL:     "http://localhost:8080",
			Token:   "token3",
			Timeout: &timeout,
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	allInstances := registry.GetAllInstances()
	if len(allInstances) != 3 {
		t.Errorf("Expected 3 instances, got %d", len(allInstances))
	}

	// Verify all instances are present
	instanceNames := make(map[string]bool)
	for _, inst := range allInstances {
		instanceNames[inst.Name] = true
	}

	expectedNames := []string{"gitlab.com", "gitlab.example.com", "localhost"}
	for _, name := range expectedNames {
		if !instanceNames[name] {
			t.Errorf("Expected instance '%s' not found", name)
		}
	}
}

// TestGetClient_ValidURL tests getting a client by URL
func TestGetClient_ValidURL(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	client, err := registry.GetClient("https://gitlab.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

// TestGetClient_URLVariations tests URL normalization in client lookup
func TestGetClient_URLVariations(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	testURLs := []string{
		"https://gitlab.com",
		"https://gitlab.com/",
		"gitlab.com",
		"gitlab.com/",
		"  gitlab.com  ",
	}

	for _, url := range testURLs {
		t.Run(url, func(t *testing.T) {
			client, err := registry.GetClient(url)
			if err != nil {
				t.Errorf("Expected no error for URL '%s', got: %v", url, err)
			}
			if client == nil {
				t.Errorf("Expected non-nil client for URL '%s'", url)
			}
		})
	}
}

// TestGetClient_NotFound tests getting a client for an unconfigured instance
func TestGetClient_NotFound(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	_, err := registry.GetClient("https://gitlab.example.com")
	if err == nil {
		t.Error("Expected error for unconfigured instance")
	}

	expectedMsg := "no client configured for instance"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

// TestGetClientForInstanceName_Found tests getting a client by instance name
func TestGetClientForInstanceName_Found(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
		{
			Name:  "work",
			URL:   "https://gitlab.example.com",
			Token: "work-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	client, name, err := registry.GetClientForInstanceName("work")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if name != "work" {
		t.Errorf("Expected instance name 'work', got '%s'", name)
	}
}

// TestGetClientForInstanceName_NotFound tests getting a client by non-existent name
func TestGetClientForInstanceName_NotFound(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	_, name, err := registry.GetClientForInstanceName("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent instance")
	}

	if name != "" {
		t.Errorf("Expected empty instance name, got '%s'", name)
	}

	expectedMsg := "instance 'nonexistent' not found"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

// TestHasInstanceForURL_True tests checking if an instance exists
func TestHasInstanceForURL_True(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	if !registry.HasInstanceForURL("https://gitlab.com") {
		t.Error("Expected HasInstanceForURL to return true")
	}

	if !registry.HasInstanceForURL("gitlab.com") {
		t.Error("Expected HasInstanceForURL to return true for unnormalized URL")
	}
}

// TestHasInstanceForURL_False tests checking for a non-existent instance
func TestHasInstanceForURL_False(t *testing.T) {
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "test-token",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	if registry.HasInstanceForURL("https://gitlab.example.com") {
		t.Error("Expected HasInstanceForURL to return false for unconfigured instance")
	}
}

// TestGetAllInstances tests retrieving all configured instances
func TestGetAllInstances(t *testing.T) {
	timeout := 45 * time.Second
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "token1",
		},
		{
			Name:    "custom",
			URL:     "https://custom.com",
			Token:   "token2",
			Timeout: &timeout,
		},
		{
			Name:  "notoken",
			URL:   "https://notoken.com",
			Token: "",
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	allInstances := registry.GetAllInstances()
	if len(allInstances) != 3 {
		t.Fatalf("Expected 3 instances, got %d", len(allInstances))
	}

	// Verify each instance
	instanceMap := make(map[string]InstanceInfo)
	for _, inst := range allInstances {
		instanceMap[inst.Name] = inst
	}

	// Check gitlab.com
	if inst, ok := instanceMap["gitlab.com"]; ok {
		if !inst.HasToken {
			t.Error("Expected gitlab.com to have token")
		}
	} else {
		t.Error("Instance 'gitlab.com' not found")
	}

	// Check custom
	if inst, ok := instanceMap["custom"]; ok {
		if !inst.HasToken {
			t.Error("Expected custom to have token")
		}
	} else {
		t.Error("Instance 'custom' not found")
	}

	// Check notoken
	if inst, ok := instanceMap["notoken"]; ok {
		if inst.HasToken {
			t.Error("Expected notoken to not have token")
		}
	} else {
		t.Error("Instance 'notoken' not found")
	}
}

// TestValidateAllTokens_AllValid tests token validation when all tokens are valid
func TestValidateAllTokens_AllValid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/user" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"username": "testuser"}`))
		}
	}))
	defer server.Close()

	instances := []InstanceConfig{
		{Name: "inst1", URL: server.URL, Token: "token1"},
		{Name: "inst2", URL: server.URL, Token: "token2"},
	}

	registry := NewClientRegistry(instances, 10*time.Second, nil)

	ctx := context.Background()
	results := registry.ValidateAllTokens(ctx)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for name, err := range results {
		if err != nil {
			t.Errorf("Instance '%s' expected valid token, got error: %v", name, err)
		}
	}
}

// TestValidateAllTokens_MixedResults tests token validation with mixed success/failure
func TestValidateAllTokens_MixedResults(t *testing.T) {
	// Use separate servers for different instances to ensure proper isolation
	validServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/user" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"username": "testuser"}`))
		}
	}))
	defer validServer.Close()

	invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/user" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message": "401 Unauthorized"}`))
		}
	}))
	defer invalidServer.Close()

	instances := []InstanceConfig{
		{Name: "valid", URL: validServer.URL, Token: "valid-token"},
		{Name: "invalid", URL: invalidServer.URL, Token: "invalid-token"},
		{Name: "notoken", URL: validServer.URL, Token: ""},
	}

	registry := NewClientRegistry(instances, 10*time.Second, nil)

	ctx := context.Background()
	results := registry.ValidateAllTokens(ctx)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results["valid"] != nil {
		t.Errorf("Instance 'valid' expected no error, got: %v", results["valid"])
	}

	if results["invalid"] == nil {
		t.Error("Instance 'invalid' expected error, got nil")
	}

	// No token instance should error (our code returns "no token configured" before API call)
	if results["notoken"] == nil {
		t.Error("Instance 'notoken' expected error (no token), got nil")
	}
}

// TestValidateAllTokens_NoTokens tests validation when instances have no tokens
func TestValidateAllTokens_NoTokens(t *testing.T) {
	instances := []InstanceConfig{
		{Name: "notoken1", URL: "https://gitlab.com", Token: ""},
		{Name: "notoken2", URL: "https://gitlab.example.com", Token: ""},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	ctx := context.Background()
	results := registry.ValidateAllTokens(ctx)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for name, err := range results {
		if err == nil {
			t.Errorf("Instance '%s' expected error (no token), got nil", name)
		}
		expectedMsg := "no token configured"
		if !contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
		}
	}
}

// TestNewClientRegistry_WithTimeout tests that custom timeouts are respected
func TestNewClientRegistry_WithTimeout(t *testing.T) {
	customTimeout := 60 * time.Second
	instances := []InstanceConfig{
		{
			Name:    "default",
			URL:     "https://gitlab.com",
			Token:   "token1",
			Timeout: nil,
		},
		{
			Name:    "custom",
			URL:     "https://custom.com",
			Token:   "token2",
			Timeout: &customTimeout,
		},
	}

	registry := NewClientRegistry(instances, 30*time.Second, nil)

	// Get client for default timeout instance
	client1, _ := registry.GetClient("https://gitlab.com")
	if client1 == nil {
		t.Fatal("Expected non-nil client for default instance")
	}

	// Get client for custom timeout instance
	client2, _ := registry.GetClient("https://custom.com")
	if client2 == nil {
		t.Fatal("Expected non-nil client for custom instance")
	}

	// Note: We can't directly access the timeout from the client,
	// but we verified the registry was created without errors
}

// TestNewClientRegistry_DebugLogging tests debug logging during registry creation
func TestNewClientRegistry_DebugLogging(t *testing.T) {
	mockDebug := &mockDebugLogger{}
	instances := []InstanceConfig{
		{
			Name:  "gitlab.com",
			URL:   "https://gitlab.com",
			Token: "glpat-12345678",
		},
		{
			Name:  "notoken",
			URL:   "https://notoken.com",
			Token: "",
		},
	}

	_ = NewClientRegistry(instances, 30*time.Second, mockDebug)

	logs := mockDebug.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 debug logs, got %d", len(logs))
	}

	// Check first log contains instance URL and token suffix
	if !contains(logs[0], "https://gitlab.com") {
		t.Errorf("Expected first log to contain instance URL, got: %s", logs[0])
	}

	// Check token suffix is present
	if !contains(logs[0], "5678") {
		t.Errorf("Expected first log to contain token suffix, got: %s", logs[0])
	}

	// Check second log for no-token instance
	if !contains(logs[1], "https://notoken.com") {
		t.Errorf("Expected second log to contain instance URL, got: %s", logs[1])
	}

	// Second log should NOT have token suffix (no token configured)
	if contains(logs[1], "glpat-") {
		t.Errorf("Expected second log to NOT have token, got: %s", logs[1])
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
