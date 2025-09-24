package ies

import (
	"fmt"
	"illuminate/internal/converter"
	"illuminate/internal/models"
	"regexp"
	"strconv"
	"strings"
)

// Parser implements the converter.Parser interface for IES files
type Parser struct {
	supportedVersions []string
}

// NewParser creates a new IES parser instance
func NewParser() *Parser {
	return &Parser{
		supportedVersions: []string{string(VersionLM631995), string(VersionLM632002)},
	}
}

// Parse converts raw IES file data to the common photometric data model
func (p *Parser) Parse(data []byte) (*models.PhotometricData, error) {
	iesFile, err := p.parseIESFile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IES file: %w", err)
	}

	commonData, err := iesFile.ToCommonModel()
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

// DetectFormat attempts to identify if the data matches IES format
func (p *Parser) DetectFormat(data []byte) (confidence float64, version string) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) == 0 {
		return 0.0, ""
	}

	// Check for IES signature
	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "IESNA:LM-63-2002") {
		return 0.95, string(VersionLM632002)
	}
	if strings.HasPrefix(firstLine, "IESNA:LM-63-1995") {
		return 0.95, string(VersionLM631995)
	}
	if strings.HasPrefix(firstLine, "IESNA91") {
		return 0.90, string(VersionLM631995)
	}

	// Look for TILT= line which is characteristic of IES files
	confidence = 0.0
	for _, line := range lines[:min(20, len(lines))] {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TILT=") {
			confidence += 0.3
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			confidence += 0.1
		}
	}

	// Look for numeric data pattern typical of IES files
	numericLineCount := 0
	for i, line := range lines {
		if i > 50 { // Don't check too many lines
			break
		}
		line = strings.TrimSpace(line)
		if isNumericLine(line) {
			numericLineCount++
		}
	}

	if numericLineCount > 5 {
		confidence += 0.2
	}

	if confidence > 0.5 {
		return confidence, string(VersionLM632002) // Default to newer version
	}

	return confidence, ""
}

// parseIESFile parses the raw IES file data into an IESFile structure
func (p *Parser) parseIESFile(data []byte) (*IESFile, error) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	iesFile := &IESFile{
		Header: IESHeader{
			Keywords: make(map[string]string),
		},
	}

	// Parse header
	lineIndex, err := p.parseHeader(lines, iesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Parse photometric data
	err = p.parsePhotometricData(lines[lineIndex:], iesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse photometric data: %w", err)
	}

	return iesFile, nil
}

// parseHeader parses the IES file header section
func (p *Parser) parseHeader(lines []string, iesFile *IESFile) (int, error) {
	lineIndex := 0

	// Parse version line
	if lineIndex < len(lines) {
		firstLine := strings.TrimSpace(lines[lineIndex])
		if strings.HasPrefix(firstLine, "IESNA:LM-63-2002") {
			iesFile.Header.Version = VersionLM632002
		} else if strings.HasPrefix(firstLine, "IESNA:LM-63-1995") {
			iesFile.Header.Version = VersionLM631995
		} else if strings.HasPrefix(firstLine, "IESNA91") {
			iesFile.Header.Version = VersionLM631995
		} else {
			// Try to detect version from keywords or default to 2002
			iesFile.Header.Version = VersionLM632002
		}
		lineIndex++
	}

	// Parse keywords
	keywordRegex := regexp.MustCompile(`^\[([A-Z_]+)\]\s*(.*)$`)
	for lineIndex < len(lines) {
		line := strings.TrimSpace(lines[lineIndex])

		// Check for TILT line (end of keywords section)
		if strings.HasPrefix(line, "TILT=") {
			iesFile.Header.TiltData = line
			lineIndex++
			break
		}

		// Parse keyword
		if matches := keywordRegex.FindStringSubmatch(line); matches != nil {
			keyword := matches[1]
			value := matches[2]

			// Handle multi-line keywords (MORE keyword)
			if keyword == "MORE" {
				// Append to the last keyword or create a description
				if len(iesFile.Header.Keywords) > 0 {
					// Find the last keyword and append
					lastKey := ""
					for k := range iesFile.Header.Keywords {
						lastKey = k // This is not ideal, but maps don't preserve order
					}
					if lastKey != "" {
						iesFile.Header.Keywords[lastKey] += " " + value
					}
				} else {
					iesFile.Header.Keywords["DESCRIPTION"] = value
				}
			} else {
				iesFile.Header.Keywords[keyword] = value
			}
		}

		lineIndex++
	}

	return lineIndex, nil
}

