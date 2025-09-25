package ldt

import (
	"illuminate/internal/models"
	"strconv"
	"strings"
)

// EULUMDATVersion represents the supported EULUMDAT format versions
type EULUMDATVersion string

const (
	Version10 EULUMDATVersion = "1.0"
)

// LDTHeader contains the header information from an LDT file
type LDTHeader struct {
	CompanyIdentification              string  `json:"company_identification"`
	TypeIndicator                      int     `json:"type_indicator"`
	SymmetryIndicator                  int     `json:"symmetry_indicator"`
	NumberOfCPlanes                    int     `json:"number_of_c_planes"`
	DistanceBetweenCPlanes             float64 `json:"distance_between_c_planes"`
	NumberOfLuminousIntensities        int     `json:"number_of_luminous_intensities"`
	DistanceBetweenLuminousIntensities float64 `json:"distance_between_luminous_intensities"`
	MeasurementReport                  string  `json:"measurement_report"`
	LuminaireName                      string  `json:"luminaire_name"`
	LuminaireNumber                    string  `json:"luminaire_number"`
	FileName                           string  `json:"file_name"`
	DateUser                           string  `json:"date_user"`
}

// LDTGeometry contains the geometric data from an LDT file
type LDTGeometry struct {
	LengthOfLuminaire         float64 `json:"length_of_luminaire"`          // mm
	WidthOfLuminaire          float64 `json:"width_of_luminaire"`           // mm
	HeightOfLuminaire         float64 `json:"height_of_luminaire"`          // mm
	LengthOfLuminousArea      float64 `json:"length_of_luminous_area"`      // mm
	WidthOfLuminousArea       float64 `json:"width_of_luminous_area"`       // mm
	HeightOfLuminousAreaC0    float64 `json:"height_of_luminous_area_c0"`   // mm
	HeightOfLuminousAreaC90   float64 `json:"height_of_luminous_area_c90"`  // mm
	HeightOfLuminousAreaC180  float64 `json:"height_of_luminous_area_c180"` // mm
	HeightOfLuminousAreaC270  float64 `json:"height_of_luminous_area_c270"` // mm
	DownwardFluxFraction      float64 `json:"downward_flux_fraction"`       // %
	LightOutputRatioLuminaire float64 `json:"light_output_ratio_luminaire"` // %
	ConversionFactor          float64 `json:"conversion_factor"`
}

// LDTElectrical contains electrical data from an LDT file
type LDTElectrical struct {
	DRIndex          int       `json:"dr_index"`
	NumberOfLampSets int       `json:"number_of_lamp_sets"`
	LampSets         []LampSet `json:"lamp_sets"`
}

// LampSet represents a set of lamps with their characteristics
type LampSet struct {
	NumberOfLamps           int     `json:"number_of_lamps"`
	Type                    string  `json:"type"`
	TotalLuminousFlux       float64 `json:"total_luminous_flux"` // lm
	ColorTemperature        string  `json:"color_temperature"`   // K
	ColorRenderingGroup     string  `json:"color_rendering_group"`
	WattageIncludingBallast float64 `json:"wattage_including_ballast"` // W
}

// LDTPhotometry contains the photometric measurement data
type LDTPhotometry struct {
	CPlaneAngles                  []float64   `json:"c_plane_angles"`
	GammaAngles                   []float64   `json:"gamma_angles"`
	LuminousIntensityDistribution [][]float64 `json:"luminous_intensity_distribution"` // [gamma][c_plane]
}

// LDTFile represents a complete LDT file structure
type LDTFile struct {
	Header     LDTHeader     `json:"header"`
	Geometry   LDTGeometry   `json:"geometry"`
	Electrical LDTElectrical `json:"electrical"`
	Photometry LDTPhotometry `json:"photometry"`
}

