package ldt

import (
	"fmt"
	"illuminate/internal/models"
	"math"
)

// Validator provides LDT-specific validation functionality
type Validator struct{}

// NewValidator creates a new LDT validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs comprehensive LDT-specific validation
func (v *Validator) Validate(data *models.PhotometricData) error {
	if data == nil {
		return fmt.Errorf("photometric data is nil")
	}

	// Validate basic structure first
	if err := data.Validate(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// LDT-specific validations
	if err := v.validatePhotometryType(data); err != nil {
		return fmt.Errorf("photometry type validation failed: %w", err)
	}

	if err := v.validateAngles(data); err != nil {
		return fmt.Errorf("angle validation failed: %w", err)
	}

	if err := v.validateCandelaValues(data); err != nil {
		return fmt.Errorf("candela values validation failed: %w", err)
	}

	if err := v.validateLuminousFlux(data); err != nil {
		return fmt.Errorf("luminous flux validation failed: %w", err)
	}

	if err := v.validateGeometry(data); err != nil {
		return fmt.Errorf("geometry validation failed: %w", err)
	}

	if err := v.validateElectricalData(data); err != nil {
		return fmt.Errorf("electrical data validation failed: %w", err)
	}

	return nil
}

// validatePhotometryType validates the photometry type for LDT compliance
func (v *Validator) validatePhotometryType(data *models.PhotometricData) error {
	// LDT format primarily supports Type C photometry
	switch data.Photometry.PhotometryType {
	case "C":
		// Valid for LDT
	case "A", "B":
		// Can be converted to Type C representation
	default:
		return fmt.Errorf("invalid photometry type '%s': LDT format primarily supports Type C", data.Photometry.PhotometryType)
	}

	// Validate angle ranges for LDT format
	if err := v.validateLDTAngles(data); err != nil {
		return err
	}

	return nil
}

// validateLDTAngles validates angles for LDT format compliance
func (v *Validator) validateLDTAngles(data *models.PhotometricData) error {
	// LDT format: Gamma angles 0-180°, C-plane angles 0-360°
	for i, angle := range data.Photometry.VerticalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("gamma angle at index %d (%f°) must be between 0° and 180°", i, angle)
		}
	}

	for i, angle := range data.Photometry.HorizontalAngles {
		if angle < 0 || angle > 360 {
			return fmt.Errorf("C-plane angle at index %d (%f°) must be between 0° and 360°", i, angle)
		}
	}

	return nil
}

// validateAngles validates angle arrays for LDT compliance
func (v *Validator) validateAngles(data *models.PhotometricData) error {
	// Check minimum angle requirements
	if len(data.Photometry.VerticalAngles) < 2 {
		return fmt.Errorf("minimum 2 gamma angles required, got %d", len(data.Photometry.VerticalAngles))
	}

	if len(data.Photometry.HorizontalAngles) < 1 {
		return fmt.Errorf("minimum 1 C-plane angle required, got %d", len(data.Photometry.HorizontalAngles))
	}

	// Check maximum angle limits for LDT format
	if len(data.Photometry.VerticalAngles) > 181 {
		return fmt.Errorf("maximum 181 gamma angles allowed in LDT format, got %d", len(data.Photometry.VerticalAngles))
	}

	if len(data.Photometry.HorizontalAngles) > 361 {
		return fmt.Errorf("maximum 361 C-plane angles allowed in LDT format, got %d", len(data.Photometry.HorizontalAngles))
	}

	// Validate angle increments
	if err := v.validateAngleIncrements(data.Photometry.VerticalAngles, "gamma"); err != nil {
		return err
	}

	if err := v.validateAngleIncrements(data.Photometry.HorizontalAngles, "C-plane"); err != nil {
		return err
	}

	// Validate angle ranges
	if err := v.validateAngleRanges(data); err != nil {
		return err
	}

	return nil
}

// validateAngleIncrements validates that angle increments are reasonable for LDT
func (v *Validator) validateAngleIncrements(angles []float64, angleType string) error {
	if len(angles) < 2 {
		return nil // Skip validation for single angle
	}

	for i := 1; i < len(angles); i++ {
		increment := angles[i] - angles[i-1]

		// Check for negative increments (angles should be ascending)
		if increment <= 0 {
			return fmt.Errorf("%s angles must be in ascending order: angle[%d]=%f, angle[%d]=%f",
				angleType, i-1, angles[i-1], i, angles[i])
		}

		// Check for unreasonably small increments
		if increment < 0.1 {
			return fmt.Errorf("%s angle increment too small: %f° between indices %d and %d",
				angleType, increment, i-1, i)
		}

		// Check for unreasonably large increments
		if increment > 90 {
			return fmt.Errorf("%s angle increment too large: %f° between indices %d and %d",
				angleType, increment, i-1, i)
		}
	}

	return nil
}

