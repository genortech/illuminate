package ldt

import (
	"fmt"
	"illuminate/internal/interfaces"
	"illuminate/internal/models"
	"strconv"
	"strings"
)

// Parser implements the interfaces.Parser interface for LDT files
type Parser struct {
	supportedVersions []string
}

// NewParser creates a new LDT parser instance
func NewParser() *Parser {
	return &Parser{
		supportedVersions: []string{string(Version10)},
	}
}

// Parse converts raw LDT file data to the common photometric data model
func (p *Parser) Parse(data []byte) (*models.PhotometricData, error) {
	ldtFile, err := p.parseLDTFile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LDT file: %w", err)
	}

	commonData, err := ldtFile.ToCommonModel()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to common model: %w", err)
	}

	return commonData, nil
}

// Validate performs format-specific validation on the parsed data
func (p *Parser) Validate(data *models.PhotometricData) error {
	validator := NewValidator()
	return validator.Validate(data)
}

// GetSupportedVersions returns the format versions supported by this parser
func (p *Parser) GetSupportedVersions() []string {
	return p.supportedVersions
}

// DetectFormat attempts to identify if the data matches LDT format
func (p *Parser) DetectFormat(data []byte) (confidence float64, version string) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) < 10 {
		return 0.0, ""
	}

	confidence = 0.0

	// Check for typical LDT structure patterns
	// Line 1: Company identification (often contains semicolon)
	if strings.Contains(lines[0], ";") {
		confidence += 0.2
	}

	// Lines 2-7 should be numeric (type indicator, symmetry, etc.)
	numericLineCount := 0
	for i := 1; i < min(7, len(lines)); i++ {
		line := strings.TrimSpace(lines[i])
		if isNumericLine(line) {
			numericLineCount++
		}
	}

	if numericLineCount >= 4 {
		confidence += 0.3
	}

	// Check for measurement report and luminaire name (lines 8-9)
	if len(lines) > 8 {
		// These lines typically contain descriptive text
		if len(strings.TrimSpace(lines[7])) > 0 && len(strings.TrimSpace(lines[8])) > 0 {
			confidence += 0.1
		}
	}

	// Check for date/user line pattern (line 13)
	if len(lines) > 12 {
		dateUserLine := strings.TrimSpace(lines[12])
		if strings.Contains(dateUserLine, "/") || isDateLike(dateUserLine) {
			confidence += 0.2
		}
	}

	// Look for large blocks of numeric data (photometric measurements)
	numericDataLines := 0
	for i := 20; i < min(len(lines), 100); i++ {
		line := strings.TrimSpace(lines[i])
		if isNumericLine(line) {
			numericDataLines++
		}
	}

	if numericDataLines > 20 {
		confidence += 0.2
	}

	// Check for European decimal separator (comma)
	commaDecimalCount := 0
	for i := 0; i < min(len(lines), 50); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, ",") && isNumericLine(line) {
			commaDecimalCount++
		}
	}

	if commaDecimalCount > 5 {
		confidence += 0.1
	}

	if confidence > 0.5 {
		return confidence, string(Version10)
	}

	return confidence, ""
}

// parseLDTFile parses the raw LDT file data into an LDTFile structure
func (p *Parser) parseLDTFile(data []byte) (*LDTFile, error) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) < 13 {
		return nil, fmt.Errorf("insufficient lines in LDT file: expected at least 13, got %d", len(lines))
	}

	ldtFile := &LDTFile{}

	// Parse header (lines 1-13)
	if err := p.parseHeader(lines, ldtFile); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Parse geometry data (lines 14-26)
	if err := p.parseGeometry(lines, ldtFile); err != nil {
		return nil, fmt.Errorf("failed to parse geometry: %w", err)
	}

	// Parse electrical data (lines 27+)
	lineIndex := 26
	if err := p.parseElectrical(lines, &lineIndex, ldtFile); err != nil {
		return nil, fmt.Errorf("failed to parse electrical data: %w", err)
	}

	// Parse photometric data
	if err := p.parsePhotometry(lines, lineIndex, ldtFile); err != nil {
		return nil, fmt.Errorf("failed to parse photometric data: %w", err)
	}

	return ldtFile, nil
}

