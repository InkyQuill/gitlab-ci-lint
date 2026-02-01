package validator

// Result represents the result of a validation operation
type Result struct {
	Valid    bool      `yaml:"valid" json:"valid"`
	Errors   []Error   `yaml:"errors,omitempty" json:"errors,omitempty"`
	Warnings []Warning `yaml:"warnings,omitempty" json:"warnings,omitempty"`
	Stage    string    `yaml:"stage" json:"stage"` // "local" or "api"
}

// Error represents a validation error
type Error struct {
	Message string `yaml:"message" json:"message"`
	Line    int    `yaml:"line,omitempty" json:"line,omitempty"`
	Column  int    `yaml:"column,omitempty" json:"column,omitempty"`
	Content string `yaml:"content,omitempty" json:"content,omitempty"` // Error line content
}

// Warning represents a validation warning
type Warning struct {
	Message string `yaml:"message" json:"message"`
	Line    int    `yaml:"line,omitempty" json:"line,omitempty"`
}

// Validator defines the interface for validation operations
type Validator interface {
	Validate(content []byte) Result
}
