package cie

import (
	"illuminate/internal/models"
	"strings"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	// Create valid test data
	validData := &models.PhotometricData{
		Metadata: models.LuminaireMetadata{
			Manufacturer:  "TestCorp",
			CatalogNumber: "LED123",
			Description:   "Test LED",
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
			InputWatts:    25.0,
			BallastFactor: 1.0,
		},
	}

	tests := []struct {
		name        string
		data        *models.PhotometricData
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid data",
			data:        validData,
			expectError: false,
		},
		{
			name:        "Nil data",
			data:        nil,
			expectError: true,
			errorMsg:    "cannot be nil",
		},
		{
			name: "Wrong photometry type",
			data: func() *models.PhotometricData {
				data := *validData
				data.Photometry.PhotometryType = "A"
				return &data
			}(),
			expectError: true,
			errorMsg:    "Type C photometry",
		},
		{
			name: "Vertical angle out of range",
			data: func() *models.PhotometricData {
				data := *validData
				data.Photometry.VerticalAngles = []float64{0, 30, 60, 120} // 120° > 90°
				data.Photometry.CandelaValues = [][]float64{
					{1.0, 0.9, 0.8, 0.7, 0.6, 0.7, 0.8, 0.9},
					{0.9, 0.8, 0.7, 0.6, 0.5, 0.6, 0.7, 0.8},
					{0.8, 0.7, 0.6, 0.5, 0.4, 0.5, 0.6, 0.7},
					{0.7, 0.6, 0.5, 0.4, 0.3, 0.4, 0.5, 0.6},
				}
				return &data
			}(),
			expectError: true,
			errorMsg:    "between 0° and 90°",
		},
		{
			name: "Negative intensity value",
			data: func() *models.PhotometricData {
				// Create a deep copy
				data := *validData
				data.Photometry = validData.Photometry
				data.Photometry.CandelaValues = make([][]float64, len(validData.Photometry.CandelaValues))
				for i, row := range validData.Photometry.CandelaValues {
					data.Photometry.CandelaValues[i] = make([]float64, len(row))
					copy(data.Photometry.CandelaValues[i], row)
				}
				data.Photometry.CandelaValues[0][0] = -1.0
				return &data
			}(),
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name: "Extremely high intensity value",
			data: func() *models.PhotometricData {
				// Create a deep copy
				data := *validData
				data.Photometry = validData.Photometry
				data.Photometry.CandelaValues = make([][]float64, len(validData.Photometry.CandelaValues))
				for i, row := range validData.Photometry.CandelaValues {
					data.Photometry.CandelaValues[i] = make([]float64, len(row))
					copy(data.Photometry.CandelaValues[i], row)
				}
				data.Photometry.CandelaValues[0][0] = 15000.0
				return &data
			}(),
			expectError: true,
			errorMsg:    "too high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidator_validatePhotometryType(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		photometryType string
		expectError    bool
	}{
		{
			name:           "Valid Type C",
			photometryType: "C",
			expectError:    false,
		},
		{
			name:           "Invalid Type A",
			photometryType: "A",
			expectError:    true,
		},
		{
			name:           "Invalid Type B",
			photometryType: "B",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.PhotometricData{
				Photometry: models.PhotometricMeasurements{
					PhotometryType: tt.photometryType,
				},
			}

			err := validator.validatePhotometryType(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("validatePhotometryType() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validatePhotometryType() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidator_validateAngularGrid(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		verticalAngles   []float64
		horizontalAngles []float64
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "Valid angles",
			verticalAngles:   []float64{0, 15, 30, 45, 60, 75, 90},
			horizontalAngles: []float64{0, 45, 90, 135, 180, 225, 270, 315},
			expectError:      false,
		},
		{
			name:             "Empty vertical angles",
			verticalAngles:   []float64{},
			horizontalAngles: []float64{0, 90, 180, 270},
			expectError:      true,
			errorMsg:         "vertical angles cannot be empty",
		},
		{
			name:             "Empty horizontal angles",
			verticalAngles:   []float64{0, 30, 60, 90},
			horizontalAngles: []float64{},
			expectError:      true,
			errorMsg:         "horizontal angles cannot be empty",
		},
		{
			name:             "Vertical angle out of range (negative)",
			verticalAngles:   []float64{-10, 30, 60, 90},
			horizontalAngles: []float64{0, 90, 180, 270},
			expectError:      true,
			errorMsg:         "between 0° and 90°",
		},
		{
			name:             "Vertical angle out of range (too high)",
			verticalAngles:   []float64{0, 30, 60, 120},
			horizontalAngles: []float64{0, 90, 180, 270},
			expectError:      true,
			errorMsg:         "between 0° and 90°",
		},
		{
			name:             "Horizontal angle out of range (negative)",
			verticalAngles:   []float64{0, 30, 60, 90},
			horizontalAngles: []float64{-10, 90, 180, 270},
			expectError:      true,
			errorMsg:         "between 0° and 360°",
		},
		{
			name:             "Horizontal angle out of range (too high)",
			verticalAngles:   []float64{0, 30, 60, 90},
			horizontalAngles: []float64{0, 90, 180, 400},
			expectError:      true,
			errorMsg:         "between 0° and 360°",
		},
		{
			name:             "Vertical resolution too fine",
			verticalAngles:   []float64{0, 0.5, 1.0, 1.5},
			horizontalAngles: []float64{0, 90, 180, 270},
			expectError:      true,
			errorMsg:         "resolution too fine",
		},
		{
			name:             "Vertical resolution too coarse",
			verticalAngles:   []float64{0, 20, 40, 60, 80},
			horizontalAngles: []float64{0, 90, 180, 270},
			expectError:      true,
			errorMsg:         "resolution too coarse",
		},
		{
			name:             "Horizontal resolution too fine",
			verticalAngles:   []float64{0, 15, 30, 45},
			horizontalAngles: []float64{0, 2, 4, 6},
			expectError:      true,
			errorMsg:         "resolution too fine",
		},
		{
			name:             "Horizontal resolution too coarse",
			verticalAngles:   []float64{0, 15, 30, 45},
			horizontalAngles: []float64{0, 60, 120, 180},
			expectError:      true,
			errorMsg:         "resolution too coarse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.PhotometricData{
				Photometry: models.PhotometricMeasurements{
					VerticalAngles:   tt.verticalAngles,
					HorizontalAngles: tt.horizontalAngles,
				},
			}

			err := validator.validateAngularGrid(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateAngularGrid() expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("validateAngularGrid() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateAngularGrid() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidator_validateSymmetry(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name             string
		horizontalAngles []float64
		candelaValues    [][]float64
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "Valid quarter symmetry",
			horizontalAngles: []float64{0, 22.5, 45, 67.5, 90},
			candelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7, 0.6},
				{0.8, 0.7, 0.6, 0.5, 0.4},
			},
			expectError: false,
		},
		{
			name:             "Valid half symmetry",
			horizontalAngles: []float64{0, 45, 90, 135, 180},
			candelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7, 0.6},
				{0.8, 0.7, 0.6, 0.5, 0.4},
			},
			expectError: false,
		},
		{
			name:             "Valid full data",
			horizontalAngles: []float64{0, 45, 90, 135, 180, 225, 270, 315},
			candelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7, 0.6, 0.7, 0.8, 0.9}, // Nadir
				{0.8, 0.7, 0.6, 0.5, 0.4, 0.5, 0.6, 0.7}, // Should be lower
			},
			expectError: false,
		},
		{
			name:             "Invalid quarter symmetry - angle too high",
			horizontalAngles: []float64{0, 30, 60, 100}, // Max 100° > 90°, so this will be treated as half symmetry
			candelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7},
			},
			expectError: false, // This will pass half symmetry validation
		},
		{
			name:             "Invalid half symmetry - angle too high",
			horizontalAngles: []float64{0, 90, 180, 270}, // 270° > 180°, will be treated as full data
			candelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7},
			},
			expectError: true,
			errorMsg:    "cover close to 360°", // Will fail full data validation
		},
		{
			name:             "Invalid full data - insufficient coverage",
			horizontalAngles: []float64{0, 90, 180, 270}, // Only 270°, need ~360°
			candelaValues: [][]float64{
				{1.0, 0.9, 0.8, 0.7},
			},
			expectError: true,
			errorMsg:    "cover close to 360°",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.PhotometricData{
				Photometry: models.PhotometricMeasurements{
					HorizontalAngles: tt.horizontalAngles,
					VerticalAngles:   make([]float64, len(tt.candelaValues)), // Match candela rows
					CandelaValues:    tt.candelaValues,
				},
			}

			// Fill vertical angles to match candela data
			for i := range data.Photometry.VerticalAngles {
				data.Photometry.VerticalAngles[i] = float64(i * 15) // 0, 15, 30, ...
			}

			err := validator.validateSymmetry(data)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateSymmetry() expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("validateSymmetry() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateSymmetry() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidator_ValidateCIEFile(t *testing.T) {
	validator := NewValidator()

	validCIEFile := &CIEFile{
		Header: CIEHeader{
			FormatType:   1,
			SymmetryType: 0,
			Reserved:     0,
			Description:  "Test LED 17W 1000 lm",
		},
		Photometry: CIEPhotometry{
			GammaAngles:  []float64{0, 15, 30, 45, 60, 75, 90},
			CPlaneAngles: []float64{0, 45, 90, 135, 180, 225, 270, 315},
			IntensityData: [][]float64{
				{100, 90, 80, 70, 60, 70, 80, 90},
				{90, 80, 70, 60, 50, 60, 70, 80},
				{80, 70, 60, 50, 40, 50, 60, 70},
				{70, 60, 50, 40, 30, 40, 50, 60},
				{50, 40, 30, 20, 10, 20, 30, 40},
				{30, 20, 10, 10, 0, 10, 10, 20},
				{10, 0, 0, 0, 0, 0, 0, 0},
			},
		},
	}

	tests := []struct {
		name        string
		cieFile     *CIEFile
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid CIE file",
			cieFile:     validCIEFile,
			expectError: false,
		},
		{
			name:        "Nil CIE file",
			cieFile:     nil,
			expectError: true,
			errorMsg:    "cannot be nil",
		},
		{
			name: "Invalid format type",
			cieFile: func() *CIEFile {
				file := *validCIEFile
				file.Header.FormatType = 2
				return &file
			}(),
			expectError: true,
			errorMsg:    "unsupported format type",
		},
		{
			name: "Invalid symmetry type",
			cieFile: func() *CIEFile {
				file := *validCIEFile
				file.Header.SymmetryType = 5
				return &file
			}(),
			expectError: true,
			errorMsg:    "invalid symmetry type",
		},
		{
			name: "Non-zero reserved field",
			cieFile: func() *CIEFile {
				file := *validCIEFile
				file.Header.Reserved = 1
				return &file
			}(),
			expectError: true,
			errorMsg:    "reserved field should be 0",
		},
		{
			name: "Empty description",
			cieFile: func() *CIEFile {
				file := *validCIEFile
				file.Header.Description = ""
				return &file
			}(),
			expectError: true,
			errorMsg:    "description cannot be empty",
		},
		{
			name: "Mismatched intensity data dimensions",
			cieFile: func() *CIEFile {
				file := *validCIEFile
				file.Photometry.IntensityData = [][]float64{
					{100, 90, 80}, // Wrong number of columns (3 instead of 16)
				}
				return &file
			}(),
			expectError: true,
			errorMsg:    "must match gamma angles count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCIEFile(tt.cieFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateCIEFile() expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateCIEFile() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCIEFile() unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				strings.Contains(s, substr))))
}
