package converter

import (
	"illuminate/internal/models"
)

// Parser defines the interface for format-specific parsers
type Parser interface {
	// Parse converts raw file data to the common photometric data model
	Parse(data []byte) (*models.PhotometricData, error)

	// Validate performs format-specific validation on the parsed data
	Validate(data *models.PhotometricData) error

	// GetSupportedVersions returns the format versions supported by this parser
	GetSupportedVersions() []string

	// DetectFormat attempts to identify if the data matches this parser's format
	DetectFormat(data []byte) (confidence float64, version string)
}

// Writer defines the interface for format-specific writers
type Writer interface {
	// Write converts the common photometric data model to format-specific output
	Write(data *models.PhotometricData) ([]byte, error)

	// SetOptions configures writer-specific options
	SetOptions(opts WriterOptions) error

	// GetDefaultOptions returns the default writer options
	GetDefaultOptions() WriterOptions

	// ValidateForWrite checks if the data can be written in this format
	ValidateForWrite(data *models.PhotometricData) error
}

// WriterOptions contains configuration options for writers
type WriterOptions struct {
	// Precision controls the decimal precision for numeric values
	Precision int `json:"precision"`

	// UseCommaDecimal indicates whether to use comma as decimal separator
	UseCommaDecimal bool `json:"use_comma_decimal"`

	// IncludeComments indicates whether to include descriptive comments
	IncludeComments bool `json:"include_comments"`

	// FormatVersion specifies the target format version
	FormatVersion string `json:"format_version"`

	// CustomHeaders allows adding custom header information
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// ConversionManager orchestrates the conversion process
type ConversionManager interface {
	// Convert performs a complete conversion from source to target format
	Convert(sourceData []byte, sourceFormat, targetFormat string) (*ConversionResult, error)

	// DetectFormat automatically detects the format of the input data
	DetectFormat(data []byte) (*FormatDetectionResult, error)

	// ValidateData validates photometric data for consistency and quality
	ValidateData(data *models.PhotometricData) *ValidationResult

	// GetSupportedFormats returns a list of supported format identifiers
	GetSupportedFormats() []string

	// GetFormatInfo returns detailed information about a specific format
	GetFormatInfo(format string) (*FormatInfo, error)
}

// ConversionResult contains the result of a conversion operation
type ConversionResult struct {
	// OutputData contains the converted file data
	OutputData []byte `json:"output_data"`

	// SourceFormat is the detected or specified source format
	SourceFormat string `json:"source_format"`

	// TargetFormat is the requested target format
	TargetFormat string `json:"target_format"`

	// Warnings contains any warnings generated during conversion
	Warnings []string `json:"warnings,omitempty"`

	// Metadata contains additional information about the conversion
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// ProcessingTime indicates how long the conversion took
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// FormatDetectionResult contains the result of format detection
type FormatDetectionResult struct {
	// Format is the detected format identifier
	Format string `json:"format"`

	// Confidence is a value between 0 and 1 indicating detection confidence
	Confidence float64 `json:"confidence"`

	// Version is the detected format version
	Version string `json:"version"`

	// Alternatives contains other possible formats with their confidence scores
	Alternatives []FormatAlternative `json:"alternatives,omitempty"`
}

// FormatAlternative represents an alternative format detection result
type FormatAlternative struct {
	Format     string  `json:"format"`
	Confidence float64 `json:"confidence"`
	Version    string  `json:"version"`
}

// FormatInfo provides detailed information about a supported format
type FormatInfo struct {
	// Identifier is the unique format identifier
	Identifier string `json:"identifier"`

	// Name is the human-readable format name
	Name string `json:"name"`

	// Description provides a brief description of the format
	Description string `json:"description"`

	// SupportedVersions lists the supported versions of this format
	SupportedVersions []string `json:"supported_versions"`

	// FileExtensions lists common file extensions for this format
	FileExtensions []string `json:"file_extensions"`

	// MimeTypes lists MIME types associated with this format
	MimeTypes []string `json:"mime_types"`

	// Standards lists the standards or specifications this format follows
	Standards []string `json:"standards"`

	// Capabilities describes what features are supported
	Capabilities FormatCapabilities `json:"capabilities"`
}

// FormatCapabilities describes the capabilities of a format
type FormatCapabilities struct {
	// SupportsPhotometryTypes lists supported photometry types (A, B, C)
	SupportsPhotometryTypes []string `json:"supports_photometry_types"`

	// SupportsAbsolutePhotometry indicates if absolute photometry is supported
	SupportsAbsolutePhotometry bool `json:"supports_absolute_photometry"`

	// SupportsRelativePhotometry indicates if relative photometry is supported
	SupportsRelativePhotometry bool `json:"supports_relative_photometry"`

	// MaxVerticalAngles indicates the maximum number of vertical angles
	MaxVerticalAngles int `json:"max_vertical_angles"`

	// MaxHorizontalAngles indicates the maximum number of horizontal angles
	MaxHorizontalAngles int `json:"max_horizontal_angles"`

	// SupportsMetadata indicates if metadata fields are supported
	SupportsMetadata bool `json:"supports_metadata"`

	// SupportsGeometry indicates if geometry information is supported
	SupportsGeometry bool `json:"supports_geometry"`

	// SupportsElectrical indicates if electrical data is supported
	SupportsElectrical bool `json:"supports_electrical"`
}