// parsePhotometricData parses the photometric data section
func (p *Parser) parsePhotometricData(lines []string, iesFile *IESFile) error {
	if len(lines) < 3 {
		return fmt.Errorf("insufficient data lines for photometric section")
	}

	lineIndex := 0

	// Parse first data line
	line1Fields := strings.Fields(strings.TrimSpace(lines[lineIndex]))
	if len(line1Fields) < 10 {
		return fmt.Errorf("invalid first data line: expected 10 fields, got %d", len(line1Fields))
	}

	iesFile.Photometric.NumberOfLamps = ParseInt(line1Fields[0])
	iesFile.Photometric.LumensPerLamp = ParseFloat(line1Fields[1])
	iesFile.Photometric.CandelaMultiplier = ParseFloat(line1Fields[2])
	iesFile.Photometric.NumVerticalAngles = ParseInt(line1Fields[3])
	iesFile.Photometric.NumHorizontalAngles = ParseInt(line1Fields[4])
	iesFile.Photometric.PhotometricType = ParseInt(line1Fields[5])
	iesFile.Photometric.UnitsType = ParseInt(line1Fields[6])
	iesFile.Photometric.Width = ParseFloat(line1Fields[7])
	iesFile.Photometric.Length = ParseFloat(line1Fields[8])
	iesFile.Photometric.Height = ParseFloat(line1Fields[9])

	lineIndex++

	// Parse second data line
	line2Fields := strings.Fields(strings.TrimSpace(lines[lineIndex]))
	if len(line2Fields) >= 3 {
		iesFile.Photometric.BallastFactor = ParseFloat(line2Fields[0])
		iesFile.Photometric.BallastLampFactor = ParseFloat(line2Fields[1])
		iesFile.Photometric.InputWatts = ParseFloat(line2Fields[2])
	}

	lineIndex++

	// Parse vertical angles
	iesFile.Photometric.VerticalAngles = make([]float64, 0, iesFile.Photometric.NumVerticalAngles)
	err := p.parseFloatArray(lines, &lineIndex, &iesFile.Photometric.VerticalAngles, iesFile.Photometric.NumVerticalAngles)
	if err != nil {
		return fmt.Errorf("failed to parse vertical angles: %w", err)
	}

	// Parse horizontal angles
	iesFile.Photometric.HorizontalAngles = make([]float64, 0, iesFile.Photometric.NumHorizontalAngles)
	err = p.parseFloatArray(lines, &lineIndex, &iesFile.Photometric.HorizontalAngles, iesFile.Photometric.NumHorizontalAngles)
	if err != nil {
		return fmt.Errorf("failed to parse horizontal angles: %w", err)
	}

	// Parse candela values matrix
	iesFile.Photometric.CandelaValues = make([][]float64, iesFile.Photometric.NumVerticalAngles)
	for i := 0; i < iesFile.Photometric.NumVerticalAngles; i++ {
		iesFile.Photometric.CandelaValues[i] = make([]float64, 0, iesFile.Photometric.NumHorizontalAngles)
		err = p.parseFloatArray(lines, &lineIndex, &iesFile.Photometric.CandelaValues[i], iesFile.Photometric.NumHorizontalAngles)
		if err != nil {
			return fmt.Errorf("failed to parse candela values for vertical angle %d: %w", i, err)
		}
	}

	return nil
}

// parseFloatArray parses floating point values from multiple lines
func (p *Parser) parseFloatArray(lines []string, lineIndex *int, target *[]float64, expectedCount int) error {
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

			value, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return fmt.Errorf("invalid float value '%s': %w", field, err)
			}
			*target = append(*target, value)
		}
		*lineIndex++
	}

	if len(*target) != expectedCount {
		return fmt.Errorf("expected %d values, got %d", expectedCount, len(*target))
	}

	return nil
}

// isNumericLine checks if a line contains primarily numeric data
func isNumericLine(line string) bool {
	if line == "" {
		return false
	}

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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure Parser implements the converter.Parser interface
var _ converter.Parser = (*Parser)(nil)
