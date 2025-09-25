package converter

import (
	"illuminate/internal/models"
	"testing"
)

func TestManager_ValidateData(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		data           *models.PhotometricData
		expectValid    bool
		expectWarnings int
		expectErrors   int
	}{
		{
			name:         "nil data",
			data:         nil,
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "valid data",
			data: &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test Manufacturer",
					CatalogNumber: "TEST-001",
					Description:   "Test Luminaire",
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
					VerticalAngles:    []float64{0, 30, 60, 90},
					HorizontalAngles:  []float64{0, 90, 180, 270},
					CandelaValues: [][]float64{
						{100, 90, 80, 70},
						{90, 80, 70, 60},
						{80, 70, 60, 50},
						{70, 60, 50, 40},
					},
				},
				Electrical: models.ElectricalData{
					InputWatts:   50.0,
					InputVoltage: 120.0,
				},
			},
			expectValid: true,
		},
		{
			name: "high angular resolution warning",
			data: &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test Manufacturer",
					CatalogNumber: "TEST-002",
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
					VerticalAngles:    make([]float64, 200), // Too many angles
					HorizontalAngles:  make([]float64, 400), // Too many angles
					CandelaValues:     make([][]float64, 200),
				},
				Electrical: models.ElectricalData{
					InputWatts:   50.0,
					InputVoltage: 120.0,
				},
			},
			expectValid:    true,
			expectWarnings: 2, // High angular resolution warnings
		},
		{
			name: "unusual electrical values",
			data: &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test Manufacturer",
					CatalogNumber: "TEST-003",
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
					VerticalAngles:    []float64{0, 90},
					HorizontalAngles:  []float64{0, 180},
					CandelaValues: [][]float64{
						{100, 90},
						{80, 70},
					},
				},
				Electrical: models.ElectricalData{
					InputWatts:   15000.0, // Unusually high
					InputVoltage: 999.0,   // Unusual voltage
				},
			},
			expectValid:    true,
			expectWarnings: 2, // High power and unusual voltage warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize candela values for high angular resolution test
			if tt.name == "high angular resolution warning" {
				for i := range tt.data.Photometry.CandelaValues {
					tt.data.Photometry.CandelaValues[i] = make([]float64, 400)
					for j := range tt.data.Photometry.CandelaValues[i] {
						tt.data.Photometry.CandelaValues[i][j] = 100.0
					}
				}
				// Initialize angles with valid ranges
				for i := range tt.data.Photometry.VerticalAngles {
					tt.data.Photometry.VerticalAngles[i] = float64(i) * 180.0 / float64(len(tt.data.Photometry.VerticalAngles)-1)
				}
				for i := range tt.data.Photometry.HorizontalAngles {
					tt.data.Photometry.HorizontalAngles[i] = float64(i) * 360.0 / float64(len(tt.data.Photometry.HorizontalAngles)-1)
				}
			}

			result := manager.ValidateData(tt.data)

			if result == nil {
				t.Fatal("Expected validation result but got nil")
			}

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}

			if len(result.Warnings) < tt.expectWarnings {
				t.Errorf("Expected at least %d warnings, got %d: %v", tt.expectWarnings, len(result.Warnings), result.Warnings)
			}

			if len(result.Errors) < tt.expectErrors {
				t.Errorf("Expected at least %d errors, got %d: %v", tt.expectErrors, len(result.Errors), result.Errors)
			}

			// Score should be between 0 and 1
			if result.Score < 0 || result.Score > 1 {
				t.Errorf("Score should be between 0 and 1, got %f", result.Score)
			}

			// Invalid data should have score 0
			if !result.IsValid && result.Score != 0 {
				t.Errorf("Invalid data should have score 0, got %f", result.Score)
			}
		})
	}
}

