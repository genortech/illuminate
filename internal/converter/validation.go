package converter

import (
	"fmt"
	"math"
	"strings"

	"illuminate/internal/interfaces"
	"illuminate/internal/models"
)

// addError adds an error to the validation result
func addError(vr *interfaces.ValidationResult, err string) {
	vr.Errors = append(vr.Errors, err)
	vr.IsValid = false
}

// addWarning adds a warning to the validation result
func addWarning(vr *interfaces.ValidationResult, warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// hasErrors returns true if there are validation errors
func hasErrors(vr *interfaces.ValidationResult) bool {
	return len(vr.Errors) > 0
}

// hasWarnings returns true if there are validation warnings
func hasWarnings(vr *interfaces.ValidationResult) bool {
	return len(vr.Warnings) > 0
}

// Validator provides comprehensive validation for photometric data
type Validator struct {
	logger *Logger
}

// NewValidator creates a new validator instance
func NewValidator(logger *Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidatePhotometricData performs comprehensive validation of photometric data
func (v *Validator) ValidatePhotometricData(data *models.PhotometricData) *interfaces.ValidationResult {
	result := &interfaces.ValidationResult{IsValid: true}

	// Basic model validation
	if err := data.Validate(); err != nil {
		addError(result, fmt.Sprintf("Model validation failed: %s", err.Error()))
	}

	// Cross-format compatibility checks
	v.validatePhotometryTypeCompatibility(data, result)
	v.validateGeometryConsistency(data, result)
	v.validateElectricalConsistency(data, result)
	v.validatePhotometricConsistency(data, result)

	// Performance and quality checks
	v.validateDataQuality(data, result)

	return result
}

// validatePhotometryTypeCompatibility checks if photometry type is compatible across formats
func (v *Validator) validatePhotometryTypeCompatibility(data *models.PhotometricData, result *interfaces.ValidationResult) {
	switch data.Photometry.PhotometryType {
	case "A":
		// Type A: Plane perpendicular to lamp axis
		if len(data.Photometry.HorizontalAngles) > 1 {
			addWarning(result, "Type A photometry typically uses single horizontal plane")
		}
	case "B":
		// Type B: Plane containing lamp axis
		if len(data.Photometry.HorizontalAngles) < 2 {
			addWarning(result, "Type B photometry should have multiple horizontal angles")
		}
	case "C":
		// Type C: Plane perpendicular to lamp axis (most common)
		if len(data.Photometry.HorizontalAngles) < 2 {
			addWarning(result, "Type C photometry should have multiple horizontal angles")
		}
	default:
		addError(result, fmt.Sprintf("Unknown photometry type: %s", data.Photometry.PhotometryType))
	}
}

// validateGeometryConsistency checks geometry parameter consistency
func (v *Validator) validateGeometryConsistency(data *models.PhotometricData, result *interfaces.ValidationResult) {
	geom := &data.Geometry

	// Check for reasonable dimensions
	if geom.Length > 10000 || geom.Width > 10000 || geom.Height > 10000 {
		addWarning(result, "Luminaire dimensions seem unusually large (>10m)")
	}

	if geom.Length < 0.001 && geom.Length > 0 {
		addWarning(result, "Luminaire length seems unusually small (<1mm)")
	}

	// Check luminous vs physical dimensions consistency
	if geom.LuminousLength > 0 && geom.Length > 0 {
		ratio := geom.LuminousLength / geom.Length
		if ratio > 1.1 {
			addError(result, "Luminous length significantly exceeds physical length")
		} else if ratio > 1.0 {
			addWarning(result, "Luminous length slightly exceeds physical length")
		}
	}

	// Similar checks for width and height
	if geom.LuminousWidth > 0 && geom.Width > 0 {
		ratio := geom.LuminousWidth / geom.Width
		if ratio > 1.1 {
			addError(result, "Luminous width significantly exceeds physical width")
		} else if ratio > 1.0 {
			addWarning(result, "Luminous width slightly exceeds physical width")
		}
	}

	if geom.LuminousHeight > 0 && geom.Height > 0 {
		ratio := geom.LuminousHeight / geom.Height
		if ratio > 1.1 {
			addError(result, "Luminous height significantly exceeds physical height")
		} else if ratio > 1.0 {
			addWarning(result, "Luminous height slightly exceeds physical height")
		}
	}
}

// validateElectricalConsistency checks electrical parameter consistency
func (v *Validator) validateElectricalConsistency(data *models.PhotometricData, result *interfaces.ValidationResult) {
	elec := &data.Electrical

	// Check for reasonable electrical values
	if elec.InputWatts > 10000 {
		addWarning(result, "Input watts seems unusually high (>10kW)")
	}

	if elec.InputVoltage > 1000 {
		addWarning(result, "Input voltage seems unusually high (>1000V)")
	}

	if elec.InputCurrent > 100 {
		addWarning(result, "Input current seems unusually high (>100A)")
	}

	// Check power factor reasonableness
	if elec.PowerFactor > 0 && elec.PowerFactor < 0.5 {
		addWarning(result, "Power factor is quite low (<0.5)")
	}

	// Check ballast factors
	if elec.BallastFactor > 0 && (elec.BallastFactor < 0.5 || elec.BallastFactor > 1.5) {
		addWarning(result, "Ballast factor outside typical range (0.5-1.5)")
	}

	if elec.BallastLampFactor > 0 && (elec.BallastLampFactor < 0.5 || elec.BallastLampFactor > 1.5) {
		addWarning(result, "Ballast lamp factor outside typical range (0.5-1.5)")
	}
}

// validatePhotometricConsistency checks photometric data consistency
func (v *Validator) validatePhotometricConsistency(data *models.PhotometricData, result *interfaces.ValidationResult) {
	photo := &data.Photometry

	// Check luminous flux consistency with candela values
	if photo.LuminousFlux > 0 {
		calculatedFlux := v.calculateLuminousFluxFromCandela(photo)
		if calculatedFlux > 0 {
			ratio := photo.LuminousFlux / calculatedFlux
			if ratio < 0.8 || ratio > 1.2 {
				addWarning(result, fmt.Sprintf("Declared luminous flux (%.0f lm) differs significantly from calculated flux (%.0f lm)",
					photo.LuminousFlux, calculatedFlux))
			}
		}
	}

	// Check for symmetry in candela distribution
	v.validateSymmetry(photo, result)

	// Check for reasonable candela values
	v.validateCandelaValues(photo, result)
}

// calculateLuminousFluxFromCandela estimates luminous flux from candela distribution
func (v *Validator) calculateLuminousFluxFromCandela(photo *models.PhotometricMeasurements) float64 {
	if len(photo.CandelaValues) == 0 || len(photo.CandelaValues[0]) == 0 {
		return 0
	}

	// Simplified integration using trapezoidal rule
	// This is an approximation - real calculation would be more complex
	totalFlux := 0.0

	for i := 0; i < len(photo.VerticalAngles); i++ {
		for j := 0; j < len(photo.HorizontalAngles); j++ {
			if i < len(photo.CandelaValues) && j < len(photo.CandelaValues[i]) {
				candela := photo.CandelaValues[i][j] * photo.CandelaMultiplier

				// Convert angles to radians
				theta := photo.VerticalAngles[i] * math.Pi / 180

				// Solid angle element (simplified)
				var dTheta, dPhi float64
				if i > 0 {
					dTheta = (photo.VerticalAngles[i] - photo.VerticalAngles[i-1]) * math.Pi / 180
				} else if i < len(photo.VerticalAngles)-1 {
					dTheta = (photo.VerticalAngles[i+1] - photo.VerticalAngles[i]) * math.Pi / 180
				}

				if j > 0 {
					dPhi = (photo.HorizontalAngles[j] - photo.HorizontalAngles[j-1]) * math.Pi / 180
				} else if j < len(photo.HorizontalAngles)-1 {
					dPhi = (photo.HorizontalAngles[j+1] - photo.HorizontalAngles[j]) * math.Pi / 180
				}

				solidAngle := math.Sin(theta) * dTheta * dPhi
				totalFlux += candela * solidAngle
			}
		}
	}

	return totalFlux
}

// validateSymmetry checks for expected symmetry in candela distribution
func (v *Validator) validateSymmetry(photo *models.PhotometricMeasurements, result *interfaces.ValidationResult) {
	// Check for horizontal symmetry (common in many luminaires)
	if len(photo.HorizontalAngles) >= 4 {
		asymmetryCount := 0
		tolerance := 0.1 // 10% tolerance

		for i := 0; i < len(photo.VerticalAngles); i++ {
			if i < len(photo.CandelaValues) {
				row := photo.CandelaValues[i]
				midPoint := len(row) / 2

				for j := 0; j < midPoint; j++ {
					oppositeJ := len(row) - 1 - j
					if oppositeJ < len(row) {
						val1 := row[j]
						val2 := row[oppositeJ]

						if val1 > 0 && val2 > 0 {
							diff := math.Abs(val1-val2) / math.Max(val1, val2)
							if diff > tolerance {
								asymmetryCount++
							}
						}
					}
				}
			}
		}

		totalComparisons := len(photo.VerticalAngles) * (len(photo.HorizontalAngles) / 2)
		if totalComparisons > 0 {
			asymmetryRatio := float64(asymmetryCount) / float64(totalComparisons)
			if asymmetryRatio > 0.3 {
				addWarning(result, "Significant asymmetry detected in candela distribution")
			}
		}
	}
}

// validateCandelaValues checks for reasonable candela values
func (v *Validator) validateCandelaValues(photo *models.PhotometricMeasurements, result *interfaces.ValidationResult) {
	maxCandela := 0.0
	minCandela := math.Inf(1)
	zeroCount := 0
	totalValues := 0

	for i := 0; i < len(photo.CandelaValues); i++ {
		for j := 0; j < len(photo.CandelaValues[i]); j++ {
			value := photo.CandelaValues[i][j] * photo.CandelaMultiplier
			totalValues++

			if value == 0 {
				zeroCount++
			} else {
				if value > maxCandela {
					maxCandela = value
				}
				if value < minCandela {
					minCandela = value
				}
			}
		}
	}

	// Check for reasonable candela range
	if maxCandela > 100000 {
		addWarning(result, "Maximum candela value seems unusually high (>100,000 cd)")
	}

	if minCandela < 0.001 && minCandela > 0 {
		addWarning(result, "Minimum candela value seems unusually low (<0.001 cd)")
	}

	// Check for excessive zero values
	if totalValues > 0 {
		zeroRatio := float64(zeroCount) / float64(totalValues)
		if zeroRatio > 0.5 {
			addWarning(result, "More than 50% of candela values are zero")
		}
	}

	// Check dynamic range
	if maxCandela > 0 && minCandela > 0 {
		dynamicRange := maxCandela / minCandela
		if dynamicRange > 10000 {
			addWarning(result, "Very high dynamic range in candela values (>10,000:1)")
		}
	}
}

// validateDataQuality performs additional data quality checks
func (v *Validator) validateDataQuality(data *models.PhotometricData, result *interfaces.ValidationResult) {
	// Check metadata completeness
	if strings.TrimSpace(data.Metadata.Manufacturer) == "" {
		addWarning(result, "Manufacturer information is missing")
	}

	if strings.TrimSpace(data.Metadata.CatalogNumber) == "" {
		addWarning(result, "Catalog number is missing")
	}

	// Check for reasonable angle increments
	photo := &data.Photometry

	if len(photo.VerticalAngles) > 1 {
		minIncrement := math.Inf(1)
		maxIncrement := 0.0

		for i := 1; i < len(photo.VerticalAngles); i++ {
			increment := photo.VerticalAngles[i] - photo.VerticalAngles[i-1]
			if increment < minIncrement {
				minIncrement = increment
			}
			if increment > maxIncrement {
				maxIncrement = increment
			}
		}

		if maxIncrement/minIncrement > 10 {
			addWarning(result, "Inconsistent vertical angle increments")
		}
	}

	if len(photo.HorizontalAngles) > 1 {
		minIncrement := math.Inf(1)
		maxIncrement := 0.0

		for i := 1; i < len(photo.HorizontalAngles); i++ {
			increment := photo.HorizontalAngles[i] - photo.HorizontalAngles[i-1]
			if increment < minIncrement {
				minIncrement = increment
			}
			if increment > maxIncrement {
				maxIncrement = increment
			}
		}

		if maxIncrement/minIncrement > 10 {
			addWarning(result, "Inconsistent horizontal angle increments")
		}
	}
}
