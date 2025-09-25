package converter

import (
	"fmt"
	"illuminate/internal/interfaces"
	"illuminate/internal/models"
	cieParser "illuminate/internal/parsers/cie"
	iesParser "illuminate/internal/parsers/ies"
	ldtParser "illuminate/internal/parsers/ldt"
	cieWriter "illuminate/internal/writers/cie"
	iesWriter "illuminate/internal/writers/ies"
	ldtWriter "illuminate/internal/writers/ldt"
	"sort"
	"strings"
)

// Manager implements the ConversionManager interface
type Manager struct {
	parsers map[string]interfaces.Parser
	writers map[string]interfaces.Writer
}

// NewManager creates a new conversion manager instance
func NewManager() *Manager {
	m := &Manager{
		parsers: make(map[string]interfaces.Parser),
		writers: make(map[string]interfaces.Writer),
	}

	// Register parsers
	m.parsers["ies"] = iesParser.NewParser()
	m.parsers["ldt"] = ldtParser.NewParser()
	m.parsers["cie"] = cieParser.NewParser()

	// Register writers
	m.writers["ies"] = iesWriter.NewWriter()
	m.writers["ldt"] = ldtWriter.NewWriter()
	m.writers["cie"] = cieWriter.NewWriter()

	return m
}

// GetSupportedFormats returns a list of supported format identifiers
func (m *Manager) GetSupportedFormats() []string {
	formats := make([]string, 0, len(m.parsers))
	for format := range m.parsers {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}

// isFormatSupported checks if a format is supported
func (m *Manager) isFormatSupported(format string) bool {
	_, exists := m.parsers[format]
	return exists
}

// Convert performs a complete conversion from source to target format
func (m *Manager) Convert(sourceData []byte, sourceFormat, targetFormat string) (*interfaces.ConversionResult, error) {
	// Implementation will be added when needed
	return nil, fmt.Errorf("not implemented yet")
}

// DetectFormat automatically detects the format of the input data
func (m *Manager) DetectFormat(data []byte) (*interfaces.FormatDetectionResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data provided")
	}

	type detectionResult struct {
		format     string
		confidence float64
		version    string
	}

	var results []detectionResult

	// Try each parser's detection method
	for formatName, parser := range m.parsers {
		confidence, version := parser.DetectFormat(data)
		if confidence > 0 {
			results = append(results, detectionResult{
				format:     formatName,
				confidence: confidence,
				version:    version,
			})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("unable to detect format: no parser recognized the data")
	}

	// Sort by confidence (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].confidence > results[j].confidence
	})

	// Build result
	bestResult := results[0]

	var alternatives []interfaces.FormatAlternative
	for i := 1; i < len(results) && i < 3; i++ { // Include up to 3 alternatives
		alternatives = append(alternatives, interfaces.FormatAlternative{
			Format:     results[i].format,
			Confidence: results[i].confidence,
			Version:    results[i].version,
		})
	}

	return &interfaces.FormatDetectionResult{
		Format:       bestResult.format,
		Confidence:   bestResult.confidence,
		Version:      bestResult.version,
		Alternatives: alternatives,
	}, nil
}

// ValidateData validates photometric data for consistency and quality
func (m *Manager) ValidateData(data *models.PhotometricData) *interfaces.ValidationResult {
	result := &interfaces.ValidationResult{
		IsValid:  true,
		Score:    1.0,
		Warnings: []string{},
		Errors:   []string{},
	}

	if data == nil {
		result.IsValid = false
		result.Score = 0.0
		result.Errors = append(result.Errors, "photometric data is nil")
		return result
	}

	// Validate basic data structure
	if err := data.Validate(); err != nil {
		result.IsValid = false
		result.Score = 0.0
		result.Errors = append(result.Errors, fmt.Sprintf("basic validation failed: %v", err))
		return result
	}

	// Check photometric data consistency
	m.validatePhotometricConsistency(data, result)

	// Check geometry parameter compatibility
	m.validateGeometryCompatibility(data, result)

	// Check electrical data consistency
	m.validateElectricalConsistency(data, result)

	// Calculate final score based on warnings and errors
	if len(result.Errors) > 0 {
		result.IsValid = false
		result.Score = 0.0
	} else {
		// Reduce score based on warnings
		warningPenalty := float64(len(result.Warnings)) * 0.1
		result.Score = 1.0 - warningPenalty
		if result.Score < 0.0 {
			result.Score = 0.0
		}
	}

	return result
}

