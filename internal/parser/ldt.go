package parser

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"illuminate/internal/database"
	"illuminate/internal/logger"
)

type LDTParser struct{}

func NewLDTParser() *LDTParser {
	return &LDTParser{}
}

func (p *LDTParser) Parse(filepath string) (*database.ParsedLuminaire, error) {
	logger.Default.Debugf("parsing LDT file: %s", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	reader := io.TeeReader(file, hash)

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	metadata := database.Luminaire{
		OriginalFilename: filepath,
		FormatType:       "LDT",
	}

	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	if len(lines) < 5 {
		return nil, fmt.Errorf("invalid LDT file: too few lines")
	}

	headerParts := strings.Split(lines[0], ";")
	if len(headerParts) < 2 || headerParts[1] != "Eulumdat2" {
		return nil, fmt.Errorf("invalid LDT file: not Eulumdat2 format")
	}

	metadata.SymmetryFlag, _ = strconv.Atoi(lines[1])
	formatType, _ := strconv.Atoi(lines[2])
	metadata.FormatType = fmt.Sprintf("Eulumdat%d", formatType)

	unitTypeLine, _ := strconv.Atoi(lines[3])
	metadata.UnitsType = "Metric"
	if unitTypeLine == 2 {
		metadata.UnitsType = "Imperial"
	}

	candelaValuesLine, _ := strconv.Atoi(lines[4])
	_ = candelaValuesLine

	lightingTechnology, _ := strconv.Atoi(lines[5])
	_ = lightingTechnology

	if len(lines) >= 8 {
		lumInfo := strings.Split(lines[7], ";")
		if len(lumInfo) >= 1 {
			metadata.LuminaireDesc = lumInfo[0]
		}
	}

	if len(lines) >= 9 {
		lumCat := strings.Split(lines[8], ";")
		if len(lumCat) >= 1 {
			metadata.Model = lumCat[0]
		}
	}

	if len(lines) >= 13 {
		if luminousFlux, err := strconv.ParseFloat(lines[12], 64); err == nil {
			metadata.LuminousFlux = luminousFlux
		}
	}

	if len(lines) >= 14 {
		if inputWatts, err := strconv.ParseFloat(lines[13], 64); err == nil {
			metadata.InputWatts = inputWatts
		}
	}

	verticalAngles := make([]float64, 0)
	horizontalAngles := make([]float64, 0)
	candelaMatrix := make([][]float64, 0)

	angleStartIdx := -1
	for i, line := range lines {
		fields := strings.Split(line, ";")
		if len(fields) > 0 {
			if val, err := strconv.ParseFloat(fields[0], 64); err == nil {
				if val >= 0 && val <= 180 {
					isAngle := true
					for _, f := range fields {
						if _, err := strconv.ParseFloat(f, 64); err != nil {
							isAngle = false
							break
						}
					}
					if isAngle && len(fields) > 1 {
						angleStartIdx = i
						break
					}
				}
			}
		}
	}

	if angleStartIdx >= 0 {
		for i := angleStartIdx; i < len(lines); i++ {
			fields := strings.Split(lines[i], ";")
			vals := make([]float64, 0)
			for _, f := range fields {
				if f == "" {
					continue
				}
				if v, err := strconv.ParseFloat(f, 64); err == nil {
					vals = append(vals, v)
				}
			}
			if len(vals) > 0 {
				if verticalAngles == nil || len(verticalAngles) == 0 {
					verticalAngles = vals
				} else if len(vals) == len(horizontalAngles) || len(horizontalAngles) == 0 {
					candelaMatrix = append(candelaMatrix, vals)
				}
			}
		}

		if len(verticalAngles) > 0 && len(candelaMatrix) > 0 {
			horizontalAngles = make([]float64, len(candelaMatrix[0]))
			for i := range horizontalAngles {
				horizontalAngles[i] = float64(i * 10)
			}
		}
	}

	fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	metadata.FileHash = fileHash

	logger.Default.Debugf("LDT parse complete: file_hash=%s, vertical_angles=%d, horizontal_angles=%d",
		fileHash, len(verticalAngles), len(horizontalAngles))

	return &database.ParsedLuminaire{
		Metadata:         metadata,
		VerticalAngles:   verticalAngles,
		HorizontalAngles: horizontalAngles,
		CandelaMatrix:    candelaMatrix,
	}, nil
}

func (p *LDTParser) Write(lum *database.ParsedLuminaire, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	writer.WriteString("WE-EF;Eulumdat2\n")
	writer.WriteString(fmt.Sprintf("%d\n", lum.Metadata.SymmetryFlag))
	writer.WriteString("3\n")

	numVert := len(lum.VerticalAngles)
	numHorz := 0
	if len(lum.CandelaMatrix) > 0 {
		numHorz = len(lum.CandelaMatrix[0])
	}
	writer.WriteString(fmt.Sprintf("%d\n", numVert))
	writer.WriteString(fmt.Sprintf("%d\n", numHorz))

	symmetry := lum.Metadata.Symmetry
	if symmetry == 0 {
		symmetry = 1
	}
	writer.WriteString(fmt.Sprintf("%d\n", symmetry))

	writer.WriteString("1\n")

	lumDesc := lum.Metadata.LuminaireDesc
	if lumDesc == "" {
		lumDesc = lum.Metadata.Model
	}
	if lumDesc == "" {
		lumDesc = "Luminaire"
	}
	writer.WriteString(fmt.Sprintf("%s\n", lumDesc))

	model := lum.Metadata.Model
	if model == "" {
		model = lumDesc
	}
	writer.WriteString(fmt.Sprintf("%s\n", model))
	writer.WriteString(fmt.Sprintf("%s\n", model))

	writer.WriteString("Generated by illuminate\n")

	flux := int(lum.Metadata.LuminousFlux)
	if flux == 0 {
		flux = 1000
	}
	writer.WriteString(fmt.Sprintf("%d\n", flux))

	watts := int(lum.Metadata.InputWatts)
	if watts == 0 {
		watts = 100
	}
	writer.WriteString(fmt.Sprintf("%d\n", watts))

	writer.WriteString("0\n")
	writer.WriteString("0\n")
	writer.WriteString("0\n")
	writer.WriteString("0\n")
	writer.WriteString("100.0\n")
	writer.WriteString("90.6\n")
	writer.WriteString("1.0\n")
	writer.WriteString("0\n")
	writer.WriteString("3\n")
	writer.WriteString("24\n")
	writer.WriteString("LED\n")

	lampType := lum.Metadata.LampType
	if lampType == "" {
		lampType = "LED"
	}
	writer.WriteString(fmt.Sprintf("%s\n", lampType))

	writer.WriteString("5800.0\n")
	writer.WriteString("4000K\n")
	writer.WriteString("70\n")
	writer.WriteString("44.5\n")
	writer.WriteString("1\n")

	writer.WriteString("0\n")
	writer.WriteString("0\n")
	writer.WriteString("0\n")
	writer.WriteString("0\n")

	for _, v := range lum.VerticalAngles {
		writer.WriteString(fmt.Sprintf("%.1f\n", v))
	}

	for _, row := range lum.CandelaMatrix {
		for i, v := range row {
			if i > 0 {
				writer.WriteString(";")
			}
			writer.WriteString(fmt.Sprintf("%.5f", v))
		}
		writer.WriteString("\n")
	}

	return nil
}
