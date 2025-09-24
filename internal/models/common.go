package models

import (
	"errors"
	"fmt"
	"math"
)

// PhotometricData represents the common internal data structure for all lighting formats
type PhotometricData struct {
	Metadata   LuminaireMetadata       `json:"metadata" validate:"required"`
	Geometry   LuminaireGeometry       `json:"geometry" validate:"required"`
	Photometry PhotometricMeasurements `json:"photometry" validate:"required"`
	Electrical ElectricalData          `json:"electrical" validate:"required"`
}

// Validate performs comprehensive validation of the photometric data
func (pd *PhotometricData) Validate() error {
	if err := pd.Metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}
	if err := pd.Geometry.Validate(); err != nil {
		return fmt.Errorf("geometry validation failed: %w", err)
	}
	if err := pd.Photometry.Validate(); err != nil {
		return fmt.Errorf("photometry validation failed: %w", err)
	}
	if err := pd.Electrical.Validate(); err != nil {
		return fmt.Errorf("electrical validation failed: %w", err)
	}
	return nil
}

// LuminaireMetadata contains general information about the luminaire
type LuminaireMetadata struct {
	Manufacturer  string `json:"manufacturer" validate:"required"`
	CatalogNumber string `json:"catalog_number" validate:"required"`
	Description   string `json:"description"`
	LuminaireType string `json:"luminaire_type"`
	TestLab       string `json:"test_lab"`
	TestDate      string `json:"test_date"`
	TestNumber    string `json:"test_number"`
}

// Validate ensures required metadata fields are present
func (lm *LuminaireMetadata) Validate() error {
	if lm.Manufacturer == "" {
		return errors.New("manufacturer is required")
	}
	if lm.CatalogNumber == "" {
		return errors.New("catalog number is required")
	}
	return nil
}

// LuminaireGeometry contains physical dimensions and mounting information
type LuminaireGeometry struct {
	Length         float64 `json:"length" validate:"min=0"`
	Width          float64 `json:"width" validate:"min=0"`
	Height         float64 `json:"height" validate:"min=0"`
	LuminousLength float64 `json:"luminous_length" validate:"min=0"`
	LuminousWidth  float64 `json:"luminous_width" validate:"min=0"`
	LuminousHeight float64 `json:"luminous_height" validate:"min=0"`
	MountingType   string  `json:"mounting_type"`
}

// Validate ensures geometry values are non-negative and logical
func (lg *LuminaireGeometry) Validate() error {
	if lg.Length < 0 {
		return errors.New("length cannot be negative")
	}
	if lg.Width < 0 {
		return errors.New("width cannot be negative")
	}
	if lg.Height < 0 {
		return errors.New("height cannot be negative")
	}
	if lg.LuminousLength < 0 {
		return errors.New("luminous length cannot be negative")
	}
	if lg.LuminousWidth < 0 {
		return errors.New("luminous width cannot be negative")
	}
	if lg.LuminousHeight < 0 {
		return errors.New("luminous height cannot be negative")
	}

	// Luminous dimensions should not exceed physical dimensions
	if lg.LuminousLength > lg.Length && lg.Length > 0 {
		return errors.New("luminous length cannot exceed physical length")
	}
	if lg.LuminousWidth > lg.Width && lg.Width > 0 {
		return errors.New("luminous width cannot exceed physical width")
	}
	if lg.LuminousHeight > lg.Height && lg.Height > 0 {
		return errors.New("luminous height cannot exceed physical height")
	}

	return nil
}

// PhotometricMeasurements contains the actual light distribution data
type PhotometricMeasurements struct {
	PhotometryType    string      `json:"photometry_type" validate:"required,oneof=A B C"`
	UnitsType         string      `json:"units_type" validate:"required,oneof=absolute relative"`
	LuminousFlux      float64     `json:"luminous_flux" validate:"min=0"`
	CandelaMultiplier float64     `json:"candela_multiplier" validate:"min=0"`
	VerticalAngles    []float64   `json:"vertical_angles" validate:"required,min=1"`
	HorizontalAngles  []float64   `json:"horizontal_angles" validate:"required,min=1"`
	CandelaValues     [][]float64 `json:"candela_values" validate:"required"`
}

