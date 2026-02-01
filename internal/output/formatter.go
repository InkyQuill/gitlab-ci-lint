package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/InkyQuill/gitlab-ci-lint/internal/validator"
	"gopkg.in/yaml.v3"
)

// Formatter handles output formatting
type Formatter struct {
	colorizer *Colorizer
	verbose   bool
	debug     *DebugLogger
}

// NewFormatter creates a new output formatter
func NewFormatter(colorSetting string, verbose bool) *Formatter {
	return &Formatter{
		colorizer: NewColorizer(colorSetting),
		verbose:   verbose,
		debug:     nil, // Debug logger initialized separately if needed
	}
}

// NewFormatterWithDebug creates a new output formatter with debug logging enabled
func NewFormatterWithDebug(colorSetting string, verbose bool, debugWriter io.Writer) *Formatter {
	return &Formatter{
		colorizer: NewColorizer(colorSetting),
		verbose:   verbose,
		debug:     NewDebugLogger(colorSetting, debugWriter),
	}
}

// SetDebug enables or disables debug logging
func (f *Formatter) SetDebug(enabled bool, writer io.Writer) {
	if enabled {
		f.debug = NewDebugLogger(f.colorizer.GetColorSetting(), writer)
	} else {
		f.debug = nil
	}
}

// GetDebugLogger returns the debug logger (may be nil if debug disabled)
func (f *Formatter) GetDebugLogger() *DebugLogger {
	return f.debug
}

// FormatResult formats and writes validation results to the output
func (f *Formatter) FormatResult(w io.Writer, format string, results []validator.Result, filename string) {
	switch format {
	case "json":
		f.formatJSON(w, results, filename)
	case "yaml":
		f.formatYAML(w, results, filename)
	default:
		f.formatText(w, results, filename)
	}
}

// formatText formats results as human-readable text
func (f *Formatter) formatText(w io.Writer, results []validator.Result, filename string) {
	allValid := true

	for _, result := range results {
		if !result.Valid {
			allValid = false
			break
		}
	}

	// Print header
	_, _ = fmt.Fprintf(w, "\n  %s\n\n", f.colorizer.Blue("GitLab CI Lint Results"))

	// Print filename
	_, _ = fmt.Fprintf(w, "  File: %s\n\n", f.colorizer.Gray(filename))

	// Print results for each stage
	for _, result := range results {
		f.formatTextResult(w, &result)
	}

	// Print overall status
	_, _ = fmt.Fprintf(w, "  ")
	if allValid {
		_, _ = fmt.Fprintf(w, "%s\n\n", f.colorizer.Green("✓ All validations passed"))
	} else {
		_, _ = fmt.Fprintf(w, "%s\n\n", f.colorizer.Red("✗ Validation failed"))
	}
}

// formatTextResult formats a single validation result as text
func (f *Formatter) formatTextResult(w io.Writer, result *validator.Result) {
	// Print stage (protect against empty stage)
	var stageName string
	if result.Stage == "" {
		stageName = "Unknown"
	} else {
		stageName = strings.ToUpper(result.Stage[:1]) + result.Stage[1:]
	}
	_, _ = fmt.Fprintf(w, "  %s: ", f.colorizer.Blue(stageName))

	if result.Valid {
		_, _ = fmt.Fprintf(w, "%s\n", f.colorizer.Green("Valid"))
	} else {
		_, _ = fmt.Fprintf(w, "%s\n", f.colorizer.Red("Invalid"))
	}

	// Print errors
	if len(result.Errors) > 0 {
		_, _ = fmt.Fprintf(w, "\n")
		for _, err := range result.Errors {
			f.formatError(w, &err)
		}
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintf(w, "\n")
		for _, warn := range result.Warnings {
			f.formatWarning(w, &warn)
		}
	}

	_, _ = fmt.Fprintf(w, "\n")
}

// formatError formats a validation error
func (f *Formatter) formatError(w io.Writer, err *validator.Error) {
	prefix := f.colorizer.Red("    ✗")

	if err.Line > 0 {
		loc := fmt.Sprintf("line %d", err.Line)
		if err.Column > 0 {
			loc += fmt.Sprintf(":%d", err.Column)
		}
		_, _ = fmt.Fprintf(w, "%s %s: %s\n", prefix, f.colorizer.Gray(loc), err.Message)
		if err.Content != "" {
			_, _ = fmt.Fprintf(w, "      %s\n", f.colorizer.Gray(err.Content))
		}
	} else {
		_, _ = fmt.Fprintf(w, "%s %s\n", prefix, err.Message)
	}
}