// validatePhotometricConsistency checks photometric data for consistency
func (m *Manager) validatePhotometricConsistency(data *models.PhotometricData, result *interfaces.ValidationResult) {
	photometry := &data.Photometry

	// Calculate maximum intensity from candela values
	maxIntensity := 0.0
	for _, row := range photometry.CandelaValues {
		for _, value := range row {
			scaledValue := value * photometry.CandelaMultiplier
			if scaledValue > maxIntensity {
				maxIntensity = scaledValue
			}
		}
	}

	// Check if intensity values are reasonable
	if maxIntensity <= 0 {
		result.Warnings = append(result.Warnings, "maximum intensity is zero or negative")
	}

	// Check for extremely high intensity values that might indicate errors
	if maxIntensity > 1000000 {
		result.Warnings = append(result.Warnings, "maximum intensity is unusually high (>1,000,000 cd)")
	}

	// Validate angular data consistency
	if len(photometry.VerticalAngles) == 0 {
		result.Errors = append(result.Errors, "no vertical angles defined")
		return
	}

	if len(photometry.HorizontalAngles) == 0 {
		result.Errors = append(result.Errors, "no horizontal angles defined")
		return
	}

	// Check intensity matrix dimensions
	if len(photometry.CandelaValues) != len(photometry.VerticalAngles) {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"candela matrix rows (%d) don't match vertical angles (%d)",
			len(photometry.CandelaValues), len(photometry.VerticalAngles)))
		return
	}

	for i, row := range photometry.CandelaValues {
		if len(row) != len(photometry.HorizontalAngles) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"candela matrix row %d length (%d) doesn't match horizontal angles (%d)",
				i, len(row), len(photometry.HorizontalAngles)))
			return
		}
	}

	// Check for negative intensity values
	negativeCount := 0
	for _, row := range photometry.CandelaValues {
		for _, intensity := range row {
			if intensity < 0 {
				negativeCount++
			}
		}
	}
	if negativeCount > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"found %d negative intensity values", negativeCount))
	}

	// Check angular range validity
	if len(photometry.VerticalAngles) > 0 {
		if photometry.VerticalAngles[0] < 0 || photometry.VerticalAngles[len(photometry.VerticalAngles)-1] > 180 {
			result.Warnings = append(result.Warnings, "vertical angles outside standard range [0, 180]")
		}
	}

	if len(photometry.HorizontalAngles) > 0 {
		if photometry.HorizontalAngles[0] < 0 || photometry.HorizontalAngles[len(photometry.HorizontalAngles)-1] > 360 {
			result.Warnings = append(result.Warnings, "horizontal angles outside standard range [0, 360]")
		}
	}
}

