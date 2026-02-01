package output

import (
	"fmt"
	"io"
	"strings"
)

// DebugLogger handles debug output with categories
type DebugLogger struct {
	colorizer *Colorizer
	writer    io.Writer
}

// NewDebugLogger creates a new debug logger
func NewDebugLogger(colorSetting string, writer io.Writer) *DebugLogger {
	return &DebugLogger{
		colorizer: NewColorizer(colorSetting),
		writer:    writer,
	}
}

// Log logs a debug message with a category
// Categories: CONFIG, DISCOVER, DETECT, API, GITLAB, VALIDATE
func (d *DebugLogger) Log(category, message string) {
	prefix := d.colorizer.Gray(fmt.Sprintf("[DEBUG:%s]", category))
	_, _ = fmt.Fprintf(d.writer, "%s %s\n", prefix, message)
}

// LogAPIRequest logs an API request with details
func (d *DebugLogger) LogAPIRequest(endpoint, method string, hasProject bool, project string) {
	projectInfo := "no project"
	if hasProject && project != "" {
		projectInfo = fmt.Sprintf("project=%s", project)
	}
	d.Log("API", fmt.Sprintf("%s %s %s", method, endpoint, projectInfo))
}

// LogProjectDetection logs project detection results
func (d *DebugLogger) LogProjectDetection(source, project, filePath string) {
	if project != "" {
		d.Log("DETECT", fmt.Sprintf("project=%s from %s for %s", project, source, filePath))
	} else {
		d.Log("DETECT", fmt.Sprintf("no project detected from %s for %s", source, filePath))
	}
}

// LogConfig logs configuration loading
func (d *DebugLogger) LogConfig(key, value string) {
	d.Log("CONFIG", fmt.Sprintf("%s=%s", key, value))
}

// LogDiscovery logs file discovery
func (d *DebugLogger) LogDiscovery(message string) {
	d.Log("DISCOVER", message)
}

// LogGitLab logs GitLab client operations
func (d *DebugLogger) LogGitLab(message string) {
	d.Log("GITLAB", message)
}

// LogValidate logs validation operations
func (d *DebugLogger) LogValidate(message string) {
	d.Log("VALIDATE", message)
}

// LogSection logs a section header for grouping debug output
func (d *DebugLogger) LogSection(title string) {
	separator := strings.Repeat("-", 60)
	_, _ = fmt.Fprintf(d.writer, "\n%s\n%s\n%s\n\n", d.colorizer.Gray(separator), d.colorizer.Gray(title), d.colorizer.Gray(separator))
}
