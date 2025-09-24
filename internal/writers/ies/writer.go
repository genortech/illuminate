package ies

import (
	"fmt"
	"illuminate/internal/converter"
	"illuminate/internal/models"
	"illuminate/internal/parsers/ies"
	"strconv"
	"strings"
)

// Writer implements the converter.Writer interface for IES files
type Writer struct {
	options converter.WriterOptions
}

// NewWriter creates a new IES writer instance
func NewWriter() *Writer {
	return &Writer{
		options: converter.WriterOptions{
			Precision:       1,
			UseCommaDecimal: false,
			IncludeComments: true,
			FormatVersion:   string(ies.VersionLM632002),
			CustomHeaders:   make(map[string]string),
		},
	}
}

// Write converts the common photometric data model to IES format output
func (w *Writer) Write(data *models.PhotometricData) ([]byte, error) {
	if err := w.ValidateForWrite(data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert common model to IES-specific format
	iesFile := &ies.IESFile{}
	if err := iesFile.FromCommonModel(data); err != nil {
		return nil, fmt.Errorf("failed to convert to IES format: %w", err)
	}

	// Override version if specified in options
	if w.options.FormatVersion != "" {
		iesFile.Header.Version = ies.IESVersion(w.options.FormatVersion)
	}

	// Add custom headers if provided
	for key, value := range w.options.CustomHeaders {
		iesFile.Header.Keywords[key] = value
	}

	// Generate IES file content
	var output strings.Builder

	// Write version header
	w.writeVersionHeader(&output, iesFile.Header.Version)

	// Write keywords
	w.writeKeywords(&output, iesFile.Header.Keywords)

	// Write TILT line
	output.WriteString(iesFile.Header.TiltData + "\n")

	// Write photometric data
	w.writePhotometricData(&output, &iesFile.Photometric)

	return []byte(output.String()), nil
}

// SetOptions configures writer-specific options
func (w *Writer) SetOptions(opts converter.WriterOptions) error {
	// Validate format version
	if opts.FormatVersion != "" {
		version := ies.IESVersion(opts.FormatVersion)
		if version != ies.VersionLM631995 && version != ies.VersionLM632002 {
			return fmt.Errorf("unsupported IES version: %s", opts.FormatVersion)
		}
	}

	// Validate precision
	if opts.Precision < 0 || opts.Precision > 10 {
		return fmt.Errorf("precision must be between 0 and 10, got %d", opts.Precision)
	}

	w.options = opts

	// Ensure custom headers map is initialized
	if w.options.CustomHeaders == nil {
		w.options.CustomHeaders = make(map[string]string)
	}

	return nil
}

// GetDefaultOptions returns the default writer options
func (w *Writer) GetDefaultOptions() converter.WriterOptions {
	return converter.WriterOptions{
		Precision:       1,
		UseCommaDecimal: false,
		IncludeComments: true,
		FormatVersion:   string(ies.VersionLM632002),
		CustomHeaders:   make(map[string]string),
	}
}

// ValidateForWrite checks if the data can be written in IES format
func (w *Writer) ValidateForWrite(data *models.PhotometricData) error {
	if data == nil {
		return fmt.Errorf("photometric data cannot be nil")
	}

	// Validate basic data structure
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid photometric data: %w", err)
	}

	// IES-specific validations
	if len(data.Photometry.VerticalAngles) == 0 {
		return fmt.Errorf("vertical angles cannot be empty")
	}

	if len(data.Photometry.HorizontalAngles) == 0 {
		return fmt.Errorf("horizontal angles cannot be empty")
	}

	if len(data.Photometry.CandelaValues) != len(data.Photometry.VerticalAngles) {
		return fmt.Errorf("candela values rows (%d) must match vertical angles count (%d)",
			len(data.Photometry.CandelaValues), len(data.Photometry.VerticalAngles))
	}

	for i, row := range data.Photometry.CandelaValues {
		if len(row) != len(data.Photometry.HorizontalAngles) {
			return fmt.Errorf("candela values row %d length (%d) must match horizontal angles count (%d)",
				i, len(row), len(data.Photometry.HorizontalAngles))
		}
	}

	// Check angle ranges
	for _, angle := range data.Photometry.VerticalAngles {
		if angle < 0 || angle > 180 {
			return fmt.Errorf("vertical angle %f is out of range (0-180)", angle)
		}
	}

	for _, angle := range data.Photometry.HorizontalAngles {
		if angle < 0 || angle > 360 {
			return fmt.Errorf("horizontal angle %f is out of range (0-360)", angle)
		}
	}

	return nil
}

