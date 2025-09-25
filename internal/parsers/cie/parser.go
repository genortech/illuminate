package cie

import (
	"fmt"
	"illuminate/internal/formats/cie"
	"illuminate/internal/interfaces"
	"illuminate/internal/models"
	"strconv"
	"strings"
)

// Parser implements the interfaces.Parser interface for CIE files
type Parser struct {
	supportedVersions []string
}

// NewParser creates a new CIE parser instance
func NewParser() *Parser {
	return &Parser{
		supportedVersions: []string{string(cie.Version102), string(cie.VersionITable)},
	}
}

// Parse converts raw CIE file data to the common photometric data model
func (p *Parser) Parse(data []byte) (*models.PhotometricData, error) {
	cieFile, err := p.parseCIEFile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIE file: %w", err)
	}

	commonData, err := cieFile.ToCommonModel()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to common model: %w", err)
	}

	return commonData, nil
}

// Validate performs format-specific validation on the parsed data
func (p *Parser) Validate(data *models.PhotometricData) error {
	validator := cie.NewValidator()
	return validator.Validate(data)
}

// GetSupportedVersions returns the format versions supported by this parser
func (p *Parser) GetSupportedVersions() []string {
	return p.supportedVersions
}

// DetectFormat attempts to identify if the data matches CIE format
func (p *Parser) DetectFormat(data []byte) (confidence float64, version string) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) < 2 {
		return 0.0, ""
	}

	confidence = 0.0

	// Check first line for CIE i-table format pattern
	firstLine := strings.TrimSpace(lines[0])
	fields := strings.Fields(firstLine)

	// CIE i-table format should have 4 integers followed by description
	if len(fields) >= 4 {
		// Check if first 4 fields are integers
		intCount := 0
		for i := 0; i < 4 && i < len(fields); i++ {
			if _, err := strconv.Atoi(fields[i]); err == nil {
				intCount++
			}
		}

		if intCount == 4 {
			confidence += 0.4

			// Check if the first integer is 1 (common for CIE i-table)
			if fields[0] == "1" {
				confidence += 0.2
			}

			// Check if there's a description after the integers
			if len(fields) > 4 {
				confidence += 0.1
			}
		}
	}

	// Check for intensity data pattern (16-17 values per row, numeric)
	dataLineCount := 0
	validDataLines := 0

	for i := 1; i < len(lines) && i < 20; i++ { // Check first 20 data lines
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		dataLineCount++
		fields := strings.Fields(line)

		// CIE i-table typically has 16-17 values per row
		if len(fields) >= 16 && len(fields) <= 17 {
			confidence += 0.02

			// Check if all fields are numeric
			numericCount := 0
			for _, field := range fields {
				if _, err := strconv.Atoi(field); err == nil {
					numericCount++
				}
			}

			if numericCount >= 16 {
				validDataLines++
				confidence += 0.02
			}
		}
	}

	// If we have many valid data lines, increase confidence
	if dataLineCount > 0 && float64(validDataLines)/float64(dataLineCount) > 0.8 {
		confidence += 0.2
	}

	// Check for typical CIE file characteristics
	if strings.Contains(strings.ToLower(firstLine), "led") ||
		strings.Contains(strings.ToLower(firstLine), "lm") ||
		strings.Contains(strings.ToLower(firstLine), "w") {
		confidence += 0.1
	}

	// Check file extension if available in description
	if strings.Contains(strings.ToLower(firstLine), ".cie") {
		confidence += 0.1
	}

	if confidence > 0.3 {
		return confidence, string(cie.VersionITable)
	}

	return confidence, ""
}

// parseCIEFile parses the raw CIE file data into a CIEFile structure
func (p *Parser) parseCIEFile(data []byte) (*cie.CIEFile, error) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient lines in CIE file: expected at least 2, got %d", len(lines))
	}

	cieFile := &cie.CIEFile{}

	// Parse header (first line)
	if err := p.parseHeader(lines[0], cieFile); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Parse photometric data (remaining lines)
	if err := p.parsePhotometry(lines[1:], cieFile); err != nil {
		return nil, fmt.Errorf("failed to parse photometric data: %w", err)
	}

	return cieFile, nil
}