// parseHeader parses the LDT file header (lines 1-13)
func (p *Parser) parseHeader(lines []string, ldtFile *LDTFile) error {
	if len(lines) < 13 {
		return fmt.Errorf("insufficient lines for header: expected 13, got %d", len(lines))
	}

	ldtFile.Header = LDTHeader{
		CompanyIdentification:              strings.TrimSpace(lines[0]),
		TypeIndicator:                      ParseInt(lines[1]),
		SymmetryIndicator:                  ParseInt(lines[2]),
		NumberOfCPlanes:                    ParseInt(lines[3]),
		DistanceBetweenCPlanes:             ParseFloat(lines[4]),
		NumberOfLuminousIntensities:        ParseInt(lines[5]),
		DistanceBetweenLuminousIntensities: ParseFloat(lines[6]),
		MeasurementReport:                  strings.TrimSpace(lines[7]),
		LuminaireName:                      strings.TrimSpace(lines[8]),
		LuminaireNumber:                    strings.TrimSpace(lines[9]),
		FileName:                           strings.TrimSpace(lines[10]),
		DateUser:                           strings.TrimSpace(lines[11]),
	}

	// Validate header values
	if ldtFile.Header.NumberOfCPlanes <= 0 {
		return fmt.Errorf("invalid number of C planes: %d", ldtFile.Header.NumberOfCPlanes)
	}
	if ldtFile.Header.NumberOfLuminousIntensities <= 0 {
		return fmt.Errorf("invalid number of luminous intensities: %d", ldtFile.Header.NumberOfLuminousIntensities)
	}

	return nil
}

// parseGeometry parses the geometry data (lines 14-26)
func (p *Parser) parseGeometry(lines []string, ldtFile *LDTFile) error {
	if len(lines) < 26 {
		return fmt.Errorf("insufficient lines for geometry: expected at least 26, got %d", len(lines))
	}

	ldtFile.Geometry = LDTGeometry{
		LengthOfLuminaire:         ParseFloat(lines[12]),
		WidthOfLuminaire:          ParseFloat(lines[13]),
		HeightOfLuminaire:         ParseFloat(lines[14]),
		LengthOfLuminousArea:      ParseFloat(lines[15]),
		WidthOfLuminousArea:       ParseFloat(lines[16]),
		HeightOfLuminousAreaC0:    ParseFloat(lines[17]),
		HeightOfLuminousAreaC90:   ParseFloat(lines[18]),
		HeightOfLuminousAreaC180:  ParseFloat(lines[19]),
		HeightOfLuminousAreaC270:  ParseFloat(lines[20]),
		DownwardFluxFraction:      ParseFloat(lines[21]),
		LightOutputRatioLuminaire: ParseFloat(lines[22]),
		ConversionFactor:          ParseFloat(lines[23]),
	}

	// Validate geometry values
	if ldtFile.Geometry.ConversionFactor <= 0 {
		ldtFile.Geometry.ConversionFactor = 1.0 // Default value
	}

	return nil
}

// parseElectrical parses the electrical data starting from the given line index
func (p *Parser) parseElectrical(lines []string, lineIndex *int, ldtFile *LDTFile) error {
	if *lineIndex >= len(lines) {
		return fmt.Errorf("insufficient lines for electrical data")
	}

	// Parse DR index (line 27)
	ldtFile.Electrical.DRIndex = ParseInt(lines[*lineIndex])
	*lineIndex++

	// Parse number of lamp sets (line 28)
	if *lineIndex >= len(lines) {
		return fmt.Errorf("missing number of lamp sets")
	}
	ldtFile.Electrical.NumberOfLampSets = ParseInt(lines[*lineIndex])
	*lineIndex++

	// Parse each lamp set (6 lines per set)
	ldtFile.Electrical.LampSets = make([]LampSet, ldtFile.Electrical.NumberOfLampSets)
	for i := 0; i < ldtFile.Electrical.NumberOfLampSets; i++ {
		if *lineIndex+5 >= len(lines) {
			return fmt.Errorf("insufficient lines for lamp set %d", i+1)
		}

		ldtFile.Electrical.LampSets[i] = LampSet{
			NumberOfLamps:           ParseInt(lines[*lineIndex]),
			Type:                    strings.TrimSpace(lines[*lineIndex+1]),
			TotalLuminousFlux:       ParseFloat(lines[*lineIndex+2]),
			ColorTemperature:        strings.TrimSpace(lines[*lineIndex+3]),
			ColorRenderingGroup:     strings.TrimSpace(lines[*lineIndex+4]),
			WattageIncludingBallast: ParseFloat(lines[*lineIndex+5]),
		}
		*lineIndex += 6
	}

	return nil
}

