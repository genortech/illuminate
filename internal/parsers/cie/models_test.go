package cie

import (
	"illuminate/internal/models"
	"testing"
)

func TestCIEFile_ToCommonModel(t *testing.T) {
	// Create a test CIE file
	cieFile := &CIEFile{
		Header: CIEHeader{
			FormatType:   1,
			SymmetryType: 0,
			Reserved:     0,
			Description:  "Test LED 17W 1000 lm",
		},
		Photometry: CIEPhotometry{
			GammaAngles:  generateStandardGammaAngles(),
			CPlaneAngles: generateStandardCPlaneAngles(),
			IntensityData: func() [][]float64 {
				// Create test data: 19x16 matrix with decreasing values
				data := make([][]float64, 19)
				for i := range data {
					data[i] = make([]float64, 16)
					for j := range data[i] {
						// Values in cd/1000lm, decreasing with angle
						data[i][j] = float64(100 - i*5)
						if data[i][j] < 0 {
							data[i][j] = 0
						}
					}
				}
				return data
			}(),
		},
	}

	result, err := cieFile.ToCommonModel()
	if err != nil {
		t.Fatalf("ToCommonModel() error = %v", err)
	}

	// Test metadata
	if result.Metadata.Manufacturer == "" {
		t.Error("Manufacturer should not be empty")
	}
	if result.Metadata.CatalogNumber == "" {
		t.Error("CatalogNumber should not be empty")
	}
	if result.Metadata.Description != "Test LED 17W 1000 lm" {
		t.Errorf("Description = %v, want %v", result.Metadata.Description, "Test LED 17W 1000 lm")
	}

	// Test geometry (should have default values)
	if result.Geometry.Length <= 0 {
		t.Error("Length should be positive")
	}
	if result.Geometry.Width <= 0 {
		t.Error("Width should be positive")
	}

	// Test photometry
	if result.Photometry.PhotometryType != "C" {
		t.Errorf("PhotometryType = %v, want C", result.Photometry.PhotometryType)
	}
	if result.Photometry.UnitsType != "absolute" {
		t.Errorf("UnitsType = %v, want absolute", result.Photometry.UnitsType)
	}
	if result.Photometry.LuminousFlux != 1000.0 {
		t.Errorf("LuminousFlux = %v, want 1000.0", result.Photometry.LuminousFlux)
	}

	// Test angles
	if len(result.Photometry.VerticalAngles) != 19 {
		t.Errorf("VerticalAngles length = %v, want 19", len(result.Photometry.VerticalAngles))
	}
	if len(result.Photometry.HorizontalAngles) != 16 {
		t.Errorf("HorizontalAngles length = %v, want 16", len(result.Photometry.HorizontalAngles))
	}

	// Test candela values conversion (should be divided by 1000)
	if len(result.Photometry.CandelaValues) != 19 {
		t.Errorf("CandelaValues rows = %v, want 19", len(result.Photometry.CandelaValues))
	}
	if len(result.Photometry.CandelaValues[0]) != 16 {
		t.Errorf("CandelaValues columns = %v, want 16", len(result.Photometry.CandelaValues[0]))
	}

	// Check that values were converted from cd/1000lm to cd/lm
	expectedFirstValue := 100.0 / 1000.0 // 100 cd/1000lm -> 0.1 cd/lm
	if result.Photometry.CandelaValues[0][0] != expectedFirstValue {
		t.Errorf("First candela value = %v, want %v", result.Photometry.CandelaValues[0][0], expectedFirstValue)
	}

	// Test electrical data
	if result.Electrical.InputWatts != 17.0 {
		t.Errorf("InputWatts = %v, want 17.0", result.Electrical.InputWatts)
	}
}