// validateAngleRanges validates that angles are within expected ranges
func (v *Validator) validateAngleRanges(data *models.PhotometricData) error {
	// Gamma angles should start at 0 and not exceed 180
	if len(data.Photometry.VerticalAngles) > 0 {
		firstGamma := data.Photometry.VerticalAngles[0]
		lastGamma := data.Photometry.VerticalAngles[len(data.Photometry.VerticalAngles)-1]

		if firstGamma != 0 {
			return fmt.Errorf("first gamma angle should be 0°, got %f°", firstGamma)
		}

		if lastGamma > 180 {
			return fmt.Errorf("last gamma angle should not exceed 180°, got %f°", lastGamma)
		}
	}

	// C-plane angles should be within 0-360 range
	if len(data.Photometry.HorizontalAngles) > 0 {
		firstCPlane := data.Photometry.HorizontalAngles[0]
		lastCPlane := data.Photometry.HorizontalAngles[len(data.Photometry.HorizontalAngles)-1]

		if firstCPlane < 0 {
			return fmt.Errorf("first C-plane angle should not be negative, got %f°", firstCPlane)
		}

		if lastCPlane > 360 {
			return fmt.Errorf("last C-plane angle should not exceed 360°, got %f°", lastCPlane)
		}
	}

	return nil
}

// validateCandelaValues validates the candela values matrix for LDT compliance
func (v *Validator) validateCandelaValues(data *models.PhotometricData) error {
	// Check for reasonable candela value ranges
	maxCandela := 0.0
	minCandela := math.Inf(1)

	for i, row := range data.Photometry.CandelaValues {
		for j, value := range row {
			if value < 0 {
				return fmt.Errorf("negative candela value at [%d][%d]: %f", i, j, value)
			}

			if value > maxCandela {
				maxCandela = value
			}
			if value < minCandela {
				minCandela = value
			}

			// Check for unreasonably high values (> 10 million cd for LDT)
			if value > 10000000 {
				return fmt.Errorf("unreasonably high candela value at [%d][%d]: %f cd", i, j, value)
			}
		}
	}

	// Check for reasonable dynamic range
	if maxCandela > 0 && minCandela >= 0 {
		dynamicRange := maxCandela / math.Max(minCandela, 0.001) // Avoid division by zero
		if dynamicRange > 10000000 {
			return fmt.Errorf("candela dynamic range too high: %f (max=%f, min=%f)",
				dynamicRange, maxCandela, minCandela)
		}
	}

	// Validate symmetry if applicable
	if err := v.validateSymmetry(data); err != nil {
		return err
	}

	return nil
}

// validateSymmetry validates symmetry in candela values if applicable
func (v *Validator) validateSymmetry(data *models.PhotometricData) error {
	// This is a simplified symmetry check
	// In practice, LDT files can have various symmetry types

	numGamma := len(data.Photometry.VerticalAngles)
	numCPlane := len(data.Photometry.HorizontalAngles)

	if numGamma == 0 || numCPlane == 0 {
		return nil // Skip if no data
	}

	// Check for obvious data inconsistencies
	for i := 0; i < numGamma; i++ {
		for j := 0; j < numCPlane; j++ {
			value := data.Photometry.CandelaValues[i][j]

			// Check for NaN or infinite values
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("invalid candela value at [%d][%d]: %f", i, j, value)
			}
		}
	}

	return nil
}

// validateLuminousFlux validates the luminous flux value for LDT compliance
func (v *Validator) validateLuminousFlux(data *models.PhotometricData) error {
	if data.Photometry.LuminousFlux < 0 {
		return fmt.Errorf("luminous flux cannot be negative: %f", data.Photometry.LuminousFlux)
	}

	// Check for reasonable flux values for LDT format (0.1 lm to 10,000,000 lm)
	if data.Photometry.LuminousFlux > 0 && data.Photometry.LuminousFlux < 0.1 {
		return fmt.Errorf("luminous flux too low: %f lm", data.Photometry.LuminousFlux)
	}

	if data.Photometry.LuminousFlux > 10000000 {
		return fmt.Errorf("luminous flux too high: %f lm", data.Photometry.LuminousFlux)
	}

	// Validate candela multiplier (conversion factor)
	if data.Photometry.CandelaMultiplier <= 0 {
		return fmt.Errorf("candela multiplier must be positive: %f", data.Photometry.CandelaMultiplier)
	}

	if data.Photometry.CandelaMultiplier > 10000000 {
		return fmt.Errorf("candela multiplier too high: %f", data.Photometry.CandelaMultiplier)
	}

	return nil
}

