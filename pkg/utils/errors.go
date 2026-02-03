package utils

import (
	"fmt"
	"os"
)

// ExitCode represents the exit code for the application
type ExitCode int

const (
	// ExitSuccess indicates successful execution
	ExitSuccess ExitCode = 0
	// ExitError indicates a general error
	ExitError ExitCode = 1
	// ExitConfigError indicates a configuration error
	ExitConfigError ExitCode = 2
	// ExitDiscordError indicates a Discord API error
	ExitDiscordError ExitCode = 3
	// ExitValidationError indicates a validation error
	ExitValidationError ExitCode = 4
	// ExitNotFound indicates a resource was not found
	ExitNotFound ExitCode = 5
)

// CLIError is the base error type for CLI errors
type CLIError struct {
	Message string
	Code    ExitCode
	Cause   error
}

// Error implements the error interface
func (e *CLIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *CLIError) Unwrap() error {
	return e.Cause
}

// Exit exits the application with the appropriate exit code
func (e *CLIError) Exit() {
	os.Exit(int(e.Code))
}

// NewError creates a new CLIError
func NewError(message string, code ExitCode) *CLIError {
	return &CLIError{
		Message: message,
		Code:    code,
	}
}

// NewErrorWithCause creates a new CLIError with a cause
func NewErrorWithCause(message string, code ExitCode, cause error) *CLIError {
	return &CLIError{
		Message: message,
		Code:    code,
		Cause:   cause,
	}
}

// Common error constructors

// ConfigError creates a configuration error
func ConfigError(message string) *CLIError {
	return NewError(message, ExitConfigError)
}

// ConfigErrorf creates a formatted configuration error
func ConfigErrorf(format string, args ...interface{}) *CLIError {
	// Convert any %w verbs to %v for compatibility
	format = convertFormatVerbs(format)
	return NewError(fmt.Sprintf(format, args...), ExitConfigError)
}

// DiscordError creates a Discord API error
func DiscordError(message string) *CLIError {
	return NewError(message, ExitDiscordError)
}

// DiscordErrorf creates a formatted Discord API error
func DiscordErrorf(format string, args ...interface{}) *CLIError {
	// Convert any %w verbs to %v for compatibility
	format = convertFormatVerbs(format)
	return NewError(fmt.Sprintf(format, args...), ExitDiscordError)
}

// ValidationError creates a validation error
func ValidationError(message string) *CLIError {
	return NewError(message, ExitValidationError)
}

// ValidationErrorf creates a formatted validation error
func ValidationErrorf(format string, args ...interface{}) *CLIError {
	// Convert any %w verbs to %v for compatibility
	format = convertFormatVerbs(format)
	return NewError(fmt.Sprintf(format, args...), ExitValidationError)
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *CLIError {
	return NewError(fmt.Sprintf("%s not found", resource), ExitNotFound)
}

// NotFoundErrorf creates a formatted not found error
func NotFoundErrorf(format string, args ...interface{}) *CLIError {
	return NewError(fmt.Sprintf(format, args...), ExitNotFound)
}

// WrapError wraps an error with a CLIError
func WrapError(err error, message string, code ExitCode) *CLIError {
	if err == nil {
		return nil
	}
	return NewErrorWithCause(message, code, err)
}

// HandleError handles an error and exits if it's a CLIError
func HandleError(err error) {
	if err == nil {
		return
	}

	if cliErr, ok := err.(*CLIError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", cliErr.Message)
		if cliErr.Cause != nil {
			fmt.Fprintf(os.Stderr, "  Cause: %v\n", cliErr.Cause)
		}
		cliErr.Exit()
	}

	// Generic error handling
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(int(ExitError))
}

// PrintError prints an error without exiting
func PrintError(err error) {
	if err == nil {
		return
	}

	if cliErr, ok := err.(*CLIError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", cliErr.Message)
		if cliErr.Cause != nil {
			fmt.Fprintf(os.Stderr, "  Cause: %v\n", cliErr.Cause)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// convertFormatVerbs converts %w verbs to %v for compatibility with fmt.Sprintf
func convertFormatVerbs(format string) string {
	return fmt.Sprintf("%s", format) // Simple pass-through, %w is handled by fmt
}

// IsCLIError checks if an error is a CLIError
func IsCLIError(err error) bool {
	_, ok := err.(*CLIError)
	return ok
}

// GetExitCode gets the exit code from an error
func GetExitCode(err error) ExitCode {
	if err == nil {
		return ExitSuccess
	}

	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Code
	}

	return ExitError
}
