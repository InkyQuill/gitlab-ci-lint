package gitlab

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DebugLogger is an interface for debug logging (avoids import cycle)
type DebugLogger interface {
	LogAPIRequest(endpoint, method string, hasProject bool, project string)
	Log(category, message string)
}

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
func (c *Client) Lint(ctx context.Context, content []byte, projectRef string, debug DebugLogger) (*LintResponse, error) {
	// Build request URL
	endpoint, err := c.buildLintEndpoint(projectRef)
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint URL: %w", err)
	}

	// Debug logging
	if debug != nil {
		debug.LogAPIRequest(endpoint, "POST", projectRef != "", projectRef)
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

	if debug != nil {
		debug.Log("API", fmt.Sprintf("request size: %d bytes", len(jsonBody)))
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	// GitLab supports two token types:
	// - PRIVATE-TOKEN: Personal Access Token (PAT) - used by default for linter
	// - JOB-TOKEN: CI/CD Job Token (limited to current project)
	//
	// We set PRIVATE-TOKEN as the primary authentication method for full API access.
	// JOB-TOKEN is also set for compatibility with GitLab servers that may require it.
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
		req.Header.Set("JOB-TOKEN", c.token) // For compatibility
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if debug != nil {
			debug.Log("API", fmt.Sprintf("request failed: %v", err))
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if debug != nil {
		debug.Log("API", fmt.Sprintf("response status: %d", resp.StatusCode))
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if debug != nil {
		debug.Log("API", fmt.Sprintf("response size: %d bytes", len(respBody)))
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

	if debug != nil {
		if lintResp.Valid {
			debug.Log("API", "validation result: VALID")
		} else {
			debug.Log("API", fmt.Sprintf("validation result: INVALID (%d errors)", len(lintResp.Errors)))
		}
	}

	return &lintResp, nil
}

// buildLintEndpoint builds the lint endpoint URL
func (c *Client) buildLintEndpoint(projectRef string) (string, error) {
	// Parse base URL to extract scheme and host
	baseURL, err := url.Parse(c.instance)
	if err != nil {
		return "", fmt.Errorf("invalid instance URL: %w", err)
	}

	if projectRef == "" {
		// Global lint endpoint: /api/v4/ci/lint
		return fmt.Sprintf("%s://%s/api/v4/ci/lint", baseURL.Scheme, baseURL.Host), nil
	}

	// Project-specific endpoint: /api/v4/projects/:id/ci/lint
	// Manually encode and build to avoid double-encoding
	encodedProject := url.PathEscape(projectRef)
	return fmt.Sprintf("%s://%s/api/v4/projects/%s/ci/lint",
		baseURL.Scheme, baseURL.Host, encodedProject), nil
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
//
// The .netrc file format:
//   machine <host> login <username> password <token>
//   machine gitlab.com login gitlab-ci-lint password glpat-...
//
// This function searches for a matching machine entry and returns the password field.
// Returns empty string with no error if .netrc doesn't exist or has no matching entry.
// Returns error only if .netrc exists but cannot be read.
func ExtractTokenFromNetrc(instance string) (string, error) {
	// Parse instance URL to get hostname
	hostname := extractHostname(instance)
	if hostname == "" {
		return "", fmt.Errorf("invalid instance URL: %s", instance)
	}

	// Find .netrc file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}

	netrcPath := filepath.Join(homeDir, ".netrc")
	file, err := os.Open(netrcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// .netrc doesn't exist, return empty (not an error)
			return "", nil
		}
		return "", fmt.Errorf("cannot open .netrc: %w", err)
	}
	defer func() {
		// Close file; error is ignored as we already have the data we need
		_ = file.Close()
	}()

	// Parse netrc format: machine <host> login <user> password <secret>
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Parse line
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Check for machine entry
		if fields[0] == "machine" && fields[1] == hostname {
			// Look for password field (index 2 is "login", 3 is username, 4 is "password", 5 is token)
			for i := 2; i < len(fields)-1; i++ {
				if fields[i] == "password" {
					return fields[i+1], nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read .netrc: %w", err)
	}

	// No matching entry found
	return "", nil
}

// extractHostname extracts the hostname from a URL
func extractHostname(instance string) string {
	// Remove protocol
	instance = strings.TrimPrefix(instance, "https://")
	instance = strings.TrimPrefix(instance, "http://")
	instance = strings.TrimPrefix(instance, "ssh://")
	instance = strings.TrimPrefix(instance, "git@")

	// Remove port and path
	if idx := strings.Index(instance, ":"); idx != -1 {
		instance = instance[:idx]
	}
	if idx := strings.Index(instance, "/"); idx != -1 {
		instance = instance[:idx]
	}

	return instance
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