// formatWarning formats a validation warning
func (f *Formatter) formatWarning(w io.Writer, warn *validator.Warning) {
	prefix := f.colorizer.Yellow("    ⚠")

	if warn.Line > 0 {
		_, _ = fmt.Fprintf(w, "%s %s (line %d)\n", prefix, warn.Message, warn.Line)
	} else {
		_, _ = fmt.Fprintf(w, "%s %s\n", prefix, warn.Message)
	}
}

// formatJSON formats results as JSON
func (f *Formatter) formatJSON(w io.Writer, results []validator.Result, filename string) {
	output := map[string]interface{}{
		"file":    filename,
		"valid":   f.allValid(results),
		"results": results,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(output)
}

// formatYAML formats results as YAML
func (f *Formatter) formatYAML(w io.Writer, results []validator.Result, filename string) {
	output := map[string]interface{}{
		"file":    filename,
		"valid":   f.allValid(results),
		"results": results,
	}

	data, err := yaml.Marshal(output)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error formatting YAML: %v\n", err)
		return
	}

	_, _ = fmt.Fprintln(w, string(data))
}

// FormatMessage formats a generic message
func (f *Formatter) FormatMessage(w io.Writer, level, message string) {
	prefix := ""
	switch level {
	case "error":
		prefix = f.colorizer.Red("Error:")
	case "warning":
		prefix = f.colorizer.Yellow("Warning:")
	case "info":
		prefix = f.colorizer.Blue("Info:")
	}

	_, _ = fmt.Fprintf(w, "%s %s\n", prefix, message)
}

// allValid checks if all results are valid
func (f *Formatter) allValid(results []validator.Result) bool {
	for _, result := range results {
		if !result.Valid {
			return false
		}
	}
	return true
}

// countValidationResults counts total, valid, and invalid results
func (f *Formatter) countValidationResults(results []validator.FileResult) (total, valid, invalid int) {
	total = len(results)
	for _, result := range results {
		if result.Valid {
			valid++
		} else {
			invalid++
		}
	}
	return
}

// FormatSummary formats and writes a summary of batch validation results
func (f *Formatter) FormatSummary(w io.Writer, format string, results []validator.FileResult) {
	switch format {
	case "json":
		f.formatSummaryJSON(w, results)
	case "yaml":
		f.formatSummaryYAML(w, results)
	default:
		f.formatSummaryText(w, results)
	}
}

// formatSummaryText formats summary as human-readable text
func (f *Formatter) formatSummaryText(w io.Writer, results []validator.FileResult) {
	total, valid, invalid := f.countValidationResults(results)

	_, _ = fmt.Fprintf(w, "\n  %s\n\n", f.colorizer.Blue("Summary"))
	_, _ = fmt.Fprintf(w, "  Total files: %d\n", total)
	_, _ = fmt.Fprintf(w, "  %s: %d\n", f.colorizer.Green("Valid"), valid)
	_, _ = fmt.Fprintf(w, "  %s: %d\n\n", f.colorizer.Red("Invalid"), invalid)
}

// formatSummaryJSON formats summary as JSON
func (f *Formatter) formatSummaryJSON(w io.Writer, results []validator.FileResult) {
	total, valid, invalid := f.countValidationResults(results)

	output := map[string]interface{}{
		"summary": map[string]int{
			"total":   total,
			"valid":   valid,
			"invalid": invalid,
		},
		"files": results,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(output)
}

// formatSummaryYAML formats summary as YAML
func (f *Formatter) formatSummaryYAML(w io.Writer, results []validator.FileResult) {
	total, valid, invalid := f.countValidationResults(results)

	output := map[string]interface{}{
		"summary": map[string]int{
			"total":   total,
			"valid":   valid,
			"invalid": invalid,
		},
		"files": results,
	}

	data, err := yaml.Marshal(output)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error formatting YAML: %v\n", err)
		return
	}

	_, _ = fmt.Fprintln(w, string(data))
}
