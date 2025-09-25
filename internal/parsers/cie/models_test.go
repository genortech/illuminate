package cie

import (
	"illuminate/internal/formats/cie"
	"illuminate/internal/models"
	"testing"
)

func TestCIEFile_ToCommonModel(t *testing.T) {
	// Create a test CIE file
	cieFile := &cie.CIEFile{
		Header: cie.CIEHeader{
			FormatType:   1,
			SymmetryType: 0,
			Reserved:     0,
			Description:  "Test LED 17W 1000 lm",
		},
		Photometry: cie.CIEPhotometry{
			GammaAngles:  cie.GenerateStandardGammaAngles(),
			CPlaneAngles: cie.GenerateStandardCPlaneAngles(),
			IntensityData: func() [][]float64 {
				// Create test data: 19x16 matrix with decreasing values
				data := make([][]float64, 19)
				for i := range data {
					data[i] = make([]float64, 16)
					for j := range data[i] {
						// Create a simple pattern: higher values at nadir, decreasing with angle
						data[i][j] = float64(1000 - i*30 - j*10)
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

	// Test metadata extraction
	if result.Metadata.Manufacturer == "" {
		t.Error("Manufacturer should not be empty")
	}
	if result.Metadata.Description != "Test LED 17W 1000 lm" {
		t.Errorf("Description = %v, want 'Test LED 17W 1000 lm'", result.Metadata.Description)
	}

	// Test geometry defaults
	if result.Geometry.Length != 1.0 {
		t.Errorf("Length = %v, want 1.0", result.Geometry.Length)
	}
	if result.Geometry.Width != 1.0 {
		t.Errorf("Width = %v, want 1.0", result.Geometry.Width)
	}

	// Test photometry
	if result.Photometry.PhotometryType != "C" {
		t.Errorf("PhotometryType = %v, want 'C'", result.Photometry.PhotometryType)
	}
	if result.Photometry.UnitsType != "absolute" {
		t.Errorf("UnitsType = %v, want 'absolute'", result.Photometry.UnitsType)
	}
	if result.Photometry.LuminousFlux != 1000.0 {
		t.Errorf("LuminousFlux = %v, want 1000.0", result.Photometry.LuminousFlux)
	}

	// Test angular data
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
	for i, row := range result.Photometry.CandelaValues {
		if len(row) != 16 {
			t.Errorf("CandelaValues row %d length = %v, want 16", i, len(row))
		}
	}

	// Test electrical data extraction
	if result.Electrical.InputWatts != 17.0 {
		t.Errorf("InputWatts = %v, want 17.0", result.Electrical.InputWatts)
	}
}

func TestCIEFile_FromCommonModel(t *testing.T) {
	// Create test common model data
	commonData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test LED Luminaire",
		},
		Geometry: models.LuminaireGeometry{
			Length: 1.2,
			Width:  0.6,
			Height: 0.1,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1500.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0, 15, 30, 45, 60, 75, 90},
			HorizontalAngles:  []float64{0, 45, 90, 135, 180, 225, 270, 315},
			CandelaValues: [][]float64{
				{100, 95, 90, 85, 80, 85, 90, 95},
				{90, 85, 80, 75, 70, 75, 80, 85},
				{80, 75, 70, 65, 60, 65, 70, 75},
				{70, 65, 60, 55, 50, 55, 60, 65},
				{60, 55, 50, 45, 40, 45, 50, 55},
				{50, 45, 40, 35, 30, 35, 40, 45},
				{40, 35, 30, 25, 20, 25, 30, 35},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:   25.0,
			InputVoltage: 120.0,
		},
	}

	cieFile := &cie.CIEFile{}
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
		t.Error("Expected some non-zero intensity values after interpolation")
	}
}
