package ldt

import (
	"illuminate/internal/models"
	"testing"
)

func TestLDTFile_ToCommonModel(t *testing.T) {
	// Create a sample LDT file
	ldtFile := &LDTFile{
		Header: LDTHeader{
			CompanyIdentification:              "WE-EF;Eulumdat2",
			TypeIndicator:                      2,
			SymmetryIndicator:                  3,
			NumberOfCPlanes:                    4,
			DistanceBetweenCPlanes:             90.0,
			NumberOfLuminousIntensities:        3,
			DistanceBetweenLuminousIntensities: 30.0,
			MeasurementReport:                  "Test measurement report",
			LuminaireName:                      "Test Luminaire",
			LuminaireNumber:                    "TEST-001",
			FileName:                           "test.ldt",
			DateUser:                           "01 Jan 2023/TestLab",
		},
		Geometry: LDTGeometry{
			LengthOfLuminaire:         600.0, // mm
			WidthOfLuminaire:          250.0, // mm
			HeightOfLuminaire:         190.0, // mm
			LengthOfLuminousArea:      180.0, // mm
			WidthOfLuminousArea:       160.0, // mm
			HeightOfLuminousAreaC0:    10.0,  // mm
			HeightOfLuminousAreaC90:   10.0,  // mm
			HeightOfLuminousAreaC180:  10.0,  // mm
			HeightOfLuminousAreaC270:  10.0,  // mm
			DownwardFluxFraction:      100.0, // %
			LightOutputRatioLuminaire: 90.0,  // %
			ConversionFactor:          1.0,
		},
		Electrical: LDTElectrical{
			DRIndex:          0,
			NumberOfLampSets: 1,
			LampSets: []LampSet{
				{
					NumberOfLamps:           1,
					Type:                    "LED",
					TotalLuminousFlux:       1000.0,
					ColorTemperature:        "3000K",
					ColorRenderingGroup:     "80",
					WattageIncludingBallast: 10.0,
				},
			},
		},
		Photometry: LDTPhotometry{
			CPlaneAngles: []float64{0.0, 90.0, 180.0, 270.0},
			GammaAngles:  []float64{0.0, 30.0, 90.0},
			LuminousIntensityDistribution: [][]float64{
				{100.0, 90.0, 80.0, 70.0},
				{95.0, 85.0, 75.0, 65.0},
				{50.0, 40.0, 30.0, 20.0},
			},
		},
	}

	// Convert to common model
	commonData, err := ldtFile.ToCommonModel()
	if err != nil {
		t.Fatalf("ToCommonModel failed: %v", err)
	}

	// Verify metadata conversion
	if commonData.Metadata.Manufacturer != "WE-EF" {
		t.Errorf("Expected manufacturer 'WE-EF', got '%s'", commonData.Metadata.Manufacturer)
	}
	if commonData.Metadata.CatalogNumber != "TEST-001" {
		t.Errorf("Expected catalog number 'TEST-001', got '%s'", commonData.Metadata.CatalogNumber)
	}
	if commonData.Metadata.Description != "Test Luminaire" {
		t.Errorf("Expected description 'Test Luminaire', got '%s'", commonData.Metadata.Description)
	}

	// Verify geometry conversion (mm to meters)
	expectedLength := 0.6 // 600mm to meters
	if commonData.Geometry.Length != expectedLength {
		t.Errorf("Expected length %f, got %f", expectedLength, commonData.Geometry.Length)
	}
	expectedWidth := 0.25 // 250mm to meters
	if commonData.Geometry.Width != expectedWidth {
		t.Errorf("Expected width %f, got %f", expectedWidth, commonData.Geometry.Width)
	}

	// Verify photometry conversion
	if commonData.Photometry.PhotometryType != "C" {
		t.Errorf("Expected photometry type 'C', got '%s'", commonData.Photometry.PhotometryType)
	}
	if commonData.Photometry.LuminousFlux != 1000.0 {
		t.Errorf("Expected luminous flux 1000.0, got %f", commonData.Photometry.LuminousFlux)
	}
	if len(commonData.Photometry.VerticalAngles) != 3 {
		t.Errorf("Expected 3 vertical angles, got %d", len(commonData.Photometry.VerticalAngles))
	}
	if len(commonData.Photometry.HorizontalAngles) != 4 {
		t.Errorf("Expected 4 horizontal angles, got %d", len(commonData.Photometry.HorizontalAngles))
	}

	// Verify electrical conversion
	if commonData.Electrical.InputWatts != 10.0 {
		t.Errorf("Expected input watts 10.0, got %f", commonData.Electrical.InputWatts)
	}
}

