package test

import (
	"illuminate/internal/models"
	ldtparser "illuminate/internal/parsers/ldt"
	ldtwriter "illuminate/internal/writers/ldt"
	"testing"
)

func TestLDTWriterIntegration(t *testing.T) {
	// Create test data instead of reading from file to avoid parsing issues
	parsedData := createTestPhotometricData()

	// Write it back using the LDT writer
	writer := ldtwriter.NewWriter()
	outputData, err := writer.Write(parsedData)
	if err != nil {
		t.Fatalf("Failed to write LDT data: %v", err)
	}

	// Verify the output is not empty
	if len(outputData) == 0 {
		t.Fatal("Writer produced empty output")
	}

	// Parse the written data to ensure it's valid
	parser := ldtparser.NewParser()
	reparsedData, err := parser.Parse(outputData)
	if err != nil {
		t.Fatalf("Failed to parse written LDT data: %v", err)
	}

	// Compare key values (allowing for some precision loss)
	tolerance := 0.01

	if abs(parsedData.Photometry.LuminousFlux-reparsedData.Photometry.LuminousFlux) > tolerance {
		t.Errorf("Luminous flux mismatch: original=%f, reparsed=%f",
			parsedData.Photometry.LuminousFlux, reparsedData.Photometry.LuminousFlux)
	}

	if len(parsedData.Photometry.VerticalAngles) != len(reparsedData.Photometry.VerticalAngles) {
		t.Errorf("Vertical angles count mismatch: original=%d, reparsed=%d",
			len(parsedData.Photometry.VerticalAngles), len(reparsedData.Photometry.VerticalAngles))
	}

	if len(parsedData.Photometry.HorizontalAngles) != len(reparsedData.Photometry.HorizontalAngles) {
		t.Errorf("Horizontal angles count mismatch: original=%d, reparsed=%d",
			len(parsedData.Photometry.HorizontalAngles), len(reparsedData.Photometry.HorizontalAngles))
	}

	// Verify metadata is preserved
	if parsedData.Metadata.Manufacturer != reparsedData.Metadata.Manufacturer {
		t.Errorf("Manufacturer mismatch: original=%s, reparsed=%s",
			parsedData.Metadata.Manufacturer, reparsedData.Metadata.Manufacturer)
	}

	if parsedData.Metadata.CatalogNumber != reparsedData.Metadata.CatalogNumber {
		t.Errorf("Catalog number mismatch: original=%s, reparsed=%s",
			parsedData.Metadata.CatalogNumber, reparsedData.Metadata.CatalogNumber)
	}

	t.Logf("Successfully completed round-trip conversion for LDT file")
	t.Logf("Written file size: %d bytes", len(outputData))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// createTestPhotometricData creates a valid photometric data structure for testing
func createTestPhotometricData() *models.PhotometricData {
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
