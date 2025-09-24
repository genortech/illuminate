package ies

import (
	"illuminate/internal/models"
	"testing"
)

func TestIESFileToCommonModel(t *testing.T) {
	iesFile := &IESFile{
		Header: IESHeader{
			Version: VersionLM632002,
			Keywords: map[string]string{
				"MANUFAC":   "Test Manufacturer",
				"LUMCAT":    "TEST-001",
				"LUMINAIRE": "Test Luminaire Description",
				"TESTLAB":   "Test Laboratory",
				"ISSUEDATE": "2023-01-01",
				"TEST":      "12345",
			},
			TiltData: "TILT=NONE",
		},
		Photometric: IESPhotometricData{
			NumberOfLamps:       1,
			LumensPerLamp:       1000.0,
			CandelaMultiplier:   1.0,
			NumVerticalAngles:   3,
			NumHorizontalAngles: 2,
			PhotometricType:     1,       // Type C
			UnitsType:           1,       // Feet
			Width:               3.28084, // 1 meter in feet
			Length:              6.56168, // 2 meters in feet
			Height:              9.84252, // 3 meters in feet
			BallastFactor:       1.0,
			BallastLampFactor:   1.0,
			InputWatts:          100.0,
			VerticalAngles:      []float64{0.0, 45.0, 90.0},
			HorizontalAngles:    []float64{0.0, 180.0},
			CandelaValues: [][]float64{
				{100.0, 200.0},
				{150.0, 250.0},
				{200.0, 300.0},
			},
		},
	}

	commonData, err := iesFile.ToCommonModel()
	if err != nil {
		t.Fatalf("ToCommonModel failed: %v", err)
	}

	// Test metadata conversion
	if commonData.Metadata.Manufacturer != "Test Manufacturer" {
		t.Errorf("Expected manufacturer 'Test Manufacturer', got '%s'", commonData.Metadata.Manufacturer)
	}

	if commonData.Metadata.CatalogNumber != "TEST-001" {
		t.Errorf("Expected catalog number 'TEST-001', got '%s'", commonData.Metadata.CatalogNumber)
	}

	if commonData.Metadata.Description != "Test Luminaire Description" {
		t.Errorf("Expected description 'Test Luminaire Description', got '%s'", commonData.Metadata.Description)
	}

	// Test photometry conversion
	if commonData.Photometry.PhotometryType != "C" {
		t.Errorf("Expected photometry type 'C', got '%s'", commonData.Photometry.PhotometryType)
	}

	if commonData.Photometry.LuminousFlux != 1000.0 {
		t.Errorf("Expected luminous flux 1000.0, got %f", commonData.Photometry.LuminousFlux)
	}

	if commonData.Photometry.CandelaMultiplier != 1.0 {
		t.Errorf("Expected candela multiplier 1.0, got %f", commonData.Photometry.CandelaMultiplier)
	}

	// Test geometry conversion (feet to meters)
	expectedWidth := 1.0  // 3.28084 feet = 1 meter
	expectedLength := 2.0 // 6.56168 feet = 2 meters
	expectedHeight := 3.0 // 9.84252 feet = 3 meters

	if abs(commonData.Geometry.Width-expectedWidth) > 0.001 {
		t.Errorf("Expected width %f meters, got %f", expectedWidth, commonData.Geometry.Width)
	}

	if abs(commonData.Geometry.Length-expectedLength) > 0.001 {
		t.Errorf("Expected length %f meters, got %f", expectedLength, commonData.Geometry.Length)
	}

	if abs(commonData.Geometry.Height-expectedHeight) > 0.001 {
		t.Errorf("Expected height %f meters, got %f", expectedHeight, commonData.Geometry.Height)
	}

	// Test electrical data conversion
	if commonData.Electrical.InputWatts != 100.0 {
		t.Errorf("Expected input watts 100.0, got %f", commonData.Electrical.InputWatts)
	}

	if commonData.Electrical.BallastFactor != 1.0 {
		t.Errorf("Expected ballast factor 1.0, got %f", commonData.Electrical.BallastFactor)
	}

	// Test angle arrays
	if len(commonData.Photometry.VerticalAngles) != 3 {
		t.Errorf("Expected 3 vertical angles, got %d", len(commonData.Photometry.VerticalAngles))
	}

	if len(commonData.Photometry.HorizontalAngles) != 2 {
		t.Errorf("Expected 2 horizontal angles, got %d", len(commonData.Photometry.HorizontalAngles))
	}

	// Test candela values matrix
	if len(commonData.Photometry.CandelaValues) != 3 {
		t.Errorf("Expected 3 candela value rows, got %d", len(commonData.Photometry.CandelaValues))
	}

	for i, row := range commonData.Photometry.CandelaValues {
		if len(row) != 2 {
			t.Errorf("Expected 2 candela values in row %d, got %d", i, len(row))
		}
	}
}

