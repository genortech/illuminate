package ldt

import (
	"illuminate/internal/interfaces"
	"illuminate/internal/models"
	"illuminate/internal/parsers/ldt"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	writer := NewWriter()
	if writer == nil {
		t.Fatal("NewWriter() returned nil")
	}

	// Check default options
	opts := writer.GetDefaultOptions()
	if opts.Precision != 1 {
		t.Errorf("Expected default precision 1, got %d", opts.Precision)
	}
	if !opts.UseCommaDecimal {
		t.Error("Expected UseCommaDecimal to be true for LDT format")
	}
	if opts.IncludeComments {
		t.Error("Expected IncludeComments to be false for LDT format")
	}
	if opts.FormatVersion != string(ldt.Version10) {
		t.Errorf("Expected format version %s, got %s", ldt.Version10, opts.FormatVersion)
	}
}

func TestSetOptions(t *testing.T) {
	writer := NewWriter()

	tests := []struct {
		name        string
		options     interfaces.WriterOptions
		expectError bool
	}{
		{
			name: "valid options",
			options: interfaces.WriterOptions{
				Precision:       2,
				UseCommaDecimal: false,
				FormatVersion:   string(ldt.Version10),
			},
			expectError: false,
		},
		{
			name: "invalid version",
			options: interfaces.WriterOptions{
				FormatVersion: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid precision - negative",
			options: interfaces.WriterOptions{
				Precision: -1,
			},
			expectError: true,
		},
		{
			name: "invalid precision - too high",
			options: interfaces.WriterOptions{
				Precision: 11,
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
			data:        createTestData([]float64{}, []float64{0, 90}),
			expectError: true,
		},
		{
			name:        "empty horizontal angles",
			data:        createTestData([]float64{0, 90}, []float64{}),
			expectError: true,
		},
		{
			name:        "mismatched candela values rows",
			data:        createMismatchedData(),
			expectError: true,
		},
		{
			name:        "invalid vertical angle range",
			data:        createTestData([]float64{-10, 90}, []float64{0, 90}),
			expectError: true,
		},
		{
			name:        "invalid horizontal angle range",
			data:        createTestData([]float64{0, 90}, []float64{-10, 90}),
			expectError: true,
		},
		{
			name:        "too many horizontal angles",
			data:        createLargeAngleData(10, 361),
			expectError: true,
		},
		{
			name:        "too many vertical angles",
			data:        createLargeAngleData(182, 10),
			expectError: true,
		},
		{
			name:        "valid data",
			data:        createValidTestData(),
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
	data := createValidTestData()

	output, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	if len(output) == 0 {
		t.Fatal("Write() returned empty output")
	}

	// Convert to string for easier testing
	content := string(output)
	lines := strings.Split(content, "\n")

	// Check that we have enough lines for a complete LDT file
	expectedMinLines := 12 + // header
		13 + // geometry
		2 + // electrical header
		6 + // one lamp set
		len(data.Photometry.HorizontalAngles) + // C plane angles
		len(data.Photometry.VerticalAngles) + // gamma angles
		len(data.Photometry.HorizontalAngles)*len(data.Photometry.VerticalAngles) // intensity data

	if len(lines) < expectedMinLines {
		t.Errorf("Expected at least %d lines, got %d", expectedMinLines, len(lines))
	}

	// Check header content
	if !strings.Contains(lines[0], data.Metadata.Manufacturer) {
		t.Error("Company identification not found in output")
	}

	// Check that numeric values use comma as decimal separator (default for LDT)
	foundCommaDecimal := false
	for _, line := range lines[4:10] { // Check some numeric lines
		if strings.Contains(line, ",") && !strings.Contains(line, ";") {
			foundCommaDecimal = true
			break
		}
	}
	if !foundCommaDecimal {
		t.Error("Expected comma decimal separator in numeric values")
	}
}

func TestWriteWithCustomOptions(t *testing.T) {
	writer := NewWriter()

	// Set custom options
	opts := interfaces.WriterOptions{
		Precision:       2,
		UseCommaDecimal: false, // Use dot instead of comma
		CustomHeaders: map[string]string{
			"company":   "Custom Manufacturer",
			"luminaire": "Custom Luminaire",
			"catalog":   "CUSTOM-001",
		},
	}

	err := writer.SetOptions(opts)
	if err != nil {
		t.Fatalf("SetOptions() failed: %v", err)
	}

	data := createValidTestData()
	output, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	// Check custom headers were applied
	if !strings.Contains(lines[0], "Custom Manufacturer") {
		t.Error("Custom manufacturer not found in output")
	}

	// Check precision and decimal separator
	foundDotDecimal := false
	for _, line := range lines[4:10] {
		if strings.Contains(line, ".") && !strings.Contains(line, ",") {
			// Check if it has 2 decimal places
			parts := strings.Split(line, ".")
			if len(parts) == 2 && len(parts[1]) == 2 {
				foundDotDecimal = true
				break
			}
		}
	}
	if !foundDotDecimal {
		t.Error("Expected dot decimal separator with 2 decimal places")
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that we can write data and parse it back successfully
	writer := NewWriter()
	parser := ldt.NewParser()

	originalData := createValidTestData()

	// Write to LDT format
	ldtOutput, err := writer.Write(originalData)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Parse it back
	parsedData, err := parser.Parse(ldtOutput)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Compare key values (allowing for some precision loss)
	tolerance := 0.01

	if abs(originalData.Photometry.LuminousFlux-parsedData.Photometry.LuminousFlux) > tolerance {
		t.Errorf("Luminous flux mismatch: original=%f, parsed=%f",
			originalData.Photometry.LuminousFlux, parsedData.Photometry.LuminousFlux)
	}

	if len(originalData.Photometry.VerticalAngles) != len(parsedData.Photometry.VerticalAngles) {
		t.Errorf("Vertical angles count mismatch: original=%d, parsed=%d",
			len(originalData.Photometry.VerticalAngles), len(parsedData.Photometry.VerticalAngles))
	}

	if len(originalData.Photometry.HorizontalAngles) != len(parsedData.Photometry.HorizontalAngles) {
		t.Errorf("Horizontal angles count mismatch: original=%d, parsed=%d",
			len(originalData.Photometry.HorizontalAngles), len(parsedData.Photometry.HorizontalAngles))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper functions for creating test data

func createTestData(verticalAngles, horizontalAngles []float64) *models.PhotometricData {
	candelaValues := make([][]float64, len(verticalAngles))
	for i := range candelaValues {
		candelaValues[i] = make([]float64, len(horizontalAngles))
		for j := range candelaValues[i] {
			candelaValues[i][j] = 100.0 // Default test value
		}
	}

	return &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
		},
		Geometry: models.LuminaireGeometry{
			Length: 0.5,
			Width:  0.3,
			Height: 0.1,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    verticalAngles,
			HorizontalAngles:  horizontalAngles,
			CandelaValues:     candelaValues,
		},
		Electrical: models.ElectricalData{
			InputWatts:        20.0,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}
}

func createMismatchedData() *models.PhotometricData {
	data := createTestData([]float64{0, 90}, []float64{0, 90, 180})
	// Create mismatched candela values (wrong number of rows)
	data.Photometry.CandelaValues = [][]float64{
		{100, 90, 80}, // Only one row instead of two
	}
	return data
}

func createLargeAngleData(verticalCount, horizontalCount int) *models.PhotometricData {
	verticalAngles := make([]float64, verticalCount)
	for i := 0; i < verticalCount; i++ {
		verticalAngles[i] = float64(i)
	}

	horizontalAngles := make([]float64, horizontalCount)
	for i := 0; i < horizontalCount; i++ {
		horizontalAngles[i] = float64(i)
	}

	return createTestData(verticalAngles, horizontalAngles)
}

func createValidTestData() *models.PhotometricData {
	verticalAngles := []float64{0, 10, 20, 30, 45, 60, 90}
	horizontalAngles := []float64{0, 45, 90, 135, 180, 225, 270, 315}

	candelaValues := make([][]float64, len(verticalAngles))
	for i := range candelaValues {
		candelaValues[i] = make([]float64, len(horizontalAngles))
		for j := range candelaValues[i] {
			// Create some realistic test data
			candelaValues[i][j] = 1000.0 * (1.0 - float64(i)/float64(len(verticalAngles)))
		}
	}

	return &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test LED Luminaire",
			TestLab:       "Test Lab",
			TestDate:      "2023-01-01",
		},
		Geometry: models.LuminaireGeometry{
			Length:         0.6,
			Width:          0.3,
			Height:         0.1,
			LuminousLength: 0.5,
			LuminousWidth:  0.25,
			LuminousHeight: 0.05,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      5000.0,
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
