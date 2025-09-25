package cie

import (
	"illuminate/internal/formats/cie"
	"illuminate/internal/interfaces"
	"illuminate/internal/models"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	writer := NewWriter()

	if writer == nil {
		t.Fatal("NewWriter() returned nil")
	}

	options := writer.GetCIEOptions()
	if options.FormatType != 1 {
		t.Errorf("Default FormatType = %d, want 1", options.FormatType)
	}
	if options.SymmetryType != 0 {
		t.Errorf("Default SymmetryType = %d, want 0", options.SymmetryType)
	}
	if options.Precision != 0 {
		t.Errorf("Default Precision = %d, want 0", options.Precision)
	}
	if !options.IncludeDescription {
		t.Error("Default IncludeDescription should be true")
	}
}

func TestWriter_SetCIEOptions(t *testing.T) {
	writer := NewWriter()

	tests := []struct {
		name        string
		options     WriterOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid options",
			options: WriterOptions{
				FormatType:         1,
				SymmetryType:       1,
				Precision:          2,
				IncludeDescription: false,
			},
			expectError: false,
		},
		{
			name: "Invalid format type",
			options: WriterOptions{
				FormatType: 0,
			},
			expectError: true,
			errorMsg:    "format type must be >= 1",
		},
		{
			name: "Invalid symmetry type - negative",
			options: WriterOptions{
				FormatType:   1,
				SymmetryType: -1,
			},
			expectError: true,
			errorMsg:    "symmetry type must be 0 or 1",
		},
		{
			name: "Invalid symmetry type - too high",
			options: WriterOptions{
				FormatType:   1,
				SymmetryType: 2,
			},
			expectError: true,
			errorMsg:    "symmetry type must be 0 or 1",
		},
		{
			name: "Invalid precision - negative",
			options: WriterOptions{
				FormatType: 1,
				Precision:  -1,
			},
			expectError: true,
			errorMsg:    "precision must be between 0 and 6",
		},
		{
			name: "Invalid precision - too high",
			options: WriterOptions{
				FormatType: 1,
				Precision:  7,
			},
			expectError: true,
			errorMsg:    "precision must be between 0 and 6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.SetCIEOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("SetCIEOptions() expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("SetCIEOptions() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("SetCIEOptions() unexpected error: %v", err)
				} else {
					// Verify options were set
					currentOptions := writer.GetCIEOptions()
					if currentOptions.FormatType != tt.options.FormatType {
						t.Errorf("FormatType = %d, want %d", currentOptions.FormatType, tt.options.FormatType)
					}
					if currentOptions.SymmetryType != tt.options.SymmetryType {
						t.Errorf("SymmetryType = %d, want %d", currentOptions.SymmetryType, tt.options.SymmetryType)
					}
					if currentOptions.Precision != tt.options.Precision {
						t.Errorf("Precision = %d, want %d", currentOptions.Precision, tt.options.Precision)
					}
					if currentOptions.IncludeDescription != tt.options.IncludeDescription {
						t.Errorf("IncludeDescription = %v, want %v", currentOptions.IncludeDescription, tt.options.IncludeDescription)
					}
				}
			}
		})
	}
}

func TestWriter_Write(t *testing.T) {
	writer := NewWriter()

	// Create valid test data
	validData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "TestCorp",
			CatalogNumber: "LED123",
			Description:   "Test LED 17W",
		},
		Geometry: models.LuminaireGeometry{
			Length: 1.0,
			Width:  0.5,
			Height: 0.1,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0, 15, 30, 45, 60, 75, 90},
			HorizontalAngles:  []float64{0, 45, 90, 135, 180, 225, 270, 315},
			CandelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7, 0.6, 0.7, 0.8, 0.9},
				{0.9, 0.8, 0.7, 0.6, 0.5, 0.6, 0.7, 0.8},
				{0.8, 0.7, 0.6, 0.5, 0.4, 0.5, 0.6, 0.7},
				{0.7, 0.6, 0.5, 0.4, 0.3, 0.4, 0.5, 0.6},
				{0.5, 0.4, 0.3, 0.2, 0.1, 0.2, 0.3, 0.4},
				{0.3, 0.2, 0.1, 0.1, 0.0, 0.1, 0.1, 0.2},
				{0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:    17.0,
			BallastFactor: 1.0,
		},
	}

	tests := []struct {
		name        string
		data        *models.PhotometricData
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid data",
			data:        validData,
			expectError: false,
		},
		{
			name:        "Nil data",
			data:        nil,
			expectError: true,
			errorMsg:    "cannot be nil",
		},
		{
			name: "Invalid data - empty manufacturer",
			data: func() *models.PhotometricData {
				data := *validData
				data.Metadata.Manufacturer = ""
				return &data
			}(),
			expectError: true,
			errorMsg:    "manufacturer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := writer.Write(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Write() expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Write() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Write() unexpected error: %v", err)
				} else {
					if result == nil {
						t.Error("Write() returned nil result")
					} else {
						// Basic validation of output format
						content := string(result)
						lines := strings.Split(content, "\n")

						if len(lines) < 2 {
							t.Errorf("Expected at least 2 lines, got %d", len(lines))
						}

						// Check header line format
						headerLine := strings.TrimSpace(lines[0])
						if !strings.HasPrefix(headerLine, "1") {
							t.Errorf("Header line should start with '1', got: %s", headerLine)
						}

						// Check that we have intensity data
						hasIntensityData := false
						for i := 1; i < len(lines); i++ {
							line := strings.TrimSpace(lines[i])
							if line != "" {
								hasIntensityData = true
								break
							}
						}
						if !hasIntensityData {
							t.Error("Expected intensity data in output")
						}
					}
				}
			}
		})
	}
}