func TestLDTFile_FromCommonModel(t *testing.T) {
	// Create common model data
	commonData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
			TestLab:       "TestLab",
			TestDate:      "01 Jan 2023",
			TestNumber:    "test.ldt",
		},
		Geometry: models.LuminaireGeometry{
			Length:         0.6,  // meters
			Width:          0.25, // meters
			Height:         0.19, // meters
			LuminousLength: 0.18, // meters
			LuminousWidth:  0.16, // meters
			LuminousHeight: 0.01, // meters
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0.0, 30.0, 60.0, 90.0},
			HorizontalAngles:  []float64{0.0, 90.0, 180.0, 270.0},
			CandelaValues: [][]float64{
				{100.0, 90.0, 80.0, 70.0},
				{95.0, 85.0, 75.0, 65.0},
				{80.0, 70.0, 60.0, 50.0},
				{50.0, 40.0, 30.0, 20.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:        10.0,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}

	// Convert from common model
	ldtFile := &LDTFile{}
	err := ldtFile.FromCommonModel(commonData)
	if err != nil {
		t.Fatalf("FromCommonModel failed: %v", err)
	}

	// Verify header conversion
	if ldtFile.Header.CompanyIdentification != "Test Manufacturer" {
		t.Errorf("Expected company identification 'Test Manufacturer', got '%s'", ldtFile.Header.CompanyIdentification)
	}
	if ldtFile.Header.LuminaireNumber != "TEST-001" {
		t.Errorf("Expected luminaire number 'TEST-001', got '%s'", ldtFile.Header.LuminaireNumber)
	}
	if ldtFile.Header.NumberOfCPlanes != 4 {
		t.Errorf("Expected 4 C planes, got %d", ldtFile.Header.NumberOfCPlanes)
	}
	if ldtFile.Header.NumberOfLuminousIntensities != 4 {
		t.Errorf("Expected 4 luminous intensities, got %d", ldtFile.Header.NumberOfLuminousIntensities)
	}

	// Verify geometry conversion (meters to mm)
	expectedLength := 600.0 // 0.6m to mm
	if ldtFile.Geometry.LengthOfLuminaire != expectedLength {
		t.Errorf("Expected length %f, got %f", expectedLength, ldtFile.Geometry.LengthOfLuminaire)
	}
	expectedWidth := 250.0 // 0.25m to mm
	if ldtFile.Geometry.WidthOfLuminaire != expectedWidth {
		t.Errorf("Expected width %f, got %f", expectedWidth, ldtFile.Geometry.WidthOfLuminaire)
	}

	// Verify electrical conversion
	if ldtFile.Electrical.NumberOfLampSets != 1 {
		t.Errorf("Expected 1 lamp set, got %d", ldtFile.Electrical.NumberOfLampSets)
	}
	if len(ldtFile.Electrical.LampSets) != 1 {
		t.Errorf("Expected 1 lamp set in array, got %d", len(ldtFile.Electrical.LampSets))
	}
	if ldtFile.Electrical.LampSets[0].TotalLuminousFlux != 1000.0 {
		t.Errorf("Expected flux 1000.0, got %f", ldtFile.Electrical.LampSets[0].TotalLuminousFlux)
	}
	if ldtFile.Electrical.LampSets[0].WattageIncludingBallast != 10.0 {
		t.Errorf("Expected wattage 10.0, got %f", ldtFile.Electrical.LampSets[0].WattageIncludingBallast)
	}

	// Verify photometry conversion
	if len(ldtFile.Photometry.CPlaneAngles) != 4 {
		t.Errorf("Expected 4 C-plane angles, got %d", len(ldtFile.Photometry.CPlaneAngles))
	}
	if len(ldtFile.Photometry.GammaAngles) != 4 {
		t.Errorf("Expected 4 gamma angles, got %d", len(ldtFile.Photometry.GammaAngles))
	}
}

