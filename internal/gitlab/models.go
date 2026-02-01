package gitlab

// LintRequest represents the request body for the GitLab CI lint API
type LintRequest struct {
	Content     string `json:"content"`
	DryRun      bool   `json:"dry_run,omitempty"`
	IncludeJobs bool   `json:"include_jobs,omitempty"`
}

// LintResponse represents the response from the GitLab CI lint API
type LintResponse struct {
	Valid      bool          `json:"valid"`
	Errors     []APIError    `json:"errors,omitempty"`
	Warnings   []string      `json:"warnings,omitempty"`
	MergedYaml string        `json:"merged_yaml,omitempty"`
}

// APIError represents an error from the GitLab API
type APIError struct {
	Message string `json:"message"`
}