// parsePhotometry parses the photometric data starting from the given line index
func (p *Parser) parsePhotometry(lines []string, lineIndex int, ldtFile *LDTFile) error {
	// Parse C plane angles
	ldtFile.Photometry.CPlaneAngles = make([]float64, 0, ldtFile.Header.NumberOfCPlanes)
	err := p.parseFloatArray(lines, &lineIndex, &ldtFile.Photometry.CPlaneAngles, ldtFile.Header.NumberOfCPlanes)
	if err != nil {
		return fmt.Errorf("failed to parse C plane angles: %w", err)
	}

	// Parse gamma angles
	ldtFile.Photometry.GammaAngles = make([]float64, 0, ldtFile.Header.NumberOfLuminousIntensities)
	err = p.parseFloatArray(lines, &lineIndex, &ldtFile.Photometry.GammaAngles, ldtFile.Header.NumberOfLuminousIntensities)
	if err != nil {
		return fmt.Errorf("failed to parse gamma angles: %w", err)
	}

	// Parse luminous intensity distribution
	// Data is organized by C-plane first, then by gamma angle
	// We need to transpose it to [gamma][c_plane] format for our common model
	tempData := make([][]float64, ldtFile.Header.NumberOfCPlanes)

	for cPlane := 0; cPlane < ldtFile.Header.NumberOfCPlanes; cPlane++ {
		tempData[cPlane] = make([]float64, 0, ldtFile.Header.NumberOfLuminousIntensities)
		err = p.parseFloatArray(lines, &lineIndex, &tempData[cPlane], ldtFile.Header.NumberOfLuminousIntensities)
		if err != nil {
			return fmt.Errorf("failed to parse luminous intensity data for C-plane %d: %w", cPlane, err)
		}
	}

	// Transpose the data to [gamma][c_plane] format
	ldtFile.Photometry.LuminousIntensityDistribution = make([][]float64, ldtFile.Header.NumberOfLuminousIntensities)
	for gamma := 0; gamma < ldtFile.Header.NumberOfLuminousIntensities; gamma++ {
		ldtFile.Photometry.LuminousIntensityDistribution[gamma] = make([]float64, ldtFile.Header.NumberOfCPlanes)
		for cPlane := 0; cPlane < ldtFile.Header.NumberOfCPlanes; cPlane++ {
			ldtFile.Photometry.LuminousIntensityDistribution[gamma][cPlane] = tempData[cPlane][gamma]
		}
	}

	return nil
}

// parseFloatArray parses floating point values from multiple lines
func (p *Parser) parseFloatArray(lines []string, lineIndex *int, target *[]float64, expectedCount int) error {
	startIndex := *lineIndex
	for len(*target) < expectedCount && *lineIndex < len(lines) {
		line := strings.TrimSpace(lines[*lineIndex])
		if line == "" {
			*lineIndex++
			continue
		}

		fields := strings.Fields(line)
		for _, field := range fields {
			if len(*target) >= expectedCount {
				break
			}

			value := ParseFloat(field)
			*target = append(*target, value)
		}
		*lineIndex++
	}

	if len(*target) != expectedCount {
		return fmt.Errorf("expected %d values, got %d (started at line %d, ended at line %d)",
			expectedCount, len(*target), startIndex+1, *lineIndex)
	}

	return nil
}

// Helper functions

// isNumericLine checks if a line contains primarily numeric data
func isNumericLine(line string) bool {
	if line == "" {
		return false
	}

	// Handle European decimal separator
	line = strings.Replace(line, ",", ".", -1)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	numericCount := 0
	for _, field := range fields {
		if _, err := strconv.ParseFloat(field, 64); err == nil {
			numericCount++
		}
	}

	// Consider it numeric if more than half the fields are numbers
	return float64(numericCount)/float64(len(fields)) > 0.5
}

// isDateLike checks if a string looks like a date
func isDateLike(s string) bool {
	// Simple heuristic: contains digits and common date separators
	hasDigits := false
	hasDateSeparators := false

	for _, char := range s {
		if char >= '0' && char <= '9' {
			hasDigits = true
		}
		if char == '/' || char == '-' || char == '.' || char == ' ' {
			hasDateSeparators = true
		}
	}

	return hasDigits && hasDateSeparators
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure Parser implements the interfaces.Parser interface
var _ interfaces.Parser = (*Parser)(nil)