// writeVersionHeader writes the IES version header line
func (w *Writer) writeVersionHeader(output *strings.Builder, version ies.IESVersion) {
	switch version {
	case ies.VersionLM631995:
		output.WriteString("IESNA91\n")
	case ies.VersionLM632002:
		output.WriteString("IESNA:LM-63-2002\n")
	default:
		output.WriteString("IESNA:LM-63-2002\n") // Default to 2002
	}
}

// writeKeywords writes the keyword section
func (w *Writer) writeKeywords(output *strings.Builder, keywords map[string]string) {
	// Define standard keyword order for better compatibility
	standardOrder := []string{
		"TEST", "TESTLAB", "ISSUEDATE", "MANUFAC", "LUMCAT", "LUMINAIRE",
		"LAMPCAT", "LAMP", "BALLAST", "MAINTCAT", "OTHER",
	}

	// Write keywords in standard order first
	writtenKeys := make(map[string]bool)
	for _, key := range standardOrder {
		if value, exists := keywords[key]; exists && value != "" {
			w.writeKeyword(output, key, value)
			writtenKeys[key] = true
		}
	}

	// Write remaining keywords
	for key, value := range keywords {
		if !writtenKeys[key] && value != "" {
			w.writeKeyword(output, key, value)
		}
	}
}

// writeKeyword writes a single keyword line
func (w *Writer) writeKeyword(output *strings.Builder, key, value string) {
	// Handle multi-line values by splitting on newlines
	lines := strings.Split(value, "\n")

	if len(lines) == 1 {
		output.WriteString(fmt.Sprintf("[%s] %s\n", key, value))
	} else {
		// First line with the keyword
		output.WriteString(fmt.Sprintf("[%s] %s\n", key, lines[0]))
		// Additional lines with MORE keyword
		for _, line := range lines[1:] {
			if strings.TrimSpace(line) != "" {
				output.WriteString(fmt.Sprintf("[MORE] %s\n", line))
			}
		}
	}
}

// writePhotometricData writes the photometric data section
func (w *Writer) writePhotometricData(output *strings.Builder, photometric *ies.IESPhotometricData) {
	// Line 1: 10 values
	output.WriteString(fmt.Sprintf("%d %s %s %d %d %d %d %s %s %s\n",
		photometric.NumberOfLamps,
		w.formatFloat(photometric.LumensPerLamp),
		w.formatFloat(photometric.CandelaMultiplier),
		photometric.NumVerticalAngles,
		photometric.NumHorizontalAngles,
		photometric.PhotometricType,
		photometric.UnitsType,
		w.formatFloat(photometric.Width),
		w.formatFloat(photometric.Length),
		w.formatFloat(photometric.Height)))

	// Line 2: 3 values
	output.WriteString(fmt.Sprintf("%s %s %s\n",
		w.formatFloat(photometric.BallastFactor),
		w.formatFloat(photometric.BallastLampFactor),
		w.formatFloat(photometric.InputWatts)))

	// Vertical angles
	w.writeFloatArray(output, photometric.VerticalAngles)

	// Horizontal angles
	w.writeFloatArray(output, photometric.HorizontalAngles)

	// Candela values matrix
	for _, row := range photometric.CandelaValues {
		w.writeFloatArray(output, row)
	}
}

// writeFloatArray writes an array of floats with proper formatting and line wrapping
func (w *Writer) writeFloatArray(output *strings.Builder, values []float64) {
	const maxValuesPerLine = 10

	for i, value := range values {
		if i > 0 && i%maxValuesPerLine == 0 {
			output.WriteString("\n")
		} else if i > 0 {
			output.WriteString(" ")
		}

		output.WriteString(w.formatFloat(value))
	}
	output.WriteString("\n")
}

// formatFloat formats a float value according to writer options
func (w *Writer) formatFloat(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', w.options.Precision, 64)

	if w.options.UseCommaDecimal {
		formatted = strings.Replace(formatted, ".", ",", 1)
	}

	return formatted
}

// Ensure Writer implements the converter.Writer interface
var _ converter.Writer = (*Writer)(nil)
