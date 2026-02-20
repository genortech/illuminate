package server

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"illuminate/internal/database"
	"illuminate/internal/logger"
	"illuminate/internal/parser"
)

type LuminaireHandler struct {
	db *sql.DB
}

func NewLuminaireHandler(db database.Service) *LuminaireHandler {
	return &LuminaireHandler{db: db.GetDB()}
}

func (h *LuminaireHandler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	logger.Default.Infof("=== UPLOAD START: filename=%s ===", file.Filename)

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open file"})
	}
	defer src.Close()

	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, "tmp_"+file.Filename)
	dst, err := os.Create(tmpPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create temp file"})
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save file"})
	}

	p, err := parser.GetParser(file.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	logger.Default.Infof("parsing file: %s", tmpPath)
	lum, err := p.Parse(tmpPath)
	if err != nil {
		logger.Default.Errorf("parse failed: filename=%s, error=%v", file.Filename, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("parse error: %v", err)})
	}

	logger.Default.Infof("parsed: manufacturer=%s, model=%s, format=%s", lum.Metadata.Manufacturer, lum.Metadata.Model, lum.Metadata.FormatType)

	lum.Metadata.OriginalFilename = file.Filename
	lum.Metadata.FormatType = parser.DetectFormat(file.Filename)

	missingFields := []string{}
	if lum.Metadata.Manufacturer == "" {
		missingFields = append(missingFields, "manufacturer")
	}
	if lum.Metadata.Model == "" {
		missingFields = append(missingFields, "model")
	}

	if len(missingFields) > 0 {
		newTmpPath := filepath.Join(tmpDir, lum.Metadata.FileHash+"_"+file.Filename)
		os.Rename(tmpPath, newTmpPath)
		logger.Default.Infof("METADATA REQUIRED: filename=%s, hash=%s, missing=%v", file.Filename, lum.Metadata.FileHash, missingFields)
		logger.Default.Infof("temp file saved as: %s", newTmpPath)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "metadata_required",
			"missing":   missingFields,
			"luminaire": lum.Metadata,
			"file_hash": lum.Metadata.FileHash,
		})
	}

	os.Remove(tmpPath)
	logger.Default.Infof("saving directly: manufacturer=%s, model=%s", lum.Metadata.Manufacturer, lum.Metadata.Model)
	lumID, err := h.saveLuminaire(lum)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	logger.Default.Infof("=== UPLOAD COMPLETE: filename=%s, luminaire_id=%d ===", file.Filename, lumID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       "uploaded",
		"luminaire_id": lumID,
	})
}

