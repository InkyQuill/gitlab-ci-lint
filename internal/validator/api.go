package validator

import (
	"context"
	"fmt"

	"github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
)

// APIValidator validates CI configuration using GitLab API
type APIValidator struct {
	client  *gitlab.Client
	project string
	debug   gitlab.DebugLogger
}

// NewAPIValidator creates a new API validator
func NewAPIValidator(client *gitlab.Client, project string) *APIValidator {
	return &APIValidator{
		client:  client,
		project: project,
		debug:   nil,
	}
}

// SetDebugLogger sets the debug logger for the validator
func (v *APIValidator) SetDebugLogger(debug gitlab.DebugLogger) {
	v.debug = debug
}

// Validate validates the CI configuration using the GitLab API
func (v *APIValidator) Validate(content []byte) Result {
	result := Result{
		Stage: "api",
	}

	// Create context with timeout from client
	ctx := context.Background()

	// Call GitLab API
	resp, err := v.client.Lint(ctx, content, v.project, v.debug)
	if err != nil {
		result.Valid = false
		result.Errors = []Error{
			{
				Message: fmt.Sprintf("API validation failed: %v", err),
			},
		}
		return result
	}

	// Parse response
	result.Valid = resp.Valid

	// Convert API errors to validator errors
	for _, apiErr := range resp.Errors {
		result.Errors = append(result.Errors, Error{
			Message: apiErr.Message,
		})
	}

	// Convert warnings
	for _, warning := range resp.Warnings {
		result.Warnings = append(result.Warnings, Warning{
			Message: warning,
		})
	}

	return result
}
