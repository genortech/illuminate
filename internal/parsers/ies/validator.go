package ies

import (
	"fmt"
	"illuminate/internal/models"
	"math"
)

// Validator provides IES-specific validation functionality
type Validator struct{}

// NewValidator creates a new IES validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs comprehensive IES-specific validation
func (v *Validator) Validate(data *models.PhotometricData) error {
	if data == nil {
		return fmt.Errorf("photometric data is nil")
	}

	// Validate basic structure first
	if err := data.Validate(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// IES-specific validations
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

// validatePhotometryType validates the photometry type for IES compliance
func (v *Validator) validatePhotometryType(data *models.PhotometricData) error {
	switch data.Photometry.PhotometryType {
	case "A", "B", "C":
		// Valid types
	default:
		return fmt.Errorf("invalid photometry type '%s': must be A, B, or C", data.Photometry.PhotometryType)
	}

	// Validate angle ranges based on photometry type
	switch data.Photometry.PhotometryType {
	case "A":
		// Type A: vertical angles 0-180°, horizontal angles 0-360°
		if err := v.validateTypeAAngles(data); err != nil {
			return err
		}
	case "B":
		// Type B: vertical angles 0-180°, horizontal angles 0-180°
		if err := v.validateTypeBAngles(data); err != nil {
			return err
		}
	case "C":
		// Type C: vertical angles 0-180°, horizontal angles 0-360°
		if err := v.validateTypeCAngles(data); err != nil {
			return err
		}
	}

	return nil
}

// validateTypeAAngles validates angles for Type A photometry
func (v *Validator) validateTypeAAngles(data *models.PhotometricData) error {
	// Type A: Vertical angles 0-180°, Horizontal angles 0-360°
	for i, angle := range data.Photometry.VerticalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("Type A vertical angle at index %d (%f°) must be between 0° and 180°", i, angle)
		}
	}

	for i, angle := range data.Photometry.HorizontalAngles {
		if angle < 0 || angle > 360 {
			return fmt.Errorf("Type A horizontal angle at index %d (%f°) must be between 0° and 360°", i, angle)
		}
	}

	return nil
}

// validateTypeBAngles validates angles for Type B photometry
func (v *Validator) validateTypeBAngles(data *models.PhotometricData) error {
	// Type B: Vertical angles 0-180°, Horizontal angles 0-180°
	for i, angle := range data.Photometry.VerticalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("Type B vertical angle at index %d (%f°) must be between 0° and 180°", i, angle)
		}
	}

	for i, angle := range data.Photometry.HorizontalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("Type B horizontal angle at index %d (%f°) must be between 0° and 180°", i, angle)
		}
	}

	return nil
}

// validateTypeCAngles validates angles for Type C photometry
func (v *Validator) validateTypeCAngles(data *models.PhotometricData) error {
	// Type C: Vertical angles 0-180°, Horizontal angles 0-360°
	for i, angle := range data.Photometry.VerticalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("Type C vertical angle at index %d (%f°) must be between 0° and 180°", i, angle)
		}
	}

	for i, angle := range data.Photometry.HorizontalAngles {
		if angle < 0 || angle > 360 {
			return fmt.Errorf("Type C horizontal angle at index %d (%f°) must be between 0° and 360°", i, angle)
		}
	}

	return nil
}

// validateAngles validates angle arrays for IES compliance
func (v *Validator) validateAngles(data *models.PhotometricData) error {
	// Check minimum angle requirements
	if len(data.Photometry.VerticalAngles) < 2 {
		return fmt.Errorf("minimum 2 vertical angles required, got %d", len(data.Photometry.VerticalAngles))
	}

	if len(data.Photometry.HorizontalAngles) < 1 {
		return fmt.Errorf("minimum 1 horizontal angle required, got %d", len(data.Photometry.HorizontalAngles))
	}

	// Check maximum angle limits (relaxed for modern IES files)
	if len(data.Photometry.VerticalAngles) > 181 {
		return fmt.Errorf("maximum 181 vertical angles allowed, got %d", len(data.Photometry.VerticalAngles))
	}

	if len(data.Photometry.HorizontalAngles) > 361 {
		return fmt.Errorf("maximum 361 horizontal angles allowed, got %d", len(data.Photometry.HorizontalAngles))
	}

	// Validate angle increments (should be reasonable)
	if err := v.validateAngleIncrements(data.Photometry.VerticalAngles, "vertical"); err != nil {
		return err
	}

	if err := v.validateAngleIncrements(data.Photometry.HorizontalAngles, "horizontal"); err != nil {
		return err
	}

	return nil
}

// validateAngleIncrements validates that angle increments are reasonable
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

// validateCandelaValues validates the candela values matrix
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

			// Check for unreasonably high values (> 1 million cd)
			if value > 1000000 {
				return fmt.Errorf("unreasonably high candela value at [%d][%d]: %f cd", i, j, value)
			}
		}
	}

	// Check for reasonable dynamic range
	if maxCandela > 0 && minCandela >= 0 {
		dynamicRange := maxCandela / math.Max(minCandela, 0.001) // Avoid division by zero
		if dynamicRange > 1000000 {
			return fmt.Errorf("candela dynamic range too high: %f (max=%f, min=%f)",
				dynamicRange, maxCandela, minCandela)
		}
	}

	return nil
}

// validateLuminousFlux validates the luminous flux value
func (v *Validator) validateLuminousFlux(data *models.PhotometricData) error {
	if data.Photometry.LuminousFlux < 0 {
		return fmt.Errorf("luminous flux cannot be negative: %f", data.Photometry.LuminousFlux)
	}

	// Check for reasonable flux values (0.1 lm to 1,000,000 lm)
	if data.Photometry.LuminousFlux > 0 && data.Photometry.LuminousFlux < 0.1 {
		return fmt.Errorf("luminous flux too low: %f lm", data.Photometry.LuminousFlux)
	}

	if data.Photometry.LuminousFlux > 1000000 {
		return fmt.Errorf("luminous flux too high: %f lm", data.Photometry.LuminousFlux)
	}

	// Validate candela multiplier
	if data.Photometry.CandelaMultiplier <= 0 {
		return fmt.Errorf("candela multiplier must be positive: %f", data.Photometry.CandelaMultiplier)
	}

	if data.Photometry.CandelaMultiplier > 1000000 {
		return fmt.Errorf("candela multiplier too high: %f", data.Photometry.CandelaMultiplier)
	}

	return nil
}

// validateGeometry validates luminaire geometry for IES compliance
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

	// Check for reasonable dimension limits (0.001m to 100m)
	maxDimension := 100.0 // 100 meters
	minDimension := 0.001 // 1 millimeter

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

	return nil
}

// validateElectricalData validates electrical parameters for IES compliance
func (v *Validator) validateElectricalData(data *models.PhotometricData) error {
	// Input watts should be reasonable
	if data.Electrical.InputWatts < 0 {
		return fmt.Errorf("input watts cannot be negative: %f", data.Electrical.InputWatts)
	}

	if data.Electrical.InputWatts > 100000 {
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

	return nil
}