// validateGeometryCompatibility checks geometry parameters for cross-format compatibility
func (m *Manager) validateGeometryCompatibility(data *models.PhotometricData, result *interfaces.ValidationResult) {
	geometry := &data.Geometry

	// Check luminaire dimensions
	if geometry.Length <= 0 || geometry.Width <= 0 || geometry.Height <= 0 {
		result.Warnings = append(result.Warnings, "luminaire dimensions contain zero or negative values")
	}

	// Check for unreasonable dimensions (too large)
	maxDimension := 100.0 // 100 meters seems reasonable as maximum
	if geometry.Length > maxDimension || geometry.Width > maxDimension || geometry.Height > maxDimension {
		result.Warnings = append(result.Warnings, "luminaire dimensions are unusually large (>100m)")
	}

	// Check photometry type compatibility across formats
	switch data.Photometry.PhotometryType {
	case "A":
		// Type A: road lighting, check if geometry is appropriate
		if geometry.Height < 0.1 {
			result.Warnings = append(result.Warnings, "Type A photometry with very low mounting height")
		}
		// Type A typically has limited horizontal angles
		if len(data.Photometry.HorizontalAngles) > 37 {
			result.Warnings = append(result.Warnings, "Type A photometry with unusually many horizontal angles (may not be compatible with all formats)")
		}
	case "B":
		// Type B: floodlighting, check for appropriate geometry
		if geometry.Length == geometry.Width {
			result.Warnings = append(result.Warnings, "Type B photometry typically uses asymmetric luminaires")
		}
		// Type B may have compatibility issues with some formats
		if len(data.Photometry.VerticalAngles) > 181 || len(data.Photometry.HorizontalAngles) > 361 {
			result.Warnings = append(result.Warnings, "Type B photometry with high angular resolution may not be compatible with all formats")
		}
	case "C":
		// Type C: interior lighting, check for reasonable dimensions
		if geometry.Height > 10.0 {
			result.Warnings = append(result.Warnings, "Type C photometry with unusually high mounting height")
		}
		// Type C is most widely supported
		if len(data.Photometry.VerticalAngles) > 181 || len(data.Photometry.HorizontalAngles) > 361 {
			result.Warnings = append(result.Warnings, "High angular resolution may require format-specific handling")
		}
	default:
		result.Errors = append(result.Errors, fmt.Sprintf("unknown photometry type: %s", data.Photometry.PhotometryType))
	}

	// Check for format-specific geometry limitations
	m.validateFormatSpecificGeometry(data, result)
}

// validateFormatSpecificGeometry checks geometry compatibility with specific formats
func (m *Manager) validateFormatSpecificGeometry(data *models.PhotometricData, result *interfaces.ValidationResult) {
	// IES format limitations
	if len(data.Photometry.VerticalAngles) > 37 {
		result.Warnings = append(result.Warnings, "IES format typically supports maximum 37 vertical angles")
	}
	if len(data.Photometry.HorizontalAngles) > 73 {
		result.Warnings = append(result.Warnings, "IES format typically supports maximum 73 horizontal angles")
	}

	// CIE format limitations
	if data.Metadata.Manufacturer == "" || data.Metadata.CatalogNumber == "" {
		result.Warnings = append(result.Warnings, "CIE format has limited metadata support - some information may be lost")
	}

	// Check for electrical data compatibility
	if data.Electrical.InputWatts == 0 && data.Electrical.InputVoltage == 0 {
		result.Warnings = append(result.Warnings, "CIE format does not support electrical data - this information will be lost")
	}

	// LDT format specific checks
	if data.Photometry.UnitsType == "relative" {
		result.Warnings = append(result.Warnings, "LDT format prefers absolute photometry - relative values may need scaling")
	}
}

// validateElectricalConsistency checks electrical data for consistency
func (m *Manager) validateElectricalConsistency(data *models.PhotometricData, result *interfaces.ValidationResult) {
	electrical := &data.Electrical

	// Check power values
	if electrical.InputWatts <= 0 {
		result.Warnings = append(result.Warnings, "input power is zero or negative")
	}

	if electrical.InputWatts > 10000 {
		result.Warnings = append(result.Warnings, "input power is unusually high (>10kW)")
	}

	// Check voltage values
	if electrical.InputVoltage <= 0 {
		result.Warnings = append(result.Warnings, "input voltage is zero or negative")
	}

	// Check for common voltage ranges
	commonVoltages := []float64{12, 24, 120, 208, 240, 277, 347, 480}
	voltageOK := false
	tolerance := 0.1
	for _, commonVolt := range commonVoltages {
		if electrical.InputVoltage >= commonVolt*(1-tolerance) && electrical.InputVoltage <= commonVolt*(1+tolerance) {
			voltageOK = true
			break
		}
	}
	if !voltageOK && electrical.InputVoltage > 0 {
		result.Warnings = append(result.Warnings, "input voltage doesn't match common standards")
	}

	// Check luminous flux consistency (from photometry data)
	if data.Photometry.LuminousFlux <= 0 {
		result.Warnings = append(result.Warnings, "luminous flux is zero or negative")
	}

	// Check efficacy (lumens per watt)
	if electrical.InputWatts > 0 && data.Photometry.LuminousFlux > 0 {
		efficacy := data.Photometry.LuminousFlux / electrical.InputWatts
		if efficacy < 10 {
			result.Warnings = append(result.Warnings, "luminous efficacy is very low (<10 lm/W)")
		}
		if efficacy > 300 {
			result.Warnings = append(result.Warnings, "luminous efficacy is unusually high (>300 lm/W)")
		}
	}

	// Check ballast factors
	if electrical.BallastFactor > 0 && (electrical.BallastFactor < 0.5 || electrical.BallastFactor > 1.5) {
		result.Warnings = append(result.Warnings, "ballast factor outside typical range (0.5-1.5)")
	}

	if electrical.BallastLampFactor > 0 && (electrical.BallastLampFactor < 0.5 || electrical.BallastLampFactor > 1.5) {
		result.Warnings = append(result.Warnings, "ballast lamp factor outside typical range (0.5-1.5)")
	}

	// Check power factor
	if electrical.PowerFactor > 0 && electrical.PowerFactor < 0.5 {
		result.Warnings = append(result.Warnings, "power factor is quite low (<0.5)")
	}
}