// parseHeader parses the CIE file header (first line)
func (p *Parser) parseHeader(headerLine string, cieFile *cie.CIEFile) error {
	line := strings.TrimSpace(headerLine)
	fields := strings.Fields(line)

	if len(fields) < 4 {
		return fmt.Errorf("invalid header line: expected at least 4 fields, got %d", len(fields))
	}

	// Parse the 4 integer fields
	formatType, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("invalid format type: %w", err)
	}

	symmetryType, err := strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("invalid symmetry type: %w", err)
	}

	reserved, err := strconv.Atoi(fields[2])
	if err != nil {
		return fmt.Errorf("invalid reserved field: %w", err)
	}

	// The description starts from the 4th field
	description := ""
	if len(fields) > 3 {
		description = strings.Join(fields[3:], " ")
	}

	cieFile.Header = cie.CIEHeader{
		FormatType:   formatType,
		SymmetryType: symmetryType,
		Reserved:     reserved,
		Description:  description,
	}

	return nil
}

// parsePhotometry parses the photometric data section
func (p *Parser) parsePhotometry(dataLines []string, cieFile *cie.CIEFile) error {
	// Set up standard CIE i-table angular grid
	cieFile.Photometry.GammaAngles = cie.GenerateStandardGammaAngles()
	cieFile.Photometry.CPlaneAngles = cie.GenerateStandardCPlaneAngles()

	// Parse intensity data
	var intensityData [][]float64

	for _, line := range dataLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		// Parse the intensity values from this line
		var rowData []float64
		for _, field := range fields {
			value, err := strconv.ParseFloat(field, 64)
			if err != nil {
				// Try parsing as integer first
				if intVal, intErr := strconv.Atoi(field); intErr == nil {
					value = float64(intVal)
				} else {
					return fmt.Errorf("invalid intensity value '%s': %w", field, err)
				}
			}
			rowData = append(rowData, value)
		}

		if len(rowData) > 0 {
			intensityData = append(intensityData, rowData)
		}
	}

	if len(intensityData) == 0 {
		return fmt.Errorf("no intensity data found")
	}

	// Validate and reshape data to match standard CIE grid
	cieFile.Photometry.IntensityData = p.reshapeIntensityData(intensityData)

	return nil
}

// reshapeIntensityData reshapes the parsed intensity data to match the standard CIE grid
func (p *Parser) reshapeIntensityData(rawData [][]float64) [][]float64 {
	expectedGammaCount := 19  // 0° to 90° in 5° steps
	expectedCPlaneCount := 16 // 0° to 337.5° in 22.5° steps

	// If the data is already in the correct shape, return as-is
	if len(rawData) == expectedGammaCount {
		allCorrectLength := true
		for _, row := range rawData {
			if len(row) != expectedCPlaneCount {
				allCorrectLength = false
				break
			}
		}
		if allCorrectLength {
			return rawData
		}
	}

	// Flatten all data into a single slice
	var flatData []float64
	for _, row := range rawData {
		flatData = append(flatData, row...)
	}

	// Calculate expected total values
	expectedTotal := expectedGammaCount * expectedCPlaneCount

	// If we don't have enough data, pad with zeros
	for len(flatData) < expectedTotal {
		flatData = append(flatData, 0.0)
	}

	// If we have too much data, truncate
	if len(flatData) > expectedTotal {
		flatData = flatData[:expectedTotal]
	}

	// Reshape into the standard grid
	result := make([][]float64, expectedGammaCount)
	dataIndex := 0

	for i := 0; i < expectedGammaCount; i++ {
		result[i] = make([]float64, expectedCPlaneCount)
		for j := 0; j < expectedCPlaneCount; j++ {
			if dataIndex < len(flatData) {
				result[i][j] = flatData[dataIndex]
				dataIndex++
			}
		}
	}

	return result
}

// Ensure Parser implements the interfaces.Parser interface
var _ interfaces.Parser = (*Parser)(nil)
