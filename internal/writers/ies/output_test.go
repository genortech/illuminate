package ies

import (
	"illuminate/internal/models"
	"strings"
	"testing"
)

func TestWriterOutputFormat(t *testing.T) {
	writer := NewWriter()

	// Create simple test data
	data := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
			TestLab:       "Test Lab",
			TestDate:      "2023-01-01",
			TestNumber:    "12345",
		},
		Geometry: models.LuminaireGeometry{
			Length:         1.64,
			Width:          0.49,
			Height:         0.00,
			LuminousLength: 1.64,
			LuminousWidth:  0.49,
			LuminousHeight: 0.00,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      5800.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0.0, 45.0, 90.0},
			HorizontalAngles:  []float64{0.0, 90.0, 180.0},
			CandelaValues: [][]float64{
				{1337.5, 1000.0, 500.0},
				{1200.0, 900.0, 400.0},
				{800.0, 600.0, 200.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:        44.5,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}

	output, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content := string(output)
	t.Logf("Generated IES content:\n%s", content)

	lines := strings.Split(content, "\n")

	// Verify basic structure elements are present
	foundVersion := false
	foundManufac := false
	foundLumcat := false
	foundTilt := false
	foundPhotometricLine1 := false
	foundPhotometricLine2 := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "IESNA:LM-63-2002") {
			foundVersion = true
		}
		if strings.Contains(line, "[MANUFAC] Test Manufacturer") {
			foundManufac = true
		}
		if strings.Contains(line, "[LUMCAT] TEST-001") {
			foundLumcat = true
		}
		if line == "TILT=NONE" {
			foundTilt = true
		}

		// Check for photometric data lines
		fields := strings.Fields(line)
		if len(fields) == 10 && fields[0] == "1" && fields[1] == "5800.0" {
			foundPhotometricLine1 = true
		}
		if len(fields) == 3 && fields[0] == "1.0" && fields[1] == "1.0" && fields[2] == "44.5" {
			foundPhotometricLine2 = true
		}
	}

	if !foundVersion {
		t.Error("Version header not found")
	}
	if !foundManufac {
		t.Error("MANUFAC keyword not found")
	}
	if !foundLumcat {
		t.Error("LUMCAT keyword not found")
	}
	if !foundTilt {
		t.Error("TILT line not found")
	}
	if !foundPhotometricLine1 {
		t.Error("First photometric data line not found")
	}
	if !foundPhotometricLine2 {
		t.Error("Second photometric data line not found")
	}

	// Verify angles are present
	foundVerticalAngles := false
	foundHorizontalAngles := false
	foundCandelaValues := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "0.0 45.0 90.0") {
			foundVerticalAngles = true
		}
		if strings.Contains(line, "0.0 90.0 180.0") {
			foundHorizontalAngles = true
		}
		if strings.Contains(line, "1337.5") {
			foundCandelaValues = true
		}
	}

	if !foundVerticalAngles {
		t.Error("Vertical angles not found in output")
	}
	if !foundHorizontalAngles {
		t.Error("Horizontal angles not found in output")
	}
	if !foundCandelaValues {
		t.Error("Candela values not found in output")
	}
}

func TestWriterCompliance(t *testing.T) {
	writer := NewWriter()

	// Test with data that matches IES standards
	data := createTestData(
		[]float64{0.0, 10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0},
		[]float64{0.0, 22.5, 45.0, 67.5, 90.0, 112.5, 135.0, 157.5, 180.0},
		generateCandelaMatrix(10, 9),
	)

	output, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	// Check IES compliance requirements

	// 1. Must start with IESNA version
	if !strings.HasPrefix(lines[0], "IESNA") {
		t.Error("File must start with IESNA version identifier")
	}

	// 2. Must have TILT line
	foundTilt := false
	for _, line := range lines {
		if strings.HasPrefix(line, "TILT=") {
			foundTilt = true
			break
		}
	}
	if !foundTilt {
		t.Error("File must contain TILT line")
	}

	// 3. Must have required keywords
	requiredKeywords := []string{"MANUFAC", "LUMCAT"}
	for _, keyword := range requiredKeywords {
		found := false
		for _, line := range lines {
			if strings.Contains(line, "["+keyword+"]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing required keyword: %s", keyword)
		}
	}

	// 4. Check photometric data format
	foundPhotometricData := false
	for _, line := range lines {
		// Look for the first photometric data line (10 values)
		fields := strings.Fields(line)
		if len(fields) == 10 {
			// This should be the first photometric line
			foundPhotometricData = true

			// Verify it contains expected number of angles
			if fields[3] != "10" || fields[4] != "9" {
				t.Errorf("Incorrect angle counts in photometric data: got %s vertical, %s horizontal", fields[3], fields[4])
			}
			break
		}
	}
	if !foundPhotometricData {
		t.Error("Missing photometric data section")
	}
}

// Helper function to generate a candela matrix
func generateCandelaMatrix(verticalCount, horizontalCount int) [][]float64 {
	matrix := make([][]float64, verticalCount)
	for i := range matrix {
		matrix[i] = make([]float64, horizontalCount)
		for j := range matrix[i] {
			// Generate some realistic candela values
			matrix[i][j] = float64(1000 - i*50 - j*20)
			if matrix[i][j] < 0 {
				matrix[i][j] = 0
			}
		}
	}
	return matrix
}
