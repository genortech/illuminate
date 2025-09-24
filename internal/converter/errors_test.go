package converter

import (
	"errors"
	"testing"
)

func TestConversionError(t *testing.T) {
	// Test creating different error types
	syntaxErr := NewSyntaxError("INVALID_FORMAT", "File format is not recognized")
	if syntaxErr.Category != SyntaxError {
		t.Errorf("Expected category %s, got %s", SyntaxError, syntaxErr.Category)
	}

	semanticErr := NewSemanticError("INVALID_VALUE", "Candela value is negative")
	if semanticErr.Category != SemanticError {
		t.Errorf("Expected category %s, got %s", SemanticError, semanticErr.Category)
	}

	// Test adding context and warnings
	syntaxErr.AddContext("line", 42)
	syntaxErr.AddContext("column", 15)
	syntaxErr.AddWarning("This might cause issues")

	if len(syntaxErr.Context) != 2 {
		t.Errorf("Expected 2 context items, got %d", len(syntaxErr.Context))
	}

	if len(syntaxErr.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(syntaxErr.Warnings))
	}

	// Test error message formatting
	errorMsg := syntaxErr.Error()
	if errorMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, SystemError, "IO_ERROR", "Failed to read file")

	if wrappedErr.Category != SystemError {
		t.Errorf("Expected category %s, got %s", SystemError, wrappedErr.Category)
	}

	if wrappedErr.Cause != originalErr {
		t.Error("Wrapped error should preserve original cause")
	}

	// Test unwrapping
	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap should return original error")
	}
}

func TestErrorRecovery(t *testing.T) {
	recovery := NewErrorRecovery()

	// Initially should have no errors or warnings
	if recovery.HasErrors() {
		t.Error("New recovery should not have errors")
	}
	if recovery.HasWarnings() {
		t.Error("New recovery should not have warnings")
	}

	// Add some errors and warnings
	recovery.AddError(errors.New("first error"))
	recovery.AddError(errors.New("second error"))
	recovery.AddWarning("first warning")

	if !recovery.HasErrors() {
		t.Error("Recovery should have errors after adding them")
	}
	if !recovery.HasWarnings() {
		t.Error("Recovery should have warnings after adding them")
	}

	if len(recovery.GetErrors()) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(recovery.GetErrors()))
	}
	if len(recovery.GetWarnings()) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(recovery.GetWarnings()))
	}

	// Test conversion to ConversionError
	convErr := recovery.ToConversionError("MULTIPLE_ERRORS", "Multiple errors occurred")
	if convErr == nil {
		t.Error("Should create ConversionError when errors exist")
	}

	if len(convErr.Warnings) != 1 {
		t.Errorf("ConversionError should have 1 warning, got %d", len(convErr.Warnings))
	}

	// Test clearing
	recovery.Clear()
	if recovery.HasErrors() || recovery.HasWarnings() {
		t.Error("Recovery should be empty after clearing")
	}
}
