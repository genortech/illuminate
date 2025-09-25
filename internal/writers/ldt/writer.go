package ldt

import (
	"fmt"
	"illuminate/internal/converter"
	"illuminate/internal/models"
	"illuminate/internal/parsers/ldt"
	"strconv"
	"strings"
)

// Writer implements the converter.Writer interface for LDT files
type Writer struct {
	options converter.WriterOptions
}

// NewWriter creates a new LDT writer instance
func NewWriter() *Writer {
	return &Writer{
		options: converter.WriterOptions{
			Precision:       1,
			UseCommaDecimal: true, // LDT format typically uses European decimal separator
			IncludeComments: false,
			FormatVersion:   string(ldt.Version10),
			CustomHeaders:   make(map[string]string),
		},
	}
}

// Write converts the common photometric data model to LDT format output
func (w *Writer) Write(data *models.PhotometricData) ([]byte, error) {
	if err := w.ValidateForWrite(data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert common model to LDT-specific format
	ldtFile := &ldt.LDTFile{}
	if err := ldtFile.FromCommonModel(data); err != nil {
		return nil, fmt.Errorf("failed to convert to LDT format: %w", err)
	}

	// Apply custom headers if provided
	w.applyCustomHeaders(ldtFile)

	// Generate LDT file content
	var output strings.Builder

	// Write header section (lines 1-13)
	w.writeHeader(&output, &ldtFile.Header)

	// Write geometry section (lines 14-26)
	w.writeGeometry(&output, &ldtFile.Geometry)

	// Write electrical section (lines 27+)
	w.writeElectrical(&output, ldtFile.Electrical)

	// Write photometric data
	w.writePhotometry(&output, ldtFile.Photometry)

	return []byte(output.String()), nil
}

// SetOptions configures writer-specific options
func (w *Writer) SetOptions(opts converter.WriterOptions) error {
	// Validate format version
	if opts.FormatVersion != "" {
		version := ldt.EULUMDATVersion(opts.FormatVersion)
		if version != ldt.Version10 {
			return fmt.Errorf("unsupported LDT version: %s", opts.FormatVersion)
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
		UseCommaDecimal: true,
		IncludeComments: false,
		FormatVersion:   string(ldt.Version10),
		CustomHeaders:   make(map[string]string),
	}
}

// ValidateForWrite checks if the data can be written in LDT format
func (w *Writer) ValidateForWrite(data *models.PhotometricData) error {
	if data == nil {
		return fmt.Errorf("photometric data cannot be nil")
	}

	// Validate basic data structure
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid photometric data: %w", err)
	}

	// LDT-specific validations
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

	// Check for reasonable data limits (EULUMDAT format constraints)
	if len(data.Photometry.HorizontalAngles) > 360 {
		return fmt.Errorf("too many horizontal angles (%d), maximum is 360", len(data.Photometry.HorizontalAngles))
	}

	if len(data.Photometry.VerticalAngles) > 181 {
		return fmt.Errorf("too many vertical angles (%d), maximum is 181", len(data.Photometry.VerticalAngles))
	}

	return nil
}

// applyCustomHeaders applies custom headers from options to the LDT file
func (w *Writer) applyCustomHeaders(ldtFile *ldt.LDTFile) {
	for key, value := range w.options.CustomHeaders {
		switch strings.ToLower(key) {
		case "company", "manufacturer":
			ldtFile.Header.CompanyIdentification = value
		case "luminaire", "description":
			ldtFile.Header.LuminaireName = value
		case "catalog", "number":
			ldtFile.Header.LuminaireNumber = value
		case "filename":
			ldtFile.Header.FileName = value
		case "report":
			ldtFile.Header.MeasurementReport = value
		case "date", "dateuser":
			ldtFile.Header.DateUser = value
		}
	}
}

// writeHeader writes the LDT header section (lines 1-13)
func (w *Writer) writeHeader(output *strings.Builder, header *ldt.LDTHeader) {
	// Line 1: Company identification
	output.WriteString(header.CompanyIdentification + "\n")

	// Line 2: Type indicator
	output.WriteString(fmt.Sprintf("%d\n", header.TypeIndicator))

	// Line 3: Symmetry indicator
	output.WriteString(fmt.Sprintf("%d\n", header.SymmetryIndicator))

	// Line 4: Number of C planes
	output.WriteString(fmt.Sprintf("%d\n", header.NumberOfCPlanes))

	// Line 5: Distance between C planes
	output.WriteString(w.formatFloat(header.DistanceBetweenCPlanes) + "\n")

	// Line 6: Number of luminous intensities
	output.WriteString(fmt.Sprintf("%d\n", header.NumberOfLuminousIntensities))

	// Line 7: Distance between luminous intensities
	output.WriteString(w.formatFloat(header.DistanceBetweenLuminousIntensities) + "\n")

	// Line 8: Measurement report
	output.WriteString(header.MeasurementReport + "\n")

	// Line 9: Luminaire name
	output.WriteString(header.LuminaireName + "\n")

	// Line 10: Luminaire number
	output.WriteString(header.LuminaireNumber + "\n")

	// Line 11: File name
	output.WriteString(header.FileName + "\n")

	// Line 12: Date/User
	output.WriteString(header.DateUser + "\n")
}

// writeGeometry writes the geometry section (lines 13-25)
func (w *Writer) writeGeometry(output *strings.Builder, geometry *ldt.LDTGeometry) {
	// Line 13: Length of luminaire
	output.WriteString(w.formatFloat(geometry.LengthOfLuminaire) + "\n")

	// Line 14: Width of luminaire
	output.WriteString(w.formatFloat(geometry.WidthOfLuminaire) + "\n")

	// Line 15: Height of luminaire
	output.WriteString(w.formatFloat(geometry.HeightOfLuminaire) + "\n")

	// Line 16: Length of luminous area
	output.WriteString(w.formatFloat(geometry.LengthOfLuminousArea) + "\n")

	// Line 17: Width of luminous area
	output.WriteString(w.formatFloat(geometry.WidthOfLuminousArea) + "\n")

	// Line 18: Height of luminous area C0
	output.WriteString(w.formatFloat(geometry.HeightOfLuminousAreaC0) + "\n")

	// Line 19: Height of luminous area C90
	output.WriteString(w.formatFloat(geometry.HeightOfLuminousAreaC90) + "\n")

	// Line 20: Height of luminous area C180
	output.WriteString(w.formatFloat(geometry.HeightOfLuminousAreaC180) + "\n")

	// Line 21: Height of luminous area C270
	output.WriteString(w.formatFloat(geometry.HeightOfLuminousAreaC270) + "\n")

	// Line 22: Downward flux fraction
	output.WriteString(w.formatFloat(geometry.DownwardFluxFraction) + "\n")

	// Line 23: Light output ratio luminaire
	output.WriteString(w.formatFloat(geometry.LightOutputRatioLuminaire) + "\n")

	// Line 24: Conversion factor
	output.WriteString(w.formatFloat(geometry.ConversionFactor) + "\n")

	// Line 25: Additional geometry field (placeholder for EULUMDAT compatibility)
	output.WriteString("0\n")

	// Line 26: Additional geometry field (placeholder for EULUMDAT compatibility)
	output.WriteString("0\n")
}

// writeElectrical writes the electrical section
func (w *Writer) writeElectrical(output *strings.Builder, electrical ldt.LDTElectrical) {
	// Line 25: DR index
	output.WriteString(fmt.Sprintf("%d\n", electrical.DRIndex))

	// Line 26: Number of lamp sets
	output.WriteString(fmt.Sprintf("%d\n", electrical.NumberOfLampSets))

	// Write each lamp set (6 lines per set)
	for _, lampSet := range electrical.LampSets {
		// Number of lamps
		output.WriteString(fmt.Sprintf("%d\n", lampSet.NumberOfLamps))

		// Type
		output.WriteString(lampSet.Type + "\n")

		// Total luminous flux
		output.WriteString(w.formatFloat(lampSet.TotalLuminousFlux) + "\n")

		// Color temperature
		output.WriteString(lampSet.ColorTemperature + "\n")

		// Color rendering group
		output.WriteString(lampSet.ColorRenderingGroup + "\n")

		// Wattage including ballast
		output.WriteString(w.formatFloat(lampSet.WattageIncludingBallast) + "\n")
	}
}

// writePhotometry writes the photometric data section
func (w *Writer) writePhotometry(output *strings.Builder, photometry ldt.LDTPhotometry) {
	// Write C plane angles
	w.writeFloatArray(output, photometry.CPlaneAngles)

	// Write gamma angles
	w.writeFloatArray(output, photometry.GammaAngles)

	// Write luminous intensity distribution
	// LDT format expects data organized by C-plane first, then by gamma angle
	// Our common model stores it as [gamma][c_plane], so we need to transpose
	for cPlane := 0; cPlane < len(photometry.CPlaneAngles); cPlane++ {
		var cPlaneData []float64
		for gamma := 0; gamma < len(photometry.GammaAngles); gamma++ {
			if gamma < len(photometry.LuminousIntensityDistribution) &&
				cPlane < len(photometry.LuminousIntensityDistribution[gamma]) {
				cPlaneData = append(cPlaneData, photometry.LuminousIntensityDistribution[gamma][cPlane])
			} else {
				cPlaneData = append(cPlaneData, 0.0) // Default to 0 if data is missing
			}
		}
		w.writeFloatArray(output, cPlaneData)
	}
}

// writeFloatArray writes an array of floats, one per line
func (w *Writer) writeFloatArray(output *strings.Builder, values []float64) {
	for _, value := range values {
		output.WriteString(w.formatFloat(value) + "\n")
	}
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
