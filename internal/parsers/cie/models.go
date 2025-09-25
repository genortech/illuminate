package cie

import (
	"illuminate/internal/models"
	"strconv"
	"strings"
)

// CIEVersion represents the supported CIE format versions
type CIEVersion string

const (
	Version102    CIEVersion = "CIE 102-1993"
	VersionITable CIEVersion = "CIE i-table"
)

// CIEHeader contains the header information from a CIE file
type CIEHeader struct {
	FormatType   int    `json:"format_type"`   // Format type indicator
	SymmetryType int    `json:"symmetry_type"` // Symmetry indicator
	Reserved     int    `json:"reserved"`      // Reserved field
	Description  string `json:"description"`   // Luminaire description
}

// CIEPhotometry contains the photometric measurement data
type CIEPhotometry struct {
	// Standard CIE i-table uses fixed angular grid
	// Gamma angles: 0° to 90° in 5° increments (19 angles)
	// C-plane angles: 0° to 360° in 22.5° increments (17 angles including 360°=0°)
	GammaAngles   []float64   `json:"gamma_angles"`
	CPlaneAngles  []float64   `json:"c_plane_angles"`
	IntensityData [][]float64 `json:"intensity_data"` // [gamma][c_plane] in cd/1000lm
}

// CIEFile represents a complete CIE file structure
type CIEFile struct {
	Header     CIEHeader     `json:"header"`
	Photometry CIEPhotometry `json:"photometry"`
}

// ToCommonModel converts CIE-specific data to the common photometric data model
func (cie *CIEFile) ToCommonModel() (*models.PhotometricData, error) {
	// Extract metadata from description
	metadata := models.LuminaireMetadata{
		Manufacturer:  extractManufacturer(cie.Header.Description),
		CatalogNumber: extractCatalogNumber(cie.Header.Description),
		Description:   cie.Header.Description,
		TestLab:       "Unknown",
		TestDate:      "",
		TestNumber:    "",
	}

	// Set default values if not provided
	if metadata.Manufacturer == "" {
		metadata.Manufacturer = "Unknown"
	}
	if metadata.CatalogNumber == "" {
		metadata.CatalogNumber = "Unknown"
	}

	// CIE files typically don't contain geometry information
	// Set default values
	geometry := models.LuminaireGeometry{
		Length:         1.0, // Default 1m
		Width:          1.0, // Default 1m
		Height:         0.1, // Default 0.1m
		LuminousLength: 1.0,
		LuminousWidth:  1.0,
		LuminousHeight: 0.1,
	}

	// Extract luminous flux from description if available
	totalLumens := extractLuminousFlux(cie.Header.Description)
	if totalLumens == 0 {
		totalLumens = 1000.0 // Default value
	}

	// Convert intensity data from cd/1000lm to cd/lm
	candelaValues := make([][]float64, len(cie.Photometry.IntensityData))
	for i, row := range cie.Photometry.IntensityData {
		candelaValues[i] = make([]float64, len(row))
		for j, value := range row {
			// Convert from cd/1000lm to cd/lm by dividing by 1000
			candelaValues[i][j] = value / 1000.0
		}
	}

	photometry := models.PhotometricMeasurements{
		PhotometryType:    "C", // CIE i-table is Type C photometry
		UnitsType:         "absolute",
		LuminousFlux:      totalLumens,
		CandelaMultiplier: 1.0,
		VerticalAngles:    cie.Photometry.GammaAngles,
		HorizontalAngles:  cie.Photometry.CPlaneAngles,
		CandelaValues:     candelaValues,
	}

	electrical := models.ElectricalData{
		InputWatts:        extractWattage(cie.Header.Description),
		BallastFactor:     1.0, // Default value for CIE
		BallastLampFactor: 1.0, // Default value for CIE
	}

	return &models.PhotometricData{
		Metadata:   metadata,
		Geometry:   geometry,
		Photometry: photometry,
		Electrical: electrical,
	}, nil
}

// FromCommonModel converts common photometric data to CIE-specific format
func (cie *CIEFile) FromCommonModel(data *models.PhotometricData) error {
	// Create header with description
	description := formatDescription(data.Metadata.Description, data.Photometry.LuminousFlux, data.Electrical.InputWatts)

	cie.Header = CIEHeader{
		FormatType:   1, // Standard format type
		SymmetryType: determineSymmetryType(data.Photometry.HorizontalAngles),
		Reserved:     0, // Always 0
		Description:  description,
	}

	// Set up standard CIE i-table angular grid
	cie.Photometry.GammaAngles = generateStandardGammaAngles()
	cie.Photometry.CPlaneAngles = generateStandardCPlaneAngles()

	// Interpolate or map the input data to the standard CIE grid
	intensityData, err := interpolateToStandardGrid(
		data.Photometry.VerticalAngles,
		data.Photometry.HorizontalAngles,
		data.Photometry.CandelaValues,
		cie.Photometry.GammaAngles,
		cie.Photometry.CPlaneAngles,
	)
	if err != nil {
		return err
	}

	// Convert from cd/lm to cd/1000lm
	cie.Photometry.IntensityData = make([][]float64, len(intensityData))
	for i, row := range intensityData {
		cie.Photometry.IntensityData[i] = make([]float64, len(row))
		for j, value := range row {
			// Convert from cd/lm to cd/1000lm by multiplying by 1000
			cie.Photometry.IntensityData[i][j] = value * 1000.0
		}
	}

	return nil
}

// Helper functions

