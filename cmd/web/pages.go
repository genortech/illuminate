package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ListResponse struct {
	Luminaires []LuminaireRow `json:"luminaires"`
}

type DetailResponse struct {
	Luminaire       map[string]interface{} `json:"luminaire"`
	PhotometricData map[string]interface{} `json:"photometric_data"`
}

type LuminaireDetail struct {
	ID               int64   `json:"id"`
	Manufacturer     string  `json:"manufacturer"`
	Model            string  `json:"model"`
	CatalogNumber    string  `json:"catalog_number"`
	LuminaireDesc    string  `json:"luminaire_description"`
	LampType         string  `json:"lamp_type"`
	LampCatalog      string  `json:"lamp_catalog"`
	Ballast          string  `json:"ballast"`
	TestLab          string  `json:"test_lab"`
	TestNumber       string  `json:"test_number"`
	IssueDate        string  `json:"issue_date"`
	TestDate         string  `json:"test_date"`
	LuminaireCandela string  `json:"luminaire_candela"`
	LampPosition     string  `json:"lamp_position"`
	Symmetry         int     `json:"symmetry"`
	PhotometricType  int     `json:"photometric_type"`
	UnitsType        string  `json:"units_type"`
	ConversionFactor float64 `json:"conversion_factor"`
	InputWatts       float64 `json:"input_watts"`
	LuminousFlux     float64 `json:"luminous_flux"`
	ColorTemp        int     `json:"color_temp"`
	CRI              int     `json:"cri"`
	FormatType       string  `json:"format_type"`
	SymmetryFlag     int     `json:"symmetry_flag"`
	FileHash         string  `json:"file_hash"`
	OriginalFilename string  `json:"original_filename"`
	CreatedAt        string  `json:"created_at"`
}

type PhotometricDataDetail struct {
	ID                  int64  `json:"id"`
	LuminaireID         int64  `json:"luminaire_id"`
	VerticalAngles      string `json:"vertical_angles"`
	HorizontalAngles    string `json:"horizontal_angles"`
	CandelaValues       string `json:"candela_values"`
	NumVerticalAngles   int    `json:"num_vertical_angles"`
	NumHorizontalAngles int    `json:"num_horizontal_angles"`
}

func UploadPageHandler(c echo.Context) error {
	return UploadForm().Render(c.Request().Context(), c.Response())
}

func ListPageHandler(c echo.Context) error {
	resp, err := http.Get("http://localhost:8080/api/v1/luminaires")
	if err != nil {
		return c.String(http.StatusOK, "Failed to fetch luminaires")
	}
	defer resp.Body.Close()

	var result ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.String(http.StatusOK, "Failed to parse luminaires")
	}

	return LuminaireList(result.Luminaires).Render(c.Request().Context(), c.Response())
}

func DetailPageHandler(c echo.Context) error {
	id := c.Param("id")
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/luminaires/%s", id))
	if err != nil {
		return c.String(http.StatusOK, "Failed to fetch luminaire")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return c.String(http.StatusOK, "Luminaire not found")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.String(http.StatusOK, "Failed to parse luminaire")
	}

	lumData, ok := result["luminaire"].(map[string]interface{})
	if !ok {
		return c.String(http.StatusOK, "Invalid luminaire data")
	}

	photoData, ok := result["photometric_data"].(map[string]interface{})
	if !ok {
		photoData = make(map[string]interface{})
	}

	lum := LuminaireDetail{
		ID:               toInt64(lumData["id"]),
		Manufacturer:     toString(lumData["manufacturer"]),
		Model:            toString(lumData["model"]),
		CatalogNumber:    toString(lumData["catalog_number"]),
		LuminaireDesc:    toString(lumData["luminaire_description"]),
		LampType:         toString(lumData["lamp_type"]),
		LampCatalog:      toString(lumData["lamp_catalog"]),
		Ballast:          toString(lumData["ballast"]),
		TestLab:          toString(lumData["test_lab"]),
		TestNumber:       toString(lumData["test_number"]),
		IssueDate:        toString(lumData["issue_date"]),
		TestDate:         toString(lumData["test_date"]),
		LuminaireCandela: toString(lumData["luminaire_candela"]),
		LampPosition:     toString(lumData["lamp_position"]),
		Symmetry:         toInt(lumData["symmetry"]),
		PhotometricType:  toInt(lumData["photometric_type"]),
		UnitsType:        toString(lumData["units_type"]),
		ConversionFactor: toFloat(lumData["conversion_factor"]),
		InputWatts:       toFloat(lumData["input_watts"]),
		LuminousFlux:     toFloat(lumData["luminous_flux"]),
		ColorTemp:        toInt(lumData["color_temp"]),
		CRI:              toInt(lumData["cri"]),
		FormatType:       toString(lumData["format_type"]),
		SymmetryFlag:     toInt(lumData["symmetry_flag"]),
		FileHash:         toString(lumData["file_hash"]),
		OriginalFilename: toString(lumData["original_filename"]),
		CreatedAt:        toString(lumData["created_at"]),
	}

	photo := PhotometricDataDetail{
		ID:                  toInt64(photoData["id"]),
		LuminaireID:         toInt64(photoData["luminaire_id"]),
		VerticalAngles:      toString(photoData["vertical_angles"]),
		HorizontalAngles:    toString(photoData["horizontal_angles"]),
		CandelaValues:       toString(photoData["candela_values"]),
		NumVerticalAngles:   toInt(photoData["num_vertical_angles"]),
		NumHorizontalAngles: toInt(photoData["num_horizontal_angles"]),
	}

	return LuminaireDetailPage(lum, photo).Render(c.Request().Context(), c.Response())
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	default:
		return 0
	}
}
