package ldt

import (
	"illuminate/internal/models"
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator() returned nil")
	}
}

func TestValidate_ValidData(t *testing.T) {
	validator := NewValidator()

	// Create valid photometric data
	data := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
		},
		Geometry: models.LuminaireGeometry{
			Length:         0.6,  // 600mm
			Width:          0.25, // 250mm
			Height:         0.19, // 190mm
			LuminousLength: 0.18, // 180mm
			LuminousWidth:  0.16, // 160mm
			LuminousHeight: 0.01, // 10mm
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

	err := validator.Validate(data)
	if err != nil {
		t.Errorf("Validation failed for valid data: %v", err)
	}
}

func TestValidate_NilData(t *testing.T) {
	validator := NewValidator()

	err := validator.Validate(nil)
	if err == nil {
		t.Error("Expected error for nil data, but got none")
	}
}

func TestValidatePhotometryType(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		photometryType string
		expectError    bool
	}{
		{"Valid Type C", "C", false},
		{"Valid Type A", "A", false},
		{"Valid Type B", "B", false},
		{"Invalid Type", "X", true},
		{"Empty Type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Photometry.PhotometryType = tt.photometryType

			err := validator.validatePhotometryType(data)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateLDTAngles(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		verticalAngles   []float64
		horizontalAngles []float64
		expectError      bool
	}{
		{
			name:             "Valid angles",
			verticalAngles:   []float64{0.0, 30.0, 60.0, 90.0, 120.0, 150.0, 180.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      false,
		},
		{
			name:             "Invalid vertical angle - negative",
			verticalAngles:   []float64{-10.0, 30.0, 60.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      true,
		},
		{
			name:             "Invalid vertical angle - over 180",
			verticalAngles:   []float64{0.0, 30.0, 60.0, 190.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      true,
		},
		{
			name:             "Invalid horizontal angle - negative",
			verticalAngles:   []float64{0.0, 30.0, 60.0, 90.0},
			horizontalAngles: []float64{-10.0, 90.0, 180.0, 270.0},
			expectError:      true,
		},
		{
			name:             "Invalid horizontal angle - over 360",
			verticalAngles:   []float64{0.0, 30.0, 60.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 370.0},
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Photometry.VerticalAngles = tt.verticalAngles
			data.Photometry.HorizontalAngles = tt.horizontalAngles
			// Update candela values to match new angle arrays
			data.Photometry.CandelaValues = make([][]float64, len(tt.verticalAngles))
			for i := range data.Photometry.CandelaValues {
				data.Photometry.CandelaValues[i] = make([]float64, len(tt.horizontalAngles))
				for j := range data.Photometry.CandelaValues[i] {
					data.Photometry.CandelaValues[i][j] = 100.0
				}
			}

			err := validator.validateLDTAngles(data)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAngles(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		verticalAngles   []float64
		horizontalAngles []float64
		expectError      bool
		errorContains    string
	}{
		{
			name:             "Valid angles",
			verticalAngles:   []float64{0.0, 30.0, 60.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      false,
		},
		{
			name:             "Too few vertical angles",
			verticalAngles:   []float64{0.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      true,
			errorContains:    "minimum 2 gamma angles",
		},
		{
			name:             "Too few horizontal angles",
			verticalAngles:   []float64{0.0, 30.0, 60.0, 90.0},
			horizontalAngles: []float64{},
			expectError:      true,
			errorContains:    "minimum 1 C-plane angle",
		},
		{
			name:             "Non-ascending vertical angles",
			verticalAngles:   []float64{0.0, 60.0, 30.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      true,
			errorContains:    "ascending order",
		},
		{
			name:             "First gamma angle not zero",
			verticalAngles:   []float64{10.0, 30.0, 60.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      true,
			errorContains:    "first gamma angle should be 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Photometry.VerticalAngles = tt.verticalAngles
			data.Photometry.HorizontalAngles = tt.horizontalAngles
			// Update candela values to match new angle arrays
			if len(tt.verticalAngles) > 0 && len(tt.horizontalAngles) > 0 {
				data.Photometry.CandelaValues = make([][]float64, len(tt.verticalAngles))
				for i := range data.Photometry.CandelaValues {
					data.Photometry.CandelaValues[i] = make([]float64, len(tt.horizontalAngles))
					for j := range data.Photometry.CandelaValues[i] {
						data.Photometry.CandelaValues[i][j] = 100.0
					}
				}
			}

			err := validator.validateAngles(data)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCandelaValues(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name          string
		candelaValues [][]float64
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid candela values",
			candelaValues: [][]float64{
				{100.0, 90.0, 80.0, 70.0},
				{95.0, 85.0, 75.0, 65.0},
				{80.0, 70.0, 60.0, 50.0},
				{50.0, 40.0, 30.0, 20.0},
			},
			expectError: false,
		},
		{
			name: "Negative candela value",
			candelaValues: [][]float64{
				{100.0, 90.0, 80.0, 70.0},
				{95.0, -10.0, 75.0, 65.0},
				{80.0, 70.0, 60.0, 50.0},
				{50.0, 40.0, 30.0, 20.0},
			},
			expectError:   true,
			errorContains: "negative candela value",
		},
		{
			name: "Extremely high candela value",
			candelaValues: [][]float64{
				{100.0, 90.0, 80.0, 70.0},
				{95.0, 20000000.0, 75.0, 65.0},
				{80.0, 70.0, 60.0, 50.0},
				{50.0, 40.0, 30.0, 20.0},
			},
			expectError:   true,
			errorContains: "unreasonably high candela value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Photometry.CandelaValues = tt.candelaValues

			err := validator.validateCandelaValues(data)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateLuminousFlux(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name              string
		luminousFlux      float64
		candelaMultiplier float64
		expectError       bool
		errorContains     string
	}{
		{
			name:              "Valid flux and multiplier",
			luminousFlux:      1000.0,
			candelaMultiplier: 1.0,
			expectError:       false,
		},
		{
			name:              "Negative luminous flux",
			luminousFlux:      -100.0,
			candelaMultiplier: 1.0,
			expectError:       true,
			errorContains:     "luminous flux cannot be negative",
		},
		{
			name:              "Too low luminous flux",
			luminousFlux:      0.05,
			candelaMultiplier: 1.0,
			expectError:       true,
			errorContains:     "luminous flux too low",
		},
		{
			name:              "Too high luminous flux",
			luminousFlux:      20000000.0,
			candelaMultiplier: 1.0,
			expectError:       true,
			errorContains:     "luminous flux too high",
		},
		{
			name:              "Zero candela multiplier",
			luminousFlux:      1000.0,
			candelaMultiplier: 0.0,
			expectError:       true,
			errorContains:     "candela multiplier must be positive",
		},
		{
			name:              "Too high candela multiplier",
			luminousFlux:      1000.0,
			candelaMultiplier: 20000000.0,
			expectError:       true,
			errorContains:     "candela multiplier too high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Photometry.LuminousFlux = tt.luminousFlux
			data.Photometry.CandelaMultiplier = tt.candelaMultiplier

			err := validator.validateLuminousFlux(data)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateGeometry(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name          string
		geometry      models.LuminaireGeometry
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid geometry",
			geometry: models.LuminaireGeometry{
				Length:         0.6,
				Width:          0.25,
				Height:         0.19,
				LuminousLength: 0.18,
				LuminousWidth:  0.16,
				LuminousHeight: 0.01,
			},
			expectError: false,
		},
		{
			name: "Negative length",
			geometry: models.LuminaireGeometry{
				Length:         -0.1,
				Width:          0.25,
				Height:         0.19,
				LuminousLength: 0.18,
				LuminousWidth:  0.16,
				LuminousHeight: 0.01,
			},
			expectError:   true,
			errorContains: "length cannot be negative",
		},
		{
			name: "Too large dimensions",
			geometry: models.LuminaireGeometry{
				Length:         2000.0, // 2000 meters
				Width:          0.25,
				Height:         0.19,
				LuminousLength: 0.18,
				LuminousWidth:  0.16,
				LuminousHeight: 0.01,
			},
			expectError:   true,
			errorContains: "length too large",
		},
		{
			name: "Luminous length exceeds physical length",
			geometry: models.LuminaireGeometry{
				Length:         0.5,
				Width:          0.25,
				Height:         0.19,
				LuminousLength: 0.6, // Larger than physical length
				LuminousWidth:  0.16,
				LuminousHeight: 0.01,
			},
			expectError:   true,
			errorContains: "luminous length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Geometry = tt.geometry

			err := validator.validateGeometry(data)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateElectricalData(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name          string
		electrical    models.ElectricalData
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid electrical data",
			electrical: models.ElectricalData{
				InputWatts:        10.0,
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
				PowerFactor:       0.9,
			},
			expectError: false,
		},
		{
			name: "Negative input watts",
			electrical: models.ElectricalData{
				InputWatts:        -10.0,
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
			},
			expectError:   true,
			errorContains: "input watts cannot be negative",
		},
		{
			name: "Too high input watts",
			electrical: models.ElectricalData{
				InputWatts:        2000000.0,
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
			},
			expectError:   true,
			errorContains: "input watts too high",
		},
		{
			name: "Invalid power factor",
			electrical: models.ElectricalData{
				InputWatts:        10.0,
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
				PowerFactor:       1.5, // > 1.0
			},
			expectError:   true,
			errorContains: "power factor cannot exceed 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidTestData()
			data.Electrical = tt.electrical

			err := validator.validateElectricalData(data)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Helper functions

func createValidTestData() *models.PhotometricData {
	return &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
			Description:   "Test Luminaire",
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
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