func TestCIEFile_FromCommonModel(t *testing.T) {
	// Create test common model data
	commonData := &models.PhotometricData{
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
			LuminousFlux:      1500.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0, 10, 20, 30, 45, 60, 90},
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
			InputWatts: 25.0,
		},
	}

	cieFile := &CIEFile{}
	err := cieFile.FromCommonModel(commonData)
	if err != nil {
		t.Fatalf("FromCommonModel() error = %v", err)
	}

	// Test header
	if cieFile.Header.FormatType != 1 {
		t.Errorf("FormatType = %v, want 1", cieFile.Header.FormatType)
	}
	if cieFile.Header.Reserved != 0 {
		t.Errorf("Reserved = %v, want 0", cieFile.Header.Reserved)
	}
	if cieFile.Header.Description == "" {
		t.Error("Description should not be empty")
	}

	// Test photometry angles
	if len(cieFile.Photometry.GammaAngles) != 19 {
		t.Errorf("GammaAngles length = %v, want 19", len(cieFile.Photometry.GammaAngles))
	}
	if len(cieFile.Photometry.CPlaneAngles) != 16 {
		t.Errorf("CPlaneAngles length = %v, want 16", len(cieFile.Photometry.CPlaneAngles))
	}

	// Test intensity data dimensions
	if len(cieFile.Photometry.IntensityData) != 19 {
		t.Errorf("IntensityData rows = %v, want 19", len(cieFile.Photometry.IntensityData))
	}
	for i, row := range cieFile.Photometry.IntensityData {
		if len(row) != 16 {
			t.Errorf("IntensityData row %d length = %v, want 16", i, len(row))
		}
	}

	// Test that values were converted from cd/lm to cd/1000lm
	// The interpolation should give us some non-zero values
	hasNonZeroValues := false
	for _, row := range cieFile.Photometry.IntensityData {
		for _, value := range row {
			if value > 0 {
				hasNonZeroValues = true
				break
			}
		}
		if hasNonZeroValues {
			break
		}
	}
	if !hasNonZeroValues {
		t.Error("Expected some non-zero intensity values after conversion")
	}
}

