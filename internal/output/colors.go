package output

import (
	"os"
)

// Color codes for terminal output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Gray   = "\033[90m"
)

// Colorizer handles color output
type Colorizer struct {
	enabled      bool
	colorSetting string
}

// NewColorizer creates a new colorizer based on the color setting
func NewColorizer(colorSetting string) *Colorizer {
	enabled := false

	switch colorSetting {
	case "always":
		enabled = true
	case "auto":
		// Detect if output is a terminal
		fileInfo, err := os.Stdout.Stat()
		if err == nil && (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			enabled = true
		}
	case "never":
		enabled = false
	}

	return &Colorizer{
		enabled:      enabled,
		colorSetting: colorSetting,
	}
}

// Color wraps text with ANSI color codes if enabled
func (c *Colorizer) Color(text, colorCode string) string {
	if !c.enabled {
		return text
	}
	return colorCode + text + Reset
}

// Red returns text in red
func (c *Colorizer) Red(text string) string {
	return c.Color(text, Red)
}

// Green returns text in green
func (c *Colorizer) Green(text string) string {
	return c.Color(text, Green)
}

// Yellow returns text in yellow
func (c *Colorizer) Yellow(text string) string {
	return c.Color(text, Yellow)
}

// Blue returns text in blue
func (c *Colorizer) Blue(text string) string {
	return c.Color(text, Blue)
}

// Gray returns text in gray
func (c *Colorizer) Gray(text string) string {
	return c.Color(text, Gray)
}

// IsEnabled returns whether coloring is enabled
func (c *Colorizer) IsEnabled() bool {
	return c.enabled
}

// GetColorSetting returns the color setting string
func (c *Colorizer) GetColorSetting() string {
	return c.colorSetting
}