func TestIESFileFromCommonModel(t *testing.T) {
	commonData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire Description",
			TestLab:       "Test Laboratory",
			TestDate:      "2023-01-01",
			TestNumber:    "12345",
		},
		Geometry: models.LuminaireGeometry{
			Width:  1.0, // 1 meter
			Length: 2.0, // 2 meters
			Height: 3.0, // 3 meters
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0.0, 45.0, 90.0},
			HorizontalAngles:  []float64{0.0, 180.0},
			CandelaValues: [][]float64{
				{100.0, 200.0},
				{150.0, 250.0},
				{200.0, 300.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:        100.0,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}

	iesFile := &IESFile{}
	err := iesFile.FromCommonModel(commonData)
	if err != nil {
		t.Fatalf("FromCommonModel failed: %v", err)
	}

	// Test header conversion
	if iesFile.Header.Version != VersionLM632002 {
		t.Errorf("Expected version %s, got %s", VersionLM632002, iesFile.Header.Version)
	}

	if iesFile.Header.Keywords["MANUFAC"] != "Test Manufacturer" {
		t.Errorf("Expected manufacturer 'Test Manufacturer', got '%s'", iesFile.Header.Keywords["MANUFAC"])
	}

	if iesFile.Header.Keywords["LUMCAT"] != "TEST-001" {
		t.Errorf("Expected catalog number 'TEST-001', got '%s'", iesFile.Header.Keywords["LUMCAT"])
	}

	// Test photometric data conversion
	if iesFile.Photometric.PhotometricType != 1 {
		t.Errorf("Expected photometric type 1 (Type C), got %d", iesFile.Photometric.PhotometricType)
	}

	if iesFile.Photometric.LumensPerLamp != 1000.0 {
		t.Errorf("Expected lumens per lamp 1000.0, got %f", iesFile.Photometric.LumensPerLamp)
	}

	if iesFile.Photometric.UnitsType != 1 {
		t.Errorf("Expected units type 1 (feet), got %d", iesFile.Photometric.UnitsType)
	}

	// Test geometry conversion (meters to feet)
	const metersToFeet = 3.28084
	expectedWidth := 1.0 * metersToFeet
	expectedLength := 2.0 * metersToFeet
	expectedHeight := 3.0 * metersToFeet

	if abs(iesFile.Photometric.Width-expectedWidth) > 0.001 {
		t.Errorf("Expected width %f feet, got %f", expectedWidth, iesFile.Photometric.Width)
	}

	if abs(iesFile.Photometric.Length-expectedLength) > 0.001 {
		t.Errorf("Expected length %f feet, got %f", expectedLength, iesFile.Photometric.Length)
	}

	if abs(iesFile.Photometric.Height-expectedHeight) > 0.001 {
		t.Errorf("Expected height %f feet, got %f", expectedHeight, iesFile.Photometric.Height)
	}

	// Test angle arrays
	if len(iesFile.Photometric.VerticalAngles) != 3 {
		t.Errorf("Expected 3 vertical angles, got %d", len(iesFile.Photometric.VerticalAngles))
	}

	if len(iesFile.Photometric.HorizontalAngles) != 2 {
		t.Errorf("Expected 2 horizontal angles, got %d", len(iesFile.Photometric.HorizontalAngles))
	}

	// Test candela values matrix
	if len(iesFile.Photometric.CandelaValues) != 3 {
		t.Errorf("Expected 3 candela value rows, got %d", len(iesFile.Photometric.CandelaValues))
	}
}

