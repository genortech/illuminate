package ies

import (
	"illuminate/internal/models"
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator() returned nil")
	}
}

func TestValidateNilData(t *testing.T) {
	validator := NewValidator()

	err := validator.Validate(nil)
	if err == nil {
		t.Fatal("Expected error for nil data, got nil")
	}

	if !strings.Contains(err.Error(), "photometric data is nil") {
		t.Errorf("Expected error about nil data, got: %v", err)
	}
}

func TestValidatePhotometryType(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		photometryType string
		expectError    bool
	}{
		{"Valid Type A", "A", false},
		{"Valid Type B", "B", false},
		{"Valid Type C", "C", false},
		{"Invalid Type D", "D", true},
		{"Invalid Type X", "X", true},
		{"Empty Type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Photometry.PhotometryType = tt.photometryType

			err := validator.Validate(data)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for photometry type %s, got nil", tt.photometryType)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for photometry type %s, got: %v", tt.photometryType, err)
			}
		})
	}
}

func TestValidateTypeAAngles(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		verticalAngles   []float64
		horizontalAngles []float64
		expectError      bool
	}{
		{
			name:             "Valid Type A angles",
			verticalAngles:   []float64{0.0, 45.0, 90.0, 135.0, 180.0},
			horizontalAngles: []float64{0.0, 90.0, 180.0, 270.0},
			expectError:      false,
		},
		{
			name:             "Invalid vertical angle > 180",
			verticalAngles:   []float64{0.0, 45.0, 190.0},
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
		},
		{
			name:             "Invalid vertical angle < 0",
			verticalAngles:   []float64{-10.0, 45.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
		},
		{
			name:             "Invalid horizontal angle >= 360",
			verticalAngles:   []float64{0.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 360.0},
			expectError:      true,
		},
		{
			name:             "Invalid horizontal angle < 0",
			verticalAngles:   []float64{0.0, 90.0},
			horizontalAngles: []float64{-10.0, 90.0},
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Photometry.PhotometryType = "A"
			data.Photometry.VerticalAngles = tt.verticalAngles
			data.Photometry.HorizontalAngles = tt.horizontalAngles
			data.Photometry.CandelaValues = createCandelaMatrix(len(tt.verticalAngles), len(tt.horizontalAngles))

			err := validator.Validate(data)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for Type A angles, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for Type A angles, got: %v", err)
			}
		})
	}
}