func TestExtractManufacturer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"WE-EF;Eulumdat2", "WE-EF"},
		{"Company Name;Additional Info", "Company Name"},
		{"Single Company", "Single Company"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractManufacturer(tt.input)
			if result != tt.expected {
				t.Errorf("extractManufacturer(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractTestLab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01 Jan 2023/TestLab", "TestLab"},
		{"Date/Lab Name", "Lab Name"},
		{"No slash", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractTestLab(tt.input)
			if result != tt.expected {
				t.Errorf("extractTestLab(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractTestDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01 Jan 2023/TestLab", "01 Jan 2023"},
		{"Date/Lab Name", "Date"},
		{"No slash", "No slash"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractTestDate(tt.input)
			if result != tt.expected {
				t.Errorf("extractTestDate(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDateUser(t *testing.T) {
	tests := []struct {
		date     string
		user     string
		expected string
	}{
		{"01 Jan 2023", "TestLab", "01 Jan 2023/TestLab"},
		{"", "TestLab", "/TestLab"},
		{"01 Jan 2023", "", "01 Jan 2023"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.date+"_"+tt.user, func(t *testing.T) {
			result := formatDateUser(tt.date, tt.user)
			if result != tt.expected {
				t.Errorf("formatDateUser(%q, %q) = %q, expected %q", tt.date, tt.user, result, tt.expected)
			}
		})
	}
}

func TestCalculateAngleIncrement(t *testing.T) {
	tests := []struct {
		name     string
		angles   []float64
		expected float64
	}{
		{
			name:     "Regular increments",
			angles:   []float64{0.0, 30.0, 60.0, 90.0},
			expected: 30.0,
		},
		{
			name:     "Irregular increments",
			angles:   []float64{0.0, 10.0, 30.0, 60.0},
			expected: 20.0, // Average of 10, 20, 30
		},
		{
			name:     "Single angle",
			angles:   []float64{0.0},
			expected: 1.0, // Default
		},
		{
			name:     "Empty array",
			angles:   []float64{},
			expected: 1.0, // Default
		},
		{
			name:     "Two angles",
			angles:   []float64{0.0, 45.0},
			expected: 45.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAngleIncrement(tt.angles)
			if result != tt.expected {
				t.Errorf("calculateAngleIncrement(%v) = %f, expected %f", tt.angles, result, tt.expected)
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Create original common data
	originalData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
			TestLab:       "TestLab",
			TestDate:      "01 Jan 2023",
			TestNumber:    "test.ldt",
		},
		Geometry: models.LuminaireGeometry{
			Length:         0.6,
			Width:          0.25,
			Height:         0.19,
			LuminousLength: 0.18,
			LuminousWidth:  0.16,
			LuminousHeight: 0.01,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0.0, 30.0, 60.0, 90.0},
			HorizontalAngles:  []float64{0.0, 90.0, 180.0, 270.0},
			CandelaValues: [][]float64{
				{100.0, 90.0, 80.0, 70.0},
				{95.0, 85.0, 75.0, 65.0},
				{80.0, 70.0, 60.0, 50.0},
				{50.0, 40.0, 30.0, 20.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:        10.0,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}

	// Convert to LDT and back
	ldtFile := &LDTFile{}
	err := ldtFile.FromCommonModel(originalData)
	if err != nil {
		t.Fatalf("FromCommonModel failed: %v", err)
	}

	convertedData, err := ldtFile.ToCommonModel()
	if err != nil {
		t.Fatalf("ToCommonModel failed: %v", err)
	}

	// Verify key values are preserved (allowing for some precision loss)
	tolerance := 0.001

	if convertedData.Metadata.Manufacturer != originalData.Metadata.Manufacturer {
		t.Errorf("Manufacturer not preserved: got '%s', expected '%s'",
			convertedData.Metadata.Manufacturer, originalData.Metadata.Manufacturer)
	}

	if convertedData.Metadata.CatalogNumber != originalData.Metadata.CatalogNumber {
		t.Errorf("Catalog number not preserved: got '%s', expected '%s'",
			convertedData.Metadata.CatalogNumber, originalData.Metadata.CatalogNumber)
	}

	if abs(convertedData.Geometry.Length-originalData.Geometry.Length) > tolerance {
		t.Errorf("Length not preserved: got %f, expected %f",
			convertedData.Geometry.Length, originalData.Geometry.Length)
	}

	if abs(convertedData.Photometry.LuminousFlux-originalData.Photometry.LuminousFlux) > tolerance {
		t.Errorf("Luminous flux not preserved: got %f, expected %f",
			convertedData.Photometry.LuminousFlux, originalData.Photometry.LuminousFlux)
	}

	if abs(convertedData.Electrical.InputWatts-originalData.Electrical.InputWatts) > tolerance {
		t.Errorf("Input watts not preserved: got %f, expected %f",
			convertedData.Electrical.InputWatts, originalData.Electrical.InputWatts)
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