func (h *LuminaireHandler) UploadWithMetadata(c echo.Context) error {
	fileHash := c.FormValue("file_hash")
	originalFilename := c.FormValue("original_filename")
	manufacturer := c.FormValue("manufacturer")
	model := c.FormValue("model")
	catalogNumber := c.FormValue("catalog_number")
	luminaireDesc := c.FormValue("luminaire_description")
	lampType := c.FormValue("lamp_type")
	testLab := c.FormValue("test_lab")
	testNumber := c.FormValue("test_number")
	issueDate := c.FormValue("issue_date")
	inputWatts := c.FormValue("input_watts")
	luminousFlux := c.FormValue("luminous_flux")

	logger.Default.Infof("=== UPLOAD WITH METADATA START ===")
	logger.Default.Infof("file_hash=%s, original_filename=%s", fileHash, originalFilename)
	logger.Default.Infof("manufacturer=%s, model=%s", manufacturer, model)
	logger.Default.Infof("catalog_number=%s, test_lab=%s, test_number=%s", catalogNumber, testLab, testNumber)
	logger.Default.Infof("input_watts=%s, luminous_flux=%s", inputWatts, luminousFlux)

	if fileHash == "" || originalFilename == "" {
		logger.Default.Error("file_hash or original_filename is empty")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file_hash and original_filename are required"})
	}

	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, fileHash+"_"+originalFilename)

	logger.Default.Infof("looking for temp file: %s", tmpPath)

	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		logger.Default.Infof("exact path not found, searching in temp dir...")
		tmpFiles, _ := os.ReadDir(tmpDir)
		for _, f := range tmpFiles {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			logger.Default.Infof("checking temp file: %s", name)
			if strings.HasPrefix(name, fileHash) {
				tmpPath = filepath.Join(tmpDir, name)
				logger.Default.Infof("found matching temp file: %s", tmpPath)
				break
			}
		}
	}

	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		logger.Default.Errorf("temp file NOT FOUND: hash=%s, filename=%s, searched=%s", fileHash, originalFilename, tmpPath)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file not found, please upload again"})
	}

	logger.Default.Infof("temp file found, parsing: %s", tmpPath)
	p, err := parser.GetParser(tmpPath)
	if err != nil {
		logger.Default.Errorf("GetParser failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	lum, err := p.Parse(tmpPath)
	if err != nil {
		logger.Default.Errorf("Parse failed: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("parse error: %v", err)})
	}

	logger.Default.Infof("parse successful, format_type=%s", lum.Metadata.FormatType)

	// Only overwrite with user input if provided
	if manufacturer != "" {
		lum.Metadata.Manufacturer = manufacturer
	}
	if model != "" {
		lum.Metadata.Model = model
	}
	if catalogNumber != "" {
		lum.Metadata.CatalogNumber = catalogNumber
	}
	if luminaireDesc != "" {
		lum.Metadata.LuminaireDesc = luminaireDesc
	}
	if lampType != "" {
		lum.Metadata.LampType = lampType
	}
	if testLab != "" {
		lum.Metadata.TestLab = testLab
	}
	if testNumber != "" {
		lum.Metadata.TestNumber = testNumber
	}
	if issueDate != "" {
		lum.Metadata.IssueDate = issueDate
	}
	lum.Metadata.OriginalFilename = originalFilename
	if w, err := strconv.ParseFloat(inputWatts, 64); err == nil && w > 0 {
		lum.Metadata.InputWatts = w
	}
	if f, err := strconv.ParseFloat(luminousFlux, 64); err == nil && f > 0 {
		lum.Metadata.LuminousFlux = f
	}

	logger.Default.Infof("saving luminaire to database: manufacturer=%s, model=%s", lum.Metadata.Manufacturer, lum.Metadata.Model)
	lumID, err := h.saveLuminaire(lum)
	if err != nil {
		logger.Default.Errorf("saveLuminaire failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	os.Remove(tmpPath)
	logger.Default.Infof("=== UPLOAD COMPLETE: luminaire_id=%d ===", lumID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       "uploaded",
		"luminaire_id": lumID,
	})
}

func (h *LuminaireHandler) saveLuminaire(lum *database.ParsedLuminaire) (int64, error) {
	db := h.db

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO luminaires (
			manufacturer, model, catalog_number, luminare_description, lamp_type,
			lamp_catalog, ballast, test_lab, test_number, issue_date, test_date,
			luminaire_candela, lamp_position, symmetry, photometric_type, units_type,
			conversion_factor, input_watts, luminous_flux, color_temp, cri,
			format_type, symmetry_flag, file_hash, original_filename
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		lum.Metadata.Manufacturer, lum.Metadata.Model, lum.Metadata.CatalogNumber,
		lum.Metadata.LuminaireDesc, lum.Metadata.LampType, lum.Metadata.LampCatalog,
		lum.Metadata.Ballast, lum.Metadata.TestLab, lum.Metadata.TestNumber,
		lum.Metadata.IssueDate, lum.Metadata.TestDate, lum.Metadata.LuminaireCandela,
		lum.Metadata.LampPosition, lum.Metadata.Symmetry, lum.Metadata.PhotometricType,
		lum.Metadata.UnitsType, lum.Metadata.ConversionFactor, lum.Metadata.InputWatts,
		lum.Metadata.LuminousFlux, lum.Metadata.ColorTemp, lum.Metadata.CRI,
		lum.Metadata.FormatType, lum.Metadata.SymmetryFlag, lum.Metadata.FileHash,
		lum.Metadata.OriginalFilename,
	)
	if err != nil {
		return 0, err
	}

	lumID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	vertAngles := fmt.Sprintf("%v", lum.VerticalAngles)
	horzAngles := fmt.Sprintf("%v", lum.HorizontalAngles)
	candelaVals := ""
	for i, row := range lum.CandelaMatrix {
		if i > 0 {
			candelaVals += ";"
		}
		for j, v := range row {
			if j > 0 {
				candelaVals += ","
			}
			candelaVals += fmt.Sprintf("%.2f", v)
		}
	}

	_, err = tx.Exec(`
		INSERT INTO photometric_data (luminaire_id, vertical_angles, horizontal_angles, candela_values, num_vertical_angles, num_horizontal_angles)
		VALUES (?, ?, ?, ?, ?, ?)`,
		lumID, vertAngles, horzAngles, candelaVals, len(lum.VerticalAngles), len(lum.HorizontalAngles),
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return lumID, nil
}

func (h *LuminaireHandler) List(c echo.Context) error {
	db := h.db

	rows, err := db.Query(`
		SELECT id, manufacturer, model, catalog_number, luminare_description,
			lamp_type, test_lab, test_number, input_watts, luminous_flux,
			format_type, original_filename, created_at
		FROM luminaires ORDER BY created_at DESC
	`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	luminaires := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var manufacturer, model, catalogNumber, lumDesc, lampType, testLab, testNumber string
		var inputWatts, luminousFlux float64
		var formatType, originalFilename, createdAt string

		err := rows.Scan(&id, &manufacturer, &model, &catalogNumber, &lumDesc,
			&lampType, &testLab, &testNumber, &inputWatts, &luminousFlux,
			&formatType, &originalFilename, &createdAt)
		if err != nil {
			continue
		}

		luminaires = append(luminaires, map[string]interface{}{
			"id":                id,
			"manufacturer":      manufacturer,
			"model":             model,
			"catalog_number":    catalogNumber,
			"description":       lumDesc,
			"lamp_type":         lampType,
			"test_lab":          testLab,
			"test_number":       testNumber,
			"input_watts":       inputWatts,
			"luminous_flux":     luminousFlux,
			"format_type":       formatType,
			"original_filename": originalFilename,
			"created_at":        createdAt,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"luminaires": luminaires,
	})
}

func (h *LuminaireHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	db := h.db

	var lum database.Luminaire
	err = db.QueryRow(`
		SELECT id, manufacturer, model, catalog_number, luminare_description,
			lamp_type, lamp_catalog, ballast, test_lab, test_number, issue_date,
			test_date, luminaire_candela, lamp_position, symmetry, photometric_type,
			units_type, conversion_factor, input_watts, luminous_flux, color_temp,
			cri, format_type, symmetry_flag, file_hash, original_filename, created_at
		FROM luminaires WHERE id = ?`, id,
	).Scan(
		&lum.ID, &lum.Manufacturer, &lum.Model, &lum.CatalogNumber, &lum.LuminaireDesc,
		&lum.LampType, &lum.LampCatalog, &lum.Ballast, &lum.TestLab, &lum.TestNumber,
		&lum.IssueDate, &lum.TestDate, &lum.LuminaireCandela, &lum.LampPosition,
		&lum.Symmetry, &lum.PhotometricType, &lum.UnitsType, &lum.ConversionFactor,
		&lum.InputWatts, &lum.LuminousFlux, &lum.ColorTemp, &lum.CRI, &lum.FormatType,
		&lum.SymmetryFlag, &lum.FileHash, &lum.OriginalFilename, &lum.CreatedAt,
	)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "luminaire not found"})
	}

	var photoData database.PhotometricData
	err = db.QueryRow(`
		SELECT id, luminaire_id, vertical_angles, horizontal_angles, candela_values,
			num_vertical_angles, num_horizontal_angles
		FROM photometric_data WHERE luminaire_id = ?`, id,
	).Scan(
		&photoData.ID, &photoData.LuminaireID, &photoData.VerticalAngles,
		&photoData.HorizontalAngles, &photoData.CandelaValues,
		&photoData.NumVerticalAngles, &photoData.NumHorizontalAngles,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get photometric data"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"luminaire":        lum,
		"photometric_data": photoData,
	})
}

func (h *LuminaireHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	db := h.db

	manufacturer := c.FormValue("manufacturer")
	model := c.FormValue("model")
	catalogNumber := c.FormValue("catalog_number")
	luminaireDesc := c.FormValue("luminaire_description")
	lampType := c.FormValue("lamp_type")
	testLab := c.FormValue("test_lab")
	testNumber := c.FormValue("test_number")
	issueDate := c.FormValue("issue_date")
	inputWatts := c.FormValue("input_watts")
	luminousFlux := c.FormValue("luminous_flux")

	_, err = db.Exec(`
		UPDATE luminaires SET
			manufacturer = COALESCE(NULLIF(?, ''), manufacturer),
			model = COALESCE(NULLIF(?, ''), model),
			catalog_number = COALESCE(NULLIF(?, ''), catalog_number),
			luminare_description = COALESCE(NULLIF(?, ''), luminare_description),
			lamp_type = COALESCE(NULLIF(?, ''), lamp_type),
			test_lab = COALESCE(NULLIF(?, ''), test_lab),
			test_number = COALESCE(NULLIF(?, ''), test_number),
			issue_date = COALESCE(NULLIF(?, ''), issue_date),
			input_watts = COALESCE(NULLIF(?, ''), input_watts),
			luminous_flux = COALESCE(NULLIF(?, ''), luminous_flux),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		manufacturer, model, catalogNumber, luminaireDesc, lampType,
		testLab, testNumber, issueDate, inputWatts, luminousFlux, id,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

func (h *LuminaireHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	db := h.db

	_, err = db.Exec("DELETE FROM luminaires WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *LuminaireHandler) Export(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	format := c.QueryParam("format")
	if format == "" {
		format = "ies"
	}

	db := h.db

	var lum database.Luminaire
	var vertAngles, horzAngles, candelaVals string

	err = db.QueryRow(`
		SELECT id, manufacturer, model, catalog_number, luminare_description,
			lamp_type, lamp_catalog, ballast, test_lab, test_number, issue_date,
			test_date, luminaire_candela, lamp_position, symmetry, photometric_type,
			units_type, conversion_factor, input_watts, luminous_flux, color_temp,
			cri, format_type, symmetry_flag, file_hash, original_filename
		FROM luminaires WHERE id = ?`, id,
	).Scan(
		&lum.ID, &lum.Manufacturer, &lum.Model, &lum.CatalogNumber, &lum.LuminaireDesc,
		&lum.LampType, &lum.LampCatalog, &lum.Ballast, &lum.TestLab, &lum.TestNumber,
		&lum.IssueDate, &lum.TestDate, &lum.LuminaireCandela, &lum.LampPosition,
		&lum.Symmetry, &lum.PhotometricType, &lum.UnitsType, &lum.ConversionFactor,
		&lum.InputWatts, &lum.LuminousFlux, &lum.ColorTemp, &lum.CRI, &lum.FormatType,
		&lum.SymmetryFlag, &lum.FileHash, &lum.OriginalFilename,
	)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "luminaire not found"})
	}

	err = db.QueryRow(`
		SELECT vertical_angles, horizontal_angles, candela_values
		FROM photometric_data WHERE luminaire_id = ?`, id,
	).Scan(&vertAngles, &horzAngles, &candelaVals)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get photometric data"})
	}

	p, err := parser.GetParser("test." + format)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	parsedLum := &database.ParsedLuminaire{
		Metadata: lum,
	}

	candelaRows := [][]float64{}
	rows := candelaVals
	if rows != "" {
		for _, rowStr := range strings.Split(rows, ";") {
			row := []float64{}
			for _, v := range strings.Split(rowStr, ",") {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					row = append(row, f)
				}
			}
			if len(row) > 0 {
				candelaRows = append(candelaRows, row)
			}
		}
	}
	parsedLum.CandelaMatrix = candelaRows

	filename := fmt.Sprintf("%s_%s.%s", lum.Manufacturer, lum.Model, format)
	if filename == "_."+format || filename == " ."+format {
		filename = fmt.Sprintf("luminaire_%d.%s", id, format)
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "application/octet-stream")

	tmpPath := filepath.Join(os.TempDir(), filename)
	if err := p.Write(parsedLum, tmpPath); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer os.Remove(tmpPath)

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Blob(http.StatusOK, "application/octet-stream", data)
}