func TestValidateTypeBAngles(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		verticalAngles   []float64
		horizontalAngles []float64
		expectError      bool
	}{
		{
			name:             "Valid Type B angles",
			verticalAngles:   []float64{0.0, 45.0, 90.0, 135.0, 180.0},
			horizontalAngles: []float64{0.0, 45.0, 90.0, 135.0, 180.0},
			expectError:      false,
		},
		{
			name:             "Invalid horizontal angle > 180",
			verticalAngles:   []float64{0.0, 90.0},
			horizontalAngles: []float64{0.0, 90.0, 190.0},
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Photometry.PhotometryType = "B"
			data.Photometry.VerticalAngles = tt.verticalAngles
			data.Photometry.HorizontalAngles = tt.horizontalAngles
			data.Photometry.CandelaValues = createCandelaMatrix(len(tt.verticalAngles), len(tt.horizontalAngles))

			err := validator.Validate(data)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for Type B angles, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for Type B angles, got: %v", err)
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
			name:             "Too few vertical angles",
			verticalAngles:   []float64{0.0},
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
			errorContains:    "minimum 2 vertical angles",
		},
		{
			name:             "No horizontal angles",
			verticalAngles:   []float64{0.0, 90.0},
			horizontalAngles: []float64{},
			expectError:      true,
			errorContains:    "horizontal angles array cannot be empty",
		},
		{
			name:             "Too many vertical angles",
			verticalAngles:   createValidVerticalAngles(182), // > 181
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
			errorContains:    "maximum 181 vertical angles",
		},
		{
			name:             "Too many horizontal angles",
			verticalAngles:   []float64{0.0, 90.0},
			horizontalAngles: createValidHorizontalAngles(362), // > 361
			expectError:      true,
			errorContains:    "maximum 361 horizontal angles",
		},
		{
			name:             "Non-ascending vertical angles",
			verticalAngles:   []float64{0.0, 90.0, 45.0},
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
			errorContains:    "ascending order",
		},
		{
			name:             "Too small angle increment",
			verticalAngles:   []float64{0.0, 0.05}, // 0.05° increment
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
			errorContains:    "increment too small",
		},
		{
			name:             "Too large angle increment",
			verticalAngles:   []float64{0.0, 95.0}, // 95° increment
			horizontalAngles: []float64{0.0, 90.0},
			expectError:      true,
			errorContains:    "increment too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Photometry.VerticalAngles = tt.verticalAngles
			data.Photometry.HorizontalAngles = tt.horizontalAngles

			// Create appropriate candela matrix if angles are valid
			if len(tt.verticalAngles) > 0 && len(tt.horizontalAngles) > 0 {
				data.Photometry.CandelaValues = createCandelaMatrix(len(tt.verticalAngles), len(tt.horizontalAngles))
			}

			err := validator.Validate(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got: %v", err)
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
				{100.0, 200.0},
				{150.0, 250.0},
			},
			expectError: false,
		},
		{
			name: "Negative candela value",
			candelaValues: [][]float64{
				{100.0, -50.0},
				{150.0, 250.0},
			},
			expectError:   true,
			errorContains: "candela value at [0][1] cannot be negative",
		},
		{
			name: "Unreasonably high candela value",
			candelaValues: [][]float64{
				{100.0, 2000000.0}, // > 1 million
				{150.0, 250.0},
			},
			expectError:   true,
			errorContains: "unreasonably high candela value",
		},
		{
			name: "Extreme dynamic range",
			candelaValues: [][]float64{
				{0.000001, 2000.0}, // Very high dynamic range
				{0.000001, 1500.0},
			},
			expectError:   true,
			errorContains: "dynamic range too high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Photometry.CandelaValues = tt.candelaValues
			data.Photometry.VerticalAngles = make([]float64, len(tt.candelaValues))
			data.Photometry.HorizontalAngles = make([]float64, len(tt.candelaValues[0]))

			// Fill angles with valid values
			for i := range data.Photometry.VerticalAngles {
				data.Photometry.VerticalAngles[i] = float64(i * 45)
			}
			for i := range data.Photometry.HorizontalAngles {
				data.Photometry.HorizontalAngles[i] = float64(i * 90)
			}

			err := validator.Validate(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got: %v", err)
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
			luminousFlux:      2000000.0,
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
			candelaMultiplier: 2000000.0,
			expectError:       true,
			errorContains:     "candela multiplier too high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Photometry.LuminousFlux = tt.luminousFlux
			data.Photometry.CandelaMultiplier = tt.candelaMultiplier

			err := validator.Validate(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got: %v", err)
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
				Length: 1.0,
				Width:  0.5,
				Height: 0.3,
			},
			expectError: false,
		},
		{
			name: "Negative length",
			geometry: models.LuminaireGeometry{
				Length: -1.0,
				Width:  0.5,
				Height: 0.3,
			},
			expectError:   true,
			errorContains: "length cannot be negative",
		},
		{
			name: "Too large width",
			geometry: models.LuminaireGeometry{
				Length: 1.0,
				Width:  150.0, // > 100m
				Height: 0.3,
			},
			expectError:   true,
			errorContains: "width too large",
		},
		{
			name: "Too small height",
			geometry: models.LuminaireGeometry{
				Length: 1.0,
				Width:  0.5,
				Height: 0.0005, // < 0.001m
			},
			expectError:   true,
			errorContains: "height too small",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Geometry = tt.geometry

			err := validator.Validate(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got: %v", err)
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
				InputWatts:        100.0,
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
			},
			expectError: false,
		},
		{
			name: "Negative input watts",
			electrical: models.ElectricalData{
				InputWatts:        -50.0,
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
			},
			expectError:   true,
			errorContains: "input watts cannot be negative",
		},
		{
			name: "Too high input watts",
			electrical: models.ElectricalData{
				InputWatts:        150000.0, // > 100kW
				BallastFactor:     1.0,
				BallastLampFactor: 1.0,
			},
			expectError:   true,
			errorContains: "input watts too high",
		},
		{
			name: "Too high ballast factor",
			electrical: models.ElectricalData{
				InputWatts:        100.0,
				BallastFactor:     6.0, // > 5.0
				BallastLampFactor: 1.0,
			},
			expectError:   true,
			errorContains: "ballast factor must be between 0 and 2",
		},
		{
			name: "Negative ballast lamp factor",
			electrical: models.ElectricalData{
				InputWatts:        100.0,
				BallastFactor:     1.0,
				BallastLampFactor: -0.5,
			},
			expectError:   true,
			errorContains: "ballast lamp factor must be between 0 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createValidPhotometricData()
			data.Electrical = tt.electrical

			err := validator.Validate(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// Helper function to create valid photometric data for testing
func createValidPhotometricData() *models.PhotometricData {
	return &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "Test Manufacturer",
			CatalogNumber: "TEST-001",
		},
		Geometry: models.LuminaireGeometry{
			Length: 1.0,
			Width:  0.5,
			Height: 0.3,
		},
		Photometry: models.PhotometricMeasurements{
			PhotometryType:    "C",
			UnitsType:         "absolute",
			LuminousFlux:      1000.0,
			CandelaMultiplier: 1.0,
			VerticalAngles:    []float64{0.0, 45.0, 90.0},
			HorizontalAngles:  []float64{0.0, 90.0, 180.0},
			CandelaValues: [][]float64{
				{100.0, 150.0, 200.0},
				{120.0, 180.0, 240.0},
				{80.0, 120.0, 160.0},
			},
		},
		Electrical: models.ElectricalData{
			InputWatts:        100.0,
			BallastFactor:     1.0,
			BallastLampFactor: 1.0,
		},
	}
}

// Helper function to create candela matrix with given dimensions
func createCandelaMatrix(verticalCount, horizontalCount int) [][]float64 {
	matrix := make([][]float64, verticalCount)
	for i := 0; i < verticalCount; i++ {
		matrix[i] = make([]float64, horizontalCount)
		for j := 0; j < horizontalCount; j++ {
			matrix[i][j] = float64((i+1)*100 + (j+1)*10) // Generate some test values
		}
	}
	return matrix
}

// Helper function to create ascending angle arrays
func createAscendingAngles(count int) []float64 {
	angles := make([]float64, count)
	increment := 180.0 / float64(count-1) // Distribute angles from 0 to 180
	if increment < 1.0 {
		increment = 1.0 // Minimum 1-degree increment
	}
	for i := 0; i < count; i++ {
		angles[i] = float64(i) * increment
		if angles[i] > 180.0 {
			angles[i] = 180.0 // Cap at 180 degrees
		}
	}
	return angles
}

// Helper function to create valid vertical angles (0-180 degrees)
func createValidVerticalAngles(count int) []float64 {
	angles := make([]float64, count)
	increment := 180.0 / float64(count-1)
	if increment < 1.0 {
		increment = 1.0
	}
	for i := 0; i < count; i++ {
		angle := float64(i) * increment
		if angle > 180.0 {
			angle = 180.0
		}
		angles[i] = angle
	}
	return angles
}

// Helper function to create valid horizontal angles (0-359 degrees)
func createValidHorizontalAngles(count int) []float64 {
	angles := make([]float64, count)
	increment := 359.0 / float64(count-1)
	if increment < 1.0 {
		increment = 1.0
	}
	for i := 0; i < count; i++ {
		angle := float64(i) * increment
		if angle >= 360.0 {
			angle = 359.0
		}
		angles[i] = angle
	}
	return angles
}
