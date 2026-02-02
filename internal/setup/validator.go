package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
)

// TokenValidationResult contains the result of token validation
type TokenValidationResult struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

// InstanceInfo contains information about a GitLab instance
type InstanceInfo struct {
	Version      string `json:"version"`
	InstanceName string `json:"instance_name,omitempty"`
	URL          string `json:"url"`
}

// ValidateToken validates a GitLab personal access token against an instance
func ValidateToken(instance, token string) (*TokenValidationResult, error) {
	// Normalize instance URL
	instanceURL := gitlab.NormalizeInstanceURL(instance)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build /api/v4/user endpoint
	endpoint, err := url.Parse(instanceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid instance URL: %w", err)
	}
	endpoint.Path = "/api/v4/user"

	// Create request
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Sprintf("Failed to connect to instance: %v", err),
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		return &TokenValidationResult{
			Valid: false,
			Error: "Authentication failed: Invalid or expired token",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Sprintf("Token validation failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// Parse response to get username
	var user struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		// Token is valid but we couldn't parse the response
		return &TokenValidationResult{
			Valid:    true,
			Username: "unknown",
		}, nil
	}

	// Use name if username is empty
	displayName := user.Username
	if displayName == "" {
		displayName = user.Name
	}

	return &TokenValidationResult{
		Valid:    true,
		Username: displayName,
	}, nil
}

// DetectInstanceType attempts to detect the GitLab instance type
func DetectInstanceType(instance string) (*InstanceInfo, error) {
	instanceURL := gitlab.NormalizeInstanceURL(instance)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try to access the API root
	endpoint, err := url.Parse(instanceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid instance URL: %w", err)
	}
	endpoint.Path = "/api/v4"

	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to instance: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	info := &InstanceInfo{
		URL: instanceURL,
	}

	// Extract version from headers
	if version := resp.Header.Get("X-GitLab-Version"); version != "" {
		info.Version = version
	}

	// Try to determine instance type
	if strings.Contains(instanceURL, "gitlab.com") {
		info.InstanceName = "GitLab.com"
	} else {
		info.InstanceName = "Self-hosted GitLab"
	}

	return info, nil
}

// TestConnection performs a quick connectivity test to the GitLab instance.
// Accepts an optional token parameter for authentication on private instances.
func TestConnection(ctx context.Context, instance string, token ...string) error {
	instanceURL := gitlab.NormalizeInstanceURL(instance)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	endpoint, err := url.Parse(instanceURL)
	if err != nil {
		return fmt.Errorf("invalid instance URL: %w", err)
	}
	endpoint.Path = "/api/v4"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add token if provided
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("PRIVATE-TOKEN", token[0])
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Accept 200 OK or 401 Unauthorized (both mean instance is reachable)
	// 401 means instance exists but requires authentication
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
