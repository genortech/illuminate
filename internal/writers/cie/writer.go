package cie

import (
	"fmt"
	"illuminate/internal/converter"
	"illuminate/internal/models"
	"illuminate/internal/parsers/cie"
	"strings"
)

// Writer implements the converter.Writer interface for CIE files
type Writer struct {
	options WriterOptions
}

// WriterOptions contains configuration options for the CIE writer
type WriterOptions struct {
	// FormatType specifies the CIE format type (default: 1 for i-table)
	FormatType int

	// SymmetryType specifies the symmetry type (0 = no symmetry, 1 = quarter symmetry)
	SymmetryType int

	// Precision specifies the number of decimal places for intensity values (default: 0 for integers)
	Precision int

	// IncludeDescription whether to include luminaire description in header
	IncludeDescription bool
}

// NewWriter creates a new CIE writer instance
func NewWriter() *Writer {
	return &Writer{
		options: WriterOptions{
			FormatType:         1,    // Standard CIE i-table format
			SymmetryType:       0,    // No symmetry (full data)
			Precision:          0,    // Integer values
			IncludeDescription: true, // Include description
		},
	}
}

// Write converts photometric data to CIE file format
func (w *Writer) Write(data *models.PhotometricData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("photometric data cannot be nil")
	}

	// Validate the data first
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("invalid photometric data: %w", err)
	}

	// Convert to CIE-specific format
	cieFile := &cie.CIEFile{}
	if err := cieFile.FromCommonModel(data); err != nil {
		return nil, fmt.Errorf("failed to convert to CIE format: %w", err)
	}

	// Override with writer options
	cieFile.Header.FormatType = w.options.FormatType
	if w.options.SymmetryType >= 0 {
		cieFile.Header.SymmetryType = w.options.SymmetryType
	}

	// Generate the CIE file content
	content, err := w.generateCIEContent(cieFile)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CIE content: %w", err)
	}

	return []byte(content), nil
}

// SetOptions configures the writer with the provided options
func (w *Writer) SetOptions(opts converter.WriterOptions) error {
	// Map converter options to CIE-specific options
	cieOpts := WriterOptions{
		FormatType:         1, // Default CIE i-table format
		SymmetryType:       0, // Default no symmetry
		Precision:          opts.Precision,
		IncludeDescription: opts.IncludeComments,
	}

	// Validate options
	if cieOpts.Precision < 0 || cieOpts.Precision > 6 {
		return fmt.Errorf("precision must be between 0 and 6, got %d", cieOpts.Precision)
	}

	w.options = cieOpts
	return nil
}

// GetDefaultOptions returns the default writer options
func (w *Writer) GetDefaultOptions() converter.WriterOptions {
	return converter.WriterOptions{
		Precision:       0,     // Integer values
		UseCommaDecimal: false, // Use dot as decimal separator
		IncludeComments: true,  // Include description in header
		FormatVersion:   "CIE i-table",
		CustomHeaders:   make(map[string]string),
	}
}

// ValidateForWrite checks if the data can be written in CIE format
func (w *Writer) ValidateForWrite(data *models.PhotometricData) error {
	return w.ValidateData(data)
}

// SetCIEOptions configures CIE-specific writer options
func (w *Writer) SetCIEOptions(opts WriterOptions) error {
	// Validate options
	if opts.FormatType < 1 {
		return fmt.Errorf("format type must be >= 1, got %d", opts.FormatType)
	}
	if opts.SymmetryType < 0 || opts.SymmetryType > 1 {
		return fmt.Errorf("symmetry type must be 0 or 1, got %d", opts.SymmetryType)
	}
	if opts.Precision < 0 || opts.Precision > 6 {
		return fmt.Errorf("precision must be between 0 and 6, got %d", opts.Precision)
	}

	w.options = opts
	return nil
}

// GetCIEOptions returns the current CIE-specific writer options
func (w *Writer) GetCIEOptions() WriterOptions {
	return w.options
}

// generateCIEContent generates the complete CIE file content
func (w *Writer) generateCIEContent(cieFile *cie.CIEFile) (string, error) {
	var content strings.Builder

	// Write header line
	headerLine := w.formatHeaderLine(cieFile.Header)
	content.WriteString(headerLine)
	content.WriteString("\n")

	// Write intensity data
	intensityContent, err := w.formatIntensityData(cieFile.Photometry.IntensityData)
	if err != nil {
		return "", fmt.Errorf("failed to format intensity data: %w", err)
	}
	content.WriteString(intensityContent)

	return content.String(), nil
}

// formatHeaderLine formats the CIE header line
func (w *Writer) formatHeaderLine(header cie.CIEHeader) string {
	// Format: "   1   0   0        Description"
	headerLine := fmt.Sprintf("%4d%4d%4d", header.FormatType, header.SymmetryType, header.Reserved)

	if w.options.IncludeDescription && header.Description != "" {
		headerLine += "        " + header.Description
	}

	return headerLine
}

// formatIntensityData formats the intensity data matrix
func (w *Writer) formatIntensityData(intensityData [][]float64) (string, error) {
	if len(intensityData) == 0 {
		return "", fmt.Errorf("intensity data cannot be empty")
	}

	var content strings.Builder

	for rowIndex, row := range intensityData {
		if len(row) == 0 {
			continue
		}

		var rowValues []string
		for _, value := range row {
			formattedValue := w.formatIntensityValue(value)
			rowValues = append(rowValues, formattedValue)
		}

		// Join values with spaces
		rowLine := strings.Join(rowValues, " ")
		content.WriteString(rowLine)

		// Add newline except for the last row
		if rowIndex < len(intensityData)-1 {
			content.WriteString("\n")
		}
	}

	return content.String(), nil
}

// formatIntensityValue formats a single intensity value according to precision settings
func (w *Writer) formatIntensityValue(value float64) string {
	if w.options.Precision == 0 {
		// Format as integer with minimum width of 4 characters
		return fmt.Sprintf("%4.0f", value)
	} else {
		// Format as float with specified precision, no minimum width for decimal values
		format := fmt.Sprintf("%%.%df", w.options.Precision)
		return fmt.Sprintf(format, value)
	}
}

// ValidateData performs CIE-specific validation on the input data
func (w *Writer) ValidateData(data *models.PhotometricData) error {
	validator := cie.NewValidator()
	return validator.Validate(data)
}

// GetSupportedFormats returns the formats supported by this writer
func (w *Writer) GetSupportedFormats() []string {
	return []string{"CIE", "cie"}
}

// GetFormatDescription returns a description of the CIE format
func (w *Writer) GetFormatDescription() string {
	return "CIE photometric file format (i-table) - International Commission on Illumination standard"
}

// Ensure Writer implements the converter.Writer interface
var _ converter.Writer = (*Writer)(nil)
