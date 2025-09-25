package cie

import (
	"errors"
	"fmt"
	"illuminate/internal/models"
	"strings"
)

// Validator provides CIE format-specific validation
type Validator struct{}

// NewValidator creates a new CIE validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs CIE format-specific validation on photometric data
func (v *Validator) Validate(data *models.PhotometricData) error {
	if data == nil {
		return errors.New("photometric data cannot be nil")
	}

	// Perform basic validation first
	if err := data.Validate(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// CIE-specific validations
	if err := v.validatePhotometryType(data); err != nil {
		return fmt.Errorf("photometry type validation failed: %w", err)
	}

	if err := v.validateAngularGrid(data); err != nil {
		return fmt.Errorf("angular grid validation failed: %w", err)
	}

	if err := v.validateIntensityData(data); err != nil {
		return fmt.Errorf("intensity data validation failed: %w", err)
	}

	if err := v.validateSymmetry(data); err != nil {
		return fmt.Errorf("symmetry validation failed: %w", err)
	}

	return nil
}

// validatePhotometryType ensures the photometry type is compatible with CIE format
func (v *Validator) validatePhotometryType(data *models.PhotometricData) error {
	// CIE i-table format typically uses Type C photometry
	if data.Photometry.PhotometryType != "C" {
		return fmt.Errorf("CIE format typically requires Type C photometry, got: %s", data.Photometry.PhotometryType)
	}

	return nil
}

// validateAngularGrid validates the angular grid for CIE compatibility
func (v *Validator) validateAngularGrid(data *models.PhotometricData) error {
	// Check vertical angles (gamma angles)
	if len(data.Photometry.VerticalAngles) == 0 {
		return errors.New("vertical angles cannot be empty")
	}

	// Vertical angles should be between 0 and 90 degrees for CIE i-table
	for i, angle := range data.Photometry.VerticalAngles {
		if angle < 0 || angle > 90 {
			return fmt.Errorf("vertical angle at index %d (%f°) must be between 0° and 90° for CIE format", i, angle)
		}
	}

	// Check horizontal angles (C-plane angles)
	if len(data.Photometry.HorizontalAngles) == 0 {
		return errors.New("horizontal angles cannot be empty")
	}

	// Horizontal angles should be between 0 and 360 degrees
	for i, angle := range data.Photometry.HorizontalAngles {
		if angle < 0 || angle > 360 {
			return fmt.Errorf("horizontal angle at index %d (%f°) must be between 0° and 360°", i, angle)
		}
	}

	// Check for reasonable angular resolution
	if len(data.Photometry.VerticalAngles) > 1 {
		minVerticalStep := data.Photometry.VerticalAngles[1] - data.Photometry.VerticalAngles[0]
		if minVerticalStep < 1.0 {
			return fmt.Errorf("vertical angular resolution too fine (%f°), CIE format typically uses 5° steps", minVerticalStep)
		}
		if minVerticalStep > 15.0 {
			return fmt.Errorf("vertical angular resolution too coarse (%f°), CIE format typically uses 5° steps", minVerticalStep)
		}
	}

	if len(data.Photometry.HorizontalAngles) > 1 {
		minHorizontalStep := data.Photometry.HorizontalAngles[1] - data.Photometry.HorizontalAngles[0]
		if minHorizontalStep < 5.0 {
			return fmt.Errorf("horizontal angular resolution too fine (%f°), CIE format typically uses 22.5° steps", minHorizontalStep)
		}
		if minHorizontalStep > 45.0 {
			return fmt.Errorf("horizontal angular resolution too coarse (%f°), CIE format typically uses 22.5° steps", minHorizontalStep)
		}
	}

	return nil
}

// validateIntensityData validates the intensity data for CIE format requirements
func (v *Validator) validateIntensityData(data *models.PhotometricData) error {
	if len(data.Photometry.CandelaValues) == 0 {
		return errors.New("candela values cannot be empty")
	}

	// Check data dimensions
	expectedRows := len(data.Photometry.VerticalAngles)
	expectedCols := len(data.Photometry.HorizontalAngles)

	if len(data.Photometry.CandelaValues) != expectedRows {
		return fmt.Errorf("candela values rows (%d) must match vertical angles count (%d)",
			len(data.Photometry.CandelaValues), expectedRows)
	}

	// Check each row
	for i, row := range data.Photometry.CandelaValues {
		if len(row) != expectedCols {
			return fmt.Errorf("candela values row %d has %d columns, expected %d",
				i, len(row), expectedCols)
		}

		// Check for reasonable intensity values
		for j, value := range row {
			if value < 0 {
				return fmt.Errorf("negative intensity value at [%d][%d]: %f", i, j, value)
			}

			// Check for extremely high values that might indicate unit errors
			if value > 10000 {
				return fmt.Errorf("intensity value at [%d][%d] seems too high: %f cd/lm (check units)", i, j, value)
			}
		}
	}

	return nil
}

// validateSymmetry validates symmetry conditions for CIE format
func (v *Validator) validateSymmetry(data *models.PhotometricData) error {
	// Check for basic symmetry conditions that are common in CIE data

	// If we have full 360° data, check for reasonable symmetry
	maxHorizontalAngle := 0.0
	if len(data.Photometry.HorizontalAngles) > 0 {
		maxHorizontalAngle = data.Photometry.HorizontalAngles[len(data.Photometry.HorizontalAngles)-1]
	}

	// If we have quarter symmetry (0-90°), validate it
	if maxHorizontalAngle <= 90 {
		return v.validateQuarterSymmetry(data)
	}

	// If we have half symmetry (0-180°), validate it
	if maxHorizontalAngle <= 180 {
		return v.validateHalfSymmetry(data)
	}

	// For full 360° data, check for reasonable patterns
	return v.validateFullData(data)
}

// validateQuarterSymmetry validates quarter symmetry conditions
func (v *Validator) validateQuarterSymmetry(data *models.PhotometricData) error {
	// For quarter symmetry, we expect the luminaire to be symmetric about both C0-C180 and C90-C270 planes
	// This means the data should represent only the first quadrant (0-90°)

	// Check that horizontal angles are within 0-90°
	for i, angle := range data.Photometry.HorizontalAngles {
		if angle > 90 {
			return fmt.Errorf("quarter symmetry data should not have horizontal angles > 90°, found %f° at index %d", angle, i)
		}
	}

	return nil
}

// validateHalfSymmetry validates half symmetry conditions
func (v *Validator) validateHalfSymmetry(data *models.PhotometricData) error {
	// For half symmetry, we expect the luminaire to be symmetric about one plane
	// This means the data should represent only half the distribution (0-180°)

	// Check that horizontal angles are within 0-180°
	for i, angle := range data.Photometry.HorizontalAngles {
		if angle > 180 {
			return fmt.Errorf("half symmetry data should not have horizontal angles > 180°, found %f° at index %d", angle, i)
		}
	}

	return nil
}

// validateFullData validates full 360° data
func (v *Validator) validateFullData(data *models.PhotometricData) error {
	// For full data, check that we have reasonable coverage
	maxHorizontalAngle := 0.0
	if len(data.Photometry.HorizontalAngles) > 0 {
		maxHorizontalAngle = data.Photometry.HorizontalAngles[len(data.Photometry.HorizontalAngles)-1]
	}

	// We should have close to 360° coverage for full data
	if maxHorizontalAngle < 315 { // Allow some tolerance
		return fmt.Errorf("full data should cover close to 360°, maximum angle found: %f°", maxHorizontalAngle)
	}

	// Check for reasonable data distribution
	// The intensity should generally decrease as we move away from nadir (0°)
	if len(data.Photometry.CandelaValues) >= 2 {
		nadirRow := data.Photometry.CandelaValues[0] // 0° vertical angle
		maxNadirIntensity := 0.0
		for _, value := range nadirRow {
			if value > maxNadirIntensity {
				maxNadirIntensity = value
			}
		}

		// Check that intensity generally decreases with increasing vertical angle
		for i := 1; i < len(data.Photometry.CandelaValues); i++ {
			row := data.Photometry.CandelaValues[i]
			maxRowIntensity := 0.0
			for _, value := range row {
				if value > maxRowIntensity {
					maxRowIntensity = value
				}
			}

			// Allow some tolerance for measurement variations
			if maxRowIntensity > maxNadirIntensity*2 {
				verticalAngle := data.Photometry.VerticalAngles[i]
				return fmt.Errorf("intensity at %f° vertical angle (%f) is much higher than nadir intensity (%f), check data",
					verticalAngle, maxRowIntensity, maxNadirIntensity)
			}
		}
	}

	return nil
}

// ValidateCIEFile validates a CIE file structure before conversion
func (v *Validator) ValidateCIEFile(cieFile *CIEFile) error {
	if cieFile == nil {
		return errors.New("CIE file cannot be nil")
	}

	// Validate header
	if err := v.validateCIEHeader(&cieFile.Header); err != nil {
		return fmt.Errorf("header validation failed: %w", err)
	}

	// Validate photometry
	if err := v.validateCIEPhotometry(&cieFile.Photometry); err != nil {
		return fmt.Errorf("photometry validation failed: %w", err)
	}

	return nil
}

// validateCIEHeader validates the CIE file header
func (v *Validator) validateCIEHeader(header *CIEHeader) error {
	// Format type should be 1 for standard CIE i-table
	if header.FormatType != 1 {
		return fmt.Errorf("unsupported format type: %d (expected 1)", header.FormatType)
	}

	// Symmetry type should be 0 or 1
	if header.SymmetryType < 0 || header.SymmetryType > 1 {
		return fmt.Errorf("invalid symmetry type: %d (expected 0 or 1)", header.SymmetryType)
	}

	// Reserved field should be 0
	if header.Reserved != 0 {
		return fmt.Errorf("reserved field should be 0, got: %d", header.Reserved)
	}

	// Description should not be empty
	if strings.TrimSpace(header.Description) == "" {
		return errors.New("description cannot be empty")
	}

	return nil
}

// validateCIEPhotometry validates the CIE photometry data
func (v *Validator) validateCIEPhotometry(photometry *CIEPhotometry) error {
	// Check angular arrays
	if len(photometry.GammaAngles) == 0 {
		return errors.New("gamma angles cannot be empty")
	}

	if len(photometry.CPlaneAngles) == 0 {
		return errors.New("C-plane angles cannot be empty")
	}

	// Check intensity data dimensions
	if len(photometry.IntensityData) != len(photometry.GammaAngles) {
		return fmt.Errorf("intensity data rows (%d) must match gamma angles count (%d)",
			len(photometry.IntensityData), len(photometry.GammaAngles))
	}

	for i, row := range photometry.IntensityData {
		if len(row) != len(photometry.CPlaneAngles) {
			return fmt.Errorf("intensity data row %d has %d columns, expected %d",
				i, len(row), len(photometry.CPlaneAngles))
		}
	}

	return nil
}
