package ies

import (
	"illuminate/internal/models"
	"strconv"
	"strings"
)

// IESVersion represents the supported IES format versions
type IESVersion string

const (
	VersionLM631995 IESVersion = "LM-63-1995"
	VersionLM632002 IESVersion = "LM-63-2002"
)

// IESHeader contains the header information from an IES file
type IESHeader struct {
	Version     IESVersion        `json:"version"`
	Keywords    map[string]string `json:"keywords"`
	TiltData    string            `json:"tilt_data"`
	RawKeywords []string          `json:"raw_keywords,omitempty"`
}

// IESPhotometricData represents the photometric data section of an IES file
type IESPhotometricData struct {
	// Line 1: Number of lamps, lumens per lamp, candela multiplier,
	// number of vertical angles, number of horizontal angles, photometric type,
	// units type, width, length, height
	NumberOfLamps       int     `json:"number_of_lamps"`
	LumensPerLamp       float64 `json:"lumens_per_lamp"`
	CandelaMultiplier   float64 `json:"candela_multiplier"`
	NumVerticalAngles   int     `json:"num_vertical_angles"`
	NumHorizontalAngles int     `json:"num_horizontal_angles"`
	PhotometricType     int     `json:"photometric_type"` // 1=Type C, 2=Type B, 3=Type A
	UnitsType           int     `json:"units_type"`       // 1=feet, 2=meters
	Width               float64 `json:"width"`
	Length              float64 `json:"length"`
	Height              float64 `json:"height"`

	// Line 2: Ballast factor, future use, input watts
	BallastFactor     float64 `json:"ballast_factor"`
	BallastLampFactor float64 `json:"ballast_lamp_factor"`
	InputWatts        float64 `json:"input_watts"`

	// Angle arrays
	VerticalAngles   []float64 `json:"vertical_angles"`
	HorizontalAngles []float64 `json:"horizontal_angles"`

	// Candela values matrix [vertical][horizontal]
	CandelaValues [][]float64 `json:"candela_values"`
}

// IESFile represents a complete IES file structure
type IESFile struct {
	Header      IESHeader          `json:"header"`
	Photometric IESPhotometricData `json:"photometric"`
}

// ToCommonModel converts IES-specific data to the common photometric data model
func (ies *IESFile) ToCommonModel() (*models.PhotometricData, error) {
	// Convert photometric type
	var photometryType string
	switch ies.Photometric.PhotometricType {
	case 1:
		photometryType = "C"
	case 2:
		photometryType = "B"
	case 3:
		photometryType = "A"
	default:
		photometryType = "C" // Default to Type C
	}

	// Convert units type
	var unitsType string
	if ies.Photometric.UnitsType == 1 {
		unitsType = "absolute" // feet
	} else {
		unitsType = "absolute" // meters
	}

	// Extract metadata from keywords
	metadata := models.LuminaireMetadata{
		Manufacturer:  getKeywordValue(ies.Header.Keywords, "MANUFAC"),
		CatalogNumber: getKeywordValue(ies.Header.Keywords, "LUMCAT"),
		Description:   getKeywordValue(ies.Header.Keywords, "LUMINAIRE"),
		TestLab:       getKeywordValue(ies.Header.Keywords, "TESTLAB"),
		TestDate:      getKeywordValue(ies.Header.Keywords, "ISSUEDATE"),
		TestNumber:    getKeywordValue(ies.Header.Keywords, "TEST"),
	}

	// Set default values if not provided
	if metadata.Manufacturer == "" {
		metadata.Manufacturer = "Unknown"
	}
	if metadata.CatalogNumber == "" {
		metadata.CatalogNumber = "Unknown"
	}

	// Convert dimensions (IES uses different units)
	var dimensionMultiplier float64 = 1.0
	if ies.Photometric.UnitsType == 1 {
		dimensionMultiplier = 0.3048 // Convert feet to meters
	}

	geometry := models.LuminaireGeometry{
		Length:         ies.Photometric.Length * dimensionMultiplier,
		Width:          ies.Photometric.Width * dimensionMultiplier,
		Height:         ies.Photometric.Height * dimensionMultiplier,
		LuminousLength: ies.Photometric.Length * dimensionMultiplier,
		LuminousWidth:  ies.Photometric.Width * dimensionMultiplier,
		LuminousHeight: ies.Photometric.Height * dimensionMultiplier,
	}

	// Calculate total luminous flux
	totalLumens := ies.Photometric.LumensPerLamp * float64(ies.Photometric.NumberOfLamps)

	photometry := models.PhotometricMeasurements{
		PhotometryType:    photometryType,
		UnitsType:         unitsType,
		LuminousFlux:      totalLumens,
		CandelaMultiplier: ies.Photometric.CandelaMultiplier,
		VerticalAngles:    ies.Photometric.VerticalAngles,
		HorizontalAngles:  ies.Photometric.HorizontalAngles,
		CandelaValues:     ies.Photometric.CandelaValues,
	}

	electrical := models.ElectricalData{
		InputWatts:        ies.Photometric.InputWatts,
		BallastFactor:     ies.Photometric.BallastFactor,
		BallastLampFactor: ies.Photometric.BallastLampFactor,
	}

	return &models.PhotometricData{
		Metadata:   metadata,
		Geometry:   geometry,
		Photometry: photometry,
		Electrical: electrical,
	}, nil
}

// FromCommonModel converts common photometric data to IES-specific format
func (ies *IESFile) FromCommonModel(data *models.PhotometricData) error {
	// Convert photometric type
	var photometricType int
	switch strings.ToUpper(data.Photometry.PhotometryType) {
	case "A":
		photometricType = 3
	case "B":
		photometricType = 2
	case "C":
		photometricType = 1
	default:
		photometricType = 1 // Default to Type C
	}

	// Set header keywords
	ies.Header.Keywords = map[string]string{
		"MANUFAC":   data.Metadata.Manufacturer,
		"LUMCAT":    data.Metadata.CatalogNumber,
		"LUMINAIRE": data.Metadata.Description,
		"TESTLAB":   data.Metadata.TestLab,
		"ISSUEDATE": data.Metadata.TestDate,
		"TEST":      data.Metadata.TestNumber,
	}

	// Set version (default to LM-63-2002)
	ies.Header.Version = VersionLM632002
	ies.Header.TiltData = "TILT=NONE"

	// Convert dimensions to feet (IES standard)
	const metersToFeet = 3.28084

	ies.Photometric = IESPhotometricData{
		NumberOfLamps:       1, // Default to 1 lamp
		LumensPerLamp:       data.Photometry.LuminousFlux,
		CandelaMultiplier:   data.Photometry.CandelaMultiplier,
		NumVerticalAngles:   len(data.Photometry.VerticalAngles),
		NumHorizontalAngles: len(data.Photometry.HorizontalAngles),
		PhotometricType:     photometricType,
		UnitsType:           1, // Use feet
		Width:               data.Geometry.Width * metersToFeet,
		Length:              data.Geometry.Length * metersToFeet,
		Height:              data.Geometry.Height * metersToFeet,
		BallastFactor:       data.Electrical.BallastFactor,
		BallastLampFactor:   data.Electrical.BallastLampFactor,
		InputWatts:          data.Electrical.InputWatts,
		VerticalAngles:      data.Photometry.VerticalAngles,
		HorizontalAngles:    data.Photometry.HorizontalAngles,
		CandelaValues:       data.Photometry.CandelaValues,
	}

	return nil
}

// getKeywordValue safely retrieves a keyword value from the map
func getKeywordValue(keywords map[string]string, key string) string {
	if value, exists := keywords[key]; exists {
		return value
	}
	return ""
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
