package cie

import (
	"illuminate/internal/formats/cie"
	"illuminate/internal/models"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	validator := cie.NewValidator()

	tests := []struct {
		name    string
		data    *models.PhotometricData
		wantErr bool
	}{
		{
			name: "Valid CIE data",
			data: &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test Manufacturer",
					CatalogNumber: "TEST-001",
					Description:   "Test LED Luminaire",
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
			},
			wantErr: false,
		},
		{
			name:    "Nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "Invalid photometry type",
			data: &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test Manufacturer",
					CatalogNumber: "TEST-001",
				},
				Geometry: models.LuminaireGeometry{
					Length: 1.0,
					Width:  0.5,
					Height: 0.1,
				},
				Photometry: models.PhotometricMeasurements{
					PhotometryType:    "A", // Invalid for CIE
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
					InputWatts:   25.0,
					InputVoltage: 120.0,
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid vertical angle range",
			data: &models.PhotometricData{
				Metadata: models.LuminaireMetadata{
					Manufacturer:  "Test Manufacturer",
					CatalogNumber: "TEST-001",
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
					VerticalAngles:    []float64{0, 30, 60, 120}, // 120° is invalid for CIE
					HorizontalAngles:  []float64{0, 90, 180, 270},
					CandelaValues: [][]float64{
						{100, 90, 80, 70},
						{90, 80, 70, 60},
						{80, 70, 60, 50},
						{70, 60, 50, 40},
					},
				},
				Electrical: models.ElectricalData{
					InputWatts:   25.0,
					InputVoltage: 120.0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidator_ValidateCIEFile(t *testing.T) {
	validator := cie.NewValidator()

	tests := []struct {
		name    string
		cieFile *cie.CIEFile
		wantErr bool
	}{
		{
			name: "Valid CIE file",
			cieFile: &cie.CIEFile{
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
						data := make([][]float64, 19)
						for i := range data {
							data[i] = make([]float64, 16)
							for j := range data[i] {
								data[i][j] = 100.0
							}
						}
						return data
					}(),
				},
			},
			wantErr: false,
		},
		{
			name:    "Nil CIE file",
			cieFile: nil,
			wantErr: true,
		},
		{
			name: "Invalid format type",
			cieFile: &cie.CIEFile{
				Header: cie.CIEHeader{
					FormatType:   2, // Invalid
					SymmetryType: 0,
					Reserved:     0,
					Description:  "Test LED",
				},
				Photometry: cie.CIEPhotometry{
					GammaAngles:   cie.GenerateStandardGammaAngles(),
					CPlaneAngles:  cie.GenerateStandardCPlaneAngles(),
					IntensityData: make([][]float64, 19),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCIEFile(tt.cieFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCIEFile() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCIEFile() unexpected error = %v", err)
				}
			}
		})
	}
}
