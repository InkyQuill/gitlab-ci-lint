package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/InkyQuill/gitlab-ci-lint/internal/validator"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter("auto", false)

	if formatter == nil { //nolint:staticcheck
		t.Fatal("Expected non-nil formatter")
	}

	if formatter.colorizer == nil { //nolint:staticcheck
		t.Error("Expected non-nil colorizer")
	}

	if formatter.verbose != false { //nolint:staticcheck
		t.Errorf("Expected verbose false, got true")
	}
}

func TestFormatter_FormatResult_Text_Success(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage:  "local",
			Valid:  true,
			Errors: []validator.Error{},
		},
		{
			Stage:  "api",
			Valid:  true,
			Errors: []validator.Error{},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "text", results, ".gitlab-ci.yml")

	output := buf.String()

	// Should contain success indicators
	if !strings.Contains(output, "✓") {
		t.Error("Expected output to contain success indicator")
	}

	// Should mention the file
	if !strings.Contains(output, ".gitlab-ci.yml") {
		t.Error("Expected output to contain filename")
	}
}

func TestFormatter_FormatResult_Text_WithErrors(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage: "local",
			Valid: false,
			Errors: []validator.Error{
				{
					Message: "invalid YAML syntax",
					Line:    10,
					Column:  5,
					Content: "  script: echo \"test",
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "text", results, ".gitlab-ci.yml")

	output := buf.String()

	// Should contain error indicator
	if !strings.Contains(output, "✗") {
		t.Error("Expected output to contain error indicator")
	}

	// Should contain error message
	if !strings.Contains(output, "invalid YAML syntax") {
		t.Error("Expected output to contain error message")
	}

	// Should contain line number
	if !strings.Contains(output, "10") {
		t.Error("Expected output to contain line number")
	}
}

func TestFormatter_FormatResult_JSON(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage:  "local",
			Valid:  true,
			Errors: []validator.Error{},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "json", results, "test.yml")

	output := buf.String()

	// Should be valid JSON
	if !strings.HasPrefix(output, "{") {
		t.Error("Expected JSON output to start with '{'")
	}

	// Should contain file name
	if !strings.Contains(output, "test.yml") {
		t.Error("Expected JSON to contain filename")
	}

	// Should contain results array
	if !strings.Contains(output, `"results"`) {
		t.Error("Expected JSON to contain results field")
	}

	// Should contain stage
	if !strings.Contains(output, `"stage"`) {
		t.Error("Expected JSON to contain stage field")
	}

	if !strings.Contains(output, `"local"`) {
		t.Error("Expected JSON to contain local stage")
	}
}

func TestFormatter_FormatResult_YAML(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage:  "api",
			Valid:  true,
			Errors: []validator.Error{},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "yaml", results, "test.yml")

	output := buf.String()

	// Should be YAML-like
	if !strings.Contains(output, "file:") {
		t.Error("Expected YAML to contain file field")
	}

	if !strings.Contains(output, "results:") {
		t.Error("Expected YAML to contain results field")
	}

	// YAML will contain stage but not necessarily with "- stage:" prefix
	if !strings.Contains(output, "stage:") {
		t.Error("Expected YAML to contain stage field")
	}

	if !strings.Contains(output, "api") {
		t.Error("Expected YAML to contain api stage value")
	}
}

func TestFormatter_FormatResult_WithWarnings(t *testing.T) {
	formatter := NewFormatter("never", true)

	results := []validator.Result{
		{
			Stage:  "api",
			Valid:  true,
			Errors: []validator.Error{},
			Warnings: []validator.Warning{
				{
					Message: "deprecated job syntax",
					Line:    5,
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "text", results, ".gitlab-ci.yml")

	output := buf.String()

	// Verbose mode should show warnings
	if !strings.Contains(output, "deprecated job syntax") {
		t.Error("Expected verbose output to contain warnings")
	}
}

func TestFormatter_FormatMessage(t *testing.T) {
	formatter := NewFormatter("never", false)

	var buf bytes.Buffer
	formatter.FormatMessage(&buf, "info", "Test message")

	output := buf.String()

	if !strings.Contains(output, "Test message") {
		t.Error("Expected output to contain message")
	}
}

func TestFormatter_FormatResult_InvalidFormat(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage:  "local",
			Valid:  true,
			Errors: []validator.Error{},
		},
	}

	var buf bytes.Buffer
	// Invalid format should fall back to text format
	formatter.FormatResult(&buf, "invalid", results, "test.yml")

	output := buf.String()
	// Should have output (since invalid falls back to text)
	if len(output) == 0 {
		t.Error("Expected output even for invalid format (should fall back to text)")
	}
}

func TestFormatter_FormatResult_MultipleResults(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage:  "local",
			Valid:  true,
			Errors: []validator.Error{},
		},
		{
			Stage: "api",
			Valid: false,
			Errors: []validator.Error{
				{
					Message: "job not found",
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "text", results, ".gitlab-ci.yml")

	output := buf.String()

	// Should show the results - text format capitalizes stage names
	if !strings.Contains(output, "Local") {
		t.Error("Expected output to mention Local stage")
	}

	if !strings.Contains(output, "Api") {
		t.Error("Expected output to mention API stage")
	}
}

func TestFormatter_FormatResult_WithVerbose(t *testing.T) {
	verboseFormatter := NewFormatter("never", true)
	normalFormatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage:  "local",
			Valid:  true,
			Errors: []validator.Error{},
		},
	}

	var verboseBuf, normalBuf bytes.Buffer
	verboseFormatter.FormatResult(&verboseBuf, "text", results, "test.yml")
	normalFormatter.FormatResult(&normalBuf, "text", results, "test.yml")

	verboseOutput := verboseBuf.String()
	normalOutput := normalBuf.String()

	// Verbose output should be longer or contain more details
	if len(verboseOutput) == 0 && len(normalOutput) == 0 {
		// Both empty - might be valid for some cases
		return
	}

	// At minimum, verbose shouldn't be shorter than normal
	if len(verboseOutput) > 0 && len(verboseOutput) < len(normalOutput) {
		t.Error("Expected verbose output to be at least as long as normal output")
	}
}

func TestFormatter_FormatResult_ErrorWithContent(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage: "local",
			Valid: false,
			Errors: []validator.Error{
				{
					Message: "syntax error",
					Line:    15,
					Column:  3,
					Content: "  invalid_field:",
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "text", results, ".gitlab-ci.yml")

	output := buf.String()

	// Should show the problematic content
	if !strings.Contains(output, "invalid_field:") {
		t.Error("Expected output to show error context/content")
	}

	// Should show line and column
	if !strings.Contains(output, "15") {
		t.Error("Expected output to show line number")
	}

	if !strings.Contains(output, "3") {
		t.Error("Expected output to show column number")
	}
}

func TestFormatter_FormatResult_JSON_Escaping(t *testing.T) {
	formatter := NewFormatter("never", false)

	results := []validator.Result{
		{
			Stage: "local",
			Valid: false,
			Errors: []validator.Error{
				{
					Message: "Error with \"quotes\" and 'apostrophes'",
					Content: "value: \"test \\ value\"",
				},
			},
		},
	}

	var buf bytes.Buffer
	formatter.FormatResult(&buf, "json", results, "test.yml")

	output := buf.String()

	// Should be valid JSON (no unescaped quotes)
	// Simple check: should end with }
	if !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("Expected valid JSON output")
	}
}