func TestExtractManufacturer(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "Standard format",
			description: "OSL0526 PLED II 17W AE 3000K 2172.2 lms",
			expected:    "OSL0526",
		},
		{
			name:        "Simple format",
			description: "Philips LED 25W",
			expected:    "Philips",
		},
		{
			name:        "Single word",
			description: "TestLED",
			expected:    "TestLED",
		},
		{
			name:        "Empty description",
			description: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractManufacturer(tt.description)
			if result != tt.expected {
				t.Errorf("extractManufacturer(%q) = %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func TestExtractCatalogNumber(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "With W suffix",
			description: "OSL0526 PLED II 17W AE 3000K 2172.2 lms",
			expected:    "17W",
		},
		{
			name:        "With LED",
			description: "Philips LED25 1000 lm",
			expected:    "LED25",
		},
		{
			name:        "Second word fallback",
			description: "Philips Model123 25W",
			expected:    "25W",
		},
		{
			name:        "No suitable pattern",
			description: "Simple Description",
			expected:    "Description",
		},
		{
			name:        "Single word",
			description: "TestLED",
			expected:    "TestLED",
		},
		{
			name:        "Empty description",
			description: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCatalogNumber(tt.description)
			if result != tt.expected {
				t.Errorf("extractCatalogNumber(%q) = %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func TestFormatDescription(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		lumens   float64
		watts    float64
		expected string
	}{
		{
			name:     "Full description",
			desc:     "LED Fixture",
			lumens:   1000.0,
			watts:    17.0,
			expected: "LED Fixture 17W 1000.0 lm",
		},
		{
			name:     "No watts",
			desc:     "LED Fixture",
			lumens:   1000.0,
			watts:    0.0,
			expected: "LED Fixture 1000.0 lm",
		},
		{
			name:     "No lumens",
			desc:     "LED Fixture",
			lumens:   0.0,
			watts:    17.0,
			expected: "LED Fixture 17W",
		},
		{
			name:     "No additional info",
			desc:     "LED Fixture",
			lumens:   0.0,
			watts:    0.0,
			expected: "LED Fixture",
		},
		{
			name:     "Empty description",
			desc:     "",
			lumens:   1000.0,
			watts:    17.0,
			expected: "LED Luminaire 17W 1000.0 lm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDescription(tt.desc, tt.lumens, tt.watts)
			if result != tt.expected {
				t.Errorf("formatDescription(%q, %v, %v) = %v, want %v",
					tt.desc, tt.lumens, tt.watts, result, tt.expected)
			}
		})
	}
}

func TestDetermineSymmetryType(t *testing.T) {
	tests := []struct {
		name             string
		horizontalAngles []float64
		expectedSymmetry int
	}{
		{
			name:             "Quarter symmetry",
			horizontalAngles: []float64{0, 22.5, 45, 67.5, 90},
			expectedSymmetry: 1,
		},
		{
			name:             "Half symmetry",
			horizontalAngles: []float64{0, 30, 60, 90, 120, 150, 180},
			expectedSymmetry: 0,
		},
		{
			name:             "Full symmetry",
			horizontalAngles: []float64{0, 45, 90, 135, 180, 225, 270, 315},
			expectedSymmetry: 0,
		},
		{
			name:             "Single angle",
			horizontalAngles: []float64{0},
			expectedSymmetry: 1,
		},
		{
			name:             "Empty angles",
			horizontalAngles: []float64{},
			expectedSymmetry: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineSymmetryType(tt.horizontalAngles)
			if result != tt.expectedSymmetry {
				t.Errorf("determineSymmetryType(%v) = %v, want %v",
					tt.horizontalAngles, result, tt.expectedSymmetry)
			}
		})
	}
}

func TestInterpolateValue(t *testing.T) {
	// Create simple test data
	inputGamma := []float64{0, 30, 60, 90}
	inputCPlane := []float64{0, 90, 180, 270}
	inputData := [][]float64{
		{1.0, 0.8, 0.6, 0.8}, // 0°
		{0.8, 0.6, 0.4, 0.6}, // 30°
		{0.4, 0.3, 0.2, 0.3}, // 60°
		{0.1, 0.1, 0.1, 0.1}, // 90°
	}

	tests := []struct {
		name         string
		targetGamma  float64
		targetCPlane float64
		expected     float64
	}{
		{
			name:         "Exact match",
			targetGamma:  0,
			targetCPlane: 0,
			expected:     1.0,
		},
		{
			name:         "Nearest neighbor gamma",
			targetGamma:  10, // Closer to 0° than 30°
			targetCPlane: 0,
			expected:     1.0,
		},
		{
			name:         "Nearest neighbor cplane",
			targetGamma:  0,
			targetCPlane: 45, // Closer to 0° than 90°
			expected:     1.0,
		},
		{
			name:         "Wraparound test",
			targetGamma:  0,
			targetCPlane: 350, // Should be close to 0° (10° difference)
			expected:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolateValue(inputGamma, inputCPlane, inputData, tt.targetGamma, tt.targetCPlane)
			if result != tt.expected {
				t.Errorf("interpolateValue(gamma=%v, cplane=%v) = %v, want %v",
					tt.targetGamma, tt.targetCPlane, result, tt.expected)
			}
		})
	}
}

func TestInterpolateValue_EdgeCases(t *testing.T) {
	t.Run("Empty data", func(t *testing.T) {
		result := interpolateValue([]float64{}, []float64{}, [][]float64{}, 0, 0)
		if result != 0.0 {
			t.Errorf("Expected 0.0 for empty data, got %v", result)
		}
	})

	t.Run("Single point", func(t *testing.T) {
		inputGamma := []float64{0}
		inputCPlane := []float64{0}
		inputData := [][]float64{{5.0}}

		result := interpolateValue(inputGamma, inputCPlane, inputData, 45, 180)
		if result != 5.0 {
			t.Errorf("Expected 5.0 for single point, got %v", result)
		}
	})
}