// ToCommonModel converts LDT-specific data to the common photometric data model
func (ldt *LDTFile) ToCommonModel() (*models.PhotometricData, error) {
	// Convert photometric type based on symmetry indicator
	var photometryType string
	switch ldt.Header.SymmetryIndicator {
	case 0:
		photometryType = "C" // No symmetry
	case 1:
		photometryType = "C" // Symmetry about vertical axis
	case 2:
		photometryType = "C" // Symmetry about C0-C180 plane
	case 3:
		photometryType = "C" // Symmetry about C90-C270 plane
	case 4:
		photometryType = "C" // Symmetry about C0-C180 and C90-C270 planes
	default:
		photometryType = "C" // Default to Type C
	}

	// Extract metadata
	metadata := models.LuminaireMetadata{
		Manufacturer:  extractManufacturer(ldt.Header.CompanyIdentification),
		CatalogNumber: ldt.Header.LuminaireNumber,
		Description:   ldt.Header.LuminaireName,
		TestLab:       extractTestLab(ldt.Header.DateUser),
		TestDate:      extractTestDate(ldt.Header.DateUser),
		TestNumber:    ldt.Header.FileName,
	}

	// Set default values if not provided
	if metadata.Manufacturer == "" {
		metadata.Manufacturer = "Unknown"
	}
	if metadata.CatalogNumber == "" {
		metadata.CatalogNumber = "Unknown"
	}

	// Convert dimensions from mm to meters
	const mmToMeters = 0.001

	geometry := models.LuminaireGeometry{
		Length:         ldt.Geometry.LengthOfLuminaire * mmToMeters,
		Width:          ldt.Geometry.WidthOfLuminaire * mmToMeters,
		Height:         ldt.Geometry.HeightOfLuminaire * mmToMeters,
		LuminousLength: ldt.Geometry.LengthOfLuminousArea * mmToMeters,
		LuminousWidth:  ldt.Geometry.WidthOfLuminousArea * mmToMeters,
		LuminousHeight: ldt.Geometry.HeightOfLuminousAreaC0 * mmToMeters, // Use C0 as representative
	}

	// Calculate total luminous flux from all lamp sets
	var totalLumens float64
	var totalWatts float64
	for _, lampSet := range ldt.Electrical.LampSets {
		totalLumens += lampSet.TotalLuminousFlux
		totalWatts += lampSet.WattageIncludingBallast
	}

	photometry := models.PhotometricMeasurements{
		PhotometryType:    photometryType,
		UnitsType:         "absolute",
		LuminousFlux:      totalLumens,
		CandelaMultiplier: ldt.Geometry.ConversionFactor,
		VerticalAngles:    ldt.Photometry.GammaAngles,
		HorizontalAngles:  ldt.Photometry.CPlaneAngles,
		CandelaValues:     ldt.Photometry.LuminousIntensityDistribution,
	}

	electrical := models.ElectricalData{
		InputWatts:        totalWatts,
		BallastFactor:     1.0, // Default value for LDT
		BallastLampFactor: 1.0, // Default value for LDT
	}

	return &models.PhotometricData{
		Metadata:   metadata,
		Geometry:   geometry,
		Photometry: photometry,
		Electrical: electrical,
	}, nil
}

