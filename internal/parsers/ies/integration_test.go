package ies

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRealIESFile(t *testing.T) {
	parser := NewParser()

	// Test with the sample IES file
	samplePath := filepath.Join("..", "..", "..", "tests", "samples", "P14W.N7P.3K.AG.ies")

	data, err := os.ReadFile(samplePath)
	if err != nil {
		t.Skipf("Sample file not found: %v", err)
		return
	}

	// Test format detection
	confidence, detectedVersion := parser.DetectFormat(data)
	if confidence < 0.9 {
		t.Errorf("Expected high confidence for real IES file, got %f", confidence)
	}

	if detectedVersion == "" {
		t.Error("Expected version to be detected")
	}

	// Test parsing
	commonData, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse real IES file: %v", err)
	}

	if commonData == nil {
		t.Fatal("Parse returned nil data")
	}

	// Test validation
	err = parser.Validate(commonData)
	if err != nil {
		t.Errorf("Validation failed for real IES file: %v", err)
	}

	// Verify some basic properties
	if commonData.Metadata.Manufacturer == "" {
		t.Error("Expected manufacturer to be parsed")
	}

	if commonData.Photometry.PhotometryType == "" {
		t.Error("Expected photometry type to be set")
	}

	if len(commonData.Photometry.VerticalAngles) == 0 {
		t.Error("Expected vertical angles to be parsed")
	}

	if len(commonData.Photometry.HorizontalAngles) == 0 {
		t.Error("Expected horizontal angles to be parsed")
	}

	if len(commonData.Photometry.CandelaValues) == 0 {
		t.Error("Expected candela values to be parsed")
	}

	// Verify matrix dimensions match
	if len(commonData.Photometry.CandelaValues) != len(commonData.Photometry.VerticalAngles) {
		t.Errorf("Candela matrix rows (%d) don't match vertical angles (%d)",
			len(commonData.Photometry.CandelaValues), len(commonData.Photometry.VerticalAngles))
	}

	for i, row := range commonData.Photometry.CandelaValues {
		if len(row) != len(commonData.Photometry.HorizontalAngles) {
			t.Errorf("Candela matrix row %d length (%d) doesn't match horizontal angles (%d)",
				i, len(row), len(commonData.Photometry.HorizontalAngles))
		}
	}

	t.Logf("Successfully parsed IES file:")
	t.Logf("  Manufacturer: %s", commonData.Metadata.Manufacturer)
	t.Logf("  Catalog: %s", commonData.Metadata.CatalogNumber)
	t.Logf("  Photometry Type: %s", commonData.Photometry.PhotometryType)
	t.Logf("  Luminous Flux: %.1f lm", commonData.Photometry.LuminousFlux)
	t.Logf("  Vertical Angles: %d", len(commonData.Photometry.VerticalAngles))
	t.Logf("  Horizontal Angles: %d", len(commonData.Photometry.HorizontalAngles))
	t.Logf("  Input Watts: %.1f W", commonData.Electrical.InputWatts)
}

func TestParseMultipleSampleFiles(t *testing.T) {
	parser := NewParser()

	samplesDir := filepath.Join("..", "..", "..", "tests", "samples")

	// Find all IES files in the samples directory
	files, err := filepath.Glob(filepath.Join(samplesDir, "*.ies"))
	if err != nil {
		t.Skipf("Could not read samples directory: %v", err)
		return
	}

	if len(files) == 0 {
		t.Skip("No IES sample files found")
		return
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", file, err)
			}

			// Test format detection
			confidence, _ := parser.DetectFormat(data)
			if confidence < 0.8 {
				t.Errorf("Low confidence (%f) for IES file %s", confidence, file)
			}

			// Test parsing
			commonData, err := parser.Parse(data)
			if err != nil {
				t.Errorf("Failed to parse %s: %v", file, err)
				return
			}

			// Test validation
			err = parser.Validate(commonData)
			if err != nil {
				t.Errorf("Validation failed for %s: %v", file, err)
			}

			t.Logf("File: %s - Manufacturer: %s, Type: %s, Flux: %.1f lm",
				filepath.Base(file),
				commonData.Metadata.Manufacturer,
				commonData.Photometry.PhotometryType,
				commonData.Photometry.LuminousFlux)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	parser := NewParser()

	samplePath := filepath.Join("..", "..", "..", "tests", "samples", "P14W.N7P.3K.AG.ies")

	data, err := os.ReadFile(samplePath)
	if err != nil {
		t.Skipf("Sample file not found: %v", err)
		return
	}

	// Parse to common model
	commonData1, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Failed to parse IES file: %v", err)
	}

	// Convert back to IES format
	iesFile := &IESFile{}
	err = iesFile.FromCommonModel(commonData1)
	if err != nil {
		t.Fatalf("Failed to convert to IES format: %v", err)
	}

	// Convert back to common model
	commonData2, err := iesFile.ToCommonModel()
	if err != nil {
		t.Fatalf("Failed to convert back to common model: %v", err)
	}

	// Compare key values (allowing for some precision loss)
	if commonData1.Metadata.Manufacturer != commonData2.Metadata.Manufacturer {
		t.Errorf("Manufacturer mismatch: %s vs %s",
			commonData1.Metadata.Manufacturer, commonData2.Metadata.Manufacturer)
	}

	if commonData1.Photometry.PhotometryType != commonData2.Photometry.PhotometryType {
		t.Errorf("Photometry type mismatch: %s vs %s",
			commonData1.Photometry.PhotometryType, commonData2.Photometry.PhotometryType)
	}

	// Allow for small differences in luminous flux due to conversion
	fluxDiff := absFloat(commonData1.Photometry.LuminousFlux - commonData2.Photometry.LuminousFlux)
	if fluxDiff > 0.1 {
		t.Errorf("Luminous flux difference too large: %f vs %f (diff: %f)",
			commonData1.Photometry.LuminousFlux, commonData2.Photometry.LuminousFlux, fluxDiff)
	}

	t.Logf("Round-trip conversion successful")
}

// Helper function for floating point comparison
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
