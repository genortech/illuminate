package ies

import (
	"illuminate/internal/converter"
	"illuminate/internal/models"
	"illuminate/internal/parsers/ies"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	writer := NewWriter()
	if writer == nil {
		t.Fatal("NewWriter() returned nil")
	}

	defaults := writer.GetDefaultOptions()
	if defaults.Precision != 1 {
		t.Errorf("Expected default precision 1, got %d", defaults.Precision)
	}
	if defaults.UseCommaDecimal {
		t.Error("Expected UseCommaDecimal to be false by default")
	}
	if !defaults.IncludeComments {
		t.Error("Expected IncludeComments to be true by default")
	}
	if defaults.FormatVersion != string(ies.VersionLM632002) {
		t.Errorf("Expected default version %s, got %s", ies.VersionLM632002, defaults.FormatVersion)
	}
}

func TestSetOptions(t *testing.T) {
	writer := NewWriter()

	tests := []struct {
		name        string
		options     converter.WriterOptions
		expectError bool
	}{
		{
			name: "valid options",
			options: converter.WriterOptions{
				Precision:       3,
				UseCommaDecimal: true,
				IncludeComments: false,
				FormatVersion:   string(ies.VersionLM631995),
			},
			expectError: false,
		},
		{
			name: "invalid version",
			options: converter.WriterOptions{
				FormatVersion: "INVALID",
			},
			expectError: true,
		},
		{
			name: "invalid precision - negative",
			options: converter.WriterOptions{
				Precision: -1,
			},
			expectError: true,
		},
		{
			name: "invalid precision - too high",
			options: converter.WriterOptions{
				Precision: 15,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.SetOptions(tt.options)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateForWrite(t *testing.T) {
	writer := NewWriter()

	tests := []struct {
		name        string
		data        *models.PhotometricData
		expectError bool
	}{
		{
			name:        "nil data",
			data:        nil,
			expectError: true,
		},
		{
			name:        "empty vertical angles",
			data:        createTestData([]float64{}, []float64{0, 90}, [][]float64{}),
			expectError: true,
		},
		{
			name:        "empty horizontal angles",
			data:        createTestData([]float64{0, 90}, []float64{}, [][]float64{}),
			expectError: true,
		},
		{
			name:        "mismatched candela values",
			data:        createTestData([]float64{0, 90}, []float64{0, 90}, [][]float64{{100}}),
			expectError: true,
		},
		{
			name:        "invalid vertical angle",
			data:        createTestData([]float64{-10, 90}, []float64{0, 90}, [][]float64{{100, 200}, {150, 250}}),
			expectError: true,
		},
		{
			name:        "invalid horizontal angle",
			data:        createTestData([]float64{0, 90}, []float64{0, 400}, [][]float64{{100, 200}, {150, 250}}),
			expectError: true,
		},
		{
			name:        "valid data",
			data:        createTestData([]float64{0, 90}, []float64{0, 90}, [][]float64{{100, 200}, {150, 250}}),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.ValidateForWrite(tt.data)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	writer := NewWriter()
	data := createTestData(
		[]float64{0, 45, 90},
		[]float64{0, 90, 180},
		[][]float64{
			{1000, 800, 600},
			{900, 700, 500},
			{800, 600, 400},
		},
	)

	output, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	// Check version header
	if !strings.HasPrefix(lines[0], "IESNA:LM-63-2002") {
		t.Errorf("Expected version header, got: %s", lines[0])
	}

	// Check for required keywords
	foundManufac := false
	foundLumcat := false
	foundTilt := false

	for _, line := range lines {
		if strings.Contains(line, "[MANUFAC]") {
			foundManufac = true
		}
		if strings.Contains(line, "[LUMCAT]") {
			foundLumcat = true
		}
		if strings.HasPrefix(line, "TILT=") {
			foundTilt = true
		}
	}

	if !foundManufac {
		t.Error("Missing MANUFAC keyword")
	}
	if !foundLumcat {
		t.Error("Missing LUMCAT keyword")
	}
	if !foundTilt {
		t.Error("Missing TILT line")
	}

	// Check that numeric data is present
	foundNumericData := false
	for _, line := range lines {
		if strings.Contains(line, "1000") {
			foundNumericData = true
			break
		}
	}
	if !foundNumericData {
		t.Error("Missing numeric candela data")
	}
}

func TestWriteWithCustomOptions(t *testing.T) {
	writer := NewWriter()

	// Set custom options
	options := converter.WriterOptions{
		Precision:       3,
		UseCommaDecimal: true,
		FormatVersion:   string(ies.VersionLM631995),
		CustomHeaders: map[string]string{
			"CUSTOM": "Test Value",
		},
	}

	err := writer.SetOptions(options)
	if err != nil {
		t.Fatalf("SetOptions() failed: %v", err)
	}

	data := createTestData(
		[]float64{0, 90},
		[]float64{0, 180},
		[][]float64{{1000.123, 800.456}, {900.789, 700.012}},
	)

	output, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	content := string(output)

	// Check version header for LM-63-1995
	if !strings.HasPrefix(content, "IESNA91") {
		t.Error("Expected IESNA91 version header")
	}

	// Check for custom header
	if !strings.Contains(content, "[CUSTOM] Test Value") {
		t.Error("Missing custom header")
	}

	// Check precision and comma decimal
	if !strings.Contains(content, "1000,123") {
		t.Error("Expected comma decimal separator with 3 decimal places")
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Create test data
	originalData := createTestData(
		[]float64{0, 30, 60, 90},
		[]float64{0, 90, 180, 270},
		[][]float64{
			{1000, 800, 600, 400},
			{900, 700, 500, 300},
			{800, 600, 400, 200},
			{700, 500, 300, 100},
		},
	)

	// Write to IES format
	writer := NewWriter()
	iesData, err := writer.Write(originalData)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Parse back from IES format
	parser := ies.NewParser()
	parsedData, err := parser.Parse(iesData)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Compare key values
	if len(parsedData.Photometry.VerticalAngles) != len(originalData.Photometry.VerticalAngles) {
		t.Errorf("Vertical angles count mismatch: expected %d, got %d",
			len(originalData.Photometry.VerticalAngles), len(parsedData.Photometry.VerticalAngles))
	}

	if len(parsedData.Photometry.HorizontalAngles) != len(originalData.Photometry.HorizontalAngles) {
		t.Errorf("Horizontal angles count mismatch: expected %d, got %d",
			len(originalData.Photometry.HorizontalAngles), len(parsedData.Photometry.HorizontalAngles))
	}

	// Check angle values (with tolerance for floating point precision)
	tolerance := 0.1
	for i, angle := range originalData.Photometry.VerticalAngles {
		if i < len(parsedData.Photometry.VerticalAngles) {
			diff := abs(angle - parsedData.Photometry.VerticalAngles[i])
			if diff > tolerance {
				t.Errorf("Vertical angle %d mismatch: expected %f, got %f",
					i, angle, parsedData.Photometry.VerticalAngles[i])
			}
		}
	}

	for i, angle := range originalData.Photometry.HorizontalAngles {
		if i < len(parsedData.Photometry.HorizontalAngles) {
			diff := abs(angle - parsedData.Photometry.HorizontalAngles[i])
			if diff > tolerance {
				t.Errorf("Horizontal angle %d mismatch: expected %f, got %f",
					i, angle, parsedData.Photometry.HorizontalAngles[i])
			}
		}
	}

	// Check candela values
	for i, row := range originalData.Photometry.CandelaValues {
		if i < len(parsedData.Photometry.CandelaValues) {
			for j, value := range row {
				if j < len(parsedData.Photometry.CandelaValues[i]) {
					diff := abs(value - parsedData.Photometry.CandelaValues[i][j])
					if diff > tolerance {
						t.Errorf("Candela value [%d][%d] mismatch: expected %f, got %f",
							i, j, value, parsedData.Photometry.CandelaValues[i][j])
					}
				}
			}
		}
	}

	// Check metadata
	if parsedData.Metadata.Manufacturer != originalData.Metadata.Manufacturer {
		t.Errorf("Manufacturer mismatch: expected %s, got %s",
			originalData.Metadata.Manufacturer, parsedData.Metadata.Manufacturer)
	}

	if parsedData.Metadata.CatalogNumber != originalData.Metadata.CatalogNumber {
		t.Errorf("Catalog number mismatch: expected %s, got %s",
			originalData.Metadata.CatalogNumber, parsedData.Metadata.CatalogNumber)
	}
}

func TestWriteFormatFloat(t *testing.T) {
	writer := NewWriter()

	tests := []struct {
		name     string
		value    float64
		options  converter.WriterOptions
		expected string
	}{
		{
			name:     "default precision",
			value:    123.456,
			options:  converter.WriterOptions{Precision: 1},
			expected: "123.5",
		},
		{
			name:     "high precision",
			value:    123.456789,
			options:  converter.WriterOptions{Precision: 4},
			expected: "123.4568",
		},
		{
			name:     "comma decimal",
			value:    123.456,
			options:  converter.WriterOptions{Precision: 2, UseCommaDecimal: true},
			expected: "123,46",
		},
		{
			name:     "zero precision",
			value:    123.456,
			options:  converter.WriterOptions{Precision: 0},
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer.SetOptions(tt.options)
			result := writer.formatFloat(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper functions

func createTestData(verticalAngles, horizontalAngles []float64, candelaValues [][]float64) *models.PhotometricData {
	return &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
			TestLab:       "Test Lab",
			TestDate:      "2023-01-01",
			TestNumber:    "12345",
		},
		Geometry: models.LuminaireGeometry{
			Length:         0.5,
			Width:          0.3,
			Height:         0.1,
			LuminousLength: 0.4,
			LuminousWidth:  0.25,
			LuminousHeight: 0.05,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      5000,
			CandelaMultiplier: 1.0,
			VerticalAngles:    verticalAngles,
			HorizontalAngles:  horizontalAngles,
			CandelaValues:     candelaValues,
		},
		Electrical: models.ElectricalData{
			InputWatts:        50.0,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
