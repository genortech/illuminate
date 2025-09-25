package cie

import (
	"illuminate/internal/models"
	"illuminate/internal/parsers/cie"
	"strings"
	"testing"
)

func TestCIE_RoundTrip(t *testing.T) {
	// Test data representing a simple CIE i-table file
	originalCIEData := `   1   0   0        Test LED 17W 1000 lm
 100  90  80  70  60  70  80  90 100  90  80  70  60  70  80  90
  90  80  70  60  50  60  70  80  90  80  70  60  50  60  70  80
  80  70  60  50  40  50  60  70  80  70  60  50  40  50  60  70
  70  60  50  40  30  40  50  60  70  60  50  40  30  40  50  60
  50  40  30  20  10  20  30  40  50  40  30  20  10  20  30  40
  30  20  10  10   0  10  10  20  30  20  10  10   0  10  10  20
  10   0   0   0   0   0   0   0  10   0   0   0   0   0   0   0
   5   0   0   0   0   0   0   0   5   0   0   0   0   0   0   0
   2   0   0   0   0   0   0   0   2   0   0   0   0   0   0   0
   1   0   0   0   0   0   0   0   1   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0`

	// Parse the original data
	parser := cie.NewParser()
	photometricData, err := parser.Parse([]byte(originalCIEData))
	if err != nil {
		t.Fatalf("Failed to parse original CIE data: %v", err)
	}

	// Write it back to CIE format
	writer := NewWriter()
	outputData, err := writer.Write(photometricData)
	if err != nil {
		t.Fatalf("Failed to write CIE data: %v", err)
	}

	// Parse the output data again
	reparsedData, err := parser.Parse(outputData)
	if err != nil {
		t.Fatalf("Failed to reparse written CIE data: %v", err)
	}

	// Compare key fields to ensure round-trip accuracy
	if photometricData.Photometry.PhotometryType != reparsedData.Photometry.PhotometryType {
		t.Errorf("PhotometryType mismatch: original=%s, reparsed=%s",
			photometricData.Photometry.PhotometryType, reparsedData.Photometry.PhotometryType)
	}

	if len(photometricData.Photometry.VerticalAngles) != len(reparsedData.Photometry.VerticalAngles) {
		t.Errorf("VerticalAngles length mismatch: original=%d, reparsed=%d",
			len(photometricData.Photometry.VerticalAngles), len(reparsedData.Photometry.VerticalAngles))
	}

	if len(photometricData.Photometry.HorizontalAngles) != len(reparsedData.Photometry.HorizontalAngles) {
		t.Errorf("HorizontalAngles length mismatch: original=%d, reparsed=%d",
			len(photometricData.Photometry.HorizontalAngles), len(reparsedData.Photometry.HorizontalAngles))
	}

	if len(photometricData.Photometry.CandelaValues) != len(reparsedData.Photometry.CandelaValues) {
		t.Errorf("CandelaValues rows mismatch: original=%d, reparsed=%d",
			len(photometricData.Photometry.CandelaValues), len(reparsedData.Photometry.CandelaValues))
	}

	// Check that the output contains expected CIE format elements
	outputString := string(outputData)
	if !containsString(outputString, "Test LED") {
		t.Error("Output should contain luminaire description")
	}

	// Verify the output starts with proper CIE header format
	lines := strings.Split(outputString, "\n")
	if len(lines) < 2 {
		t.Errorf("Output should have at least 2 lines, got %d", len(lines))
	}

	headerLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(headerLine, "1") {
		t.Errorf("Header line should start with format type '1', got: %s", headerLine)
	}
}

func TestCIE_WriterWithDifferentPrecision(t *testing.T) {
	// Create test photometric data
	photometricData := createTestPhotometricData()

	writer := NewWriter()

	// Test with different precision settings
	precisionTests := []struct {
		precision int
		name      string
	}{
		{0, "Integer precision"},
		{1, "One decimal place"},
		{2, "Two decimal places"},
	}

	for _, tt := range precisionTests {
		t.Run(tt.name, func(t *testing.T) {
			// Set precision
			err := writer.SetCIEOptions(WriterOptions{
				FormatType:         1,
				SymmetryType:       0,
				Precision:          tt.precision,
				IncludeDescription: true,
			})
			if err != nil {
				t.Fatalf("Failed to set options: %v", err)
			}

			// Write the data
			outputData, err := writer.Write(photometricData)
			if err != nil {
				t.Fatalf("Failed to write data: %v", err)
			}

			// Verify output is not empty
			if len(outputData) == 0 {
				t.Error("Output data is empty")
			}

			// Parse it back to ensure it's valid
			parser := cie.NewParser()
			_, err = parser.Parse(outputData)
			if err != nil {
				t.Errorf("Failed to parse output data: %v", err)
			}
		})
	}
}

// Helper function to create test photometric data
func createTestPhotometricData() *models.PhotometricData {
	return &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "TestCorp",
			CatalogNumber: "LED123",
			Description:   "Test LED Fixture",
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
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