// extractManufacturer extracts manufacturer from description
func extractManufacturer(description string) string {
	// Simple heuristic: first word or phrase before model number
	parts := strings.Fields(description)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// extractCatalogNumber extracts catalog/model number from description
func extractCatalogNumber(description string) string {
	// Look for patterns like "17W", "P14W", etc.
	parts := strings.Fields(description)
	for _, part := range parts {
		if strings.Contains(part, "W") {
			return part
		}
	}
	for _, part := range parts {
		if strings.Contains(part, "LED") {
			return part
		}
	}
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// extractLuminousFlux extracts luminous flux from description
func extractLuminousFlux(description string) float64 {
	// Look for patterns like "2172.2 lms", "1289 lms", "2458 lm"
	parts := strings.Fields(description)
	for i, part := range parts {
		if strings.Contains(part, "lm") {
			// Try to parse the previous part as a number
			if i > 0 {
				if flux, err := strconv.ParseFloat(parts[i-1], 64); err == nil {
					return flux
				}
			}
		}
		// Also try parsing numbers followed by "lm"
		if strings.HasSuffix(part, "lm") || strings.HasSuffix(part, "lms") {
			numStr := strings.TrimSuffix(strings.TrimSuffix(part, "lms"), "lm")
			if flux, err := strconv.ParseFloat(numStr, 64); err == nil {
				return flux
			}
		}
	}
	return 0.0
}

// extractWattage extracts wattage from description
func extractWattage(description string) float64 {
	// Look for patterns like "17W", "14W"
	parts := strings.Fields(description)
	for _, part := range parts {
		if strings.HasSuffix(part, "W") && len(part) > 1 {
			wattStr := strings.TrimSuffix(part, "W")
			if watts, err := strconv.ParseFloat(wattStr, 64); err == nil {
				return watts
			}
		}
	}
	return 0.0
}

// formatDescription formats a description for CIE output
func formatDescription(desc string, lumens, watts float64) string {
	if desc == "" {
		desc = "LED Luminaire"
	}

	result := desc
	if watts > 0 {
		result += " " + strconv.FormatFloat(watts, 'f', 0, 64) + "W"
	}
	if lumens > 0 {
		result += " " + strconv.FormatFloat(lumens, 'f', 1, 64) + " lm"
	}

	return result
}

// determineSymmetryType determines symmetry type from horizontal angles
func determineSymmetryType(horizontalAngles []float64) int {
	if len(horizontalAngles) <= 1 {
		return 1 // Full symmetry
	}

	// Check for quarter symmetry (0-90°)
	maxAngle := horizontalAngles[len(horizontalAngles)-1]
	if maxAngle <= 90 {
		return 1 // Quarter symmetry
	}

	// Check for half symmetry (0-180°)
	if maxAngle <= 180 {
		return 0 // Half symmetry
	}

	return 0 // No symmetry (full 360°)
}

// generateStandardGammaAngles generates the standard CIE i-table gamma angles (0° to 90° in 5° steps)
func generateStandardGammaAngles() []float64 {
	angles := make([]float64, 19) // 0, 5, 10, ..., 90
	for i := 0; i < 19; i++ {
		angles[i] = float64(i * 5)
	}
	return angles
}

// generateStandardCPlaneAngles generates the standard CIE i-table C-plane angles (0° to 337.5° in 22.5° steps)
func generateStandardCPlaneAngles() []float64 {
	angles := make([]float64, 16) // 0, 22.5, 45, ..., 337.5 (16 angles, not 17)
	for i := 0; i < 16; i++ {
		angles[i] = float64(i) * 22.5
	}
	return angles
}

// interpolateToStandardGrid interpolates input data to the standard CIE grid
func interpolateToStandardGrid(
	inputGamma, inputCPlane []float64,
	inputData [][]float64,
	targetGamma, targetCPlane []float64,
) ([][]float64, error) {
	result := make([][]float64, len(targetGamma))

	for i, gamma := range targetGamma {
		result[i] = make([]float64, len(targetCPlane))

		for j, cplane := range targetCPlane {
			// Simple nearest neighbor interpolation for now
			// In a production system, you might want bilinear interpolation
			value := interpolateValue(inputGamma, inputCPlane, inputData, gamma, cplane)
			result[i][j] = value
		}
	}

	return result, nil
}

// interpolateValue performs simple nearest neighbor interpolation
func interpolateValue(inputGamma, inputCPlane []float64, inputData [][]float64, targetGamma, targetCPlane float64) float64 {
	if len(inputData) == 0 || len(inputData[0]) == 0 {
		return 0.0
	}

	// Find nearest gamma angle
	gammaIdx := 0
	minGammaDiff := 1000.0
	for i, angle := range inputGamma {
		diff := abs(angle - targetGamma)
		if diff < minGammaDiff {
			minGammaDiff = diff
			gammaIdx = i
		}
	}

	// Find nearest C-plane angle (handle 360° wraparound)
	cplaneIdx := 0
	minCPlaneDiff := 1000.0
	for i, angle := range inputCPlane {
		diff := abs(angle - targetCPlane)
		// Handle wraparound (e.g., 350° and 10° are only 20° apart)
		wrapDiff := abs(abs(angle-targetCPlane) - 360.0)
		if wrapDiff < diff {
			diff = wrapDiff
		}
		if diff < minCPlaneDiff {
			minCPlaneDiff = diff
			cplaneIdx = i
		}
	}

	// Bounds check
	if gammaIdx >= len(inputData) {
		gammaIdx = len(inputData) - 1
	}
	if cplaneIdx >= len(inputData[gammaIdx]) {
		cplaneIdx = len(inputData[gammaIdx]) - 1
	}

	return inputData[gammaIdx][cplaneIdx]
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ParseFloat safely parses a float64 from string, returning 0 on error
func ParseFloat(s string) float64 {
	if val, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return val
	}
	return 0.0
}

// ParseInt safely parses an int from string, returning 0 on error
func ParseInt(s string) int {
	if val, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return val
	}
	return 0
}