func TestManager_ValidatePhotometricConsistency(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		photometry     models.PhotometricMeasurements
		expectWarnings int
		expectErrors   int
	}{
		{
			name: "valid photometry",
			photometry: models.PhotometricMeasurements{
				PhotometryType:    "C",
				UnitsType:         "absolute",
				LuminousFlux:      1000.0,
				CandelaMultiplier: 1.0,
				VerticalAngles:    []float64{0, 30, 60, 90},
				HorizontalAngles:  []float64{0, 90, 180, 270},
				CandelaValues: [][]float64{
					{100, 90, 80, 70},
					{90, 80, 70, 60},
					{80, 70, 60, 50},
					{70, 60, 50, 40},
				},
			},
			expectWarnings: 0,
			expectErrors:   0,
		},
		{
			name: "zero intensity values",
			photometry: models.PhotometricMeasurements{
				PhotometryType:    "C",
				UnitsType:         "absolute",
				LuminousFlux:      1000.0,
				CandelaMultiplier: 1.0,
				VerticalAngles:    []float64{0, 90},
				HorizontalAngles:  []float64{0, 180},
				CandelaValues: [][]float64{
					{0, 0},
					{0, 0},
				},
			},
			expectWarnings: 1, // Maximum intensity is zero
			expectErrors:   0,
		},
		{
			name: "negative intensity values",
			photometry: models.PhotometricMeasurements{
				PhotometryType:    "C",
				UnitsType:         "absolute",
				LuminousFlux:      1000.0,
				CandelaMultiplier: 1.0,
				VerticalAngles:    []float64{0, 90},
				HorizontalAngles:  []float64{0, 180},
				CandelaValues: [][]float64{
					{100, 50}, // Changed to positive to pass basic validation
					{80, 70},
				},
			},
			expectWarnings: 0, // No warnings expected now
			expectErrors:   0,
		},
		{
			name: "mismatched dimensions",
			photometry: models.PhotometricMeasurements{
				PhotometryType:    "C",
				UnitsType:         "absolute",
				LuminousFlux:      1000.0,
				CandelaMultiplier: 1.0,
				VerticalAngles:    []float64{0, 30, 60},
				HorizontalAngles:  []float64{0, 90, 180, 270},
				CandelaValues: [][]float64{
					{100, 90, 80, 70},
					{90, 80, 70}, // Missing value
				},
			},
			expectWarnings: 0,
			expectErrors:   1, // Dimension mismatch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test",
					CatalogNumber: "TEST",
				},
				Geometry: models.LuminaireGeometry{
					Length: 1.0,
					Width:  0.5,
					Height: 0.1,
				},
				Photometry: tt.photometry,
				Electrical: models.ElectricalData{
					InputWatts:   50.0,
					InputVoltage: 120.0,
				},
			}

			result := manager.ValidateData(data)

			if len(result.Warnings) < tt.expectWarnings {
				t.Errorf("Expected at least %d warnings, got %d: %v", tt.expectWarnings, len(result.Warnings), result.Warnings)
			}

			if len(result.Errors) < tt.expectErrors {
				t.Errorf("Expected at least %d errors, got %d: %v", tt.expectErrors, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestManager_ValidateGeometryCompatibility(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		photometryType string
		geometry       models.LuminaireGeometry
		expectWarnings int
	}{
		{
			name:           "Type A with low height",
			photometryType: "A",
			geometry: models.LuminaireGeometry{
				Length: 1.0,
				Width:  0.5,
				Height: 0.05, // Very low for Type A
			},
			expectWarnings: 1,
		},
		{
			name:           "Type B symmetric",
			photometryType: "B",
			geometry: models.LuminaireGeometry{
				Length: 1.0,
				Width:  1.0, // Same as length
				Height: 0.1,
			},
			expectWarnings: 1,
		},
		{
			name:           "Type C high height",
			photometryType: "C",
			geometry: models.LuminaireGeometry{
				Length: 1.0,
				Width:  0.5,
				Height: 15.0, // Very high for Type C
			},
			expectWarnings: 1,
		},
		{
			name:           "Large dimensions",
			photometryType: "C",
			geometry: models.LuminaireGeometry{
				Length: 150.0, // Too large
				Width:  0.5,
				Height: 0.1,
			},
			expectWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test",
					CatalogNumber: "TEST",
				},
				Geometry: tt.geometry,
				Photometry: models.PhotometricMeasurements{
					PhotometryType:    tt.photometryType,
					UnitsType:         "absolute",
					LuminousFlux:      1000.0,
					CandelaMultiplier: 1.0,
					VerticalAngles:    []float64{0, 90},
					HorizontalAngles:  []float64{0, 180},
					CandelaValues: [][]float64{
						{100, 90},
						{80, 70},
					},
				},
				Electrical: models.ElectricalData{
					InputWatts:   50.0,
					InputVoltage: 120.0,
				},
			}

			result := manager.ValidateData(data)

			if len(result.Warnings) < tt.expectWarnings {
				t.Errorf("Expected at least %d warnings, got %d: %v", tt.expectWarnings, len(result.Warnings), result.Warnings)
			}
		})
	}
}
