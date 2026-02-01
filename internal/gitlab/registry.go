package gitlab

import (
	"context"
	"fmt"
	"time"
)

// InstanceConfig represents a GitLab instance configuration (avoids import cycle with config package)
type InstanceConfig struct {
	Name     string
	URL      string
	Token    string
	Timeout  *time.Duration
}

// ClientRegistry manages multiple GitLab clients for different instances.
// It provides automatic routing of files to their corresponding GitLab instances
// based on .git/config detection.
type ClientRegistry struct {
	clients map[string]*Client               // key: normalized instance URL
	config  map[string]*InstanceConfig       // key: instance name
	timeout time.Duration
	debug   DebugLogger
}

// NewClientRegistry creates a new client registry with the given instance configurations.
//
// Parameters:
//   - instances: Slice of instance configurations from config file
//   - defaultTimeout: Default timeout for instances without specific timeout
//   - debug: Optional debug logger (can be nil)
//
// The registry initializes clients for all configured instances during creation.
// Clients are keyed by normalized instance URL for easy lookup.
func NewClientRegistry(instances []InstanceConfig, defaultTimeout time.Duration, debug DebugLogger) *ClientRegistry {
	registry := &ClientRegistry{
		clients: make(map[string]*Client),
		config:  make(map[string]*InstanceConfig),
		timeout: defaultTimeout,
		debug:   debug,
	}

	// Build config map and initialize clients
	for i := range instances {
		inst := &instances[i]
		registry.config[inst.Name] = inst

		timeout := defaultTimeout
		if inst.Timeout != nil {
			timeout = *inst.Timeout
		}

		client := NewClient(inst.URL, inst.Token, timeout)
		normalized := NormalizeInstanceURL(inst.URL)
		registry.clients[normalized] = client

		if debug != nil {
			tokenPreview := ""
			if inst.Token != "" {
				tokenPreview = fmt.Sprintf(" (glpat-***%s)", safeTokenSuffix(inst.Token, 4))
			}
			debug.Log("REGISTRY", fmt.Sprintf("initialized %s%s", normalized, tokenPreview))
		}
	}

	return registry
}

// GetAllInstances returns all configured instance names and URLs.
// Useful for listing instances in CLI output.
func (r *ClientRegistry) GetAllInstances() []InstanceInfo {
	result := make([]InstanceInfo, 0, len(r.config))
	for name, inst := range r.config {
		result = append(result, InstanceInfo{
			Name:     name,
			URL:      inst.URL,
			HasToken: inst.Token != "",
		})
	}
	return result
}

// InstanceInfo holds information about an instance for display purposes.
type InstanceInfo struct {
	Name     string
	URL      string
	HasToken bool
}

// GetClient returns a client for the specified instance URL.
// The URL is normalized before lookup.
//
// Returns an error if no client is configured for the instance.
func (r *ClientRegistry) GetClient(instanceURL string) (*Client, error) {
	normalized := NormalizeInstanceURL(instanceURL)
	client, exists := r.clients[normalized]
	if !exists {
		return nil, fmt.Errorf("no client configured for instance: %s", normalized)
	}
	return client, nil
}

// GetClientForInstanceName returns a client for the specified instance name.
//
// Returns (client, instanceName, error). If not found, returns (nil, "", error).
func (r *ClientRegistry) GetClientForInstanceName(name string) (*Client, string, error) {
	inst, exists := r.config[name]
	if !exists {
		return nil, "", fmt.Errorf("instance '%s' not found in configuration", name)
	}

	client, err := r.GetClient(inst.URL)
	return client, name, err
}

// ValidateAllTokens validates tokens for all configured instances.
// Returns a map of instance name to validation error (nil if validation succeeded).
//
// This is useful for setup wizard to test all configurations.
func (r *ClientRegistry) ValidateAllTokens(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, inst := range r.config {
		if inst.Token == "" {
			results[name] = fmt.Errorf("no token configured")
			continue
		}

		client, err := r.GetClient(inst.URL)
		if err != nil {
			results[name] = err
			continue
		}

		err = client.ValidateToken(ctx)
		results[name] = err
	}

	return results
}

// HasInstanceForURL checks if a client is configured for the given instance URL.
func (r *ClientRegistry) HasInstanceForURL(instanceURL string) bool {
	normalized := NormalizeInstanceURL(instanceURL)
	_, exists := r.clients[normalized]
	return exists
}

// safeTokenSuffix returns the last n characters of a token, or fewer if token is shorter.
// Returns empty string if n <= 0.
func safeTokenSuffix(token string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(token) <= n {
		return token
	}
	return token[len(token)-n:]
}