func TestPhotometryTypeConversion(t *testing.T) {
	tests := []struct {
		name           string
		iesType        int
		expectedCommon string
		commonType     string
		expectedIES    int
	}{
		{"Type A", 3, "A", "A", 3},
		{"Type B", 2, "B", "B", 2},
		{"Type C", 1, "C", "C", 1},
		{"Invalid IES type", 99, "C", "C", 1},   // Should default to C
		{"Invalid common type", 1, "C", "X", 1}, // Should default to C
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test IES to Common conversion
			iesFile := &IESFile{
				Photometric: IESPhotometricData{
					PhotometricType:     tt.iesType,
					NumberOfLamps:       1,
					LumensPerLamp:       1000.0,
					CandelaMultiplier:   1.0,
					NumVerticalAngles:   1,
					NumHorizontalAngles: 1,
					UnitsType:           1,
					VerticalAngles:      []float64{0.0},
					HorizontalAngles:    []float64{0.0},
					CandelaValues:       [][]float64{{100.0}},
				},
				Header: IESHeader{
					Keywords: map[string]string{
						"MANUFAC": "Test",
						"LUMCAT":  "Test",
					},
				},
			}

			commonData, err := iesFile.ToCommonModel()
			if err != nil {
				t.Fatalf("ToCommonModel failed: %v", err)
			}

			if commonData.Photometry.PhotometryType != tt.expectedCommon {
				t.Errorf("Expected common type %s, got %s", tt.expectedCommon, commonData.Photometry.PhotometryType)
			}

			// Test Common to IES conversion
			commonData.Photometry.PhotometryType = tt.commonType
			newIESFile := &IESFile{}
			err = newIESFile.FromCommonModel(commonData)
			if err != nil {
				t.Fatalf("FromCommonModel failed: %v", err)
			}

			if newIESFile.Photometric.PhotometricType != tt.expectedIES {
				t.Errorf("Expected IES type %d, got %d", tt.expectedIES, newIESFile.Photometric.PhotometricType)
			}
		})
	}
}

func TestGetKeywordValue(t *testing.T) {
	keywords := map[string]string{
		"MANUFAC": "Test Manufacturer",
		"LUMCAT":  "TEST-001",
	}

	// Test existing key
	value := getKeywordValue(keywords, "MANUFAC")
	if value != "Test Manufacturer" {
		t.Errorf("Expected 'Test Manufacturer', got '%s'", value)
	}

	// Test non-existing key
	value = getKeywordValue(keywords, "NONEXISTENT")
	if value != "" {
		t.Errorf("Expected empty string for non-existent key, got '%s'", value)
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Valid float", "123.45", 123.45},
		{"Valid integer", "123", 123.0},
		{"With whitespace", "  123.45  ", 123.45},
		{"Invalid string", "abc", 0.0},
		{"Empty string", "", 0.0},
		{"Negative number", "-123.45", -123.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFloat(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Valid integer", "123", 123},
		{"With whitespace", "  123  ", 123},
		{"Invalid string", "abc", 0},
		{"Empty string", "", 0},
		{"Negative number", "-123", -123},
		{"Float string", "123.45", 0}, // Should fail for non-integer
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInt(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestToCommonModelWithMissingKeywords(t *testing.T) {
	iesFile := &IESFile{
		Header: IESHeader{
			Version:  VersionLM632002,
			Keywords: map[string]string{}, // Empty keywords
			TiltData: "TILT=NONE",
		},
		Photometric: IESPhotometricData{
			NumberOfLamps:       1,
			LumensPerLamp:       1000.0,
			CandelaMultiplier:   1.0,
			NumVerticalAngles:   1,
			NumHorizontalAngles: 1,
			PhotometricType:     1,
			UnitsType:           1,
			VerticalAngles:      []float64{0.0},
			HorizontalAngles:    []float64{0.0},
			CandelaValues:       [][]float64{{100.0}},
		},
	}

	commonData, err := iesFile.ToCommonModel()
	if err != nil {
		t.Fatalf("ToCommonModel failed: %v", err)
	}

	// Should use default values for missing keywords
	if commonData.Metadata.Manufacturer != "Unknown" {
		t.Errorf("Expected default manufacturer 'Unknown', got '%s'", commonData.Metadata.Manufacturer)
	}

	if commonData.Metadata.CatalogNumber != "Unknown" {
		t.Errorf("Expected default catalog number 'Unknown', got '%s'", commonData.Metadata.CatalogNumber)
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
