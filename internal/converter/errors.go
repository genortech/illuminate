package converter

import (
	"fmt"
	"strings"
)

// ErrorCategory represents different types of conversion errors
type ErrorCategory string

const (
	// SyntaxError indicates invalid file format or corrupted data
	SyntaxError ErrorCategory = "syntax_error"
	// SemanticError indicates invalid photometric values or incompatible parameters
	SemanticError ErrorCategory = "semantic_error"
	// ConversionError indicates format limitations or data loss warnings
	ConversionError ErrorCategory = "conversion_error"
	// SystemError indicates file I/O, memory limitations, or other system issues
	SystemError ErrorCategory = "system_error"
	// ValidationError indicates data validation failures
	ValidationError ErrorCategory = "validation_error"
)

// ConversionError represents a structured error with context and warnings
type ConversionErr struct {
	Code     string                 `json:"code"`
	Category ErrorCategory          `json:"category"`
	Message  string                 `json:"message"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Cause    error                  `json:"-"` // Original error, not serialized
}

// Error implements the error interface
func (e *ConversionErr) Error() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s:%s] %s", e.Category, e.Code, e.Message))

	if len(e.Context) > 0 {
		var contextParts []string
		for k, v := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("context: %s", strings.Join(contextParts, ", ")))
	}

	if len(e.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("warnings: %s", strings.Join(e.Warnings, "; ")))
	}

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("cause: %s", e.Cause.Error()))
	}

	return strings.Join(parts, " | ")
}

// Unwrap returns the underlying cause error
func (e *ConversionErr) Unwrap() error {
	return e.Cause
}

// AddWarning adds a warning message to the error
func (e *ConversionErr) AddWarning(warning string) {
	e.Warnings = append(e.Warnings, warning)
}

// AddContext adds context information to the error
func (e *ConversionErr) AddContext(key string, value interface{}) {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
}

// NewSyntaxError creates a new syntax error
func NewSyntaxError(code, message string) *ConversionErr {
	return &ConversionErr{
		Code:     code,
		Category: SyntaxError,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// NewSemanticError creates a new semantic error
func NewSemanticError(code, message string) *ConversionErr {
	return &ConversionErr{
		Code:     code,
		Category: SemanticError,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// NewConversionError creates a new conversion error
func NewConversionError(code, message string) *ConversionErr {
	return &ConversionErr{
		Code:     code,
		Category: ConversionError,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// NewSystemError creates a new system error
func NewSystemError(code, message string, cause error) *ConversionErr {
	return &ConversionErr{
		Code:     code,
		Category: SystemError,
		Message:  message,
		Context:  make(map[string]interface{}),
		Cause:    cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string, cause error) *ConversionErr {
	return &ConversionErr{
		Code:     code,
		Category: ValidationError,
		Message:  message,
		Context:  make(map[string]interface{}),
		Cause:    cause,
	}
}

// WrapError wraps an existing error with conversion error context
func WrapError(err error, category ErrorCategory, code, message string) *ConversionErr {
	return &ConversionErr{
		Code:     code,
		Category: category,
		Message:  message,
		Context:  make(map[string]interface{}),
		Cause:    err,
	}
}

// ErrorRecovery provides mechanisms for error recovery and reporting
type ErrorRecovery struct {
	errors   []error
	warnings []string
}

// NewErrorRecovery creates a new error recovery instance
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{
		errors:   make([]error, 0),
		warnings: make([]string, 0),
	}
}

// AddError adds an error to the recovery context
func (er *ErrorRecovery) AddError(err error) {
	er.errors = append(er.errors, err)
}

// AddWarning adds a warning to the recovery context
func (er *ErrorRecovery) AddWarning(warning string) {
	er.warnings = append(er.warnings, warning)
}

// HasErrors returns true if there are any errors
func (er *ErrorRecovery) HasErrors() bool {
	return len(er.errors) > 0
}

// HasWarnings returns true if there are any warnings
func (er *ErrorRecovery) HasWarnings() bool {
	return len(er.warnings) > 0
}

// GetErrors returns all collected errors
func (er *ErrorRecovery) GetErrors() []error {
	return er.errors
}

// GetWarnings returns all collected warnings
func (er *ErrorRecovery) GetWarnings() []string {
	return er.warnings
}

// ToConversionError converts the recovery context to a single ConversionError
func (er *ErrorRecovery) ToConversionError(code, message string) *ConversionErr {
	if !er.HasErrors() {
		return nil
	}

	convErr := NewSystemError(code, message, er.errors[0])
	convErr.Warnings = er.warnings

	// Add additional errors as context
	if len(er.errors) > 1 {
		var errorMessages []string
		for i, err := range er.errors[1:] {
			errorMessages = append(errorMessages, fmt.Sprintf("error_%d: %s", i+2, err.Error()))
		}
		convErr.AddContext("additional_errors", errorMessages)
	}

	return convErr
}

// Clear resets the error recovery context
func (er *ErrorRecovery) Clear() {
	er.errors = er.errors[:0]
	er.warnings = er.warnings[:0]
}
