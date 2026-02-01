package validator

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// LocalValidator validates YAML syntax and basic structure
type LocalValidator struct {
	Strict bool
}

// NewLocalValidator creates a new local YAML validator
func NewLocalValidator(strict bool) *LocalValidator {
	return &LocalValidator{
		Strict: strict,
	}
}

// Validate validates the YAML content
func (v *LocalValidator) Validate(content []byte) Result {
	result := Result{
		Stage: "local",
	}

	// Parse YAML
	var node yaml.Node
	errorMsg := ""

	// First, check for basic YAML syntax errors
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(v.Strict)
	err := decoder.Decode(&node)

	if err != nil {
		errorMsg = err.Error()
	} else {
		// Check for extra content after first document (should only have one document)
		var extra interface{}
		extraDecoder := yaml.NewDecoder(bytes.NewReader(content))
		_ = extraDecoder.Decode(&node)
		if extraDecoder.Decode(&extra) == nil {
			result.Errors = append(result.Errors, Error{
				Message: "CI config should contain only one YAML document",
			})
			result.Valid = false
			return result
		}

		// Basic structure validation
		structureErrors := v.validateStructure(&node)
		if len(structureErrors) > 0 {
			result.Errors = structureErrors
			result.Valid = false
			return result
		}
	}

	if errorMsg != "" {
		result.Errors = v.parseYAMLError(errorMsg, content)
		result.Valid = false
		return result
	}

	result.Valid = true
	return result
}

// validateStructure performs basic structure validation on parsed YAML
func (v *LocalValidator) validateStructure(node *yaml.Node) []Error {
	var errors []Error

	// Root must be a mapping
	// Note: When decoding into a yaml.Node, the node itself contains the parsed document
	// The Kind field should be yaml.MappingNode for a proper GitLab CI config
	if node.Kind == 0 {
		// Node is empty or wasn't parsed correctly
		return append(errors, Error{
			Message: "Failed to parse YAML structure",
		})
	}

	// For a GitLab CI config, we expect a mapping at root level
	// But we'll be lenient here and let the API validation catch structural issues
	// since GitLab CI has many valid formats

	return errors
}

// parseYAMLError parses YAML error messages to extract line/column info
func (v *LocalValidator) parseYAMLError(errorMsg string, content []byte) []Error {
	var errors []Error

	// Parse common YAML error formats
	// Examples:
	// "yaml: line 10: found unexpected end of file"
	// "yaml: line 15: could not find expected ':'"
	// "yaml: unmarshal errors:\n  line 10: field test not found"

	lines := strings.Split(string(content), "\n")

	// Try to extract line number from error message
	line := 0
	column := 0

	// Look for line number
	if strings.Contains(errorMsg, "line ") {
		parts := strings.Split(errorMsg, "line ")
		if len(parts) > 1 {
			lineStr := strings.Split(parts[1], ":")[0]
			_, _ = fmt.Sscanf(lineStr, "%d", &line)
		}
	}

	// Look for column number
	if strings.Contains(errorMsg, "column ") {
		parts := strings.Split(errorMsg, "column ")
		if len(parts) > 1 {
			colStr := strings.Split(parts[1], " ")[0]
			_, _ = fmt.Sscanf(colStr, "%d", &column)
		}
	}

	// Extract error message
	msg := errorMsg
	if strings.HasPrefix(msg, "yaml:") {
		msg = strings.TrimPrefix(msg, "yaml:")
		msg = strings.TrimSpace(msg)
	}

	// Get content of the error line if line number is available
	lineContent := ""
	if line > 0 && line <= len(lines) {
		lineContent = strings.TrimSpace(lines[line-1])
	}

	errors = append(errors, Error{
		Message: msg,
		Line:    line,
		Column:  column,
		Content: lineContent,
	})

	return errors
}