// Validate ensures photometric measurements are consistent and valid
func (pm *PhotometricMeasurements) Validate() error {
	// Check required fields
	if pm.PhotometryType == "" {
		return errors.New("photometry type is required")
	}
	if pm.PhotometryType != "A" && pm.PhotometryType != "B" && pm.PhotometryType != "C" {
		return errors.New("photometry type must be A, B, or C")
	}

	if pm.UnitsType == "" {
		return errors.New("units type is required")
	}
	if pm.UnitsType != "absolute" && pm.UnitsType != "relative" {
		return errors.New("units type must be 'absolute' or 'relative'")
	}

	// Check numeric values
	if pm.LuminousFlux < 0 {
		return errors.New("luminous flux cannot be negative")
	}
	if pm.CandelaMultiplier < 0 {
		return errors.New("candela multiplier cannot be negative")
	}

	// Check angle arrays
	if len(pm.VerticalAngles) == 0 {
		return errors.New("vertical angles array cannot be empty")
	}
	if len(pm.HorizontalAngles) == 0 {
		return errors.New("horizontal angles array cannot be empty")
	}

	// Validate angle ranges
	for i, angle := range pm.VerticalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("vertical angle at index %d (%f) must be between 0 and 180 degrees", i, angle)
		}
		if i > 0 && angle <= pm.VerticalAngles[i-1] {
			return fmt.Errorf("vertical angles must be in ascending order")
		}
	}

	for i, angle := range pm.HorizontalAngles {
		if angle < 0 || angle > 360 {
			return fmt.Errorf("horizontal angle at index %d (%f) must be between 0 and 360 degrees", i, angle)
		}
		if i > 0 && angle <= pm.HorizontalAngles[i-1] {
			return fmt.Errorf("horizontal angles must be in ascending order")
		}
	}

	// Check candela values array dimensions
	if len(pm.CandelaValues) != len(pm.VerticalAngles) {
		return fmt.Errorf("candela values array length (%d) must match vertical angles length (%d)",
			len(pm.CandelaValues), len(pm.VerticalAngles))
	}

	for i, row := range pm.CandelaValues {
		if len(row) != len(pm.HorizontalAngles) {
			return fmt.Errorf("candela values row %d length (%d) must match horizontal angles length (%d)",
				i, len(row), len(pm.HorizontalAngles))
		}

		// Check for valid candela values
		for j, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("invalid candela value at [%d][%d]: %f", i, j, value)
			}
			if value < 0 {
				return fmt.Errorf("candela value at [%d][%d] cannot be negative: %f", i, j, value)
			}
		}
	}

	return nil
}

// ElectricalData contains electrical characteristics
type ElectricalData struct {
	InputWatts        float64 `json:"input_watts" validate:"min=0"`
	BallastFactor     float64 `json:"ballast_factor" validate:"min=0,max=2"`
	BallastLampFactor float64 `json:"ballast_lamp_factor" validate:"min=0,max=2"`
	InputVoltage      float64 `json:"input_voltage" validate:"min=0"`
	InputCurrent      float64 `json:"input_current" validate:"min=0"`
	PowerFactor       float64 `json:"power_factor" validate:"min=0,max=1"`
}

// Validate ensures electrical data values are within reasonable ranges
func (ed *ElectricalData) Validate() error {
	if ed.InputWatts < 0 {
		return errors.New("input watts cannot be negative")
	}
	if ed.BallastFactor < 0 || ed.BallastFactor > 2 {
		return errors.New("ballast factor must be between 0 and 2")
	}
	if ed.BallastLampFactor < 0 || ed.BallastLampFactor > 2 {
		return errors.New("ballast lamp factor must be between 0 and 2")
	}
	if ed.InputVoltage < 0 {
		return errors.New("input voltage cannot be negative")
	}
	if ed.InputCurrent < 0 {
		return errors.New("input current cannot be negative")
	}
	if ed.PowerFactor < 0 || ed.PowerFactor > 1 {
		return errors.New("power factor must be between 0 and 1")
	}

	// Basic electrical validation: P = V * I * PF
	if ed.InputVoltage > 0 && ed.InputCurrent > 0 && ed.PowerFactor > 0 {
		calculatedWatts := ed.InputVoltage * ed.InputCurrent * ed.PowerFactor
		tolerance := 0.1 // 10% tolerance
		if ed.InputWatts > 0 && math.Abs(ed.InputWatts-calculatedWatts)/ed.InputWatts > tolerance {
			return fmt.Errorf("electrical values inconsistent: watts=%f, calculated=%f (V*I*PF)",
				ed.InputWatts, calculatedWatts)
		}
	}

	return nil
}