// GetFormatInfo returns detailed information about a specific format
func (m *Manager) GetFormatInfo(format string) (*interfaces.FormatInfo, error) {
	format = strings.ToLower(strings.TrimSpace(format))

	parser, exists := m.parsers[format]
	if !exists {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	switch format {
	case "ies":
		return &interfaces.FormatInfo{
			Identifier:        "ies",
			Name:              "IES Photometric Data",
			Description:       "Illuminating Engineering Society standard photometric data format",
			SupportedVersions: parser.GetSupportedVersions(),
			FileExtensions:    []string{".ies"},
			MimeTypes:         []string{"text/plain", "application/x-ies"},
			Standards:         []string{"LM-63-1995", "LM-63-2002"},
			Capabilities: interfaces.FormatCapabilities{
				SupportsPhotometryTypes:    []string{"A", "B", "C"},
				SupportsAbsolutePhotometry: true,
				SupportsRelativePhotometry: true,
				MaxVerticalAngles:          37,
				MaxHorizontalAngles:        73,
				SupportsMetadata:           true,
				SupportsGeometry:           true,
				SupportsElectrical:         true,
			},
		}, nil

	case "ldt":
		return &interfaces.FormatInfo{
			Identifier:        "ldt",
			Name:              "EULUMDAT",
			Description:       "European standard for luminaire photometric data",
			SupportedVersions: parser.GetSupportedVersions(),
			FileExtensions:    []string{".ldt"},
			MimeTypes:         []string{"text/plain", "application/x-ldt"},
			Standards:         []string{"EULUMDAT 1.0"},
			Capabilities: interfaces.FormatCapabilities{
				SupportsPhotometryTypes:    []string{"A", "B", "C"},
				SupportsAbsolutePhotometry: true,
				SupportsRelativePhotometry: true,
				MaxVerticalAngles:          181,
				MaxHorizontalAngles:        361,
				SupportsMetadata:           true,
				SupportsGeometry:           true,
				SupportsElectrical:         true,
			},
		}, nil

	case "cie":
		return &interfaces.FormatInfo{
			Identifier:        "cie",
			Name:              "CIE Photometric Data",
			Description:       "Commission Internationale de l'Éclairage photometric data format",
			SupportedVersions: parser.GetSupportedVersions(),
			FileExtensions:    []string{".cie"},
			MimeTypes:         []string{"text/plain", "application/x-cie"},
			Standards:         []string{"CIE 102", "CIE i-table"},
			Capabilities: interfaces.FormatCapabilities{
				SupportsPhotometryTypes:    []string{"A", "B", "C"},
				SupportsAbsolutePhotometry: true,
				SupportsRelativePhotometry: false,
				MaxVerticalAngles:          181,
				MaxHorizontalAngles:        361,
				SupportsMetadata:           false,
				SupportsGeometry:           true,
				SupportsElectrical:         false,
			},
		}, nil

	default:
		return nil, fmt.Errorf("format info not available for: %s", format)
	}
}
