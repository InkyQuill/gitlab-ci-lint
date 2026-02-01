package exit

// Exit codes
const (
	ExitSuccess         = 0  // All validations passed
	ExitGeneralError    = 1  // Runtime error (file not found, auth failed, etc)
	ExitValidationError = 10 // CI configuration invalid
)