func TestWriter_formatHeaderLine(t *testing.T) {
	tests := []struct {
		name    string
		options WriterOptions
		header  struct {
			FormatType   int
			SymmetryType int
			Reserved     int
			Description  string
		}
		expected string
	}{
		{
			name: "Standard header with description",
			options: WriterOptions{
				IncludeDescription: true,
			},
			header: struct {
				FormatType   int
				SymmetryType int
				Reserved     int
				Description  string
			}{
				FormatType:   1,
				SymmetryType: 0,
				Reserved:     0,
				Description:  "Test LED 17W 1000 lm",
			},
			expected: "   1   0   0        Test LED 17W 1000 lm",
		},
		{
			name: "Header without description",
			options: WriterOptions{
				IncludeDescription: false,
			},
			header: struct {
				FormatType   int
				SymmetryType int
				Reserved     int
				Description  string
			}{
				FormatType:   1,
				SymmetryType: 1,
				Reserved:     0,
				Description:  "Test LED 17W 1000 lm",
			},
			expected: "   1   1   0",
		},
		{
			name: "Header with empty description",
			options: WriterOptions{
				IncludeDescription: true,
			},
			header: struct {
				FormatType   int
				SymmetryType int
				Reserved     int
				Description  string
			}{
				FormatType:   1,
				SymmetryType: 0,
				Reserved:     0,
				Description:  "",
			},
			expected: "   1   0   0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewWriter()
			writer.SetCIEOptions(WriterOptions{
				FormatType:         1,
				SymmetryType:       0,
				Precision:          0,
				IncludeDescription: tt.options.IncludeDescription,
			})

			// Create a mock CIE header
			header := struct {
				FormatType   int
				SymmetryType int
				Reserved     int
				Description  string
			}{
				FormatType:   tt.header.FormatType,
				SymmetryType: tt.header.SymmetryType,
				Reserved:     tt.header.Reserved,
				Description:  tt.header.Description,
			}

			// Create a proper CIE header for the actual method
			cieHeader := cie.CIEHeader{
				FormatType:   header.FormatType,
				SymmetryType: header.SymmetryType,
				Reserved:     header.Reserved,
				Description:  header.Description,
			}

			result := writer.formatHeaderLine(cieHeader)
			if result != tt.expected {
				t.Errorf("formatHeaderLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestWriter_formatIntensityValue(t *testing.T) {
	tests := []struct {
		name      string
		precision int
		value     float64
		expected  string
	}{
		{
			name:      "Integer precision",
			precision: 0,
			value:     123.456,
			expected:  " 123",
		},
		{
			name:      "Integer precision - small value",
			precision: 0,
			value:     5.7,
			expected:  "   6",
		},
		{
			name:      "One decimal place",
			precision: 1,
			value:     123.456,
			expected:  "123.5",
		},
		{
			name:      "Two decimal places",
			precision: 2,
			value:     123.456,
			expected:  "123.46",
		},
		{
			name:      "Zero value",
			precision: 0,
			value:     0.0,
			expected:  "   0",
		},
		{
			name:      "Large value",
			precision: 0,
			value:     9999.0,
			expected:  "9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewWriter()
			writer.SetCIEOptions(WriterOptions{
				FormatType:   1,
				SymmetryType: 0,
				Precision:    tt.precision,
			})

			result := writer.formatIntensityValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatIntensityValue(%f) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestWriter_formatIntensityData(t *testing.T) {
	writer := NewWriter()

	tests := []struct {
		name        string
		data        [][]float64
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid intensity data",
			data: [][]float64{
				{100, 90, 80},
				{70, 60, 50},
				{40, 30, 20},
			},
			expectError: false,
		},
		{
			name:        "Empty data",
			data:        [][]float64{},
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name: "Data with empty row",
			data: [][]float64{
				{100, 90, 80},
				{}, // Empty row
				{40, 30, 20},
			},
			expectError: false, // Should skip empty rows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := writer.formatIntensityData(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("formatIntensityData() expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("formatIntensityData() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("formatIntensityData() unexpected error: %v", err)
				} else {
					// Basic validation
					if result == "" && len(tt.data) > 0 {
						t.Error("formatIntensityData() returned empty result for non-empty data")
					}

					// Check that result contains expected number of lines
					if tt.name == "Valid intensity data" {
						lines := strings.Split(result, "\n")
						expectedLines := 3 // 3 rows of data
						if len(lines) != expectedLines {
							t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
						}
					}
				}
			}
		})
	}
}

func TestWriter_ValidateData(t *testing.T) {
	writer := NewWriter()

	validData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "TestCorp",
			CatalogNumber: "LED123",
		},
		Geometry: models.LuminaireGeometry{
			Length: 1.0,
			Width:  0.5,
			Height: 0.1,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0, 15, 30, 45, 60, 75, 90},
			HorizontalAngles:  []float64{0, 45, 90, 135, 180, 225, 270, 315},
			CandelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7, 0.6, 0.7, 0.8, 0.9},
				{0.9, 0.8, 0.7, 0.6, 0.5, 0.6, 0.7, 0.8},
				{0.8, 0.7, 0.6, 0.5, 0.4, 0.5, 0.6, 0.7},
				{0.7, 0.6, 0.5, 0.4, 0.3, 0.4, 0.5, 0.6},
				{0.5, 0.4, 0.3, 0.2, 0.1, 0.2, 0.3, 0.4},
				{0.3, 0.2, 0.1, 0.1, 0.0, 0.1, 0.1, 0.2},
				{0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:    17.0,
			BallastFactor: 1.0,
		},
	}

	tests := []struct {
		name        string
		data        *models.PhotometricData
		expectError bool
	}{
		{
			name:        "Valid data",
			data:        validData,
			expectError: false,
		},
		{
			name: "Invalid photometry type",
			data: func() *models.PhotometricData {
				data := *validData
				data.Photometry.PhotometryType = "A"
				return &data
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.ValidateData(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateData() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateData() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestWriter_GetSupportedFormats(t *testing.T) {
	writer := NewWriter()
	formats := writer.GetSupportedFormats()

	expectedFormats := []string{"CIE", "cie"}

	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}

	for i, expected := range expectedFormats {
		if i < len(formats) && formats[i] != expected {
			t.Errorf("Format %d = %v, want %v", i, formats[i], expected)
		}
	}
}

func TestWriter_GetFormatDescription(t *testing.T) {
	writer := NewWriter()
	description := writer.GetFormatDescription()

	if description == "" {
		t.Error("GetFormatDescription() returned empty string")
	}

	if !strings.Contains(strings.ToLower(description), "cie") {
		t.Errorf("Description should contain 'CIE', got: %s", description)
	}
}

func TestWriter_GetDefaultOptions(t *testing.T) {
	writer := NewWriter()
	options := writer.GetDefaultOptions()

	if options.Precision != 0 {
		t.Errorf("Default precision = %d, want 0", options.Precision)
	}
	if options.UseCommaDecimal {
		t.Error("Default UseCommaDecimal should be false")
	}
	if !options.IncludeComments {
		t.Error("Default IncludeComments should be true")
	}
	if options.FormatVersion != "CIE i-table" {
		t.Errorf("Default FormatVersion = %s, want 'CIE i-table'", options.FormatVersion)
	}
}

func TestWriter_SetOptions_StandardInterface(t *testing.T) {
	writer := NewWriter()

	opts := interfaces.WriterOptions{
		Precision:       2,
		UseCommaDecimal: false,
		IncludeComments: false,
		FormatVersion:   "CIE i-table",
	}

	err := writer.SetOptions(opts)
	if err != nil {
		t.Errorf("SetOptions() unexpected error: %v", err)
	}

	// Check that CIE-specific options were updated
	cieOpts := writer.GetCIEOptions()
	if cieOpts.Precision != 2 {
		t.Errorf("Precision = %d, want 2", cieOpts.Precision)
	}
	if cieOpts.IncludeDescription {
		t.Error("IncludeDescription should be false when IncludeComments is false")
	}
}

func TestWriter_ValidateForWrite(t *testing.T) {
	writer := NewWriter()

	validData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "TestCorp",
			CatalogNumber: "LED123",
		},
		Geometry: models.LuminaireGeometry{
			Length: 1.0,
			Width:  0.5,
			Height: 0.1,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0, 15, 30, 45, 60, 75, 90},
			HorizontalAngles:  []float64{0, 45, 90, 135, 180, 225, 270, 315},
			CandelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7, 0.6, 0.7, 0.8, 0.9},
				{0.9, 0.8, 0.7, 0.6, 0.5, 0.6, 0.7, 0.8},
				{0.8, 0.7, 0.6, 0.5, 0.4, 0.5, 0.6, 0.7},
				{0.7, 0.6, 0.5, 0.4, 0.3, 0.4, 0.5, 0.6},
				{0.5, 0.4, 0.3, 0.2, 0.1, 0.2, 0.3, 0.4},
				{0.3, 0.2, 0.1, 0.1, 0.0, 0.1, 0.1, 0.2},
				{0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:    17.0,
			BallastFactor: 1.0,
		},
	}

	err := writer.ValidateForWrite(validData)
	if err != nil {
		t.Errorf("ValidateForWrite() unexpected error: %v", err)
	}
}