// FromCommonModel converts common photometric data to LDT-specific format
func (ldt *LDTFile) FromCommonModel(data *models.PhotometricData) error {
	// Set header information
	ldt.Header = LDTHeader{
		CompanyIdentification:              data.Metadata.Manufacturer,
		TypeIndicator:                      1, // Point source with symmetry
		SymmetryIndicator:                  0, // No symmetry (most general case)
		NumberOfCPlanes:                    len(data.Photometry.HorizontalAngles),
		DistanceBetweenCPlanes:             calculateAngleIncrement(data.Photometry.HorizontalAngles),
		NumberOfLuminousIntensities:        len(data.Photometry.VerticalAngles),
		DistanceBetweenLuminousIntensities: calculateAngleIncrement(data.Photometry.VerticalAngles),
		MeasurementReport:                  data.Metadata.Description,
		LuminaireName:                      data.Metadata.Description,
		LuminaireNumber:                    data.Metadata.CatalogNumber,
		FileName:                           data.Metadata.TestNumber,
		DateUser:                           formatDateUser(data.Metadata.TestDate, data.Metadata.TestLab),
	}

	// Convert dimensions from meters to mm
	const metersToMm = 1000.0

	ldt.Geometry = LDTGeometry{
		LengthOfLuminaire:         data.Geometry.Length * metersToMm,
		WidthOfLuminaire:          data.Geometry.Width * metersToMm,
		HeightOfLuminaire:         data.Geometry.Height * metersToMm,
		LengthOfLuminousArea:      data.Geometry.LuminousLength * metersToMm,
		WidthOfLuminousArea:       data.Geometry.LuminousWidth * metersToMm,
		HeightOfLuminousAreaC0:    data.Geometry.LuminousHeight * metersToMm,
		HeightOfLuminousAreaC90:   data.Geometry.LuminousHeight * metersToMm,
		HeightOfLuminousAreaC180:  data.Geometry.LuminousHeight * metersToMm,
		HeightOfLuminousAreaC270:  data.Geometry.LuminousHeight * metersToMm,
		DownwardFluxFraction:      100.0, // Default to 100%
		LightOutputRatioLuminaire: 100.0, // Default to 100%
		ConversionFactor:          data.Photometry.CandelaMultiplier,
	}

	// Set electrical data
	ldt.Electrical = LDTElectrical{
		DRIndex:          0, // No direct ratios
		NumberOfLampSets: 1, // Single lamp set
		LampSets: []LampSet{
			{
				NumberOfLamps:           1,
				Type:                    "LED",
				TotalLuminousFlux:       data.Photometry.LuminousFlux,
				ColorTemperature:        "3000K",
				ColorRenderingGroup:     "80",
				WattageIncludingBallast: data.Electrical.InputWatts,
			},
		},
	}

	// Set photometric data
	ldt.Photometry = LDTPhotometry{
		CPlaneAngles:                  data.Photometry.HorizontalAngles,
		GammaAngles:                   data.Photometry.VerticalAngles,
		LuminousIntensityDistribution: data.Photometry.CandelaValues,
	}

	return nil
}

// Helper functions

// extractManufacturer extracts manufacturer from company identification
func extractManufacturer(companyId string) string {
	parts := strings.Split(companyId, ";")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return companyId
}

// extractTestLab extracts test lab from date/user string
func extractTestLab(dateUser string) string {
	parts := strings.Split(dateUser, "/")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// extractTestDate extracts test date from date/user string
func extractTestDate(dateUser string) string {
	parts := strings.Split(dateUser, "/")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return dateUser
}

// formatDateUser formats date and user into LDT format
func formatDateUser(date, user string) string {
	if date == "" && user == "" {
		return ""
	}
	if user == "" {
		return date
	}
	if date == "" {
		return "/" + user
	}
	return date + "/" + user
}

// calculateAngleIncrement calculates the typical increment between angles
func calculateAngleIncrement(angles []float64) float64 {
	if len(angles) < 2 {
		return 1.0
	}

	// Calculate average increment
	totalIncrement := 0.0
	count := 0
	for i := 1; i < len(angles); i++ {
		increment := angles[i] - angles[i-1]
		if increment > 0 {
			totalIncrement += increment
			count++
		}
	}

	if count > 0 {
		return totalIncrement / float64(count)
	}
	return 1.0
}

// ParseFloat safely parses a float64 from string, handling European decimal separator
func ParseFloat(s string) float64 {
	// Replace comma with dot for European decimal separator
	s = strings.Replace(strings.TrimSpace(s), ",", ".", -1)
	if val, err := strconv.ParseFloat(s, 64); err == nil {
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
