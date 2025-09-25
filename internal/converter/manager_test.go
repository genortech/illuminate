package converter

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if len(manager.parsers) == 0 {
		t.Error("Manager has no parsers registered")
	}

	if len(manager.writers) == 0 {
		t.Error("Manager has no writers registered")
	}

	// Check that all expected formats are registered
	expectedFormats := []string{"ies", "ldt", "cie"}
	for _, format := range expectedFormats {
		if _, exists := manager.parsers[format]; !exists {
			t.Errorf("Parser for format %s not registered", format)
		}
		if _, exists := manager.writers[format]; !exists {
			t.Errorf("Writer for format %s not registered", format)
		}
	}
}

func TestGetSupportedFormats(t *testing.T) {
	manager := NewManager()
	formats := manager.GetSupportedFormats()

	expectedFormats := []string{"cie", "ies", "ldt"} // Should be sorted
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}

	for i, expected := range expectedFormats {
		if i >= len(formats) || formats[i] != expected {
			t.Errorf("Expected format %s at position %d, got %s", expected, i, formats[i])
		}
	}
}

func TestGetFormatInfo(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		format      string
		expectError bool
	}{
		{"ies", false},
		{"ldt", false},
		{"cie", false},
		{"IES", false}, // Should handle case insensitivity
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			info, err := manager.GetFormatInfo(tt.format)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if info == nil {
				t.Error("Expected format info but got nil")
				return
			}

			// Validate format info structure
			if info.Identifier == "" {
				t.Error("Format identifier is empty")
			}
			if info.Name == "" {
				t.Error("Format name is empty")
			}
			if len(info.SupportedVersions) == 0 {
				t.Error("No supported versions listed")
			}
			if len(info.FileExtensions) == 0 {
				t.Error("No file extensions listed")
			}
		})
	}
}

func TestDetectFormat_EmptyData(t *testing.T) {
	manager := NewManager()

	result, err := manager.DetectFormat([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
	if result != nil {
		t.Error("Expected nil result for empty data")
	}
}

func TestDetectFormat_ValidData(t *testing.T) {
	manager := NewManager()

	// Test with sample IES data
	iesData := `IESNA:LM-63-2002
[TEST] Test Luminaire
[MANUFAC] Test Manufacturer
TILT=NONE
1 1000 1 1 1 1 1 -1 1 1 1
0 0 90
0 90
1000`

	result, err := manager.DetectFormat([]byte(iesData))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected detection result but got nil")
		return
	}

	if result.Format != "ies" {
		t.Errorf("Expected format 'ies', got '%s'", result.Format)
	}

	if result.Confidence <= 0 {
		t.Errorf("Expected positive confidence, got %f", result.Confidence)
	}
}

func TestValidateData_NilData(t *testing.T) {
	manager := NewManager()

	result := manager.ValidateData(nil)
	if result == nil {
		t.Error("Expected validation result but got nil")
		return
	}

	if result.IsValid {
		t.Error("Expected validation to fail for nil data")
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score 0.0 for nil data, got %f", result.Score)
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for nil data")
	}
}
