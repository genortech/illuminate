package database

import "time"

type PhotometricType int

const (
	PhotometricTypeC PhotometricType = 1
	PhotometricTypeB PhotometricType = 2
	PhotometricTypeA PhotometricType = 3
)

type UnitsType string

const (
	UnitsMetric   UnitsType = "Metric"
	UnitsImperial UnitsType = "Imperial"
)

type Luminaire struct {
	ID               int64           `json:"id"`
	Manufacturer     string          `json:"manufacturer"`
	Model            string          `json:"model"`
	CatalogNumber    string          `json:"catalog_number"`
	LuminaireDesc    string          `json:"luminaire_description"`
	LampType         string          `json:"lamp_type"`
	LampCatalog      string          `json:"lamp_catalog"`
	Ballast          string          `json:"ballast"`
	TestLab          string          `json:"test_lab"`
	TestNumber       string          `json:"test_number"`
	IssueDate        string          `json:"issue_date"`
	TestDate         string          `json:"test_date"`
	LuminaireCandela string          `json:"luminaire_candela"`
	LampPosition     string          `json:"lamp_position"`
	Symmetry         int             `json:"symmetry"`
	PhotometricType  PhotometricType `json:"photometric_type"`
	UnitsType        UnitsType       `json:"units_type"`
	ConversionFactor float64         `json:"conversion_factor"`
	InputWatts       float64         `json:"input_watts"`
	LuminousFlux     float64         `json:"luminous_flux"`
	ColorTemp        int             `json:"color_temp"`
	CRI              int             `json:"cri"`
	FormatType       string          `json:"format_type"`
	SymmetryFlag     int             `json:"symmetry_flag"`
	FileHash         string          `json:"file_hash"`
	OriginalFilename string          `json:"original_filename"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type PhotometricData struct {
	ID                  int64     `json:"id"`
	LuminaireID         int64     `json:"luminaire_id"`
	VerticalAngles      string    `json:"vertical_angles"`
	HorizontalAngles    string    `json:"horizontal_angles"`
	CandelaValues       string    `json:"candela_values"`
	NumVerticalAngles   int       `json:"num_vertical_angles"`
	NumHorizontalAngles int       `json:"num_horizontal_angles"`
	CreatedAt           time.Time `json:"created_at"`
}

type ParsedLuminaire struct {
	Metadata         Luminaire
	VerticalAngles   []float64
	HorizontalAngles []float64
	CandelaMatrix    [][]float64
}
