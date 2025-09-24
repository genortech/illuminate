package models

import (
	"testing"
)

func TestPhotometricDataValidation(t *testing.T) {
	// Test valid photometric data
	validData := &PhotometricData{
		Metadata: LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
		},
		Geometry: LuminaireGeometry{
			Length:         100.0,
			Width:          50.0,
			Height:         25.0,
			LuminousLength: 90.0,
			LuminousWidth:  45.0,
			LuminousHeight: 20.0,
		},
		Photometry: PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0, 30, 60, 90},
			HorizontalAngles:  []float64{0, 90, 180, 270},
			CandelaValues: [][]float64{
				{100, 90, 80, 90},
				{80, 70, 60, 70},
				{50, 40, 30, 40},
				{10, 5, 0, 5},
			},
		},
		Electrical: ElectricalData{
			InputWatts:    20.0,
			InputVoltage:  120.0,
			InputCurrent:  0.2,
			PowerFactor:   0.85,
			BallastFactor: 1.0,
		},
	}

	err := validData.Validate()
	if err != nil {
		t.Errorf("Valid data should not produce validation error: %v", err)
	}
}

func TestLuminaireMetadataValidation(t *testing.T) {
	tests := []struct {
		name        string
		metadata    LuminaireMetadata
		expectError bool
	}{
		{
			name: "valid metadata",
			metadata: LuminaireMetadata{
				Manufacturer:  "Test Manufacturer",
				CatalogNumber: "TEST-001",
			},
			expectError: false,
		},
		{
			name: "missing manufacturer",
			metadata: LuminaireMetadata{
				CatalogNumber: "TEST-001",
			},
			expectError: true,
		},
		{
			name: "missing catalog number",
			metadata: LuminaireMetadata{
				Manufacturer: "Test Manufacturer",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestLuminaireGeometryValidation(t *testing.T) {
	tests := []struct {
		name        string
		geometry    LuminaireGeometry
		expectError bool
	}{
		{
			name: "valid geometry",
			geometry: LuminaireGeometry{
				Length:         100.0,
				Width:          50.0,
				Height:         25.0,
				LuminousLength: 90.0,
				LuminousWidth:  45.0,
				LuminousHeight: 20.0,
			},
			expectError: false,
		},
		{
			name: "negative length",
			geometry: LuminaireGeometry{
				Length: -10.0,
			},
			expectError: true,
		},
		{
			name: "luminous length exceeds physical",
			geometry: LuminaireGeometry{
				Length:         100.0,
				LuminousLength: 150.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.geometry.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestPhotometricMeasurementsValidation(t *testing.T) {
	tests := []struct {
		name        string
		photometry  PhotometricMeasurements
		expectError bool
	}{
		{
			name: "valid photometry",
			photometry: PhotometricMeasurements{
				PhotometryType:    "C",
				UnitsType:         "absolute",
				LuminousFlux:      1000.0,
				CandelaMultiplier: 1.0,
				VerticalAngles:    []float64{0, 30, 60, 90},
				HorizontalAngles:  []float64{0, 90, 180, 270},
				CandelaValues: [][]float64{
					{100, 90, 80, 90},
					{80, 70, 60, 70},
					{50, 40, 30, 40},
					{10, 5, 0, 5},
				},
			},
			expectError: false,
		},
		{
			name: "invalid photometry type",
			photometry: PhotometricMeasurements{
				PhotometryType:   "X",
				UnitsType:        "absolute",
				VerticalAngles:   []float64{0, 30},
				HorizontalAngles: []float64{0, 90},
				CandelaValues: [][]float64{
					{100, 90},
					{80, 70},
				},
			},
			expectError: true,
		},
		{
			name: "mismatched candela array dimensions",
			photometry: PhotometricMeasurements{
				PhotometryType:   "C",
				UnitsType:        "absolute",
				VerticalAngles:   []float64{0, 30, 60},
				HorizontalAngles: []float64{0, 90},
				CandelaValues: [][]float64{
					{100, 90},
					{80, 70},
				}, // Missing third row
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.photometry.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}
