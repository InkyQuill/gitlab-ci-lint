package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client represents a GitLab API client
type Client struct {
	instance   string
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitLab API client
func NewClient(instance, token string, timeout time.Duration) *Client {
	return &Client{
		instance: instance,
		token:    token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Lint validates a CI configuration using the GitLab API
func (c *Client) Lint(ctx context.Context, content []byte, projectRef string) (*LintResponse, error) {
	// Build request URL
	endpoint, err := c.buildLintEndpoint(projectRef)
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint URL: %w", err)
	}

	// Create request body
	reqBody := LintRequest{
		Content: string(content),
		DryRun:  true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("JOB-TOKEN", c.token)
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var lintResp LintResponse
	if err := json.Unmarshal(respBody, &lintResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &lintResp, nil
}

// buildLintEndpoint builds the lint endpoint URL
func (c *Client) buildLintEndpoint(projectRef string) (string, error) {
	// Parse base URL
	baseURL, err := url.Parse(c.instance)
	if err != nil {
		return "", fmt.Errorf("invalid instance URL: %w", err)
	}

	// Build API path
	apiPath := "/api/v4/ci/lint"

	if projectRef != "" {
		// Use project-specific endpoint: /api/v4/projects/:id/ci/lint
		// URL encode the project reference (handles both IDs and paths like "group/project")
		apiPath = path.Join("/api/v4", "projects", url.PathEscape(projectRef), "ci", "lint")
	}

	// Join base URL with API path
	baseURL.Path = apiPath

	return baseURL.String(), nil
}

// ValidateToken checks if the current token is valid by making a simple API call
func (c *Client) ValidateToken(ctx context.Context) error {
	// Try to access the /user endpoint which requires authentication
	endpoint, err := url.Parse(c.instance)
	if err != nil {
		return fmt.Errorf("invalid instance URL: %w", err)
	}

	endpoint.Path = "/api/v4/user"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid or expired token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}

	return nil
}

// ExtractTokenFromNetrc attempts to find credentials for the given host in .netrc file
func ExtractTokenFromNetrc(instance string) (string, error) {
	// This is a simplified implementation
	// In a production tool, you might use a proper .netrc parser
	return "", fmt.Errorf("netrc support not yet implemented")
}

// NormalizeInstanceURL ensures the instance URL has a proper format
func NormalizeInstanceURL(instance string) string {
	instance = strings.TrimSpace(instance)

	// Add scheme if missing
	if !strings.HasPrefix(instance, "http://") && !strings.HasPrefix(instance, "https://") {
		instance = "https://" + instance
	}

	// Remove trailing slash
	instance = strings.TrimSuffix(instance, "/")

	return instance
}