// validateGeometry validates luminaire geometry for LDT compliance
func (v *Validator) validateGeometry(data *models.PhotometricData) error {
	// All dimensions should be non-negative
	if data.Geometry.Length < 0 {
		return fmt.Errorf("luminaire length cannot be negative: %f", data.Geometry.Length)
	}
	if data.Geometry.Width < 0 {
		return fmt.Errorf("luminaire width cannot be negative: %f", data.Geometry.Width)
	}
	if data.Geometry.Height < 0 {
		return fmt.Errorf("luminaire height cannot be negative: %f", data.Geometry.Height)
	}

	// Check for reasonable dimension limits (0.001m to 1000m for LDT)
	maxDimension := 1000.0 // 1000 meters
	minDimension := 0.001  // 1 millimeter

	if data.Geometry.Length > maxDimension {
		return fmt.Errorf("luminaire length too large: %f m", data.Geometry.Length)
	}
	if data.Geometry.Width > maxDimension {
		return fmt.Errorf("luminaire width too large: %f m", data.Geometry.Width)
	}
	if data.Geometry.Height > maxDimension {
		return fmt.Errorf("luminaire height too large: %f m", data.Geometry.Height)
	}

	// Only check minimum if dimensions are specified (non-zero)
	if data.Geometry.Length > 0 && data.Geometry.Length < minDimension {
		return fmt.Errorf("luminaire length too small: %f m", data.Geometry.Length)
	}
	if data.Geometry.Width > 0 && data.Geometry.Width < minDimension {
		return fmt.Errorf("luminaire width too small: %f m", data.Geometry.Width)
	}
	if data.Geometry.Height > 0 && data.Geometry.Height < minDimension {
		return fmt.Errorf("luminaire height too small: %f m", data.Geometry.Height)
	}

	// Validate luminous area dimensions
	if data.Geometry.LuminousLength > data.Geometry.Length && data.Geometry.Length > 0 {
		return fmt.Errorf("luminous length (%f m) cannot exceed physical length (%f m)",
			data.Geometry.LuminousLength, data.Geometry.Length)
	}
	if data.Geometry.LuminousWidth > data.Geometry.Width && data.Geometry.Width > 0 {
		return fmt.Errorf("luminous width (%f m) cannot exceed physical width (%f m)",
			data.Geometry.LuminousWidth, data.Geometry.Width)
	}
	if data.Geometry.LuminousHeight > data.Geometry.Height && data.Geometry.Height > 0 {
		return fmt.Errorf("luminous height (%f m) cannot exceed physical height (%f m)",
			data.Geometry.LuminousHeight, data.Geometry.Height)
	}

	return nil
}

// validateElectricalData validates electrical parameters for LDT compliance
func (v *Validator) validateElectricalData(data *models.PhotometricData) error {
	// Input watts should be reasonable
	if data.Electrical.InputWatts < 0 {
		return fmt.Errorf("input watts cannot be negative: %f", data.Electrical.InputWatts)
	}

	if data.Electrical.InputWatts > 1000000 {
		return fmt.Errorf("input watts too high: %f W", data.Electrical.InputWatts)
	}

	// Ballast factor should be reasonable (typically 0.5 to 2.0)
	if data.Electrical.BallastFactor < 0 {
		return fmt.Errorf("ballast factor cannot be negative: %f", data.Electrical.BallastFactor)
	}

	if data.Electrical.BallastFactor > 5.0 {
		return fmt.Errorf("ballast factor too high: %f", data.Electrical.BallastFactor)
	}

	// Ballast lamp factor should be reasonable
	if data.Electrical.BallastLampFactor < 0 {
		return fmt.Errorf("ballast lamp factor cannot be negative: %f", data.Electrical.BallastLampFactor)
	}

	if data.Electrical.BallastLampFactor > 5.0 {
		return fmt.Errorf("ballast lamp factor too high: %f", data.Electrical.BallastLampFactor)
	}

	// Validate power factor if specified
	if data.Electrical.PowerFactor > 0 {
		if data.Electrical.PowerFactor > 1.0 {
			return fmt.Errorf("power factor cannot exceed 1.0: %f", data.Electrical.PowerFactor)
		}
	}

	return nil
}
