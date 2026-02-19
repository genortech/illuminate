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

	_ = result
	return nil
}
